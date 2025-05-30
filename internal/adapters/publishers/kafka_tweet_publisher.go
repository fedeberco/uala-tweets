package publishers

import (
	"context"
	"encoding/json"
	"fmt"
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
		Key:   fmt.Appendf(nil, "%d", tweet.UserID),
		Value: data,
	}

	return p.writer.WriteMessages(ctx, msg)
}

func (p *KafkaTweetPublisher) Close() error {
	return p.writer.Close()
}
