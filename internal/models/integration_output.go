package models

import (
	"gorm.io/gorm"
)

// IntegrationOutput represents a single output (target) for an integration
// One integration can have multiple outputs, each with its own mapping and target API
type IntegrationOutput struct {
	gorm.Model
	IntegrationID uint   `gorm:"not null;index" json:"integration_id"`
	Name          string `gorm:"not null" json:"name"` // "CRM Output", "Analytics Output"
	Description   string `gorm:"type:text" json:"description"`
	
	// Target API Configuration
	TargetAPI  string `gorm:"not null" json:"target_api"`
	HTTPMethod string `gorm:"default:'POST'" json:"http_method"`
	
	// Mapping Configuration (same as Integration)
	MappingConfig  string `gorm:"type:text" json:"mapping_config"`
	OutputTemplate string `gorm:"type:text" json:"output_template"`
	TemplateType   string `gorm:"default:'json'" json:"template_type"` // json, xml, text, custom
	
	// Execution Control
	Condition string `gorm:"type:text" json:"condition"` // Условие выполнения (опционально)
	Priority  int    `gorm:"default:0" json:"priority"`  // Порядок выполнения (меньше = раньше)
	Enabled   bool   `gorm:"default:true" json:"enabled"`
	
	// Authentication (копия из Integration)
	AuthType           string `gorm:"default:'none'" json:"auth_type"` // none, oauth2, bearer, basic
	OAuth2TokenURL     string `json:"oauth2_token_url"`
	OAuth2ClientID     string `json:"oauth2_client_id"`
	OAuth2ClientSecret string `json:"oauth2_client_secret"`
	OAuth2Scope        string `json:"oauth2_scope"`
	OAuth2GrantType    string `gorm:"default:'client_credentials'" json:"oauth2_grant_type"`
	BearerToken        string `json:"bearer_token"`
	BasicAuthUser      string `json:"basic_auth_user"`
	BasicAuthPass      string `json:"basic_auth_pass"`
	
	// OAuth2 Runtime Data
	OAuth2AccessToken  string `json:"-"`
	OAuth2RefreshToken string `json:"-"`
	OAuth2ExpiresAt    int64  `json:"-"`
	
	// Custom Headers
	CustomHeaders string `gorm:"type:text" json:"custom_headers"` // JSON map of custom HTTP headers
	
	// Relationship
	Integration Integration `gorm:"constraint:OnDelete:CASCADE;" json:"-"`
}
