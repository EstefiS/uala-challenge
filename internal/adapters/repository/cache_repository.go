package repository

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
	"github.com/redis/go-redis/v9"
)

type CachingRepository struct {
	redisClient      *redis.Client
	nextUserRepo     ports.UserRepository
	nextTweetRepo    ports.TweetRepository
	nextTimelineRepo ports.TimelineRepository
	logger           *slog.Logger
	ttl              time.Duration
}

func NewCachingRepository(
	client *redis.Client,
	userRepo ports.UserRepository,
	tweetRepo ports.TweetRepository,
	timelineRepo ports.TimelineRepository,
	logger *slog.Logger,
) *CachingRepository {
	return &CachingRepository{
		redisClient:      client,
		nextUserRepo:     userRepo,
		nextTweetRepo:    tweetRepo,
		nextTimelineRepo: timelineRepo,
		logger:           logger.With("component", "CachingRepository"),
		ttl:              2 * time.Minute,
	}
}

func timelineCacheKey(userID string) string {
	return "timeline:" + userID
}

func (r *CachingRepository) Get(ctx context.Context, userID string, limit int) ([]domain.Tweet, error) {
	cacheKey := timelineCacheKey(userID)

	val, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		r.logger.Debug("Cache HIT for user's timeline", "userID", userID)
		var timeline []domain.Tweet
		if json.Unmarshal([]byte(val), &timeline) == nil {
			return timeline, nil
		}
	}

	if err != redis.Nil {
		r.logger.Warn("Redis error on GET (not a cache miss)", "error", err, "key", cacheKey)
	}

	r.logger.Debug("Cache MISS for user's timeline", "userID", userID)
	timeline, err := r.nextTimelineRepo.Get(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	if len(timeline) > 0 {
		go func() {
			bgCtx := context.Background()

			data, marshalErr := json.Marshal(timeline)
			if marshalErr != nil {
				r.logger.Error("Background cache population: failed to marshal timeline", "error", marshalErr, "userID", userID)
				return
			}

			if err := r.redisClient.Set(bgCtx, cacheKey, data, r.ttl).Err(); err != nil {
				r.logger.Error("Background cache population: failed to set cache", "error", err, "userID", userID)
			}
		}()
	}

	return timeline, nil
}

func (r *CachingRepository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	err := r.nextTweetRepo.PublishTx(ctx, tweet)
	if err != nil {
		return err
	}

	followers, err := r.nextUserRepo.GetFollowers(ctx, tweet.UserID)
	if err != nil {
		r.logger.Error("Failed to get followers for cache invalidation", "error", err, "userID", tweet.UserID)
		return nil
	}

	if len(followers) == 0 {
		return nil
	}

	pipe := r.redisClient.Pipeline()
	for _, followerID := range followers {
		pipe.Del(ctx, timelineCacheKey(followerID))
	}
	_, err = pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		r.logger.Error("Failed to execute cache invalidation pipeline", "error", err)
	}

	r.logger.Info("Cache invalidated for follower timelines", "count", len(followers))

	return nil
}

func (r *CachingRepository) FollowTx(ctx context.Context, userID, userToFollowID string) error {
	err := r.nextUserRepo.FollowTx(ctx, userID, userToFollowID)
	if err == nil {
		r.logger.Info("Invalidating timeline cache for new follower", "userID", userID)
		if err := r.redisClient.Del(ctx, timelineCacheKey(userID)).Err(); err != nil {
			r.logger.Warn("Failed to invalidate cache on follow", "error", err, "userID", userID)
		}
	}
	return err
}

func (r *CachingRepository) GetFollowers(ctx context.Context, userID string) ([]string, error) {
	return r.nextUserRepo.GetFollowers(ctx, userID)
}
