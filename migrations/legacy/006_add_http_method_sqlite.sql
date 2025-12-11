-- Add HTTP method field to integrations table (SQLite)
ALTER TABLE integrations ADD COLUMN http_method TEXT DEFAULT 'POST';
