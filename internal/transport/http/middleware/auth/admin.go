// Package auth provides authentication middleware for HTTP routes.
package auth

import (
	"encoding/json"
	"net/http"
)

// AdminAuth middleware protects admin routes using session-based authentication only.
// No Bearer token fallback - admin access requires login via web UI.
func AdminAuth(sessions *SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Session management required
			if sessions == nil {
				writeUnauthorized(w, "session management not configured")
				return
			}

			// Check for valid session cookie
			cookie, err := r.Cookie("goatway_session")
			if err != nil || cookie.Value == "" {
				writeUnauthorized(w, "session required")
				return
			}

			session := sessions.Get(cookie.Value)
			if session == nil {
				writeUnauthorized(w, "invalid or expired session")
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
