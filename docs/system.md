# System Architecture
This is a ducument for describing system architecture in a structred way in natural language. NOT a software document, the detailed document for software development should be document in other file, such as API document, software specification, flow chart, class diagram, ... etc.

The system is devided into Three layer: Frontent (UI/Client) Layer, Service Layer and Data Layer.
Frontend communicate with backend through Service Layer, then Service Layer interact with data through Data Layer.


## Data Objects Design
### Local Storage  
- **Summary** <br>
Only have images slide.

```bash
── summary/
    ├── topic/
    │   └── {topic_id}_{composite_rss_hash}/
    │       └── slide/
    │           ├── 1.png
    │           └── 2.png
    └── post/
        └── {post_id}/
            └── slide/
                ├── 1.png
                └── 2.png
```
## DB Schema

### RSS Feed
Stores the raw, untouched RSS XML exactly as fetched from the publisher.

### Post
Stores post metadata together with the processed plain-text content extracted from HTML.

### Topic
Logical grouping of one or more RSS feeds.

### Topic ↔ RSS
Topics and RSS feeds have a many-to-many relationship through `rss_topics`.

### Post Summary
Stores generated briefing text for an individual post.

### Topic Summary
Stores generated briefing text for an entire topic.

### Slides
Slide assets are generated externally and referenced by the service layer rather than stored directly in SQLite.

---

# DTO Layer

The DTO package defines the API contract between the frontend, service layer, and data layer.

## Response DTOs

### BriefingTextResponse

Represents generated briefing text.

```go
Body
CreatedAt
```

### BriefingSlideResponse

Represents generated slide assets.

```go
Slides []string
CreatedAt
```

### TopicResponse

Lightweight topic information.

```go
TopicID
Name
Rss []RssItem
Summary *BriefingTextResponse
SummaryID
CreatedAt
```

Summary is optional.
A nil pointer indicates the summary has not been generated.

---

### TopicAllInOne

Complete topic payload returned when opening a topic.

```go
TopicID
Rss []RssDetailResponse
Summary *BriefingTextResponse
Slide *BriefingSlideResponse
CreatedAt
```

Unlike TopicResponse, this DTO contains complete RSS and Post information.

---

### RssResponse

```go
Info RssItem
Posts []PostItem
```

---

### RssDetailResponse

```go
Info RssItem
Posts []PostDetailResponse
```

---

### PostDetailResponse

```go
ID
Title
PublishedAt
Content
Summary *BriefingTextResponse
Slide *BriefingSlideResponse
```

---

### PostSummaryResponse

```go
Meta PostItem
Summary *BriefingTextResponse
```

---

## Mutation DTOs

### TopicPayload

```go
Name
```

---

### RssPayload

```go
Title
Url
Xml
```

---

### PostPayload

```go
RssID
Title
Url
Content
PublishedAt
```

---

### SummaryPayload

Generic payload used by both TopicSummary and PostSummary.

```go
ID
Content
```

---

### TopicSlidePayload

```go
TopicID
Slide [][]byte
RssHash
```

---

### PostSlidePayload

```go
PostID
Slide [][]byte
```

---

### LinkRssTopicPayload

```go
RssID
TopicID
```

---

### MutationResult

Returned by every mutation.

```go
GeneratedID
Err
```

---

# Service Layer

The Service layer coordinates data retrieval, generation, and persistence while keeping business logic independent from storage implementation.

---

## Summary Service

Retrieval is intentionally separated from AI generation.

```
GetBriefing()
GenerateBriefing()
```

Supported briefing scopes:

- Topic
- RSS
- Post
- Arbitrary collection of Posts

### GetBriefing()

Fast read path.

Queries DBReader for an existing generated briefing and immediately returns the corresponding DTO.

No AI inference is executed.

---

### GenerateTopicBriefing()

Workflow:

1. Read all RSS under the topic.
2. Retrieve required posts.
3. Load processed post contents.
4. Batch content into prompt windows.
5. Execute local LLM inference.
6. Persist TopicSummary.
7. Optionally generate slides.
8. Return TopicAllInOne.

---

### GeneratePostBriefing()

Workflow:

1. Retrieve processed post content.
2. Execute inference.
3. Store PostSummary.
4. Optionally generate slides.

---

### GenerateRssBriefing()

Workflow:

1. Retrieve all posts under an RSS.
2. Aggregate contents.
3. Generate briefing.
4. Return BriefingTextResponse.

---

# Topic Service

Responsible for Topic management.

### APIs

- GetTopics()
- GetTopicResponse()
- GetTopicAllInOne()
- CreateTopic()
- UpdateTopic()
- DeleteTopic()
- LinkRss()
- UnlinkRss()

---

# RSS Service

Responsible for RSS lifecycle.

### APIs

- SubmitRssUrl()
- GetRss()
- CheckUpdate()
- UpdateRss()
- DeleteRss()

---

# Post Service

Responsible for Post lifecycle.

### APIs

- GetContent()
- GetPost()
- CheckUpdate()
- DeletePost()

GetContent() first checks local storage. If unavailable, HTML is downloaded, processed into plain text, persisted, and then returned.

---

# Processor

Processors convert external resources into normalized content suitable for downstream LLM inference.

---

## Processor Interface

```
Processor
```

---

## RSSProcessor

Responsibilities

- Download RSS XML
- Parse feed metadata
- Detect new or updated posts
- Produce PostPayload batches

---

## HtmlProcessor

Transforms raw HTML into clean plain text.

Responsibilities

- Remove advertisements
- Remove duplicated titles
- Remove navigation
- Remove unrelated links
- Preserve article structure

Because publishers use different HTML structures, the processor should support pluggable extraction policies.

---

# Data Layer

Implements CQRS.

SQLite permits concurrent readers but only a single writer.

All write requests are serialized through a dedicated writer goroutine.

---

## Write Workflow

1. Service constructs DTO payload.
2. DTO is wrapped into a WriteTask.
3. Orchestrator queues the task.
4. DBWriter converts DTO → Model.
5. SQL executes mutation.
6. Local assets are written if necessary.

Database schema details never leak outside DBWriter.

---

## Read Workflow

1. Service requests DBReader.
2. Reader executes SQL.
3. Reader assembles response DTOs.
4. Multi-source DTOs are composed inside the Service layer.

---
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
   └── LocalReader---
```

# DBReader

Read-only interface.

### Topic

- GetTopics()
- GetTopicByID()

### RSS

- GetRssByID()
- GetRssByTopic()

### Post

- GetPostByID()
- GetPostsByRss()
- GetPostsByTopic()

### Summary

- GetTopicSummary()
- GetPostSummary()

---

# DBWriter

Responsible for Create / Update / Delete operations.

### Topic

- CreateTopic()
- UpdateTopic()
- DeleteTopic()

---

### RSS

- CreateRss()
- UpdateRss()
- DeleteRss()

---

### Post

- CreatePost()
- CreatePostsBatch()
- UpdatePostTitle()
- DeletePost()

---

### Summary

#### Topic

- CreateTopicSummary()
- UpdateTopicSummary()
- DeleteTopicSummary()

#### Post

- CreatePostSummary()
- UpdatePostSummary()
- DeletePostSummary()

---

### Topic Relations

- LinkRssTopic()
- UnlinkRssTopic()

---

# Local Storage

Large assets are stored on disk instead of SQLite.

```


summary/
    post/{postID}.txt
    topic/{topicID}.txt

slides/
    post/{postID}/
    topic/{topicID}/
```

Update operations should use atomic replacement:

```
Delete → Create
```

to prevent partially written files.

