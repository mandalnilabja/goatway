package handler

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/storage"
)

// Version is set at build time
var Version = "dev"

// AdminHealth handles GET /api/admin/health
func (h *Repo) AdminHealth(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	dbStatus := "connected"

	// Check database connectivity by listing credentials
	if _, err := h.Storage.ListCredentials(); err != nil {
		status = "degraded"
		dbStatus = "error: " + err.Error()
	}

	writeJSON(w, map[string]any{
		"status":    status,
		"database":  dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, http.StatusOK)
}

// AdminInfo handles GET /api/admin/info
func (h *Repo) AdminInfo(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.StartTime)

	// Get quick stats
	stats, _ := h.Storage.GetUsageStats(storage.StatsFilter{})
	creds, _ := h.Storage.ListCredentials()

	writeJSON(w, map[string]any{
		"version":     Version,
		"go_version":  runtime.Version(),
		"uptime":      uptime.String(),
		"uptime_secs": int64(uptime.Seconds()),
		"data_dir":    config.DataDir(),
		"stats": map[string]any{
			"total_credentials": len(creds),
			"total_requests":    stats.TotalRequests,
			"total_tokens":      stats.TotalTokens,
		},
	}, http.StatusOK)
}

// ChangePasswordRequest is the request body for changing admin password.
type ChangePasswordRequest struct {
	NewPassword string `json:"new_password"`
}

// ChangeAdminPassword changes the admin password (PUT /api/admin/password).
func (h *Repo) ChangeAdminPassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !IsValidAdminPassword(req.NewPassword) {
		writeJSONError(w, "password must be alphanumeric, min 8 characters", http.StatusBadRequest)
		return
	}

	hash, err := storage.HashPassword(req.NewPassword, storage.DefaultArgon2Params())
	if err != nil {
		writeJSONError(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.Storage.SetAdminPasswordHash(hash); err != nil {
		writeJSONError(w, "failed to save password", http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]string{"message": "password updated"}, http.StatusOK)
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

// writeJSON writes a JSON response with the given status code
func writeJSON(w http.ResponseWriter, data any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

// writeJSONError writes a JSON error response
func writeJSONError(w http.ResponseWriter, message string, status int) {
	writeJSON(w, map[string]any{
		"error": map[string]any{
			"message": message,
			"code":    status,
		},
	}, status)
}
