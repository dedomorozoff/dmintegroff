-- Add webhook signature fields for security (SQLite version)
ALTER TABLE integrations ADD COLUMN webhook_signature_enabled INTEGER DEFAULT 0;
ALTER TABLE integrations ADD COLUMN webhook_signature_secret TEXT;
ALTER TABLE integrations ADD COLUMN webhook_signature_header TEXT DEFAULT 'X-Webhook-Signature';
ALTER TABLE integrations ADD COLUMN webhook_signature_algorithm TEXT DEFAULT 'sha256';
