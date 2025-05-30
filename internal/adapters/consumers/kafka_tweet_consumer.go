package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/repositories"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type KafkaTweetConsumer struct {
	reader   KafkaReader
	tweetRepo repositories.TweetRepository
}

func NewKafkaTweetConsumer(reader KafkaReader, tweetRepo repositories.TweetRepository) *KafkaTweetConsumer {
	return &KafkaTweetConsumer{
		reader:   reader,
		tweetRepo: tweetRepo,
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
	}
}

func (c *KafkaTweetConsumer) Close() error {
	return c.reader.Close()
}
