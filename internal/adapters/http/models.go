package http

type PublishTweetRequest struct {
	Text string `json:"text" binding:"required,max=280"`
}
