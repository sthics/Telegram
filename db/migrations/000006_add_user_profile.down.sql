-- Remove profile fields from users table
DROP INDEX IF EXISTS idx_users_username;

ALTER TABLE users 
DROP COLUMN IF EXISTS username,
DROP COLUMN IF EXISTS avatar_url,
DROP COLUMN IF EXISTS bio;
