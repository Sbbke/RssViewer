package meta

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"RssViewer/internal/dto"
	"RssViewer/internal/model"
)

type DBWriter struct {
	db *sql.DB
}

func NewDBWriter(database *sql.DB) (*DBWriter, error) {
	if database == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &DBWriter{db: database}, nil
}

// ---------------------------------------------------------------------------
// Topic
// ---------------------------------------------------------------------------

func (w *DBWriter) CreateTopic(payload dto.TopicPayload) (dto.MutationResult, error) {
	m := topicPayloadToModel(payload)

	const q = `INSERT INTO topic (name, created_at) VALUES (?, ?)`
	res, err := w.db.Exec(q, m.Name, m.CreatedAt)
	if err != nil {
		return dto.MutationResult{}, fmt.Errorf("CreateTopic: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return dto.MutationResult{}, fmt.Errorf("CreateTopic: retrieve id: %w", err)
	}
	return dto.MutationResult{GeneratedID: id}, nil
}

func (w *DBWriter) UpdateTopic(id int64, payload dto.TopicPayload) error {
	const q = `UPDATE topic SET name = ? WHERE id = ?`
	res, err := w.db.Exec(q, payload.Name, id)
	if err != nil {
		return fmt.Errorf("UpdateTopic: %w", err)
	}
	return requireOneRow(res, "UpdateTopic", id)
}

func (w *DBWriter) DeleteTopic(id int64) error {
	const q = `DELETE FROM topic WHERE id = ?`
	res, err := w.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("DeleteTopic: %w", err)
	}
	return requireOneRow(res, "DeleteTopic", id)
}

// ---------------------------------------------------------------------------
// RSS
// ---------------------------------------------------------------------------

func (w *DBWriter) CreateRss(payload dto.RssPayload) (dto.MutationResult, error) {
    if strings.TrimSpace(payload.Url) == "" {
        return dto.MutationResult{}, fmt.Errorf("CreateRss: payload URL cannot be empty")
    }
	fmt.Print(payload.Url)
	m := rssPayloadToModel(payload)
	fmt.Println(m.Url)
	const q = `INSERT INTO rss ( title, xml, url, created_at) VALUES (?, ?, ?, ?)`
	res, err := w.db.Exec(q, m.Title, m.Xml, m.Url, m.CreatedAt)
	if err != nil {
		return dto.MutationResult{}, fmt.Errorf("CreateRss: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return dto.MutationResult{}, fmt.Errorf("CreateRss: retrieve id: %w", err)
	}
	return dto.MutationResult{GeneratedID: id}, nil
}

func (w *DBWriter) UpdateRss(id int64, payload dto.RssPayload) error {
	const q = `UPDATE rss SET title = ?, xml = ? WHERE id = ?`
	res, err := w.db.Exec(q, payload.Title, payload.Xml, id)
	if err != nil {
		return fmt.Errorf("UpdateRss: %w", err)
	}
	return requireOneRow(res, "UpdateRss", id)
}

func (w *DBWriter) DeleteRss(id int64) error {
	const q = `DELETE FROM rss WHERE id = ?`
	res, err := w.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("DeleteRss: %w", err)
	}
	return requireOneRow(res, "DeleteRss", id)
}

// ---------------------------------------------------------------------------
// Post
// ---------------------------------------------------------------------------

func (w *DBWriter) CreatePost(payload dto.PostPayload) (dto.MutationResult, error) {
	m := postPayloadToModel(payload)

	const q = `
		INSERT INTO post (source_id, title, url, content, created_at, published_at)
		VALUES (?, ?, ?, ?, ?, ?)`
	res, err := w.db.Exec(q, m.RssID, m.Title, m.Url, m.Content, m.CreatedAt, m.PublishedAt)
	if err != nil {
		return dto.MutationResult{}, fmt.Errorf("CreatePost: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return dto.MutationResult{}, fmt.Errorf("CreatePost: retrieve id: %w", err)
	}
	return dto.MutationResult{GeneratedID: id}, nil
}

func (w *DBWriter) UpdatePostTitle(id int64, title string) error {
	const q = `UPDATE post SET title = ? WHERE id = ?`
	res, err := w.db.Exec(q, title, id)
	if err != nil {
		return fmt.Errorf("UpdatePostTitle: %w", err)
	}
	return requireOneRow(res, "UpdatePostTitle", id)
}

func (w *DBWriter) DeletePost(id int64) error {
	const q = `DELETE FROM post WHERE id = ?`
	res, err := w.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("DeletePost: %w", err)
	}
	return requireOneRow(res, "DeletePost", id)
}

// CreatePostsBatch inserts a slice of PostPayloads in a single transaction.
// On any failure the transaction is rolled back; no partial writes reach the DB.
// Each MutationResult in the returned slice corresponds by index to the input payload.

func (w *DBWriter) CreatePostsBatch(payloads []dto.PostPayload) (_ []dto.MutationResult, finalErr error) {
	if len(payloads) == 0 {
		return nil, nil
	}

	tx, err := w.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("CreatePostsBatch: begin tx: %w", err)
	}
	// Secure deferred stack unwind safety
	defer func() {
		if finalErr != nil {
			_ = tx.Rollback()
		}
	}()

	// 1. Fully aligned 6 positional argument mappings matching destination tables
	const q = `
		INSERT INTO post (source_id, title, url, content, created_at, published_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(source_id, url) DO NOTHING
		`

	stmt, err := tx.Prepare(q)
	if err != nil {
		return nil, fmt.Errorf("CreatePostsBatch: prepare statement layout: %w", err)
	}
	defer func() {
		if closeErr := stmt.Close(); closeErr != nil && finalErr == nil {
			finalErr = fmt.Errorf("CreatePostsBatch: close prepared statement handle: %w", closeErr)
		}
	}()

	now := time.Now()
	results := make([]dto.MutationResult, 0, len(payloads))

	for _, p := range payloads {
		// Convert DTO layer structural elements down to Model domain values
		m := postPayloadToModel(p)

		// Fallback timestamp generation validations
		if m.CreatedAt.IsZero() {
			m.CreatedAt = now
		}

		// 2. FIXED: Explicitly provide all 6 bound input parameter rows including content
		res, execErr := stmt.Exec(m.RssID, m.Title, m.Url, m.Content, m.CreatedAt, m.PublishedAt)
		if execErr != nil {
			finalErr = fmt.Errorf("CreatePostsBatch: insert transaction processing failure for %q: %w", m.Title, execErr)
			return nil, finalErr
		}

		id, idErr := res.LastInsertId()
		if idErr != nil {
			finalErr = fmt.Errorf("CreatePostsBatch: retrieve generated database primary key identity for %q: %w", m.Title, idErr)
			return nil, finalErr
		}

		results = append(results, dto.MutationResult{GeneratedID: id})
	}

	if err = tx.Commit(); err != nil {
		finalErr = fmt.Errorf("CreatePostsBatch: commit transaction state boundaries: %w", err)
		return nil, finalErr
	}

	return results, nil
}

// summary
func (w *DBWriter) CreatePostSummary(p dto.SummaryPayload) (dto.MutationResult, error) {
    const q = `
        INSERT INTO post_summary (post_id, body, created_at, updated_at)
        VALUES (?, ?, ?, ?)`
    now := time.Now()
    res, err := w.db.Exec(q, p.ID, p.Content, now, now)
    if err != nil {
        return dto.MutationResult{}, fmt.Errorf("CreatePostSummary: %w", err)
    }
    id, err := res.LastInsertId()
    if err != nil {
        return dto.MutationResult{}, fmt.Errorf("CreatePostSummary: retrieve id: %w", err)
    }
    return dto.MutationResult{GeneratedID: id}, nil
}

func (w *DBWriter) UpdatePostSummary(p dto.SummaryPayload) error {
    const q = `UPDATE post_summary SET body = ?, updated_at = ? WHERE post_id = ?`
    res, err := w.db.Exec(q, p.Content, time.Now(), p.ID)
    if err != nil {
        return fmt.Errorf("UpdatePostSummary: %w", err)
    }
    return requireOneRow(res, "UpdatePostSummary", p.ID)
}

func (w *DBWriter) DeletePostSummary(postID int64) error {
    const q = `DELETE FROM post_summary WHERE post_id = ?`
    res, err := w.db.Exec(q, postID)
    if err != nil {
        return fmt.Errorf("DeletePostSummary: %w", err)
    }
    return requireOneRow(res, "DeletePostSummary", postID)
}

func (w *DBWriter) CreateTopicSummary(p dto.SummaryPayload) (dto.MutationResult, error) {
    const q = `
        INSERT INTO topic_summary (topic_id, body, created_at, updated_at)
        VALUES (?, ?, ?, ?)`
    now := time.Now()
    res, err := w.db.Exec(q, p.ID, p.Content, now, now)
    if err != nil {
        return dto.MutationResult{}, fmt.Errorf("CreateTopicSummary: %w", err)
    }
    id, err := res.LastInsertId()
    if err != nil {
        return dto.MutationResult{}, fmt.Errorf("CreateTopicSummary: retrieve id: %w", err)
    }
    return dto.MutationResult{GeneratedID: id}, nil
}

func (w *DBWriter) UpdateTopicSummary(p dto.SummaryPayload) error {
    const q = `UPDATE topic_summary SET body = ?, updated_at = ? WHERE topic_id = ?`
    res, err := w.db.Exec(q, p.ID, time.Now(), p.Content)
    if err != nil {
        return fmt.Errorf("UpdateTopicSummary: %w", err)
    }
    return requireOneRow(res, "UpdateTopicSummary", p.ID)
}

func (w *DBWriter) DeleteTopicSummary(topicID int64) error {
    const q = `DELETE FROM topic_summary WHERE topic_id = ?`
    res, err := w.db.Exec(q, topicID)
    if err != nil {
        return fmt.Errorf("DeleteTopicSummary: %w", err)
    }
    return requireOneRow(res, "DeleteTopicSummary", topicID)
}

func (w *DBWriter) LinkRssTopic(rssID, topicID int64) error{
	const q = `INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`
	if _, err := w.db.Exec(q, rssID, topicID); err != nil {
		return fmt.Errorf("LinkRssTopic rss=%d topic=%d: %w", rssID, topicID, err)
	}
	return nil
}

 func (w *DBWriter) UnlinkRssTopic(rssID, topicID int64) error {
	const q = `DELETE FROM rss_topics WHERE rss_id = ? AND topic_id = ?`
	if _, err := w.db.Exec(q, rssID, topicID); err != nil {
		return fmt.Errorf("UnlinkRssTopic rss=%d topic=%d: %w", rssID, topicID, err)
	}
	return nil
}
// ---------------------------------------------------------------------------
// Private mappers — DTO → model
//
// time.Now() is stamped here so callers never need to think about created_at.
// These are the only places in the package that construct db.*Model values.
// ---------------------------------------------------------------------------

func topicPayloadToModel(p dto.TopicPayload) model.TopicModel {
	return model.TopicModel{
		Name:      p.Name,
		CreatedAt: time.Now(),
	}
}

func rssPayloadToModel(p dto.RssPayload) model.RSSModel {
	return model.RSSModel{
		Title:     p.Title,
		Url:       p.Url,
		Xml:       p.Xml,
		CreatedAt: time.Now(),
	}
}

func postPayloadToModel(p dto.PostPayload) model.PostModel {
	return model.PostModel{
		RssID:       p.RssID,
		Title:       p.Title,
		Url:         p.Url,
		Content:     p.Content,
		CreatedAt:   time.Now(),
		PublishedAt: p.PublishedAt,
	}
}

// ---------------------------------------------------------------------------
// requireOneRow evaluates mutation metrics and raises an error if the target
// record was missing (zero rows affected).
// ---------------------------------------------------------------------------

func requireOneRow(res sql.Result, op string, id int64) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: rows affected lookup failed: %w", op, err)
	}
	if n == 0 {
		return fmt.Errorf("%s: entity id %d does not exist", op, id)
	}
	return nil
}
