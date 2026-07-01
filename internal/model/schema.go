package model

import (
	"database/sql"
	"fmt"
	"time"
)

// TopicModel represents the "Topic" table entity
type TopicModel struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time`db:"created_at"`	
}

// RSSModel represents the "RSS" table entity
type RSSModel struct {
	ID        int64     `db:"id"`
	Title     string    `db:"title"`
	Url       string    `db:"url"` // The raw XML feed source URL
	Xml []byte`db:"xml"` // raw xml file
	CreatedAt time.Time`db:"created_at"`	

}


// PostModel represents the "Post" table entity pointer
type PostModel struct {
	ID        int64     `db:"id"`
	RssID	  int64     `db:"source_id"` // Foreign Key -> RSSModel.ID
	Title     string    `db:"title"`
	Url       string    `db:"url"` // The unique target website landing page
	Content string `db:"content"` // raw processed content
	CreatedAt time.Time `db:"created_at"`	
	PublishedAt string `db:"published_at"`
}

const SchemaSQL = `
PRAGMA journal_mode=WAL;
PRAGMA foreign_keys=ON;
PRAGMA busy_timeout=5000;

CREATE TABLE IF NOT EXISTS topic (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    name       TEXT     NOT NULL UNIQUE,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS rss (
    id         INTEGER PRIMARY KEY AUTOINCREMENT,
    title      TEXT     NOT NULL,
    xml        BLOB,
    url        TEXT     NOT NULL,
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS rss_topics (
    rss_id   INTEGER NOT NULL REFERENCES rss(id) ON DELETE CASCADE,
    topic_id INTEGER NOT NULL REFERENCES topic(id) ON DELETE CASCADE,
    PRIMARY KEY (rss_id, topic_id)
);

CREATE TABLE IF NOT EXISTS post (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id    INTEGER NOT NULL REFERENCES rss(id) ON DELETE CASCADE,
    title        TEXT    NOT NULL,
    url          TEXT    NOT NULL,
    content      TEXT    NOT NULL,
    created_at   DATETIME NOT NULL,
    published_at DATETIME NOT NULL
);
`

func ApplySchema(db *sql.DB) error {
    if _, err := db.Exec(SchemaSQL); err != nil {
        return fmt.Errorf("ApplySchema: %w", err)
    }
    return nil
}
