-- Consolidated SQLite Schema for dmIntegroff
-- Description: Complete database schema with all migrations
-- Date: 2024-12-09

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    username TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    role TEXT NOT NULL DEFAULT 'specialist',
    is_demo INTEGER DEFAULT 0,
    expires_at DATETIME
);

CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_expires_at ON users(expires_at);

-- ============================================================================
-- PROJECTS TABLE
-- ============================================================================
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

-- ============================================================================
-- INTEGRATIONS TABLE
-- ============================================================================
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
    project_id INTEGER NOT NULL,
    output_template TEXT,
    http_method TEXT DEFAULT 'POST',
    
    -- Authentication
    auth_type TEXT DEFAULT 'none',
    oauth2_token_url TEXT,
    oauth2_client_id TEXT,
    oauth2_client_secret TEXT,
    oauth2_scope TEXT,
    oauth2_grant_type TEXT DEFAULT 'client_credentials',
    bearer_token TEXT,
    basic_auth_user TEXT,
    basic_auth_pass TEXT,
    oauth2_access_token TEXT,
    oauth2_refresh_token TEXT,
    oauth2_expires_at INTEGER DEFAULT 0,
    
    -- Webhook Security
    webhook_signature_enabled INTEGER DEFAULT 0,
    webhook_signature_secret TEXT,
    webhook_signature_header TEXT DEFAULT 'X-Webhook-Signature',
    webhook_signature_algorithm TEXT DEFAULT 'sha256',
    webhook_http_methods TEXT DEFAULT '*',
    
    -- GraphQL Support
    api_type TEXT DEFAULT 'rest',
    graphql_endpoint TEXT,
    graphql_query TEXT,
    graphql_variables TEXT,
    graphql_operation_name TEXT,
    graphql_schema TEXT,
    graphql_schema_updated_at DATETIME,
    
    -- GraphQL Enrichment
    enrichment_enabled INTEGER DEFAULT 0,
    enrichment_endpoint TEXT,
    enrichment_query TEXT,
    enrichment_variables TEXT,
    enrichment_merge_mode TEXT DEFAULT 'merge',
    
    -- Other
    custom_headers TEXT,
    template_type TEXT DEFAULT 'json',
    hide_in_logs INTEGER DEFAULT 0,
    
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_integrations_deleted_at ON integrations(deleted_at);
CREATE INDEX IF NOT EXISTS idx_integrations_webhook_token ON integrations(webhook_token);
CREATE INDEX IF NOT EXISTS idx_integrations_mode ON integrations(mode);
CREATE INDEX IF NOT EXISTS idx_integrations_project_id ON integrations(project_id);
CREATE INDEX IF NOT EXISTS idx_integrations_api_type ON integrations(api_type);
CREATE INDEX IF NOT EXISTS idx_integrations_enrichment_enabled ON integrations(enrichment_enabled);

-- ============================================================================
-- INTEGRATION OUTPUTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS integration_outputs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    integration_id INTEGER NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    
    -- Target API Configuration
    target_api TEXT NOT NULL,
    http_method TEXT DEFAULT 'POST',
    
    -- Mapping Configuration
    mapping_config TEXT,
    output_template TEXT,
    template_type TEXT DEFAULT 'json',
    
    -- Execution Control
    condition TEXT,
    priority INTEGER DEFAULT 0,
    enabled INTEGER DEFAULT 1,
    
    -- Authentication
    auth_type TEXT DEFAULT 'none',
    oauth2_token_url TEXT,
    oauth2_client_id TEXT,
    oauth2_client_secret TEXT,
    oauth2_scope TEXT,
    oauth2_grant_type TEXT DEFAULT 'client_credentials',
    bearer_token TEXT,
    basic_auth_user TEXT,
    basic_auth_pass TEXT,
    
    -- OAuth2 Runtime Data
    oauth2_access_token TEXT,
    oauth2_refresh_token TEXT,
    oauth2_expires_at INTEGER,
    
    -- Custom Headers
    custom_headers TEXT,
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_integration_outputs_integration_id ON integration_outputs(integration_id);
CREATE INDEX IF NOT EXISTS idx_integration_outputs_enabled ON integration_outputs(enabled);
CREATE INDEX IF NOT EXISTS idx_integration_outputs_priority ON integration_outputs(priority);

-- ============================================================================
-- REQUEST LOGS TABLE
-- ============================================================================
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
    error_message TEXT,
    request_headers TEXT,
    output_name TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_request_logs_deleted_at ON request_logs(deleted_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_created_at ON request_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_request_logs_log_type ON request_logs(log_type);
CREATE INDEX IF NOT EXISTS idx_request_logs_status_code ON request_logs(status_code);

-- ============================================================================
-- WEBHOOK TESTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS webhook_tests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    token TEXT NOT NULL UNIQUE,
    name TEXT,
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    created_by_id INTEGER,
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_webhook_tests_token ON webhook_tests(token);
CREATE INDEX IF NOT EXISTS idx_webhook_tests_expires_at ON webhook_tests(expires_at);

-- ============================================================================
-- WEBHOOK TEST REQUESTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS webhook_test_requests (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    webhook_test_id INTEGER NOT NULL,
    method TEXT,
    headers TEXT,
    body TEXT,
    query_params TEXT,
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (webhook_test_id) REFERENCES webhook_tests(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_webhook_test_requests_webhook_test_id ON webhook_test_requests(webhook_test_id);
CREATE INDEX IF NOT EXISTS idx_webhook_test_requests_received_at ON webhook_test_requests(received_at);

-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

-- Insert default admin user
-- Password: admin (bcrypt hash)
INSERT OR IGNORE INTO users (created_at, updated_at, username, password, role)
VALUES (datetime('now'), datetime('now'), 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin');
