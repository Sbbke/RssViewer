package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type sqliteDb struct{
	db *sql.DB
}
func InitDB(dbPath string) (*sqliteDb, error) {
    if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
        return nil, fmt.Errorf("failed to create directory tree for database file: %w", err)
	}
    // Open connection and enforce foreign key constraints along with WAL mode
    db, err := sql.Open("sqlite3", dbPath)
    if err != nil {
        return nil, err
    }
    
    if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
        return nil, err
    }
	db.SetMaxOpenConns(1)    
	return &sqliteDb{ db : db}, nil
}

