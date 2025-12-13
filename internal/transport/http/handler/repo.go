package handler

import (
	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/provider"
)

// Repo holds the dependencies for HTTP handlers
type Repo struct {
	Cache    *ristretto.Cache[string, any]
	Provider provider.Provider
}

// NewRepo creates a new instance of the handler repository
func NewRepo(cache *ristretto.Cache[string, any], provider provider.Provider) *Repo {
	return &Repo{
		Cache:    cache,
		Provider: provider,
	}
}
