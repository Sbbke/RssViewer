package processor

import (
	"RssViewer/internal/dto"
	"fmt"
	"io"
	"time"

	"github.com/mmcdole/gofeed"
)

type RssProcessor struct {
	parser *gofeed.Parser
}

func NewRssProcessor() *RssProcessor{
	return &RssProcessor{
		parser: gofeed.NewParser(),
	}
}

// Process parses the RSS body and transforms items into DTO payloads.
// rssID is injected by the service after the feed row is persisted —
// the processor itself has no knowledge of the DB.
func (r *RssProcessor) Run(body io.Reader, rssID int64) (dto.RssUpdatePayload, []dto.PostPayload,error) {
	// Drain into buffer so we can both parse and store the raw XML.
	raw, err := io.ReadAll(body)
	if err != nil {
		return dto.RssUpdatePayload{},[]dto.PostPayload{},fmt.Errorf("rss processor: read body: %w", err)
	}

	feed, err := r.parser.ParseString(string(raw))
	if err != nil {
		return dto.RssUpdatePayload{},[]dto.PostPayload{},fmt.Errorf("rss processor: parse feed: %w", err)
	}

	posts := make([]dto.PostPayload, 0, len(feed.Items))
	for _, item := range feed.Items {
		publishedAt := time.Now()
		if item.PublishedParsed != nil {
			publishedAt = *item.PublishedParsed
		}
		posts = append(posts, dto.PostPayload{
			RssID:       rssID,
			Title:       item.Title,
			Url:         item.Link,
			Content:     item.Description,
			PublishedAt: publishedAt.Format(time.RFC3339),
		})
	}
	
	rp := dto.RssPayload{
		Title: feed.Title,
		Xml: raw,
	}

	nr := dto.RssUpdatePayload{
		Id: rssID,
		Body: rp,
	}

	return nr, posts, nil
}
