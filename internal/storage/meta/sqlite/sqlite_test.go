package meta_test

import (
	"database/sql"
	"errors"
	"testing"
	"time"
	 
	_ "github.com/mattn/go-sqlite3"
 
	"RssViewer/internal/dto"
	"RssViewer/internal/storage/meta/sqlite"
)


func testDB(t *testing.T) (*meta.DBReader, *meta.DBWriter){

	t.Helper()
	db,err := sql.Open("sqlite3", ":memory:")

	if err != nil{
		t.Fatalf("open: memory : %v",err)
	}

	t.Cleanup(func() {db.Close()})

	if _, err := db.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		t.Fatalf("enable foreign keys: %v", err)
	}

		schema := `
		CREATE TABLE topic (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			name       TEXT    NOT NULL,
			created_at DATETIME NOT NULL
		);
		CREATE TABLE rss (
			id         INTEGER PRIMARY KEY AUTOINCREMENT,
			topic_id   INTEGER NOT NULL REFERENCES topic(id) ON DELETE CASCADE,
			title      TEXT    NOT NULL,
			url        TEXT    NOT NULL,
			created_at DATETIME NOT NULL
		);
		CREATE TABLE post (
			id           INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id    INTEGER NOT NULL REFERENCES rss(id) ON DELETE CASCADE,
			title        TEXT    NOT NULL,
			url          TEXT    NOT NULL,
			created_at   DATETIME NOT NULL,
			published_at DATETIME NOT NULL
		);`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}
 
	r, err := meta.NewDBReader(db)
	if err != nil {
		t.Fatalf("new reader: %v", err)
	}
	w, err := meta.NewDBWriter(db)
	if err != nil {
		t.Fatalf("new writer: %v", err)
	}
	return r, w
}
 
// seed inserts one topic → one rss → one post and returns their IDs.
// Used by tests that need existing rows without caring about the insert itself.
func seed(t *testing.T, w *meta.DBWriter) (topicID, rssID, postID int64) {
	t.Helper()
 
	tr, err := w.CreateTopic(dto.TopicPayload{Name: "AI"})
	if err != nil {
		t.Fatalf("seed topic: %v", err)
	}
	topicID = tr.GeneratedID
 
	rr, err := w.CreateRss(dto.RssPayload{
		TopicID: topicID,
		Title:   "arXiv ML",
		Url:     "https://arxiv.org/rss/cs.LG",
	})
	if err != nil {
		t.Fatalf("seed rss: %v", err)
	}
	rssID = rr.GeneratedID
 
	pr, err := w.CreatePost(dto.PostPayload{
		RssID:       rssID,
		Title:       "Attention Is All You Need",
		Url:         "https://arxiv.org/abs/1706.03762",
		PublishedAt: time.Date(2017, 6, 12, 0, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("seed post: %v", err)
	}
	postID = pr.GeneratedID
	return
}
 
// ---------------------------------------------------------------------------
// Writer — Topic
// ---------------------------------------------------------------------------
 
func TestCreateTopic(t *testing.T) {
	_, w := testDB(t)
 
	result, err := w.CreateTopic(dto.TopicPayload{Name: "ML"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GeneratedID <= 0 {
		t.Errorf("expected positive GeneratedID, got %d", result.GeneratedID)
	}
}
 
func TestUpdateTopic(t *testing.T) {
	r, w := testDB(t)
	topicID, _, _ := seed(t, w)
 
	if err := w.UpdateTopic(topicID, dto.TopicPayload{Name: "Renamed"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
 
	topic, err := r.GetTopicByID(topicID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if topic.Name != "Renamed" {
		t.Errorf("name: got %q want %q", topic.Name, "Renamed")
	}
}
 
func TestUpdateTopic_NotFound(t *testing.T) {
	_, w := testDB(t)
 
	err := w.UpdateTopic(9999, dto.TopicPayload{Name: "Ghost"})
	if err == nil {
		t.Fatal("expected error for missing topic, got nil")
	}
}
 
func TestDeleteTopic(t *testing.T) {
	r, w := testDB(t)
	topicID, _, _ := seed(t, w)
 
	if err := w.DeleteTopic(topicID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
 
	_, err := r.GetTopicByID(topicID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}
 
func TestDeleteTopic_CascadesToRssAndPost(t *testing.T) {
	r, w := testDB(t)
	topicID, rssID, _ := seed(t, w)
 
	if err := w.DeleteTopic(topicID); err != nil {
		t.Fatalf("delete topic: %v", err)
	}
 
	// Cascade should have removed the child RSS row
	_, err := r.GetRssByID(rssID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected rss row to be cascade-deleted, got %v", err)
	}
}
 
// ---------------------------------------------------------------------------
// Writer — RSS
// ---------------------------------------------------------------------------
 
func TestCreateRss(t *testing.T) {
	_, w := testDB(t)
	topicID, _, _ := seed(t, w)
 
	result, err := w.CreateRss(dto.RssPayload{
		TopicID: topicID,
		Title:   "Second feed",
		Url:     "https://example.com/feed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GeneratedID <= 0 {
		t.Errorf("expected positive GeneratedID, got %d", result.GeneratedID)
	}
}
 
func TestCreateRss_ForeignKeyViolation(t *testing.T) {
	_, w := testDB(t)
 
	// topic_id 9999 does not exist — FK constraint must fire
	_, err := w.CreateRss(dto.RssPayload{
		TopicID: 9999,
		Title:   "Orphan",
		Url:     "https://example.com",
	})
	if err == nil {
		t.Fatal("expected FK violation error, got nil")
	}
}
 
func TestUpdateRss(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	err := w.UpdateRss(rssID, dto.RssPayload{
		Title: "Updated Title",
		Url:   "https://new.example.com/feed",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
 
	resp, err := r.GetRssByID(rssID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if resp.Info.Title != "Updated Title" {
		t.Errorf("title: got %q want %q", resp.Info.Title, "Updated Title")
	}
}
 
func TestDeleteRss_CascadesToPost(t *testing.T) {
	r, w := testDB(t)
	_, rssID, postID := seed(t, w)
 
	if err := w.DeleteRss(rssID); err != nil {
		t.Fatalf("delete rss: %v", err)
	}
 
	_, err := r.GetPostByID(postID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected post to be cascade-deleted, got %v", err)
	}
}
 
// ---------------------------------------------------------------------------
// Writer — Post
// ---------------------------------------------------------------------------
 
func TestCreatePost(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	result, err := w.CreatePost(dto.PostPayload{
		RssID:       rssID,
		Title:       "New Post",
		Url:         "https://example.com/new",
		PublishedAt: time.Now(),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
 
	post, err := r.GetPostByID(result.GeneratedID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if post.Title != "New Post" {
		t.Errorf("title: got %q want %q", post.Title, "New Post")
	}
}
 
func TestUpdatePostTitle(t *testing.T) {
	r, w := testDB(t)
	_, _, postID := seed(t, w)
 
	if err := w.UpdatePostTitle(postID, "Revised Title"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
 
	post, err := r.GetPostByID(postID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	if post.Title != "Revised Title" {
		t.Errorf("title: got %q want %q", post.Title, "Revised Title")
	}
}
 
func TestUpdatePostTitle_NotFound(t *testing.T) {
	_, w := testDB(t)
 
	err := w.UpdatePostTitle(9999, "Ghost")
	if err == nil {
		t.Fatal("expected error for missing post, got nil")
	}
}
 
func TestDeletePost(t *testing.T) {
	r, w := testDB(t)
	_, _, postID := seed(t, w)
 
	if err := w.DeletePost(postID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
 
	_, err := r.GetPostByID(postID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows after delete, got %v", err)
	}
}
 
func TestCreatePostsBatch(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	now := time.Now()
	payloads := []dto.PostPayload{
		{RssID: rssID, Title: "Batch A", Url: "https://a.example.com", PublishedAt: now},
		{RssID: rssID, Title: "Batch B", Url: "https://b.example.com", PublishedAt: now},
		{RssID: rssID, Title: "Batch C", Url: "https://c.example.com", PublishedAt: now},
	}
 
	results, err := w.CreatePostsBatch(payloads)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != len(payloads) {
		t.Fatalf("result count: got %d want %d", len(results), len(payloads))
	}
 
	// Every generated ID must be positive and unique
	seen := make(map[int64]bool)
	for i, res := range results {
		if res.GeneratedID <= 0 {
			t.Errorf("results[%d]: non-positive ID %d", i, res.GeneratedID)
		}
		if seen[res.GeneratedID] {
			t.Errorf("results[%d]: duplicate ID %d", i, res.GeneratedID)
		}
		seen[res.GeneratedID] = true
	}
 
	// Cross-check: posts should all appear under the rss
	posts, err := r.GetPostsByRss(rssID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	// seed inserted one post already; batch adds three more
	if len(posts) != 4 {
		t.Errorf("post count: got %d want 4", len(posts))
	}
}
 
func TestCreatePostsBatch_Empty(t *testing.T) {
	_, w := testDB(t)
 
	results, err := w.CreatePostsBatch(nil)
	if err != nil {
		t.Fatalf("empty batch should not error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for empty batch, got %v", results)
	}
}
 
func TestCreatePostsBatch_RollbackOnBadRow(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	// Mix valid and invalid (bad FK) payloads. The whole batch must roll back.
	payloads := []dto.PostPayload{
		{RssID: rssID, Title: "Good", Url: "https://good.example.com", PublishedAt: time.Now()},
		{RssID: 9999, Title: "Bad FK", Url: "https://bad.example.com", PublishedAt: time.Now()},
	}
 
	_, err := w.CreatePostsBatch(payloads)
	if err == nil {
		t.Fatal("expected error for FK violation, got nil")
	}
 
	// "Good" must not have been committed
	posts, err := r.GetPostsByRss(rssID)
	if err != nil {
		t.Fatalf("read back: %v", err)
	}
	// Only the original seed post should exist
	if len(posts) != 1 {
		t.Errorf("rollback failed: got %d posts, want 1", len(posts))
	}
}
 
// ---------------------------------------------------------------------------
// Reader — Topic
// ---------------------------------------------------------------------------
 
func TestGetTopics_Empty(t *testing.T) {
	r, _ := testDB(t)
 
	topics, err := r.GetTopics()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topics) != 0 {
		t.Errorf("expected empty slice, got %d topics", len(topics))
	}
}
 
func TestGetTopics_MultipleRows(t *testing.T) {
	r, w := testDB(t)
 
	for _, name := range []string{"AI", "Security", "Systems"} {
		if _, err := w.CreateTopic(dto.TopicPayload{Name: name}); err != nil {
			t.Fatalf("create topic %q: %v", name, err)
		}
	}
 
	topics, err := r.GetTopics()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topics) != 3 {
		t.Errorf("count: got %d want 3", len(topics))
	}
}
 
func TestGetTopicByID_NotFound(t *testing.T) {
	r, _ := testDB(t)
 
	_, err := r.GetTopicByID(9999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}
 
func TestGetTopicWithRss(t *testing.T) {
	r, w := testDB(t)
	topicID, _, _ := seed(t, w)
 
	topic, err := r.GetTopicWithRss(topicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(topic.Rss) != 1 {
		t.Errorf("rss count: got %d want 1", len(topic.Rss))
	}
	if topic.Rss[0].Title != "arXiv ML" {
		t.Errorf("rss title: got %q want %q", topic.Rss[0].Title, "arXiv ML")
	}
}
 
// ---------------------------------------------------------------------------
// Reader — RSS
// ---------------------------------------------------------------------------
 
func TestGetRssByTopic_EmptySliceNotNull(t *testing.T) {
	r, w := testDB(t)
	result, _ := w.CreateTopic(dto.TopicPayload{Name: "Empty"})
 
	items, err := r.GetRssByTopic(result.GeneratedID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Must be an empty slice (JSON → []) not nil (JSON → null)
	if items == nil {
		t.Error("expected empty non-nil slice, got nil")
	}
}
 
func TestGetRssURL(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	url, err := r.GetRssURL(rssID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://arxiv.org/rss/cs.LG" {
		t.Errorf("url: got %q want %q", url, "https://arxiv.org/rss/cs.LG")
	}
}
 
func TestGetRssURL_NotFound(t *testing.T) {
	r, _ := testDB(t)
 
	_, err := r.GetRssURL(9999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}
 
// ---------------------------------------------------------------------------
// Reader — Post
// ---------------------------------------------------------------------------
 
func TestGetPostsByRss_OrderedByPublishedAtDesc(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	for i, title := range []string{"Oldest", "Middle", "Newest"} {
		_, err := w.CreatePost(dto.PostPayload{
			RssID:       rssID,
			Title:       title,
			Url:         "https://example.com/" + title,
			PublishedAt: base.Add(time.Duration(i) * 24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("create post %q: %v", title, err)
		}
	}
 
	posts, err := r.GetPostsByRss(rssID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// seed + 3 new posts = 4 total; newest first
	if posts[0].Title != "Newest" {
		t.Errorf("first post: got %q want %q", posts[0].Title, "Newest")
	}
}
 
func TestGetPostsByTopic(t *testing.T) {
	r, w := testDB(t)
	topicID, rssID, _ := seed(t, w)
 
	// Add a second RSS under the same topic with two posts
	rr, _ := w.CreateRss(dto.RssPayload{TopicID: topicID, Title: "Second", Url: "https://b.example.com"})
	for _, title := range []string{"Post B1", "Post B2"} {
		_, err := w.CreatePost(dto.PostPayload{
			RssID: rr.GeneratedID, Title: title,
			Url: "https://b.example.com/" + title, PublishedAt: time.Now(),
		})
		if err != nil {
			t.Fatalf("create post: %v", err)
		}
	}
 
	_ = rssID // used implicitly via seed
	posts, err := r.GetPostsByTopic(topicID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 1 from seed + 2 from second rss
	if len(posts) != 3 {
		t.Errorf("count: got %d want 3", len(posts))
	}
}
 
func TestGetPostsByTopicInWindow(t *testing.T) {
	r, w := testDB(t)
	topicID, rssID, _ := seed(t, w)
 
	base := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	titles := []string{"Jan1", "Jan5", "Jan10", "Feb1"}
	dates := []time.Time{
		base,
		base.Add(4 * 24 * time.Hour),
		base.Add(9 * 24 * time.Hour),
		base.Add(31 * 24 * time.Hour),
	}
	for i, title := range titles {
		_, err := w.CreatePost(dto.PostPayload{
			RssID: rssID, Title: title,
			Url: "https://example.com/" + title, PublishedAt: dates[i],
		})
		if err != nil {
			t.Fatalf("create post %q: %v", title, err)
		}
	}
 
	// Window covers January only
	from := base
	to := time.Date(2024, 2, 1, 0, 0, 0, 0, time.UTC)
	posts, err := r.GetPostsByTopicInWindow(topicID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 3 {
		t.Errorf("window count: got %d want 3", len(posts))
	}
	// Must be ASC for coherent narrative order into the inference context window
	for i := 1; i < len(posts); i++ {
		if posts[i].PublishedAt.Before(posts[i-1].PublishedAt) {
			t.Errorf("posts[%d] (%v) is before posts[%d] (%v): want ASC order",
				i, posts[i].PublishedAt, i-1, posts[i-1].PublishedAt)
		}
	}
}
 
func TestGetPostsByRssInWindow(t *testing.T) {
	r, w := testDB(t)
	_, rssID, _ := seed(t, w)
 
	base := time.Date(2024, 3, 1, 0, 0, 0, 0, time.UTC)
	for i, title := range []string{"Week1", "Week2", "Week5"} {
		_, err := w.CreatePost(dto.PostPayload{
			RssID: rssID, Title: title,
			Url:         "https://example.com/" + title,
			PublishedAt: base.Add(time.Duration(i*7) * 24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("create post %q: %v", title, err)
		}
	}
 
	// Window covers first two weeks only
	from := base
	to := base.Add(14 * 24 * time.Hour)
	posts, err := r.GetPostsByRssInWindow(rssID, from, to)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("window count: got %d want 2", len(posts))
	}
}
 
func TestGetPostURL(t *testing.T) {
	r, w := testDB(t)
	_, _, postID := seed(t, w)
 
	url, err := r.GetPostURL(postID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url != "https://arxiv.org/abs/1706.03762" {
		t.Errorf("url: got %q want %q", url, "https://arxiv.org/abs/1706.03762")
	}
}
 
func TestGetPostURL_NotFound(t *testing.T) {
	r, _ := testDB(t)
 
	_, err := r.GetPostURL(9999)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("expected ErrNoRows, got %v", err)
	}
}
 
