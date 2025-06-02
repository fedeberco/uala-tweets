package consumers

import (
	"context"
	"encoding/json"
	"log"

	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/repositories"
)

type KafkaFollowConsumer struct {
	reader        KafkaReader
	timelineCache repositories.TimelineCache
	tweetRepo     repositories.TweetRepository
}

func NewKafkaFollowConsumer(reader KafkaReader, timelineCache repositories.TimelineCache, tweetRepo repositories.TweetRepository) *KafkaFollowConsumer {
	return &KafkaFollowConsumer{
		reader:        reader,
		timelineCache: timelineCache,
		tweetRepo:     tweetRepo,
	}
}

// Start starts the consumer loop. It should be run as a goroutine.
func (c *KafkaFollowConsumer) Start(ctx context.Context) error {
	log.Println("Starting Kafka follow consumer...")
	defer log.Println("Stopped Kafka follow consumer")

	for {
		select {
		case <-ctx.Done():
			log.Printf("Context done, stopping consumer: %v", ctx.Err())
			return ctx.Err()
		default:
			m, err := c.reader.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					log.Printf("Context error, stopping consumer: %v", ctx.Err())
					return ctx.Err()
				}
				log.Printf("Error reading from kafka: %v", err)
				continue
			}

			log.Printf("Received follow event - Topic: %s, Partition: %d, Offset: %d", m.Topic, m.Partition, m.Offset)

			var event domain.FollowEvent
			if err := json.Unmarshal(m.Value, &event); err != nil {
				log.Printf("Error unmarshaling follow event: %v, Raw: %s", err, string(m.Value))
				continue
			}

			log.Printf("Processing follow event - Type: %s, FollowerID: %d, FollowedID: %d",
				map[bool]string{true: "FOLLOW", false: "UNFOLLOW"}[event.Following],
				event.FollowerID,
				event.FollowedID)

			if event.FollowerID == 0 || event.FollowedID == 0 {
				log.Printf("Invalid follow event - missing IDs: %+v", event)
				continue
			}

			if event.Following {
				// On follow: Add followed user's tweets to follower's timeline
				tweetIDs, err := c.tweetRepo.GetTweetIDsByUser(event.FollowedID)
				if err != nil {
					log.Printf("Error getting tweet IDs for user %d: %v", event.FollowedID, err)
					continue
				}

				log.Printf("Adding %d tweets to user %d's timeline from user %d",
					len(tweetIDs), event.FollowerID, event.FollowedID)

				for _, tweetID := range tweetIDs {
					if err := c.timelineCache.AddToTimeline(event.FollowerID, tweetID); err != nil {
						log.Printf("Error adding tweet %d to timeline for user %d: %v",
							tweetID, event.FollowerID, err)
					} else {
						log.Printf("Successfully added tweet %d to user %d's timeline",
							tweetID, event.FollowerID)
					}
				}
			} else {
				// On unfollow: Remove followed user's tweets from follower's timeline
				tweetIDs, err := c.tweetRepo.GetTweetIDsByUser(event.FollowedID)
				if err != nil {
					log.Printf("Error getting tweet IDs for user %d: %v", event.FollowedID, err)
					continue
				}

				log.Printf("Removing %d tweets from user %d's timeline (unfollowing user %d)",
					len(tweetIDs), event.FollowerID, event.FollowedID)

				for _, tweetID := range tweetIDs {
					if err := c.timelineCache.RemoveFromTimeline(event.FollowerID, tweetID); err != nil {
						log.Printf("Error removing tweet %d from timeline for user %d: %v",
							tweetID, event.FollowerID, err)
					} else {
						log.Printf("Successfully removed tweet %d from user %d's timeline",
							tweetID, event.FollowerID)
					}
				}
			}

			log.Printf("Completed processing follow event - FollowerID: %d, FollowedID: %d",
				event.FollowerID, event.FollowedID)
		}
	}
}

func (c *KafkaFollowConsumer) Close() error {
	return c.reader.Close()
}
