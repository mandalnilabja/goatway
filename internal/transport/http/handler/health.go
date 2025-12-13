package handler

import (
	"encoding/json"
	"net/http"
)

// Home handler returns a welcome message
func (h *Repo) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to Goatway API!"))
}

// HealthCheck handler returns the application health status
func (h *Repo) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "active",
		"app":    "goatway",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
