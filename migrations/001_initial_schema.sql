-- Migration: 001_initial_schema
-- Description: Initial database schema for dmIntegroff
-- Date: 2024-11-29

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    created_at DATETIME(3) NULL,
    updated_at DATETIME(3) NULL,
    deleted_at DATETIME(3) NULL,
    username VARCHAR(191) NOT NULL UNIQUE,
    password VARCHAR(191) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'specialist',
    INDEX idx_users_deleted_at (deleted_at),
    INDEX idx_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Integrations table
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
    INDEX idx_integrations_deleted_at (deleted_at),
    INDEX idx_integrations_webhook_token (webhook_token),
    INDEX idx_integrations_mode (mode),
    FOREIGN KEY (created_by_id) REFERENCES users(id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Request logs table
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
    INDEX idx_request_logs_deleted_at (deleted_at),
    INDEX idx_request_logs_created_at (created_at),
    INDEX idx_request_logs_log_type (log_type),
    INDEX idx_request_logs_status_code (status_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Insert default admin user
-- Password: admin (bcrypt hash)
INSERT INTO users (created_at, updated_at, username, password, role)
VALUES (NOW(), NOW(), 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'admin')
ON DUPLICATE KEY UPDATE username=username;
