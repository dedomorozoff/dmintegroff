package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username  string     `gorm:"uniqueIndex;not null" json:"username"`
	Password  string     `gorm:"not null" json:"-"` // Store hashed password
	Role      string     `gorm:"not null;default:'specialist'" json:"role"` // 'admin' or 'specialist'
	IsDemo    bool       `gorm:"default:false" json:"is_demo"` // Demo user flag
	ExpiresAt *time.Time `gorm:"index" json:"expires_at,omitempty"` // Expiration time for demo users
}
