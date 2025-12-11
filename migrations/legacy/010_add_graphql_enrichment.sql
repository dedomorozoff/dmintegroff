-- Migration: Add GraphQL Enrichment support
-- Description: Adds GraphQL enrichment fields for REST integrations

-- Add GraphQL enrichment columns
ALTER TABLE integrations ADD COLUMN enrichment_enabled BOOLEAN DEFAULT FALSE AFTER graphql_schema_updated_at;
ALTER TABLE integrations ADD COLUMN enrichment_endpoint VARCHAR(500) AFTER enrichment_enabled;
ALTER TABLE integrations ADD COLUMN enrichment_query TEXT AFTER enrichment_endpoint;
ALTER TABLE integrations ADD COLUMN enrichment_variables TEXT AFTER enrichment_query;
ALTER TABLE integrations ADD COLUMN enrichment_merge_mode VARCHAR(20) DEFAULT 'merge' AFTER enrichment_variables;

-- Add index for enrichment_enabled
CREATE INDEX idx_integrations_enrichment_enabled ON integrations(enrichment_enabled);
