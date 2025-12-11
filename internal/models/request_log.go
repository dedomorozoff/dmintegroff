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
	ResponseHeaders string      `gorm:"type:text" json:"response_headers"` // Добавлено для заголовков ответа
	StatusCode     int          `json:"status_code"`
	LogType        string       `json:"log_type"` // "request", "error", "webhook", "incoming"
	ErrorMessage   string       `gorm:"type:text" json:"error_message"`
	OutputName     string       `json:"output_name"` // Название дополнительного выхода (если применимо)
	ResponseTime   int64        `json:"response_time"` // Время ответа в миллисекундах
	RequestSize    int          `json:"request_size"`  // Размер запроса в байтах
	ResponseSize   int          `json:"response_size"` // Размер ответа в байтах
}
