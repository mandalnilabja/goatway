package main

import (
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/handlers"
)

func main() {
	// 1. Initialize Cache (Keeping your existing logic)
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}

	// 2. Initialize Repo
	h := handlers.NewRepo(cache)

	// 3. Router
	mux := http.NewServeMux()

	// Existing routes
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /api/health", h.HealthCheck)
	mux.HandleFunc("GET /api/data", h.GetCachedData)

	// --- NEW PROXY ROUTE ---
	// We map POST /v1/chat/completions to our proxy handler
	mux.HandleFunc("POST /v1/chat/completions", h.OpenAIProxy)

	// 4. Server Config
	srv := &http.Server{
		Addr:    ":8080",
		Handler: mux,
		// IMPORTANT: ReadTimeout can kill long streams!
		// If you expect long AI generations, you might need to bump this or remove it.
		// For now, let's bump it to 5 minutes for safety with LLMs.
		ReadTimeout:  300 * time.Second,
		WriteTimeout: 300 * time.Second,
	}

	log.Println("Goatway server starting on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
