package app

import (
	"net/http"

	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
)

// NewRouter creates and configures the HTTP router with all application routes
func NewRouter(repo *handler.Repo) *http.ServeMux {
	mux := http.NewServeMux()

	// Basic routes
	mux.HandleFunc("GET /", repo.Home)
	mux.HandleFunc("GET /api/health", repo.HealthCheck)
	mux.HandleFunc("GET /api/data", repo.GetCachedData)

	// Proxy route - forwards to configured LLM provider
	mux.HandleFunc("POST /v1/chat/completions", repo.OpenAIProxy)

	return mux
}
