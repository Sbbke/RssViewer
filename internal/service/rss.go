package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/processor"
	"RssViewer/internal/storage"
	"context"
	"fmt"
	"net/http"
	"time"
)

type RssService struct {
	orch   *storage.DataOrch
	client *http.Client
	processor *processor.RssProcessor
}

func NewRssService(orch *storage.DataOrch) *RssService {
	return &RssService{
		orch:   orch,
		client: &http.Client{Timeout: 10 * time.Second},
		processor: processor.NewRssProcessor(),
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

	// 2. Register feed shell first to get the DB-generated ID
	//    Title is unknown until after parsing, so we do a two-phase write:
	//    persist URL now, update title after processing.
	mutation, err := s.orch.AddRss(dto.RssPayload{Url: rssURL})
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: register feed: %w", err)
	}

	// 3. Process — parse XML and build post payloads stamped with the real ID
	rssUpdate, posts, err := s.processor.Run(resp.Body, mutation.GeneratedID)
	if err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: process feed: %w", err)
	}

	// 4. Patch the feed row with title + raw XML now that we have them

	if err := s.orch.UpdateRss(rssUpdate); err != nil {
		return dto.RssItem{}, fmt.Errorf("rss service: update feed metadata: %w", err)
	}

	// 5. Persist posts
	if len(posts) > 0 {
		if err := s.orch.AddPostBatch(posts); err != nil {
			return dto.RssItem{}, fmt.Errorf("rss service: persist posts: %w", err)
		}
	}

	return dto.RssItem{
		ID:           mutation.GeneratedID,
		Title:        rssUpdate.Body.Title,
		SubscribedAt: time.Now(),
	}, nil
}

func (s *RssService) CheckUpdate() error{
	return nil
}

func (s *RssService) RemoveRss(id int64) error{
	err := s.orch.DeleteRss(id)
	if err != nil{
		return fmt.Errorf("error during rss deletion: %v", err)
	}
	return nil
}

