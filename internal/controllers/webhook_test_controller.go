package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"dmintegroff/internal/cache"
	"dmintegroff/internal/database"
	"dmintegroff/internal/models"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CreateWebhookTest creates a new test webhook
func CreateWebhookTest(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")

	// Generate unique token
	token := generateWebhookTestToken()
	expiresAt := time.Now().Add(24 * time.Hour)

	// Try Redis first
	if cache.IsRedisAvailable() {
		data := cache.WebhookTestData{
			Token:     token,
			UserID:    userID.(uint),
			ExpiresAt: expiresAt,
			IsActive:  true,
			CreatedAt: time.Now(),
		}

		if err := cache.SaveWebhookTest(token, data); err != nil {
			fmt.Printf("Redis error, falling back to DB: %v\n", err)
		} else {
			// Build webhook URL
			baseURL := os.Getenv("BASE_URL")
			if baseURL == "" {
				baseURL = "http://localhost:8080"
			}
			webhookURL := fmt.Sprintf("%s/webhook/test/%s", baseURL, token)

			c.JSON(http.StatusOK, gin.H{
				"id":          0, // Redis doesn't use DB ID
				"token":       token,
				"webhook_url": webhookURL,
				"expires_at":  expiresAt,
				"storage":     "redis",
			})
			return
		}
	}

	// Fallback to database
	webhookTest := models.WebhookTest{
		Token:     token,
		UserID:    userID.(uint),
		ExpiresAt: expiresAt,
		IsActive:  true,
	}

	if err := database.DB.Create(&webhookTest).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create test webhook"})
		return
	}

	// Build webhook URL
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	webhookURL := fmt.Sprintf("%s/webhook/test/%s", baseURL, token)

	c.JSON(http.StatusOK, gin.H{
		"id":          webhookTest.ID,
		"token":       token,
		"webhook_url": webhookURL,
		"expires_at":  webhookTest.ExpiresAt,
		"storage":     "database",
	})
}

// WebhookTestPage displays the test webhook page
func WebhookTestPage(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	username := session.Get("username")
	role := session.Get("role")
	token := c.Param("token")

	// Check that webhook belongs to user
	var webhookTest models.WebhookTest
	if err := database.DB.Where("token = ? AND user_id = ?", token, userID).First(&webhookTest).Error; err != nil {
		c.HTML(http.StatusNotFound, "pages/404.html", gin.H{
			"title": "Not Found",
		})
		return
	}

	// Check if expired
	if time.Now().After(webhookTest.ExpiresAt) || !webhookTest.IsActive {
		c.HTML(http.StatusGone, "pages/webhook_test.html", gin.H{
			"title":       "Test Webhook",
			"username":    username,
			"role":        role,
			"CurrentPage": "dashboard",
			"expired":     true,
		})
		return
	}

	// Get all requests to this webhook
	var requests []models.WebhookTestRequest
	database.DB.Where("webhook_test_id = ?", webhookTest.ID).
		Order("created_at desc").
		Limit(100).
		Find(&requests)

	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	webhookURL := fmt.Sprintf("%s/webhook/test/%s", baseURL, token)

	c.HTML(http.StatusOK, "pages/webhook_test.html", gin.H{
		"title":        "Test Webhook",
		"username":     username,
		"role":         role,
		"CurrentPage":  "dashboard",
		"webhook_test": webhookTest,
		"webhook_url":  webhookURL,
		"requests":     requests,
		"token":        token,
	})
}

// HandleWebhookTest handles incoming requests to test webhook
func HandleWebhookTest(c *gin.Context) {
	token := c.Param("token")

	// Read request body
	body, _ := io.ReadAll(c.Request.Body)

	// Collect headers
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}
	headersJSON, _ := json.Marshal(headers)

	// Collect query params
	queryParams := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			queryParams[key] = values[0]
		}
	}
	queryParamsJSON, _ := json.Marshal(queryParams)

	// Try Redis first
	if cache.IsRedisAvailable() {
		webhookData, err := cache.GetWebhookTest(token)
		if err == nil && webhookData != nil {
			// Check expiration
			if time.Now().After(webhookData.ExpiresAt) || !webhookData.IsActive {
				c.JSON(http.StatusGone, gin.H{"error": "Webhook expired"})
				return
			}

			// Save request to Redis
			requestData := cache.WebhookRequestData{
				ID:          uuid.New().String(),
				Method:      c.Request.Method,
				URL:         c.Request.URL.String(),
				Headers:     string(headersJSON),
				Body:        string(body),
				QueryParams: string(queryParamsJSON),
				ClientIP:    c.ClientIP(),
				CreatedAt:   time.Now(),
			}

			if err := cache.SaveWebhookRequest(token, requestData); err != nil {
				fmt.Printf("Redis error saving request: %v\n", err)
			} else {
				c.JSON(http.StatusOK, gin.H{
					"status":  "success",
					"message": "Request received and logged",
					"id":      requestData.ID,
					"storage": "redis",
				})
				return
			}
		}
	}

	// Fallback to database
	var webhookTest models.WebhookTest
	if err := database.DB.Where("token = ?", token).First(&webhookTest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// Check expiration
	if time.Now().After(webhookTest.ExpiresAt) || !webhookTest.IsActive {
		c.JSON(http.StatusGone, gin.H{"error": "Webhook expired"})
		return
	}

	// Save request to database
	request := models.WebhookTestRequest{
		WebhookTestID: webhookTest.ID,
		Method:        c.Request.Method,
		URL:           c.Request.URL.String(),
		Headers:       string(headersJSON),
		Body:          string(body),
		QueryParams:   string(queryParamsJSON),
		ClientIP:      c.ClientIP(),
	}

	if err := database.DB.Create(&request).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save request"})
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"status":  "success",
		"message": "Request received and logged",
		"id":      request.ID,
		"storage": "database",
	})
}

// GetWebhookTestRequests returns list of requests to test webhook (API)
func GetWebhookTestRequests(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	token := c.Param("token")

	// Try Redis first
	if cache.IsRedisAvailable() {
		webhookData, err := cache.GetWebhookTest(token)
		if err == nil && webhookData != nil && webhookData.UserID == userID.(uint) {
			requests, err := cache.GetWebhookRequests(token, 100)
			if err == nil {
				c.JSON(http.StatusOK, gin.H{
					"requests": requests,
					"storage":  "redis",
				})
				return
			}
			fmt.Printf("Redis error getting requests: %v\n", err)
		}
	}

	// Fallback to database
	var webhookTest models.WebhookTest
	if err := database.DB.Where("token = ? AND user_id = ?", token, userID).First(&webhookTest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// Get requests
	var requests []models.WebhookTestRequest
	database.DB.Where("webhook_test_id = ?", webhookTest.ID).
		Order("created_at desc").
		Limit(100).
		Find(&requests)

	c.JSON(http.StatusOK, gin.H{
		"requests": requests,
		"storage":  "database",
	})
}

// DeleteWebhookTest deletes test webhook
func DeleteWebhookTest(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	token := c.Param("token")

	// Check that webhook belongs to user
	var webhookTest models.WebhookTest
	if err := database.DB.Where("token = ? AND user_id = ?", token, userID).First(&webhookTest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// Delete all requests
	database.DB.Where("webhook_test_id = ?", webhookTest.ID).Delete(&models.WebhookTestRequest{})

	// Delete webhook
	database.DB.Delete(&webhookTest)

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

// generateWebhookTestToken generates random token for test webhook
func generateWebhookTestToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// UseRequestAsSample saves a webhook test request as sample payload for an integration
func UseRequestAsSample(c *gin.Context) {
	session := sessions.Default(c)
	userID := session.Get("user_id")
	token := c.Param("token")
	requestID := c.Param("request_id")

	// Check that webhook belongs to user
	var webhookTest models.WebhookTest
	if err := database.DB.Where("token = ? AND user_id = ?", token, userID).First(&webhookTest).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Webhook not found"})
		return
	}

	// Get the request
	var request models.WebhookTestRequest
	if err := database.DB.Where("id = ? AND webhook_test_id = ?", requestID, webhookTest.ID).First(&request).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Request not found"})
		return
	}

	// Get integration ID from request body
	var requestBody struct {
		IntegrationID uint `json:"integration_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&requestBody); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Integration ID is required"})
		return
	}

	// Check that integration belongs to user
	var integration models.Integration
	if err := database.DB.Where("id = ? AND created_by_id = ?", requestBody.IntegrationID, userID).First(&integration).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Integration not found"})
		return
	}

	// Update integration with sample payload
	integration.SamplePayload = request.Body
	if err := database.DB.Save(&integration).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save sample payload"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":         "success",
		"message":        "Request saved as sample payload",
		"integration_id": integration.ID,
		"redirect_url":   fmt.Sprintf("/integrations/%d/configure", integration.ID),
	})
}

// CleanupExpiredWebhooks removes expired test webhooks and their requests
func CleanupExpiredWebhooks() {
	// Delete expired webhooks (older than 24 hours)
	result := database.DB.Where("expires_at < ? OR (is_active = ? AND created_at < ?)", 
		time.Now(), 
		false, 
		time.Now().Add(-24*time.Hour)).
		Delete(&models.WebhookTest{})
	
	if result.Error != nil {
		fmt.Printf("Error cleaning up expired webhooks: %v\n", result.Error)
	} else if result.RowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired test webhooks\n", result.RowsAffected)
	}
}
