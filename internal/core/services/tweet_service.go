package services

import (
	"context"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
)

type tweetService struct {
	tweetRepo ports.TweetRepository
}

func NewTweetService(tweetRepo ports.TweetRepository) ports.TweetService {
	return &tweetService{tweetRepo: tweetRepo}
}

func (s *tweetService) PublishTweet(ctx context.Context, userID, text string) (*domain.Tweet, error) {
	tweet, err := domain.NewTweet(userID, text)
	if err != nil {
		return nil, err
	}
	return tweet, s.tweetRepo.PublishTx(ctx, tweet)
}
