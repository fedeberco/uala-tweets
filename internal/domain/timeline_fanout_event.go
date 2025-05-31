package domain

type TimelineFanoutEvent struct {
    TweetID int64   `json:"tweet_id"`
    UserIDs []int   `json:"user_ids"`
}
