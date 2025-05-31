package publishers

import (
	"context"
	"encoding/json"
	"fmt"
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
		return fmt.Errorf("failed to marshal fanout event: %w", err)
	}

	msg := kafka.Message{
		Key:   []byte("fanout"),
		Value: data,
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaTimelineFanoutPublisher) Close() error {
	return p.writer.Close()
}
