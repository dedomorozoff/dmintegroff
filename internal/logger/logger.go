package logger

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

var Log = logrus.New()

func Init() {
	// Set log level
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}
	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logLevel = logrus.InfoLevel
	}
	Log.SetLevel(logLevel)

	// Set formatter based on environment
	logFormat := os.Getenv("LOG_FORMAT")
	if logFormat == "json" {
		Log.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
		})
	} else {
		Log.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
		})
	}

	// Set output
	logFile := os.Getenv("LOG_FILE")
	if logFile != "" {
		// Создаем директорию для логов если не существует
		logDir := filepath.Dir(logFile)
		if err := os.MkdirAll(logDir, 0755); err != nil {
			Log.Warn("Failed to create log directory: " + err.Error())
		}

		file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err == nil {
			// Логи идут и в файл, и в stdout
			Log.SetOutput(io.MultiWriter(os.Stdout, file))
			Log.Info("Logging to file: " + logFile)
		} else {
			Log.Warn("Failed to log to file, using stdout only: " + err.Error())
			Log.SetOutput(os.Stdout)
		}
	} else {
		Log.SetOutput(os.Stdout)
	}

	Log.WithFields(logrus.Fields{
		"level":  logLevel.String(),
		"format": logFormat,
	}).Info("Logger initialized")
}

// LogWebhookRequest логирует входящий webhook запрос
func LogWebhookRequest(integrationID uint, method, url, contentType string, bodySize int, statusCode int) {
	Log.WithFields(logrus.Fields{
		"type":           "webhook_request",
		"integration_id": integrationID,
		"method":         method,
		"url":            url,
		"content_type":   contentType,
		"body_size":      bodySize,
		"status_code":    statusCode,
	}).Info("Webhook request processed")
}

// LogWebhookResponse логирует ответ на webhook запрос
func LogWebhookResponse(integrationID uint, targetURL string, statusCode int, responseTime int64, error string) {
	fields := logrus.Fields{
		"type":           "webhook_response",
		"integration_id": integrationID,
		"target_url":     targetURL,
		"status_code":    statusCode,
		"response_time":  responseTime, // в миллисекундах
	}

	if error != "" {
		fields["error"] = error
		Log.WithFields(fields).Error("Webhook response failed")
	} else {
		Log.WithFields(fields).Info("Webhook response successful")
	}
}

// LogRateLimitExceeded логирует превышение rate limit
func LogRateLimitExceeded(clientIP, path, method string) {
	Log.WithFields(logrus.Fields{
		"type":      "rate_limit_exceeded",
		"client_ip": clientIP,
		"path":      path,
		"method":    method,
	}).Warn("Rate limit exceeded")
}

// LogSystemError логирует системные ошибки
func LogSystemError(component, operation, error string, details map[string]interface{}) {
	fields := logrus.Fields{
		"type":      "system_error",
		"component": component,
		"operation": operation,
		"error":     error,
	}

	// Добавляем дополнительные детали
	for k, v := range details {
		fields[k] = v
	}

	Log.WithFields(fields).Error("System error occurred")
}

// LogIntegrationEvent логирует события интеграций
func LogIntegrationEvent(integrationID uint, event, details string) {
	Log.WithFields(logrus.Fields{
		"type":           "integration_event",
		"integration_id": integrationID,
		"event":          event,
		"details":        details,
	}).Info("Integration event")
}

// SanitizeForLog очищает строку для безопасного логирования
func SanitizeForLog(s string) string {
	// Удаляем потенциально чувствительные данные
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	
	// Ограничиваем длину
	if len(s) > 500 {
		s = s[:500] + "..."
	}
	
	return s
}
