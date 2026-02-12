package handler

import (
	"encoding/json"
	"net/http"
)

// RootStatus returns JSON status and version information at /
func (h *Repo) RootStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":    "goatway",
		"version": Version,
		"status":  "running",
		"web_ui":  "/web",
		"api":     "/v1",
		"admin":   "/api/admin",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
