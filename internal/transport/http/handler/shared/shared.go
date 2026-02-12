package shared

import (
	"encoding/json"
	"net/http"
)

// WriteJSON writes a JSON response with the given status code.
func WriteJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// WriteJSONError writes a JSON error response.
func WriteJSONError(w http.ResponseWriter, message string, status int) {
	WriteJSON(w, map[string]any{
		"error": map[string]any{
			"message": message,
			"code":    status,
		},
	}, status)
}

// IsValidAdminPassword validates the admin password format.
// Password must be alphanumeric (a-z, A-Z, 0-9) with minimum 8 characters.
func IsValidAdminPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	for _, c := range password {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}
