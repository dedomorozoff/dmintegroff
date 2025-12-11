-- Add HTTP method field to integrations table
ALTER TABLE integrations ADD COLUMN http_method VARCHAR(10) DEFAULT 'POST';
