package storage

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage/reader"
	"RssViewer/internal/storage/writer"
	"fmt"
)

type DataOrch struct {
	dbLayer     SQLAccessor
	dbWriter    *writer.DBWriter
	localWriter *writer.LocalWriter
	dbReader    *reader.DBReader
	localReader *reader.LocalReader
	taskCh      chan WriteTask
	done        chan struct{}
}

func NewDataOrch(db SQLAccessor, baseDiskPath string) (*DataOrch, error) {
	raw := db.GetDB()

	dbw, err := writer.NewDBWriter(raw)
	if err != nil {
		return nil, fmt.Errorf("NewDataOrch: db writer: %w", err)
	}
	lw := writer.NewLocalWriter(baseDiskPath)

	dbr, err := reader.NewDBReader(raw)
	if err != nil {
		return nil, fmt.Errorf("NewDataOrch: db reader: %w", err)
	}
	lr := reader.NewLocalReader(baseDiskPath)

	do := &DataOrch{
		dbLayer:     db,
		dbWriter:    dbw,
		localWriter: lw,
		dbReader:    dbr,
		localReader: lr,
		// Buffer gives callers headroom to enqueue without blocking when the
		// worker is mid-write. 64 is a reasonable starting value for a
		// single-user desktop app; tune upward if CheckUpdate floods the queue.
		taskCh: make(chan WriteTask, 64),
		done:   make(chan struct{}),
	}

	go do.runWorker()
	return do, nil
}

func (do *DataOrch) Shutdown() {
	close(do.taskCh)
	<-do.done
}

// GetReader returns the SQL reader for direct use by services.
// Reads bypass the channel — SQLite handles concurrent reads natively.
func (do *DataOrch) GetReader() *reader.DBReader {
	return do.dbReader
}

// GetLocalReader returns the local-file reader for direct use by services.
func (do *DataOrch) GetLocalReader() *reader.LocalReader {
	return do.localReader
}

// SubmitWrite enqueues a task onto the single worker goroutine.
// It is safe to call from any goroutine. The caller owns ErrChan and must
// receive from it to avoid the worker blocking on a full channel.
func (do *DataOrch) SubmitWrite(task WriteTask) {
	do.taskCh <- task
}

func (do *DataOrch) RequestRead() SQLAccessor {
	return do.dbLayer
}

// ---------------------------------------------------------------------------
// Worker loop — single goroutine, serializes all writes
// ---------------------------------------------------------------------------

func (do *DataOrch) runWorker() {
	defer close(do.done)
	for task := range do.taskCh {
		err := do.executeInternalMutation(task)
		if task.ErrChan != nil {
			task.ErrChan <- err
		}
	}
}

// ---------------------------------------------------------------------------
// Mutation dispatch
// ---------------------------------------------------------------------------

func (do *DataOrch) executeInternalMutation(task WriteTask) error {
	switch task.Type {

	// -----------------------------------------------------------------------
	case TaskCreateTopic:
		p, ok := task.Payload.(dto.TopicPayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		_, err := do.dbWriter.CreateTopic(p)
		return err

	// -----------------------------------------------------------------------
	// Dual-target sequence:
	//   1. Insert DB row to get the generated RSS ID.
	//   2. Scaffold local directory so the crawler has a home for rss.xml.
	case TaskCreateRss:
		p, ok := task.Payload.(dto.RssPayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		result, err := do.dbWriter.CreateRss(p)
		if err != nil {
			return err
		}
		if err := do.localWriter.CreateRss(result.GeneratedID, p.Content); err != nil {
			return fmt.Errorf("%s: local write rss_id=%d: %w", task.Type, result.GeneratedID, err)
		}
		return nil

	// -----------------------------------------------------------------------
	// DB-only: the post pointer is stored; content lives on disk and is
	// fetched lazily by the HTML processor on first GetContent call.
	case TaskCreatePost:
		p, ok := task.Payload.(dto.PostPayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		_, err := do.dbWriter.CreatePost(p)
		return err

	// -----------------------------------------------------------------------
	case TaskCreatePostBatch:
		p, ok := task.Payload.([]dto.PostPayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		_, err := do.dbWriter.CreatePostsBatch(p)
		return err

	// -----------------------------------------------------------------------
	case TaskUpdatePostTitle:
		p, ok := task.Payload.(UpdatePostTitlePayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		return do.dbWriter.UpdatePostTitle(p.PostID, p.Title)

	// -----------------------------------------------------------------------
	// Primary write is local (large text blob); no DB row needed for content.
	// The summary service has already generated the text and passes it here
	// as a dto.PostSummaryPayload (body string + post ID for the path).
	case TaskCreatePostSummary:
		p, ok := task.Payload.(dto.PostSummaryPayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		return do.localWriter.CreatePostSummary(p.PostID, p.Body)

	// -----------------------------------------------------------------------
	// Primary write is local (PNG files); paths are written back to DB so the
	// UI can reference them without re-reading the filesystem on every render.
	// Sequence:
	//   1. Write PNG files to disk via localWriter.
	case TaskCreatePostSlide:
		p, ok := task.Payload.(dto.PostSlidePayload)
		if !ok {
			return fmt.Errorf("%s: unexpected payload type %T", task.Type, task.Payload)
		}
		return do.localWriter.CreatePostSlide(p.PostID, p.Slide)
		// -----------------------------------------------------------------------

	case TaskCreateTopicSummary:
		p, ok := task.Payload.(dto.TopicSummaryPayload)
		if !ok {
			return unexpectedPayload(task)
		}
		return do.localWriter.CreatePostSummary(p.TopicID, p.Body)

	case TaskCreateTopicSlide:
		p, ok := task.Payload.(dto.TopicSlidePayload)
		if !ok {
			return unexpectedPayload(task)
		}
		return do.localWriter.CreatePostSlide(p.TopicID, p.Slide)

	default:
		return fmt.Errorf("executeInternalMutation: unrecognized task type %q", task.Type)
	}
}

func unexpectedPayload(t WriteTask) error {
	return fmt.Errorf("%s: unexpected payload type %T", t.Type, t.Payload)

}
