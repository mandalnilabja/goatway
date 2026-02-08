package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/mandalnilabja/goatway/internal/types"
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

// handleStreamingResponse processes SSE streaming responses.
func (p *OpenRouterProvider) handleStreamingResponse(w http.ResponseWriter, resp *http.Response, result *ProxyResult) (*ProxyResult, error) {
	// Copy headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		result.Error = io.ErrNoProgress
		return result, nil
	}

	// Process stream while forwarding to client
	processor := NewStreamProcessor()
	err := processor.ProcessReader(resp.Body, func(chunk []byte) error {
		if _, wErr := w.Write(chunk); wErr != nil {
			return wErr
		}
		flusher.Flush()
		return nil
	})

	// Extract results from processor
	result.FinishReason = processor.GetFinishReason()
	if processor.GetModel() != "" {
		result.Model = processor.GetModel()
	}

	// Use upstream usage if available
	if usage := processor.GetUsage(); usage != nil {
		result.PromptTokens = usage.PromptTokens
		result.CompletionTokens = usage.CompletionTokens
		result.TotalTokens = usage.TotalTokens
	}

	if err != nil {
		result.Error = err
	}
	return result, err
}

// handleJSONResponse processes non-streaming JSON responses.
func (p *OpenRouterProvider) handleJSONResponse(w http.ResponseWriter, resp *http.Response, result *ProxyResult) (*ProxyResult, error) {
	// Read full response for parsing
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = err
		http.Error(w, "Failed to read response", http.StatusBadGateway)
		return result, err
	}

	// Parse response to extract usage
	var completion types.ChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err == nil {
		if completion.Usage != nil {
			result.PromptTokens = completion.Usage.PromptTokens
			result.CompletionTokens = completion.Usage.CompletionTokens
			result.TotalTokens = completion.Usage.TotalTokens
		}
		if len(completion.Choices) > 0 {
			result.FinishReason = completion.Choices[0].FinishReason
		}
		if completion.Model != "" {
			result.Model = completion.Model
		}
	}

	// Forward response to client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)

	return result, nil
}

// handleErrorResponse forwards error responses and extracts error info.
func (p *OpenRouterProvider) handleErrorResponse(w http.ResponseWriter, resp *http.Response, result *ProxyResult) (*ProxyResult, error) {
	body, _ := io.ReadAll(resp.Body)

	// Try to extract error message
	var apiErr types.APIError
	if err := json.Unmarshal(body, &apiErr); err == nil {
		result.ErrorMessage = apiErr.Error.Message
	}

	// Forward error to client
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	_, _ = w.Write(body)

	return result, nil
}
