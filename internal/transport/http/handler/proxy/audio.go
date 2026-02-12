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

// TextToSpeech handles POST /v1/audio/speech requests.
// Converts text to spoken audio using TTS models.
func (h *Handlers) TextToSpeech(w http.ResponseWriter, r *http.Request) {
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
	var req types.AudioSpeechRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("invalid request format"))
		return
	}

	// Validate required fields
	if req.Model == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("model is required"))
		return
	}
	if req.Input == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("input is required"))
		return
	}
	if req.Voice == "" {
		types.WriteError(w, http.StatusBadRequest, types.ErrInvalidRequest("voice is required"))
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
		IsStreaming: false,
		Body:        bytes.NewReader(bodyBytes),
	}

	// Proxy the request
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Log asynchronously
	go h.logSimpleRequest(requestID, credID, req.Model, result, startTime)
}

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

	// Resolve API key
	apiKey, credID := h.resolveAPIKey(r)
	if apiKey == "" {
		h.writeError(w, "No API key provided.", http.StatusUnauthorized)
		return
	}

	// Build proxy options - body is nil for multipart, provider handles it
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
