package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"uala-tweets/internal/domain"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)


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
			name: "invalid JSON does not add to timeline",
			msgValue: []byte("not json"),
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
		{
			name: "AddToTimeline returns error",
			msgValue: func() []byte {
				b, _ := json.Marshal(&domain.TimelineFanoutEvent{TweetID: 2, UserID: 99})
				return b
			}(),
			setupMock: func(m *MockTimelineCache) {
				m.On("AddToTimeline", 99, int64(2)).Return(errors.New("cache error"))
			},
			assertions: func(t *testing.T, cache *MockTimelineCache) {
				cache.AssertCalled(t, "AddToTimeline", 99, int64(2))
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockReader := &MockKafkaReader{msg: kafka.Message{Value: tc.msgValue}}
			mockCache := new(MockTimelineCache)
			tc.setupMock(mockCache)

			consumer := NewKafkaTimelineFanoutConsumer(mockReader, mockCache, &MockFollowRepository{})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			err := consumer.Start(ctx)
			assert.Error(t, err)
			tc.assertions(t, mockCache)
		})
	}
}
