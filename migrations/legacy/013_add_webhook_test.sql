-- Migration 013: Add webhook test tables for MySQL/PostgreSQL
-- This migration adds tables for webhook testing functionality

-- Create webhook_tests table
CREATE TABLE IF NOT EXISTS webhook_tests (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    token VARCHAR(255) NOT NULL UNIQUE,
    user_id BIGINT UNSIGNED NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    INDEX idx_webhook_tests_token (token),
    INDEX idx_webhook_tests_user_id (user_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create webhook_test_requests table
CREATE TABLE IF NOT EXISTS webhook_test_requests (
    id BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY,
    webhook_test_id BIGINT UNSIGNED NOT NULL,
    method VARCHAR(10),
    url TEXT,
    headers TEXT,
    body LONGTEXT,
    query_params TEXT,
    client_ip VARCHAR(45),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (webhook_test_id) REFERENCES webhook_tests(id) ON DELETE CASCADE,
    INDEX idx_webhook_test_requests_webhook_test_id (webhook_test_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
