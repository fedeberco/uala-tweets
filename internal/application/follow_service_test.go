package application_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"uala-tweets/internal/application"
)

func TestFollowService_Follow(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*application.MockUserRepository, *application.MockFollowRepository)
		followerID int
		followedID int
		expectErr  bool
		errType    error
	}{
		{
			name: "successful follow",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 2).Return(true, nil)
				followRepo.On("IsFollowing", 1, 2).Return(false, nil)
				followRepo.On("Follow", 1, 2).Return(nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  false,
		},
		{
			name: "cannot follow self",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				// No repository calls expected for self-follow
			},
			followerID: 1,
			followedID: 1,
			expectErr:  true,
			errType:    &application.ErrAlreadyFollowing{},
		},
		{
			name: "follower not found",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 999,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "followed user not found",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 1,
			followedID: 999,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "already following",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 2).Return(true, nil)
				followRepo.On("IsFollowing", 1, 2).Return(true, nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrAlreadyFollowing{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &application.MockUserRepository{}
			mockFollowRepo := &application.MockFollowRepository{}
			tt.setupMocks(mockUserRepo, mockFollowRepo)

			service := application.NewFollowService(mockUserRepo, mockFollowRepo)

			err := service.Follow(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockFollowRepo.AssertExpectations(t)
		})
	}
}

func TestFollowService_Unfollow(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*application.MockUserRepository, *application.MockFollowRepository)
		followerID int
		followedID int
		expectErr  bool
		errType    error
	}{
		{
			name: "successful unfollow",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 2).Return(true, nil)
				followRepo.On("IsFollowing", 1, 2).Return(true, nil)
				followRepo.On("Unfollow", 1, 2).Return(nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  false,
		},
		{
			name: "follower not found",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 999,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "followed user not found",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 1,
			followedID: 999,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "not following",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 2).Return(true, nil)
				followRepo.On("IsFollowing", 1, 2).Return(false, nil)
			},
			followerID: 1,
			followedID: 2,
			expectErr:  true,
			errType:    &application.ErrNotFollowing{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &application.MockUserRepository{}
			mockFollowRepo := &application.MockFollowRepository{}
			tt.setupMocks(mockUserRepo, mockFollowRepo)

			service := application.NewFollowService(mockUserRepo, mockFollowRepo)

			err := service.Unfollow(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
			}

			mockUserRepo.AssertExpectations(t)
			mockFollowRepo.AssertExpectations(t)
		})
	}
}

func TestFollowService_IsFollowing(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*application.MockUserRepository, *application.MockFollowRepository)
		followerID int
		followedID int
		expect     bool
		expectErr  bool
		errType    error
	}{
		{
			name: "is following",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 2).Return(true, nil)
				followRepo.On("IsFollowing", 1, 2).Return(true, nil)
			},
			followerID: 1,
			followedID: 2,
			expect:     true,
			expectErr:  false,
		},
		{
			name: "is not following",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 2).Return(true, nil)
				followRepo.On("IsFollowing", 1, 2).Return(false, nil)
			},
			followerID: 1,
			followedID: 2,
			expect:     false,
			expectErr:  false,
		},
		{
			name: "follower not found",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 999,
			followedID: 2,
			expect:     false,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
		{
			name: "followed user not found",
			setupMocks: func(userRepo *application.MockUserRepository, followRepo *application.MockFollowRepository) {
				userRepo.On("Exists", 1).Return(true, nil)
				userRepo.On("Exists", 999).Return(false, nil)
			},
			followerID: 1,
			followedID: 999,
			expect:     false,
			expectErr:  true,
			errType:    &application.ErrUserNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUserRepo := &application.MockUserRepository{}
			mockFollowRepo := &application.MockFollowRepository{}
			tt.setupMocks(mockUserRepo, mockFollowRepo)

			service := application.NewFollowService(mockUserRepo, mockFollowRepo)

			isFollowing, err := service.IsFollowing(tt.followerID, tt.followedID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, isFollowing)
			}

			mockUserRepo.AssertExpectations(t)
			mockFollowRepo.AssertExpectations(t)
		})
	}
}
