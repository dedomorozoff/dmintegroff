package models

import (
	"gorm.io/gorm"
)

type Integration struct {
	gorm.Model
	Name          string `gorm:"not null" json:"name"`
	SourceAPI     string `gorm:"not null" json:"source_api"`
	TargetAPI     string `gorm:"not null" json:"target_api"`
	MappingConfig string `gorm:"type:text" json:"mapping_config"` // JSON string for field mapping
	Status        string `gorm:"default:'active'" json:"status"`  // 'active', 'inactive'
	CreatedByID   uint   `json:"created_by_id"`
	CreatedBy     User   `gorm:"foreignKey:CreatedByID" json:"created_by"`
}
