package infra

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// Handlers holds the dependencies for infrastructure HTTP handlers.
type Handlers struct {
	Cache     *ristretto.Cache[string, any]
	StartTime time.Time
}

// New creates a new instance of infrastructure handlers.
func New(cache *ristretto.Cache[string, any], startTime time.Time) *Handlers {
	return &Handlers{
		Cache:     cache,
		StartTime: startTime,
	}
}
