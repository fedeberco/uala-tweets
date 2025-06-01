package domain

type FollowEvent struct {
	FollowerID int  `json:"follower_id"`
	FollowedID int  `json:"followed_id"`
	Following  bool `json:"following"` // true for follow, false for unfollow
}

func (e *FollowEvent) TopicName() string {
	return "user.follow"
}
