package upload

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateToken generates a random token string.
func GenerateToken() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
