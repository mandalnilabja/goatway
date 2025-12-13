package provider

import (
	"context"
	"io"
	"net/http"
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
	// OpenRouter requires these headers for rankings/visibility
	req.Header.Set("HTTP-Referer", "https://github.com/mandalnilabja/goatway")
	req.Header.Set("X-Title", "Goatway Proxy")
	return nil
}

// ProxyRequest handles the streaming proxy to OpenRouter
// CRITICAL: Maintains streaming semantics with no buffering
func (p *OpenRouterProvider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
	// Create the upstream request to OpenRouter
	upstreamReq, err := http.NewRequestWithContext(ctx, req.Method, p.BaseURL(), req.Body)
	if err != nil {
		http.Error(w, "Failed to create request", http.StatusInternalServerError)
		return err
	}

	// Header Passthrough (Surgical)
	// Copy all headers except hop-by-hop headers
	for k, v := range req.Header {
		// Skip hop-by-hop headers which shouldn't be forwarded
		if k == "Content-Length" || k == "Connection" || k == "Host" {
			continue
		}
		upstreamReq.Header[k] = v
	}

	// Add provider-specific headers
	if err := p.PrepareRequest(ctx, upstreamReq); err != nil {
		http.Error(w, "Failed to prepare request", http.StatusInternalServerError)
		return err
	}

	// Setup the Client
	// CRITICAL: DisableCompression is required for correct streaming.
	// If compression is enabled, Go asks for gzip. If we just copy gzip bytes to the client,
	// the client (expecting text/event-stream) will fail to parse the chunks.
	client := &http.Client{
		Transport: &http.Transport{
			DisableCompression: true,
		},
	}

	// Execute Request
	resp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, "Bad Gateway: "+err.Error(), http.StatusBadGateway)
		return err
	}
	defer resp.Body.Close()

	// Copy Response Headers
	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)

	// Stream Data (The "Pump")
	// We assert the writer is a Flusher to support streaming
	flusher, ok := w.(http.Flusher)
	if !ok {
		// This should practically never happen in standard Go HTTP servers
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return nil
	}

	// Create a buffer (32KB is a standard clear balance between CPU and Syscalls)
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			// Write the chunk we just read
			if _, wErr := w.Write(buf[:n]); wErr != nil {
				// Client disconnected or network error
				return wErr
			}
			// FLUSH IMMEDIATELY. This is the "async" magic.
			// Without this, Go might buffer 4KB before sending, causing "laggy" streams.
			flusher.Flush()
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			// Log error if needed, but the stream is already dirty
			return err
		}
	}

	return nil
}
