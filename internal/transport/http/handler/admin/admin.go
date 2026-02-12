package admin

import (
	"time"

	"github.com/mandalnilabja/goatway/internal/storage"
)

// Handlers holds the dependencies for admin HTTP handlers.
type Handlers struct {
	Storage   storage.Storage
	StartTime time.Time
}

// New creates a new instance of admin handlers.
func New(store storage.Storage, startTime time.Time) *Handlers {
	return &Handlers{
		Storage:   store,
		StartTime: startTime,
	}
}
