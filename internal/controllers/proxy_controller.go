package controllers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// ProxyRequest represents the request structure for proxying
type ProxyRequest struct {
	URL         string            `json:"url" binding:"required"`
	Method      string            `json:"method" binding:"required"`
	Headers     map[string]string `json:"headers"`
	Body        string            `json:"body"`
	QueryParams map[string]string `json:"query_params"`
}

// ProxyResponse represents the response structure
type ProxyResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	Duration   int64             `json:"duration_ms"`
	Error      string            `json:"error,omitempty"`
}

// ProxyHTTPRequest proxies HTTP requests to avoid CORS issues
func ProxyHTTPRequest(c *gin.Context) {
	// Check authentication
	session := sessions.Default(c)
	userID := session.Get("user_id")
	if userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	var req ProxyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Validate URL
	parsedURL, err := url.Parse(req.URL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid URL format"})
		return
	}

	// Security check - prevent requests to local/private networks, except our own server
	if isLocalOrPrivateIP(parsedURL.Hostname()) && !isOwnServer(c, parsedURL) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Requests to local/private networks are not allowed"})
		return
	}

	// Add query parameters to URL
	if len(req.QueryParams) > 0 {
		query := parsedURL.Query()
		for key, value := range req.QueryParams {
			query.Add(key, value)
		}
		parsedURL.RawQuery = query.Encode()
	}

	// Prepare request body
	var bodyReader io.Reader
	if req.Body != "" && (req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH") {
		bodyReader = strings.NewReader(req.Body)
	}

	// Create HTTP request
	httpReq, err := http.NewRequest(req.Method, parsedURL.String(), bodyReader)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Set headers
	if req.Headers != nil {
		for key, value := range req.Headers {
			httpReq.Header.Set(key, value)
		}
	}

	// Always set User-Agent if not provided
	if httpReq.Header.Get("User-Agent") == "" {
		httpReq.Header.Set("User-Agent", "DmIntegroff")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Record start time
	startTime := time.Now()

	// Send request
	resp, err := client.Do(httpReq)
	if err != nil {
		duration := time.Since(startTime).Milliseconds()
		c.JSON(http.StatusOK, ProxyResponse{
			Status:   0,
			Error:    fmt.Sprintf("Request failed: %v", err),
			Duration: duration,
		})
		return
	}
	defer resp.Body.Close()

	// Record duration
	duration := time.Since(startTime).Milliseconds()

	// Read response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusOK, ProxyResponse{
			Status:     resp.StatusCode,
			StatusText: resp.Status,
			Error:      fmt.Sprintf("Failed to read response body: %v", err),
			Duration:   duration,
		})
		return
	}

	// Collect response headers
	responseHeaders := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			responseHeaders[key] = values[0]
		}
	}

	// Return response
	c.JSON(http.StatusOK, ProxyResponse{
		Status:     resp.StatusCode,
		StatusText: resp.Status,
		Headers:    responseHeaders,
		Body:       string(bodyBytes),
		Duration:   duration,
	})
}

// isLocalOrPrivateIP checks if the hostname is a local or private IP
func isLocalOrPrivateIP(hostname string) bool {
	// Basic check for localhost and private networks
	if hostname == "localhost" || hostname == "127.0.0.1" || hostname == "::1" {
		return true
	}
	
	// Check for private IP ranges (basic implementation)
	if strings.HasPrefix(hostname, "192.168.") ||
		strings.HasPrefix(hostname, "10.") ||
		strings.HasPrefix(hostname, "172.16.") ||
		strings.HasPrefix(hostname, "172.17.") ||
		strings.HasPrefix(hostname, "172.18.") ||
		strings.HasPrefix(hostname, "172.19.") ||
		strings.HasPrefix(hostname, "172.20.") ||
		strings.HasPrefix(hostname, "172.21.") ||
		strings.HasPrefix(hostname, "172.22.") ||
		strings.HasPrefix(hostname, "172.23.") ||
		strings.HasPrefix(hostname, "172.24.") ||
		strings.HasPrefix(hostname, "172.25.") ||
		strings.HasPrefix(hostname, "172.26.") ||
		strings.HasPrefix(hostname, "172.27.") ||
		strings.HasPrefix(hostname, "172.28.") ||
		strings.HasPrefix(hostname, "172.29.") ||
		strings.HasPrefix(hostname, "172.30.") ||
		strings.HasPrefix(hostname, "172.31.") {
		return true
	}
	
	return false
}

// isOwnServer checks if the URL points to our own server
func isOwnServer(c *gin.Context, parsedURL *url.URL) bool {
	// Get the host from the current request
	currentHost := c.Request.Host
	targetHost := parsedURL.Host
	
	// If no port specified in target, add default port
	if !strings.Contains(targetHost, ":") {
		if parsedURL.Scheme == "https" {
			targetHost += ":443"
		} else {
			targetHost += ":80"
		}
	}
	
	// If no port specified in current host, add default port
	if !strings.Contains(currentHost, ":") {
		currentHost += ":80" // Assume HTTP for development
	}
	
	// Direct match
	if currentHost == targetHost {
		return true
	}
	
	// Check if target is localhost/127.0.0.1 and current is localhost
	if (parsedURL.Hostname() == "localhost" || parsedURL.Hostname() == "127.0.0.1") &&
		(strings.HasPrefix(currentHost, "localhost") || strings.HasPrefix(currentHost, "127.0.0.1")) {
		// Check if ports match
		currentPort := "80"
		targetPort := "80"
		
		if strings.Contains(currentHost, ":") {
			parts := strings.Split(currentHost, ":")
			if len(parts) > 1 {
				currentPort = parts[1]
			}
		}
		
		if strings.Contains(targetHost, ":") {
			parts := strings.Split(targetHost, ":")
			if len(parts) > 1 {
				targetPort = parts[1]
			}
		}
		
		return currentPort == targetPort
	}
	
	return false
}