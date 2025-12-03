package services

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestRetryConfig_isRetryableStatus(t *testing.T) {
	config := DefaultRetryConfig()

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"429 Too Many Requests", 429, true},
		{"500 Internal Server Error", 500, true},
		{"502 Bad Gateway", 502, true},
		{"503 Service Unavailable", 503, true},
		{"504 Gateway Timeout", 504, true},
		{"200 OK", 200, false},
		{"400 Bad Request", 400, false},
		{"401 Unauthorized", 401, false},
		{"404 Not Found", 404, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.isRetryableStatus(tt.statusCode); got != tt.want {
				t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestRetryConfig_calculateDelay(t *testing.T) {
	config := RetryConfig{
		InitialDelay: 1 * time.Second,
		MaxDelay:     10 * time.Second,
		Multiplier:   2.0,
	}

	tests := []struct {
		name    string
		attempt int
		want    time.Duration
	}{
		{"First retry", 0, 1 * time.Second},
		{"Second retry", 1, 2 * time.Second},
		{"Third retry", 2, 4 * time.Second},
		{"Fourth retry", 3, 8 * time.Second},
		{"Fifth retry (capped)", 4, 10 * time.Second},
		{"Sixth retry (capped)", 5, 10 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := config.calculateDelay(tt.attempt); got != tt.want {
				t.Errorf("calculateDelay(%d) = %v, want %v", tt.attempt, got, tt.want)
			}
		})
	}
}

func TestRetryWithBackoff_Success(t *testing.T) {
	// Create test server that succeeds immediately
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}
	config := RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    100 * time.Millisecond,
		MaxDelay:        1 * time.Second,
		Multiplier:      2.0,
		RetryableStatus: []int{500, 502, 503},
	}

	resp, err := retryWithBackoff(req, client, config)
	if err != nil {
		t.Fatalf("Expected success, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestRetryWithBackoff_SuccessAfterRetries(t *testing.T) {
	var attemptCount int32

	// Create test server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte(`{"error":"service unavailable"}`))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"ok"}`))
		}
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}
	config := RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        500 * time.Millisecond,
		Multiplier:      2.0,
		RetryableStatus: []int{503},
	}

	start := time.Now()
	resp, err := retryWithBackoff(req, client, config)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// Should have waited at least 50ms + 100ms = 150ms
	if elapsed < 150*time.Millisecond {
		t.Errorf("Expected at least 150ms elapsed, got %v", elapsed)
	}
}

func TestRetryWithBackoff_AllRetriesFail(t *testing.T) {
	var attemptCount int32

	// Create test server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}
	config := RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        500 * time.Millisecond,
		Multiplier:      2.0,
		RetryableStatus: []int{503},
	}

	resp, err := retryWithBackoff(req, client, config)
	
	if err == nil {
		t.Fatal("Expected error after all retries failed")
	}

	// Response should be nil when all retries fail
	if resp != nil {
		defer resp.Body.Close()
	}

	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	if !strings.Contains(err.Error(), "failed after 3 attempts") {
		t.Errorf("Expected error message about 3 attempts, got: %v", err)
	}
}

func TestRetryWithBackoff_NonRetryableStatus(t *testing.T) {
	var attemptCount int32

	// Create test server that returns 400 (non-retryable)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"bad request"}`))
	}))
	defer server.Close()

	req, _ := http.NewRequest("GET", server.URL, nil)
	client := &http.Client{}
	config := RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        500 * time.Millisecond,
		Multiplier:      2.0,
		RetryableStatus: []int{500, 502, 503},
	}

	resp, err := retryWithBackoff(req, client, config)
	
	if err != nil {
		t.Fatalf("Expected no error for non-retryable status, got: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", resp.StatusCode)
	}

	// Should only attempt once (no retries for non-retryable status)
	if atomic.LoadInt32(&attemptCount) != 1 {
		t.Errorf("Expected 1 attempt, got %d", attemptCount)
	}
}

func TestRetryWithBackoff_WithRequestBody(t *testing.T) {
	var attemptCount int32
	var lastBody string

	// Create test server that fails once then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		
		// Read body
		body := make([]byte, r.ContentLength)
		r.Body.Read(body)
		lastBody = string(body)
		
		if count < 2 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	requestBody := `{"test":"data"}`
	req, _ := http.NewRequest("POST", server.URL, strings.NewReader(requestBody))
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{}
	config := RetryConfig{
		MaxAttempts:     3,
		InitialDelay:    50 * time.Millisecond,
		MaxDelay:        500 * time.Millisecond,
		Multiplier:      2.0,
		RetryableStatus: []int{503},
	}

	resp, err := retryWithBackoff(req, client, config)
	
	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}
	defer resp.Body.Close()

	if atomic.LoadInt32(&attemptCount) != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}

	// Verify body was sent correctly on retry
	if lastBody != requestBody {
		t.Errorf("Expected body %q, got %q", requestBody, lastBody)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected MaxAttempts=3, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 1*time.Second {
		t.Errorf("Expected InitialDelay=1s, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 30*time.Second {
		t.Errorf("Expected MaxDelay=30s, got %v", config.MaxDelay)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Expected Multiplier=2.0, got %f", config.Multiplier)
	}

	expectedStatuses := []int{429, 500, 502, 503, 504}
	if len(config.RetryableStatus) != len(expectedStatuses) {
		t.Errorf("Expected %d retryable statuses, got %d", len(expectedStatuses), len(config.RetryableStatus))
	}
}
