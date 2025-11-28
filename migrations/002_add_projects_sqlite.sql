-- Migration: 002_add_projects
-- Description: Add projects table and link integrations to projects (SQLite version)
-- Date: 2024-11-29

-- Projects table
CREATE TABLE IF NOT EXISTS projects (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NULL,
    updated_at DATETIME NULL,
    deleted_at DATETIME NULL,
    name TEXT NOT NULL,
    description TEXT,
    created_by_id INTEGER NOT NULL,
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_projects_deleted_at ON projects(deleted_at);
CREATE INDEX IF NOT EXISTS idx_projects_name ON projects(name);

-- Add project_id column to integrations table
-- SQLite doesn't support ALTER TABLE ADD FOREIGN KEY directly
-- We need to check if column exists first
PRAGMA foreign_keys=off;

-- Create new table with project_id
CREATE TABLE IF NOT EXISTS integrations_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME NULL,
    updated_at DATETIME NULL,
    deleted_at DATETIME NULL,
    name TEXT NOT NULL,
    webhook_token TEXT NOT NULL UNIQUE,
    source_api TEXT,
    target_api TEXT NOT NULL,
    mapping_config TEXT,
    sample_payload TEXT,
    mode TEXT NOT NULL DEFAULT 'listening',
    status TEXT,
    created_by_id INTEGER NOT NULL,
    project_id INTEGER NULL,
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL
);

-- Copy data from old table
INSERT INTO integrations_new (id, created_at, updated_at, deleted_at, name, webhook_token, source_api, target_api, mapping_config, sample_payload, mode, status, created_by_id)
SELECT id, created_at, updated_at, deleted_at, name, webhook_token, source_api, target_api, mapping_config, sample_payload, mode, status, created_by_id
FROM integrations;

-- Drop old table and rename new one
DROP TABLE integrations;
ALTER TABLE integrations_new RENAME TO integrations;

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_integrations_deleted_at ON integrations(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_integrations_webhook_token ON integrations(webhook_token);
CREATE INDEX IF NOT EXISTS idx_integrations_mode ON integrations(mode);
CREATE INDEX IF NOT EXISTS idx_integrations_project_id ON integrations(project_id);

PRAGMA foreign_keys=on;
