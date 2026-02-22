package proxy

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

// Embeddings handles POST /v1/embeddings requests.
// Proxies to the upstream provider and logs usage.
func (h *Handlers) Embeddings(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Read and buffer the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("failed to read request body"))
		return
	}
	r.Body.Close()

	// Parse request to extract model
	var req types.EmbeddingsRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request format"))
		return
	}

	// Validate required fields
	if req.Model == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("model is required"))
		return
	}
	if len(req.Input.Values) == 0 {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("input is required"))
		return
	}

	// Build proxy options (credential resolved by Router)
	opts := &provider.ProxyOptions{
		RequestID:   requestID,
		Model:       req.Model,
		IsStreaming: false, // Embeddings don't support streaming
		Body:        bytes.NewReader(bodyBytes),
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log the request asynchronously
	go h.logEmbeddingsRequest(requestID, opts, req.Model, result, startTime)
}

// logEmbeddingsRequest logs an embeddings request to storage.
func (h *Handlers) logEmbeddingsRequest(requestID string, opts *provider.ProxyOptions, model string, result *provider.ProxyResult, startTime time.Time) {
	if h.Storage == nil || result == nil {
		return
	}

	credentialID := ""
	if opts.Credential != nil {
		credentialID = opts.Credential.ID
	}

	duration := time.Since(startTime)

	log := &storage.RequestLog{
		ID:           uuid.New().String(),
		RequestID:    requestID,
		CredentialID: credentialID,
		Model:        model,
		Provider:     h.Provider.Name(),
		PromptTokens: result.PromptTokens,
		TotalTokens:  result.TotalTokens,
		IsStreaming:  false,
		StatusCode:   result.StatusCode,
		ErrorMessage: result.ErrorMessage,
		DurationMs:   duration.Milliseconds(),
		CreatedAt:    time.Now(),
	}

	_ = h.Storage.LogRequest(log)

	// Update daily usage
	h.updateDailyUsage(credentialID, result, result.PromptTokens, 0, result.TotalTokens)
}
