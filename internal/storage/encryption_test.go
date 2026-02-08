package storage

import (
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Create encryptor with a known key for testing
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc, err := NewEncryptorWithKey(key)
	if err != nil {
		t.Fatalf("NewEncryptorWithKey failed: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"simple", "hello world"},
		{"api key", "sk-or-v1-abcdefghijklmnopqrstuvwxyz"},
		{"empty", ""},
		{"special chars", "!@#$%^&*()_+-=[]{}|;':\",./<>?"},
		{"unicode", "Hello 世界 🌍"},
		{"long", string(make([]byte, 10000))},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			encrypted, err := enc.Encrypt(tc.plaintext)
			if err != nil {
				t.Fatalf("Encrypt failed: %v", err)
			}

			// Encrypted should be different from plaintext (unless empty)
			if tc.plaintext != "" && encrypted == tc.plaintext {
				t.Error("encrypted text should differ from plaintext")
			}

			decrypted, err := enc.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt failed: %v", err)
			}

			if decrypted != tc.plaintext {
				t.Errorf("expected %q, got %q", tc.plaintext, decrypted)
			}
		})
	}
}

func TestEncryptionDifferentEachTime(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc, err := NewEncryptorWithKey(key)
	if err != nil {
		t.Fatalf("NewEncryptorWithKey failed: %v", err)
	}

	plaintext := "same plaintext"
	encrypted1, _ := enc.Encrypt(plaintext)
	encrypted2, _ := enc.Encrypt(plaintext)

	// Due to random nonce, encryptions should differ
	if encrypted1 == encrypted2 {
		t.Error("encryptions of same plaintext should produce different ciphertexts")
	}

	// But both should decrypt to the same plaintext
	decrypted1, _ := enc.Decrypt(encrypted1)
	decrypted2, _ := enc.Decrypt(encrypted2)

	if decrypted1 != plaintext || decrypted2 != plaintext {
		t.Error("both ciphertexts should decrypt to original plaintext")
	}
}

func TestDecryptInvalidData(t *testing.T) {
	key := make([]byte, 32)
	for i := range key {
		key[i] = byte(i)
	}

	enc, err := NewEncryptorWithKey(key)
	if err != nil {
		t.Fatalf("NewEncryptorWithKey failed: %v", err)
	}

	// Invalid base64
	_, err = enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Error("expected error for invalid base64")
	}

	// Valid base64 but invalid ciphertext
	_, err = enc.Decrypt("aGVsbG8gd29ybGQ=") // "hello world" in base64
	if err == nil {
		t.Error("expected error for invalid ciphertext")
	}
}

func TestNewEncryptorInvalidKeyLength(t *testing.T) {
	// Key too short
	_, err := NewEncryptorWithKey([]byte("short"))
	if err == nil {
		t.Error("expected error for key too short")
	}

	// Key too long
	longKey := make([]byte, 64)
	_, err = NewEncryptorWithKey(longKey)
	if err == nil {
		t.Error("expected error for key too long")
	}
}

func TestNewEncryptorDefault(t *testing.T) {
	// Default encryptor should work with machine-derived key
	enc, err := NewEncryptor()
	if err != nil {
		t.Fatalf("NewEncryptor failed: %v", err)
	}

	plaintext := "test api key"
	encrypted, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	decrypted, err := enc.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("expected %q, got %q", plaintext, decrypted)
	}
}
