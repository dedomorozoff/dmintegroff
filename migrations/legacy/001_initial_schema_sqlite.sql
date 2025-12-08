-- Migration: 001_initial_schema (SQLite version)
-- Description: Initial database schema for dmIntegroff
-- Date: 2024-11-29

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'specialist'
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

-- Integrations table
CREATE TABLE IF NOT EXISTS integrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    name TEXT NOT NULL,
    webhook_token TEXT NOT NULL UNIQUE,
    source_api TEXT,
    target_api TEXT NOT NULL,
    mapping_config TEXT,
    sample_payload TEXT,
    mode TEXT NOT NULL DEFAULT 'listening',
    status TEXT,
    created_by_id INTEGER NOT NULL,
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_integrations_deleted_at ON integrations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_integrations_webhook_token ON integrations(webhook_token);
CREATE INDEX IF NOT EXISTS idx_integrations_mode ON integrations(mode);

-- Request logs table
CREATE TABLE IF NOT EXISTS request_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    integration_id INTEGER,
    method TEXT,
    url TEXT,
    request_body TEXT,
    response_body TEXT,
    status_code INTEGER,
    log_type TEXT DEFAULT 'request',
    error_message TEXT
);

CREATE INDEX IF NOT EXISTS idx_request_logs_deleted_at ON request_logs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_log_type ON request_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_request_logs_status_code ON request_logs(status_code);

-- Insert default admin user
-- Password: admin (bcrypt hash)
INSERT OR IGNORE INTO users (created_at, updated_at, username, password, role)
VALUES (datetime('now'), datetime('now'), 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin');
