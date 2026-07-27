# Orchestrator

The Orchestrator (`DataOrch`) is the central coordinator of the data layer.

It manages:

- SQL read access
- SQL write serialization
- Local asset storage access
- Mutation task dispatching
- Goroutine lifecycle

The Orchestrator provides a unified interface for the Service layer while hiding the concrete storage implementation.

---

# Responsibilities

## Read Management

The Orchestrator provides direct access to read-only components.

SQLite supports concurrent reads, therefore read operations bypass the task queue.

```text
Service
   |
   ▼
DataOrch
   |
   ├── DBReader
   └── LocalReader
```

### GetReader()

```go
GetReader() *meta.DBReader
```

Returns the SQL reader.

The returned reader can execute concurrent read operations safely.

---

### GetLocalReader()

```go
GetLocalReader() *images.LocalReader
```

Returns the local filesystem reader.

Used for retrieving generated assets such as:

* Slides
* Local cached files

---

# Write Management

SQLite only supports one writer at a time.

The Orchestrator serializes all mutations through a single worker goroutine.

```
Multiple Services
        |
        ▼
 SubmitWrite()
        |
        ▼
  WriteTask Channel
        |
        ▼
 Single Worker Goroutine
        |
        ▼
 executeInternalMutation()
        |
        ├── DBWriter
        └── LocalWriter
```

This prevents:

```
SQLITE_BUSY
database is locked
```

errors caused by concurrent writes.

---

# Internal Components

```go
type DataOrch struct {
    dbAccessor   SQLAccessor
    dbWriter     *DBWriter
    dbReader     *DBReader

    localWriter  *LocalWriter
    localReader  *LocalReader

    taskCh       chan WriteTask
    done         chan struct{}
}
```

---

# Lifecycle

## NewDataOrch()

```go
NewDataOrch(
    db SQLAccessor,
    baseDiskPath string,
) (*DataOrch, error)
```

Initializes:

* SQL Reader
* SQL Writer
* Local Reader
* Local Writer
* Write task queue
* Worker goroutine

After initialization:

```
DataOrch
   |
   ├── DBReader
   ├── DBWriter
   ├── LocalReader
   └── LocalWriter
```

---

## Shutdown()

```go
Shutdown()
```

Stops the worker goroutine.

Workflow:

1. Close task channel.
2. Wait until worker finishes remaining tasks.
3. Release resources.

---

# WriteTask Workflow

Every mutation is converted into a `WriteTask`.

A task contains:

* Operation type
* Payload
* Error channel
* Optional result channel

Example:

```
AddPost()
    |
    ▼
WriteTask{
    Type: TaskCreatePost,
    Payload: PostPayload
}
    |
    ▼
taskCh
    |
    ▼
runWorker()
```

---

# SubmitWrite()

```go
SubmitWrite(
    task WriteTask,
)
```

Adds a mutation task to the single writer queue.

Properties:

* Thread-safe.
* Can be called from any goroutine.
* Does not execute mutations directly.
* Caller waits through response channels.

---

# Public Mutation APIs

The Orchestrator exposes high-level mutation functions.

Each function:

1. Creates a `WriteTask`.
2. Submits it.
3. Waits for worker completion.
4. Returns result/error.

---

# Topic Operations

## AddTopic

```go
AddTopic(
    p dto.TopicPayload,
) (dto.MutationResult, error)
```

Creates a Topic.

Delegates to:

```
DBWriter.CreateTopic()
```

---

## DeleteTopic

```go
DeleteTopic(
    id int64,
) error
```

Deletes a Topic.

Delegates to:

```
DBWriter.DeleteTopic()
```

---

# RSS Operations

## AddRss

```go
AddRss(
    p dto.RssPayload,
) (dto.MutationResult, error)
```

Creates an RSS feed.

Delegates to:

```
DBWriter.CreateRss()
```

---

## UpdateRss

```go
UpdateRss(
    p dto.RssUpdatePayload,
) error
```

Updates RSS metadata and XML.

Delegates to:

```
DBWriter.UpdateRss()
```

---

## DeleteRss

```go
DeleteRss(
    id int64,
) error
```

Deletes an RSS feed.

---

# Post Operations

## AddPost

```go
AddPost(
    p dto.PostPayload,
) (dto.MutationResult, error)
```

Creates a single Post.

---

## AddPostBatch

```go
AddPostBatch(
    posts []dto.PostPayload,
) error
```

Creates multiple Posts atomically.

The batch operation is executed inside:

```
DBWriter.CreatePostsBatch()
```

---

## DeletePost

```go
DeletePost(
    id int64,
) error
```

Deletes a Post.

---

# Summary Operations

## AddPostSummary

```go
AddPostSummary(
    p dto.SummaryPayload,
) error
```

Creates a Post summary.

---

## UpdatePostSummary

```go
UpdatePostSummary(
    p dto.SummaryPayload,
) error
```

Updates a Post summary.

---

## DeletePostSummary

```go
DeletePostSummary(
    id int64,
) error
```

Deletes a Post summary.

---

## AddTopicSummary

```go
AddTopicSummary(
    p dto.SummaryPayload,
) error
```

Creates a Topic summary.

---

## UpdateTopicSummary

```go
UpdateTopicSummary(
    p dto.SummaryPayload,
) error
```

Updates a Topic summary.

---

## DeleteTopicSummary

```go
DeleteTopicSummary(
    id int64,
) error
```

Deletes a Topic summary.

---

# Slide Operations

Slides are stored through `LocalWriter`.

They are not persisted as SQL blobs.

---

## Post Slides

### AddPostSlide

```go
AddPostSlide(
    p dto.PostSlidePayload,
) error
```

Creates Post slide assets.

---

### UpdatePostSlide

```go
UpdatePostSlide(
    p dto.PostSlidePayload,
) error
```

Replaces Post slide assets.

---

### DeletePostSlide

```go
DeletePostSlide(
    id int64,
) error
```

Deletes Post slide assets.

---

## Topic Slides

### AddTopicSlide

```go
AddTopicSlide(
    p dto.TopicSlidePayload,
) error
```

Creates Topic slide assets.

---

### UpdateTopicSlide

```go
UpdateTopicSlide(
    p dto.TopicSlidePayload,
) error
```

Updates Topic slide assets.

---

### DeleteTopicSlide

```go
DeleteTopicSlide(
    p dto.TopicSlidePayload,
) error
```

Deletes Topic slide assets.

---

# RSS Topic Relationship

## LinkRssTopic

```go
LinkRssTopic(
    rssID int64,
    topicID int64,
) error
```

Creates RSS ↔ Topic relationship.

Delegates to:

```
DBWriter.LinkRssTopic()
```

---

## UnlinkRssTopic

```go
UnlinkRssTopic(
    rssID int64,
    topicID int64,
) error
```

Removes RSS ↔ Topic relationship.

---

# Worker Execution

The worker loop is responsible for executing every mutation.

```go
runWorker()
```

Workflow:

```
taskCh
   |
   ▼
runWorker()
   |
   ▼
executeInternalMutation()
   |
   ├── DBWriter
   └── LocalWriter
```

---

# Mutation Dispatch

`executeInternalMutation()` maps task types to concrete storage operations.

Example:

```
TaskCreatePost
        |
        ▼
DBWriter.CreatePost()

TaskCreateTopicSlide
        |
        ▼
LocalWriter.CreateTopicSlide()
```

The Service layer never directly accesses writers.

---

# Design Rules

## Services should:

* Use DBReader for reads.
* Use DataOrch for writes.
* Never access DBWriter directly.
* Never access LocalWriter directly.

## Orchestrator should:

* Coordinate storage operations.
* Serialize writes.
* Hide storage implementation details.

## Writers should:

* Perform actual persistence.
* Handle schema/filesystem details.
