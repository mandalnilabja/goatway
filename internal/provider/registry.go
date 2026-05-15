package provider

import (
	"github.com/mandalnilabja/goatway/internal/provider/azurefoundry"
	"github.com/mandalnilabja/goatway/internal/provider/openrouter"
)

// NewProviders returns a map of all available LLM providers.
// The map key is the provider identifier used in config routing.
func NewProviders() map[string]Provider {
	return map[string]Provider{
		"openrouter":   openrouter.New(),
		"azurefoundry": azurefoundry.New(),
		// Future providers:
		// "openai": openai.New(),
		// "ollama": ollama.New(),
	}
}
