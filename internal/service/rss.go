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
func (s *RssService) SubmitRssUrl(ctx context.Context, rssURL string, topicID *int64) (dto.RssItem, error) {
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
	if topicID != nil {
		if err := s.orch.LinkRssTopic(mutation.GeneratedID, *topicID); err != nil {
			return dto.RssItem{}, fmt.Errorf("rss service: link topic: %w", err)
		}
	}

	return dto.RssItem{
		ID:           mutation.GeneratedID,
		Title:        rssUpdate.Body.Title,
		SubscribedAt: time.Now(),
	}, nil
}

func (s *RssService) LinkRssToTopic(rssID, topicID int64) error {
	if err := s.orch.LinkRssTopic(rssID, topicID); err != nil {
		return fmt.Errorf("rss service: link topic: %w", err)
	}
	return nil
}

func (s *RssService) CheckUpdate(ctx context.Context, rssID int64) ( error) {
	// 1. Load the feed's source URL.
	existing, err := s.orch.GetReader().GetRssURL(rssID)
	if err != nil {
		return fmt.Errorf("rss service: checkUpdate: load feed %d: %w", rssID, err)
	}
	if existing == "" {
		return  fmt.Errorf("rss service: checkUpdate: feed %d has no source url on record", rssID)
	}
 
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, existing, nil)
	if err != nil {
		return  fmt.Errorf("rss service: checkUpdate: build request: %w", err)
	}
 
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("rss service: checkUpdate: fetch %s: %w", existing, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rss service: checkUpdate: server returned status %d", resp.StatusCode)
	}
 
	// 2. Parse using the same pipeline as initial subscribe, stamped
	//    with the existing ID so posts link to the right feed row.
	rssUpdate, posts, err := s.processor.Run(resp.Body, rssID)
	if err != nil {
		return fmt.Errorf("rss service: checkUpdate: process feed: %w", err)
	}
 
	// 3. Patch feed metadata — title/raw XML may have changed even if
	//    no posts are new.
	if err := s.orch.UpdateRss(rssUpdate); err != nil {
		return fmt.Errorf("rss service: checkUpdate: update feed metadata: %w", err)
	}
 
	// 4. Persist posts; the DB-level (rss_id, url) uniqueness handles
	//    dedupe, so this can safely submit the full parsed set every
	//    time rather than pre-filtering in application code.
	if len(posts) == 0 {
		return nil
	}

	if  err := s.orch.AddPostBatch(posts); err != nil {
		return fmt.Errorf("rss service: checkUpdate: persist posts: %w", err)
	}
 
	return nil
}
 
func (s *RssService) RemoveRss(id int64) error{
	err := s.orch.DeleteRss(id)
	if err != nil{
		return fmt.Errorf("error during rss deletion: %v", err)
	}
	return nil
}
func (s *RssService) UnlinkRssFromTopic(rssID, topicID int64) error {
	if err := s.orch.UnlinkRssTopic(rssID, topicID); err != nil {
		return fmt.Errorf("rss service: unlink topic: %w", err)
	}
	return nil
}

// GetAllRss returns every RSS feed, regardless of topic linkage.
func (s *RssService) GetAllRss() ([]dto.RssItem, error) {
	items, err := s.orch.GetReader().GetAllRss()
	if err != nil {
		return nil, fmt.Errorf("rss service: getAllRss: %w", err)
	}
	return items, nil
}

// GetStandaloneRss returns feeds that are not linked to any topic.
// Composed entirely from existing reader methods: fetch every topic,
// union the rss IDs linked to each, then diff against the full rss set.
func (s *RssService) GetStandaloneRss() ([]dto.RssItem, error) {
	all, err := s.orch.GetReader().GetAllRss()
	if err != nil {
		return nil, fmt.Errorf("rss service: getStandaloneRss: all rss: %w", err)
	}

	topics, err := s.orch.GetReader().GetTopics()
	if err != nil {
		return nil, fmt.Errorf("rss service: getStandaloneRss: topics: %w", err)
	}

	linked := make(map[int64]struct{})
	for _, t := range topics {
		rssForTopic, err := s.orch.GetReader().GetRssByTopic(t.TopicID)
		if err != nil {
			return nil, fmt.Errorf("rss service: getStandaloneRss: rss for topic %d: %w", t.TopicID, err)
		}
		for _, r := range rssForTopic {
			linked[r.ID] = struct{}{}
		}
	}

	standalone := make([]dto.RssItem, 0, len(all))
	for _, r := range all {
		if _, ok := linked[r.ID]; !ok {
			standalone = append(standalone, r)
		}
	}
	return standalone, nil
}
func (s *RssService) GetRssDetail(rssID int64) (dto.RssResponse, error) {
	r, err := s.orch.GetReader().GetRssByID(rssID)
	if err != nil {
		return dto.RssResponse{}, fmt.Errorf("rss service: getRssDetail: rssID=%d: %w", rssID, err)
	}
	return r, nil
}

func (s *RssService) UpdateRss(rssID int64) (dto.RssResponse, error) {
	res := dto.RssResponse{}
	return res, nil
}
