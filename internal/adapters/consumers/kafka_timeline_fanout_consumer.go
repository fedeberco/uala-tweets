package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/repositories"
)

type KafkaTimelineFanoutConsumer struct {
	reader        KafkaReader
	timelineCache repositories.TimelineCache
	followRepo    repositories.FollowRepository
}

func NewKafkaTimelineFanoutConsumer(reader KafkaReader, timelineCache repositories.TimelineCache, followRepo repositories.FollowRepository) *KafkaTimelineFanoutConsumer {
	return &KafkaTimelineFanoutConsumer{
		reader:        reader,
		timelineCache: timelineCache,
		followRepo:    followRepo,
	}
}

// Start starts the consumer loop. It should be run as a goroutine.
func (c *KafkaTimelineFanoutConsumer) Start(ctx context.Context) error {
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading from kafka: %w", err)
		}
		var event domain.TimelineFanoutEvent
		if err := json.Unmarshal(m.Value, &event); err != nil {
			// Optionally log and skip
			continue
		}
		if event.UserID == 0 {
			continue // invalid event
		}
		if err := c.timelineCache.AddToTimeline(event.UserID, event.TweetID); err != nil {
			// Optionally log error
			continue
		}
	}
}

func (c *KafkaTimelineFanoutConsumer) Close() error {
	return c.reader.Close()
}
