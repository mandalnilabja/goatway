// Package ratelimit provides rate limiting middleware using token bucket algorithm.
package ratelimit

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/mandalnilabja/goatway/internal/transport/http/middleware/auth"
)

// bucket represents a token bucket for rate limiting.
type bucket struct {
	tokens   float64
	lastFill time.Time
	mu       sync.Mutex
}

// Limiter tracks rate limits per API key.
type Limiter struct {
	buckets sync.Map // map[keyID]*bucket
}

// New creates a new rate limiter.
func New() *Limiter {
	return &Limiter{}
}

// Allow checks if a request is allowed under the rate limit.
// Returns true if allowed, false if rate limited.
func (l *Limiter) Allow(keyID string, rateLimit int) bool {
	if rateLimit <= 0 {
		return true // 0 = unlimited
	}

	// Get or create bucket for this key
	val, _ := l.buckets.LoadOrStore(keyID, &bucket{
		tokens:   float64(rateLimit),
		lastFill: time.Now(),
	})
	b := val.(*bucket)

	b.mu.Lock()
	defer b.mu.Unlock()

	// Refill tokens based on elapsed time
	now := time.Now()
	elapsed := now.Sub(b.lastFill).Seconds()
	refillRate := float64(rateLimit) / 60.0 // tokens per second
	b.tokens += elapsed * refillRate
	if b.tokens > float64(rateLimit) {
		b.tokens = float64(rateLimit) // cap at max capacity
	}
	b.lastFill = now

	// Check if we have tokens available
	if b.tokens >= 1.0 {
		b.tokens--
		return true
	}
	return false
}

// Middleware returns an HTTP middleware that enforces rate limits.
// Must be used after APIKeyAuth middleware (needs key in context).
func Middleware(limiter *Limiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := auth.GetAPIKey(r.Context())
			if key == nil {
				// No key in context = not authenticated, let handler decide
				next.ServeHTTP(w, r)
				return
			}

			if !limiter.Allow(key.ID, key.RateLimit) {
				writeTooManyRequests(w)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// writeTooManyRequests writes a JSON 429 response.
func writeTooManyRequests(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Retry-After", "60")
	w.WriteHeader(http.StatusTooManyRequests)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"message": "rate limit exceeded",
			"type":    "rate_limit_error",
		},
	})
}
