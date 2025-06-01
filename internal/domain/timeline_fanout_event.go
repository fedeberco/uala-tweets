package domain

type TimelineFanoutEvent struct {
    TweetID int64   `json:"tweet_id"`
    UserID  int     `json:"user_id"`
}
