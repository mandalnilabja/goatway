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

// LegacyCompletion handles POST /v1/completions requests.
// This is the legacy text completion endpoint (deprecated but needed for compatibility).
func (h *Repo) LegacyCompletion(w http.ResponseWriter, r *http.Request) {
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
	var req types.CompletionRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request format"))
		return
	}

	// Validate required fields
	if req.Model == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("model is required"))
		return
	}
	if len(req.Prompt.Values) == 0 {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("prompt is required"))
		return
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
		Model:       req.Model,
		IsStreaming: req.Stream,
		Body:        bytes.NewReader(bodyBytes),
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logCompletionRequest(requestID, credID, result, startTime)
}

// logCompletionRequest logs a completion request to storage.
func (h *Repo) logCompletionRequest(requestID, credentialID string, result *provider.ProxyResult, startTime time.Time) {
	if h.Storage == nil || result == nil {
		return
	}

	duration := time.Since(startTime)

	// Use token counts from result
	prompt := result.PromptTokens
	completion := result.CompletionTokens
	total := result.TotalTokens
	if total == 0 {
		total = prompt + completion
	}

	log := &storage.RequestLog{
		ID:               uuid.New().String(),
		RequestID:        requestID,
		CredentialID:     credentialID,
		Model:            result.Model,
		Provider:         h.Provider.Name(),
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      total,
		IsStreaming:      result.IsStreaming,
		StatusCode:       result.StatusCode,
		ErrorMessage:     result.ErrorMessage,
		DurationMs:       duration.Milliseconds(),
		CreatedAt:        time.Now(),
	}

	_ = h.Storage.LogRequest(log)

	// Update daily usage
	h.updateDailyUsage(credentialID, result, prompt, completion, total)
}
