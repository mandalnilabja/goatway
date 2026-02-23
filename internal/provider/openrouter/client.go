// Package openrouter implements the OpenRouter LLM provider.
package openrouter

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mandalnilabja/goatway/internal/types"
)

// Provider implements the provider.Provider interface for OpenRouter.
// API key is resolved per-request from storage, not stored on the provider.
type Provider struct{}

// New creates a new OpenRouter provider instance.
// API key is resolved per-request from storage via ProxyOptions.
func New() *Provider {
	return &Provider{}
}

// Name returns the provider identifier
func (p *Provider) Name() string {
	return "openrouter"
}

// BaseURL returns the OpenRouter API endpoint
func (p *Provider) BaseURL() string {
	return "https://openrouter.ai/api/v1/chat/completions"
}

// PrepareRequest adds OpenRouter-specific headers to the request
func (p *Provider) PrepareRequest(ctx context.Context, req *http.Request) error {
	req.Header.Set("HTTP-Referer", "https://github.com/mandalnilabja/goatway")
	req.Header.Set("X-Title", "Goatway Proxy")
	return nil
}

// ProxyRequest handles the proxy to OpenRouter with result tracking.
// CRITICAL: Maintains streaming semantics with no buffering.
func (p *Provider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *types.ProxyOptions) (*types.ProxyResult, error) {
	startTime := time.Now()
	result := &types.ProxyResult{
		Model:        opts.Model,
		PromptTokens: opts.PromptTokens,
		IsStreaming:  opts.IsStreaming,
	}

	// API key must be provided via credential (resolved by Router)
	if opts.Credential == nil {
		result.Error = types.ErrNoAPIKey
		result.StatusCode = http.StatusUnauthorized
		http.Error(w, "No credential configured", http.StatusUnauthorized)
		return result, types.ErrNoAPIKey
	}
	apiKey := opts.Credential.GetAPIKey()

	// Read and rewrite body with resolved model name
	body, err := rewriteModelInBody(opts.Body, req.Body, opts.Model)
	if err != nil {
		result.Error = err
		result.StatusCode = http.StatusBadRequest
		http.Error(w, "Failed to process request body", http.StatusBadRequest)
		return result, err
	}

	// Create upstream request
	upstreamReq, err := http.NewRequestWithContext(ctx, req.Method, p.BaseURL(), body)
	if err != nil {
		result.Error = err
		result.StatusCode = http.StatusInternalServerError
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return result, err
	}

	// Copy headers (skip hop-by-hop)
	for k, v := range req.Header {
		if k == "Content-Length" || k == "Connection" || k == "Host" || k == "Authorization" {
			continue
		}
		upstreamReq.Header[k] = v
	}

	// Set authorization with the resolved API key
	upstreamReq.Header.Set("Authorization", "Bearer "+apiKey)

	// Add provider-specific headers
	if err := p.PrepareRequest(ctx, upstreamReq); err != nil {
		result.Error = err
		result.StatusCode = http.StatusInternalServerError
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return result, err
	}

	// Setup client (DisableCompression required for streaming)
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	// Execute request
	resp, err := client.Do(upstreamReq)
	if err != nil {
		result.Error = err
		result.StatusCode = http.StatusBadGateway
		http.Error(w, "Bad Gateway: "+err.Error(), http.StatusBadGateway)
		return result, err
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	result.Duration = time.Since(startTime)

	// Handle error responses
	if resp.StatusCode >= 400 {
		return handleErrorResponse(w, resp, result)
	}

	// Route based on content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") {
		return handleStreamingResponse(w, resp, result)
	}
	return handleJSONResponse(w, resp, result)
}
