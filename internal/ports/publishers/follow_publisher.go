package publishers

import "uala-tweets/internal/domain"

type FollowPublisher interface {
	PublishFollowEvent(event domain.FollowEvent) error
}
