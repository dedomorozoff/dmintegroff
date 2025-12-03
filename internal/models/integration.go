package models

import (
	"gorm.io/gorm"
)

type Integration struct {
	gorm.Model
	Name           string   `gorm:"not null" json:"name"`
	WebhookToken   string   `gorm:"uniqueIndex;not null" json:"webhook_token"` // Unique token for webhook URL
	SourceAPI      string   `json:"source_api"`                                 // Optional, for documentation
	TargetAPI      string   `gorm:"not null" json:"target_api"`
	HTTPMethod     string   `gorm:"default:'POST'" json:"http_method"`          // HTTP method: GET, POST, PUT, PATCH, DELETE
	Mode           string   `gorm:"default:'listening'" json:"mode"`            // listening, active, inactive
	SamplePayload  string   `gorm:"type:text" json:"sample_payload"`            // JSON sample from first request
	MappingConfig  string   `gorm:"type:text" json:"mapping_config"`            // JSON string for field mapping
	OutputTemplate string   `gorm:"type:text" json:"output_template"`           // JSON template with {{field.path}} placeholders
	Status         string   `gorm:"default:'active'" json:"status"`             // active, inactive (deprecated, use Mode)
	ProjectID      uint     `gorm:"not null;index" json:"project_id"`           // Required project assignment
	Project        Project  `gorm:"constraint:OnDelete:CASCADE;" json:"project"`
	CreatedByID    uint     `gorm:"not null;index" json:"created_by_id"`
	CreatedBy      User     `gorm:"constraint:OnDelete:CASCADE;" json:"created_by"`
	
	// OAuth 2.0 Configuration
	AuthType         string `gorm:"default:'none'" json:"auth_type"`           // none, oauth2, bearer, basic
	OAuth2TokenURL   string `json:"oauth2_token_url"`                          // OAuth2 token endpoint
	OAuth2ClientID   string `json:"oauth2_client_id"`                          // OAuth2 client ID
	OAuth2ClientSecret string `json:"oauth2_client_secret"`                    // OAuth2 client secret (encrypted)
	OAuth2Scope      string `json:"oauth2_scope"`                              // OAuth2 scopes (space-separated)
	OAuth2GrantType  string `gorm:"default:'client_credentials'" json:"oauth2_grant_type"` // client_credentials, password, etc.
	BearerToken      string `json:"bearer_token"`                              // Static bearer token
	BasicAuthUser    string `json:"basic_auth_user"`                           // Basic auth username
	BasicAuthPass    string `json:"basic_auth_pass"`                           // Basic auth password
	
	// OAuth2 Runtime Data (managed by system)
	OAuth2AccessToken  string `json:"-"` // Current access token (not exposed in JSON)
	OAuth2RefreshToken string `json:"-"` // Refresh token (not exposed in JSON)
	OAuth2ExpiresAt    int64  `json:"-"` // Token expiration timestamp
}
