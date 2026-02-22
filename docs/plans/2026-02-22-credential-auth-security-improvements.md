# Goatway Credential & Auth Security Improvements Plan

## Context

This plan addresses security hardening and feature completion for Goatway's credential and API key system. The current implementation has several gaps:

1. **Auth passthrough vulnerability** - Non-Goatway keys bypass authentication entirely
2. **Credential inflexibility** - Only supports single API key strings, not JSON/structured credentials
3. **Admin auth inconsistency** - Bearer token fallback weakens session-only requirement
4. **Missing rate limiting** - Field exists but enforcement is absent
5. **No cache invalidation** - Revoked keys remain valid for up to 5 minutes

---

## Implementation Plan

### 1. Enforce Goatway-Only Authentication

**Goal:** Reject all non-`gw_*` keys in the API key middleware.

**Files to modify:**
- [internal/transport/http/middleware/auth/apikey.go](../../internal/transport/http/middleware/auth/apikey.go)

**Changes:**
```go
// In APIKeyAuth function, around line 36-40
// REMOVE: Skip logic for non-goatway keys
// REPLACE: Return 401 for any non-gw_ key

// Current (REMOVE):
// Skip if not a goatway key (pass upstream keys through)
if !strings.HasPrefix(apiKey, storage.APIKeyPrefix) {
    next.ServeHTTP(w, r)
    return
}

// New (ADD):
// Reject non-goatway keys
if !strings.HasPrefix(apiKey, storage.APIKeyPrefix) {
    writeUnauthorized(w, "only Goatway API keys (gw_*) are accepted")
    return
}
```

**Impact:** All clients must use Goatway-issued keys. Direct upstream keys no longer work.

---

### 2. Flexible Credential Schema for Multiple Providers

**Goal:** Support JSON-structured credentials (e.g., Azure needs endpoint + key + deployment).

**Files to modify:**
- [internal/storage/models/credential.go](../../internal/storage/models/credential.go) - Schema change
- [internal/storage/sqlite/credentials_write.go](../../internal/storage/sqlite/credentials_write.go) - Storage logic
- [internal/storage/sqlite/credentials_read.go](../../internal/storage/sqlite/credentials_read.go) - Retrieval
- [internal/transport/http/handler/admin/credentials_modify.go](../../internal/transport/http/handler/admin/credentials_modify.go) - API
- [internal/transport/http/handler/proxy/proxy.go](../../internal/transport/http/handler/proxy/proxy.go) - Resolution

**Schema change:**
```go
// credential.go
type Credential struct {
    ID         string          `json:"id"`
    Provider   string          `json:"provider"`
    Name       string          `json:"name"`
    // CHANGE: From APIKey string to flexible CredentialData
    Data       json.RawMessage `json:"data"` // Encrypted at rest, provider-specific format
    IsDefault  bool            `json:"is_default"`
    CreatedAt  time.Time       `json:"created_at"`
    UpdatedAt  time.Time       `json:"updated_at"`
}

// Provider-specific credential formats
type OpenRouterCredential struct {
    APIKey string `json:"api_key"`
}

type AzureCredential struct {
    Endpoint   string `json:"endpoint"`
    APIKey     string `json:"api_key"`
    Deployment string `json:"deployment"`
    APIVersion string `json:"api_version"`
}

type AnthropicCredential struct {
    APIKey string `json:"api_key"`
}
```

**Migration:** Add `data` column, migrate existing `api_key` values to `{"api_key": "..."}` format.

**Credential resolution by model:**
- Router resolves model slug → provider
- Storage fetches default credential for that provider
- Provider-specific code extracts needed fields from credential data

---

### 3. Admin Session-Only Authentication

**Goal:** Remove Bearer token fallback from AdminAuth middleware.

**Files to modify:**
- [internal/transport/http/middleware/auth/admin.go](../../internal/transport/http/middleware/auth/admin.go)

**Changes:**
```go
// AdminAuth - REMOVE Bearer token fallback
func AdminAuth(sessions *SessionStore) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // ONLY session cookie authentication
            if sessions == nil {
                writeUnauthorized(w, "session management not configured")
                return
            }

            cookie, err := r.Cookie("goatway_session")
            if err != nil || cookie.Value == "" {
                writeUnauthorized(w, "session required")
                return
            }

            session := sessions.Get(cookie.Value)
            if session == nil {
                writeUnauthorized(w, "invalid or expired session")
                return
            }

            next.ServeHTTP(w, r)
        })
    }
}
```

**Impact:** Admin API only accessible via web UI login flow. No programmatic Bearer token access.

---

### 4. Rate Limiting Implementation

**Goal:** Enforce requests-per-minute limit stored in `ClientAPIKey.RateLimit`.

**New files:**
- `internal/transport/http/middleware/ratelimit/ratelimit.go` - Token bucket or sliding window

**Files to modify:**
- [internal/transport/http/middleware/auth/apikey.go](../../internal/transport/http/middleware/auth/apikey.go) - Add to context
- [internal/app/router.go](../../internal/app/router.go) - Wire middleware

**Implementation approach:**
```go
// ratelimit/ratelimit.go
type RateLimiter struct {
    cache   *ristretto.Cache[string, *tokenBucket]
    mu      sync.RWMutex
}

type tokenBucket struct {
    tokens    float64
    lastFill  time.Time
    rateLimit int // requests per minute
}

func (rl *RateLimiter) Allow(keyID string, rateLimit int) bool {
    if rateLimit <= 0 {
        return true // 0 = unlimited
    }
    // Token bucket algorithm
    // Refill at rate of (rateLimit / 60) tokens per second
    // Max capacity = rateLimit
}
```

**Middleware chain:**
```
Request → APIKeyAuth → RateLimit → Handler
```

Key available in context from APIKeyAuth, rate limit from `key.RateLimit`.

---

### 5. Cache Invalidation on Key Modification

**Goal:** Immediately invalidate cached API keys when revoked/modified.

**Files to modify:**
- [internal/storage/sqlite/apikeys_write.go](../../internal/storage/sqlite/apikeys_write.go) - Add invalidation hook
- [internal/transport/http/handler/admin/apikeys_modify.go](../../internal/transport/http/handler/admin/apikeys_modify.go) - Call invalidation
- [internal/transport/http/middleware/auth/apikey.go](../../internal/transport/http/middleware/auth/apikey.go) - Add invalidation method

**Approach:**
1. Pass cache reference to admin handlers
2. On Update/Delete, call `cache.Del("apikey:" + key.KeyPrefix)`
3. For credential changes, invalidate by prefix pattern

```go
// admin/apikeys_modify.go
func (h *Handlers) UpdateAPIKey(...) {
    // ... existing update logic ...

    // Invalidate cache
    if h.APIKeyCache != nil {
        h.APIKeyCache.Del("apikey:" + key.KeyPrefix)
    }
}
```

---

## File Impact Summary

| File | Change Type | Description |
|------|------------|-------------|
| `middleware/auth/apikey.go` | Modify | Reject non-gw_ keys, add cache invalidation |
| `middleware/auth/admin.go` | Modify | Remove Bearer fallback |
| `middleware/ratelimit/ratelimit.go` | **New** | Token bucket rate limiter |
| `storage/models/credential.go` | Modify | Change to flexible JSON data |
| `storage/sqlite/credentials_*.go` | Modify | Handle JSON credential data |
| `handler/admin/apikeys_modify.go` | Modify | Cache invalidation on modify |
| `handler/admin/credentials_*.go` | Modify | Support new credential schema |
| `handler/proxy/proxy.go` | Modify | Updated credential resolution |
| `app/router.go` | Modify | Wire rate limit middleware |

---

## Verification

1. **Auth enforcement:**
   - `curl -H "Authorization: Bearer sk-xxx" /v1/models` → 401
   - `curl -H "Authorization: Bearer gw_valid" /v1/models` → 200

2. **Credential flexibility:**
   - Create Azure credential with JSON data
   - Model routed to Azure provider uses credential fields correctly

3. **Admin session-only:**
   - `curl -H "Authorization: Bearer adminpass" /api/admin/...` → 401
   - Login via web UI, use session cookie → 200

4. **Rate limiting:**
   - Create key with rate_limit=10
   - Send 15 requests in 1 minute → 10 succeed, 5 get 429

5. **Cache invalidation:**
   - Create and use key (cached)
   - Revoke key via admin API
   - Immediate next request → 401 (not 5min delay)

---

## Questions Addressed

| User Requirement | Solution |
|-----------------|----------|
| Reject non-gw_ keys | Section 1: Remove passthrough logic |
| Flexible credentials per provider | Section 2: JSON credential data schema |
| Session-only admin auth | Section 3: Remove Bearer fallback |
| Rate limiting | Section 4: Token bucket middleware |
| Cache invalidation | Section 5: Explicit cache.Del on modify |
