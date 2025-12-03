-- Migration: Add OAuth 2.0 support to integrations (MySQL)
-- Date: 2025-12-03

-- Add OAuth and authentication fields to integrations table
ALTER TABLE integrations 
ADD COLUMN auth_type VARCHAR(50) DEFAULT 'none',
ADD COLUMN oauth2_token_url VARCHAR(500),
ADD COLUMN oauth2_client_id VARCHAR(255),
ADD COLUMN oauth2_client_secret TEXT,
ADD COLUMN oauth2_scope VARCHAR(500),
ADD COLUMN oauth2_grant_type VARCHAR(50) DEFAULT 'client_credentials',
ADD COLUMN bearer_token TEXT,
ADD COLUMN basic_auth_user VARCHAR(255),
ADD COLUMN basic_auth_pass TEXT,
ADD COLUMN oauth2_access_token TEXT,
ADD COLUMN oauth2_refresh_token TEXT,
ADD COLUMN oauth2_expires_at BIGINT DEFAULT 0;
