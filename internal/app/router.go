package app

import (
	"log/slog"
	"net/http"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/ratelimit"
)

// RouterOptions configures the HTTP router behavior.
type RouterOptions struct {
	EnableWebUI  bool
	Logger       *slog.Logger
	Storage      storage.Storage
	APIKeyCache  *ristretto.Cache[string, *auth.CachedAPIKey]
	SessionStore *auth.SessionStore
	RateLimiter  *ratelimit.Limiter
}

// NewRouter creates and configures the HTTP router with all application routes.
// Returns an http.Handler with middleware applied.
// opts must not be nil - all routes require authentication configuration.
func NewRouter(repo *handler.Repo, opts *RouterOptions) http.Handler {
	mux := http.NewServeMux()

	// Public routes (no auth)
	mux.HandleFunc("GET /api/health", repo.Infra.HealthCheck)
	mux.HandleFunc("GET /api/data", repo.Infra.GetCachedData)

	// Create middleware chain for proxy routes: auth → rate limit
	apiKeyAuth := auth.APIKeyAuth(opts.Storage, opts.APIKeyCache)
	rateLimitMw := ratelimit.Middleware(opts.RateLimiter)

	// withProxy chains auth and rate limiting for proxy handlers
	withProxy := func(h http.HandlerFunc) http.Handler {
		return apiKeyAuth(rateLimitMw(h))
	}

	// Proxy routes (require API key auth + rate limiting)
	mux.Handle("POST /v1/chat/completions", withProxy(repo.Proxy.ChatCompletions))
	mux.Handle("GET /v1/models", withProxy(repo.Proxy.ListModels))
	mux.Handle("GET /v1/models/{model}", withProxy(repo.Proxy.GetModel))
	mux.Handle("POST /v1/embeddings", withProxy(repo.Proxy.Embeddings))
	mux.Handle("POST /v1/audio/speech", withProxy(repo.Proxy.TextToSpeech))
	mux.Handle("POST /v1/audio/transcriptions", withProxy(repo.Proxy.Transcription))
	mux.Handle("POST /v1/audio/translations", withProxy(repo.Proxy.Translation))
	mux.Handle("POST /v1/images/generations", withProxy(repo.Proxy.ImageGeneration))
	mux.Handle("POST /v1/images/edits", withProxy(repo.Proxy.ImageEdit))
	mux.Handle("POST /v1/images/variations", withProxy(repo.Proxy.ImageVariation))
	mux.Handle("POST /v1/completions", withProxy(repo.Proxy.LegacyCompletion))
	mux.Handle("POST /v1/moderations", withProxy(repo.Proxy.Moderation))

	// Admin API routes (require admin auth)
	registerAdminRoutes(mux, repo, opts)

	// Root returns JSON status (per PRD requirement)
	mux.HandleFunc("GET /", repo.Infra.RootStatus)

	// Web UI routes (if enabled)
	if opts.EnableWebUI {
		registerWebUIRoutes(mux, repo, opts)
	}

	// Apply middleware chain (order: outer to inner)
	var h http.Handler = mux

	// Request logging (if logger provided)
	if opts.Logger != nil {
		h = middleware.RequestLogger(opts.Logger)(h)
	}

	// Request ID (always applied)
	h = middleware.RequestID(h)

	// CORS (always applied for Web UI compatibility)
	h = middleware.CORS(h)

	return h
}

// registerAdminRoutes adds all admin API routes to the router.
func registerAdminRoutes(mux *http.ServeMux, repo *handler.Repo, opts *RouterOptions) {
	// Create admin auth middleware using session store (session-only, no Bearer fallback)
	adminAuth := auth.AdminAuth(opts.SessionStore)

	// Helper to wrap handler with admin auth
	withAuth := func(h http.HandlerFunc) http.Handler {
		return adminAuth(h)
	}

	// Credential management
	mux.Handle("POST /api/admin/credentials", withAuth(repo.Admin.CreateCredential))
	mux.Handle("GET /api/admin/credentials", withAuth(repo.Admin.ListCredentials))
	mux.Handle("GET /api/admin/credentials/{id}", withAuth(repo.Admin.GetCredential))
	mux.Handle("PUT /api/admin/credentials/{id}", withAuth(repo.Admin.UpdateCredential))
	mux.Handle("DELETE /api/admin/credentials/{id}", withAuth(repo.Admin.DeleteCredential))

	// API key management
	mux.Handle("POST /api/admin/apikeys", withAuth(repo.Admin.CreateAPIKey))
	mux.Handle("GET /api/admin/apikeys", withAuth(repo.Admin.ListAPIKeys))
	mux.Handle("GET /api/admin/apikeys/{id}", withAuth(repo.Admin.GetAPIKeyByID))
	mux.Handle("PUT /api/admin/apikeys/{id}", withAuth(repo.Admin.UpdateAPIKey))
	mux.Handle("DELETE /api/admin/apikeys/{id}", withAuth(repo.Admin.DeleteAPIKey))
	mux.Handle("POST /api/admin/apikeys/{id}/rotate", withAuth(repo.Admin.RotateAPIKey))

	// Password management
	mux.Handle("PUT /api/admin/password", withAuth(repo.Admin.ChangeAdminPassword))

	// Usage and logs
	mux.Handle("GET /api/admin/usage", withAuth(repo.Admin.GetUsageStats))
	mux.Handle("GET /api/admin/usage/daily", withAuth(repo.Admin.GetDailyUsage))
	mux.Handle("GET /api/admin/logs", withAuth(repo.Admin.GetRequestLogs))
	mux.Handle("DELETE /api/admin/logs", withAuth(repo.Admin.DeleteRequestLogs))

	// System info
	mux.Handle("GET /api/admin/health", withAuth(repo.Admin.AdminHealth))
	mux.Handle("GET /api/admin/info", withAuth(repo.Admin.AdminInfo))
}

// registerWebUIRoutes adds web UI routes with session auth support.
func registerWebUIRoutes(mux *http.ServeMux, repo *handler.Repo, opts *RouterOptions) {
	webUI := repo.WebUI.ServeWebUI()
	sessionAuth := auth.SessionAuth(opts.SessionStore)

	// Login routes (no auth required)
	mux.HandleFunc("GET /web/login", repo.WebUI.LoginPage)
	mux.HandleFunc("POST /web/login", repo.WebUI.Login)
	mux.HandleFunc("POST /web/logout", repo.WebUI.Logout)

	// Static files (no auth)
	mux.Handle("GET /web/static/", webUI)

	// Protected Web UI routes
	mux.Handle("GET /web", sessionAuth(webUI))
	mux.Handle("GET /web/", sessionAuth(webUI))
	mux.Handle("GET /web/credentials", sessionAuth(webUI))
	mux.Handle("GET /web/usage", sessionAuth(webUI))
	mux.Handle("GET /web/logs", sessionAuth(webUI))
	mux.Handle("GET /web/apikeys", sessionAuth(webUI))
	mux.Handle("GET /web/settings", sessionAuth(webUI))
}
