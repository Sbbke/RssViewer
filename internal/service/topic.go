package service

import (
	"RssViewer/internal/dto"
	"RssViewer/internal/storage"
	"fmt"
)


type TopicService struct{
	orch *storage.DataOrch
}

// return all the topics 
func (s *TopicService) GetTopics() ( []dto.TopicResponse,  error) {
	rows, err := s.orch.GetReader().GetTopics()
	if err != nil{
		return nil, fmt.Errorf("topics service: getTopics: %w",err)
	}

	return rows, nil
}

// return topic
func (s *TopicService) GetTopicResponse(topicID int64) (dto.TopicResponse, error) {
	r, err := s.orch.GetReader().GetTopicWithRss(topicID)
	if err != nil {
		return dto.TopicResponse{} , fmt.Errorf("TopicService.GetTopicResponse: topic=%d: %w", topicID, err)
	}
	return r, nil
}
 
func (s *TopicService) CreateTopic(name string) (dto.TopicResponse, error){
	req := dto.TopicPayload{
		Name: name,
	}
	mr, err := s.orch.AddTopic(req)

	if err != nil{
		return dto.TopicResponse{}, fmt.Errorf("error create topic:%w",err)
	}
		
	r, err := s.orch.GetReader().GetTopicByID(mr.GeneratedID)
	if err != nil{
		return dto.TopicResponse{}, fmt.Errorf("error fetching created topic: %w",err)
	}

	return r, err
}
