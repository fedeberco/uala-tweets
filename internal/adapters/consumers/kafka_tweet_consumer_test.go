package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"uala-tweets/internal/domain"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


// MockFollowRepository mocks the FollowRepository interface for injection in tests.

// MockTimelineFanoutPublisher mocks the TimelineFanoutPublisher interface
// for injection into KafkaTweetConsumer during tests.
type MockTimelineFanoutPublisher struct{}

func (m *MockTimelineFanoutPublisher) PublishFanoutEvent(ctx context.Context, event *domain.TimelineFanoutEvent) error {
	return nil
}

// MockTweetRepository mocks the TweetRepository interface
type MockTweetRepository struct {
	mock.Mock
}

func (m *MockTweetRepository) Create(tweet *domain.Tweet) error {
	args := m.Called(tweet)
	return args.Error(0)
}

func (m *MockTweetRepository) GetByID(id int64) (*domain.Tweet, error) {
	args := m.Called(id)
	if tw, ok := args.Get(0).(*domain.Tweet); ok {
		return tw, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTweetRepository) GetByUserID(userID int64) ([]*domain.Tweet, error) {
	args := m.Called(userID)
	if tweets, ok := args.Get(0).([]*domain.Tweet); ok {
		return tweets, args.Error(1)
	}
	return nil, args.Error(1)
}


func (m *MockTweetRepository) GetTweetIDsByUser(userID int) ([]int64, error) {
	args := m.Called(userID)
	if tweetIDs, ok := args.Get(0).([]int64); ok {
		return tweetIDs, args.Error(1)
	}
	return nil, args.Error(1)
}

func TestKafkaTweetConsumer_Start(t *testing.T) {
	testCases := []struct {
		name          string
		msgValue      []byte
		setupRepoMock func(m *MockTweetRepository)
		assertions    func(t *testing.T, repo *MockTweetRepository)
	}{
		{
			name:     "successfully consumes and stores tweet",
			msgValue: func() []byte { b, _ := json.Marshal(&domain.Tweet{ID: 1, UserID: 42, Content: "hello"}); return b }(),
			setupRepoMock: func(m *MockTweetRepository) {
				m.On("Create", mock.AnythingOfType("*domain.Tweet")).Return(nil)
			},
			assertions: func(t *testing.T, repo *MockTweetRepository) {
				repo.AssertCalled(t, "Create", mock.MatchedBy(func(tw *domain.Tweet) bool {
					return tw.Content == "hello" && tw.UserID == 42
				}))
			},
		},
		{
			name:          "invalid JSON does not call Create",
			msgValue:      []byte("not json"),
			setupRepoMock: func(m *MockTweetRepository) {},
			assertions: func(t *testing.T, repo *MockTweetRepository) {
				repo.AssertNotCalled(t, "Create", mock.Anything)
			},
		},
		{
			name:     "repo returns error",
			msgValue: func() []byte { b, _ := json.Marshal(&domain.Tweet{ID: 1, UserID: 42, Content: "fail"}); return b }(),
			setupRepoMock: func(m *MockTweetRepository) {
				m.On("Create", mock.AnythingOfType("*domain.Tweet")).Return(errors.New("db error"))
			},
			assertions: func(t *testing.T, repo *MockTweetRepository) {
				repo.AssertCalled(t, "Create", mock.AnythingOfType("*domain.Tweet"))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockReader := NewMockKafkaReader(kafka.Message{Value: tc.msgValue})
			mockRepo := new(MockTweetRepository)
			tc.setupRepoMock(mockRepo)

			consumer := NewKafkaTweetConsumer(mockReader, mockRepo, &MockTimelineFanoutPublisher{}, &MockFollowRepository{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Start the consumer in a goroutine
			errCh := make(chan error, 1)
			go func() {
				errCh <- consumer.Start(ctx)
			}()

			// Wait for the consumer to process the message
			mockReader.WaitForRead()

			// Give some time for the consumer to process the message
			time.Sleep(10 * time.Millisecond)

			// Verify the assertions
			tc.assertions(t, mockRepo)

			// Cancel the context to stop the consumer
			cancel()


			// Check for errors with a timeout
			select {
			case err := <-errCh:
				if tc.name == "repo returns error" {
					assert.Error(t, err)
				} else {
					assert.ErrorIs(t, err, context.Canceled)
				}
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timed out waiting for consumer to stop")
			}
		})
	}
}
