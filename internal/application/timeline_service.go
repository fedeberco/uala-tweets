package application

import (
	"uala-tweets/internal/ports/repositories"
)

type TimelineService struct {
	cache repositories.TimelineCache
}

func NewTimelineService(cache repositories.TimelineCache) *TimelineService {
	return &TimelineService{cache: cache}
}

func (s *TimelineService) AddTweet(userID int, tweetID int64) error {
	return s.cache.AddToTimeline(userID, tweetID)
}

func (s *TimelineService) GetTimeline(userID int, limit int) ([]int64, error) {
	return s.cache.GetTimeline(userID, limit)
}

func (s *TimelineService) ClearTimeline(userID int) error {
	return s.cache.ClearTimeline(userID)
}
