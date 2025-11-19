-- +goose Up
ALTER TABLE messages ADD COLUMN media_url TEXT;
ALTER TABLE messages ADD COLUMN reply_to_id BIGINT REFERENCES messages(id);
ALTER TABLE messages ADD COLUMN reactions JSONB DEFAULT '{}';

-- +goose Down
ALTER TABLE messages DROP COLUMN reactions;
ALTER TABLE messages DROP COLUMN reply_to_id;
ALTER TABLE messages DROP COLUMN media_url;
