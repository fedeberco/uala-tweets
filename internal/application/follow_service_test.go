package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"uala-tweets/internal/application"
	"uala-tweets/internal/domain"
)

type MockFollowPublisher struct {
	mock.Mock
}

func (m *MockFollowPublisher) PublishFollowEvent(event domain.FollowEvent) error {
	args := m.Called(event)
	return args.Error(0)
}

type testMocks struct {
	userRepo   *application.MockUserRepository
	followRepo *application.MockFollowRepository
	followPub  *MockFollowPublisher
}

func newTestMocks() *testMocks {
	return &testMocks{
		userRepo:   &application.MockUserRepository{},
		followRepo: &application.MockFollowRepository{},
		followPub:  &MockFollowPublisher{},
	}
}

func TestFollowService_Follow(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*testMocks)
		followerID int
		followedID int
		expectErr  bool
		errType    error
	}{
		{
			name: "successful follow",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 2).Return(true, nil)
				m.followRepo.On("IsFollowing", 1, 2).Return(false, nil)
				m.followRepo.On("Follow", 1, 2).Return(nil)
				m.followPub.On("PublishFollowEvent", mock.MatchedBy(func(e domain.FollowEvent) bool {
					return e.FollowerID == 1 && e.FollowedID == 2 && e.Following
				})).Return(nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  false,
		},
		{
			name: "cannot follow self",
			setupMocks: func(m *testMocks) {
				// No repository calls expected for self-follow
			},
			followerID: 1,
			followedID: 1,
			expectErr:  true,
			errType:    &application.ErrAlreadyFollowing{},
		},
		{
			name: "follower not found",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 999,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "followed user not found",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 1,
			followedID: 999,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "already following",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 2).Return(true, nil)
				m.followRepo.On("IsFollowing", 1, 2).Return(true, nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrAlreadyFollowing{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMocks()
			tt.setupMocks(m)

			service := application.NewFollowService(m.userRepo, m.followRepo, m.followPub)

			err := service.Follow(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			m.userRepo.AssertExpectations(t)
			m.followRepo.AssertExpectations(t)
			m.followPub.AssertExpectations(t)
		})
	}
}

func TestFollowService_Unfollow(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(m *testMocks)
		followerID int
		followedID int
		expectErr  bool
		errType    error
	}{
		{
			name: "successful unfollow",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 2).Return(true, nil)
				m.followRepo.On("IsFollowing", 1, 2).Return(true, nil)
				m.followRepo.On("Unfollow", 1, 2).Return(nil)
				m.followPub.On("PublishFollowEvent", mock.MatchedBy(func(e domain.FollowEvent) bool {
					return e.FollowerID == 1 && e.FollowedID == 2 && !e.Following
				})).Return(nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  false,
		},
		{
			name: "follower not found",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 999,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "followed user not found",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 1,
			followedID: 999,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "not following",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 2).Return(true, nil)
				m.followRepo.On("IsFollowing", 1, 2).Return(false, nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrNotFollowing{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMocks()
			tt.setupMocks(m)

			service := application.NewFollowService(m.userRepo, m.followRepo, m.followPub)

			err := service.Unfollow(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			m.userRepo.AssertExpectations(t)
			m.followRepo.AssertExpectations(t)
			m.followPub.AssertExpectations(t)
		})
	}
}

func TestFollowService_IsFollowing(t *testing.T) {
	tests := []struct {
		name        string
		setupMocks  func(m *testMocks)
		followerID  int
		followedID  int
		expect      bool
		expectError bool
	}{
		{
			name: "is following",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 2).Return(true, nil)
				m.followRepo.On("IsFollowing", 1, 2).Return(true, nil)
			},
			followerID:  1,
			followedID:  2,
			expect:      true,
			expectError: false,
		},
		{
			name: "is not following",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 2).Return(true, nil)
				m.followRepo.On("IsFollowing", 1, 2).Return(false, nil)
			},
			followerID:  1,
			followedID:  2,
			expect:      false,
			expectError: false,
		},
		{
			name: "follower not found",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID:  999,
			followedID:  2,
			expect:      false,
			expectError: true,
		},
		{
			name: "followed user not found",
			setupMocks: func(m *testMocks) {
				m.userRepo.On("Exists", 1).Return(true, nil)
				m.userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID:  1,
			followedID:  999,
			expect:      false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newTestMocks()
			tt.setupMocks(m)

			service := application.NewFollowService(m.userRepo, m.followRepo, m.followPub)

			result, err := service.IsFollowing(tt.followerID, tt.followedID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, result)
			}

			m.userRepo.AssertExpectations(t)
			m.followRepo.AssertExpectations(t)
		})
	}
}
