package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

type RssService struct {
	orch   *storage.DataOrch
	client *http.Client
	parser *gofeed.Parser
}

func NewRssService(orch *storage.DataOrch) *RssService {
	return &RssService{
		orch:   orch,
		client: &http.Client{Timeout: 10 * time.Second},
		parser: gofeed.NewParser(),
	}
}

// ReceiveRssFromFrontend orchestrates the external HTTP fetch, parsing, and pipeline submission
func (s *RssService) SubmitRssUrl(ctx context.Context, rssURL string) (dto.RssItem, error) {
	// 1. Perform network I/O tasks entirely outside the database write loop to prevent pipeline blocking
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rssURL, nil)
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: failed to construct request: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: network fetch failed for %s: %w", rssURL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return dto.RssItem{}, fmt.Errorf("rss service: server returned non-200 status: %d", resp.StatusCode)
	}
	respBody := resp.Body
	// 2. Parse raw XML content into structured entities
	feed, err := s.parser.Parse(respBody)
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: failed to parse XML document feed: %w", err)
	}
	body, err := io.ReadAll(respBody)
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("error parsing resp body")
	}
	// 3. Register the newly verified feed metadata source in your database via the orchestrator
	rssSourceInfo := dto.RssPayload{
		Title: feed.Title,
		Url:   rssURL,
		Xml:   body,
	}

	mutationResult, err := s.orch.AddRss(rssSourceInfo)
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: failed to preserve feed registration: %w", err)
	}

	// 4. Transform gofeed native items into your system's data layer DTO payload array
	postPayloads := make([]dto.PostPayload, 0, len(feed.Items))
	for _, item := range feed.Items {
		publishedTime := time.Now()
		if item.PublishedParsed != nil {
			publishedTime = *item.PublishedParsed
		}

		postPayloads = append(postPayloads, dto.PostPayload{
			RssID:       mutationResult.GeneratedID, // Map the newly created identity constraint
			Title:       item.Title,
			Url:         item.Link,
			Content:     item.Description,
			PublishedAt: publishedTime.Format(time.RFC3339),
		})
	}

	// 5. Submit the array to the single-writer database channel queue
	if len(postPayloads) > 0 {
		if err := s.orch.AddPostBatch(postPayloads); err != nil {
			return dto.RssItem{}, fmt.Errorf("rss service: transaction processing batch allocation failed: %w", err)
		}
	}
	subscribedTime := time.Now()
	rssResult := dto.RssItem{
		ID:           mutationResult.GeneratedID,
		Title:        rssSourceInfo.Title,
		SubscribedAt: subscribedTime,
	}
	return rssResult, nil
}

func (s *RssService) RemoveRss(id int64) error{
	err := s.orch.DeleteRss(id)
	if err != nil{
		return fmt.Errorf("error during rss deletion: %v", err)
	}
	return nil
}
