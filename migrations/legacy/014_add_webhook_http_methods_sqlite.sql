-- Add webhook_http_methods column to integrations table
ALTER TABLE integrations ADD COLUMN webhook_http_methods TEXT DEFAULT '*';
