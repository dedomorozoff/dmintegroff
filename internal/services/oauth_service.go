package services

import (
	"dmintegroff/internal/database"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// OAuth2TokenResponse represents the response from OAuth2 token endpoint
type OAuth2TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// RetryConfig defines retry behavior with exponential backoff
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of retry attempts
	InitialDelay    time.Duration // Initial delay before first retry
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Exponential backoff multiplier
	RetryableStatus []int         // HTTP status codes that should trigger retry
}

// DefaultRetryConfig returns default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		RetryableStatus: []int{
			http.StatusTooManyRequests,      // 429
			http.StatusInternalServerError,  // 500
			http.StatusBadGateway,           // 502
			http.StatusServiceUnavailable,   // 503
			http.StatusGatewayTimeout,       // 504
		},
	}
}

// isRetryableStatus checks if the HTTP status code should trigger a retry
func (rc RetryConfig) isRetryableStatus(statusCode int) bool {
	for _, code := range rc.RetryableStatus {
		if code == statusCode {
			return true
		}
	}
	return false
}

// calculateDelay calculates the delay for the given attempt using exponential backoff
func (rc RetryConfig) calculateDelay(attempt int) time.Duration {
	delay := float64(rc.InitialDelay) * math.Pow(rc.Multiplier, float64(attempt))
	if delay > float64(rc.MaxDelay) {
		delay = float64(rc.MaxDelay)
	}
	return time.Duration(delay)
}

// retryWithBackoff executes an HTTP request with exponential backoff retry logic
func retryWithBackoff(req *http.Request, client *http.Client, config RetryConfig) (*http.Response, error) {
	var lastErr error
	var resp *http.Response
	
	// Save original body for retries
	var bodyBytes []byte
	if req.Body != nil {
		bodyBytes, _ = io.ReadAll(req.Body)
		req.Body.Close()
	}

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		// Restore body for each attempt
		if bodyBytes != nil {
			req.Body = io.NopCloser(strings.NewReader(string(bodyBytes)))
		}
		
		// Execute request
		resp, lastErr = client.Do(req)
		
		// Success - return immediately
		if lastErr == nil && !config.isRetryableStatus(resp.StatusCode) {
			return resp, nil
		}

		// Log retry attempt
		if lastErr != nil {
			logger.Log.WithFields(map[string]interface{}{
				"attempt": attempt + 1,
				"max":     config.MaxAttempts,
				"error":   lastErr.Error(),
				"url":     req.URL.String(),
			}).Warn("Request failed, will retry")
		} else {
			logger.Log.WithFields(map[string]interface{}{
				"attempt":     attempt + 1,
				"max":         config.MaxAttempts,
				"status_code": resp.StatusCode,
				"url":         req.URL.String(),
			}).Warn("Request returned retryable status, will retry")
			
			// Close response body before retry
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
		}

		// Don't sleep after last attempt
		if attempt < config.MaxAttempts-1 {
			delay := config.calculateDelay(attempt)
			logger.Log.WithFields(map[string]interface{}{
				"delay_seconds": delay.Seconds(),
			}).Debug("Waiting before retry")
			time.Sleep(delay)
		}
	}

	// All retries exhausted
	if lastErr != nil {
		return nil, fmt.Errorf("request failed after %d attempts: %w", config.MaxAttempts, lastErr)
	}
	
	// Return the last response with error
	return nil, fmt.Errorf("request failed after %d attempts with status %d", config.MaxAttempts, resp.StatusCode)
}

// GetAccessToken retrieves a valid access token for the integration
// It will use cached token if valid, or fetch a new one if expired
func GetAccessToken(integration *models.Integration) (string, error) {
	if integration.AuthType != "oauth2" {
		return "", nil // No OAuth2 configured
	}

	// Check if we have a valid cached token
	if integration.OAuth2AccessToken != "" && integration.OAuth2ExpiresAt > time.Now().Unix() {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
		}).Debug("Using cached OAuth2 access token")
		return integration.OAuth2AccessToken, nil
	}

	// Token expired or doesn't exist, fetch a new one
	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integration.ID,
	}).Info("Fetching new OAuth2 access token")

	return fetchNewAccessToken(integration)
}

// fetchNewAccessToken requests a new access token from the OAuth2 provider
func fetchNewAccessToken(integration *models.Integration) (string, error) {
	if integration.OAuth2TokenURL == "" {
		return "", errors.New("OAuth2 token URL not configured")
	}

	if integration.OAuth2ClientID == "" || integration.OAuth2ClientSecret == "" {
		return "", errors.New("OAuth2 client credentials not configured")
	}

	// Prepare request body based on grant type
	data := url.Values{}
	grantType := integration.OAuth2GrantType
	if grantType == "" {
		grantType = "client_credentials"
	}
	data.Set("grant_type", grantType)

	if integration.OAuth2Scope != "" {
		data.Set("scope", integration.OAuth2Scope)
	}

	// Create request
	req, err := http.NewRequest("POST", integration.OAuth2TokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	
	// Add client credentials (Basic Auth)
	auth := base64.StdEncoding.EncodeToString([]byte(
		integration.OAuth2ClientID + ":" + integration.OAuth2ClientSecret,
	))
	req.Header.Set("Authorization", "Basic "+auth)

	// Execute request with retry logic
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	
	retryConfig := DefaultRetryConfig()
	resp, err := retryWithBackoff(req, client, retryConfig)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"error":          err.Error(),
		}).Error("Failed to request OAuth2 token after retries")
		return "", fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"status_code":    resp.StatusCode,
			"response":       string(body),
		}).Error("OAuth2 token request failed")
		return "", fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var tokenResp OAuth2TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", errors.New("no access token in response")
	}

	// Calculate expiration time (with 60 second buffer)
	expiresAt := time.Now().Unix() + int64(tokenResp.ExpiresIn) - 60

	// Update integration with new token
	integration.OAuth2AccessToken = tokenResp.AccessToken
	integration.OAuth2RefreshToken = tokenResp.RefreshToken
	integration.OAuth2ExpiresAt = expiresAt

	if err := database.DB.Model(integration).Updates(map[string]interface{}{
		"oauth2_access_token":  tokenResp.AccessToken,
		"oauth2_refresh_token": tokenResp.RefreshToken,
		"oauth2_expires_at":    expiresAt,
	}).Error; err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"error":          err.Error(),
		}).Warn("Failed to save OAuth2 token to database")
		// Don't fail the request, we still have the token in memory
	}

	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integration.ID,
		"expires_in":     tokenResp.ExpiresIn,
	}).Info("Successfully obtained OAuth2 access token")

	return tokenResp.AccessToken, nil
}

// AddAuthHeaders adds authentication headers to the HTTP request based on integration config
func AddAuthHeaders(req *http.Request, integration *models.Integration) error {
	switch integration.AuthType {
	case "oauth2":
		token, err := GetAccessToken(integration)
		if err != nil {
			return fmt.Errorf("failed to get OAuth2 token: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		
	case "bearer":
		if integration.BearerToken == "" {
			return errors.New("bearer token not configured")
		}
		req.Header.Set("Authorization", "Bearer "+integration.BearerToken)
		
	case "basic":
		if integration.BasicAuthUser == "" || integration.BasicAuthPass == "" {
			return errors.New("basic auth credentials not configured")
		}
		auth := base64.StdEncoding.EncodeToString([]byte(
			integration.BasicAuthUser + ":" + integration.BasicAuthPass,
		))
		req.Header.Set("Authorization", "Basic "+auth)
		
	case "none", "":
		// No authentication
		
	default:
		return fmt.Errorf("unsupported auth type: %s", integration.AuthType)
	}
	
	return nil
}

// AddCustomHeaders adds custom HTTP headers to the request from integration config
func AddCustomHeaders(req *http.Request, integration *models.Integration) error {
	if integration.CustomHeaders == "" {
		logger.Log.Debug("No custom headers configured")
		return nil
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"custom_headers_raw": integration.CustomHeaders,
	}).Debug("Processing custom headers")
	
	var headers map[string]string
	if err := json.Unmarshal([]byte(integration.CustomHeaders), &headers); err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
			"raw":   integration.CustomHeaders,
		}).Error("Failed to parse custom headers")
		return fmt.Errorf("failed to parse custom headers: %w", err)
	}
	
	for name, value := range headers {
		if name != "" && value != "" {
			req.Header.Set(name, value)
			logger.Log.WithFields(map[string]interface{}{
				"header": name,
				"value":  value,
			}).Debug("Added custom header")
		}
	}
	
	logger.Log.WithFields(map[string]interface{}{
		"count": len(headers),
	}).Info("Custom headers applied")
	
	return nil
}

// TestOAuth2Connection tests the OAuth2 configuration by attempting to get a token
func TestOAuth2Connection(integration *models.Integration) error {
	if integration.AuthType != "oauth2" {
		return errors.New("integration is not configured for OAuth2")
	}

	// Clear cached token to force a fresh request
	integration.OAuth2AccessToken = ""
	integration.OAuth2ExpiresAt = 0

	_, err := fetchNewAccessToken(integration)
	return err
}
