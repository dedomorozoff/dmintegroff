package services

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"dmintegroff/internal/logger"
	"dmintegroff/internal/models"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net/http"
	"strings"
)

// SignatureAlgorithm represents supported HMAC algorithms
type SignatureAlgorithm string

const (
	AlgorithmSHA256 SignatureAlgorithm = "sha256"
	AlgorithmSHA512 SignatureAlgorithm = "sha512"
	AlgorithmSHA1   SignatureAlgorithm = "sha1"
)

// GenerateSignature generates HMAC signature for the given payload
func GenerateSignature(payload []byte, secret string, algorithm SignatureAlgorithm) (string, error) {
	if secret == "" {
		return "", errors.New("signature secret is empty")
	}

	var h hash.Hash

	switch algorithm {
	case AlgorithmSHA256:
		h = hmac.New(sha256.New, []byte(secret))
	case AlgorithmSHA512:
		h = hmac.New(sha512.New, []byte(secret))
	case AlgorithmSHA1:
		h = hmac.New(sha1.New, []byte(secret))
	default:
		return "", fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	h.Write(payload)
	signature := hex.EncodeToString(h.Sum(nil))

	return signature, nil
}

// VerifySignature verifies HMAC signature from the request
func VerifySignature(payload []byte, receivedSignature string, secret string, algorithm SignatureAlgorithm) bool {
	expectedSignature, err := GenerateSignature(payload, secret, algorithm)
	if err != nil {
		logger.Log.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to generate signature for verification")
		return false
	}

	// Use constant-time comparison to prevent timing attacks
	return hmac.Equal([]byte(receivedSignature), []byte(expectedSignature))
}

// AddSignatureToRequest adds webhook signature to outgoing HTTP request
func AddSignatureToRequest(req *http.Request, payload []byte, integration *models.Integration) error {
	if !integration.WebhookSignatureEnabled {
		return nil // Signature not enabled
	}

	if integration.WebhookSignatureSecret == "" {
		return errors.New("webhook signature enabled but secret is empty")
	}

	algorithm := SignatureAlgorithm(integration.WebhookSignatureAlgorithm)
	if algorithm == "" {
		algorithm = AlgorithmSHA256 // Default
	}

	signature, err := GenerateSignature(payload, integration.WebhookSignatureSecret, algorithm)
	if err != nil {
		return fmt.Errorf("failed to generate signature: %w", err)
	}

	headerName := integration.WebhookSignatureHeader
	if headerName == "" {
		headerName = "X-Webhook-Signature"
	}

	// Add signature with algorithm prefix (e.g., "sha256=abc123...")
	req.Header.Set(headerName, fmt.Sprintf("%s=%s", algorithm, signature))

	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integration.ID,
		"header":         headerName,
		"algorithm":      algorithm,
	}).Debug("Added webhook signature to request")

	return nil
}

// VerifyIncomingSignature verifies signature from incoming webhook request
func VerifyIncomingSignature(req *http.Request, payload []byte, integration *models.Integration) error {
	if !integration.WebhookSignatureEnabled {
		return nil // Signature verification not enabled
	}

	if integration.WebhookSignatureSecret == "" {
		return errors.New("webhook signature enabled but secret is not configured")
	}

	headerName := integration.WebhookSignatureHeader
	if headerName == "" {
		headerName = "X-Webhook-Signature"
	}

	receivedSignature := req.Header.Get(headerName)
	if receivedSignature == "" {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"header":         headerName,
		}).Warn("Webhook signature header missing")
		return fmt.Errorf("signature header '%s' is missing", headerName)
	}

	// Parse algorithm from signature (format: "sha256=abc123...")
	var algorithm SignatureAlgorithm
	var signatureValue string

	if strings.Contains(receivedSignature, "=") {
		parts := strings.SplitN(receivedSignature, "=", 2)
		algorithm = SignatureAlgorithm(parts[0])
		signatureValue = parts[1]
	} else {
		// No algorithm prefix, use configured algorithm
		algorithm = SignatureAlgorithm(integration.WebhookSignatureAlgorithm)
		if algorithm == "" {
			algorithm = AlgorithmSHA256
		}
		signatureValue = receivedSignature
	}

	// Verify signature
	if !VerifySignature(payload, signatureValue, integration.WebhookSignatureSecret, algorithm) {
		logger.Log.WithFields(map[string]interface{}{
			"integration_id": integration.ID,
			"algorithm":      algorithm,
		}).Warn("Webhook signature verification failed")
		return errors.New("invalid webhook signature")
	}

	logger.Log.WithFields(map[string]interface{}{
		"integration_id": integration.ID,
		"algorithm":      algorithm,
	}).Debug("Webhook signature verified successfully")

	return nil
}

// GenerateRandomSecret generates a random secret for webhook signatures
func GenerateRandomSecret(length int) (string, error) {
	if length < 16 {
		length = 32 // Minimum secure length
	}

	// Use crypto/rand for secure random generation
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random secret: %w", err)
	}

	return hex.EncodeToString(bytes), nil
}

// ValidateSignatureConfig validates webhook signature configuration
func ValidateSignatureConfig(integration *models.Integration) error {
	if !integration.WebhookSignatureEnabled {
		return nil // Not enabled, no validation needed
	}

	if integration.WebhookSignatureSecret == "" {
		return errors.New("webhook signature secret is required when signature is enabled")
	}

	if len(integration.WebhookSignatureSecret) < 16 {
		return errors.New("webhook signature secret must be at least 16 characters")
	}

	algorithm := SignatureAlgorithm(integration.WebhookSignatureAlgorithm)
	if algorithm != "" && algorithm != AlgorithmSHA256 && algorithm != AlgorithmSHA512 && algorithm != AlgorithmSHA1 {
		return fmt.Errorf("unsupported signature algorithm: %s", algorithm)
	}

	return nil
}
