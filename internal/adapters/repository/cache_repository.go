package repository

import (
	"context"
	"encoding/json"
	"log"
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
	ttl              time.Duration
}

func NewCachingRepository(
	client *redis.Client,
	userRepo ports.UserRepository,
	tweetRepo ports.TweetRepository,
	timelineRepo ports.TimelineRepository,
) *CachingRepository {
	return &CachingRepository{
		redisClient:      client,
		nextUserRepo:     userRepo,
		nextTweetRepo:    tweetRepo,
		nextTimelineRepo: timelineRepo,
		ttl:              2 * time.Minute,
	}
}

func (r *CachingRepository) Get(ctx context.Context, userID string, limit int) ([]domain.Tweet, error) {
	cacheKey := "timeline:" + userID

	val, err := r.redisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		log.Printf("Cache HIT for user's timeline: %s", userID)
		var timeline []domain.Tweet
		if json.Unmarshal([]byte(val), &timeline) == nil {
			return timeline, nil
		}
	}

	if err != redis.Nil {
		log.Printf("Redis error proceeding to DB : %v", err)
	}

	log.Printf("Cache MISS for user's timeline: %s. Querying the DB.", userID)
	timeline, err := r.nextTimelineRepo.Get(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	if len(timeline) > 0 {
		data, marshalErr := json.Marshal(timeline)
		if marshalErr == nil {
			r.redisClient.Set(ctx, cacheKey, data, r.ttl)
		}
	}

	return timeline, nil
}

func (r *CachingRepository) PublishTx(ctx context.Context, tweet *domain.Tweet) error {
	err := r.nextTweetRepo.PublishTx(ctx, tweet)
	if err != nil {
		return err
	}

	log.Printf("Invalidating timeline cache for author: %s", tweet.UserID)
	authorCacheKey := "timeline:" + tweet.UserID
	r.redisClient.Del(ctx, authorCacheKey)

	return nil
}

func (r *CachingRepository) FollowTx(ctx context.Context, userID, userToFollowID string) error {
	err := r.nextUserRepo.FollowTx(ctx, userID, userToFollowID)
	if err == nil {
		log.Printf("Invalidating timeline cache for new follower: %s", userID)
		r.redisClient.Del(ctx, "timeline:"+userID)
	}
	return err
}
