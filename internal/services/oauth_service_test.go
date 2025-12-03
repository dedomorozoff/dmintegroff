package services

import (
	"dmintegroff/internal/models"
	"net/http"
	"testing"
)

func TestAddAuthHeaders_None(t *testing.T) {
	integration := &models.Integration{
		AuthType: "none",
	}

	req, _ := createTestRequest()
	err := AddAuthHeaders(req, integration)

	if err != nil {
		t.Errorf("Expected no error for auth_type=none, got: %v", err)
	}

	if req.Header.Get("Authorization") != "" {
		t.Error("Expected no Authorization header for auth_type=none")
	}
}

func TestAddAuthHeaders_Bearer(t *testing.T) {
	integration := &models.Integration{
		AuthType:    "bearer",
		BearerToken: "test_token_123",
	}

	req, _ := createTestRequest()
	err := AddAuthHeaders(req, integration)

	if err != nil {
		t.Errorf("Expected no error for bearer token, got: %v", err)
	}

	expected := "Bearer test_token_123"
	if req.Header.Get("Authorization") != expected {
		t.Errorf("Expected Authorization header '%s', got '%s'", expected, req.Header.Get("Authorization"))
	}
}

func TestAddAuthHeaders_Basic(t *testing.T) {
	integration := &models.Integration{
		AuthType:      "basic",
		BasicAuthUser: "user",
		BasicAuthPass: "pass",
	}

	req, _ := createTestRequest()
	err := AddAuthHeaders(req, integration)

	if err != nil {
		t.Errorf("Expected no error for basic auth, got: %v", err)
	}

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Error("Expected Authorization header for basic auth")
	}

	// Should start with "Basic "
	if len(auth) < 6 || auth[:6] != "Basic " {
		t.Errorf("Expected Authorization header to start with 'Basic ', got '%s'", auth)
	}
}

func TestAddAuthHeaders_BearerMissingToken(t *testing.T) {
	integration := &models.Integration{
		AuthType:    "bearer",
		BearerToken: "",
	}

	req, _ := createTestRequest()
	err := AddAuthHeaders(req, integration)

	if err == nil {
		t.Error("Expected error for missing bearer token")
	}
}

func TestAddAuthHeaders_BasicMissingCredentials(t *testing.T) {
	integration := &models.Integration{
		AuthType:      "basic",
		BasicAuthUser: "",
		BasicAuthPass: "",
	}

	req, _ := createTestRequest()
	err := AddAuthHeaders(req, integration)

	if err == nil {
		t.Error("Expected error for missing basic auth credentials")
	}
}

func TestAddAuthHeaders_UnsupportedType(t *testing.T) {
	integration := &models.Integration{
		AuthType: "unsupported_type",
	}

	req, _ := createTestRequest()
	err := AddAuthHeaders(req, integration)

	if err == nil {
		t.Error("Expected error for unsupported auth type")
	}
}

// Helper function to create a test HTTP request
func createTestRequest() (*http.Request, error) {
	return http.NewRequest("POST", "https://example.com/api", nil)
}
