-- Add demo user fields to users table (MySQL)
ALTER TABLE users ADD COLUMN is_demo BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN expires_at DATETIME NULL;

-- Create index on expires_at for faster cleanup queries
CREATE INDEX idx_users_expires_at ON users(expires_at);
