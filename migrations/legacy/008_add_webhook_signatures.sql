-- Add webhook signature fields for security
ALTER TABLE integrations ADD COLUMN webhook_signature_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE integrations ADD COLUMN webhook_signature_secret VARCHAR(255);
ALTER TABLE integrations ADD COLUMN webhook_signature_header VARCHAR(100) DEFAULT 'X-Webhook-Signature';
ALTER TABLE integrations ADD COLUMN webhook_signature_algorithm VARCHAR(20) DEFAULT 'sha256';
