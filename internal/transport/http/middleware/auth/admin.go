// Package auth provides authentication middleware for HTTP routes.
package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mandalnilabja/goatway/internal/storage"
)

// AdminAuth middleware protects admin routes using stored password hash.
// Requires Bearer token authentication with the admin password.
func AdminAuth(store storage.Storage) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract Bearer token
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				writeUnauthorized(w, "authorization required")
				return
			}
			password := strings.TrimPrefix(auth, "Bearer ")

			// Get stored hash and verify
			hash, err := store.GetAdminPasswordHash()
			if err != nil {
				writeUnauthorized(w, "server error")
				return
			}
			if hash == "" {
				writeUnauthorized(w, "admin not configured")
				return
			}

			valid, err := storage.VerifyPassword(password, hash)
			if err != nil || !valid {
				writeUnauthorized(w, "invalid credentials")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeUnauthorized writes a JSON 401 response.
func writeUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"message": message,
			"type":    "authentication_error",
		},
	})
}
