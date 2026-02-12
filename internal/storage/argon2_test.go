package storage

import (
	"strings"
	"testing"
)

func TestDefaultArgon2Params(t *testing.T) {
	params := DefaultArgon2Params()

	if params.Memory != 64*1024 {
		t.Errorf("expected memory 64MB, got %d KB", params.Memory)
	}
	if params.Iterations != 1 {
		t.Errorf("expected iterations 1, got %d", params.Iterations)
	}
	if params.Parallelism != 4 {
		t.Errorf("expected parallelism 4, got %d", params.Parallelism)
	}
	if params.SaltLength != 16 {
		t.Errorf("expected salt length 16, got %d", params.SaltLength)
	}
	if params.KeyLength != 32 {
		t.Errorf("expected key length 32, got %d", params.KeyLength)
	}
}

func TestHashPassword(t *testing.T) {
	password := "testpassword123"
	params := DefaultArgon2Params()

	hash, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Verify hash format: $argon2id$v=19$m=65536,t=1,p=4$<salt>$<hash>
	if !strings.HasPrefix(hash, "$argon2id$v=") {
		t.Errorf("hash should start with $argon2id$v=, got: %s", hash)
	}

	parts := strings.Split(hash, "$")
	if len(parts) != 6 {
		t.Errorf("expected 6 parts in hash, got %d", len(parts))
	}
}

func TestHashPasswordNilParams(t *testing.T) {
	password := "testpassword123"

	hash, err := HashPassword(password, nil)
	if err != nil {
		t.Fatalf("HashPassword with nil params failed: %v", err)
	}

	if hash == "" {
		t.Error("expected non-empty hash")
	}
}

func TestHashPasswordUniqueness(t *testing.T) {
	password := "samepassword"
	params := DefaultArgon2Params()

	hash1, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("first hash failed: %v", err)
	}

	hash2, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("second hash failed: %v", err)
	}

	// Same password should produce different hashes (different salts)
	if hash1 == hash2 {
		t.Error("hashing same password twice should produce different hashes")
	}
}

func TestVerifyPasswordCorrect(t *testing.T) {
	password := "correctpassword"
	params := DefaultArgon2Params()

	hash, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}

	if !valid {
		t.Error("correct password should verify as valid")
	}
}

func TestVerifyPasswordIncorrect(t *testing.T) {
	password := "correctpassword"
	wrongPassword := "wrongpassword"
	params := DefaultArgon2Params()

	hash, err := HashPassword(password, params)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	valid, err := VerifyPassword(wrongPassword, hash)
	if err != nil {
		t.Fatalf("VerifyPassword failed: %v", err)
	}

	if valid {
		t.Error("incorrect password should not verify as valid")
	}
}

func TestVerifyPasswordInvalidHash(t *testing.T) {
	testCases := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"wrong format", "notahash"},
		{"wrong algorithm", "$argon2i$v=19$m=65536,t=1,p=4$c2FsdA$aGFzaA"},
		{"missing parts", "$argon2id$v=19$m=65536"},
		{"invalid base64 salt", "$argon2id$v=19$m=65536,t=1,p=4$!!!$aGFzaA"},
		{"invalid base64 hash", "$argon2id$v=19$m=65536,t=1,p=4$c2FsdA$!!!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := VerifyPassword("password", tc.hash)
			if err == nil {
				t.Error("expected error for invalid hash")
			}
		})
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	lengths := []uint32{16, 32, 64}

	for _, length := range lengths {
		bytes, err := GenerateRandomBytes(length)
		if err != nil {
			t.Fatalf("GenerateRandomBytes(%d) failed: %v", length, err)
		}

		if uint32(len(bytes)) != length {
			t.Errorf("expected %d bytes, got %d", length, len(bytes))
		}
	}
}

func TestGenerateRandomBytesUniqueness(t *testing.T) {
	seen := make(map[string]bool)

	for i := 0; i < 100; i++ {
		bytes, err := GenerateRandomBytes(16)
		if err != nil {
			t.Fatalf("GenerateRandomBytes failed: %v", err)
		}

		key := string(bytes)
		if seen[key] {
			t.Error("generated duplicate random bytes")
		}
		seen[key] = true
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "benchmarkpassword"
	params := DefaultArgon2Params()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = HashPassword(password, params)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "benchmarkpassword"
	params := DefaultArgon2Params()
	hash, _ := HashPassword(password, params)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyPassword(password, hash)
	}
}
