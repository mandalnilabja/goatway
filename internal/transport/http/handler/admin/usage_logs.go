package admin

import (
	"net/http"
	"strconv"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/shared"
)

// GetRequestLogs handles GET /api/admin/logs.
func (h *Handlers) GetRequestLogs(w http.ResponseWriter, r *http.Request) {
	filter := parseLogFilter(r)

	logs, err := h.Storage.GetRequestLogs(filter)
	if err != nil {
		shared.WriteJSONError(w, "Failed to get request logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, map[string]any{
		"logs":   logs,
		"limit":  filter.Limit,
		"offset": filter.Offset,
	}, http.StatusOK)
}

// DeleteRequestLogs handles DELETE /api/admin/logs.
func (h *Handlers) DeleteRequestLogs(w http.ResponseWriter, r *http.Request) {
	beforeDate := r.URL.Query().Get("before_date")
	if beforeDate == "" {
		shared.WriteJSONError(w, "before_date query parameter is required (format: YYYY-MM-DD)", http.StatusBadRequest)
		return
	}

	// Validate date format
	if _, err := time.Parse("2006-01-02", beforeDate); err != nil {
		shared.WriteJSONError(w, "Invalid date format. Use YYYY-MM-DD", http.StatusBadRequest)
		return
	}

	deleted, err := h.Storage.DeleteRequestLogs(beforeDate)
	if err != nil {
		shared.WriteJSONError(w, "Failed to delete logs: "+err.Error(), http.StatusInternalServerError)
		return
	}

	shared.WriteJSON(w, map[string]any{
		"deleted_count": deleted,
		"before_date":   beforeDate,
	}, http.StatusOK)
}

// parseLogFilter creates a LogFilter from query parameters.
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
