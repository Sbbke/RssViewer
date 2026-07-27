# DBReader

`DBReader` is the read-only implementation of the CQRS data layer.

Unlike `DBWriter`, the reader may assemble response DTOs from multiple SQL tables when all required information resides in SQLite. It never performs business logic, AI generation, or filesystem access.

The reader is responsible for:

- Executing SQL queries.
- Mapping database models into DTOs.
- Assembling SQL-backed DTOs.
- Returning lightweight objects optimized for the Service layer.

Cross-storage aggregation (such as combining SQLite data with local summaries or generated slide assets) belongs to the Service layer.

---

# Topic Operations

## GetTopics

Returns every Topic in the system.

### Signature

```go
GetTopics() ([]dto.TopicResponse, error)
```

### Returns

| Type | Description |
|------|-------------|
| `[]dto.TopicResponse` | Every Topic ordered by creation time (newest first). |
| `error` | SQL execution failure. |

### Behavior

- Reads the `topic` table.
- Converts rows into `TopicResponse`.
- Does not populate RSS, Summary, or Slide information.

---

## GetTopicByID

Returns a single Topic.

### Signature

```go
GetTopicByID(
    id int64,
) (dto.TopicResponse, error)
```

### Input

| Parameter | Type | Description |
|-----------|------|-------------|
| id | int64 | Target Topic ID |

### Returns

| Type | Description |
|------|-------------|
| `dto.TopicResponse` | Topic metadata. |
| `error` | Topic not found or SQL error. |

### Behavior

- Queries the `topic` table.
- Returns a lightweight Topic DTO.

---

## GetTopicWithRss

Returns a Topic together with every subscribed RSS feed.

### Signature

```go
GetTopicWithRss(
    id int64,
) (dto.TopicResponse, error)
```

### Returns

| Type | Description |
|------|-------------|
| `dto.TopicResponse` | Topic with nested RSS items. |
| `error` | Query failure. |

### Behavior

- Retrieves Topic metadata.
- Retrieves associated RSS subscriptions through `rss_topics`.
- Assembles a complete `TopicResponse`.

Because every required dataset exists in SQLite, assembly belongs to `DBReader`.

---

# RSS Operations

## GetAllRss

Returns every RSS subscription.

### Signature

```go
GetAllRss() ([]dto.RssItem, error)
```

### Returns

| Type | Description |
|------|-------------|
| `[]dto.RssItem` | All RSS subscriptions. |
| `error` | SQL execution failure. |

---

## GetRssByID

Returns an RSS feed together with all Posts.

### Signature

```go
GetRssByID(
    rssID int64,
) (dto.RssResponse, error)
```

### Input

| Parameter | Type | Description |
|-----------|------|-------------|
| rssID | int64 | RSS ID |

### Returns

| Type | Description |
|------|-------------|
| `dto.RssResponse` | RSS metadata and nested Posts. |
| `error` | RSS not found or SQL failure. |

### Behavior

- Reads RSS metadata.
- Retrieves every Post under the RSS.
- Returns an assembled `RssResponse`.

---

## GetRssByTopic

Returns every RSS subscribed to a Topic.

### Signature

```go
GetRssByTopic(
    topicID int64,
) ([]dto.RssItem, error)
```

### Input

| Parameter | Type |
|-----------|------|
| topicID | int64 |

### Returns

`[]dto.RssItem`

### Behavior

- Joins `rss_topics`.
- Returns RSS metadata only.
- Ordered by subscription time.

---

## GetRssURL

Returns the original feed URL.

### Signature

```go
GetRssURL(
    rssID int64,
) (string, error)
```

### Returns

| Type | Description |
|------|-------------|
| `string` | Original RSS feed URL. |
| `error` | RSS not found. |

### Behavior

This helper function exists primarily for the RSS crawler when checking feed updates.

---

# Post Operations

## GetPostsByRss

Returns every Post belonging to an RSS feed.

### Signature

```go
GetPostsByRss(
    rssID int64,
) ([]dto.PostItem, error)
```

### Returns

Posts ordered by publication time (newest first).

---

## GetPostByID

Returns one Post.

### Signature

```go
GetPostByID(
    postID int64,
) (dto.PostItem, error)
```

### Returns

| Type | Description |
|------|-------------|
| `dto.PostItem` | Post metadata. |
| `error` | Post not found. |

### Behavior

Returns only metadata.

Raw content is retrieved separately through the content storage layer.

---

## GetPostURL

Returns the original article URL.

### Signature

```go
GetPostURL(
    postID int64,
) (string, error)
```

### Behavior

Used internally by the HTML processor as the starting point for downloading and processing article content.

---

## GetPostsByTopic

Returns every Post under a Topic.

### Signature

```go
GetPostsByTopic(
    topicID int64,
) ([]dto.PostItem, error)
```

### Behavior

- Joins Topic → RSS → Post.
- Returns all Posts belonging to every RSS under the Topic.
- Ordered by publication date (newest first).

---

## GetPostsByTopicInWindow

Returns Posts inside a time window.

### Signature

```go
GetPostsByTopicInWindow(
    topicID int64,
    from time.Time,
    to time.Time,
) ([]dto.PostItem, error)
```

### Input

| Parameter | Description |
|-----------|-------------|
| topicID | Topic ID |
| from | Inclusive lower bound |
| to | Exclusive upper bound |

### Returns

Chronologically ordered Posts.

### Behavior

Designed for `GenerateTopicBriefing`.

The Service specifies a weekly or monthly window, and the returned Posts become the inference input for the LLM.

Ascending order preserves chronological narrative.

---

## GetPostsByRssInWindow

Returns Posts inside a time window for a single RSS feed.

### Signature

```go
GetPostsByRssInWindow(
    rssID int64,
    from time.Time,
    to time.Time,
) ([]dto.PostItem, error)
```

### Behavior

Equivalent to `GetPostsByTopicInWindow`, but scoped to a single RSS source.

Typically used for feed-specific briefing generation.

---

# Internal Mapping

The reader completely hides the underlying database schema.

```
SQLite
    │
    ▼
Database Model
    │
    ▼
DTO Mapper
    │
    ▼
Response DTO
```

Database model types (`model.TopicModel`, `model.RSSModel`, `model.PostModel`) never escape the storage layer.

---

# DTO Assembly Policy

The following guideline determines where DTO composition occurs.

## Inside DBReader

Assemble DTOs when every required component originates from SQLite.

Examples:

- Topic + RSS
- RSS + Posts
- Topic metadata
- RSS metadata

---

## Inside Service Layer

Assemble DTOs when data originates from multiple storage backends.

Examples:

- Topic + Summary
- Topic + Slides
- TopicAllInOne
- PostDetailResponse
- BriefingSlideResponse
- BriefingTextResponse

This separation keeps `DBReader` responsible only for relational data while the Service layer coordinates external resources such as generated summaries, slide assets, and local storage.
