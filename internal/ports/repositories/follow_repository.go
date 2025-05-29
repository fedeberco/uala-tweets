package repositories

type FollowRepository interface {
	Follow(followerID, followedID int) error
	Unfollow(followerID, followedID int) error
	IsFollowing(followerID, followedID int) (bool, error)
}
