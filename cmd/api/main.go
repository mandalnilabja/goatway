package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/app"
	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/tokenizer"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
)

// Version information - set via ldflags at build time
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

func main() {
	// Parse CLI flags
	var (
		addr         = flag.String("addr", "", "Server address (overrides SERVER_ADDR)")
		dataDir      = flag.String("data-dir", "", "Data directory path (overrides GOATWAY_DATA_DIR)")
		showVer      = flag.Bool("version", false, "Print version and exit")
		versionShort = flag.Bool("v", false, "Print version and exit (shorthand)")
	)
	flag.Parse()

	// Handle version flag
	if *showVer || *versionShort {
		printVersion()
		os.Exit(0)
	}

	// Apply CLI flag overrides to environment (they take precedence)
	if *addr != "" {
		os.Setenv("SERVER_ADDR", *addr)
	}
	if *dataDir != "" {
		os.Setenv("GOATWAY_DATA_DIR", *dataDir)
	}

	// 1. Load Configuration
	cfg := config.Load()

	// 2. Initialize Data Directory
	if err := config.EnsureDataDir(); err != nil {
		log.Fatal("Failed to create data directory:", err)
	}

	// 3. Initialize Storage
	store, err := storage.NewSQLiteStorage(config.DBPath())
	if err != nil {
		log.Fatal("Failed to initialize storage:", err)
	}
	defer store.Close()

	// Run database migrations
	if err := store.Migrate(); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// 4. Initialize Cache
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}

	// 5. Initialize Provider based on configuration
	var llmProvider provider.Provider
	switch cfg.Provider {
	case "openrouter":
		llmProvider = provider.NewOpenRouterProvider(cfg.OpenRouterAPIKey)
	// Future providers can be added here:
	// case "openai":
	//     llmProvider = provider.NewOpenAIProvider(cfg.OpenAIAPIKey, cfg.OpenAIOrg)
	// case "azure":
	//     llmProvider = provider.NewAzureProvider(cfg.AzureAPIKey, cfg.AzureEndpoint)
	default:
		// Default to OpenRouter if unspecified or unknown
		llmProvider = provider.NewOpenRouterProvider(cfg.OpenRouterAPIKey)
	}

	// 6. Initialize Tokenizer for token counting
	tok := tokenizer.New()

	// 7. Initialize Handler Repository with dependencies
	repo := handler.NewRepo(cache, llmProvider, store, tok)

	// 8. Setup Logger for request logging
	logger := setupLogger(cfg)

	// 9. Setup Router with all routes
	routerOpts := &app.RouterOptions{
		EnableWebUI:   cfg.EnableWebUI,
		AdminPassword: cfg.AdminPassword,
		Logger:        logger,
	}
	router := app.NewRouter(repo, routerOpts)

	// 10. Print startup info
	printStartupBanner(cfg)

	// 11. Create and Start Server
	server := app.NewServer(cfg, router)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

func printVersion() {
	fmt.Printf("goatway %s\n", Version)
	fmt.Printf("  commit:  %s\n", Commit)
	fmt.Printf("  built:   %s\n", BuildTime)
}

func printStartupBanner(cfg *config.Config) {
	fmt.Fprintf(os.Stderr, "Goatway %s - Local OpenAI-Compatible Proxy\n", Version)
	fmt.Fprintln(os.Stderr, "========================================")
	if cfg.EnableWebUI {
		fmt.Fprintf(os.Stderr, "Web UI:     http://localhost%s\n", cfg.ServerAddr)
	}
	fmt.Fprintf(os.Stderr, "Proxy API:  http://localhost%s/v1/chat/completions\n", cfg.ServerAddr)
	fmt.Fprintf(os.Stderr, "Admin API:  http://localhost%s/api/admin/\n", cfg.ServerAddr)
	fmt.Fprintf(os.Stderr, "Data:       %s\n", config.DataDir())
	fmt.Fprintln(os.Stderr, "========================================")
}

func setupLogger(cfg *config.Config) *slog.Logger {
	var handler slog.Handler
	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.LogLevel),
	}

	if cfg.LogFormat == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
