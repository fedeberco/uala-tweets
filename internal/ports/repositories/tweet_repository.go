package repositories

import "uala-tweets/internal/domain"

type TweetRepository interface {
	Create(tweet *domain.Tweet) error
	GetByID(id int64) (*domain.Tweet, error)
	GetByUserID(userID int64) ([]*domain.Tweet, error)
	GetTweetIDsByUser(userID int) ([]int64, error)
}
