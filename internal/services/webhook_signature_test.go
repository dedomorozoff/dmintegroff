package services

import (
	"bytes"
	"dmintegroff/internal/models"
	"net/http"
	"strings"
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "my-secret-key"

	tests := []struct {
		name      string
		algorithm SignatureAlgorithm
		wantErr   bool
	}{
		{"SHA256", AlgorithmSHA256, false},
		{"SHA512", AlgorithmSHA512, false},
		{"SHA1", AlgorithmSHA1, false},
		{"Invalid", SignatureAlgorithm("invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := GenerateSignature(payload, secret, tt.algorithm)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if signature == "" {
				t.Error("Expected non-empty signature")
			}
			
			// Verify signature length based on algorithm
			switch tt.algorithm {
			case AlgorithmSHA256:
				if len(signature) != 64 { // 32 bytes = 64 hex chars
					t.Errorf("Expected SHA256 signature length 64, got %d", len(signature))
				}
			case AlgorithmSHA512:
				if len(signature) != 128 { // 64 bytes = 128 hex chars
					t.Errorf("Expected SHA512 signature length 128, got %d", len(signature))
				}
			case AlgorithmSHA1:
				if len(signature) != 40 { // 20 bytes = 40 hex chars
					t.Errorf("Expected SHA1 signature length 40, got %d", len(signature))
				}
			}
		})
	}
}

func TestGenerateSignature_EmptySecret(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	_, err := GenerateSignature(payload, "", AlgorithmSHA256)
	
	if err == nil {
		t.Error("Expected error for empty secret")
	}
}

func TestVerifySignature(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "my-secret-key"
	
	// Generate valid signature
	validSignature, _ := GenerateSignature(payload, secret, AlgorithmSHA256)
	
	tests := []struct {
		name      string
		signature string
		secret    string
		want      bool
	}{
		{"Valid signature", validSignature, secret, true},
		{"Invalid signature", "invalid", secret, false},
		{"Wrong secret", validSignature, "wrong-secret", false},
		{"Empty signature", "", secret, false},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifySignature(payload, tt.signature, tt.secret, AlgorithmSHA256)
			if result != tt.want {
				t.Errorf("VerifySignature() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestVerifySignature_DifferentAlgorithms(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "my-secret-key"
	
	algorithms := []SignatureAlgorithm{AlgorithmSHA256, AlgorithmSHA512, AlgorithmSHA1}
	
	for _, algo := range algorithms {
		t.Run(string(algo), func(t *testing.T) {
			signature, err := GenerateSignature(payload, secret, algo)
			if err != nil {
				t.Fatalf("Failed to generate signature: %v", err)
			}
			
			if !VerifySignature(payload, signature, secret, algo) {
				t.Error("Signature verification failed")
			}
		})
	}
}

func TestAddSignatureToRequest(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	
	tests := []struct {
		name        string
		integration *models.Integration
		wantHeader  bool
		wantErr     bool
	}{
		{
			name: "Signature enabled",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    "my-secret-key",
				WebhookSignatureHeader:    "X-Webhook-Signature",
				WebhookSignatureAlgorithm: "sha256",
			},
			wantHeader: true,
			wantErr:    false,
		},
		{
			name: "Signature disabled",
			integration: &models.Integration{
				WebhookSignatureEnabled: false,
			},
			wantHeader: false,
			wantErr:    false,
		},
		{
			name: "Signature enabled but no secret",
			integration: &models.Integration{
				WebhookSignatureEnabled: true,
				WebhookSignatureSecret:  "",
			},
			wantHeader: false,
			wantErr:    true,
		},
		{
			name: "Custom header name",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    "my-secret-key",
				WebhookSignatureHeader:    "X-Custom-Signature",
				WebhookSignatureAlgorithm: "sha512",
			},
			wantHeader: true,
			wantErr:    false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "https://example.com/webhook", bytes.NewReader(payload))
			
			err := AddSignatureToRequest(req, payload, tt.integration)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if tt.wantHeader {
				headerName := tt.integration.WebhookSignatureHeader
				if headerName == "" {
					headerName = "X-Webhook-Signature"
				}
				
				signature := req.Header.Get(headerName)
				if signature == "" {
					t.Errorf("Expected signature header '%s' but got none", headerName)
				}
				
				// Verify signature format (algorithm=signature)
				if !strings.Contains(signature, "=") {
					t.Error("Expected signature format 'algorithm=signature'")
				}
			}
		})
	}
}

func TestVerifyIncomingSignature(t *testing.T) {
	payload := []byte(`{"test":"data"}`)
	secret := "my-secret-key"
	
	// Generate valid signature
	validSignature, _ := GenerateSignature(payload, secret, AlgorithmSHA256)
	
	tests := []struct {
		name        string
		integration *models.Integration
		signature   string
		wantErr     bool
	}{
		{
			name: "Valid signature",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    secret,
				WebhookSignatureHeader:    "X-Webhook-Signature",
				WebhookSignatureAlgorithm: "sha256",
			},
			signature: "sha256=" + validSignature,
			wantErr:   false,
		},
		{
			name: "Invalid signature",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    secret,
				WebhookSignatureHeader:    "X-Webhook-Signature",
				WebhookSignatureAlgorithm: "sha256",
			},
			signature: "sha256=invalid",
			wantErr:   true,
		},
		{
			name: "Missing signature header",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    secret,
				WebhookSignatureHeader:    "X-Webhook-Signature",
				WebhookSignatureAlgorithm: "sha256",
			},
			signature: "",
			wantErr:   true,
		},
		{
			name: "Signature disabled",
			integration: &models.Integration{
				WebhookSignatureEnabled: false,
			},
			signature: "",
			wantErr:   false,
		},
		{
			name: "Signature without algorithm prefix",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    secret,
				WebhookSignatureHeader:    "X-Webhook-Signature",
				WebhookSignatureAlgorithm: "sha256",
			},
			signature: validSignature,
			wantErr:   false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("POST", "https://example.com/webhook", bytes.NewReader(payload))
			
			if tt.signature != "" {
				headerName := tt.integration.WebhookSignatureHeader
				if headerName == "" {
					headerName = "X-Webhook-Signature"
				}
				req.Header.Set(headerName, tt.signature)
			}
			
			err := VerifyIncomingSignature(req, payload, tt.integration)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestGenerateRandomSecret(t *testing.T) {
	tests := []struct {
		name   string
		length int
		want   int // Expected hex string length
	}{
		{"Default length", 32, 64},  // 32 bytes = 64 hex chars
		{"Custom length", 16, 32},   // 16 bytes = 32 hex chars
		{"Small length (auto-adjusted)", 8, 64}, // Should use minimum 32 bytes
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			secret, err := GenerateRandomSecret(tt.length)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			
			if len(secret) != tt.want {
				t.Errorf("Expected secret length %d, got %d", tt.want, len(secret))
			}
			
			// Verify it's valid hex
			for _, c := range secret {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Errorf("Secret contains non-hex character: %c", c)
					break
				}
			}
		})
	}
}

func TestGenerateRandomSecret_Uniqueness(t *testing.T) {
	secret1, _ := GenerateRandomSecret(32)
	secret2, _ := GenerateRandomSecret(32)
	
	if secret1 == secret2 {
		t.Error("Generated secrets should be unique")
	}
}

func TestValidateSignatureConfig(t *testing.T) {
	tests := []struct {
		name        string
		integration *models.Integration
		wantErr     bool
	}{
		{
			name: "Valid config",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    "my-very-secure-secret-key",
				WebhookSignatureAlgorithm: "sha256",
			},
			wantErr: false,
		},
		{
			name: "Disabled signature",
			integration: &models.Integration{
				WebhookSignatureEnabled: false,
			},
			wantErr: false,
		},
		{
			name: "Missing secret",
			integration: &models.Integration{
				WebhookSignatureEnabled: true,
				WebhookSignatureSecret:  "",
			},
			wantErr: true,
		},
		{
			name: "Secret too short",
			integration: &models.Integration{
				WebhookSignatureEnabled: true,
				WebhookSignatureSecret:  "short",
			},
			wantErr: true,
		},
		{
			name: "Invalid algorithm",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    "my-very-secure-secret-key",
				WebhookSignatureAlgorithm: "md5",
			},
			wantErr: true,
		},
		{
			name: "Valid SHA512",
			integration: &models.Integration{
				WebhookSignatureEnabled:   true,
				WebhookSignatureSecret:    "my-very-secure-secret-key",
				WebhookSignatureAlgorithm: "sha512",
			},
			wantErr: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSignatureConfig(tt.integration)
			
			if tt.wantErr {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}
