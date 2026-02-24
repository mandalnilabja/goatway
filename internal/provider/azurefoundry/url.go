// Package azurefoundry implements the Azure AI Foundry LLM provider.
package azurefoundry

import (
	"fmt"
	"net/url"
	"strings"
)

// buildTargetURL constructs the Azure AI Foundry target URL from an endpoint.
// Handles endpoints with or without https:// prefix.
func buildTargetURL(endpoint, apiVersion string) (string, error) {
	// Strip protocol prefix if present
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")

	// Parse to validate and extract components
	parsed, err := url.Parse("https://" + endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid endpoint: %w", err)
	}

	// Build target URL with just the host and the correct path
	return fmt.Sprintf("https://%s/models/chat/completions?api-version=%s",
		parsed.Host, apiVersion), nil
}
