package consumers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"uala-tweets/internal/domain"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTimelineCache is a mock implementation of the TimelineCache interface.
type MockTimelineCache struct {
	mock.Mock
}

func (m *MockTimelineCache) AddToTimeline(userID int, tweetID int64) error {
	args := m.Called(userID, tweetID)
	return args.Error(0)
}

func (m *MockTimelineCache) ClearTimeline(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockTimelineCache) GetTimeline(userID int, limit int) ([]int64, error) {
	args := m.Called(userID, limit)
	if timeline, ok := args.Get(0).([]int64); ok {
		return timeline, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockTimelineCache) RemoveFromTimeline(userID int, tweetID int64) error {
	args := m.Called(userID, tweetID)
	return args.Error(0)
}

func TestKafkaTimelineFanoutConsumer_Start(t *testing.T) {
	testCases := []struct {
		name       string
		msgValue   []byte
		setupMock  func(m *MockTimelineCache)
		assertions func(t *testing.T, cache *MockTimelineCache)
	}{
		{
			name: "successfully adds to timeline",
			msgValue: func() []byte {
				b, _ := json.Marshal(&domain.TimelineFanoutEvent{TweetID: 1, UserID: 42})
				return b
			}(),
			setupMock: func(m *MockTimelineCache) {
				m.On("AddToTimeline", 42, int64(1)).Return(nil)
			},
			assertions: func(t *testing.T, cache *MockTimelineCache) {
				cache.AssertCalled(t, "AddToTimeline", 42, int64(1))
			},
		},
		{
			name:      "invalid JSON does not add to timeline",
			msgValue:  []byte("not json"),
			setupMock: func(m *MockTimelineCache) {},
			assertions: func(t *testing.T, cache *MockTimelineCache) {
				cache.AssertNotCalled(t, "AddToTimeline", mock.Anything, mock.Anything)
			},
		},
		{
			name: "invalid event (UserID == 0)",
			msgValue: func() []byte {
				b, _ := json.Marshal(&domain.TimelineFanoutEvent{TweetID: 1, UserID: 0})
				return b
			}(),
			setupMock: func(m *MockTimelineCache) {},
			assertions: func(t *testing.T, cache *MockTimelineCache) {
				cache.AssertNotCalled(t, "AddToTimeline", mock.Anything, mock.Anything)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockReader := NewMockKafkaReader(kafka.Message{Value: tc.msgValue})
			mockCache := new(MockTimelineCache)
			tc.setupMock(mockCache)

			consumer := NewKafkaTimelineFanoutConsumer(mockReader, mockCache, &MockFollowRepository{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			errCh := make(chan error, 1)
			go func() {
				errCh <- consumer.Start(ctx)
			}()

			// Wait for the consumer to process the message
			mockReader.WaitForRead()

			// Add a small delay to ensure the message is processed
			time.Sleep(10 * time.Millisecond)

			// Verify the assertions
			tc.assertions(t, mockCache)

			// Cancel the context to stop the consumer
			cancel()

			// Check for errors with a timeout
			select {
			case err := <-errCh:
				assert.NoError(t, err)
			case <-time.After(100 * time.Millisecond):
				t.Fatal("Timed out waiting for consumer to stop")
			}
		})
	}
}
