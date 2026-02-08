package provider

import (
	"context"
	"io"
	"net/http"
	"strings"
	"time"
)

// OpenRouterProvider implements the Provider interface for OpenRouter
type OpenRouterProvider struct {
	APIKey string
}

// NewOpenRouterProvider creates a new OpenRouter provider instance
func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
	return &OpenRouterProvider{
		APIKey: apiKey,
	}
}

// Name returns the provider identifier
func (p *OpenRouterProvider) Name() string {
	return "openrouter"
}

// BaseURL returns the OpenRouter API endpoint
func (p *OpenRouterProvider) BaseURL() string {
	return "https://openrouter.ai/api/v1/chat/completions"
}

// PrepareRequest adds OpenRouter-specific headers to the request
func (p *OpenRouterProvider) PrepareRequest(ctx context.Context, req *http.Request) error {
	req.Header.Set("HTTP-Referer", "https://github.com/mandalnilabja/goatway")
	req.Header.Set("X-Title", "Goatway Proxy")
	return nil
}

// ProxyRequest handles the proxy to OpenRouter with result tracking.
// CRITICAL: Maintains streaming semantics with no buffering.
func (p *OpenRouterProvider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *ProxyOptions) (*ProxyResult, error) {
	startTime := time.Now()
	result := &ProxyResult{
		Model:        opts.Model,
		PromptTokens: opts.PromptTokens,
		IsStreaming:  opts.IsStreaming,
	}

	// Determine API key: from options or provider default
	apiKey := opts.APIKey
	if apiKey == "" {
		apiKey = p.APIKey
	}

	// Determine request body source
	var body io.Reader = req.Body
	if opts.Body != nil {
		body = opts.Body
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
		return p.handleErrorResponse(w, resp, result)
	}

	// Route based on content type
	contentType := resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "text/event-stream") {
		return p.handleStreamingResponse(w, resp, result)
	}
	return p.handleJSONResponse(w, resp, result)
}
