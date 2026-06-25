package reader

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"

	"RssViewer/internal/dto"
	"RssViewer/internal/model"
)

type DBReader struct {
	db *sql.DB
}

func NewDBReader(database *sql.DB) (*DBReader, error) {
	if database == nil {
		return nil, fmt.Errorf("database cannot be nil")
	}
	return &DBReader{db: database}, nil
}

// ---------------------------------------------------------------------------
// Topic
// ---------------------------------------------------------------------------

func (r *DBReader) GetTopics() ([]dto.TopicResponse, error) {
	const q = `SELECT id, name, created_at FROM topic ORDER BY created_at DESC`

	rows, err := r.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("GetTopics: %w", err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	var topics []dto.TopicResponse
	for rows.Next() {
		m, err := scanTopic(rows)
		if err != nil {
			return nil, fmt.Errorf("GetTopics: %w", err)
		}
		topics = append(topics, topicModelToResponse(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetTopics: rows: %w", err)
	}
	return topics, nil
}

func (r *DBReader) GetTopicByID(id int64) (dto.TopicResponse, error) {
	const q = `SELECT id, name, created_at FROM topic WHERE id = ?`

	var m model.TopicModel
	err := r.db.QueryRow(q, id).Scan(&m.ID, &m.Name, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.TopicResponse{}, fmt.Errorf("GetTopicByID %d: %w", id, sql.ErrNoRows)
		}
		return dto.TopicResponse{}, fmt.Errorf("GetTopicByID %d: %w", id, err)
	}
	return topicModelToResponse(m), nil
}

// GetTopicWithRss returns a TopicResponse with its nested RssItems populated.
// All data is SQL-sourced so assembly belongs here, not in the service layer.
func (r *DBReader) GetTopicWithRss(id int64) (dto.TopicResponse, error) {
	topic, err := r.GetTopicByID(id)
	if err != nil {
		return dto.TopicResponse{}, fmt.Errorf("GetTopicWithRss %d: %w", id, err)
	}
	rssItems, err := r.GetRssByTopic(id)
	if err != nil {
		return dto.TopicResponse{}, fmt.Errorf("GetTopicWithRss %d: rss: %w", id, err)
	}
	topic.Rss = rssItems
	return topic, nil
}

// ---------------------------------------------------------------------------
// RSS
// ---------------------------------------------------------------------------

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
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	items := make([]dto.RssItem, 0) // empty slice: JSON → [] not null
	for rows.Next() {
		m, err := scanRss(rows)
		if err != nil {
			return nil, fmt.Errorf("GetRssByTopic %d: %w", topicID, err)
		}
		items = append(items, rssModelToItem(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetRssByTopic %d: rows: %w", topicID, err)
	}
	return items, nil
}

func (r *DBReader) GetRssByID(rssID int64) (dto.RssResponse, error) {
	const q = `SELECT id, title, created_at FROM rss WHERE id = ?`

	var m model.RSSModel
	err := r.db.QueryRow(q, rssID).Scan(&m.ID, &m.Title, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.RssResponse{}, fmt.Errorf("GetRssByID %d: not found: %w", rssID, sql.ErrNoRows)
		}
		return dto.RssResponse{}, fmt.Errorf("GetRssByID %d: %w", rssID, err)
	}
	posts, err := r.GetPostsByRss(rssID)
	if err != nil {
		return dto.RssResponse{}, fmt.Errorf("GetRssByID %d: posts: %w", rssID, err)
	}
	return dto.RssResponse{
		Info:  rssModelToItem(m),
		Posts: posts,
	}, nil
}

// GetRssURL returns the raw XML feed URL for a given RSS ID.
// Used by the crawler (CheckUpdate) to know where to fetch the remote feed.
// Returns a plain string — no DTO wrapper needed for a single scalar value.
func (r *DBReader) GetRssURL(rssID int64) (string, error) {
	const q = `SELECT url FROM rss WHERE id = ?`

	var url string
	err := r.db.QueryRow(q, rssID).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("GetRssURL %d: not found: %w", rssID, sql.ErrNoRows)
		}
		return "", fmt.Errorf("GetRssURL %d: %w", rssID, err)
	}
	return url, nil
}

// ---------------------------------------------------------------------------
// Post
// ---------------------------------------------------------------------------

func (r *DBReader) GetPostsByRss(rssID int64) ([]dto.PostItem, error) {
	const q = `
		SELECT id, title, published_at
		FROM post
		WHERE source_id = ?
		ORDER BY published_at DESC`

	rows, err := r.db.Query(q, rssID)
	if err != nil {
		return nil, fmt.Errorf("GetPostsByRss %d: query: %w", rssID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	posts := make([]dto.PostItem, 0, 32)
	for rows.Next() {
		m, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("GetPostsByRss %d: %w", rssID, err)
		}
		posts = append(posts, postModelToItem(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsByRss %d: rows: %w", rssID, err)
	}
	return posts, nil
}

func (r *DBReader) GetPostByID(postID int64) (dto.PostItem, error) {
	const q = `SELECT id, title, published_at FROM post WHERE id = ?`

	var m model.PostModel
	err := r.db.QueryRow(q, postID).Scan(&m.ID, &m.Title, &m.PublishedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return dto.PostItem{}, fmt.Errorf("GetPostByID %d: not found: %w", postID, sql.ErrNoRows)
		}
		return dto.PostItem{}, fmt.Errorf("GetPostByID %d: %w", postID, err)
	}
	return postModelToItem(m), nil
}

// GetPostURL returns the landing page URL for a post.
// Used by the HTML processor (GetContent) as its fetch starting point.
// Returns a plain string — no DTO wrapper needed for a single scalar value.
func (r *DBReader) GetPostURL(postID int64) (string, error) {
	const q = `SELECT url FROM post WHERE id = ?`

	var url string
	err := r.db.QueryRow(q, postID).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("GetPostURL %d: not found: %w", postID, sql.ErrNoRows)
		}
		return "", fmt.Errorf("GetPostURL %d: %w", postID, err)
	}
	return url, nil
}

func (r *DBReader) GetPostsByTopic(topicID int64) ([]dto.PostItem, error) {
	const q = `
		SELECT p.id, p.title, p.published_at
		FROM post p
		JOIN rss r ON p.source_id = r.id
		WHERE r.topic_id = ?
		ORDER BY p.published_at DESC`

	rows, err := r.db.Query(q, topicID)
	if err != nil {
		return nil, fmt.Errorf("GetPostsByTopic %d: query: %w", topicID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	posts := make([]dto.PostItem, 0, 64)
	for rows.Next() {
		m, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("GetPostsByTopic %d: %w", topicID, err)
		}
		posts = append(posts, postModelToItem(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsByTopic %d: rows: %w", topicID, err)
	}
	return posts, nil
}

// GetPostsByTopicInWindow returns posts published within [from, to) for a topic.
// Primary feed for GenerateTopicBriefing — the summary service passes a week or
// month window and batches the result into inference context windows.
// ASC order gives the LLM a coherent chronological narrative.
func (r *DBReader) GetPostsByTopicInWindow(topicID int64, from, to time.Time) ([]dto.PostItem, error) {
	const q = `
		SELECT p.id, p.title, p.published_at
		FROM post p
		JOIN rss r ON p.source_id = r.id
		WHERE r.topic_id = ?
		  AND p.published_at >= ?
		  AND p.published_at <  ?
		ORDER BY p.published_at ASC`

	rows, err := r.db.Query(q, topicID, from, to)
	if err != nil {
		return nil, fmt.Errorf("GetPostsByTopicInWindow topic=%d: query: %w", topicID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	posts := make([]dto.PostItem, 0, 64)
	for rows.Next() {
		m, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("GetPostsByTopicInWindow topic=%d: %w", topicID, err)
		}
		posts = append(posts, postModelToItem(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsByTopicInWindow topic=%d: rows: %w", topicID, err)
	}
	return posts, nil
}

// GetPostsByRssInWindow is the per-feed equivalent of GetPostsByTopicInWindow.
// Used by GeneratePostBriefing when scoped to a single RSS source.
func (r *DBReader) GetPostsByRssInWindow(rssID int64, from, to time.Time) ([]dto.PostItem, error) {
	const q = `
		SELECT id, title, published_at
		FROM post
		WHERE source_id = ?
		  AND published_at >= ?
		  AND published_at <  ?
		ORDER BY published_at ASC`

	rows, err := r.db.Query(q, rssID, from, to)
	if err != nil {
		return nil, fmt.Errorf("GetPostsByRssInWindow rss=%d: query: %w", rssID, err)
	}
	defer func() {
		if err := rows.Close(); err != nil {
			log.Printf("failed to close rows: %v", err)
		}
	}()
	posts := make([]dto.PostItem, 0, 32)
	for rows.Next() {
		m, err := scanPost(rows)
		if err != nil {
			return nil, fmt.Errorf("GetPostsByRssInWindow rss=%d: %w", rssID, err)
		}
		posts = append(posts, postModelToItem(m))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsByRssInWindow rss=%d: rows: %w", rssID, err)
	}
	return posts, nil
}

// ---------------------------------------------------------------------------
// Private scanners
//
// Each scanner owns the Scan call for its model type. Adding a column means
// updating one function rather than every query loop that returns that type.
// The *sql.Rows parameter satisfies both Query loops and could be adapted to
// QueryRow via an interface if needed.
// ---------------------------------------------------------------------------

type scanner interface {
	Scan(dest ...any) error
}

func scanTopic(s scanner) (model.TopicModel, error) {
	var m model.TopicModel
	if err := s.Scan(&m.ID, &m.Name, &m.CreatedAt); err != nil {
		return model.TopicModel{}, fmt.Errorf("scan topic: %w", err)
	}
	return m, nil
}

func scanRss(s scanner) (model.RSSModel, error) {
	var m model.RSSModel
	if err := s.Scan(&m.ID, &m.Title, &m.CreatedAt); err != nil {
		return model.RSSModel{}, fmt.Errorf("scan rss: %w", err)
	}
	return m, nil
}

func scanPost(s scanner) (model.PostModel, error) {
	var m model.PostModel
	if err := s.Scan(&m.ID, &m.Title, &m.PublishedAt); err != nil {
		return model.PostModel{}, fmt.Errorf("scan post: %w", err)
	}
	return m, nil
}

// ---------------------------------------------------------------------------
// Private mappers — model → DTO
//
// Callers above the storage layer never see db.*Model types.
// Schema changes (column rename, new field) are isolated to these functions.
// ---------------------------------------------------------------------------

func topicModelToResponse(m model.TopicModel) dto.TopicResponse {
	return dto.TopicResponse{
		TopicID:   m.ID,
		Name:      m.Name,
		CreatedAt: m.CreatedAt,
		// Rss, Summary, SummaryID left at zero — caller populates if needed
	}
}

func rssModelToItem(m model.RSSModel) dto.RssItem {
	return dto.RssItem{
		ID:           m.ID,
		Title:        m.Title,
		SubscribedAt: m.CreatedAt,
	}
}

func postModelToItem(m model.PostModel) dto.PostItem {
	return dto.PostItem{
		ID:          m.ID,
		Title:       m.Title,
		PublishedAt: m.PublishedAt,
	}
}
