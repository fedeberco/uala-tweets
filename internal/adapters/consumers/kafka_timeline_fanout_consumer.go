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
		if len(event.UserIDs) == 0 {
			continue // invalid event
		}
		authorID := event.UserIDs[0]
		followers, err := c.followRepo.GetFollowers(authorID)
		if err != nil {
			// Optionally log error
			continue
		}
		targets := append([]int{authorID}, followers...)
		for _, userID := range targets {
			if err := c.timelineCache.AddToTimeline(userID, event.TweetID); err != nil {
				// Optionally log error
				continue
			}
		}
	}
}

func (c *KafkaTimelineFanoutConsumer) Close() error {
	return c.reader.Close()
}
