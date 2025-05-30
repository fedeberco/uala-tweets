package application

import (
	"uala-tweets/internal/domain"

	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(id int) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Exists(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

type MockFollowRepository struct {
	mock.Mock
}

func (m *MockFollowRepository) Follow(followerID, followedID int) error {
	args := m.Called(followerID, followedID)
	return args.Error(0)
}

func (m *MockFollowRepository) Unfollow(followerID, followedID int) error {
	args := m.Called(followerID, followedID)
	return args.Error(0)
}

func (m *MockFollowRepository) IsFollowing(followerID, followedID int) (bool, error) {
	args := m.Called(followerID, followedID)
	return args.Bool(0), args.Error(1)
}
