// Package azurefoundry implements the Azure AI Foundry LLM provider.
package azurefoundry

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/mandalnilabja/goatway/internal/types"
)

const defaultAPIVersion = "2024-05-01-preview"

// Provider implements the provider.Provider interface for Azure AI Foundry.
type Provider struct{}

// New creates a new Azure AI Foundry provider instance.
func New() *Provider {
	return &Provider{}
}

// Name returns the provider identifier.
func (p *Provider) Name() string {
	return "azurefoundry"
}

// BaseURL returns empty string as URL is built dynamically from credential.
func (p *Provider) BaseURL() string {
	return ""
}

// PrepareRequest is a no-op for Azure Foundry (no provider-specific headers needed).
func (p *Provider) PrepareRequest(ctx context.Context, req *http.Request) error {
	return nil
}

// ProxyRequest handles the proxy to Azure AI Foundry with result tracking.
func (p *Provider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *types.ProxyOptions) (*types.ProxyResult, error) {
	startTime := time.Now()
	result := &types.ProxyResult{
		Model:        opts.Model,
		PromptTokens: opts.PromptTokens,
		IsStreaming:  opts.IsStreaming,
	}

	// Must have credential
	if opts.Credential == nil {
		result.Error = types.ErrNoAPIKey
		result.StatusCode = http.StatusUnauthorized
		http.Error(w, "No credential configured", http.StatusUnauthorized)
		return result, types.ErrNoAPIKey
	}

	// Extract Azure-specific credential
	azureCred, err := opts.Credential.GetAzureCredential()
	if err != nil {
		result.Error = err
		result.StatusCode = http.StatusUnauthorized
		http.Error(w, "Invalid Azure credential", http.StatusUnauthorized)
		return result, err
	}

	// Build URL with api-version
	apiVersion := azureCred.APIVersion
	if apiVersion == "" {
		apiVersion = defaultAPIVersion
	}
	targetURL, err := buildTargetURL(azureCred.Endpoint, apiVersion)
	if err != nil {
		result.Error = err
		result.StatusCode = http.StatusBadRequest
		http.Error(w, "Invalid endpoint URL", http.StatusBadRequest)
		return result, err
	}

	// Rewrite body with resolved model
	body, err := rewriteModelInBody(opts.Body, req.Body, opts.Model)
	if err != nil {
		result.Error = err
		result.StatusCode = http.StatusBadRequest
		http.Error(w, "Failed to process request body", http.StatusBadRequest)
		return result, err
	}

	// Create upstream request
	upstreamReq, err := http.NewRequestWithContext(ctx, req.Method, targetURL, body)
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

	// Set Azure-style authentication
	upstreamReq.Header.Set("api-key", azureCred.APIKey)

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
