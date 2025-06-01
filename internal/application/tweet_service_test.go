package application_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"uala-tweets/internal/application"
	"uala-tweets/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTweetRepository struct {
	mock.Mock
}

func (m *MockTweetRepository) Create(tweet *domain.Tweet) error {
	args := m.Called(tweet)
	return args.Error(0)
}

func (m *MockTweetRepository) GetByID(id int64) (*domain.Tweet, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Tweet), args.Error(1)
}

func (m *MockTweetRepository) GetByUserID(userID int64) ([]*domain.Tweet, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Tweet), args.Error(1)
}


func (m *MockTweetRepository) GetTweetIDsByUser(userID int) ([]int64, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]int64), args.Error(1)
}

type MockTweetPublisher struct {
	mock.Mock
}

type MockTimelineFanoutPublisher struct {
	mock.Mock
}

func (m *MockTimelineFanoutPublisher) PublishFanoutEvent(ctx context.Context, event *domain.TimelineFanoutEvent) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func (m *MockTweetPublisher) Publish(ctx context.Context, tweet *domain.Tweet) error {
	args := m.Called(ctx, tweet)
	return args.Error(0)
}

func TestTweetService_CreateTweet(t *testing.T) {
	tests := []struct {
		name          string
		input         application.CreateTweetInput
		mockSetup     func(*MockTweetRepository, *MockTweetPublisher, *sync.WaitGroup)
		expectedError string
	}{
		{
			name: "successful tweet creation",
			input: application.CreateTweetInput{
				UserID:  1,
				Content: "Hello, world!",
			},
			mockSetup: func(repo *MockTweetRepository, pub *MockTweetPublisher, wg *sync.WaitGroup) {
				wg.Add(1)
				pub.On("Publish", mock.Anything, mock.AnythingOfType("*domain.Tweet")).
					Run(func(args mock.Arguments) { wg.Done() }).
					Return(nil)
			},
		},
		{
			name: "empty content",
			input: application.CreateTweetInput{
				UserID:  1,
				Content: "",
			},
			mockSetup: func(repo *MockTweetRepository, pub *MockTweetPublisher, wg *sync.WaitGroup) {
				// Prevent panic from async Publish
				pub.On("Publish", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: "tweet content cannot be empty",
		},
		{
			name: "content too long",
			input: application.CreateTweetInput{
				UserID: 1,
				Content: "This is a very long tweet that exceeds the maximum allowed length of 280 characters. " +
					"Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam auctor, nisl eget " +
					"ultricies tincidunt, nunc nisl aliquam nunc, vitae aliquam nisl nunc vitae nisl. " +
					"Sed vitae nisl eget nisl aliquam tincidunt. Nullam auctor, nisl eget ultricies tincidunt.",
			},
			mockSetup: func(repo *MockTweetRepository, pub *MockTweetPublisher, wg *sync.WaitGroup) {
				pub.On("Publish", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: "tweet content is too long",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTweetRepository)
			mockPub := new(MockTweetPublisher)
			var wg sync.WaitGroup

			tt.mockSetup(mockRepo, mockPub, &wg)

			service := application.NewTweetService(mockRepo, mockPub)

			tweet, err := service.CreateTweet(context.Background(), tt.input)

			done := make(chan struct{})
			go func() {
				wg.Wait()
				close(done)
			}()
			select {
			case <-done:
				// ok
			case <-time.After(1 * time.Second):
				t.Fatal("timed out waiting for Publish goroutine")
			}

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, tweet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tweet)
				assert.Equal(t, tt.input.UserID, tweet.UserID)
				assert.Equal(t, tt.input.Content, tweet.Content)
				assert.False(t, tweet.CreatedAt.IsZero())

				mockPub.AssertExpectations(t)
			}
		})
	}
}

func TestTweetService_GetTweet(t *testing.T) {
	tests := []struct {
		name          string
		tweetID       int64
		setupMock     func(*MockTweetRepository)
		expectedTweet *domain.Tweet
		expectedError string
	}{
		{
			name:    "successful get tweet",
			tweetID: 1,
			setupMock: func(repo *MockTweetRepository) {
				repo.On("GetByID", int64(1)).Return(
					&domain.Tweet{
						ID:        1,
						UserID:    1,
						Content:   "Test tweet",
						CreatedAt: time.Now(),
					},
					nil,
				)
			},
			expectedTweet: &domain.Tweet{
				ID:      1,
				UserID:  1,
				Content: "Test tweet",
			},
		},
		{
			name:    "tweet not found",
			tweetID: 999,
			setupMock: func(repo *MockTweetRepository) {
				repo.On("GetByID", int64(999)).Return(
					(*domain.Tweet)(nil),
					errors.New("tweet not found"),
				)
			},
			expectedError: "tweet not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTweetRepository)
			mockPub := new(MockTweetPublisher)

			tt.setupMock(mockRepo)

			service := application.NewTweetService(mockRepo, mockPub)
			tweet, err := service.GetTweet(context.Background(), tt.tweetID)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, tweet)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, tweet)
				assert.Equal(t, tt.expectedTweet.ID, tweet.ID)
				assert.Equal(t, tt.expectedTweet.UserID, tweet.UserID)
				assert.Equal(t, tt.expectedTweet.Content, tweet.Content)
			}

			mockRepo.AssertExpectations(t)
			mockPub.AssertExpectations(t)
		})
	}
}

func TestTweetService_GetUserTweets(t *testing.T) {
	tests := []struct {
		name           string
		userID         int64
		setupMock      func(*MockTweetRepository)
		expectedTweets []*domain.Tweet
		expectedError  string
	}{
		{
			name:   "successful get user tweets",
			userID: 1,
			setupMock: func(repo *MockTweetRepository) {
				repo.On("GetByUserID", int64(1)).Return(
					[]*domain.Tweet{
						{ID: 1, UserID: 1, Content: "First tweet"},
						{ID: 2, UserID: 1, Content: "Second tweet"},
					},
					nil,
				)
			},
			expectedTweets: []*domain.Tweet{
				{ID: 1, UserID: 1, Content: "First tweet"},
				{ID: 2, UserID: 1, Content: "Second tweet"},
			},
		},
		{
			name:   "user has no tweets",
			userID: 2,
			setupMock: func(repo *MockTweetRepository) {
				repo.On("GetByUserID", int64(2)).Return(
					[]*domain.Tweet{},
					nil,
				)
			},
			expectedTweets: []*domain.Tweet{},
		},
		{
			name:   "error from repository",
			userID: 3,
			setupMock: func(repo *MockTweetRepository) {
				repo.On("GetByUserID", int64(3)).Return(
					([]*domain.Tweet)(nil),
					errors.New("database error"),
				)
			},
			expectedError: "database error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTweetRepository)
			mockPub := new(MockTweetPublisher)

			tt.setupMock(mockRepo)

			service := application.NewTweetService(mockRepo, mockPub)

			tweets, err := service.GetUserTweets(context.Background(), tt.userID)

			if tt.expectedError != "" {
				assert.ErrorContains(t, err, tt.expectedError)
				assert.Nil(t, tweets)
			} else {
				assert.NoError(t, err)
				assert.Len(t, tweets, len(tt.expectedTweets))
				for i, tweet := range tweets {
					assert.Equal(t, tt.expectedTweets[i].ID, tweet.ID)
					assert.Equal(t, tt.expectedTweets[i].UserID, tweet.UserID)
					assert.Equal(t, tt.expectedTweets[i].Content, tweet.Content)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
