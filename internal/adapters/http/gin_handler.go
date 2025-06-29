package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
	"github.com/EstefiS/uala-challenge/internal/core/services"
	"github.com/gin-gonic/gin"
)

type GinHandler struct {
	deps   HandlerDependencies
	logger *slog.Logger
}

func NewGinHandler(deps HandlerDependencies) *GinHandler {
	return &GinHandler{
		deps:   deps,
		logger: deps.Logger,
	}
}

func (h *GinHandler) badRequest(c *gin.Context, errorCode, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{
		ErrorCode: errorCode,
		Message:   message,
	})
}

func (h *GinHandler) internalServerError(c *gin.Context, err error, attributes ...slog.Attr) {
	h.logger.Error("Internal server error", "error", err, "attributes", attributes)
	c.JSON(http.StatusInternalServerError, ErrorResponse{
		ErrorCode: "INTERNAL_SERVER_ERROR",
		Message:   "An unexpected server error has occurred.",
	})
}

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

func (h *GinHandler) SetupRoutes(router *gin.Engine) {
	api := router.Group("/api/v1")
	api.Use(extractUserID())
	{
		api.POST("/tweets", h.publishTweet)
		api.POST("/users/:id/follow", h.followUser)
		api.GET("/timeline", h.getTimeline)
	}
}

func (h *GinHandler) publishTweet(c *gin.Context) {
	userID := c.GetString("userID")

	var req PublishTweetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.badRequest(c, "INVALID_REQUEST_BODY", err.Error())
		return
	}

	tweet, err := h.deps.TweetSvc.PublishTweet(c.Request.Context(), userID, req.Text)
	if err != nil {
		if errors.Is(err, domain.ErrTweetTooLong) {
			h.badRequest(c, "TWEET_TOO_LONG", err.Error())
			return
		}
		h.internalServerError(c, err, slog.String("userID", userID))
		return
	}

	c.JSON(http.StatusCreated, tweet)
}

func (h *GinHandler) followUser(c *gin.Context) {
	currentUserID := c.GetString("userID")
	userToFollowID := c.Param("id")

	if currentUserID == userToFollowID {
		h.badRequest(c, "INVALID_OPERATION", "A user cannot follow themselves.")
		return
	}

	if err := h.deps.FollowSvc.FollowUser(c.Request.Context(), currentUserID, userToFollowID); err != nil {
		if errors.Is(err, services.ErrSelfFollow) {
			h.badRequest(c, "INVALID_OPERATION", err.Error())
			return
		}
		h.internalServerError(c, err, slog.String("follower", currentUserID), slog.String("followee", userToFollowID))
		return
	}

	c.JSON(http.StatusOK, StatusResponse{Status: "ok"})
}

func (h *GinHandler) getTimeline(c *gin.Context) {
	userID := c.GetString("userID")

	timeline, err := h.deps.TimelineSvc.GetUserTimeline(c.Request.Context(), userID)
	if err != nil {
		h.internalServerError(c, err, slog.String("userID", userID))
		return
	}

	c.JSON(http.StatusOK, timeline)
}
