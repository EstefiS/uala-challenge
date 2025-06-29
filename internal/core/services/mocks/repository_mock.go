package mocks

import (
	"context"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type Repository struct {
	mock.Mock
}

// Implementación de ports.UserRepository
func (m *Repository) FollowTx(ctx context.Context, userID, userToFollowID string) error {
	args := m.Called(ctx, userID, userToFollowID)
	return args.Error(0)
}

func (m *Repository) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	args := m.Called(ctx, userID)
	// Hacemos un type assertion para devolver el slice de strings o nil
	if followers, ok := args.Get(0).([]string); ok {
		return followers, args.Error(1)
	}
	return nil, args.Error(1)
}

// Implementación de ports.TweetRepository
func (m *Repository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	args := m.Called(ctx, tweet)
	return args.Error(0)
}

// Implementación de ports.TimelineRepository
func (m *Repository) Get(ctx context.Context, userID string, limit int) ([]domain.Tweet, error) {
	args := m.Called(ctx, userID, limit)
	if timeline, ok := args.Get(0).([]domain.Tweet); ok {
		return timeline, args.Error(1)
	}
	return nil, args.Error(1)
}
