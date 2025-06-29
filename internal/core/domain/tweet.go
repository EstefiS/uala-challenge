package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const MaxTweetLength = 280

var ErrTweetTooLong = errors.New("tweet exceeds 280 character limit")

type Tweet struct {
	ID        string
	UserID    string
	Text      string
	CreatedAt time.Time
}

func NewTweet(userID, text string) (*Tweet, error) {
	if len(text) > MaxTweetLength {
		return nil, ErrTweetTooLong
	}
	return &Tweet{
		ID:        uuid.NewString(),
		UserID:    userID,
		Text:      text,
		CreatedAt: time.Now(),
	}, nil
}
