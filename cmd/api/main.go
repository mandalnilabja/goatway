package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/app"
	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/provider/openrouter"
	"github.com/mandalnilabja/goatway/internal/storage"
	"github.com/mandalnilabja/goatway/internal/tokenizer"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
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
		showVer      = flag.Bool("version", false, "Print version and exit")
		versionShort = flag.Bool("v", false, "Print version and exit (shorthand)")
	)
	flag.Parse()

	// Handle version flag
	if *showVer || *versionShort {
		printVersion()
		os.Exit(0)
	}

	// Apply CLI flag override for address
	if *addr != "" {
		os.Setenv("SERVER_ADDR", *addr)
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

	// 8. Initialize Provider (API key resolved per-request from storage)
	llmProvider := openrouter.New()

	// 9. Initialize Tokenizer for token counting
	tok := tokenizer.New()

	// 10. Initialize Handler Repository with dependencies
	repo := handler.NewRepo(cache, llmProvider, store, tok)
	repo.SetSessionStore(sessionStore)

	// 11. Setup Logger for request logging
	logger := setupLogger()

	// 12. Setup Router with all routes
	routerOpts := &app.RouterOptions{
		EnableWebUI:  cfg.EnableWebUI,
		Logger:       logger,
		Storage:      store,
		APIKeyCache:  apiKeyCache,
		SessionStore: sessionStore,
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

// ensureAdminPassword prompts for admin password on first run
func ensureAdminPassword(store storage.Storage) error {
	hasPassword, err := store.HasAdminPassword()
	if err != nil {
		return fmt.Errorf("failed to check admin password: %w", err)
	}

	if hasPassword {
		return nil
	}

	fmt.Println()
	fmt.Println("╔════════════════════════════════════════════════════════════╗")
	fmt.Println("║              FIRST-TIME SETUP REQUIRED                     ║")
	fmt.Println("╚════════════════════════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("No admin password configured. Please set one now.")
	fmt.Println("This password protects the Web UI and Admin API.")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("Enter admin password (alphanumeric, min 8 chars): ")
		password, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		password = strings.TrimSpace(password)

		if !isValidAdminPassword(password) {
			fmt.Println("❌ Password must be alphanumeric with at least 8 characters.")
			fmt.Println()
			continue
		}

		fmt.Print("Confirm password: ")
		confirm, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		confirm = strings.TrimSpace(confirm)

		if password != confirm {
			fmt.Println("❌ Passwords do not match. Please try again.")
			fmt.Println()
			continue
		}

		hash, err := storage.HashPassword(password, storage.DefaultArgon2Params())
		if err != nil {
			return fmt.Errorf("failed to hash password: %w", err)
		}

		if err := store.SetAdminPasswordHash(hash); err != nil {
			return fmt.Errorf("failed to save password: %w", err)
		}

		fmt.Println()
		fmt.Println("✓ Admin password saved successfully!")
		fmt.Println()
		return nil
	}
}

// isValidAdminPassword validates the admin password format
func isValidAdminPassword(password string) bool {
	if len(password) < 8 {
		return false
	}
	for _, c := range password {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')) {
			return false
		}
	}
	return true
}

func printVersion() {
	fmt.Printf("goatway %s\n", Version)
	fmt.Printf("  commit:  %s\n", Commit)
	fmt.Printf("  built:   %s\n", BuildTime)
}

func printStartupBanner(cfg *config.Config) {
	fmt.Fprintf(os.Stderr, "\n")
	fmt.Fprintf(os.Stderr, "🐐 Goatway %s - Local OpenAI-Compatible Proxy\n", Version)
	fmt.Fprintln(os.Stderr, "════════════════════════════════════════════════")
	if cfg.EnableWebUI {
		fmt.Fprintf(os.Stderr, "Web UI:     http://localhost%s/web\n", cfg.ServerAddr)
	}
	fmt.Fprintf(os.Stderr, "Proxy API:  http://localhost%s/v1/chat/completions\n", cfg.ServerAddr)
	fmt.Fprintf(os.Stderr, "Admin API:  http://localhost%s/api/admin/\n", cfg.ServerAddr)
	fmt.Fprintf(os.Stderr, "Data:       %s\n", config.DataDir())
	fmt.Fprintln(os.Stderr, "════════════════════════════════════════════════")
	fmt.Fprintf(os.Stderr, "\n")
}

func setupLogger() *slog.Logger {
	// Use sensible defaults: info level, text format
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	return slog.New(handler)
}
