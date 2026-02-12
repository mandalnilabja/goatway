package proxy

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/types"
)

// Moderation handles POST /v1/moderations requests.
// Classifies text for potential policy violations.
func (h *Handlers) Moderation(w http.ResponseWriter, r *http.Request) {
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
	go h.logSimpleRequest(requestID, credID, model, result, startTime)
}
