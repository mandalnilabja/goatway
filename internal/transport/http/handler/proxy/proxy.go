package proxy

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/google/uuid"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/tokenizer"
)

// Handlers holds the dependencies for proxy HTTP handlers.
type Handlers struct {
	Provider  provider.Provider
	Storage   storage.Storage
	Tokenizer tokenizer.Tokenizer
	Cache     *ristretto.Cache[string, any]
}

// New creates a new instance of proxy handlers.
func New(prov provider.Provider, store storage.Storage, tok tokenizer.Tokenizer, cache *ristretto.Cache[string, any]) *Handlers {
	return &Handlers{
		Provider:  prov,
		Storage:   store,
		Tokenizer: tok,
		Cache:     cache,
	}
}

// updateDailyUsage updates the daily usage aggregate for a request.
func (h *Handlers) updateDailyUsage(credentialID string, result *provider.ProxyResult, prompt, completion, total int) {
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

// logRequestBase creates a base request log entry.
func (h *Handlers) logRequestBase(requestID, credentialID, model string, result *provider.ProxyResult, startTime time.Time) *storage.RequestLog {
	duration := time.Since(startTime)

	return &storage.RequestLog{
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
}

// logSimpleRequest logs a simple request (no token counts) to storage.
func (h *Handlers) logSimpleRequest(requestID string, opts *provider.ProxyOptions, model string, result *provider.ProxyResult, startTime time.Time) {
	if h.Storage == nil || result == nil {
		return
	}

	credentialID := ""
	if opts.Credential != nil {
		credentialID = opts.Credential.ID
	}

	log := h.logRequestBase(requestID, credentialID, model, result, startTime)
	_ = h.Storage.LogRequest(log)

	// Update daily usage
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
