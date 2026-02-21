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
