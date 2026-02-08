package handler

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/tokenizer"
)

// Repo holds the dependencies for HTTP handlers
type Repo struct {
	Cache     *ristretto.Cache[string, any]
	Provider  provider.Provider
	Storage   storage.Storage
	Tokenizer tokenizer.Tokenizer
	StartTime time.Time
}

// NewRepo creates a new instance of the handler repository
func NewRepo(cache *ristretto.Cache[string, any], prov provider.Provider, store storage.Storage, tok tokenizer.Tokenizer) *Repo {
	return &Repo{
		Cache:     cache,
		Provider:  prov,
		Storage:   store,
		Tokenizer: tok,
		StartTime: time.Now(),
	}
}
