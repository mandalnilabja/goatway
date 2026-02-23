package provider

import (
	"sync"
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/storage/models"
)

// CredentialResolver resolves and caches credentials by name.
type CredentialResolver struct {
	storage storage.Storage
	cache   map[string]*cachedCredential
	mu      sync.RWMutex
	ttl     time.Duration
}

type cachedCredential struct {
	credential *models.Credential
	expiresAt  time.Time
}

// NewCredentialResolver creates a resolver with the given TTL.
func NewCredentialResolver(store storage.Storage, ttl time.Duration) *CredentialResolver {
	return &CredentialResolver{
		storage: store,
		cache:   make(map[string]*cachedCredential),
		ttl:     ttl,
	}
}

// Resolve returns the credential by name (cached).
func (r *CredentialResolver) Resolve(credentialName string) (*models.Credential, error) {
	// Check cache first
	r.mu.RLock()
	if cached, ok := r.cache[credentialName]; ok && time.Now().Before(cached.expiresAt) {
		r.mu.RUnlock()
		return cached.credential, nil
	}
	r.mu.RUnlock()

	// Cache miss or expired - fetch from storage
	cred, err := r.storage.GetCredentialByName(credentialName)
	if err != nil {
		return nil, err
	}

	// Update cache
	r.mu.Lock()
	r.cache[credentialName] = &cachedCredential{
		credential: cred,
		expiresAt:  time.Now().Add(r.ttl),
	}
	r.mu.Unlock()

	return cred, nil
}

// Invalidate removes a cached credential (call after credential update).
func (r *CredentialResolver) Invalidate(credentialName string) {
	r.mu.Lock()
	delete(r.cache, credentialName)
	r.mu.Unlock()
}
