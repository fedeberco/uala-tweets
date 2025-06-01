package consumers

import (
	"context"
	"encoding/json"
	"fmt"
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
	for {
		m, err := c.reader.ReadMessage(ctx)
		if err != nil {
			return fmt.Errorf("error reading from kafka: %w", err)
		}
		var tweet domain.Tweet
		if err := json.Unmarshal(m.Value, &tweet); err != nil {
			// Optionally log and skip
			continue
		}
		if err := c.tweetRepo.Create(&tweet); err != nil {
			// Optionally log error
			continue
		}

		// After successful persistence, publish one fan-out event per user (author + followers)
		followers, err := c.followRepo.GetFollowers(int(tweet.UserID))
		if err != nil {
			// Optionally log error
			followers = []int{}
		}
		userIDs := append([]int{int(tweet.UserID)}, followers...)
		for _, userID := range userIDs {
			func(uid int) {
				event := &domain.TimelineFanoutEvent{
					TweetID: tweet.ID,
					UserID:  uid,
				}
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				_ = c.fanoutPub.PublishFanoutEvent(ctx, event)
			}(userID)
		}
	}
}

func (c *KafkaTweetConsumer) Close() error {
	return c.reader.Close()
}
