package storage

import (
	"crypto/rand"
	"math/big"
)

const (
	// APIKeyPrefix is the prefix for all Goatway API keys
	APIKeyPrefix = "gw_"
	// APIKeyLength is the number of random characters after the prefix
	APIKeyLength = 64
	// APIKeyPrefixLen is the length of the identifying prefix (e.g., "gw_a1B2c3D4")
	APIKeyPrefixLen = 11 // "gw_" + 8 chars
)

// base62Alphabet contains characters for key generation (0-9, A-Z, a-z)
var base62Alphabet = []byte("0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")

// GenerateAPIKey creates a new API key with format: gw_ + 64 base62 chars
func GenerateAPIKey() (string, error) {
	result := make([]byte, APIKeyLength)
	alphabetLen := big.NewInt(int64(len(base62Alphabet)))

	for i := 0; i < APIKeyLength; i++ {
		idx, err := rand.Int(rand.Reader, alphabetLen)
		if err != nil {
			return "", err
		}
		result[i] = base62Alphabet[idx.Int64()]
	}

	return APIKeyPrefix + string(result), nil
}

// ExtractKeyPrefix returns the first 11 chars of a key for identification
// Format: "gw_" + first 8 random chars (e.g., "gw_a1B2c3D4")
func ExtractKeyPrefix(key string) string {
	if len(key) < APIKeyPrefixLen {
		return key
	}
	return key[:APIKeyPrefixLen]
}
