package application

import "fmt"

type (
	ErrUserNotFound struct {
		UserID int
	}

	ErrUserAlreadyExists struct {
		Username string
	}

	ErrAlreadyFollowing struct {
		FollowerID int
		FollowedID int
	}

	ErrNotFollowing struct {
		FollowerID int
		FollowedID int
	}
)

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found with id: %d", e.UserID)
}

func (e ErrUserAlreadyExists) Error() string {
	return fmt.Sprintf("user with username '%s' already exists", e.Username)
}

func (e ErrAlreadyFollowing) Error() string {
	return fmt.Sprintf("user %d is already following user %d", e.FollowerID, e.FollowedID)
}

func (e ErrNotFollowing) Error() string {
	return fmt.Sprintf("user %d is not following user %d", e.FollowerID, e.FollowedID)
}

func NewErrUserNotFound(userID int) error {
	return &ErrUserNotFound{UserID: userID}
}

func NewErrUserAlreadyExists(username string) error {
	return &ErrUserAlreadyExists{Username: username}
}

func NewErrAlreadyFollowing(followerID, followedID int) error {
	return &ErrAlreadyFollowing{
		FollowerID: followerID,
		FollowedID: followedID,
	}
}

func NewErrNotFollowing(followerID, followedID int) error {
	return &ErrNotFollowing{
		FollowerID: followerID,
		FollowedID: followedID,
	}
}
