-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    email CITEXT UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE chats(
    id BIGSERIAL PRIMARY KEY,
    type SMALLINT CHECK (type IN (1,2)) NOT NULL, -- 1=direct 2=group
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE chat_members(
    chat_id BIGINT REFERENCES chats(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    last_read_msg_id BIGINT DEFAULT 0,
    joined_at TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (chat_id, user_id)
);

CREATE TABLE messages(
    id BIGSERIAL PRIMARY KEY,
    chat_id BIGINT REFERENCES chats(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id),
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE receipts(
    msg_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    status SMALLINT CHECK (status IN (1,2,3)) NOT NULL, -- 1=sent 2=delivered 3=read
    ts TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (msg_id, user_id)
);

-- Indexes for performance
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
CREATE INDEX idx_receipts_msg ON receipts(msg_id);
CREATE INDEX idx_chat_members_user ON chat_members(user_id);
CREATE INDEX idx_users_email ON users(email);
-- +goose Down
DROP TABLE IF EXISTS receipts;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS chat_members;
DROP TABLE IF EXISTS chats;
DROP TABLE IF EXISTS users;
DROP EXTENSION IF EXISTS citext;
