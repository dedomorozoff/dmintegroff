-- Migration: Add OAuth 2.0 support to integrations
-- Date: 2025-12-03

-- Add OAuth and authentication fields to integrations table
ALTER TABLE integrations ADD COLUMN auth_type TEXT DEFAULT 'none';
ALTER TABLE integrations ADD COLUMN oauth2_token_url TEXT;
ALTER TABLE integrations ADD COLUMN oauth2_client_id TEXT;
ALTER TABLE integrations ADD COLUMN oauth2_client_secret TEXT;
ALTER TABLE integrations ADD COLUMN oauth2_scope TEXT;
ALTER TABLE integrations ADD COLUMN oauth2_grant_type TEXT DEFAULT 'client_credentials';
ALTER TABLE integrations ADD COLUMN bearer_token TEXT;
ALTER TABLE integrations ADD COLUMN basic_auth_user TEXT;
ALTER TABLE integrations ADD COLUMN basic_auth_pass TEXT;

-- OAuth2 runtime data (managed by system)
ALTER TABLE integrations ADD COLUMN oauth2_access_token TEXT;
ALTER TABLE integrations ADD COLUMN oauth2_refresh_token TEXT;
ALTER TABLE integrations ADD COLUMN oauth2_expires_at INTEGER DEFAULT 0;
