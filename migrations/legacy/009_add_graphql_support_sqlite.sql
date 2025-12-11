-- Migration: Add GraphQL support (SQLite version)
-- Description: Adds GraphQL configuration fields to integrations table

-- SQLite doesn't support ALTER TABLE ADD COLUMN with AFTER clause
-- Add columns one by one

ALTER TABLE integrations ADD COLUMN api_type VARCHAR(20) DEFAULT 'rest';
ALTER TABLE integrations ADD COLUMN graphql_endpoint VARCHAR(500);
ALTER TABLE integrations ADD COLUMN graphql_query TEXT;
ALTER TABLE integrations ADD COLUMN graphql_variables TEXT;
ALTER TABLE integrations ADD COLUMN graphql_operation_name VARCHAR(255);
ALTER TABLE integrations ADD COLUMN graphql_schema TEXT;
ALTER TABLE integrations ADD COLUMN graphql_schema_updated_at DATETIME;

-- Add index for api_type
CREATE INDEX IF NOT EXISTS idx_integrations_api_type ON integrations(api_type);
