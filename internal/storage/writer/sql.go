package writer

import (
	"RssViewer/internal/db"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type DBWriter struct {
	db *sql.DB
}

func NewDBwriter(db *sql.DB) (*DBWriter, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	return &DBWriter{db: db}, nil
}

func (w *DBWriter) CreateTopic(model *db.TopicModel) error {
	const q = `
		INSERT INTO topic (name, created_at)
		VALUES (?,?)
	`
	res, err := w.db.Exec(q, model.Name, model.CreatedAt)
	if err != nil {
		return fmt.Errorf("CreateTopic: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("CreateTopic: retrieve id: %w", err)
	}
	model.ID = id
	return nil
}
func (w *DBWriter) UpdateTopic(model *db.TopicModel) error {
	const q = `
		UPDATE topic
		SET    name = ?
		WHERE  id   = ?
	`
	res, err := w.db.Exec(q, model.Name, model.ID)
	if err != nil {
		return fmt.Errorf("UpdateTopic: %w", err)
	}
	return requireOneRow(res, "UpdateTopic", model.ID)
}

func (w *DBWriter) DeleteTopic(id int64) error {
	const q = `DELETE FROM topic WHERE id = ?`
	res, err := w.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("delete topic: %w", err)
	}

	return requireOneRow(res, "DeleteTopic", id)
}

func (w *DBWriter) CreateRss(model *db.RSSModel) error {
	const q = `
		INSERT INTO rss (topic_id, title, url, created_at)
		VALUES (?,?,?,?)
		`
	res, err := w.db.Exec(q, model.TopicID, model.Title, model.Url, model.CreatedAt)
	if err != nil {
		return fmt.Errorf("create rss: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("create rss: retrieve id: %w", err)
	}

	model.ID = id

	return nil
}

func (w *DBWriter) UpdateRss(model *db.RSSModel) error {
	const q = `
		UPDATE rss
		SET title = ?
		    url = ?
		WHERE id = ?
		`

	res, err := w.db.Exec(q, model.Title, model.Url, model.ID)
	if err != nil {
		return fmt.Errorf("update rss: %w", err)
	}

	return requireOneRow(res, "UpdateRss", model.ID)

}

func (w *DBWriter) DeleteRss(id int64) error {
	const q = `DELETE FROM rss WHERE id = ?`
	res, err := w.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("delete rss: %w", err)
	}

	return requireOneRow(res, "DeleteRss", id)
}
func (w *DBWriter) CreatePost(model *db.PostModel) error {
	const q = `
		INSERT INTO post (source_id, title, url, created_at, published_at)
		VALUES (?,?,?,?,?)
		`

	res, err := w.db.Exec(q, model.RssID, model.Title, model.Url, model.CreatedAt, model.PublishedAt)
	if err != nil {
		return fmt.Errorf("create post: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("create post: retrieve id : %w", err)
	}
	model.ID = id

	return nil
}

func (w *DBWriter) UpdatePostTitle(title string, id int64) error {
	const q = `
		UPDATE post SET title = ? WHERE id = ?
		`
	res, err := w.db.Exec(q, title, id)
	if err != nil {
		return fmt.Errorf("update post : %w", err)
	}

	return requireOneRow(res, "UpdatePostTitle", id)
}
func (w *DBWriter) DeletePost(id int64) error {
	const q = `DELETE FROM post WHERE id = ?`
	res, err := w.db.Exec(q, id)
	if err != nil {
		return fmt.Errorf("delete post: %w", err)
	}

	return requireOneRow(res, "DeletePost", id)
}

func (w *DBWriter) CreatePostsBatch(models []*db.PostModel) (finalErr error) {
	if len(models) == 0 {
		return nil
	}
	tx, err := w.db.Begin()
	if err != nil {
		return fmt.Errorf("begin post batch: begin tx: %w", err)
	}

	defer func() {
		if finalErr != nil {
			_ = tx.Rollback()
		}
	}()

	const q = `
		INSERT INTO post (source_id, title, url, created_at, published_at)
		VALUES (?,?,?,?,?)`

	stmt, err := tx.Prepare(q)

	if err != nil {
		return fmt.Errorf("create batch: prepare: %w", err)
	}
	defer func() {
		if err := stmt.Close(); err != nil && finalErr == nil {
			finalErr = fmt.Errorf("error closing stmt")
		}
	}()

	now := time.Now()

	for _, m := range models {
		if m.CreatedAt.IsZero() {
			m.CreatedAt = now
		}
		res, exeErr := stmt.Exec(m.RssID, m.Title, m.Url, m.CreatedAt, m.PublishedAt)

		if exeErr != nil {
			finalErr = fmt.Errorf("create post batch: insert %s:%w", m.Title, exeErr)
			return finalErr
		}

		id, idErr := res.LastInsertId()
		if idErr != nil {
			finalErr = fmt.Errorf("create post batch: retrieve id for %s: %w", m.Title, err)
			return finalErr

		}

		m.ID = id
	}

	if err = tx.Commit(); err != nil {
		finalErr = fmt.Errorf("create post batch: commit: %w", err)
		return finalErr
	}

	return nil
}

// requireOneRow evaluates mutation metrics and raises errors if an entity reference was missing
func requireOneRow(res sql.Result, op string, id int64) error {
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: rows affected lookup failed: %w", op, err)
	}
	if n == 0 {
		return fmt.Errorf("%s: target database record not modified, entity id %d does not exist", op, id)
	}
	return nil
}
