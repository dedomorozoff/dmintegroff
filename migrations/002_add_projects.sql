-- Migration: 002_add_projects
-- Description: Add projects table and link integrations to projects
-- Date: 2024-11-29

-- Projects table
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

-- Add project_id column to integrations table
ALTER TABLE integrations 
ADD COLUMN project_id BIGINT UNSIGNED NULL,
ADD INDEX idx_integrations_project_id (project_id),
ADD FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE SET NULL;
