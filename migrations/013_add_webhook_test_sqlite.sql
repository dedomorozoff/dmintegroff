-- Migration 013: Add webhook test tables for SQLite
-- This migration adds tables for webhook testing functionality

-- Create webhook_tests table
CREATE TABLE IF NOT EXISTS webhook_tests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    expires_at DATETIME NOT NULL,
    is_active INTEGER DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_tests_token ON webhook_tests(token);
CREATE INDEX IF NOT EXISTS idx_webhook_tests_user_id ON webhook_tests(user_id);

-- Create webhook_test_requests table
CREATE TABLE IF NOT EXISTS webhook_test_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_test_id INTEGER NOT NULL,
    method TEXT,
    url TEXT,
    headers TEXT,
    body TEXT,
    query_params TEXT,
    client_ip TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (webhook_test_id) REFERENCES webhook_tests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_test_requests_webhook_test_id ON webhook_test_requests(webhook_test_id);
