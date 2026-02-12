package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

// tokenCountTimeout is the maximum time to wait for token counting before proceeding.
const tokenCountTimeout = 100 * time.Millisecond

// OpenAIProxy forwards requests to the configured LLM provider with logging.
// Token counting runs in parallel with the proxy request to minimize latency.
func (h *Repo) OpenAIProxy(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	// Read and buffer the request body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}
	r.Body.Close()

	// Parse request to extract model and messages
	var req types.ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	// Resolve API key from Authorization header or default credential
	apiKey, credID := h.resolveAPIKey(r)
	if apiKey == "" {
		h.writeError(w, "No API key provided. Set Authorization header or configure default credential.", http.StatusUnauthorized)
		return
	}

	// Start token counting in background goroutine (non-blocking)
	// This allows the proxy request to start immediately without waiting for token counting
	tokensChan := make(chan int, 1)
	go func() {
		defer close(tokensChan)
		if h.Tokenizer != nil {
			if tokens, err := h.Tokenizer.CountRequest(&req); err == nil {
				tokensChan <- tokens
			}
		}
	}()

	// Build proxy options (prompt tokens will be collected after proxy completes)
	opts := &provider.ProxyOptions{
		APIKey:       apiKey,
		RequestID:    requestID,
		PromptTokens: 0, // Will be populated from upstream response or background count
		Model:        req.Model,
		IsStreaming:  req.Stream,
		Body:         bytes.NewReader(bodyBytes),
	}

	// Proxy the request immediately - don't wait for token counting
	result, _ := h.Provider.ProxyRequest(r.Context(), w, r, opts)

	// Collect token count with timeout (100ms max wait)
	// Token counting may already be done, or we give it a short grace period
	var promptTokens int
	select {
	case tokens, ok := <-tokensChan:
		if ok {
			promptTokens = tokens
		}
	case <-time.After(tokenCountTimeout):
		// Token counting took too long, proceed with 0 (upstream may provide it)
	}

	// Log the request asynchronously
	go h.logRequest(requestID, credID, result, promptTokens)
}

// resolveAPIKey extracts API key from Authorization header or default credential.
func (h *Repo) resolveAPIKey(r *http.Request) (apiKey string, credentialID string) {
	// Check Authorization header
	auth := r.Header.Get("Authorization")
	if auth != "" {
		if strings.HasPrefix(auth, "Bearer ") {
			return strings.TrimPrefix(auth, "Bearer "), ""
		}
	}

	// Fall back to default credential from storage
	if h.Storage != nil {
		cred, err := h.Storage.GetDefaultCredential(h.Provider.Name())
		if err == nil && cred != nil {
			return cred.APIKey, cred.ID
		}
	}

	return "", ""
}

// logRequest logs the proxy request to storage asynchronously.
func (h *Repo) logRequest(requestID, credentialID string, result *provider.ProxyResult, promptTokens int) {
	if h.Storage == nil || result == nil {
		return
	}

	// Use upstream token counts if available, otherwise use pre-calculated
	prompt := result.PromptTokens
	if prompt == 0 {
		prompt = promptTokens
	}
	completion := result.CompletionTokens
	total := result.TotalTokens
	if total == 0 {
		total = prompt + completion
	}

	// Create request log entry
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
		DurationMs:       result.Duration.Milliseconds(),
		CreatedAt:        time.Now(),
	}

	// Log to storage (ignore errors in async context)
	_ = h.Storage.LogRequest(log)

	// Update daily usage aggregates
	h.updateDailyUsage(credentialID, result, prompt, completion, total)
}

// updateDailyUsage updates the daily usage aggregate for this request.
func (h *Repo) updateDailyUsage(credentialID string, result *provider.ProxyResult, prompt, completion, total int) {
	today := time.Now().Format("2006-01-02")

	errorCount := 0
	if result.StatusCode >= 400 {
		errorCount = 1
	}

	usage := &storage.DailyUsage{
		Date:             today,
		CredentialID:     credentialID,
		Model:            result.Model,
		RequestCount:     1,
		PromptTokens:     prompt,
		CompletionTokens: completion,
		TotalTokens:      total,
		ErrorCount:       errorCount,
	}

	_ = h.Storage.UpdateDailyUsage(usage)
}

// writeError writes an OpenAI-compatible error response.
func (h *Repo) writeError(w http.ResponseWriter, message string, status int) {
	types.WriteError(w, status, types.ErrAuthentication(message))
}
