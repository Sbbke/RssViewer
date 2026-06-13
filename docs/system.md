# System Architecture
This is a ducument for describing system architecture in a structred way in natural language. NOT a software document, the detailed document for software development should be document in other file, such as API document, software specification, flow chart, class diagram, ... etc.


## Data Objects Design
### Local Storage  

- **Rss Feed** <br>
The local disk is strictly an asset cache holding the raw, untouched .xml payload exactly as it was fetched from the publisher's server.

- **Summary** <br>
have two types of summary: raw text and images slide.
- **Post Content** <br>
Processed content from html of a post.

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


## datalayer 
CQRS
SQLite handles concurrent reads brilliantly, but it only allows one single write operation at a time (it locks the database file). The automated scrapers might try to write newly crawled post pointers or summaries concurrently.

The Fix: Ensure DBWriter is backed by a single connection pool configured with a busy timeout, or channel all write operations through a single Go worker goroutine to prevent "database is locked" (SQLITE_BUSY) errors.
- DBReader
- DBWriter
