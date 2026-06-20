package writer

import (
	"database/sql"
	"RssViewer/internal/db"
	_ "github.com/mattn/go-sqlite3"
)

type DBWriter struct {
	db *sql.DB
}

func NewDBwriter(db *sql.DB) (*DBWriter, error) {
	return &DBWriter{db: db}, nil
}

func (w *DBWriter) CreateTopic(model *db.TopicModel) error {
	
	return nil
}

func (w *DBWriter) CreateRss(model *db.RSSModel) error {

	return nil
}

func (w *DBWriter) CreatePost(model *db.PostModel) error {

	return nil
}



func (w *DBWriter) Update() error {
	return nil
}

func (w *DBWriter) Delete() error {
	return nil
}
