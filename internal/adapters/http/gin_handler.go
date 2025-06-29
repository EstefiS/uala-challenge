package http

import (
	"errors"
	"log"
	"net/http"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/ports"
	"github.com/gin-gonic/gin"
)

// GinHandler wraps all the services and provides the HTTP handlers.
type GinHandler struct {
	tweetSvc    ports.TweetService
	followSvc   ports.FollowService
	timelineSvc ports.TimelineService
}

// NewGinHandler creates a new instance of GinHandler.
func NewGinHandler(tweetSvc ports.TweetService, followSvc ports.FollowService, timelineSvc ports.TimelineService) *GinHandler {
	return &GinHandler{tweetSvc, followSvc, timelineSvc}
}

// extractUserID is a middleware to get the User ID from the "X-User-ID" header.
func extractUserID() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, ErrorResponse{
				ErrorCode: "UNAUTHORIZED",
				Message:   "X-User-ID header is required",
			})
			return
		}
		c.Set("userID", userID)
		c.Next()
	}
}

// SetupRoutes configures all the API routes on the Gin router.
func (h *GinHandler) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.Use(extractUserID()) // Apply middleware to the entire API group
	{
		api.POST("/tweets", h.publishTweet)
		api.POST("/users/:id/follow", h.followUser)
		api.GET("/timeline", h.getTimeline)
	}
}

// publishTweet creates a new tweet.
// @Summary      Publish a Tweet
// @Description  Allows an authenticated user to post a new message (tweet).
// @Tags         Tweets
// @Accept       json
// @Produce      json
// @Param        X-User-ID  header    string                  true  "ID of the user publishing the tweet"
// @Param        tweet      body      PublishTweetRequest     true  "Tweet Content"
// @Success      201        {object}  domain.Tweet            "Tweet created successfully"
// @Failure      400        {object}  http.ErrorResponse      "Bad request (e.g., tweet too long, invalid JSON)"
// @Failure      401        {object}  http.ErrorResponse      "Unauthorized (missing X-User-ID header)"
// @Failure      500        {object}  http.ErrorResponse      "Internal Server Error"
// @Router       /tweets [post]
func (h *GinHandler) publishTweet(c *gin.Context) {
	userID := c.GetString("userID")

	var req PublishTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			ErrorCode: "INVALID_REQUEST_BODY",
			Message:   err.Error(),
		})
		return
	}

	tweet, err := h.tweetSvc.PublishTweet(c.Request.Context(), userID, req.Text)
	if err != nil {
		if errors.Is(err, domain.ErrTweetTooLong) {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				ErrorCode: "TWEET_TOO_LONG",
				Message:   err.Error(),
			})
			return
		}

		log.Printf("Unexpected error publishing tweet: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			ErrorCode: "INTERNAL_SERVER_ERROR",
			Message:   "An unexpected server error has occurred.",
		})
		return
	}

	c.JSON(http.StatusCreated, tweet)
}

// followUser allows a user to follow another user.
// @Summary      Follow a User
// @Description  The current user (identified by X-User-ID) starts following another user (identified by their ID in the URL).
// @Tags         Users
// @Produce      json
// @Param        X-User-ID  header    string  true  "ID of the user performing the action"
// @Param        id         path      string  true  "ID of the user to follow"
// @Success      200        {object}  http.StatusResponse     "Successfully followed user"
// @Failure      401        {object}  http.ErrorResponse      "Unauthorized (missing X-User-ID header)"
// @Failure      500        {object}  http.ErrorResponse      "Internal Server Error"
// @Router       /users/{id}/follow [post]
func (h *GinHandler) followUser(c *gin.Context) {
	currentUserID := c.GetString("userID")
	userToFollowID := c.Param("id")

	if err := h.followSvc.FollowUser(c.Request.Context(), currentUserID, userToFollowID); err != nil {
		log.Printf("Unexpected error following user: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			ErrorCode: "INTERNAL_SERVER_ERROR",
			Message:   "An unexpected server error has occurred.",
		})
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "ok"})
}

// getTimeline gets the user's tweet feed.
// @Summary      Get Timeline
// @Description  Returns a list of the most recent tweets from users the current user follows.
// @Tags         Timeline
// @Produce      json
// @Param        X-User-ID  header    string  true  "ID of the user whose timeline is being requested"
// @Success      200        {array}   domain.Tweet            "A list of tweets"
// @Failure      401        {object}  http.ErrorResponse      "Unauthorized (missing X-User-ID header)"
// @Failure      500        {object}  http.ErrorResponse      "Internal Server Error"
// @Router       /timeline [get]
func (h *GinHandler) getTimeline(c *gin.Context) {
	userID := c.GetString("userID")

	timeline, err := h.timelineSvc.GetUserTimeline(c.Request.Context(), userID)
	if err != nil {
		log.Printf("Unexpected error getting timeline: %v", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			ErrorCode: "INTERNAL_SERVER_ERROR",
			Message:   "An unexpected server error has occurred.",
		})
		return
	}

	c.JSON(http.StatusOK, timeline)
}
