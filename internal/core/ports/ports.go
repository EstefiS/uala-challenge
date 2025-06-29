package ports

import (
	"context"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
)

type UserRepository interface {
	FollowTx(ctx context.Context, userID, userToFollowID string) error
	GetFollowers(ctx context.Context, userID string) ([]string, error)
}

type TweetRepository interface {
	PublishTx(ctx context.Context, tweet *domain.Tweet) error
}

type TimelineRepository interface {
	Get(ctx context.Context, userID string, limit int) ([]domain.Tweet, error)
}

// ==========================

type TweetService interface {
	PublishTweet(ctx context.Context, userID, text string) (*domain.Tweet, error)
}

type FollowService interface {
	FollowUser(ctx context.Context, currentUserID, userToFollowID string) error
}

type TimelineService interface {
	GetUserTimeline(ctx context.Context, userID string) ([]domain.Tweet, error)
}
