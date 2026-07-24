package storage

import (
	"RssViewer/internal/model"
	"database/sql"
	"fmt"
)


type sqliteAccessor struct{
	DB *sql.DB
}

func (s *sqliteAccessor) GetDB() *sql.DB {
	return s.DB
}

func NewSqliteDB(path string) (*sqliteAccessor, error) {
	dsn := fmt.Sprintf("file:%s?_foreign_keys=on", path)

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}

	if err := model.ApplySchema(db); err != nil {
		_ = db.Close()
		return nil, err
	}

	return &sqliteAccessor{DB: db}, nil
}
