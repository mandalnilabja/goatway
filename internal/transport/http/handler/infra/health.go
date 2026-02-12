package infra

import (
	"encoding/json"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/version"
)

// RootStatus returns JSON status and version information at /.
func (h *Handlers) RootStatus(w http.ResponseWriter, r *http.Request) {
	response := map[string]interface{}{
		"name":    "goatway",
		"version": version.Version,
		"status":  "running",
		"web_ui":  "/web",
		"api":     "/v1",
		"admin":   "/api/admin",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// HealthCheck handler returns the application health status.
func (h *Handlers) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "active",
		"app":    "goatway",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
