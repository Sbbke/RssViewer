
# DBWriter

`DBWriter` owns every database mutation. It is the only component allowed to modify SQLite state.

The writer accepts DTO payloads from the Service layer, converts them into internal database models, performs validation, executes SQL mutations, and returns mutation results. The database schema is intentionally hidden from callers.

---

## Topic Operations

### CreateTopic

Creates a new Topic.

**Signature**

```go
CreateTopic(payload dto.TopicPayload)
    (dto.MutationResult, error)
```

**Input**

| Parameter | Type | Description |
|-----------|------|-------------|
| payload | TopicPayload | Name of the new Topic. |

**Returns**

| Type | Description |
|------|-------------|
| MutationResult | Contains the generated Topic ID. |
| error | Returned when insertion fails. |

**Behavior**

- Generates `created_at`
- Inserts a Topic row
- Returns the generated primary key

---

### UpdateTopic

Updates an existing Topic.

**Signature**

```go
UpdateTopic(
    id int64,
    payload dto.TopicPayload,
) error
```

**Input**

| Parameter | Type | Description |
|-----------|------|-------------|
| id | int64 | Target Topic ID |
| payload | TopicPayload | Updated Topic information |

**Returns**

`error`

**Behavior**

- Updates Topic metadata
- Verifies exactly one row was affected

---

### DeleteTopic

Deletes a Topic.

```go
DeleteTopic(id int64) error
```

Deletes the Topic record.

Associated relationships should be handled by foreign-key cascade or service orchestration.

---

# RSS Operations

## CreateRss

Creates a new RSS subscription.

```go
CreateRss(
    payload dto.RssPayload,
) (dto.MutationResult, error)
```

### Input

| Field | Description |
|------|-------------|
| Title | Feed title |
| Url | RSS URL |
| Xml | Original RSS XML |

### Output

Returns

```go
MutationResult
```

containing the generated RSS ID.

### Behavior

- Stores original XML
- Stores RSS metadata
- Generates `created_at`

---

## UpdateRss

```go
UpdateRss(
    id int64,
    payload dto.RssPayload,
) error
```

Updates RSS metadata and replaces the stored XML.

---

## DeleteRss

```go
DeleteRss(id int64) error
```

Deletes the RSS subscription.

---

# Post Operations

## CreatePost

Creates one Post.

```go
CreatePost(
    payload dto.PostPayload,
) (dto.MutationResult, error)
```

### Input

| Field | Description |
|------|-------------|
| RssID | Parent RSS |
| Title | Post title |
| Url | Original URL |
| Content | Processed plain text |
| PublishedAt | Publication timestamp |

### Output

Returns

```go
MutationResult
```

containing the generated Post ID.

### Behavior

- Generates `created_at`
- Persists processed content
- Creates one Post record

---

## CreatePostsBatch

Creates multiple Posts inside a single SQL transaction.

```go
CreatePostsBatch(
    payloads []dto.PostPayload,
) ([]dto.MutationResult, error)
```

### Input

A slice of `PostPayload`.

### Output

Returns one `MutationResult` for each inserted Post.

### Behavior

- Opens one SQL transaction
- Inserts every Post
- Rolls back on any failure
- Guarantees atomicity
- Preserves input ordering in returned results

---

## UpdatePostTitle

```go
UpdatePostTitle(
    id int64,
    title string,
) error
```

Updates only the title of an existing Post.

---

## DeletePost

```go
DeletePost(id int64) error
```

Deletes a Post.

---

# Summary Operations

Topic and Post summaries share the same DTO because both represent generated briefing text.

---

## CreatePostSummary

```go
CreatePostSummary(
    payload dto.SummaryPayload,
) (dto.MutationResult, error)
```

### Input

```go
SummaryPayload
```

| Field | Description |
|------|-------------|
| ID | Target Post ID |
| Content | Generated summary |

### Output

Generated Summary ID.

### Behavior

Creates a new Post Summary.

---

## UpdatePostSummary

```go
UpdatePostSummary(
    payload dto.SummaryPayload,
) error
```

Updates the generated summary for a Post.

---

## DeletePostSummary

```go
DeletePostSummary(
    postID int64,
) error
```

Deletes a Post Summary.

---

## CreateTopicSummary

```go
CreateTopicSummary(
    payload dto.SummaryPayload,
) (dto.MutationResult, error)
```

Creates a generated Topic summary.

Input `ID` refers to the Topic ID.

---

## UpdateTopicSummary

```go
UpdateTopicSummary(
    payload dto.SummaryPayload,
) error
```

Updates an existing Topic Summary.

---

## DeleteTopicSummary

```go
DeleteTopicSummary(
    topicID int64,
) error
```

Deletes a Topic Summary.

---

# Relationship Operations

## LinkRssTopic

Associates an RSS feed with a Topic.

```go
LinkRssTopic(
    rssID int64,
    topicID int64,
) error
```

### Behavior

Creates one record inside the `rss_topics` junction table.

---

## UnlinkRssTopic

Removes an RSS ↔ Topic association.

```go
UnlinkRssTopic(
    rssID int64,
    topicID int64,
) error
```

Deletes the corresponding row from `rss_topics`.

---

# Error Handling

Every mutation follows the same conventions.

- Returns `error` when SQL execution fails.
- Returns `error` when zero rows are affected during an update or delete.
- Create operations return a `MutationResult` containing the generated primary key.
- Batch operations are fully transactional; partial writes are never committed.

---

# Internal DTO Mapping

Before any SQL execution, every public DTO is converted into an internal database model.

```
Service
    │
    ▼
DTO Payload
    │
    ▼
DBWriter
    │
    ├── DTO → Model
    ├── Generate timestamps
    ├── SQL execution
    └── MutationResult
```

This separation ensures that database schema changes remain isolated within the writer implementation and never propagate to the Service layer.

