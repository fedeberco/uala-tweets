package application

import (
	"errors"
	"testing"

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
func (m *MockTimelineCache) GetTimeline(userID int, limit int) ([]int64, error) {
	args := m.Called(userID, limit)
	if timeline, ok := args.Get(0).([]int64); ok {
		return timeline, args.Error(1)
	}
	return nil, args.Error(1)
}
func (m *MockTimelineCache) ClearTimeline(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestTimelineService_AddTweet(t *testing.T) {
	mockCache := new(MockTimelineCache)
	service := NewTimelineService(mockCache)
	mockCache.On("AddToTimeline", 1, int64(42)).Return(nil)
	err := service.AddTweet(1, 42)
	assert.NoError(t, err)
	mockCache.AssertCalled(t, "AddToTimeline", 1, int64(42))
}

func TestTimelineService_GetTimeline(t *testing.T) {
	mockCache := new(MockTimelineCache)
	service := NewTimelineService(mockCache)
	mockCache.On("GetTimeline", 1, 10).Return([]int64{101, 102}, nil)
	timeline, err := service.GetTimeline(1, 10)
	assert.NoError(t, err)
	assert.Equal(t, []int64{101, 102}, timeline)
	mockCache.AssertCalled(t, "GetTimeline", 1, 10)
}

func TestTimelineService_ClearTimeline(t *testing.T) {
	mockCache := new(MockTimelineCache)
	service := NewTimelineService(mockCache)
	mockCache.On("ClearTimeline", 1).Return(nil)
	err := service.ClearTimeline(1)
	assert.NoError(t, err)
	mockCache.AssertCalled(t, "ClearTimeline", 1)
}

func TestTimelineService_Errors(t *testing.T) {
	mockCache := new(MockTimelineCache)
	service := NewTimelineService(mockCache)
	mockCache.On("AddToTimeline", 2, int64(43)).Return(errors.New("fail"))
	err := service.AddTweet(2, 43)
	assert.Error(t, err)

	mockCache.On("GetTimeline", 2, 5).Return(nil, errors.New("fail"))
	_, err = service.GetTimeline(2, 5)
	assert.Error(t, err)

	mockCache.On("ClearTimeline", 2).Return(errors.New("fail"))
	err = service.ClearTimeline(2)
	assert.Error(t, err)
}
