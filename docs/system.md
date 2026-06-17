# System Architecture
This is a ducument for describing system architecture in a structred way in natural language. NOT a software document, the detailed document for software development should be document in other file, such as API document, software specification, flow chart, class diagram, ... etc.


## Data Objects Design
### Local Storage  

- **Rss Feed** <br>
The local disk is strictly an asset cache holding the raw, untouched .xml payload exactly as it was fetched from the publisher's server.

- **Post Content** <br>
Processed content from html of a post.

- **Summary** <br>
have two types of summary: raw text and images slide.

```bash
/
├── rss/
│   └── {rss_id}/
|       ├── rss.xml 
│       ├── {post_id}.txt
│       └── {post_id}.html
└── summary/
    ├── topic/
    │   └── {topic_id}_{composite_rss_hash}/
    │       ├── summary.txt
    │       └── slide/
    │           ├── 1.png
    │           └── 2.png
    └── post/
        └── {post_id}/
            ├── summary.txt
            └── slide/
                ├── 1.png
                └── 2.png
```

### DB schema

### DTO
Creat both hydrated and id-reference dto object.

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

### Rss
Handle required functionalities to render content from rss
- GetPosts(): <br>
return all posts from a RSS
- CheckUpdate(): <br>
actively check if the publisher have new posts, if so then update the local rss to the new one.

### Posts
Handle the posts of a RSS.
- GetContent(): <br>
Return the raw text content of a post. First try to fetch from DB(either local file path or postgress, definately not SQLite), if it doesn't appears in DB, then call processor to fetch and process the html, save raw text content to DB, if every succesful, then try to read from DB.
- CheckUpdate(): <br>
Actively check the posts, compare it to the rss, if there is an update then re-run the process logic.

## Processor

- <interface>Processor

### RssProcessor

### HtmlProcessor
The processor is responsible for parse the raw html file to raw text to reduce the token size for generating summary.
- Flexibility: <br>
The raw html come from different sources, each have their own naming of html element class, to process the html into much more precise and accurate content (without too much noise content, such as ads, extra links, dupilicate title ... etc.), the html processor should implement different parsing policy to identified html source.

## datalayer 
CQRS
SQLite handles concurrent reads brilliantly, but it only allows one single write operation at a time (it locks the database file). The automated scrapers might try to write newly crawled post pointers or summaries concurrently.

### Orchestrator
Manage goroutine for data manipulation.
- DBReader <br>
Dedicated to Read operation
> Services can request a reader for read operation
- DBWriter <br>
Implement basic CUD operation to 1.DB, and 2. Local Disk.
> Channel all write operations through a single Go worker goroutine to prevent "database is locked" (SQLITE_BUSY) errors.

### DB

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

- For every Update operation, call Delete operation to remove target endpoint(rss, post summary, topic summary, or topic slide) then call Create operations.



