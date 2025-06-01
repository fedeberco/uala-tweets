package consumers

import (
	"context"
	"errors"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/mock"
)

type MockKafkaReader struct {
	mock.Mock
	msg       kafka.Message
	readCount int
}

type MockFollowRepository struct{}

func (m *MockFollowRepository) Follow(followerID, followedID int) error { return nil }
func (m *MockFollowRepository) Unfollow(followerID, followedID int) error { return nil }
func (m *MockFollowRepository) IsFollowing(followerID, followedID int) (bool, error) { return false, nil }
func (m *MockFollowRepository) GetFollowers(userID int) ([]int, error) { return []int{}, nil }

func (m *MockKafkaReader) ReadMessage(ctx context.Context) (kafka.Message, error) {
	m.readCount++
	if m.readCount == 1 {
		return m.msg, nil
	}
	return kafka.Message{}, errors.New("no more messages")
}

func (m *MockKafkaReader) Close() error { return nil }
