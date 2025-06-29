package services

import (
	"context"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
)

type timelineService struct {
	timelineRepo ports.TimelineRepository
}

func NewTimelineService(timelineRepo ports.TimelineRepository) ports.TimelineService {
	return &timelineService{timelineRepo: timelineRepo}
}

func (s *timelineService) GetUserTimeline(ctx context.Context, userID string) ([]domain.Tweet, error) {
	return s.timelineRepo.Get(ctx, userID, 50)
}
