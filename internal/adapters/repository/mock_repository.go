package repository

import (
	"context"
	"sort"
	"sync"

	"github.com/EstefiS/uala-challenge/internal/core/domain"
)

type MockRepository struct {
	mu        sync.RWMutex
	users     map[string]bool
	followers map[string]map[string]bool
	tweets    map[string]*domain.Tweet
	timelines map[string][]*domain.Tweet
}

func NewMockRepository() *MockRepository {
	return &MockRepository{
		users:     make(map[string]bool),
		followers: make(map[string]map[string]bool),
		tweets:    make(map[string]*domain.Tweet),
		timelines: make(map[string][]*domain.Tweet),
	}
}

func (r *MockRepository) ensureUserExists(userID string) {
	if !r.users[userID] {
		r.users[userID] = true
		r.followers[userID] = make(map[string]bool)
	}
}

// --- UserRepository ---
func (r *MockRepository) FollowTx(_ context.Context, userID, userToFollowID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ensureUserExists(userID)
	r.ensureUserExists(userToFollowID)

	r.followers[userToFollowID][userID] = true
	return nil
}

func (r *MockRepository) GetFollowers(_ context.Context, userID string) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var followers []string
	if followerMap, ok := r.followers[userID]; ok {
		for id := range followerMap {
			followers = append(followers, id)
		}
	}
	return followers, nil
}

// --- TweetRepository ---
func (r *MockRepository) PublishTx(_ context.Context, tweet *domain.Tweet) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.ensureUserExists(tweet.UserID)
	r.tweets[tweet.ID] = tweet

	followersToUpdate := []string{}
	if followers, ok := r.followers[tweet.UserID]; ok {
		for id := range followers {
			followersToUpdate = append(followersToUpdate, id)
		}
	}

	followersToUpdate = append(followersToUpdate, tweet.UserID)

	for _, id := range followersToUpdate {
		r.timelines[id] = append(r.timelines[id], tweet)
	}
	return nil
}

// --- TimelineRepository ---
func (r *MockRepository) Get(_ context.Context, userID string, limit int) ([]domain.Tweet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	timelinePointers, ok := r.timelines[userID]
	if !ok {
		return []domain.Tweet{}, nil
	}

	sort.SliceStable(timelinePointers, func(i, j int) bool {
		return timelinePointers[i].CreatedAt.After(timelinePointers[j].CreatedAt)
	})
	if len(timelinePointers) > limit {
		timelinePointers = timelinePointers[:limit]
	}

	resultTimeline := make([]domain.Tweet, len(timelinePointers))
	for i, tweetPtr := range timelinePointers {
		resultTimeline[i] = *tweetPtr
	}

	return resultTimeline, nil
}
