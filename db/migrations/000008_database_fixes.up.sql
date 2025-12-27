-- Migration 008: Database fixes for production readiness
-- Fixes: duplicate reactions storage, missing indexes, missing constraints

-- 1. Remove duplicate JSONB reactions column (using reactions table instead)
ALTER TABLE messages DROP COLUMN IF EXISTS reactions;

-- 2. Add username uniqueness constraint
-- First, handle any existing duplicates by appending ID
DO $$
DECLARE
    dup RECORD;
BEGIN
    FOR dup IN
        SELECT id, username, ROW_NUMBER() OVER (PARTITION BY username ORDER BY id) as rn
        FROM users
        WHERE username IS NOT NULL
    LOOP
        IF dup.rn > 1 THEN
            UPDATE users SET username = dup.username || '_' || dup.id WHERE id = dup.id;
        END IF;
    END LOOP;
END $$;

ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);

-- 3. Add composite index for message pagination (chat_id + created_at DESC)
CREATE INDEX IF NOT EXISTS idx_messages_chat_created ON messages(chat_id, created_at DESC);

-- 4. Add index for user's chats lookup
CREATE INDEX IF NOT EXISTS idx_chat_members_user_id ON chat_members(user_id);

-- 5. Add index for messages by user (moderation, profile views)
CREATE INDEX IF NOT EXISTS idx_messages_user_id ON messages(user_id);

-- 6. Add check constraint for chat type (1=direct, 2=group)
ALTER TABLE chats ADD CONSTRAINT chats_type_check CHECK (type IN (1, 2));

-- 7. Add check constraint for role
ALTER TABLE chat_members ADD CONSTRAINT chat_members_role_check
    CHECK (role IN ('owner', 'admin', 'member'));

-- 8. Fix reply_to_id deletion behavior (SET NULL instead of blocking delete)
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_reply_to_id_fkey;
ALTER TABLE messages ADD CONSTRAINT messages_reply_to_id_fkey
    FOREIGN KEY (reply_to_id) REFERENCES messages(id) ON DELETE SET NULL;

-- 9. Add created_at to device_tokens for tracking first registration
ALTER TABLE device_tokens ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;
