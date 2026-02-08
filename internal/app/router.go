package app

import (
	"log/slog"
	"net/http"

	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware"
)

// RouterOptions configures the HTTP router behavior
type RouterOptions struct {
	EnableWebUI   bool
	AdminPassword string       // Optional password for admin routes
	Logger        *slog.Logger // Logger for request logging
}

// NewRouter creates and configures the HTTP router with all application routes.
// Returns an http.Handler with middleware applied.
func NewRouter(repo *handler.Repo, opts *RouterOptions) http.Handler {
	mux := http.NewServeMux()

	// API routes
	mux.HandleFunc("GET /api/health", repo.HealthCheck)
	mux.HandleFunc("GET /api/data", repo.GetCachedData)

	// Proxy route - forwards to configured LLM provider
	mux.HandleFunc("POST /v1/chat/completions", repo.OpenAIProxy)

	// Models routes - list available models
	mux.HandleFunc("GET /v1/models", repo.ListModels)
	mux.HandleFunc("GET /v1/models/{model}", repo.GetModel)

	// Admin API routes (only register if storage is available)
	if repo.Storage != nil {
		registerAdminRoutes(mux, repo, opts)
	}

	// Web UI or basic home route
	if opts != nil && opts.EnableWebUI {
		registerWebUIRoutes(mux, repo)
	} else {
		mux.HandleFunc("GET /", repo.Home)
	}

	// Apply middleware chain (order: outer to inner)
	var handler http.Handler = mux

	// Request logging (if logger provided)
	if opts != nil && opts.Logger != nil {
		handler = middleware.RequestLogger(opts.Logger)(handler)
	}

	// Request ID (always applied)
	handler = middleware.RequestID(handler)

	// CORS (always applied for Web UI compatibility)
	handler = middleware.CORS(handler)

	return handler
}

// registerAdminRoutes adds all admin API routes to the router
func registerAdminRoutes(mux *http.ServeMux, repo *handler.Repo, opts *RouterOptions) {
	// Get admin password for optional auth
	adminPassword := ""
	if opts != nil {
		adminPassword = opts.AdminPassword
	}

	// Create admin auth middleware (no-op if password is empty)
	adminAuth := middleware.AdminAuth(adminPassword)

	// Helper to wrap handler with admin auth
	withAuth := func(h http.HandlerFunc) http.Handler {
		return adminAuth(h)
	}

	// Credential management
	mux.Handle("POST /api/admin/credentials", withAuth(repo.CreateCredential))
	mux.Handle("GET /api/admin/credentials", withAuth(repo.ListCredentials))
	mux.Handle("GET /api/admin/credentials/{id}", withAuth(repo.GetCredential))
	mux.Handle("PUT /api/admin/credentials/{id}", withAuth(repo.UpdateCredential))
	mux.Handle("DELETE /api/admin/credentials/{id}", withAuth(repo.DeleteCredential))
	mux.Handle("POST /api/admin/credentials/{id}/default", withAuth(repo.SetDefaultCredential))

	// Usage and logs
	mux.Handle("GET /api/admin/usage", withAuth(repo.GetUsageStats))
	mux.Handle("GET /api/admin/usage/daily", withAuth(repo.GetDailyUsage))
	mux.Handle("GET /api/admin/logs", withAuth(repo.GetRequestLogs))
	mux.Handle("DELETE /api/admin/logs", withAuth(repo.DeleteRequestLogs))

	// System info
	mux.Handle("GET /api/admin/health", withAuth(repo.AdminHealth))
	mux.Handle("GET /api/admin/info", withAuth(repo.AdminInfo))
}

// registerWebUIRoutes adds web UI routes with SPA fallback support
func registerWebUIRoutes(mux *http.ServeMux, repo *handler.Repo) {
	webUI := repo.ServeWebUI()

	// Static files
	mux.Handle("GET /static/", webUI)

	// SPA routes - all served by WebUI handler which returns index.html
	mux.Handle("GET /", webUI)
	mux.Handle("GET /credentials", webUI)
	mux.Handle("GET /usage", webUI)
	mux.Handle("GET /logs", webUI)
	mux.Handle("GET /settings", webUI)
}
