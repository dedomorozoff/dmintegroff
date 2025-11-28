package models

import (
	"gorm.io/gorm"
)

type RequestLog struct {
	gorm.Model
	IntegrationID uint   `json:"integration_id"`
	Method        string `json:"method"`
	URL           string `json:"url"`
	RequestBody   string `gorm:"type:text" json:"request_body"`
	ResponseBody  string `gorm:"type:text" json:"response_body"`
	StatusCode    int    `json:"status_code"`
	LogType       string `json:"log_type"` // "request", "error", "webhook"
	ErrorMessage  string `gorm:"type:text" json:"error_message"`
}
