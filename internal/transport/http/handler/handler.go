package handler

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/tokenizer"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/admin"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/infra"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/proxy"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler/webui"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// Repo composes all domain-specific handlers.
type Repo struct {
	Admin *admin.Handlers
	WebUI *webui.Handlers
	Proxy *proxy.Handlers
	Infra *infra.Handlers
}

// NewRepo creates a new instance of the composed handler repository.
func NewRepo(cache *ristretto.Cache[string, any], prov provider.Provider, store storage.Storage, tok tokenizer.Tokenizer) *Repo {
	startTime := time.Now()
	return &Repo{
		Admin: admin.New(store, startTime),
		WebUI: webui.New(store, nil), // SessionStore set later
		Proxy: proxy.New(prov, store, tok, cache),
		Infra: infra.New(cache, startTime),
	}
}

// SetSessionStore sets the session store for web UI authentication.
func (r *Repo) SetSessionStore(store *auth.SessionStore) {
	r.WebUI.SetSessionStore(store)
}
