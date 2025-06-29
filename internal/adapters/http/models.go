package http

import (
	"log/slog"

	"github.com/EstefiS/uala-challenge/internal/core/ports"
)

type PublishTweetRequest struct {
	Text string `json:"text" binding:"required"`
}

type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}

type HandlerDependencies struct {
	TweetSvc    ports.TweetService
	FollowSvc   ports.FollowService
	TimelineSvc ports.TimelineService
	Logger      *slog.Logger
}
