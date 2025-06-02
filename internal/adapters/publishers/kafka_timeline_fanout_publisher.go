package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"uala-tweets/internal/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaTimelineFanoutPublisher struct {
	writer *kafka.Writer
}

func NewKafkaTimelineFanoutPublisher(writer *kafka.Writer) *KafkaTimelineFanoutPublisher {
	return &KafkaTimelineFanoutPublisher{writer: writer}
}

func (p *KafkaTimelineFanoutPublisher) PublishFanoutEvent(ctx context.Context, event *domain.TimelineFanoutEvent) error {

	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Failed to marshal fanout event (TweetID: %d, UserID: %d): %v", event.TweetID, event.UserID, err)
		return fmt.Errorf("failed to marshal fanout event: %w", err)
	}

	msg := kafka.Message{
		Key:   fmt.Appendf(nil, "fanout_%d_%d", event.UserID, event.TweetID),
		Value: data,
	}

	log.Printf("Publishing fanout event - TweetID: %d, UserID: %d, Topic: %s",
		event.TweetID, event.UserID, p.writer.Topic)

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Failed to publish fanout event (TweetID: %d, UserID: %d): %v",
			event.TweetID, event.UserID, err)
		return fmt.Errorf("failed to publish fanout event: %w", err)
	}

	log.Printf("Successfully published fanout event - TweetID: %d, UserID: %d",
		event.TweetID, event.UserID)
	return nil
}

func (p *KafkaTimelineFanoutPublisher) Close() error {
	if err := p.writer.Close(); err != nil {
		log.Printf("Error closing Kafka writer for topic %s: %v", p.writer.Topic, err)
		return fmt.Errorf("error closing kafka writer: %w", err)
	}
	return nil
}
