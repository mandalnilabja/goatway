# Universal Credential Resolution for Multi-Provider Support

## Context

**Problem:** The current credential resolution is broken for multi-provider routing:

1. `resolveAPIKey()` in proxy handler calls `h.Provider.Name()` which returns `"router"` when using the Router
2. `GetDefaultCredential("router")` fails because no credential exists for provider "router"
3. `ProxyOptions` only has `APIKey string` - Azure needs endpoint, deployment, api_version in addition to the key
4. No credential caching - every request hits the database

**Goal:** Enable any provider (OpenRouter, Azure, future OpenAI/Ollama) to receive the correct credentials from a single credential store, with caching for performance.

---

## Current Architecture

```
Request → Handler.resolveAPIKey() → Storage.GetDefaultCredential(h.Provider.Name())
                                                     ↓
                                           Problem: Returns "router" not actual provider
                                                     ↓
                                        ProxyOptions{APIKey: ""} → Provider fails
```

**Key Files:**
- [proxy.go](../../internal/transport/http/handler/proxy/proxy.go) - `resolveAPIKey()` at line 35
- [provider.go](../../internal/types/provider.go) - `ProxyOptions` struct at line 31
- [router.go](../../internal/provider/router.go) - Router delegates to resolved provider at line 78
- [credential.go](../../internal/storage/models/credential.go) - Credential model with GetAPIKey/GetAzureCredential
- [credentials_read.go](../../internal/storage/sqlite/credentials_read.go) - GetDefaultCredential with decryption

---

## Solution Design

**New Flow:**
```
Request → Handler (no credential resolution)
             ↓
         Router.ProxyRequest(opts)
             ↓
         Router resolves model → target provider
             ↓
         Router resolves credential for target provider (cached)
             ↓
         opts.Credential = resolved credential
             ↓
         target.Provider.ProxyRequest(opts)
             ↓
         Provider extracts what it needs from opts.Credential
```

---

## Implementation Plan

### Step 1: Add Credential to ProxyOptions

**File:** [internal/types/provider.go](../../internal/types/provider.go)

```go
type ProxyOptions struct {
    // Credential from storage (replaces APIKey for providers that need more)
    Credential *models.Credential

    // APIKey is deprecated, use Credential.GetAPIKey() instead
    // Kept for backward compatibility during migration
    APIKey string

    // ... existing fields ...
}
```

---

### Step 2: Create CredentialResolver

**New File:** `internal/provider/credential.go` (~50 lines)

```go
package provider

import (
    "sync"
    "time"
    "github.com/mandalnilabja/goatway/internal/storage"
    "github.com/mandalnilabja/goatway/internal/storage/models"
)

// CredentialResolver resolves and caches credentials by provider name.
type CredentialResolver struct {
    storage storage.Storage
    cache   map[string]*cachedCredential
    mu      sync.RWMutex
    ttl     time.Duration
}

type cachedCredential struct {
    credential *models.Credential
    expiresAt  time.Time
}

func NewCredentialResolver(store storage.Storage, ttl time.Duration) *CredentialResolver

// Resolve returns the default credential for a provider (cached).
func (r *CredentialResolver) Resolve(provider string) (*models.Credential, error)

// Invalidate removes a cached credential (call after credential update).
func (r *CredentialResolver) Invalidate(provider string)
```

---

### Step 3: Update Router to Resolve Credentials

**File:** [internal/provider/router.go](../../internal/provider/router.go)

Add `credResolver *CredentialResolver` to Router struct.

Update `NewRouter` signature:
```go
func NewRouter(providers map[string]types.Provider, cfg *config.Config, store storage.Storage) *Router
```

Update `ProxyRequest`:
```go
func (r *Router) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request, opts *types.ProxyOptions) (*types.ProxyResult, error) {
    resolved, err := r.resolveModel(opts.Model)
    if err != nil {
        // ... error handling ...
    }

    // Resolve credential for the target provider
    cred, err := r.credResolver.Resolve(resolved.provider.Name())
    if err != nil {
        http.Error(w, "No credential configured for provider: "+resolved.provider.Name(), http.StatusUnauthorized)
        return &types.ProxyResult{StatusCode: http.StatusUnauthorized, Error: err}, err
    }

    // Set credential on options
    opts.Credential = cred
    opts.APIKey = cred.GetAPIKey() // Backward compat
    opts.Model = resolved.model

    return resolved.provider.ProxyRequest(ctx, w, req, opts)
}
```

---

### Step 4: Update OpenRouter Provider

**File:** [internal/provider/openrouter/client.go](../../internal/provider/openrouter/client.go)

Change line 52 to use credential:
```go
// Get API key from credential (preferred) or legacy APIKey field
apiKey := opts.APIKey
if opts.Credential != nil {
    apiKey = opts.Credential.GetAPIKey()
}
```

---

### Step 5: Remove resolveAPIKey from Proxy Handler

**File:** [internal/transport/http/handler/proxy/proxy.go](../../internal/transport/http/handler/proxy/proxy.go)

- Remove `resolveAPIKey` method (Router handles this now)
- Update `ChatCompletions` and other handlers to not call `resolveAPIKey`
- The `credID` for logging can come from `opts.Credential.ID` after the proxy call

---

### Step 6: Update main.go

**File:** [cmd/api/main.go](../../cmd/api/main.go)

Pass storage to Router:
```go
llmProvider := provider.NewRouter(providers, cfg, store)
```

---

### Step 7: Invalidate Cache on Credential Changes

**File:** [internal/transport/http/handler/admin/credentials_modify.go](../../internal/transport/http/handler/admin/credentials_modify.go)

Pass CredentialResolver to admin handlers. After create/update/delete credential operations:
```go
if h.credResolver != nil {
    h.credResolver.Invalidate(credential.Provider)
}
```

**File:** [internal/transport/http/handler/admin/admin.go](../../internal/transport/http/handler/admin/admin.go)

Add `CredResolver *provider.CredentialResolver` to Handlers struct.

---

## Files to Modify

| File | Changes |
|------|---------|
| `internal/types/provider.go` | Add `Credential *models.Credential` to ProxyOptions |
| `internal/provider/credential.go` | **New file** - CredentialResolver |
| `internal/provider/router.go` | Add credential resolution after model resolution |
| `internal/provider/openrouter/client.go` | Use opts.Credential |
| `internal/transport/http/handler/proxy/proxy.go` | Remove resolveAPIKey |
| `internal/transport/http/handler/proxy/chat.go` | Update credential handling |
| `cmd/api/main.go` | Pass storage to Router |
| `internal/transport/http/handler/admin/credentials_modify.go` | Cache invalidation |
| `internal/transport/http/handler/admin/admin.go` | Add CredResolver field |

---

## Testing

1. **Unit Tests:**
   - `credential_test.go` - Test caching behavior, TTL expiration, invalidation
   - `router_test.go` - Test credential resolution per-provider

2. **Integration Test:**
   ```bash
   # Create credential for openrouter
   curl -X POST http://localhost:8080/admin/credentials \
     -H "Authorization: Bearer admin-token" \
     -d '{"provider":"openrouter","name":"default","data":{"api_key":"sk-or-xxx"},"is_default":true}'

   # Make a chat request (should use cached credential)
   curl http://localhost:8080/v1/chat/completions \
     -H "Authorization: Bearer gw_xxx" \
     -d '{"model":"gpt-4o","messages":[{"role":"user","content":"hi"}]}'
   ```

3. **Verification:**
   - `make test` - All tests pass
   - `make lint` - No lint errors
   - Manual test: Create credential, make requests, verify caching (check logs for DB hits)

---

## Future Azure Support

With this design, adding Azure is straightforward:

1. Create `internal/provider/azure/` implementing Provider interface
2. In `ProxyRequest`, extract `opts.Credential.GetAzureCredential()` for endpoint/deployment/api_version
3. Register in `provider.NewProviders()`
4. Store Azure credentials as JSON: `{"endpoint":"...","api_key":"...","deployment":"...","api_version":"..."}`
