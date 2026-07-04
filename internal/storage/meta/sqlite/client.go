package meta

import (
	"RssViewer/internal/model"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type SqliteDb struct {
	Db *sql.DB
}

// GetDB exposes the underlying raw connection pool pointer safely to your internal readers and writers
func (s *SqliteDb) GetDB() *sql.DB {
	return s.Db
}

func InitDB(dbPath string) (*SqliteDb, error) {
	if dbPath == "" {
		return nil, errors.New("database path configuration must not be empty")
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory tree structure for database: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to establish connection connection handle: %w", err)
	}

	pragmas := []string{
		"PRAGMA journal_mode=WAL;",
		"PRAGMA foreign_keys=ON;",
		"PRAGMA busy_timeout=5000;",
	}
	for _, p := range pragmas {
		if _, err := db.Exec(p); err != nil {
			_ = db.Close() // Close connection to prevent socket leakage on failure loops
			return nil, fmt.Errorf("failed to execute pragma configuration instruction %q: %w", p, err)
		}
	}

	// DO NOT set open connections to 1.
	// Your single goroutine worker pattern in orchestrate.go regulates sequential writes.
	// Allowing open connections here unlocks true parallel performance for non-blocking WAL reads.
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(10)
	if err := model.ApplySchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &SqliteDb{Db: db}, nil
}
