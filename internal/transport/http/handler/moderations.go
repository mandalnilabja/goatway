package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

// Moderation handles POST /v1/moderations requests.
// Classifies text for potential policy violations.
func (h *Repo) Moderation(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Read and buffer the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("failed to read request body"))
		return
	}
	r.Body.Close()

	// Parse request
	var req types.ModerationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request format"))
		return
	}

	// Validate required fields
	if len(req.Input.Values) == 0 {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("input is required"))
		return
	}

	// Default model if not specified
	model := req.Model
	if model == "" {
		model = "omni-moderation-latest"
	}

	// Resolve API key
	apiKey, credID := h.resolveAPIKey(r)
	if apiKey == "" {
		h.writeError(w, "No API key provided.", http.StatusUnauthorized)
		return
	}

	// Build proxy options
	opts := &provider.ProxyOptions{
		APIKey:      apiKey,
		RequestID:   requestID,
		Model:       model,
		IsStreaming: false, // Moderations don't support streaming
		Body:        bytes.NewReader(bodyBytes),
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logModerationRequest(requestID, credID, model, result, startTime)
}

// logModerationRequest logs a moderation request to storage.
func (h *Repo) logModerationRequest(requestID, credentialID, model string, result *provider.ProxyResult, startTime time.Time) {
	if h.Storage == nil || result == nil {
		return
	}

	duration := time.Since(startTime)

	log := &storage.RequestLog{
		ID:           uuid.New().String(),
		RequestID:    requestID,
		CredentialID: credentialID,
		Model:        model,
		Provider:     h.Provider.Name(),
		IsStreaming:  false,
		StatusCode:   result.StatusCode,
		ErrorMessage: result.ErrorMessage,
		DurationMs:   duration.Milliseconds(),
		CreatedAt:    time.Now(),
	}

	_ = h.Storage.LogRequest(log)

	// Update daily usage (moderations don't have token counts)
	errorCount := 0
	if result.StatusCode >= 400 {
		errorCount = 1
	}

	usage := &storage.DailyUsage{
		Date:         time.Now().Format("2006-01-02"),
		CredentialID: credentialID,
		Model:        model,
		RequestCount: 1,
		ErrorCount:   errorCount,
	}

	_ = h.Storage.UpdateDailyUsage(usage)
}
