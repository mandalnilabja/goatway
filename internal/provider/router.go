package provider

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/types"
)

// ErrModelNotFound is returned when a model slug cannot be resolved.
var ErrModelNotFound = errors.New("model not found")

// resolvedRoute holds a pre-resolved provider and model for fast lookup.
type resolvedRoute struct {
	provider types.Provider
	model    string
}

// Router routes requests to the appropriate provider based on model aliases.
// It implements the types.Provider interface.
type Router struct {
	providers    map[string]types.Provider
	slugMap      map[string]*resolvedRoute // Pre-resolved for O(1) lookup
	default_     *config.DefaultRoute
	credResolver *CredentialResolver
}

// NewRouter creates a Router with pre-resolved model aliases and credential resolution.
func NewRouter(providers map[string]types.Provider, cfg *config.Config, store storage.Storage) *Router {
	r := &Router{
		providers:    providers,
		slugMap:      make(map[string]*resolvedRoute),
		default_:     cfg.Default,
		credResolver: NewCredentialResolver(store, 5*time.Minute),
	}

	// Build slug map at startup (not per-request)
	for _, alias := range cfg.Models {
		if p, ok := providers[alias.Provider]; ok {
			r.slugMap[alias.Slug] = &resolvedRoute{
				provider: p,
				model:    alias.Model,
			}
		}
	}
	return r
}

// Name returns the router identifier.
func (r *Router) Name() string {
	return "router"
}

// BaseURL returns empty since the router delegates to actual providers.
func (r *Router) BaseURL() string {
	return ""
}

// PrepareRequest is a no-op; the actual provider handles preparation.
func (r *Router) PrepareRequest(ctx context.Context, req *http.Request) error {
	return nil
}

// ProxyRequest resolves the model and credentials, then delegates to the appropriate provider.
func (r *Router) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *types.ProxyOptions) (*types.ProxyResult, error) {
	resolved, err := r.resolveModel(opts.Model)
	if err != nil {
		http.Error(w, "Model not found: "+opts.Model, http.StatusBadRequest)
		return &types.ProxyResult{
			Model:      opts.Model,
			StatusCode: http.StatusBadRequest,
			Error:      err,
		}, err
	}

	// Resolve credential for the target provider
	cred, err := r.credResolver.Resolve(resolved.provider.Name())
	if err != nil {
		http.Error(w, "No credential configured for provider: "+resolved.provider.Name(), http.StatusUnauthorized)
		return &types.ProxyResult{
			Model:      opts.Model,
			StatusCode: http.StatusUnauthorized,
			Error:      err,
		}, err
	}

	// Set credential and model, then delegate
	opts.Credential = cred
	opts.Model = resolved.model
	return resolved.provider.ProxyRequest(ctx, w, req, opts)
}

// resolveModel performs O(1) lookup for a model slug.
func (r *Router) resolveModel(slug string) (*resolvedRoute, error) {
	// Check explicit aliases first
	if route, ok := r.slugMap[slug]; ok {
		return route, nil
	}

	// Fall back to default provider if configured
	if r.default_ != nil {
		if p, ok := r.providers[r.default_.Provider]; ok {
			return &resolvedRoute{
				provider: p,
				model:    slug, // Use original slug as model name
			}, nil
		}
	}

	return nil, ErrModelNotFound
}

// CredentialResolver returns the credential resolver for cache invalidation.
func (r *Router) CredentialResolver() *CredentialResolver {
	return r.credResolver
}
