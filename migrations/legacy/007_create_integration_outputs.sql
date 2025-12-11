-- +migrate Up
CREATE TABLE IF NOT EXISTS integration_outputs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    integration_id INTEGER NOT NULL,
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
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT 1,
    
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

-- +migrate Down
DROP TABLE IF EXISTS integration_outputs;
