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

// ImageGeneration handles POST /v1/images/generations requests.
// Generates images from text prompts using DALL-E or similar models.
func (h *Handlers) ImageGeneration(w http.ResponseWriter, r *http.Request) {
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
	var req types.ImageGenerationRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request format"))
		return
	}

	// Validate required fields
	if req.Prompt == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("prompt is required"))
		return
	}

	// Default model if not specified
	model := req.Model
	if model == "" {
		model = "dall-e-2"
	}

	// Build proxy options (credential resolved by Router)
	opts := &provider.ProxyOptions{
		RequestID:   requestID,
		Model:       model,
		IsStreaming: false,
		Body:        bytes.NewReader(bodyBytes),
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, opts, model, result, startTime)
}
