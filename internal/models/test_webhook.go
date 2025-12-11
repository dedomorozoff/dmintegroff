package models

import (
	"time"

	"gorm.io/gorm"
)

// WebhookTest represents a test webhook for debugging
type WebhookTest struct {
	gorm.Model
	Token     string    `gorm:"uniqueIndex;not null" json:"token"`
	UserID    uint      `gorm:"not null;index" json:"user_id"`
	User      User      `gorm:"constraint:OnDelete:CASCADE;" json:"user"`
	ExpiresAt time.Time `gorm:"not null" json:"expires_at"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
}

// WebhookTestRequest represents a request to a test webhook
type WebhookTestRequest struct {
	gorm.Model
	WebhookTestID uint         `gorm:"not null;index" json:"webhook_test_id"`
	WebhookTest   *WebhookTest `gorm:"foreignKey:WebhookTestID" json:"webhook_test,omitempty"`
	Method        string       `json:"method"`
	URL           string       `json:"url"`
	Headers       string       `gorm:"type:text" json:"headers"`
	Body          string       `gorm:"type:text" json:"body"`
	QueryParams   string       `gorm:"type:text" json:"query_params"`
	ClientIP      string       `json:"client_ip"`
}
