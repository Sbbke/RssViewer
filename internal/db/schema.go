package db

import (
	"time"
)

// TopicModel represents the "Topic" table entity
type TopicModel struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`	
}

// RSSModel represents the "RSS" table entity
type RSSModel struct {
	ID        int64     `db:"id"`
	TopicID	  int64	    `db:"tag"` // Foreign Key -> TopicModel.ID
	Title     string    `db:"title"`
	Url       string    `db:"url"` // The raw XML feed source URL
	CreatedAt time.Time `db:"created_at"`	
}


// PostModel represents the "Post" table entity pointer
type PostModel struct {
	ID        int64     `db:"id"`
	RssID	  int64     `db:"source_id"` // Foreign Key -> RSSModel.ID
	Title     string    `db:"title"`
	Url       string    `db:"url"` // The unique target website landing page
	CreatedAt time.Time `db:"created_at"`	
}
