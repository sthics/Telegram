CREATE TABLE IF NOT EXISTS device_tokens (
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL,
    platform VARCHAR(50) NOT NULL, -- 'ios', 'android', 'web'
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (user_id, token)
);

CREATE INDEX idx_device_tokens_user_id ON device_tokens(user_id);
