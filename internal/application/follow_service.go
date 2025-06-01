package application

import (
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/publishers"
	"uala-tweets/internal/ports/repositories"
)

type FollowService struct {
	userRepo   repositories.UserRepository
	followRepo repositories.FollowRepository
	followPub  publishers.FollowPublisher
}

func NewFollowService(
	userRepo repositories.UserRepository,
	followRepo repositories.FollowRepository,
	followPub publishers.FollowPublisher,
) *FollowService {
	return &FollowService{
		userRepo:   userRepo,
		followRepo: followRepo,
		followPub:  followPub,
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

	if err := s.followRepo.Follow(followerID, followedID); err != nil {
		return err
	}

	// Publish follow event
	event := domain.FollowEvent{
		FollowerID: followerID,
		FollowedID: followedID,
		Following:  true,
	}
	if err := s.followPub.PublishFollowEvent(event); err != nil {
		// Log the error but don't fail the operation
		// The system can still function without the event being published
	}

	return nil
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

	if err := s.followRepo.Unfollow(followerID, followedID); err != nil {
		return err
	}

	// Publish unfollow event
	event := domain.FollowEvent{
		FollowerID: followerID,
		FollowedID: followedID,
		Following:  false,
	}
	if err := s.followPub.PublishFollowEvent(event); err != nil {
		// Log the error but don't fail the operation
		// The system can still function without the event being published
	}

	return nil
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
