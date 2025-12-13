package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// Repo holds the dependencies for your handlers (like Cache, DB, etc.)
// We use [string, any] so we can cache any type of data, not just strings.
type Repo struct {
	Cache *ristretto.Cache[string, any]
}

// NewRepo creates a new instance of your handler repository
func NewRepo(cache *ristretto.Cache[string, any]) *Repo {
	return &Repo{
		Cache: cache,
	}
}

// Home is now a method of *Repo
func (h *Repo) Home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to Goatway API!"))
}

// HealthCheck is now a method of *Repo
func (h *Repo) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "active",
		"app":    "goatway",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCachedData demonstrates the caching logic
func (h *Repo) GetCachedData(w http.ResponseWriter, r *http.Request) {
	key := "heavy_computation_result"

	// 1. Check Cache
	value, found := h.Cache.Get(key)
	if found {
		fmt.Println("Cache Hit!")
		w.Header().Set("X-Cache", "HIT")
		// Since we stored it as 'any', we assert it is a string for printing
		w.Write([]byte(fmt.Sprintf("Value from Cache: %v", value)))
		return
	}

	// 2. If not found, simulate heavy work (e.g., DB call)
	fmt.Println("Cache Miss - Computing...")
	time.Sleep(2 * time.Second) // Simulate delay
	computedValue := "This data was computed at " + time.Now().Format(time.RFC3339)

	// 3. Set to Cache
	// Cost is 1 (simple item). We ignore the boolean return here for simplicity.
	h.Cache.Set(key, computedValue, 1)

	// Wait logic is usually for tests, but Ristretto is async.
	// In production HTTP, you don't usually wait, but for this demo,
	// it ensures the next immediate refresh might find it.
	h.Cache.Wait()

	w.Header().Set("X-Cache", "MISS")
	w.Write([]byte(fmt.Sprintf("Value Computed: %v", computedValue)))
}
