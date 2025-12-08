-- Add custom headers support to integrations (SQLite)
ALTER TABLE integrations ADD COLUMN custom_headers TEXT;
