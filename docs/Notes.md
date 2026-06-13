# Questions
## question 1 
When pull the actual content of the post, the result is in their own html file, should I (1) process the html, transfer to DTO suitable for MY application frontend that can utilize the customized style and UI, or (2) render the html directly in frontend since it is webui
- Short answer: First option is more ideal.
- But discarded for now, since rendering the post's content isn't the first priority for this application, even with the name of RSSViewer.
- However, we still need the raw content from the posts for generating the summary.


## DTO
The dto design start with using samll objecy with only id reference and title, besides from the benefit of small initial payload and lazy loading, it allows the adaptability if the hydrated version need to be adopt in the future.



