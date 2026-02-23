package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/app"
	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/tokenizer"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/ratelimit"
	"github.com/mandalnilabja/goatway/internal/version"
)

func main() {
	// Parse CLI flags
	var (
		port         = flag.String("port", "", "Server port (overrides SERVER_PORT)")
		showVer      = flag.Bool("version", false, "Print version and exit")
		versionShort = flag.Bool("v", false, "Print version and exit (shorthand)")
	)
	flag.Parse()

	// Handle version flag
	if *showVer || *versionShort {
		printVersion()
		os.Exit(0)
	}

	// Apply CLI flag override for port
	if *port != "" {
		os.Setenv("SERVER_PORT", *port)
	}

	// 1. Load Configuration
	cfg := config.Load()

	// 2. Initialize Data Directory
	if err := config.EnsureDataDir(); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// 3. Initialize Config File (creates template on first run)
	if err := config.EnsureConfigFile(); err != nil {
		log.Fatal("Failed to create config file:", err)
	}

	// 4. Initialize Storage
	store, err := storage.NewSQLiteStorage(config.DBPath())
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	defer store.Close()

	// 4. First-run admin password setup
	if err := ensureAdminPassword(store); err != nil {
		log.Fatal("Failed to setup admin password:", err)
	}

	// 5. Initialize Cache
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}

	// 6. Initialize API Key Cache for authentication
	apiKeyCache, err := ristretto.NewCache(&ristretto.Config[string, *auth.CachedAPIKey]{
		NumCounters: 1e5,
		MaxCost:     1 << 20,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal("Failed to initialize API key cache:", err)
	}

	// 7. Initialize Session Store for Web UI
	sessionStore := auth.NewSessionStore(24 * time.Hour) // 24 hour session TTL

	// 8. Initialize Rate Limiter for API key rate limiting
	rateLimiter := ratelimit.New()

	// 8. Initialize Provider Router (routes models to appropriate providers)
	providers := provider.NewProviders()
	llmProvider := provider.NewRouter(providers, cfg, store)

	// 9. Initialize Tokenizer for token counting
	tok := tokenizer.New()

	// 10. Initialize Handler Repository with dependencies
	repo := handler.NewRepo(cache, llmProvider, store, tok, apiKeyCache)
	repo.SetSessionStore(sessionStore)
	repo.SetCredentialResolver(llmProvider.CredentialResolver())

	// 11. Setup Logger for request logging
	logger := setupLogger()

	// 13. Setup Router with all routes
	routerOpts := &app.RouterOptions{
		EnableWebUI:  cfg.EnableWebUI,
		Logger:       logger,
		Storage:      store,
		APIKeyCache:  apiKeyCache,
		SessionStore: sessionStore,
		RateLimiter:  rateLimiter,
	}
	router := app.NewRouter(repo, routerOpts)

	// 13. Print startup info
	printStartupBanner(cfg)

	// 14. Create and Start Server
	server := app.NewServer(cfg, router)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func printVersion() {
	fmt.Printf("goatway %s\n", version.Version)
	fmt.Printf("  commit:  %s\n", version.Commit)
	fmt.Printf("  built:   %s\n", version.BuildTime)
}
