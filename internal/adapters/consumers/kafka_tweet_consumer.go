package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"time"
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/publishers"
	"uala-tweets/internal/ports/repositories"

	"github.com/segmentio/kafka-go"
)

type KafkaReader interface {
	ReadMessage(ctx context.Context) (kafka.Message, error)
	Close() error
}

type KafkaTweetConsumer struct {
	reader     KafkaReader
	tweetRepo  repositories.TweetRepository
	fanoutPub  publishers.TimelineFanoutPublisher
	followRepo repositories.FollowRepository
}

func NewKafkaTweetConsumer(reader KafkaReader, tweetRepo repositories.TweetRepository, fanoutPub publishers.TimelineFanoutPublisher, followRepo repositories.FollowRepository) *KafkaTweetConsumer {
	return &KafkaTweetConsumer{
		reader:     reader,
		tweetRepo:  tweetRepo,
		fanoutPub:  fanoutPub,
		followRepo: followRepo,
	}
}

// Start starts the consumer loop. It should be run as a goroutine.
func (c *KafkaTweetConsumer) Start(ctx context.Context) error {
	log.Println("Starting Tweet consumer...")
	defer log.Println("Tweet consumer stopped")

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

			log.Printf("Received tweet message - Topic: %s, Partition: %d, Offset: %d",
				m.Topic, m.Partition, m.Offset)

			var tweet domain.Tweet
			if err := json.Unmarshal(m.Value, &tweet); err != nil {
				log.Printf("Error unmarshaling tweet: %v, Raw: %s", err, string(m.Value))
				continue
			}

			log.Printf("Processing new tweet - ID: %d, UserID: %d, Content: %.50s...",
				tweet.ID, tweet.UserID, tweet.Content)

			// Persist the tweet
			if err := c.tweetRepo.Create(&tweet); err != nil {
				log.Printf("Error persisting tweet %d: %v", tweet.ID, err)
				continue
			}

			log.Printf("Successfully persisted tweet %d, finding followers for user %d",
				tweet.ID, tweet.UserID)

			// After successful persistence, publish one fan-out event per user (author + followers)
			followers, err := c.followRepo.GetFollowers(int(tweet.UserID))
			if err != nil {
				log.Printf("Error getting followers for user %d: %v. Will only fanout to author.",
					tweet.UserID, err)
				followers = []int{}
			}

			userIDs := append([]int{int(tweet.UserID)}, followers...)
			log.Printf("Fanning out tweet %d to %d users (author + %d followers)",
				tweet.ID, len(userIDs), len(followers))

			for _, userID := range userIDs {
				func(uid int) {
					event := &domain.TimelineFanoutEvent{
						TweetID: tweet.ID,
						UserID:  uid,
					}

					log.Printf("Publishing fanout event - TweetID: %d, UserID: %d",
						event.TweetID, event.UserID)

					fanoutCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()

					if err := c.fanoutPub.PublishFanoutEvent(fanoutCtx, event); err != nil {
						log.Printf("Error publishing fanout event for tweet %d to user %d: %v",
							event.TweetID, event.UserID, err)
					} else {
						log.Printf("Successfully published fanout event for tweet %d to user %d",
							event.TweetID, event.UserID)
					}
				}(userID)
			}

			log.Printf("Completed processing tweet %d", tweet.ID)
		}
	}
}

func (c *KafkaTweetConsumer) Close() error {
	return c.reader.Close()
}
