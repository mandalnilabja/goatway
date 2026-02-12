package app

import (
	"log/slog"
	"net/http"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// RouterOptions configures the HTTP router behavior
type RouterOptions struct {
	EnableWebUI  bool
	Logger       *slog.Logger
	Storage      storage.Storage
	APIKeyCache  *ristretto.Cache[string, *auth.CachedAPIKey]
	SessionStore *auth.SessionStore
}

// NewRouter creates and configures the HTTP router with all application routes.
// Returns an http.Handler with middleware applied.
func NewRouter(repo *handler.Repo, opts *RouterOptions) http.Handler {
	mux := http.NewServeMux()

	// Public routes (no auth)
	mux.HandleFunc("GET /api/health", repo.HealthCheck)
	mux.HandleFunc("GET /api/data", repo.GetCachedData)

	// Create API key auth middleware for proxy routes
	var apiKeyAuth func(http.Handler) http.Handler
	if opts != nil && opts.Storage != nil && opts.APIKeyCache != nil {
		apiKeyAuth = auth.APIKeyAuth(opts.Storage, opts.APIKeyCache)
	}

	// Proxy routes (require API key auth if configured)
	if apiKeyAuth != nil {
		// Chat completions
		mux.Handle("POST /v1/chat/completions", apiKeyAuth(http.HandlerFunc(repo.OpenAIProxy)))

		// Models
		mux.Handle("GET /v1/models", apiKeyAuth(http.HandlerFunc(repo.ListModels)))
		mux.Handle("GET /v1/models/{model}", apiKeyAuth(http.HandlerFunc(repo.GetModel)))

		// Embeddings
		mux.Handle("POST /v1/embeddings", apiKeyAuth(http.HandlerFunc(repo.Embeddings)))

		// Audio
		mux.Handle("POST /v1/audio/speech", apiKeyAuth(http.HandlerFunc(repo.TextToSpeech)))
		mux.Handle("POST /v1/audio/transcriptions", apiKeyAuth(http.HandlerFunc(repo.Transcription)))
		mux.Handle("POST /v1/audio/translations", apiKeyAuth(http.HandlerFunc(repo.Translation)))

		// Images
		mux.Handle("POST /v1/images/generations", apiKeyAuth(http.HandlerFunc(repo.ImageGeneration)))
		mux.Handle("POST /v1/images/edits", apiKeyAuth(http.HandlerFunc(repo.ImageEdit)))
		mux.Handle("POST /v1/images/variations", apiKeyAuth(http.HandlerFunc(repo.ImageVariation)))

		// Legacy completions
		mux.Handle("POST /v1/completions", apiKeyAuth(http.HandlerFunc(repo.LegacyCompletion)))

		// Moderations
		mux.Handle("POST /v1/moderations", apiKeyAuth(http.HandlerFunc(repo.Moderation)))
	} else {
		// No auth configured - direct access (backwards compatibility)
		// Chat completions
		mux.HandleFunc("POST /v1/chat/completions", repo.OpenAIProxy)

		// Models
		mux.HandleFunc("GET /v1/models", repo.ListModels)
		mux.HandleFunc("GET /v1/models/{model}", repo.GetModel)

		// Embeddings
		mux.HandleFunc("POST /v1/embeddings", repo.Embeddings)

		// Audio
		mux.HandleFunc("POST /v1/audio/speech", repo.TextToSpeech)
		mux.HandleFunc("POST /v1/audio/transcriptions", repo.Transcription)
		mux.HandleFunc("POST /v1/audio/translations", repo.Translation)

		// Images
		mux.HandleFunc("POST /v1/images/generations", repo.ImageGeneration)
		mux.HandleFunc("POST /v1/images/edits", repo.ImageEdit)
		mux.HandleFunc("POST /v1/images/variations", repo.ImageVariation)

		// Legacy completions
		mux.HandleFunc("POST /v1/completions", repo.LegacyCompletion)

		// Moderations
		mux.HandleFunc("POST /v1/moderations", repo.Moderation)
	}

	// Admin API routes (only register if storage is available)
	if repo.Storage != nil && opts != nil && opts.Storage != nil {
		registerAdminRoutes(mux, repo, opts)
	}

	// Root returns JSON status (per PRD requirement)
	mux.HandleFunc("GET /", repo.RootStatus)

	// Web UI routes (if enabled)
	if opts != nil && opts.EnableWebUI {
		registerWebUIRoutes(mux, repo, opts)
	}

	// Apply middleware chain (order: outer to inner)
	var h http.Handler = mux

	// Request logging (if logger provided)
	if opts != nil && opts.Logger != nil {
		h = middleware.RequestLogger(opts.Logger)(h)
	}

	// Request ID (always applied)
	h = middleware.RequestID(h)

	// CORS (always applied for Web UI compatibility)
	h = middleware.CORS(h)

	return h
}

// registerAdminRoutes adds all admin API routes to the router
func registerAdminRoutes(mux *http.ServeMux, repo *handler.Repo, opts *RouterOptions) {
	// Create admin auth middleware using stored password hash
	adminAuth := auth.AdminAuth(opts.Storage)

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

	// API key management
	mux.Handle("POST /api/admin/apikeys", withAuth(repo.CreateAPIKey))
	mux.Handle("GET /api/admin/apikeys", withAuth(repo.ListAPIKeys))
	mux.Handle("GET /api/admin/apikeys/{id}", withAuth(repo.GetAPIKeyByID))
	mux.Handle("PUT /api/admin/apikeys/{id}", withAuth(repo.UpdateAPIKey))
	mux.Handle("DELETE /api/admin/apikeys/{id}", withAuth(repo.DeleteAPIKey))
	mux.Handle("POST /api/admin/apikeys/{id}/rotate", withAuth(repo.RotateAPIKey))

	// Password management
	mux.Handle("PUT /api/admin/password", withAuth(repo.ChangeAdminPassword))

	// Usage and logs
	mux.Handle("GET /api/admin/usage", withAuth(repo.GetUsageStats))
	mux.Handle("GET /api/admin/usage/daily", withAuth(repo.GetDailyUsage))
	mux.Handle("GET /api/admin/logs", withAuth(repo.GetRequestLogs))
	mux.Handle("DELETE /api/admin/logs", withAuth(repo.DeleteRequestLogs))

	// System info
	mux.Handle("GET /api/admin/health", withAuth(repo.AdminHealth))
	mux.Handle("GET /api/admin/info", withAuth(repo.AdminInfo))
}

// registerWebUIRoutes adds web UI routes with session auth support
func registerWebUIRoutes(mux *http.ServeMux, repo *handler.Repo, opts *RouterOptions) {
	webUI := repo.ServeWebUI()

	// If session store is available, use session auth
	if opts.SessionStore != nil {
		sessionAuth := auth.SessionAuth(opts.SessionStore)

		// Login routes (no auth required)
		mux.HandleFunc("GET /web/login", repo.LoginPage)
		mux.HandleFunc("POST /web/login", repo.Login)
		mux.HandleFunc("POST /web/logout", repo.Logout)

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
	} else {
		// No session auth - serve web UI without protection (legacy behavior)
		mux.Handle("GET /web/static/", webUI)
		mux.Handle("GET /web", webUI)
		mux.Handle("GET /web/", webUI)
		mux.Handle("GET /web/credentials", webUI)
		mux.Handle("GET /web/usage", webUI)
		mux.Handle("GET /web/logs", webUI)
		mux.Handle("GET /web/settings", webUI)
	}
}
