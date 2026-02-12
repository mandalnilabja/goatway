package storage

import (
	"strings"
	"testing"
)

func TestGenerateAPIKey(t *testing.T) {
	key, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	// Check prefix
	if !strings.HasPrefix(key, APIKeyPrefix) {
		t.Errorf("key should start with %q, got: %s", APIKeyPrefix, key)
	}

	// Check total length: "gw_" (3) + 64 chars = 67
	expectedLen := len(APIKeyPrefix) + APIKeyLength
	if len(key) != expectedLen {
		t.Errorf("expected key length %d, got %d", expectedLen, len(key))
	}

	// Check all chars after prefix are base62
	suffix := key[len(APIKeyPrefix):]
	for i, c := range suffix {
		if !isBase62(byte(c)) {
			t.Errorf("invalid character at position %d: %c", i, c)
		}
	}
}

func TestGenerateAPIKeyUniqueness(t *testing.T) {
	seen := make(map[string]bool)
	count := 1000

	for i := 0; i < count; i++ {
		key, err := GenerateAPIKey()
		if err != nil {
			t.Fatalf("GenerateAPIKey failed on iteration %d: %v", i, err)
		}

		if seen[key] {
			t.Errorf("duplicate key generated: %s", key)
		}
		seen[key] = true
	}
}

func TestExtractKeyPrefix(t *testing.T) {
	testCases := []struct {
		name     string
		key      string
		expected string
	}{
		{
			name:     "full key",
			key:      "gw_a1B2c3D4e5F6g7H8i9J0k1L2m3N4o5P6q7R8s9T0u1V2w3X4y5Z6a7B8c9D0e1F2",
			expected: "gw_a1B2c3D4",
		},
		{
			name:     "exact prefix length",
			key:      "gw_a1B2c3D4",
			expected: "gw_a1B2c3D4",
		},
		{
			name:     "shorter than prefix",
			key:      "gw_abc",
			expected: "gw_abc",
		},
		{
			name:     "empty",
			key:      "",
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := ExtractKeyPrefix(tc.key)
			if result != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestExtractKeyPrefixFromGenerated(t *testing.T) {
	key, err := GenerateAPIKey()
	if err != nil {
		t.Fatalf("GenerateAPIKey failed: %v", err)
	}

	prefix := ExtractKeyPrefix(key)

	// Prefix should be 11 chars: "gw_" + 8 random chars
	if len(prefix) != APIKeyPrefixLen {
		t.Errorf("expected prefix length %d, got %d", APIKeyPrefixLen, len(prefix))
	}

	// Prefix should match the start of the key
	if !strings.HasPrefix(key, prefix) {
		t.Errorf("key %q should start with prefix %q", key, prefix)
	}
}

func isBase62(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z')
}

func BenchmarkGenerateAPIKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = GenerateAPIKey()
	}
}
