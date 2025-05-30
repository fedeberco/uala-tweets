package application_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"uala-tweets/internal/application"
	"uala-tweets/internal/domain"
)

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		setupMock   func(*application.MockUserRepository)
		expectErr   bool
		errType     error
		errContains string
	}{
		{
			name:     "successful user creation",
			username: "testuser",
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("Create", mock.MatchedBy(func(user *domain.User) bool {
					user.ID = 123 // Set a non-zero ID
					return user.Username == "testuser"
				})).Return(nil)
			},
			expectErr: false,
		},
		{
			name:     "empty username",
			username: "",
			setupMock: func(repo *application.MockUserRepository) {
				// No repository calls expected for empty username
			},
			expectErr:   true,
			errType:     &application.ErrInvalidInput{},
			errContains: "username cannot be empty",
		},
		{
			name:     "repository error",
			username: "testuser",
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("Create", mock.Anything).Return(assert.AnError)
			},
			expectErr:   true,
			errContains: "failed to create user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &application.MockUserRepository{}
			tt.setupMock(mockRepo)

			service := application.NewUserService(mockRepo)

			user, err := service.CreateUser(application.CreateUserInput{
				Username: tt.username,
			})

			if tt.expectErr {
				require.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.username, user.Username)
				assert.NotZero(t, user.ID)
				assert.WithinDuration(t, time.Now(), user.CreatedAt, time.Second)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		setupMock func(*application.MockUserRepository)
		expect    *domain.User
		expectErr bool
		errType   error
	}{
		{
			name:   "successful get user",
			userID: 1,
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("GetByID", 1).Return(&domain.User{
					ID:        1,
					Username:  "testuser",
					CreatedAt: time.Now(),
				}, nil)
			},
			expect: &domain.User{
				ID:       1,
				Username: "testuser",
			},
			expectErr: false,
		},
		{
			name:   "user not found",
			userID: 999,
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("GetByID", 999).Return((*domain.User)(nil), nil)
			},
			expect:    nil,
			expectErr: true,
			errType:   &application.ErrUserNotFound{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &application.MockUserRepository{}
			tt.setupMock(mockRepo)

			service := application.NewUserService(mockRepo)

			user, err := service.GetUser(tt.userID)

			if tt.expectErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.IsType(t, tt.errType, err)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect.ID, user.ID)
				assert.Equal(t, tt.expect.Username, user.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UserExists(t *testing.T) {
	tests := []struct {
		name      string
		userID    int
		setupMock func(*application.MockUserRepository)
		expect    bool
		expectErr bool
		errType   error
	}{
		{
			name:   "user exists",
			userID: 1,
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("Exists", 1).Return(true, nil)
			},
			expect:    true,
			expectErr: false,
		},
		{
			name:   "user does not exist",
			userID: 999,
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("Exists", 999).Return(false, nil)
			},
			expect:    false,
			expectErr: false,
		},
		{
			name:   "repository error",
			userID: 1,
			setupMock: func(repo *application.MockUserRepository) {
				repo.On("Exists", 1).Return(false, assert.AnError)
			},
			expect:    false,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &application.MockUserRepository{}
			tt.setupMock(mockRepo)

			service := application.NewUserService(mockRepo)

			exists, err := service.UserExists(tt.userID)

			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expect, exists)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
