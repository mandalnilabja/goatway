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
		IsStreaming: false,
		Body:        bytes.NewReader(bodyBytes),
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, credID, model, result, startTime)
}

// ImageEdit handles POST /v1/images/edits requests.
// Edits images based on a prompt and optional mask.
func (h *Handlers) ImageEdit(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Image edit uses multipart/form-data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("failed to parse multipart form"))
		return
	}

	// Verify image file is present
	_, _, err := r.FormFile("image")
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("image file is required"))
		return
	}

	// Verify prompt is present
	prompt := r.FormValue("prompt")
	if prompt == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("prompt is required"))
		return
	}

	// Get model (optional)
	model := r.FormValue("model")
	if model == "" {
		model = "dall-e-2"
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
		IsStreaming: false,
		Body:        nil, // Multipart form is passed through r directly
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, credID, model, result, startTime)
}

// ImageVariation handles POST /v1/images/variations requests.
// Creates variations of an existing image.
func (h *Handlers) ImageVariation(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Image variation uses multipart/form-data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("failed to parse multipart form"))
		return
	}

	// Verify image file is present
	_, _, err := r.FormFile("image")
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("image file is required"))
		return
	}

	// Get model (optional)
	model := r.FormValue("model")
	if model == "" {
		model = "dall-e-2"
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
		IsStreaming: false,
		Body:        nil, // Multipart form is passed through r directly
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, credID, model, result, startTime)
}
