package repositories

type TimelineCache interface {
	AddToTimeline(userID int, tweetID int64) error
	GetTimeline(userID int, limit int) ([]int64, error)
	ClearTimeline(userID int) error
	RemoveFromTimeline(userID int, tweetID int64) error
}
