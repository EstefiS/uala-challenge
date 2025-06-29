package mocks

import (
	"context"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type TweetService struct {
	mock.Mock
}

func (m *TweetService) PublishTweet(ctx context.Context, userID, text string) (*domain.Tweet, error) {
	args := m.Called(ctx, userID, text)
	if tweet, ok := args.Get(0).(*domain.Tweet); ok {
		return tweet, args.Error(1)
	}
	return nil, args.Error(1)
}

type FollowService struct {
	mock.Mock
}

func (m *FollowService) FollowUser(ctx context.Context, currentUserID, userToFollowID string) error {
	args := m.Called(ctx, currentUserID, userToFollowID)
	return args.Error(0)
}

type TimelineService struct {
	mock.Mock
}

func (m *TimelineService) GetUserTimeline(ctx context.Context, userID string) ([]domain.Tweet, error) {
	args := m.Called(ctx, userID)
	if timeline, ok := args.Get(0).([]domain.Tweet); ok {
		return timeline, args.Error(1)
	}
	return nil, args.Error(1)
}
