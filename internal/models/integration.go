package models

import (
	"gorm.io/gorm"
)

type Integration struct {
	gorm.Model
	Name          string `gorm:"not null" json:"name"`
	WebhookToken  string `gorm:"uniqueIndex;not null" json:"webhook_token"` // Unique token for webhook URL
	SourceAPI     string `json:"source_api"`                                 // Optional, for documentation
	TargetAPI     string `gorm:"not null" json:"target_api"`
	Mode          string `gorm:"default:'listening'" json:"mode"`      // listening, active, inactive
	SamplePayload string `gorm:"type:text" json:"sample_payload"`      // JSON sample from first request
	MappingConfig string `gorm:"type:text" json:"mapping_config"`      // JSON string for field mapping
	Status        string `gorm:"default:'active'" json:"status"`       // active, inactive (deprecated, use Mode)
	CreatedByID   uint   `json:"created_by_id"`
	CreatedBy     User   `gorm:"foreignKey:CreatedByID" json:"created_by"`
}
