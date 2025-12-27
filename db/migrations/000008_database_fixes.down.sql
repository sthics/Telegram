-- Rollback Migration 008: Database fixes

-- 9. Remove created_at from device_tokens
ALTER TABLE device_tokens DROP COLUMN IF EXISTS created_at;

-- 8. Restore original reply_to_id constraint (no ON DELETE clause)
ALTER TABLE messages DROP CONSTRAINT IF EXISTS messages_reply_to_id_fkey;
ALTER TABLE messages ADD CONSTRAINT messages_reply_to_id_fkey
    FOREIGN KEY (reply_to_id) REFERENCES messages(id);

-- 7. Remove role check constraint
ALTER TABLE chat_members DROP CONSTRAINT IF EXISTS chat_members_role_check;

-- 6. Remove chat type check constraint
ALTER TABLE chats DROP CONSTRAINT IF EXISTS chats_type_check;

-- 5. Remove messages user_id index
DROP INDEX IF EXISTS idx_messages_user_id;

-- 4. Remove chat_members user_id index
DROP INDEX IF EXISTS idx_chat_members_user_id;

-- 3. Remove message pagination index
DROP INDEX IF EXISTS idx_messages_chat_created;

-- 2. Remove username uniqueness constraint
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_username_unique;

-- 1. Restore JSONB reactions column
ALTER TABLE messages ADD COLUMN IF NOT EXISTS reactions JSONB DEFAULT '{}';
