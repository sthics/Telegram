-- +goose Up
ALTER TABLE chat_members ADD COLUMN role VARCHAR(20) DEFAULT 'member';

-- +goose Down
ALTER TABLE chat_members DROP COLUMN role;
