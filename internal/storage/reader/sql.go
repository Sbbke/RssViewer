package reader

import (
	"database/sql"
	"errors"
	"fmt"

	"RssViewer/internal/db"
	"RssViewer/internal/dto"
)

type DBReader struct {
	db *sql.DB
}

func NewDBReader(database *sql.DB) (*DBReader, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return &DBReader{
		db: database,
	}, nil
}

func (r *DBReader) GetTopics() ([]dto.TopicResponse, error) {
	const q = `SELECT id, name, created_at FROM topic ORDER BY created_at DESC`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("GetTopics: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	var topics []dto.TopicResponse
	for rows.Next() {
		var m db.TopicModel
		if err := rows.Scan(&m.ID, &m.Name, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetTopics: scan: %w", err)
		}
		topics = append(topics, dto.TopicResponse{
			TopicID:   m.ID,
			Name:      m.Name,
			CreatedAt: m.CreatedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetTopics dataset iteration disrupted: %w", err)
	}

	return topics, nil
}

func (r *DBReader) GetTopicByID(id int64) (dto.TopicResponse, error) {
	const q = `SELECT id, name, created_at FROM topic WHERE id = ?`

	var m db.TopicModel
	err := r.db.QueryRow(q, id).Scan(&m.ID, &m.Name, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.TopicResponse{}, fmt.Errorf("GetTopicByID %d: %w", id, sql.ErrNoRows)
		}

		return dto.TopicResponse{}, fmt.Errorf("GetTopicByID %d: %w", id, err)
	}

	return dto.TopicResponse{
		TopicID:   m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
	}, nil

}

func (r *DBReader) GetRssByTopic(topicID int64) ([]dto.RssItem, error) {
	const q = `
		SELECT id, title, created_at
		FROM rss
		WHERE topic_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.Query(q, topicID)
	if err != nil {
		return nil, fmt.Errorf("GetRssByTopic %d: query: %w", topicID, err)
	}
	defer rows.Close()

	// Initialize with an empty slice instead of a nil slice so JSON serializations 
	// return an empty array "[]" instead of "null" to the frontend client application
	items := make([]dto.RssItem, 0)
	
	for rows.Next() {
		var m db.RSSModel
		if err := rows.Scan(&m.ID, &m.Title, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetRssByTopic %d: scan: %w", topicID, err)
		}
		items = append(items, dto.RssItem{
			ID:           m.ID,
			Title:        m.Title,
			SubscribedAt: m.CreatedAt,
		})
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetRssByTopic %d: rows iteration error: %w", topicID, err)
	}
	
	return items, nil
}

// GetRssByID extracts a consolidated feed model alongside its nested structural post entities
func (r *DBReader) GetRssByID(rssID int64) (dto.RssResponse, error) {
	const qRss = `SELECT id, title, created_at FROM rss WHERE id = ?`

	var m db.RSSModel
	err := r.db.QueryRow(qRss, rssID).Scan(&m.ID, &m.Title, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.RssResponse{}, fmt.Errorf("GetRssByID %d: reference entry missing: %w", rssID, err)
		}
		return dto.RssResponse{}, fmt.Errorf("GetRssByID %d: database evaluation failed: %w", rssID, err)
	}

	posts, err := r.GetPostsByRss(rssID)
	if err != nil {
		return dto.RssResponse{}, fmt.Errorf("GetRssByID %d: nested payload extraction failure: %w", rssID, err)
	}

	return dto.RssResponse{
		Info: dto.RssItem{
			ID:           m.ID,
			Title:        m.Title,
			SubscribedAt: m.CreatedAt,
		},
		Posts: posts,
	}, nil
}

func (r *DBReader) GetPostsByRss(rssID int64) ([]dto.PostItem, error) {
	const q = `
		SELECT id, title, published_at
		FROM post
		WHERE source_id = ?
		ORDER BY published_at DESC`
 
	rows, err := r.db.Query(q, rssID)
	if err != nil {
		return nil, fmt.Errorf("GetPostsByRss %d: execution query failed: %w", rssID, err)
	}
	defer rows.Close()
 
	// Preallocate slice with initial capacity bucket to minimize runtime allocations
	posts := make([]dto.PostItem, 0, 32)
	for rows.Next() {
		var m db.PostModel
		if err := rows.Scan(&m.ID, &m.Title, &m.PublishedAt); err != nil {
			return nil, fmt.Errorf("GetPostsByRss %d: row iteration scan failed: %w", rssID, err)
		}
		posts = append(posts, dto.PostItem{
			ID:          m.ID,
			Title:       m.Title,
			PublishedAt: m.PublishedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsByRss %d: final rows iteration context error: %w", rssID, err)
	}

	return posts, nil
}
 
// GetPostByID returns a PostItem for a single post
// Use this when the service needs metadata before deciding whether to load the full content from disk
func (r *DBReader) GetPostByID(postID int64) (dto.PostItem, error) {
	const q = `SELECT id, title, published_at FROM post WHERE id = ?`
 
	var m db.PostModel
	err := r.db.QueryRow(q, postID).Scan(&m.ID, &m.Title, &m.PublishedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.PostItem{}, fmt.Errorf("GetPostByID %d: targeted entry not found: %w", postID, err)
		}
		return dto.PostItem{}, fmt.Errorf("GetPostByID %d: scanning failed: %w", postID, err)
	}

	return dto.PostItem{
		ID:          m.ID,
		Title:       m.Title,
		PublishedAt: m.PublishedAt,
	}, nil
}
 
// GetPostsByTopic joins posts against the RSS configuration mapping
// Used when the summary service aggregates post pointers into a single processing context window
func (r *DBReader) GetPostsByTopic(topicID int64) ([]dto.PostItem, error) {
	const q = `
		SELECT p.id, p.title, p.published_at
		FROM post p
		JOIN rss r ON p.source_id = r.id
		WHERE r.topic_id = ?
		ORDER BY p.published_at DESC`
 
	rows, err := r.db.Query(q, topicID)
	if err != nil {
		return nil, fmt.Errorf("GetPostsByTopic %d: execution query failed: %w", topicID, err)
	}
	defer rows.Close()
 
	posts := make([]dto.PostItem, 0, 64)
	for rows.Next() {
		var m db.PostModel
		if err := rows.Scan(&m.ID, &m.Title, &m.PublishedAt); err != nil {
			return nil, fmt.Errorf("GetPostsByTopic %d: row iteration scan failed: %w", topicID, err)
		}
		posts = append(posts, dto.PostItem{
			ID:          m.ID,
			Title:       m.Title,
			PublishedAt: m.PublishedAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsByTopic %d: final rows iteration context error: %w", topicID, err)
	}

	return posts, nil
}
