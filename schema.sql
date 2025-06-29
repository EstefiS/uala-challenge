DROP TABLE IF EXISTS timelines;
DROP TABLE IF EXISTS followers;
DROP TABLE IF EXISTS tweets;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE followers (
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    follower_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, follower_id)
);
CREATE INDEX idx_followers_user_id ON followers(user_id);

CREATE TABLE tweets (
    id VARCHAR(255) PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    text VARCHAR(280) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE timelines (
    user_id VARCHAR(255) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tweet_id VARCHAR(255) NOT NULL REFERENCES tweets(id) ON DELETE CASCADE,
    tweet_created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (user_id, tweet_id)
);
CREATE INDEX idx_timelines_user_created_at ON timelines(user_id, tweet_created_at DESC);