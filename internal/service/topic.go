package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
	"fmt"
)

type TopicService struct {
	orch *storage.DataOrch
}

// return all the topics
func (s *TopicService) GetTopics() ([]dto.TopicResponse, error) {
	rows, err := s.orch.GetReader().GetTopics()
	if err != nil {
		return nil, fmt.Errorf("topics service: getTopics: %w", err)
	}

	return rows, nil
}

// return topic
func (s *TopicService) GetTopicResponse(topicID int64) (dto.TopicResponse, error) {
	r, err := s.orch.GetReader().GetTopicWithRss(topicID)
	if err != nil {
		return dto.TopicResponse{}, fmt.Errorf("TopicService.GetTopicResponse: topic=%d: %w", topicID, err)
	}
	return r, nil
}
