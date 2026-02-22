package admin

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// Handlers holds the dependencies for admin HTTP handlers.
type Handlers struct {
	Storage     storage.Storage
	StartTime   time.Time
	APIKeyCache *ristretto.Cache[string, *auth.CachedAPIKey]
}

// New creates a new instance of admin handlers.
func New(store storage.Storage, startTime time.Time, apiKeyCache *ristretto.Cache[string, *auth.CachedAPIKey]) *Handlers {
	return &Handlers{
		Storage:     store,
		StartTime:   startTime,
		APIKeyCache: apiKeyCache,
	}
}

// InvalidateAPIKeyCache removes a cached API key entry by its prefix.
func (h *Handlers) InvalidateAPIKeyCache(keyPrefix string) {
	if h.APIKeyCache != nil && keyPrefix != "" {
		h.APIKeyCache.Del("apikey:" + keyPrefix)
	}
}
