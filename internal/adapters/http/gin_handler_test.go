package http

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EstefiS/uala-challenge/internal/adapters/http/mocks"
	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupRouter(h *GinHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	h.SetupRoutes(router)
	return router
}

func TestGinHandler_publishTweet(t *testing.T) {
	t.Run("Success: should return 201 Created on successful tweet publication", func(t *testing.T) {
		mockTweetSvc := new(mocks.TweetService)
		mockFollowSvc := new(mocks.FollowService)
		mockTimelineSvc := new(mocks.TimelineService)
		discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

		deps := HandlerDependencies{
			TweetSvc:    mockTweetSvc,
			FollowSvc:   mockFollowSvc,
			TimelineSvc: mockTimelineSvc,
			Logger:      discardLogger,
		}

		handler := NewGinHandler(deps)
		router := setupRouter(handler)

		userID := "user-1"
		tweetText := "Mi primer tweet de prueba"

		expectedTweet := &domain.Tweet{
			ID:        uuid.NewString(),
			UserID:    userID,
			Text:      tweetText,
			CreatedAt: time.Now(),
		}

		mockTweetSvc.On("PublishTweet", mock.Anything, userID, tweetText).Return(expectedTweet, nil)

		body, _ := json.Marshal(PublishTweetRequest{Text: tweetText})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/tweets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
		var responseTweet domain.Tweet
		err := json.Unmarshal(w.Body.Bytes(), &responseTweet)
		assert.NoError(t, err)
		assert.Equal(t, expectedTweet.ID, responseTweet.ID)

		mockTweetSvc.AssertExpectations(t)
	})

	t.Run("Failure: should return 400 Bad Request if tweet is too long", func(t *testing.T) {
		mockTweetSvc := new(mocks.TweetService)
		mockFollowSvc := new(mocks.FollowService)
		mockTimelineSvc := new(mocks.TimelineService)
		discardLogger := slog.New(slog.NewTextHandler(io.Discard, nil))

		deps := HandlerDependencies{
			TweetSvc:    mockTweetSvc,
			FollowSvc:   mockFollowSvc,
			TimelineSvc: mockTimelineSvc,
			Logger:      discardLogger,
		}

		handler := NewGinHandler(deps)
		router := setupRouter(handler)

		userID := "user-1"
		longTweetText := strings.Repeat("a", 281)

		mockTweetSvc.On("PublishTweet", mock.Anything, userID, longTweetText).Return(nil, domain.ErrTweetTooLong)

		body, _ := json.Marshal(PublishTweetRequest{Text: longTweetText})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/tweets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockTweetSvc.AssertExpectations(t)
	})
}
