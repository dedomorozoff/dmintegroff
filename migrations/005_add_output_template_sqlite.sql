-- Migration: 005_add_output_template (SQLite version)
-- Description: Add output_template field for custom JSON formatting
-- Date: 2024-11-30

-- Add output_template column to integrations table
ALTER TABLE integrations ADD COLUMN output_template TEXT;

-- Note: output_template will store a JSON template with placeholders like {{field.path}}
-- Example: {"user": {"name": "{{user.name}}", "email": "{{user.email}}"}}
