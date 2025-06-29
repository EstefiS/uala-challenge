package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
	"github.com/gin-gonic/gin"
)

type GinHandler struct {
	tweetSvc    ports.TweetService
	followSvc   ports.FollowService
	timelineSvc ports.TimelineService
}

func NewGinHandler(tweetSvc ports.TweetService, followSvc ports.FollowService, timelineSvc ports.TimelineService) *GinHandler {
	return &GinHandler{tweetSvc, followSvc, timelineSvc}
}

func extractUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "header X-User-ID is required"})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

func (h *GinHandler) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.Use(extractUserID())
	{
		api.POST("/tweets", h.publishTweet)
		api.POST("/users/:id/follow", h.followUser)
		api.GET("/timeline", h.getTimeline)
	}
}

// publishTweet creates a new tweet.
// @Summary      Publish a Tweet
// @Description  Allows an authenticated user to publish a new message (tweet).
// @Tags         Tweets
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string                  true  "ID of the user publishing the tweet"
// @Param        tweet      body      PublishTweetRequest     true  "Tweet content"
// @Success      201        {object}  domain.Tweet            "Tweet created successfully"
// @Failure      400        {object}  map[string]string       "Bad request (e.g., tweet too long)"
// @Failure      401        {object}  map[string]string       "Unauthorized (missing X-User-ID header)"
// @Failure      500        {object}  map[string]string       "Internal server error"
// @Router       /tweets [post]
func (h *GinHandler) publishTweet(c *gin.Context) {
	userID := c.GetString("userID")
	var req PublishTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tweet, err := h.tweetSvc.PublishTweet(c.Request.Context(), userID, req.Text)
	if err != nil {
		if errors.Is(err, domain.ErrTweetTooLong) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error_code": "TWEET_TOO_LONG",
				"message":    err.Error(),
			})
			return
		}

		log.Printf("Error inesperado al publicar tweet: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error_code": "INTERNAL_SERVER_ERROR",
			"message":    "An unexpected server error has occurred",
		})

		return
	}

	c.JSON(http.StatusCreated, tweet)
}

// followUser allows a user to follow another.
// @Summary      Follow a User
// @Description  The current user (identified by X-User-ID) starts following another user (identified by their ID in the URL).
// @Tags         Users
// @Produce      json
// @Param        X-User-ID  header    string  true  "ID of the user performing the action"
// @Param        id         path      string  true  "ID of the user to follow"
// @Success      200        {object}  map[string]string
// @Failure      401        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /users/{id}/follow [post]
func (h *GinHandler) followUser(c *gin.Context) {
	currentUserID := c.GetString("userID")
	userToFollowID := c.Param("id")
	if err := h.followSvc.FollowUser(c.Request.Context(), currentUserID, userToFollowID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// getTimeline retrieves the user's tweet feed.
// @Summary      Get Timeline
// @Description  Returns a list of the most recent tweets from users the current user follows.
// @Tags         Timeline
// @Produce      json
// @Param        X-User-ID  header    string  true  "ID of the user whose timeline is being requested"
// @Success      200        {array}   domain.Tweet
// @Failure      401        {object}  map[string]string
// @Failure      500        {object}  map[string]string
// @Router       /timeline [get]
func (h *GinHandler) getTimeline(c *gin.Context) {
	userID := c.GetString("userID")
	timeline, err := h.timelineSvc.GetUserTimeline(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, timeline)
}
