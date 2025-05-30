package application

import (
	"errors"

	"uala-tweets/internal/ports/repositories"
)

type FollowService struct {
	userRepo   repositories.UserRepository
	followRepo repositories.FollowRepository
}

func NewFollowService(userRepo repositories.UserRepository, followRepo repositories.FollowRepository) *FollowService {
	return &FollowService{
		userRepo:   userRepo,
		followRepo: followRepo,
	}
}

func (s *FollowService) Follow(followerID, followedID int) error {
	followerExists, err := s.userRepo.Exists(followerID)
	if err != nil {
		return err
	}
	if !followerExists {
		return errors.New("follower user does not exist")
	}

	followedExists, err := s.userRepo.Exists(followedID)
	if err != nil {
		return err
	}
	if !followedExists {
		return errors.New("followed user does not exist")
	}

	isFollowing, err := s.followRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return err
	}
	if isFollowing {
		return errors.New("already following this user")
	}

	return s.followRepo.Follow(followerID, followedID)
}

func (s *FollowService) Unfollow(followerID, followedID int) error {
	followerExists, err := s.userRepo.Exists(followerID)
	if err != nil {
		return err
	}
	if !followerExists {
		return errors.New("follower user does not exist")
	}

	followedExists, err := s.userRepo.Exists(followedID)
	if err != nil {
		return err
	}
	if !followedExists {
		return errors.New("followed user does not exist")
	}

	isFollowing, err := s.followRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return err
	}
	if !isFollowing {
		return errors.New("not currently following this user")
	}

	return s.followRepo.Unfollow(followerID, followedID)
}
