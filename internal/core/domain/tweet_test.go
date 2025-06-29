package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewTweet(t *testing.T) {
	t.Run("Success: should create a new tweet when text is valid", func(t *testing.T) {
		userID := "user-123"
		text := "Este es un tweet válido."

		tweet, err := NewTweet(userID, text)

		assert.NoError(t, err)
		assert.NotNil(t, tweet)
		assert.Equal(t, userID, tweet.UserID)
		assert.Equal(t, text, tweet.Text)
		assert.NotEmpty(t, tweet.ID)
		assert.NotZero(t, tweet.CreatedAt)
	})

	t.Run("Failure: should return an error when text is too long", func(t *testing.T) {
		userID := "user-123"
		// Creamos un texto de 281 caracteres
		longText := strings.Repeat("a", 281)

		tweet, err := NewTweet(userID, longText)

		assert.Error(t, err)
		assert.Nil(t, tweet)
		assert.Equal(t, ErrTweetTooLong, err)
	})
}
