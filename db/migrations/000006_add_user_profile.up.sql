-- Add profile fields to users table
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS username VARCHAR(50),
ADD COLUMN IF NOT EXISTS avatar_url TEXT,
ADD COLUMN IF NOT EXISTS bio TEXT;

-- Create index for username search
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
