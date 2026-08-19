package service

import (
	"RssViewer/internal/storage"
	"context"
)

type BriefingService struct {
	orch *storage.DataOrch
}

func (s *BriefingService) GetLatestBriefing(ctx context.Context, topicID int64) error {
	return nil
}

func (s *BriefingService) GenerateBriefing(ctx context.Context, targetID int64, targetType string) error {
	return nil
}
func (s *BriefingService) GetBreifingByTopic(ctx context.Context, topicID int64) error{
	return nil
}

func (s *BriefingService) GetBreifingByPost(ctx context.Context, postID int64, rssID int64) error{
	return nil
}
