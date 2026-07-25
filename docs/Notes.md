# Questions
## question 1 
When pull the actual content of the post, the result is in their own html file, should I (1) process the html, transfer to DTO suitable for MY application frontend that can utilize the customized style and UI, or (2) render the html directly in frontend since it is webui
- Short answer: First option is more ideal.
- But discarded for now, since rendering the post's content isn't the first priority for this application, even with the name of RSSViewer.
- However, we still need the raw content from the posts for generating the summary.


## DTO
The dto design start with using samll objecy with only id reference and title, besides from the benefit of small initial payload and lazy loading, it allows the adaptability if the hydrated version need to be adopt in the future.


## Architectural Refinement: Separation by Data Characteristics
The current architecture does not utilize the Strategy Pattern (e.g., a unified type DatalayerWorker interface { GetReader(); GetWriter() }) to swap between localized and SQL-based workers. This choice is intentional. In our specific use cases, the SQLite database and local disk storage serve entirely different asset-storage purposes and must operate alongside each other within the same workflow. Consequently, the DataOrch (Orchestrator) needs to maintain instances of both layers simultaneously.

Because the underlying reason for splitting our storage into a local file system and an SQLite database is to separate data based on its structural characteristics, metadata is channeled into SQL, while heavy content blobs (large files) are routed to local disk storage.

To reflect this logic, we can restructure our package layout into storage/meta (for SQL) and storage/content (for local files). The reasoning behind this design is that even if we migrate the system in the future, the raw data formats and underlying characteristics will remain unchanged. Separating metadata—which is optimized for quick indexing, lookups, and sorting—from large asset files allows us to achieve two critical performance goals:

Relieve data layer congestion by keeping heavy payload operations out of the relational engine.

Optimize database utilization by ensuring each storage engine fits its ideal technical use case, preventing performance degradation.

Structural Evolution Matrix
Plaintext
Current Component Design:
DataOrch {
    reader.SQL
    reader.Local
    writer.SQL
    writer.Local
}

Proposed Structural Design (Based on Characteristics):
DataOrch {
    meta.SQL
    content.Local
}

WriteJob Dispatch Lifecycle:
WriteJob {
    meta.SQL.Writer()
    content.Local.Writer()
}
## Orchestration and Concurrency Invariants
The Orchestrator manages the data layer workers by enforcing a strict single-writer-at-a-time invariant. All mutating write tasks submitted to DataOrch enter a centralized worker queue. The Orchestrator runs an internal background loop to pull tasks sequentially from this queue, dispatching them to the correct writer sub-module depending on the specific task type.

In contrast, the read path is far simpler: the system comfortably handles multiple concurrent reads executing simultaneously, regardless of whether they hit the SQL database or the local filesystem blocks. However, to ensure total system consistency, we must explicitly design safeguards to handle edge cases involving "write-after-read" scenarios or handling concurrent writes while a long-running read operation is still active.

## Rss and Topic
The execution flow of the application is ambigious, The schema's rss_topics is a many-to-many junction table, but RssService.SubmitRssUrl only calls s.orch.AddRss(...) — it never inserts into rss_topics. So after calling SubmitRssUrl, the new feed exists in rss but is orphaned; it won't show up under any topic's GetTopicWithRss. May need either:

    - a topicID parameter threaded through SubmitRssUrl → orch.AddRss (or a follow-up call) that inserts the rss_topics row, or
    - a separate LinkRssToTopic(rssID, topicID) method on the service/orch, called right after submission.

