package admin

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// Handlers holds the dependencies for admin HTTP handlers.
type Handlers struct {
	Storage      storage.Storage
	StartTime    time.Time
	APIKeyCache  *ristretto.Cache[string, *auth.CachedAPIKey]
	CredResolver *provider.CredentialResolver
}

// New creates a new instance of admin handlers.
func New(store storage.Storage, startTime time.Time, apiKeyCache *ristretto.Cache[string, *auth.CachedAPIKey]) *Handlers {
	return &Handlers{
		Storage:     store,
		StartTime:   startTime,
		APIKeyCache: apiKeyCache,
	}
}

// SetCredentialResolver sets the credential resolver for cache invalidation.
func (h *Handlers) SetCredentialResolver(cr *provider.CredentialResolver) {
	h.CredResolver = cr
}

// InvalidateAPIKeyCache removes a cached API key entry by its prefix.
func (h *Handlers) InvalidateAPIKeyCache(keyPrefix string) {
	if h.APIKeyCache != nil && keyPrefix != "" {
		h.APIKeyCache.Del("apikey:" + keyPrefix)
	}
}

// InvalidateCredentialCache removes a cached credential for a provider.
func (h *Handlers) InvalidateCredentialCache(providerName string) {
	if h.CredResolver != nil && providerName != "" {
		h.CredResolver.Invalidate(providerName)
	}
}
