package publishers

import (
	"context"
	"uala-tweets/internal/domain"
)

type TweetPublisher interface {
	Publish(ctx context.Context, tweet *domain.Tweet) error
}
