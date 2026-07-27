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
