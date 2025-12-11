-- Migration: Add GraphQL support
-- Description: Adds GraphQL configuration fields to integrations table

-- Add GraphQL configuration columns
ALTER TABLE integrations ADD COLUMN api_type VARCHAR(20) DEFAULT 'rest' AFTER http_method;
-- api_type: rest, graphql, graphql_server

ALTER TABLE integrations ADD COLUMN graphql_endpoint VARCHAR(500) AFTER api_type;
-- GraphQL endpoint URL (for client mode)

ALTER TABLE integrations ADD COLUMN graphql_query TEXT AFTER graphql_endpoint;
-- GraphQL query template with {{variables}}

ALTER TABLE integrations ADD COLUMN graphql_variables TEXT AFTER graphql_query;
-- JSON mapping for GraphQL variables

ALTER TABLE integrations ADD COLUMN graphql_operation_name VARCHAR(255) AFTER graphql_variables;
-- Optional operation name

ALTER TABLE integrations ADD COLUMN graphql_schema TEXT AFTER graphql_operation_name;
-- Cached GraphQL schema (from introspection)

ALTER TABLE integrations ADD COLUMN graphql_schema_updated_at DATETIME AFTER graphql_schema;
-- Last schema update timestamp

-- Add index for api_type
CREATE INDEX idx_integrations_api_type ON integrations(api_type);
