package storage

import (
	"RssViewer/internal/dto"
	"database/sql"
)

// ---------------------------------------------------------------------------
// Interfaces
// ---------------------------------------------------------------------------

// SQLAccessor is satisfied by the *sqliteDb returned from InitDB.
// The orchestrator only needs the live *sql.DB handle, keeping it decoupled
// from the concrete wrapper type.
type SQLAccessor interface {
	GetDB() *sql.DB
}

// ---------------------------------------------------------------------------
// Task types
// ---------------------------------------------------------------------------

// TaskType is a typed constant so the compiler catches typos at the call site.
type TaskType string

const (
	TaskCreateTopic        TaskType = "CREATE_TOPIC"
	TaskCreateRss          TaskType = "CREATE_RSS"        // dual: DB + local dir scaffold
	TaskCreatePost         TaskType = "CREATE_POST"       // DB only — no summary/slide yet
	TaskCreatePostBatch    TaskType = "CREATE_POST_BATCH" // DB only batch insert
	TaskUpdatePostTitle    TaskType = "UPDATE_POST_TITLE"
	TaskCreatePostSummary  TaskType = "CREATE_POST_SUMMARY" // primary write: local txt
	TaskCreatePostSlide    TaskType = "CREATE_POST_SLIDE"   // primary write: local png(s)
	TaskCreateTopicSummary TaskType = "CREATE_TOPIC_SUMMARY"
	TaskCreateTopicSlide   TaskType = "CREATE_TOPIC_SLIDE"
)

// Delete task types
const (
	TaskDeleteRss   TaskType = "delete_rss"
	TaskDeleteTopic TaskType = "delete_topic"
	TaskDeletePost  TaskType = "delete_post"
)

type WriteTask struct {
	Type       TaskType
	Payload    any // concrete dto.* type; each case asserts the expected type
	ResultChan chan dto.MutationResult
	ErrChan    chan error
}
type UpdatePostTitlePayload struct {
	PostID int64
	Title  string
}
