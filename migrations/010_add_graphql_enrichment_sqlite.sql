-- Migration: Add GraphQL Enrichment support (SQLite version)
-- Description: Adds GraphQL enrichment fields for REST integrations

-- SQLite doesn't support ALTER TABLE ADD COLUMN with AFTER clause
ALTER TABLE integrations ADD COLUMN enrichment_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE integrations ADD COLUMN enrichment_endpoint VARCHAR(500);
ALTER TABLE integrations ADD COLUMN enrichment_query TEXT;
ALTER TABLE integrations ADD COLUMN enrichment_variables TEXT;
ALTER TABLE integrations ADD COLUMN enrichment_merge_mode VARCHAR(20) DEFAULT 'merge';

-- Add index for enrichment_enabled
CREATE INDEX IF NOT EXISTS idx_integrations_enrichment_enabled ON integrations(enrichment_enabled);
