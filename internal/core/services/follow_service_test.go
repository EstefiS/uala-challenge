package services

import (
	"context"
	"errors"
	"testing"

	"github.com/EstefiS/uala-challenge/internal/core/services/mocks"
	"github.com/stretchr/testify/assert"
)

func TestFollowService_FollowUser(t *testing.T) {
	ctx := context.Background()

	t.Run("Success: should follow a user", func(t *testing.T) {
		mockRepo := new(mocks.Repository)
		followService := NewFollowService(mockRepo)

		userID := "user-pepita"
		userToFollowID := "user-pepito"

		mockRepo.On("FollowTx", ctx, userID, userToFollowID).Return(nil)

		err := followService.FollowUser(ctx, userID, userToFollowID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure: should not allow self-follow", func(t *testing.T) {
		mockRepo := new(mocks.Repository)
		followService := NewFollowService(mockRepo)

		userID := "user-pepita"

		err := followService.FollowUser(ctx, userID, userID)

		assert.Error(t, err)
		assert.EqualError(t, err, "a user cannot follow themselves")
		mockRepo.AssertNotCalled(t, "FollowTx")
	})

	t.Run("Failure: repository returns an error", func(t *testing.T) {
		mockRepo := new(mocks.Repository)
		followService := NewFollowService(mockRepo)

		userID := "user-pepita"
		userToFollowID := "user-pepito"
		expectedError := errors.New("db connection error")

		mockRepo.On("FollowTx", ctx, userID, userToFollowID).Return(expectedError)

		err := followService.FollowUser(ctx, userID, userToFollowID)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		mockRepo.AssertExpectations(t)
	})
}
