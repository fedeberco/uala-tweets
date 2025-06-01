package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/repositories"
	"uala-tweets/internal/ports/publishers"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type KafkaTweetConsumer struct {
	reader     KafkaReader
	tweetRepo  repositories.TweetRepository
	fanoutPub  publishers.TimelineFanoutPublisher
}

func NewKafkaTweetConsumer(reader KafkaReader, tweetRepo repositories.TweetRepository, fanoutPub publishers.TimelineFanoutPublisher) *KafkaTweetConsumer {
	return &KafkaTweetConsumer{
		reader:    reader,
		tweetRepo: tweetRepo,
		fanoutPub: fanoutPub,
	}
}

// Start starts the consumer loop. It should be run as a goroutine.
func (c *KafkaTweetConsumer) Start(ctx context.Context) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading from kafka: %w", err)
		}
		var tweet domain.Tweet
		if err := json.Unmarshal(m.Value, &tweet); err != nil {
			// Optionally log and skip
			continue
		}
		if err := c.tweetRepo.Create(&tweet); err != nil {
			// Optionally log error
			continue
		}

		// After successful persistence, publish fan-out event for timeline
		go func(tweetID int64, authorID int64) {
			event := &domain.TimelineFanoutEvent{
				TweetID: tweetID,
				UserIDs: []int{int(authorID)}, // Only the author for now; consumer will expand
			}
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = c.fanoutPub.PublishFanoutEvent(ctx, event)
		}(tweet.ID, tweet.UserID)
	}
}

func (c *KafkaTweetConsumer) Close() error {
	return c.reader.Close()
}
