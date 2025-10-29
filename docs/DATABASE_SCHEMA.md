# Database Design

## Schema (PostgreSQL 15)
```sql
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    email CITEXT UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE chats(
    id BIGSERIAL PRIMARY KEY,
    type SMALLINT CHECK (type IN (1,2)), -- 1=direct 2=group
    created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE chat_members(
    chat_id BIGINT REFERENCES chats(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    last_read_msg_id BIGINT DEFAULT 0,
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
    status SMALLINT CHECK (status IN (1,2,3)), -- 1=sent 2=delivered 3=read
    ts TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (msg_id, user_id)
);
```

## Indexes
```sql
CREATE INDEX idx_messages_chat_created ON messages(chat_id, created_at DESC);
CREATE INDEX idx_receipts_msg        ON receipts(msg_id);
CREATE INDEX idx_chat_members_user   ON chat_members(user_id);
```

## Migration Strategy
- Local/dev: GORM `AutoMigrate` (safe additive)
- Staging/Prod: `goose` SQL files checked into `db/migrations/`

## Migration Example (goose)
File: `db/migrations/001_init.up.sql`
```sql
-- +goose Up
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE users(
    id BIGSERIAL PRIMARY KEY,
    email CITEXT UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);

-- +goose Down
DROP TABLE users;
```
Run:  
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
goose postgres "$DSN" up
```

## Common Queries
```sql
-- Paginate history
SELECT id, user_id, body, created_at
FROM messages
WHERE chat_id = $1
ORDER BY id DESC
LIMIT 50;

-- Unread count for user
SELECT COUNT(*) FROM messages m
JOIN chat_members cm ON cm.chat_id = m.chat_id
WHERE cm.user_id = $1
  AND m.id > cm.last_read_msg_id;
```
