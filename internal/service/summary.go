package service

import "context"

type BriefingService struct {
}

func (s *BriefingService) GetLatestBriefing(ctx context.Context, topicID int64) error {
	return nil
}

func (s *BriefingService) GenerateBriefing(ctx context.Context, targetID int64, targetType string) error {
	return nil
}

