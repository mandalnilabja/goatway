package main

import (
	"log"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"github.com/mandalnilabja/goatway/internal/handlers"
)

func main() {
	// 1. Initialize Ristretto Cache
	cache, err := ristretto.NewCache(&ristretto.Config[string, any]{
		NumCounters: 1e7,     // 10M keys to track frequency
		MaxCost:     1 << 30, // 1GB max cost
		BufferItems: 64,      // Number of keys per Get buffer
	})
	if err != nil {
		panic(err)
	}
	// Ideally close cache on exit, though main exit kills it anyway.
	// In a real graceful shutdown, you'd handle this.
	// defer cache.Close()

	// 2. Initialize Handlers with the Cache
	// We inject the cache into our handler repository
	h := handlers.NewRepo(cache)

	// 3. Initialize Router
	mux := http.NewServeMux()

	// 4. Register Routes
	// Note: We now use 'h.Home' instead of 'handlers.Home'
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /api/health", h.HealthCheck)
	mux.HandleFunc("GET /api/data", h.GetCachedData) // New route

	// 5. Server Configuration
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// 6. Start Server
	log.Println("Goatway server starting on http://localhost:8080")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}
