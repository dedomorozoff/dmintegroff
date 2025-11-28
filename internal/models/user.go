package models

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Username string `gorm:"uniqueIndex;not null" json:"username"`
	Password string `gorm:"not null" json:"-"` // Store hashed password
	Role     string `gorm:"not null;default:'specialist'" json:"role"` // 'admin' or 'specialist'
}
