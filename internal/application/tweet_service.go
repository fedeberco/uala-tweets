package application

import (
	"context"
	"errors"
	"time"
	"uala-tweets/internal/domain"
	"uala-tweets/internal/ports/publishers"
	"uala-tweets/internal/ports/repositories"
)

var (
	ErrTweetContentEmpty   = errors.New("tweet content cannot be empty")
	ErrTweetContentTooLong = errors.New("tweet content is too long (max 280 characters)")
)

type TweetService struct {
	tweetRepo repositories.TweetRepository
	tweetPub  publishers.TweetPublisher
}

func NewTweetService(tweetRepo repositories.TweetRepository, tweetPub publishers.TweetPublisher) *TweetService {
	return &TweetService{
		tweetRepo: tweetRepo,
		tweetPub:  tweetPub,
	}
}

type CreateTweetInput struct {
	UserID  int64
	Content string
}

func (s *TweetService) CreateTweet(ctx context.Context, input CreateTweetInput) (*domain.Tweet, error) {
	if input.Content == "" {
		return nil, ErrTweetContentEmpty
	}

	if len(input.Content) > 280 {
		return nil, ErrTweetContentTooLong
	}

	tweet := &domain.Tweet{
		UserID:    input.UserID,
		Content:   input.Content,
		CreatedAt: time.Now(),
	}

	// Publish tweet asynchronously (event-driven persistence)
	go func(tweet *domain.Tweet) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = s.tweetPub.Publish(ctx, tweet)
	}(tweet)

	return tweet, nil
}

func (s *TweetService) GetTweet(ctx context.Context, id int64) (*domain.Tweet, error) {
	return s.tweetRepo.GetByID(id)
}

func (s *TweetService) GetUserTweets(ctx context.Context, userID int64) ([]*domain.Tweet, error) {
	return s.tweetRepo.GetByUserID(userID)
}
