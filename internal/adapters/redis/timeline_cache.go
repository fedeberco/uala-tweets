package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type TimelineCacheRedis struct {
	client *redis.Client
}

func NewTimelineCacheRedis(client *redis.Client) *TimelineCacheRedis {
	return &TimelineCacheRedis{client: client}
}

func timelineKey(userID int) string {
	return fmt.Sprintf("timeline:%d", userID)
}

func (r *TimelineCacheRedis) AddToTimeline(userID int, tweetID int64) error {
	ctx := context.Background()
	key := timelineKey(userID)
	return r.client.LPush(ctx, key, tweetID).Err()
}

func (r *TimelineCacheRedis) GetTimeline(userID int, limit int) ([]int64, error) {
	ctx := context.Background()
	key := timelineKey(userID)
	values, err := r.client.LRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}
	result := make([]int64, 0, len(values))
	for _, v := range values {
		var id int64
		_, err := fmt.Sscan(v, &id)
		if err != nil {
			return nil, err
		}
		result = append(result, id)
	}
	return result, nil
}

func (r *TimelineCacheRedis) ClearTimeline(userID int) error {
	ctx := context.Background()
	key := timelineKey(userID)
	return r.client.Del(ctx, key).Err()
}
