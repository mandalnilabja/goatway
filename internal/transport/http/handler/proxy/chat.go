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

// tokenCountTimeout is the maximum time to wait for token counting before proceeding.
const tokenCountTimeout = 100 * time.Millisecond

// ChatCompletions forwards requests to the configured LLM provider with logging.
// Token counting runs in parallel with the proxy request to minimize latency.
func (h *Handlers) ChatCompletions(w http.ResponseWriter, r *http.Request) {
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

	// Build proxy options (credential resolved by Router)
	opts := &provider.ProxyOptions{
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

	// Log the request asynchronously (credential ID from opts set by Router)
	go h.logChatRequest(requestID, opts, result, promptTokens)
}

// logChatRequest logs the proxy request to storage asynchronously.
func (h *Handlers) logChatRequest(requestID string, opts *provider.ProxyOptions, result *provider.ProxyResult, promptTokens int) {
	if h.Storage == nil || result == nil {
		return
	}

	// Get credential ID from opts (set by Router)
	credentialID := ""
	if opts.Credential != nil {
		credentialID = opts.Credential.ID
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
