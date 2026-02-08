package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
)

// GetUsageStats handles GET /api/admin/usage
func (h *Repo) GetUsageStats(w http.ResponseWriter, r *http.Request) {
	filter := parseStatsFilter(r)

	stats, err := h.Storage.GetUsageStats(filter)
	if err != nil {
		writeJSONError(w, "Failed to get usage stats: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, stats, http.StatusOK)
}

// GetDailyUsage handles GET /api/admin/usage/daily
func (h *Repo) GetDailyUsage(w http.ResponseWriter, r *http.Request) {
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
		writeJSONError(w, "Failed to get daily usage: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"daily_usage": usage,
		"start_date":  startDate,
		"end_date":    endDate,
	}, http.StatusOK)
}

// GetRequestLogs handles GET /api/admin/logs
func (h *Repo) GetRequestLogs(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilter(r)

	logs, err := h.Storage.GetRequestLogs(filter)
	if err != nil {
		writeJSONError(w, "Failed to get request logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"logs":   logs,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}, http.StatusOK)
}

// DeleteRequestLogs handles DELETE /api/admin/logs
func (h *Repo) DeleteRequestLogs(w http.ResponseWriter, r *http.Request) {
	beforeDate := r.URL.Query().Get("before_date")
	if beforeDate == "" {
		writeJSONError(w, "before_date query parameter is required (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", beforeDate); err != nil {
		writeJSONError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	deleted, err := h.Storage.DeleteRequestLogs(beforeDate)
	if err != nil {
		writeJSONError(w, "Failed to delete logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, map[string]any{
		"deleted_count": deleted,
		"before_date":   beforeDate,
	}, http.StatusOK)
}

// parseLogFilter creates a LogFilter from query parameters
func parseLogFilter(r *http.Request) storage.LogFilter {
	filter := storage.LogFilter{
		Limit:  50, // default
		Offset: 0,
	}

	if v := r.URL.Query().Get("credential_id"); v != "" {
		filter.CredentialID = v
	}
	if v := r.URL.Query().Get("model"); v != "" {
		filter.Model = v
	}
	if v := r.URL.Query().Get("provider"); v != "" {
		filter.Provider = v
	}
	if v := r.URL.Query().Get("status_code"); v != "" {
		if code, err := strconv.Atoi(v); err == nil {
			filter.StatusCode = &code
		}
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		if limit, err := strconv.Atoi(v); err == nil && limit > 0 {
			filter.Limit = limit
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if offset, err := strconv.Atoi(v); err == nil && offset >= 0 {
			filter.Offset = offset
		}
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

// parseStatsFilter creates a StatsFilter from query parameters
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
