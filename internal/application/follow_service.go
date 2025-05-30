package application

import (
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
	if followerID == followedID {
		return NewErrAlreadyFollowing(followerID, followedID)
	}

	followerExists, err := s.userRepo.Exists(followerID)
	if err != nil {
		return err
	}
	if !followerExists {
		return NewErrUserNotFound(followerID)
	}

	followedExists, err := s.userRepo.Exists(followedID)
	if err != nil {
		return err
	}
	if !followedExists {
		return NewErrUserNotFound(followedID)
	}

	isFollowing, err := s.followRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return err
	}
	if isFollowing {
		return NewErrAlreadyFollowing(followerID, followedID)
	}

	return s.followRepo.Follow(followerID, followedID)
}

func (s *FollowService) Unfollow(followerID, followedID int) error {
	followerExists, err := s.userRepo.Exists(followerID)
	if err != nil {
		return err
	}
	if !followerExists {
		return NewErrUserNotFound(followerID)
	}

	followedExists, err := s.userRepo.Exists(followedID)
	if err != nil {
		return err
	}
	if !followedExists {
		return NewErrUserNotFound(followedID)
	}

	isFollowing, err := s.followRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return err
	}
	if !isFollowing {
		return NewErrNotFollowing(followerID, followedID)
	}

	return s.followRepo.Unfollow(followerID, followedID)
}

func (s *FollowService) IsFollowing(followerID, followedID int) (bool, error) {
	followerExists, err := s.userRepo.Exists(followerID)
	if err != nil {
		return false, err
	}
	if !followerExists {
		return false, NewErrUserNotFound(followerID)
	}

	followedExists, err := s.userRepo.Exists(followedID)
	if err != nil {
		return false, err
	}
	if !followedExists {
		return false, NewErrUserNotFound(followedID)
	}

	isFollowing, err := s.followRepo.IsFollowing(followerID, followedID)
	if err != nil {
		return false, err
	}

	return isFollowing, nil
}
