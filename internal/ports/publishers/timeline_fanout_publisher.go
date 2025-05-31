package publishers

import (
	"context"
	"uala-tweets/internal/domain"
)

type TimelineFanoutPublisher interface {
	PublishFanoutEvent(ctx context.Context, event *domain.TimelineFanoutEvent) error
}
