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
	Mode           string   `gorm:"default:'listening'" json:"mode"`      // listening, active, inactive
	SamplePayload  string   `gorm:"type:text" json:"sample_payload"`      // JSON sample from first request
	MappingConfig  string   `gorm:"type:text" json:"mapping_config"`      // JSON string for field mapping
	OutputTemplate string   `gorm:"type:text" json:"output_template"`     // JSON template with {{field.path}} placeholders
	Status         string  `gorm:"default:'active'" json:"status"`       // active, inactive (deprecated, use Mode)
	ProjectID      uint    `gorm:"not null;index" json:"project_id"`     // Required project assignment
	Project        Project `gorm:"constraint:OnDelete:CASCADE;" json:"project"`
	CreatedByID    uint    `gorm:"not null;index" json:"created_by_id"`
	CreatedBy      User    `gorm:"constraint:OnDelete:CASCADE;" json:"created_by"`
}
