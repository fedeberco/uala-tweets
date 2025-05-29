-- Create follows table
CREATE TABLE IF NOT EXISTS follows (
    follower_id INTEGER NOT NULL,
    followed_id INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (follower_id, followed_id),
    CONSTRAINT fk_follower
        FOREIGN KEY (follower_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_followed
        FOREIGN KEY (followed_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    -- Prevent self-follows
    CONSTRAINT check_not_self_follow 
        CHECK (follower_id != followed_id)
);

-- Create indexes for faster lookups
CREATE INDEX IF NOT EXISTS idx_follows_follower_id ON follows (follower_id);
CREATE INDEX IF NOT EXISTS idx_follows_followed_id ON follows (followed_id);
