CREATE TABLE IF NOT EXISTS tweets (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT fk_tweets_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,

    CONSTRAINT chk_tweet_content_not_empty
        CHECK (content <> ''),
    CONSTRAINT chk_tweet_content_length
        CHECK (LENGTH(content) <= 280) -- Standard tweet length limit
);

-- Create index on user_id for faster lookups of a user's tweets
CREATE INDEX IF NOT EXISTS idx_tweets_user_id ON tweets(user_id);

-- Create index on created_at for chronological ordering
CREATE INDEX IF NOT EXISTS idx_tweets_created_at ON tweets(created_at);
