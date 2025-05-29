package domain

import "time"

type Follow struct {
	FollowerID int       `json:"follower_id"`
	FollowedID int       `json:"followed_id"`
	CreatedAt  time.Time `json:"created_at"`
}
