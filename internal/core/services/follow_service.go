package services

import (
	"context"
	"errors"

	"github.com/EstefiS/uala-challenge/internal/core/ports"
)

type followService struct {
	userRepo ports.UserRepository
}

func NewFollowService(userRepo ports.UserRepository) ports.FollowService {
	return &followService{userRepo: userRepo}
}

func (s *followService) FollowUser(ctx context.Context, currentUserID, userToFollowID string) error {
	if currentUserID == userToFollowID {
		return errors.New("a user cannot follow themselves")
	}
	return s.userRepo.FollowTx(ctx, currentUserID, userToFollowID)
}
