package storage

import "testing"

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
