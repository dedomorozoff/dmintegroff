package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken generates a random hex token of specified byte length
func GenerateToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
