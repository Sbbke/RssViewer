package service_test

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/model"
	"RssViewer/internal/service"
	"RssViewer/internal/storage"
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Define a minimal valid RSS XML mockup string matching gofeed's properties
const mockRssXML = `<?xml version="1.0" encoding="UTF-8" ?>
<rss version="2.0">
<channel>
    <title>Machine Learning Developments</title>
    <link>https://example.com/ml</link>
    <description>Latest updates in ML domain</description>
    <item>
        <title>Accelerating LLM Inference Efficiency</title>
        <link>https://example.com/ml/inference-efficiency</link>
        <description>Technical evaluation of matrix multiplication steps.</description>
        <pubDate>Mon, 04 Jun 2026 12:00:00 GMT</pubDate>
    </item>
</channel>
</rss>`

// FakeAccessor implements SQLAccessor explicitly for test isolation layers
type FakeAccessor struct {
	DB *sql.DB
}

func (f *FakeAccessor) GetDB() *sql.DB {
	return f.DB
}
func NewFakeDB(db *sql.DB) *FakeAccessor {
	return &FakeAccessor{
		DB: db,
	}
}

func testOrch(t *testing.T) *storage.DataOrch {
    t.Helper()

    db, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        t.Fatalf("open: %v", err)
    }
    t.Cleanup(func() { db.Close() })

    if err := model.ApplySchema(db); err != nil {
        t.Fatalf("apply schema: %v", err)
    }

    orch, err := storage.NewDataOrch(NewFakeDB(db), t.TempDir())
    if err != nil {
        t.Fatalf("new orch: %v", err)
    }
    return orch
}


// seed inserts one topic → one rss → one post and returns their IDs.
// Used by tests that need existing rows without caring about the insert itself.
func seed(t *testing.T, orch *storage.DataOrch) (topicID, rssID, postID int64) {
	t.Helper()

	// Seed via Orchestrator wrappers to ensure channel synchronization remains intact
	topicRes, err := orch.AddTopic(dto.TopicPayload{Name: "AI Development"})
	if err != nil {
		t.Fatalf("failed to seed baseline test topic: %v", err)
	}
	topicID = topicRes.GeneratedID

	rssRes, err := orch.AddRss(dto.RssPayload{
		Title:   "arXiv Machine Learning",
		Url:     "https://arxiv.org/rss/cs.LG",
	})
	if err != nil {
		t.Fatalf("failed to seed baseline test target resource: %v", err)
	}
	rssID = rssRes.GeneratedID

	err = orch.AddPostBatch([]dto.PostPayload{
		{
			RssID:       rssID,
			Title:       "Attention Is All You Need",
			Url:         "https://arxiv.org/abs/1706.03762",
			Content:     "Introduction of Transformer architecture parameters.",
			PublishedAt: time.Date(2017, 6, 12, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
		},
	})
	if err != nil {
		t.Fatalf("failed to seed baseline post allocation arrays: %v", err)
	}

	return topicID, rssID, 1
} // ---------------------------------------------------------------------------
// TestPost_ContentPersistedCorrectly verifies that post content is stored and
// retrieved without truncation or column-shift (guards against missing bind args).
func TestPost_ContentPersistedCorrectly(t *testing.T) {
    orch := testOrch(t)
    defer orch.Shutdown()

    db := orch.DB()

    _, rssID, _ := seed(t, orch)

    const (
        wantTitle   = "BERT: Pre-training of Deep Bidirectional Transformers"
        wantURL     = "https://arxiv.org/abs/1810.04805"
        wantContent = "Bidirectional encoder representations from transformers."
    )
    wantPublished := time.Date(2018, 10, 11, 0, 0, 0, 0, time.UTC)

    err := orch.AddPostBatch([]dto.PostPayload{
        {
            RssID:       rssID,
            Title:       wantTitle,
            Url:         wantURL,
            Content:     wantContent,
            PublishedAt: wantPublished.Format(time.RFC3339),
        },
    })
    if err != nil {
        t.Fatalf("AddPostBatch: %v", err)
    }

    var gotTitle, gotURL, gotContent, gotPublished string
    err = db.QueryRow(
        `SELECT title, url, content, published_at FROM post WHERE url = ?`, wantURL,
    ).Scan(&gotTitle, &gotURL, &gotContent, &gotPublished)
    if err != nil {
        t.Fatalf("query post: %v", err)
    }

    if gotTitle != wantTitle {
        t.Errorf("title: got %q, want %q", gotTitle, wantTitle)
    }
    if gotURL != wantURL {
        t.Errorf("url: got %q, want %q", gotURL, wantURL)
    }
    if gotContent != wantContent {
        // This is the field that was being silently dropped/shifted.
        t.Errorf("content: got %q, want %q", gotContent, wantContent)
    }
    parsedPublished, parseErr := time.Parse(time.RFC3339, gotPublished)
    if parseErr != nil {
        t.Fatalf("parse published_at %q: %v", gotPublished, parseErr)
    }
    if !parsedPublished.Equal(wantPublished) {
        t.Errorf("published_at: got %v, want %v", parsedPublished, wantPublished)
    }
}


func TestForeignKey_PostCascadesOnRssDelete(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB() // expose *sql.DB via a thin accessor on DataOrch

	_, rssID, _ := seed(t, orch)

	// Confirm the post exists before deletion.

	dr := orch.GetReader()
	posts, err := dr.GetPostsByRss(rssID)
	if err != nil{
		t.Fatalf("error retrieving posts:%v", err)
	}
	if len(posts) == 0 {
		t.Fatal("pre-condition failed: expected at least one post before rss deletion")
	}

	// Delete the parent rss row.
	if _, err := db.Exec(`DELETE FROM rss WHERE id = ?`, rssID); err != nil {
		t.Fatalf("delete rss: %v", err)
	}
		
	// Child posts must be gone.
	posts, err = dr.GetPostsByRss(rssID)
	if err != nil{
		t.Fatalf("error retrieving posts:%v", err)
	}
	if len(posts) != 0 {
		t.Fatal("deletion failed: expected at no post after rss deletion")
	}


}


func TestSubmitRssUrl_Success(t *testing.T) {
	// 1. Establish an isolated HTTP Test server to serve mock XML bytes local to the machine partition
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(mockRssXML))
	}))
	defer server.Close()

	// 2. Setup database infrastructure stubs.
	targetOrch := testOrch(t)
	defer targetOrch.Shutdown()


	// Create context boundary
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Instantiate service
	// targetOrch := storage.NewDataOrch(mockDB, t.TempDir(), 10) // Example real initialization

	// Pre-requisite initialization setup shortcut for pure isolation validation:
	s := service.NewRssService(targetOrch)
	// Because targetOrch calls are tightly coupled inside SubmitRssUrl, ensure your mock
	// implementation returns a populated mutation result with GeneratedID set to 42.
	rssRes, err := s.SubmitRssUrl(ctx, server.URL)
	if err != nil {
		t.Fatalf("submit rss url error : %v", err)
	}

	var expectedId int64 = 1

	if rssRes.ID != expectedId {
		t.Fatalf("error rss item, expected: %d, got: %d", expectedId, rssRes.ID)
	}
}
// TestForeignKey_RssTopicsCascadesOnRssDelete verifies that deleting an rss row
// removes its join-table rows (rss_topics.rss_id → rss.id ON DELETE CASCADE).
func TestForeignKey_RssTopicsCascadesOnRssDelete(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB()

	topicID, rssID, _ := seed(t, orch)

	// Link rss ↔ topic through the join table.
	if _, err := db.Exec(`INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`, rssID, topicID); err != nil {
		t.Fatalf("insert rss_topics: %v", err)
	}

	// Confirm the join row exists.
	var linkCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_topics WHERE rss_id = ?`, rssID).Scan(&linkCount); err != nil {
		t.Fatalf("count rss_topics before delete: %v", err)
	}
	if linkCount == 0 {
		t.Fatal("pre-condition failed: expected rss_topics row before rss deletion")
	}

	// Delete the parent rss row.
	if _, err := db.Exec(`DELETE FROM rss WHERE id = ?`, rssID); err != nil {
		t.Fatalf("delete rss: %v", err)
	}

	// Join rows must be gone.
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_topics WHERE rss_id = ?`, rssID).Scan(&linkCount); err != nil {
		t.Fatalf("count rss_topics after delete: %v", err)
	}
	if linkCount != 0 {
		t.Errorf("cascade delete failed: expected 0 rss_topics rows after rss deletion, got %d", linkCount)
	}
}

// TestForeignKey_RssTopicsCascadesOnTopicDelete verifies that deleting a topic
// removes its join-table rows (rss_topics.topic_id → topic.id ON DELETE CASCADE).
func TestForeignKey_RssTopicsCascadesOnTopicDelete(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB()

	topicID, rssID, _ := seed(t, orch)

	if _, err := db.Exec(`INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`, rssID, topicID); err != nil {
		t.Fatalf("insert rss_topics: %v", err)
	}

	// Delete the parent topic row.
	if _, err := db.Exec(`DELETE FROM topic WHERE id = ?`, topicID); err != nil {
		t.Fatalf("delete topic: %v", err)
	}

	// Join rows on the topic side must be gone.
	var linkCount int
	if err := db.QueryRow(`SELECT COUNT(*) FROM rss_topics WHERE topic_id = ?`, topicID).Scan(&linkCount); err != nil {
		t.Fatalf("count rss_topics after topic delete: %v", err)
	}
	if linkCount != 0 {
		t.Errorf("cascade delete failed: expected 0 rss_topics rows after topic deletion, got %d", linkCount)
	}
}


// TestForeignKey_PostRejectsOrphanRssID verifies that inserting a post that
// references a non-existent rss row is rejected by the foreign key constraint.
func TestForeignKey_PostRejectsOrphanRssID(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB()

	const nonExistentRssID = 99999
	now := time.Now().UTC().Format(time.RFC3339)

	_, err := db.Exec(
		`INSERT INTO post (source_id, title, url, content, created_at, published_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		nonExistentRssID, "Ghost Post", "https://example.com/ghost",
		"content", now, now,
	)
	if err == nil {
		t.Error("expected foreign key violation inserting post with non-existent rss id, got nil error")
	}
}

// TestForeignKey_RssTopicsRejectsOrphanRssID verifies that rss_topics rejects
// a row whose rss_id does not exist.
func TestForeignKey_RssTopicsRejectsOrphanRssID(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB()

	topicRes, err := orch.AddTopic(dto.TopicPayload{Name: "Orphan RSS Test Topic"})
	if err != nil {
		t.Fatalf("add topic: %v", err)
	}

	const nonExistentRssID = 99999
	_, err = db.Exec(
		`INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`,
		nonExistentRssID, topicRes.GeneratedID,
	)
	if err == nil {
		t.Error("expected foreign key violation inserting rss_topics with non-existent rss_id, got nil error")
	}
}

// TestForeignKey_RssTopicsRejectsOrphanTopicID verifies that rss_topics rejects
// a row whose topic_id does not exist.
func TestForeignKey_RssTopicsRejectsOrphanTopicID(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB()

	rssRes, err := orch.AddRss(dto.RssPayload{
		Title: "Orphan Topic Test Feed",
		Url:   "https://example.com/orphan-topic-test",
	})
	if err != nil {
		t.Fatalf("add rss: %v", err)
	}

	const nonExistentTopicID = 99999
	_, err = db.Exec(
		`INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`,
		rssRes.GeneratedID, nonExistentTopicID,
	)
	if err == nil {
		t.Error("expected foreign key violation inserting rss_topics with non-existent topic_id, got nil error")
	}
}

// TestForeignKey_RssTopicsUniqueness verifies that the composite primary key
// on rss_topics prevents duplicate (rss_id, topic_id) pairs.
func TestForeignKey_RssTopicsUniqueness(t *testing.T) {
	orch := testOrch(t)
	defer orch.Shutdown()

	db := orch.DB()

	topicID, rssID, _ := seed(t, orch)

	if _, err := db.Exec(`INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`, rssID, topicID); err != nil {
		t.Fatalf("first insert into rss_topics: %v", err)
	}

	// Second insert of the same pair must fail.
	_, err := db.Exec(`INSERT INTO rss_topics (rss_id, topic_id) VALUES (?, ?)`, rssID, topicID)
	if err == nil {
		t.Error("expected uniqueness violation on duplicate rss_topics entry, got nil error")
	}
}
