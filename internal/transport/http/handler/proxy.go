package handler

import (
	"net/http"
)

// OpenAIProxy forwards requests to the configured LLM provider
// This handler is provider-agnostic and delegates to the provider implementation
func (h *Repo) OpenAIProxy(w http.ResponseWriter, r *http.Request) {
	// Delegate to the provider's ProxyRequest method
	// The provider handles all streaming logic, headers, and provider-specific details
	if err := h.Provider.ProxyRequest(r.Context(), w, r); err != nil {
		// Error handling is already done in the provider implementation
		// The response may already be partially written, so we can't use http.Error here
		return
	}
}
