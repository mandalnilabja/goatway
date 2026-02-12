package webui

import (
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// Handlers holds the dependencies for web UI HTTP handlers.
type Handlers struct {
	Storage      storage.Storage
	SessionStore *auth.SessionStore
}

// New creates a new instance of web UI handlers.
func New(store storage.Storage, sessionStore *auth.SessionStore) *Handlers {
	return &Handlers{
		Storage:      store,
		SessionStore: sessionStore,
	}
}

// SetSessionStore sets the session store for web UI authentication.
func (h *Handlers) SetSessionStore(store *auth.SessionStore) {
	h.SessionStore = store
}
