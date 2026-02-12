package admin

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/shared"
	"github.com/mandalnilabja/goatway/internal/version"
)

// AdminHealth handles GET /api/admin/health.
func (h *Handlers) AdminHealth(w http.ResponseWriter, r *http.Request) {
	status := "healthy"
	dbStatus := "connected"

	// Check database connectivity by listing credentials
	if _, err := h.Storage.ListCredentials(); err != nil {
		status = "degraded"
		dbStatus = "error: " + err.Error()
	}

	shared.WriteJSON(w, map[string]any{
		"status":    status,
		"database":  dbStatus,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}, http.StatusOK)
}

// AdminInfo handles GET /api/admin/info.
func (h *Handlers) AdminInfo(w http.ResponseWriter, r *http.Request) {
	uptime := time.Since(h.StartTime)

	// Get quick stats
	stats, _ := h.Storage.GetUsageStats(storage.StatsFilter{})
	creds, _ := h.Storage.ListCredentials()

	shared.WriteJSON(w, map[string]any{
		"version":     version.Version,
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
func (h *Handlers) ChangeAdminPassword(w http.ResponseWriter, r *http.Request) {
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		shared.WriteJSONError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if !shared.IsValidAdminPassword(req.NewPassword) {
		shared.WriteJSONError(w, "password must be alphanumeric, min 8 characters", http.StatusBadRequest)
		return
	}

	hash, err := storage.HashPassword(req.NewPassword, storage.DefaultArgon2Params())
	if err != nil {
		shared.WriteJSONError(w, "failed to hash password", http.StatusInternalServerError)
		return
	}

	if err := h.Storage.SetAdminPasswordHash(hash); err != nil {
		shared.WriteJSONError(w, "failed to save password", http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, map[string]string{"message": "password updated"}, http.StatusOK)
}
