-- Add demo user fields to users table
ALTER TABLE users ADD COLUMN is_demo BOOLEAN DEFAULT 0;
ALTER TABLE users ADD COLUMN expires_at DATETIME;

-- Create index on expires_at for faster cleanup queries
CREATE INDEX idx_users_expires_at ON users(expires_at);
