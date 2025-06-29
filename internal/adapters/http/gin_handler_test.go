package http

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
	// Ponemos Gin en modo de prueba
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	h.SetupRoutes(router)
	return router
}

func TestGinHandler_publishTweet(t *testing.T) {
	t.Run("Success: should return 201 Created on successful tweet publication", func(t *testing.T) {
		// Setup
		mockTweetSvc := new(mocks.TweetService)
		// Los otros mocks no son necesarios para este test específico, pero el handler los requiere
		mockFollowSvc := new(mocks.FollowService)
		mockTimelineSvc := new(mocks.TimelineService)

		handler := NewGinHandler(mockTweetSvc, mockFollowSvc, mockTimelineSvc)
		router := setupRouter(handler)

		userID := "user-1"
		tweetText := "Mi primer tweet de prueba"

		// El tweet que esperamos que el servicio devuelva
		expectedTweet := &domain.Tweet{
			ID:        uuid.NewString(),
			UserID:    userID,
			Text:      tweetText,
			CreatedAt: time.Now(),
		}

		// Mocking: Le decimos al servicio mock qué devolver cuando sea llamado
		mockTweetSvc.On("PublishTweet", mock.Anything, userID, tweetText).Return(expectedTweet, nil)

		// Creamos el cuerpo de la petición JSON
		body, _ := json.Marshal(PublishTweetRequest{Text: tweetText})

		// Creamos la petición HTTP de prueba
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/tweets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID)

		// Usamos un ResponseRecorder para capturar la respuesta
		w := httptest.NewRecorder()

		// Execute
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusCreated, w.Code)

		var responseTweet domain.Tweet
		err := json.Unmarshal(w.Body.Bytes(), &responseTweet)
		assert.NoError(t, err)
		assert.Equal(t, expectedTweet.ID, responseTweet.ID)
		assert.Equal(t, expectedTweet.Text, responseTweet.Text)

		mockTweetSvc.AssertExpectations(t)
	})

	t.Run("Failure: should return 400 Bad Request if tweet is too long", func(t *testing.T) {
		mockTweetSvc := new(mocks.TweetService)
		mockFollowSvc := new(mocks.FollowService)
		mockTimelineSvc := new(mocks.TimelineService)

		handler := NewGinHandler(mockTweetSvc, mockFollowSvc, mockTimelineSvc)
		router := setupRouter(handler)

		userID := "user-1"
		tweetText := "Este texto es demasiado largo"

		// Mocking: Configuramos el mock para que devuelva el error específico
		mockTweetSvc.On("PublishTweet", mock.Anything, userID, tweetText).Return(nil, domain.ErrTweetTooLong)

		body, _ := json.Marshal(PublishTweetRequest{Text: tweetText})
		req, _ := http.NewRequest(http.MethodPost, "/api/v1/tweets", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-User-ID", userID)

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockTweetSvc.AssertExpectations(t)
	})
}
