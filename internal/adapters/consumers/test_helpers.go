package consumers

import (
	"context"
	"sync"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type MockKafkaReader struct {
	mock.Mock
	msg        kafka.Message
	once       sync.Once
	readCalled chan struct{}
	readCount  int
}

type MockFollowRepository struct{}

func (m *MockFollowRepository) Follow(followerID, followedID int) error   { return nil }
func (m *MockFollowRepository) Unfollow(followerID, followedID int) error { return nil }
func (m *MockFollowRepository) IsFollowing(followerID, followedID int) (bool, error) {
	return false, nil
}
func (m *MockFollowRepository) GetFollowers(userID int) ([]int, error) { return []int{}, nil }

func NewMockKafkaReader(msg kafka.Message) *MockKafkaReader {
	return &MockKafkaReader{
		msg:        msg,
		readCalled: make(chan struct{}, 1),
	}
}

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	m.readCount++
	if m.readCount == 1 {
		// First call - return the message
		if m.readCalled != nil {
			close(m.readCalled)
		}
		return m.msg, nil
	}

	// For subsequent calls, check if context is already done
	select {
	case <-ctx.Done():
		return kafka.Message{}, ctx.Err()
	default:
		// If context is not done, wait for it to be done
		<-ctx.Done()
		return kafka.Message{}, ctx.Err()
	}
}

func (m *MockKafkaReader) Close() error {
	return nil
}

// WaitForRead waits for the ReadMessage method to be called
func (m *MockKafkaReader) WaitForRead() {
	if m.readCalled != nil {
		<-m.readCalled
	}
}
