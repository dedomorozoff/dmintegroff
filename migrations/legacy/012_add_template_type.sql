-- Migration: Add template_type field to integrations table (PostgreSQL/MySQL)
-- This allows users to choose between JSON, XML, plain text, or custom template formats

-- Add template_type column with default value 'json'
ALTER TABLE integrations ADD COLUMN template_type VARCHAR(50) DEFAULT 'json';

-- Update existing records to have 'json' as template_type
UPDATE integrations SET template_type = 'json' WHERE template_type IS NULL OR template_type = '';
