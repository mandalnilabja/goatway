package storage

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

// AESEncryptor implements AES-256-GCM encryption
type AESEncryptor struct {
	key []byte
}

// NewEncryptor creates a new AES encryptor with a derived key
// Priority: GOATWAY_ENCRYPTION_KEY env var > machine-derived key
func NewEncryptor() (*AESEncryptor, error) {
	var keyMaterial string

	if envKey := os.Getenv("GOATWAY_ENCRYPTION_KEY"); envKey != "" {
		keyMaterial = envKey
	} else {
		keyMaterial = deriveMachineKey()
	}

	// Derive a 256-bit key using SHA-256
	hash := sha256.Sum256([]byte(keyMaterial))
	return &AESEncryptor{key: hash[:]}, nil
}

// NewEncryptorWithKey creates an encryptor with a specific key (for testing)
func NewEncryptorWithKey(key []byte) (*AESEncryptor, error) {
	if len(key) != 32 {
		return nil, errors.New("key must be 32 bytes for AES-256")
	}
	return &AESEncryptor{key: key}, nil
}

// Encrypt encrypts plaintext using AES-256-GCM
func (e *AESEncryptor) Encrypt(plaintext string) (string, error) {
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
func (e *AESEncryptor) Decrypt(ciphertext string) (string, error) {
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
