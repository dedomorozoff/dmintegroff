-- Add custom headers support to integrations
ALTER TABLE integrations ADD COLUMN custom_headers TEXT;
