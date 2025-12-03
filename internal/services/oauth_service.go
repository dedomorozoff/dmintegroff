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

	// Execute request
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"error":          err.Error(),
		}).Error("Failed to request OAuth2 token")
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
