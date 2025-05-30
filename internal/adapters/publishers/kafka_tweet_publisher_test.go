package publishers_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/stretchr/testify/assert"

	"uala-tweets/internal/adapters/publishers"
	"uala-tweets/internal/domain"
)

func TestKafkaTweetPublisher_Publish(t *testing.T) {
	broker := os.Getenv("KAFKA_BROKER")
	if broker == "" {
		broker = "localhost:9092"
	}
	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "new_tweets"
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	pub := publishers.NewKafkaTweetPublisher(writer)
	defer pub.Close()

	tweet := &domain.Tweet{
		ID:        123,
		UserID:    1,
		Content:   "integration test tweet",
		CreatedAt: time.Now(),
	}

	ctx := context.Background()
	err := pub.Publish(ctx, tweet)
	assert.NoError(t, err)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		Topic:       topic,
		Partition:   0,
		MinBytes:    1,
		MaxBytes:    10e6,
		StartOffset: kafka.FirstOffset,
		GroupID:     "test-group",
	})
	defer reader.Close()

	found := false
	timeout := time.After(5 * time.Second)
	for !found {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for message from kafka")
		case <-time.After(100 * time.Millisecond):
			msg, err := reader.ReadMessage(ctx)
			if err != nil {
				continue // keep waiting
			}
			var received domain.Tweet
			unmarshalErr := json.Unmarshal(msg.Value, &received)
			if unmarshalErr == nil && received.ID == tweet.ID && received.Content == tweet.Content && received.UserID == tweet.UserID {
				found = true
				assert.WithinDuration(t, tweet.CreatedAt, received.CreatedAt, time.Second)
			}
		}
	}
}
