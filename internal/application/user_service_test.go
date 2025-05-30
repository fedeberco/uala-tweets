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

type mockUserRepository struct {
	mock.Mock
}

func (m *mockUserRepository) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *mockUserRepository) GetByID(id int) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) Exists(id int) (bool, error) {
	args := m.Called(id)
	return args.Bool(0), args.Error(1)
}

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mockUserRepository)
		input      application.CreateUserInput
		expectErr  bool
		expectUser *domain.User
		expectFunc func(*testing.T, *domain.User, error)
	}{
		{
			name: "successful user creation",
			setupMock: func(m *mockUserRepository) {
				m.On("Create", mock.AnythingOfType("*domain.User")).
					Run(func(args mock.Arguments) {
						user := args.Get(0).(*domain.User)
						user.ID = 1
					}).
					Return(nil)
			},
			input: application.CreateUserInput{
				Username: "testuser",
			},
			expectErr: false,
			expectFunc: func(t *testing.T, user *domain.User, err error) {
				assert.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, 1, user.ID)
				assert.Equal(t, "testuser", user.Username)
				assert.False(t, user.CreatedAt.IsZero())
				assert.False(t, user.UpdatedAt.IsZero())
			},
		},
		{
			name: "empty username",
			setupMock: func(m *mockUserRepository) {
				// No repository calls expected
			},
			input: application.CreateUserInput{
				Username: "",
			},
			expectErr: true,
			expectFunc: func(t *testing.T, user *domain.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)
			},
		},
		{
			name: "repository error",
			setupMock: func(m *mockUserRepository) {
				m.On("Create", mock.AnythingOfType("*domain.User")).
					Return(assert.AnError)
			},
			input: application.CreateUserInput{
				Username: "testuser",
			},
			expectErr: true,
			expectFunc: func(t *testing.T, user *domain.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(mockUserRepository)
			tt.setupMock(mockRepo)

			service := application.NewUserService(mockRepo)

			// Execute
			user, err := service.CreateUser(tt.input)

			// Verify
			if tt.expectFunc != nil {
				tt.expectFunc(t, user, err)
			}

			// Assert that all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUser(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mockUserRepository, int)
		userID     int
		expectErr  bool
		expectFunc func(*testing.T, *domain.User, error)
	}{
		{
			name: "user found",
			setupMock: func(m *mockUserRepository, id int) {
				m.On("GetByID", id).
					Return(&domain.User{
						ID:        id,
						Username:  "testuser",
						CreatedAt: time.Now().UTC(),
						UpdatedAt: time.Now().UTC(),
					}, nil)
			},
			userID:    1,
			expectErr: false,
			expectFunc: func(t *testing.T, user *domain.User, err error) {
				assert.NoError(t, err)
				require.NotNil(t, user)
				assert.Equal(t, 1, user.ID)
				assert.Equal(t, "testuser", user.Username)
			},
		},
		{
			name: "user not found",
			setupMock: func(m *mockUserRepository, id int) {
				m.On("GetByID", id).
					Return((*domain.User)(nil), assert.AnError)
			},
			userID:    999,
			expectErr: true,
			expectFunc: func(t *testing.T, user *domain.User, err error) {
				assert.Error(t, err)
				assert.Nil(t, user)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(mockUserRepository)
			tt.setupMock(mockRepo, tt.userID)

			service := application.NewUserService(mockRepo)

			// Execute
			user, err := service.GetUser(tt.userID)

			// Verify
			if tt.expectFunc != nil {
				tt.expectFunc(t, user, err)
			}

			// Assert that all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UserExists(t *testing.T) {
	tests := []struct {
		name       string
		setupMock  func(*mockUserRepository, int)
		userID     int
		expectErr  bool
		expectFunc func(*testing.T, bool, error)
	}{
		{
			name: "user exists",
			setupMock: func(m *mockUserRepository, id int) {
				m.On("Exists", id).Return(true, nil)
			},
			userID:    1,
			expectErr: false,
			expectFunc: func(t *testing.T, exists bool, err error) {
				assert.NoError(t, err)
				assert.True(t, exists)
			},
		},
		{
			name: "user does not exist",
			setupMock: func(m *mockUserRepository, id int) {
				m.On("Exists", id).Return(false, nil)
			},
			userID:    999,
			expectErr: false,
			expectFunc: func(t *testing.T, exists bool, err error) {
				assert.NoError(t, err)
				assert.False(t, exists)
			},
		},
		{
			name: "repository error",
			setupMock: func(m *mockUserRepository, id int) {
				m.On("Exists", id).Return(false, assert.AnError)
			},
			userID:    1,
			expectErr: true,
			expectFunc: func(t *testing.T, exists bool, err error) {
				assert.Error(t, err)
				assert.False(t, exists)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockRepo := new(mockUserRepository)
			tt.setupMock(mockRepo, tt.userID)

			service := application.NewUserService(mockRepo)

			// Execute
			exists, err := service.UserExists(tt.userID)

			// Verify
			if tt.expectFunc != nil {
				tt.expectFunc(t, exists, err)
			}

			// Assert that all expectations were met
			mockRepo.AssertExpectations(t)
		})
	}
}
