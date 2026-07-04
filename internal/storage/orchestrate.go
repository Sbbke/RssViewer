package storage

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage/images/local"
	"RssViewer/internal/storage/meta/sqlite"
	"database/sql"
	"fmt"
)

type DataOrch struct {
	dbAcessor   SQLAccessor
	dbWriter    *meta.DBWriter
	dbReader    *meta.DBReader
	localWriter *images.LocalWriter
	localReader *images.LocalReader
	taskCh      chan WriteTask
	done        chan struct{}
}

func NewDataOrch(db SQLAccessor, baseDiskPath string) (*DataOrch, error) {
	raw := db.GetDB()

	dbw, err := meta.NewDBWriter(raw)
	if err != nil {
		return nil, fmt.Errorf("NewDataOrch: db writer: %w", err)
	}
	lw := images.NewLocalWriter(baseDiskPath)

	dbr, err := meta.NewDBReader(raw)
	if err != nil {
		return nil, fmt.Errorf("NewDataOrch: db reader: %w", err)
	}
	lr := images.NewLocalReader(baseDiskPath)

	do := &DataOrch{
		dbAcessor:   db,
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
func (do *DataOrch) GetReader() *meta.DBReader {
	return do.dbReader
}

// GetLocalReader returns the local-file reader for direct use by services.
func (do *DataOrch) GetLocalReader() *images.LocalReader {
	return do.localReader
}

// SubmitWrite enqueues  a task onto the single worker goroutine.
// It is safe to call from any goroutine. The caller owns ErrChan and must
// receive from it to avoid the worker blocking on a full channel.
func (do *DataOrch) SubmitWrite(task WriteTask) {
	do.taskCh <- task
}

func (do *DataOrch) DB() *sql.DB {
	return do.dbAcessor.GetDB()
}

// ---------------------------------------------------------------------------
// Worker loop — single goroutine, serializes all writes
// ---------------------------------------------------------------------------
func (do *DataOrch) AddTopic(p dto.TopicPayload) (dto.MutationResult, error) {

	task := WriteTask{
		Type:       TaskCreateTopic,
		Payload:    p,
		ErrChan:    make(chan error, 1),
		ResultChan: make(chan dto.MutationResult, 1),
	}

	// Dispatch to the serialized loop channel
	do.SubmitWrite(task)

	// Block until the single-writer loop processes it and responds
	select {
	case err := <-task.ErrChan:
		return dto.MutationResult{}, err
	case result := <-task.ResultChan:
		return result, nil
	}
}

func (do *DataOrch) AddPost(p dto.PostPayload) (dto.MutationResult, error) {

	task := WriteTask{
		Type:       TaskCreatePost,
		Payload:    p,
		ErrChan:    make(chan error, 1),
		ResultChan: make(chan dto.MutationResult, 1),
	}

	// Dispatch to the serialized loop channel
	do.SubmitWrite(task)

	// Block until the single-writer loop processes it and responds
	select {
	case err := <-task.ErrChan:
		return dto.MutationResult{}, err
	case result := <-task.ResultChan:
		return result, nil
	}
}

// AddRss registers an RSS feed source and returns the generated DB record metadata
func (do *DataOrch) AddRss(p dto.RssPayload) (dto.MutationResult, error) {
	task := WriteTask{
		Type:       TaskCreateRss,
		Payload:    p,
		ErrChan:    make(chan error, 1),
		ResultChan: make(chan dto.MutationResult, 1),
	}
	do.SubmitWrite(task)

	select {
	case err := <-task.ErrChan:
		return dto.MutationResult{}, err
	case result := <-task.ResultChan:
		return result, nil
	}
}

func (do *DataOrch) UpdateRss( p dto.RssUpdatePayload) error{
	task := WriteTask{
		Type: TaskUpdateRss,
		Payload: p,
		ErrChan: make(chan error, 1),
	}
	do.SubmitWrite(task)
	return <- task.ErrChan
}
// AddPostBatch inserts parsed feed elements inside a single atomic database transaction
func (do *DataOrch) AddPostBatch(posts []dto.PostPayload) error {
	task := WriteTask{
		Type:    TaskCreatePostBatch,
		Payload: posts,
		ErrChan: make(chan error, 1),
	}
	do.SubmitWrite(task)

	return <-task.ErrChan

}

func (do *DataOrch) DeleteRss(id int64) error {
	task := WriteTask{
		Type:    TaskDeleteRss,
		Payload: id,
		ErrChan: make(chan error, 1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}

func (do *DataOrch) DeleteTopic(id int64) error {
	task := WriteTask{
		Type:    TaskDeleteTopic,
		Payload: id,
		ErrChan: make(chan error, 1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}

func (do *DataOrch) DeletePost(id int64) error {
	task := WriteTask{
		Type:    TaskDeletePost,
		Payload: id,
		ErrChan: make(chan error, 1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}


// summary

func (do *DataOrch) AddPostSlide(p dto.PostSlidePayload) error{

	task := WriteTask{
		Type: TaskCreatePostSlide,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}


func (do *DataOrch) AddTopicSlide(p dto.TopicSlidePayload) error{

	task := WriteTask{
		Type: TaskCreateTopicSlide,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}


func (do *DataOrch) UpdateTopicSlide(p dto.TopicSlidePayload) error{

	task := WriteTask{
		Type: TaskUpdateTopicSlide,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}



func (do *DataOrch) UpdatePostSlide(p dto.TopicSlidePayload) error{

	task := WriteTask{
		Type: TaskUpdatePostSlide,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}


func (do *DataOrch) DeleteTopicSlide(p dto.TopicSlidePayload) error{
	task := WriteTask{
		Type: TaskDeleteTopicSlide,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <- task.ErrChan
}


func (do *DataOrch) DeletePostSlide(id int64) error{
	task := WriteTask{
		Type: TaskDeletePostSlide,
		Payload: id,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <- task.ErrChan
}


func (do *DataOrch) AddPostSummary(p dto.SummaryPayload) error{

	task := WriteTask{
		Type: TaskCreatePostSummary,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}

func (do *DataOrch) AddTopicSummary(p dto.SummaryPayload) error{

	task := WriteTask{
		Type: TaskCreateTopicSummary,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}

func (do *DataOrch) UpdatePostSummary(p dto.SummaryPayload) error{

	task := WriteTask{
		Type: TaskUpdatePostSummary,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}

func (do *DataOrch) UpdateTopicSummary(p dto.SummaryPayload) error{

	task := WriteTask{
		Type: TaskUpdateTopicSummary,
		Payload: p,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <-task.ErrChan
}

func (do *DataOrch) DeleteTopicSummary(id int64) error{
	task := WriteTask{
		Type: TaskDeleteTopicSummary,
		Payload: id,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <- task.ErrChan
}

func (do *DataOrch) DeletePostSummary(id int64) error{
	task := WriteTask{
		Type: TaskDeletePostSummary,
		Payload: id,
		ErrChan: make(chan error,1),
	}
	do.SubmitWrite(task)
	return <- task.ErrChan
}

func (do *DataOrch) runWorker() {
	defer close(do.done)
	for task := range do.taskCh {
		result, err := do.executeInternalMutation(task)
		if err != nil {
			task.ErrChan <- err
		} else if task.ResultChan != nil {
			task.ResultChan <- result
		} else {
			task.ErrChan <- nil // unblock callers like AddPostBatch that only listen on ErrChan
		}
	}
}

// ---------------------------------------------------------------------------
// Mutation dispatch
// ---------------------------------------------------------------------------

func (do *DataOrch) executeInternalMutation(task WriteTask) (dto.MutationResult, error) {
	switch task.Type {

	// -----------------------------------------------------------------------
	case TaskCreateTopic:
		p, ok := task.Payload.(dto.TopicPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)

		}
		return do.dbWriter.CreateTopic(p)

	// -----------------------------------------------------------------------
	// Dual-target sequence:
	//   1. Insert DB row to get the generated RSS ID.
	//   2. Scaffold local directory so the crawler has a home for rss.xml.
	case TaskCreateRss:
		p, ok := task.Payload.(dto.RssPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return do.dbWriter.CreateRss(p)

	// -----------------------------------------------------------------------
	// DB-only: the post pointer is stored; content lives on disk and is
	// fetched lazily by the HTML processor on first GetContent call.
	case TaskCreatePost:
		p, ok := task.Payload.(dto.PostPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return do.dbWriter.CreatePost(p)

	// -----------------------------------------------------------------------
	case TaskCreatePostBatch:
		p, ok := task.Payload.([]dto.PostPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		_, err := do.dbWriter.CreatePostsBatch(p)
		if err != nil {
			return dto.MutationResult{}, err
		}
		return dto.MutationResult{}, nil

	// -----------------------------------------------------------------------
	case TaskUpdatePostTitle:
		p, ok := task.Payload.(UpdatePostTitlePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.UpdatePostTitle(p.PostID, p.Title)

	// -----------------------------------------------------------------------
	// Primary write is local (large text blob); no DB row needed for content.
	// The summary service has already generated the text and passes it here
	// as a dto.PostSummaryPayload (body string + post ID for the path).
	case TaskCreatePostSummary:
		p, ok := task.Payload.(dto.SummaryPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		r, err := do.dbWriter.CreatePostSummary(p)
		if err != nil {
			return dto.MutationResult{}, err
		}
		
		return r, nil

	case TaskUpdatePostSummary:
		p, ok := task.Payload.(dto.SummaryPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
	
		return dto.MutationResult{}, do.dbWriter.UpdatePostSummary(p) 


	case TaskDeletePostSummary:
		id, ok := task.Payload.(int64)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.DeletePostSummary(id)

	// -----------------------------------------------------------------------
	// Primary write is local (PNG files); paths are written back to DB so the
	// UI can reference them without re-reading the filesystem on every render.
	// Sequence:
	//   1. Write PNG files to disk via localWriter.
	case TaskCreatePostSlide:
		p, ok := task.Payload.(dto.PostSlidePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.localWriter.CreatePostSlide(p.PostID, p.Slide)
		// -----------------------------------------------------------------------
	case TaskUpdatePostSlide:
		p, ok := task.Payload.(dto.PostSlidePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.localWriter.UpdatePostSlide(p.PostID, p.Slide)


	case TaskDeletePostSlide:
		id, ok := task.Payload.(int64)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.localWriter.DeletePostSlide(id)



	case TaskCreateTopicSummary:
		p, ok := task.Payload.(dto.SummaryPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		mutationResult , err := do.dbWriter.CreateTopicSummary(p)
		if err != nil{
			return dto.MutationResult{}, err
		}
		return mutationResult, nil 

	case TaskUpdateTopicSummary:
		p, ok := task.Payload.(dto.SummaryPayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}

		return dto.MutationResult{}, do.dbWriter.UpdateTopicSummary(p) 
	
	case TaskDeleteTopicSummary:
		id, ok := task.Payload.(int64)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.DeleteTopicSummary(id)
	

	case TaskCreateTopicSlide:
		p, ok := task.Payload.(dto.TopicSlidePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.localWriter.CreateTopicSlide(p.TopicID,p.RssHash, p.Slide)

	case TaskUpdateTopicSlide:
		p, ok := task.Payload.(dto.TopicSlidePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.localWriter.UpdateTopicSlide(p.TopicID, p.RssHash, p.Slide)



	case TaskDeleteTopicSlide:
		p, ok := task.Payload.(dto.TopicSlidePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.localWriter.DeleteTopicSlide(p.TopicID, p.RssHash)


	case TaskDeleteRss:
		id, ok := task.Payload.(int64)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.DeleteRss(id)

	case TaskDeleteTopic:
		id, ok := task.Payload.(int64)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.DeleteTopic(id)

	case TaskDeletePost:
		id, ok := task.Payload.(int64)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.DeletePost(id)
	case TaskUpdateRss:
		p, ok := task.Payload.(dto.RssUpdatePayload)
		if !ok {
			return dto.MutationResult{}, unexpectedPayload(task)
		}
		return dto.MutationResult{}, do.dbWriter.UpdateRss(p.Id,p.Body)

	default:
		return dto.MutationResult{}, unexpectedPayload(task)

	}
}

func unexpectedPayload(t WriteTask) error {
	return fmt.Errorf("%s: unexpected payload type %T", t.Type, t.Payload)

}
