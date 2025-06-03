package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
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
	log.Printf("Starting TimelineFanout consumer...")
	defer log.Printf("TimelineFanout consumer stopped")

	for {
		select {
		case <-ctx.Done():
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Printf("Context canceled, stopping consumer")
				return nil
			}
			log.Printf("Context error, stopping consumer: %v", ctx.Err())
			return ctx.Err()
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					log.Printf("Context canceled, stopping consumer")
					return nil
				}
				log.Printf("Error reading message from Kafka: %v", err)
				continue
			}

			log.Printf("Received message - Topic: %s, Partition: %d, Offset: %d", m.Topic, m.Partition, m.Offset)

			var event domain.TimelineFanoutEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("Error unmarshaling message: %v, Message: %s", err, string(m.Value))
				continue
			}

			if event.UserID == 0 {
				log.Printf("Received invalid event with UserID 0")
				continue
			}

			log.Printf("Processing fanout event - UserID: %d, TweetID: %d", event.UserID, event.TweetID)

			if err := c.timelineCache.AddToTimeline(event.UserID, event.TweetID); err != nil {
				log.Printf("Error adding to timeline - UserID: %d, TweetID: %d, Error: %v", event.UserID, event.TweetID, err)
				continue
			}

			log.Printf("Successfully processed fanout event - UserID: %d, TweetID: %d", event.UserID, event.TweetID)
		}
	}
}

func (c *KafkaTimelineFanoutConsumer) Close() error {
	return c.reader.Close()
}
