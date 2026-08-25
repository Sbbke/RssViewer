package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
	"fmt"
)

// TopicService provides topic-related operations on top of the DataOrch,
// composing DBReader (via GetReader) for reads and DataOrch's serialized
// write API (AddTopic, DeleteTopic) for mutations.
type TopicService struct {
	orch *storage.DataOrch
}

func NewTopicService(orch *storage.DataOrch) *TopicService {
	return &TopicService{
		orch: orch,
	}
}

// GetTopics returns all topics.
func (s *TopicService) GetTopics() ([]dto.TopicResponse, error) {
	rows, err := s.orch.GetReader().GetTopics()
	if err != nil {
		return nil, fmt.Errorf("topic service: getTopics: %w", err)
	}
	return rows, nil
}

// GetTopicResponse returns a single topic along with its linked RSS feeds.
func (s *TopicService) GetTopicResponse(topicID int64) (dto.TopicResponse, error) {
	r, err := s.orch.GetReader().GetTopicWithRss(topicID)
	if err != nil {
		return dto.TopicResponse{}, fmt.Errorf("topic service: getTopicResponse: topic=%d: %w", topicID, err)
	}
	return r, nil
}

// CreateTopic creates a new topic and returns the persisted record.
func (s *TopicService) CreateTopic(name string) (dto.TopicResponse, error) {
	req := dto.TopicPayload{
		Name: name,
	}
	mr, err := s.orch.AddTopic(req)
	if err != nil {
		return dto.TopicResponse{}, fmt.Errorf("topic service: createTopic: %w", err)
	}

	r, err := s.orch.GetReader().GetTopicByID(mr.GeneratedID)
	if err != nil {
		return dto.TopicResponse{}, fmt.Errorf("topic service: createTopic: fetch created: %w", err)
	}
	return r, nil
}

// DeleteTopic deletes a topic by ID. Linked RSS relationships are removed
// via ON DELETE CASCADE on rss_topics; the RSS feeds themselves are untouched.
func (s *TopicService) DeleteTopic(id int64) error {
	if err := s.orch.DeleteTopic(id); err != nil {
		return fmt.Errorf("topic service: deleteTopic: %w", err)
	}
	return nil
}



func (s *TopicService) GetTopicDetail(id int64) (dto.TopicAllInOne, error) {
	reader := s.orch.GetReader()
 
	// 1. Topic metadata.
	topics, err := reader.GetTopics()
	if err != nil {
		return dto.TopicAllInOne{}, fmt.Errorf("topic service: getTopicDetail: load topics: %w", err)
	}
	var createdAt string
	found := false
	for _, t := range topics {
		if t.TopicID == id {
			createdAt = t.CreatedAt
			found = true
			break
		}
	}
	if !found {
		return dto.TopicAllInOne{}, fmt.Errorf("topic service: getTopicDetail: topic %d not found", id)
	}
 
	// 2. Feeds linked to this topic.
	rssItems, err := reader.GetRssByTopic(id)
	if err != nil {
		return dto.TopicAllInOne{}, fmt.Errorf("topic service: getTopicDetail: rss for topic %d: %w", id, err)
	}
 
	// 3. Build RssDetailResponse from the existing shallow RssResponse
	//    read — no new reader method required.
	rssDetails := make([]dto.RssDetailResponse, 0, len(rssItems))
	for _, item := range rssItems {
		rssResp, err := reader.GetRssByID(item.ID)
		if err != nil {
			return dto.TopicAllInOne{}, fmt.Errorf("topic service: getTopicDetail: rss %d: %w", item.ID, err)
		}
 
		posts := make([]dto.PostDetailResponse, 0, len(rssResp.Posts))
		for _, p := range rssResp.Posts {
			posts = append(posts, dto.PostDetailResponse{
				ID:          p.ID,
				Title:       p.Title,
				PublishedAt: p.PublishedAt,
				Content:     p.Content,
				// Summary, Slide: intentionally nil — per-post
				// hydration isn't implemented yet.
			})
		}
 
		rssDetails = append(rssDetails, dto.RssDetailResponse{
			Info:  rssResp.Info,
			Posts: posts,
		})
	}
 
	return dto.TopicAllInOne{
		TopicID: id,
		Rss:     rssDetails,
		// Summary, Slide: intentionally nil — topic-level briefing
		// generation isn't implemented yet.
		Summary:   nil,
		Slide:     nil,
		CreatedAt: createdAt,
	}, nil
}

