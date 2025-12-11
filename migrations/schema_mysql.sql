-- Consolidated MySQL Schema for dmIntegroff
-- Description: Complete database schema with all migrations
-- Date: 2024-12-09

-- ============================================================================
-- USERS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    username VARCHAR(191) NOT NULL UNIQUE,
    password VARCHAR(191) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'specialist',
    is_demo BOOLEAN DEFAULT FALSE,
    expires_at DATETIME NULL,
    INDEX idx_users_deleted_at (deleted_at),
    INDEX idx_users_username (username),
    INDEX idx_users_expires_at (expires_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- PROJECTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS projects (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(191) NOT NULL,
    description TEXT,
    created_by_id BIGINT UNSIGNED NOT NULL,
    INDEX idx_projects_deleted_at (deleted_at),
    INDEX idx_projects_name (name),
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- INTEGRATIONS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS integrations (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    name VARCHAR(191) NOT NULL,
    webhook_token VARCHAR(191) NOT NULL UNIQUE,
    source_api TEXT,
    target_api TEXT NOT NULL,
    mapping_config TEXT,
    sample_payload LONGTEXT,
    mode VARCHAR(50) NOT NULL DEFAULT 'listening',
    status VARCHAR(50),
    created_by_id BIGINT UNSIGNED NOT NULL,
    project_id BIGINT UNSIGNED NOT NULL,
    output_template TEXT,
    http_method VARCHAR(10) DEFAULT 'POST',
    
    -- Authentication
    auth_type VARCHAR(20) DEFAULT 'none',
    oauth2_token_url VARCHAR(500),
    oauth2_client_id VARCHAR(255),
    oauth2_client_secret VARCHAR(255),
    oauth2_scope VARCHAR(500),
    oauth2_grant_type VARCHAR(50) DEFAULT 'client_credentials',
    bearer_token VARCHAR(500),
    basic_auth_user VARCHAR(255),
    basic_auth_pass VARCHAR(255),
    oauth2_access_token TEXT,
    oauth2_refresh_token TEXT,
    oauth2_expires_at BIGINT DEFAULT 0,
    
    -- Webhook Security
    webhook_signature_enabled BOOLEAN DEFAULT FALSE,
    webhook_signature_secret VARCHAR(255),
    webhook_signature_header VARCHAR(100) DEFAULT 'X-Webhook-Signature',
    webhook_signature_algorithm VARCHAR(20) DEFAULT 'sha256',
    webhook_http_methods TEXT,
    
    -- GraphQL Support
    api_type VARCHAR(20) DEFAULT 'rest',
    graphql_endpoint VARCHAR(500),
    graphql_query TEXT,
    graphql_variables TEXT,
    graphql_operation_name VARCHAR(255),
    graphql_schema TEXT,
    graphql_schema_updated_at DATETIME,
    
    -- GraphQL Enrichment
    enrichment_enabled BOOLEAN DEFAULT FALSE,
    enrichment_endpoint VARCHAR(500),
    enrichment_query TEXT,
    enrichment_variables TEXT,
    enrichment_merge_mode VARCHAR(20) DEFAULT 'merge',
    
    -- Other
    custom_headers TEXT,
    template_type VARCHAR(50) DEFAULT 'json',
    hide_in_logs BOOLEAN DEFAULT FALSE,
    
    INDEX idx_integrations_deleted_at (deleted_at),
    INDEX idx_integrations_webhook_token (webhook_token),
    INDEX idx_integrations_mode (mode),
    INDEX idx_integrations_project_id (project_id),
    INDEX idx_integrations_api_type (api_type),
    INDEX idx_integrations_enrichment_enabled (enrichment_enabled),
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- INTEGRATION OUTPUTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS integration_outputs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    integration_id BIGINT UNSIGNED NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    
    -- Target API Configuration
    target_api VARCHAR(500) NOT NULL,
    http_method VARCHAR(10) DEFAULT 'POST',
    
    -- Mapping Configuration
    mapping_config TEXT,
    output_template TEXT,
    template_type VARCHAR(20) DEFAULT 'json',
    
    -- Execution Control
    condition TEXT,
    priority INT DEFAULT 0,
    enabled BOOLEAN DEFAULT TRUE,
    
    -- Authentication
    auth_type VARCHAR(20) DEFAULT 'none',
    oauth2_token_url VARCHAR(500),
    oauth2_client_id VARCHAR(255),
    oauth2_client_secret VARCHAR(255),
    oauth2_scope VARCHAR(500),
    oauth2_grant_type VARCHAR(50) DEFAULT 'client_credentials',
    bearer_token VARCHAR(500),
    basic_auth_user VARCHAR(255),
    basic_auth_pass VARCHAR(255),
    
    -- OAuth2 Runtime Data
    oauth2_access_token TEXT,
    oauth2_refresh_token TEXT,
    oauth2_expires_at BIGINT,
    
    -- Custom Headers
    custom_headers TEXT,
    
    -- Timestamps
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    
    INDEX idx_integration_outputs_integration_id (integration_id),
    INDEX idx_integration_outputs_enabled (enabled),
    INDEX idx_integration_outputs_priority (priority),
    FOREIGN KEY (integration_id) REFERENCES integrations(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- REQUEST LOGS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS request_logs (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    integration_id BIGINT UNSIGNED,
    method VARCHAR(10),
    url TEXT,
    request_body LONGTEXT,
    response_body LONGTEXT,
    status_code INT,
    log_type VARCHAR(50) DEFAULT 'request',
    error_message TEXT,
    request_headers TEXT,
    output_name VARCHAR(255) DEFAULT '',
    INDEX idx_request_logs_deleted_at (deleted_at),
    INDEX idx_request_logs_created_at (created_at),
    INDEX idx_request_logs_log_type (log_type),
    INDEX idx_request_logs_status_code (status_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- WEBHOOK TESTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS webhook_tests (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255),
    description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME,
    created_by_id BIGINT UNSIGNED,
    INDEX idx_webhook_tests_token (token),
    INDEX idx_webhook_tests_expires_at (expires_at),
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE SET NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- WEBHOOK TEST REQUESTS TABLE
-- ============================================================================
CREATE TABLE IF NOT EXISTS webhook_test_requests (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    webhook_test_id BIGINT UNSIGNED NOT NULL,
    method VARCHAR(10),
    headers TEXT,
    body LONGTEXT,
    query_params TEXT,
    received_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_webhook_test_requests_webhook_test_id (webhook_test_id),
    INDEX idx_webhook_test_requests_received_at (received_at),
    FOREIGN KEY (webhook_test_id) REFERENCES webhook_tests(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- ============================================================================
-- DEFAULT DATA
-- ============================================================================

-- Insert default admin user
-- Password: admin (bcrypt hash)
INSERT INTO users (created_at, updated_at, username, password, role)
VALUES (NOW(), NOW(), 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin')
ON DUPLICATE KEY UPDATE username=username;
