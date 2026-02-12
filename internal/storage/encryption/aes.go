// Package encryption provides AES-256-GCM encryption for sensitive data.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"runtime"
)

// Encryptor provides encryption/decryption for sensitive data
type Encryptor interface {
	Encrypt(plaintext string) (string, error)
	Decrypt(ciphertext string) (string, error)
}

// AES implements AES-256-GCM encryption
type AES struct {
	key []byte
}

// New creates a new AES encryptor with a derived key
// Priority: GOATWAY_ENCRYPTION_KEY env var > machine-derived key
func New() (*AES, error) {
	var keyMaterial string

	if envKey := os.Getenv("GOATWAY_ENCRYPTION_KEY"); envKey != "" {
		keyMaterial = envKey
	} else {
		keyMaterial = deriveMachineKey()
	}

	// Derive a 256-bit key using SHA-256
	hash := sha256.Sum256([]byte(keyMaterial))
	return &AES{key: hash[:]}, nil
}

// NewWithKey creates an encryptor with a specific key (for testing)
func NewWithKey(key []byte) (*AES, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	return &AES{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *AES) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts ciphertext using AES-256-GCM
func (e *AES) Decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	nonce, ciphertextBytes := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertextBytes, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// deriveMachineKey creates a machine-specific key from available identifiers
func deriveMachineKey() string {
	// Combine multiple sources for a machine-specific key
	// This provides basic protection without requiring user configuration
	material := "goatway-default-key"

	// Add hostname
	if hostname, err := os.Hostname(); err == nil {
		material += hostname
	}

	// Add user home directory
	if home, err := os.UserHomeDir(); err == nil {
		material += home
	}

	// Add OS/arch info
	material += runtime.GOOS + runtime.GOARCH

	return material
}
