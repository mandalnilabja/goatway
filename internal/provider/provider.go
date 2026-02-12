package provider

import (
	"context"
	"errors"
	"io"
	"net/http"
	"time"
)

// ErrNoAPIKey is returned when no API key is configured for a request
var ErrNoAPIKey = errors.New("no API key configured")

// Provider defines the interface all LLM providers must implement
type Provider interface {
	// Name returns the provider identifier
	Name() string

	// BaseURL returns the provider's API endpoint
	BaseURL() string

	// PrepareRequest adds provider-specific headers and modifications
	PrepareRequest(ctx context.Context, req *http.Request) error

	// ProxyRequest handles the streaming proxy to the provider
	// MUST maintain streaming semantics (no buffering)
	// Returns ProxyResult with request metadata for logging
	ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *ProxyOptions) (*ProxyResult, error)
}

// ProxyOptions contains options for proxying a request
type ProxyOptions struct {
	// APIKey to use for this request (from credential or header)
	APIKey string

	// RequestID for tracing
	RequestID string

	// PromptTokens pre-calculated by the handler
	PromptTokens int

	// Model from the parsed request
	Model string

	// IsStreaming indicates if this is a streaming request
	IsStreaming bool

	// Body is the request body (already read, needs to be replayed)
	Body io.Reader
}

// ProxyResult contains the result of a proxied request
type ProxyResult struct {
	// Model used for the request
	Model string

	// Token counts (from upstream or calculated)
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int

	// Request metadata
	StatusCode   int
	FinishReason string
	Duration     time.Duration
	IsStreaming  bool

	// Error info (if any)
	Error        error
	ErrorMessage string
}
