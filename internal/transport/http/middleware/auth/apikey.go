package auth

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/storage"
)

// APIKeyContextKey is the context key for authenticated API key.
type APIKeyContextKey struct{}

// CachedAPIKey holds validated key info for caching.
type CachedAPIKey struct {
	Key        *storage.ClientAPIKey
	ValidUntil time.Time
}

// APIKeyAuth middleware authenticates requests using Goatway API keys.
// Only keys starting with "gw_" are accepted; all other keys are rejected.
func APIKeyAuth(store storage.Storage, cache *ristretto.Cache[string, *CachedAPIKey]) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Extract key from Authorization header
			auth := r.Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				writeUnauthorized(w, "API key required")
				return
			}
			apiKey := strings.TrimPrefix(auth, "Bearer ")

			// Reject non-goatway keys (all clients must use gw_* keys)
			if !strings.HasPrefix(apiKey, storage.APIKeyPrefix) {
				writeUnauthorized(w, "only Goatway API keys (gw_*) are accepted")
				return
			}

			// 2. Check cache first
			prefix := storage.ExtractKeyPrefix(apiKey)
			cacheKey := "apikey:" + prefix

			if cache != nil {
				if cached, found := cache.Get(cacheKey); found {
					if time.Now().Before(cached.ValidUntil) {
						valid, _ := storage.VerifyPassword(apiKey, cached.Key.KeyHash)
						if valid && cached.Key.IsActive && !cached.Key.IsExpired() {
							ctx := context.WithValue(r.Context(), APIKeyContextKey{}, cached.Key)
							next.ServeHTTP(w, r.WithContext(ctx))
							return
						}
					}
				}
			}

			// 3. Lookup in database by prefix
			keys, err := store.GetAPIKeyByPrefix(prefix)
			if err != nil || len(keys) == 0 {
				writeUnauthorized(w, "invalid API key")
				return
			}

			// 4. Verify hash against all matching keys
			var validKey *storage.ClientAPIKey
			for _, k := range keys {
				valid, _ := storage.VerifyPassword(apiKey, k.KeyHash)
				if valid {
					validKey = k
					break
				}
			}

			if validKey == nil || !validKey.IsActive || validKey.IsExpired() {
				writeUnauthorized(w, "invalid or expired API key")
				return
			}

			// 5. Cache valid key for 5 minutes
			if cache != nil {
				cache.Set(cacheKey, &CachedAPIKey{
					Key:        validKey,
					ValidUntil: time.Now().Add(5 * time.Minute),
				}, 1)
			}

			// 6. Update last used timestamp (async)
			go func() { _ = store.UpdateAPIKeyLastUsed(validKey.ID) }()

			// 7. Add to context and proceed
			ctx := context.WithValue(r.Context(), APIKeyContextKey{}, validKey)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetAPIKey retrieves the authenticated API key from context.
func GetAPIKey(ctx context.Context) *storage.ClientAPIKey {
	if key, ok := ctx.Value(APIKeyContextKey{}).(*storage.ClientAPIKey); ok {
		return key
	}
	return nil
}

// RequireScope middleware checks if the authenticated API key has a required scope.
func RequireScope(scope string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := GetAPIKey(r.Context())
			if key == nil {
				writeUnauthorized(w, "authentication required")
				return
			}

			if !key.HasScope(scope) {
				writeForbidden(w, "insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeForbidden writes a JSON 403 response.
func writeForbidden(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_, _ = w.Write([]byte(`{"error":{"message":"` + message + `","type":"permission_error"}}`))
}
