package http

type PublishTweetRequest struct {
	Text string `json:"text" binding:"required,max=280"`
}

type ErrorResponse struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}

type StatusResponse struct {
	Status string `json:"status" example:"ok"`
}
