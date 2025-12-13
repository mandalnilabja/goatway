package provider

import (
	"context"
	"net/http"
)

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
	ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error
}
