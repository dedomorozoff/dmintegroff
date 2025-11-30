package models

import (
	"gorm.io/gorm"
)

type Project struct {
	gorm.Model
	Name         string        `gorm:"not null" json:"name"`
	Description  string        `gorm:"type:text" json:"description"`
	CreatedByID  uint          `json:"created_by_id"`
	CreatedBy    User          `gorm:"foreignKey:CreatedByID" json:"created_by"`
	Integrations []Integration `gorm:"foreignKey:ProjectID;constraint:OnDelete:CASCADE" json:"integrations"`
}
