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
