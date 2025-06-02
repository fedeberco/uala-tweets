package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"uala-tweets/internal/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaTweetPublisher struct {
	writer *kafka.Writer
}

func NewKafkaTweetPublisher(writer *kafka.Writer) *KafkaTweetPublisher {
	return &KafkaTweetPublisher{
		writer: writer,
	}
}

func (p *KafkaTweetPublisher) Publish(ctx context.Context, tweet *domain.Tweet) error {
	data, err := json.Marshal(tweet)
	if err != nil {
		return fmt.Errorf("failed to marshal tweet: %w", err)
	}

	msg := kafka.Message{
		Key:   fmt.Appendf(nil, "tweet_%d_%d", tweet.UserID, tweet.ID),
		Value: data,
	}

	log.Printf("Publishing tweet %d for user %d to topic %s", tweet.ID, tweet.UserID, p.writer.Topic)

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Failed to publish tweet %d: %v", tweet.ID, err)
		return fmt.Errorf("failed to publish message: %w", err)
	}

	log.Printf("Successfully published tweet %d for user %d", tweet.ID, tweet.UserID)
	return nil
}

func (p *KafkaTweetPublisher) Close() error {
	return p.writer.Close()
}
