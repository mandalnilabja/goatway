package storage

import "testing"

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
