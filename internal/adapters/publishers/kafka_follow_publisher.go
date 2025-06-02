package publishers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"uala-tweets/internal/domain"

	"github.com/segmentio/kafka-go"
)

type KafkaFollowPublisher struct {
	writer *kafka.Writer
}

func NewKafkaFollowPublisher(writer *kafka.Writer) *KafkaFollowPublisher {
	return &KafkaFollowPublisher{
		writer: writer,
	}
}

func (p *KafkaFollowPublisher) PublishFollowEvent(event domain.FollowEvent) error {
	ctx := context.Background()
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling follow event: %v", err)
		return fmt.Errorf("error marshaling follow event: %w", err)
	}

	msg := kafka.Message{
		Key:   fmt.Appendf(nil, "follow_%d_%d_%v", event.FollowerID, event.FollowedID, event.Following),
		Value: data,
	}

	log.Printf("Publishing follow event: %+v", event)

	if err := p.writer.WriteMessages(ctx, msg); err != nil {
		log.Printf("Error publishing follow event: %v", err)
		return fmt.Errorf("error publishing follow event: %w", err)
	}

	log.Printf("Successfully published follow event: %+v", event)
	return nil
}
