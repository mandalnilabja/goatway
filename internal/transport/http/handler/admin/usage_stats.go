package admin

import (
	"net/http"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/shared"
)

// GetUsageStats handles GET /api/admin/usage.
func (h *Handlers) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	filter := parseStatsFilter(r)

	stats, err := h.Storage.GetUsageStats(filter)
	if err != nil {
		shared.WriteJSONError(w, "Failed to get usage stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, stats, http.StatusOK)
}

// GetDailyUsage handles GET /api/admin/usage/daily.
func (h *Handlers) GetDailyUsage(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	endDate := r.URL.Query().Get("end_date")

	// Default to last 30 days if not specified
	if startDate == "" {
		startDate = time.Now().AddDate(0, 0, -30).Format("2006-01-02")
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}

	usage, err := h.Storage.GetDailyUsage(startDate, endDate)
	if err != nil {
		shared.WriteJSONError(w, "Failed to get daily usage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, map[string]any{
		"daily_usage": usage,
		"start_date":  startDate,
		"end_date":    endDate,
	}, http.StatusOK)
}

// parseStatsFilter creates a StatsFilter from query parameters.
func parseStatsFilter(r *http.Request) storage.StatsFilter {
	filter := storage.StatsFilter{}

	if v := r.URL.Query().Get("credential_id"); v != "" {
		filter.CredentialID = v
	}
	if v := r.URL.Query().Get("model"); v != "" {
		filter.Model = v
	}
	if v := r.URL.Query().Get("provider"); v != "" {
		filter.Provider = v
	}
	if v := r.URL.Query().Get("start_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.StartDate = &t
		}
	}
	if v := r.URL.Query().Get("end_date"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			filter.EndDate = &t
		}
	}

	return filter
}
