package main

import (
	"log"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/app"
	"github.com/mandalnilabja/goatway/internal/config"
	"github.com/mandalnilabja/goatway/internal/provider"
	"github.com/mandalnilabja/goatway/internal/transport/http/handler"
)

func main() {
	// 1. Load Configuration
	cfg := config.Load()

	// 2. Initialize Cache
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		log.Fatal("Failed to initialize cache:", err)
	}

	// 3. Initialize Provider based on configuration
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

	// 4. Initialize Handler Repository with dependencies
	repo := handler.NewRepo(cache, llmProvider)

	// 5. Setup Router with all routes
	router := app.NewRouter(repo)

	// 6. Create and Start Server
	server := app.NewServer(cfg, router)
	if err := server.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
