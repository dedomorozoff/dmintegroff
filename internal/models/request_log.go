package models

import (
	"gorm.io/gorm"
)

type RequestLog struct {
	gorm.Model
	IntegrationID  uint         `json:"integration_id"`
	Integration    *Integration `gorm:"foreignKey:IntegrationID" json:"integration,omitempty"`
	Method         string       `json:"method"`
	URL            string       `json:"url"`
	RequestBody    string       `gorm:"type:text" json:"request_body"`
	RequestHeaders string       `gorm:"type:text" json:"request_headers"`
	ResponseBody   string       `gorm:"type:text" json:"response_body"`
	StatusCode     int          `json:"status_code"`
	LogType        string       `json:"log_type"` // "request", "error", "webhook"
	ErrorMessage   string       `gorm:"type:text" json:"error_message"`
}
