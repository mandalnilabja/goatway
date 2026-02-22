package proxy

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/types"
)

// Transcription handles POST /v1/audio/transcriptions requests.
// Converts audio to text using Whisper models.
func (h *Handlers) Transcription(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Audio transcription uses multipart/form-data
	// Parse up to 32MB of file data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("failed to parse multipart form"))
		return
	}

	// Get model from form
	model := r.FormValue("model")
	if model == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("model is required"))
		return
	}

	// Verify file is present
	_, _, err := r.FormFile("file")
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("audio file is required"))
		return
	}

	// Build proxy options (credential resolved by Router)
	opts := &provider.ProxyOptions{
		RequestID:   requestID,
		Model:       model,
		IsStreaming: false,
		Body:        nil, // Multipart form is passed through r directly
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, opts, model, result, startTime)
}

// Translation handles POST /v1/audio/translations requests.
// Translates audio to English text using Whisper models.
func (h *Handlers) Translation(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()
	startTime := time.Now()

	// Audio translation uses multipart/form-data
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("failed to parse multipart form"))
		return
	}

	// Get model from form
	model := r.FormValue("model")
	if model == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("model is required"))
		return
	}

	// Verify file is present
	_, _, err := r.FormFile("file")
	if err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("audio file is required"))
		return
	}

	// Build proxy options (credential resolved by Router)
	opts := &provider.ProxyOptions{
		RequestID:   requestID,
		Model:       model,
		IsStreaming: false,
		Body:        nil, // Multipart form is passed through r directly
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, opts, model, result, startTime)
}
