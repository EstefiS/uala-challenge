package services

import (
	"context"
	"errors"
	"testing"

	"github.com/EstefiS/uala-challenge/internal/core/services/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTweetService_PublishTweet(t *testing.T) {
	ctx := context.Background()

	t.Run("Success: should publish a valid tweet", func(t *testing.T) {
		// Setup
		mockRepo := new(mocks.Repository)
		tweetService := NewTweetService(mockRepo)

		userID := "user-1"
		text := "Hola mundo"
		mockRepo.On("PublishTx", ctx, mock.AnythingOfType("*domain.Tweet")).Return(nil)

		// Execute
		tweet, err := tweetService.PublishTweet(ctx, userID, text)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, tweet)
		assert.Equal(t, userID, tweet.UserID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure: repository returns an error", func(t *testing.T) {
		// Setup
		mockRepo := new(mocks.Repository)
		tweetService := NewTweetService(mockRepo)

		expectedError := errors.New("database is down")

		// Mocking
		mockRepo.On("PublishTx", ctx, mock.AnythingOfType("*domain.Tweet")).Return(expectedError)

		// Execute
		_, err := tweetService.PublishTweet(ctx, "user-1", "un tweet")

		// Assert
		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}
