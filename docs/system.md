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

### DB schema

- **Rss Feed** <br>
The holding the raw, untouched .xml payload exactly as it was fetched from the publisher's server.

- **Post Content** <br>
Processed content from html of a post.


### DTO

# Service Layer

## Services
### summary 
Separating retrieval (GetBriefing) from computational processing (GenerateBriefing), and seperate briefing from atom response to avoid huge chunk of data transfer


To generate a briefing, we should be able to generate briefing of
1. Topic
2. Post (s)
3. Rss (s)

+ <Interface>GenerateBriefing

- GetBriefing: <br>
*GetBriefing* should be a fast, low-overhead read path. It queries *DBReader* to pull the already generated weekly/monthly summary block from the SQLite Topic table and returns the DTO immediately to the UI.


- GenerateTopicBriefing <br>
*GenerateBriefing* coordinates reading the raw .xml files under a topic from disk, grab the content by week or moth depends on the requiest, batching that content into prompt context windows, executing inference via local SLM engine, and passing the final payload to *DBWriter* to update state.

- GeneratePostBriefing <br>

### Topic
Handle read activities relating to topic
- GetTopics(): <br>
return every topics store in the system 
- GetTopicResponse: <br>
return the ID(s) of rss(s) of a topic
- DeleteTopic: <br>
delete given topic id

### Rss
Handle required functionalities to render content from rss
- GetPosts(): <br>
return all posts from a RSS
- CheckUpdate(): <br>
actively check if the publisher have new posts, if so then update the local rss to the new one.
- SubmitRssUrl: <br>
with given rss url, crawl raw xml, then pocess and save corresponding posts meta into db.
- RemoveRss(): <br>
delete given rss.

### Posts
Handle the posts of a RSS.
- GetContent(): <br>
Return the raw text content of a post. First try to fetch from DB(either local file path or postgress, definately not SQLite), if it doesn't appears in DB, then call processor to fetch and process the html, save raw text content to DB, if every succesful, then try to read from DB.
- CheckUpdate(): <br>
Actively check the posts, compare it to the rss, if there is an update then re-run the process logic.
- DeletePost(): <br>
With given post id, delete it from db

## Processor

- <interface>Processor

### RssProcessor

### HtmlProcessor
The processor is responsible for parse the raw html file to raw text to reduce the token size for generating summary.
- Flexibility: <br>
The raw html come from different sources, each have their own naming of html element class, to process the html into much more precise and accurate content (without too much noise content, such as ads, extra links, dupilicate title ... etc.), the html processor should implement different parsing policy to identified html source.

# datalayer 
CQRS
SQLite handles concurrent reads brilliantly, but it only allows one single write operation at a time (it locks the database file). The automated scrapers might try to write newly crawled post pointers or summaries concurrently.

1. The Write Path Workflow
For mutations, the input data enters the system as a DTO payload constructed by your upstream application or API services.

The upstream service builds the specific payload structural type, such as PostPayload.

The Orchestrator ingests this DTO directly, wraps it within the WriteTask envelope, and delivers it to the worker channel pool without modification.

The concrete storage writer implementation inside writer/sql.go unpacks the DTO and handles the mapping to the schema variables right before binding the values to the database driver execution context.

This ordering ensures that if your database schema changes, the modification is completely contained within writer/sql.go, leaving the core Orchestrator code untouched.

2. The Read Path Workflow
For retrieval operations, your application circumvents the write serialization machinery entirely by leveraging parallel reads.

The upstream service requests a read connection handle from the Orchestrator via RequestRead.

The dedicated reader module executes the SQL query against the database engine.

The reader scans the row data directly into the properties of your DTO struct.

The internal database schema model should never escape the perimeter of your read execution files, ensuring the API boundary receives clean, aggregated data structures optimized for presentation.

## Orchestrator
Manage goroutine for data manipulation. Controlling over writer and reader goroutine, limiting only 1 writer. All data manipulation should go through the orchetrator, basically a datalayer manager.

- RequestReadTopic()
- RequestReadRss()
- RequestReadPost()
- RequestWriteTopic()
- RequestWriteRss()
- RequestWritePost()


## CQRS
- DBReader <br>
Dedicated to Read operation
> Services can request a reader for read operation
- DBWriter <br>
Implement basic CUD operation to 1.DB, and 2. Local Disk.
> Channel all write operations through a single Go worker goroutine to prevent "database is locked" (SQLITE_BUSY) errors.

``` bash
Functional grouping:
├── interface.go
├── orchestrate.go
├── reader
│   ├── local.go
│   ├── local_test.go
│   └── sql.go
├── sqlinit.go
└── writer
    ├── local.go
    ├── local_test.go
    └── sql.go
Technology grouping:
├── interface.go
├── orchestrate.go
├──local images
    ├── reader.go
│   └── writer.go
├── sql meta
    ├── writer.go
    └── reader.go
```
### DB
single-source DTOs are assembled in the reader; multi-source DTOs are assembled in the service.

** Sqlite Concrete Implementation **
- GetDB(): return sql.DB
- initDB(): take a path as input and init sqlite db

** reader **
- GetTopics
- GetTopicByID
- GetRssByTopic
- GetRssByID
- GetPostsByRss
- GetPostByID
- GetPostsByTopic

** writer **
- CreateTopic, UpdataTopic, DeleteTopic
- CreateRss, UpdataRss, DeleteRss
- CreatePost, UpdatePostTitle, DeletePost
- CreatePostBatch

### Local
Since every modification will go through db layer first (checking duplication, fetching essential data such as ID and title ...etc.), so implment fairly simple local file manipulation (CURD).

All path is initialized as the operation executed (if not exist) as defined above.

- CreateRss() : <br>
Save rss.xml file to the folder with given ID and Content.
- CreatePostSummary() : <br>
Save Post summary to the folder with given post.ID and summary.
- CreateTopicSummary() : <br>
Save Topic raw text summary to the folder with given topic.ID, related rss.ID and summary.
- CreateTopicSlide() : <br>
Save Topic raw image slide file to the folder with given topic.ID, related rss.ID and summary.

- For every Update operation, implement atomic swap: call Delete operation to remove (move to temp) target endpoint(rss, post summary, topic summary, or topic slide) then call Create operations.



