# Provider Registry & Model-Based Routing Plan

## Context

Currently, `main.go` directly imports and instantiates `openrouter.New()`. As more providers are added (OpenAI, Azure, etc.), this approach doesn't scale. The user wants:

1. **Explicit provider registry** - All providers declared in one file
2. **Model-based routing** - Model names like `openrouter-gpt-5` route to `gpt-5` on OpenRouter

## Architecture Overview

```
main.go
    ↓ calls
provider.NewRouter(providers)     ← registry.go creates all providers
    ↓ implements Provider interface
proxy.Handlers uses Router        ← transparent to existing code
    ↓ routes based on model
Individual providers (openrouter, openai, etc.)
```

## Implementation

### 1. Create Provider Registry (`internal/provider/registry.go`)

Explicit initialization of all providers in one place:

```go
package provider

import "github.com/.../openrouter"

// NewProviders creates all available providers.
// Add new providers here when implementing them.
func NewProviders() map[string]Provider {
    return map[string]Provider{
        "openrouter": openrouter.New(),
        // "openai": openai.New(),
        // "azure": azure.New(),
    }
}
```

### 2. Create Provider Router (`internal/provider/router.go`)

Routes requests based on model prefix:

```go
package provider

type Router struct {
    providers map[string]Provider
    fallback  string // default provider if no prefix
}

func NewRouter(providers map[string]Provider, fallback string) *Router

// Implements Provider interface
func (r *Router) Name() string
func (r *Router) BaseURL() string
func (r *Router) PrepareRequest(ctx, req) error
func (r *Router) ProxyRequest(ctx, w, req, opts) (*ProxyResult, error)

// Internal: parses "openrouter-gpt-5" → ("openrouter", "gpt-5")
func (r *Router) resolveProvider(model string) (Provider, string, error)
```

**Model parsing logic:**
- `openrouter-gpt-5` → provider=`openrouter`, model=`gpt-5`
- `gpt-4` (no prefix) → uses fallback provider, model unchanged
- Unknown prefix → error

### 3. Update main.go

Replace direct provider instantiation:

```go
// Before:
llmProvider := openrouter.New()

// After:
providers := provider.NewProviders()
llmProvider := provider.NewRouter(providers, "openrouter") // fallback
```

### 4. Modify ProxyOptions Model Handling

The `Router.ProxyRequest` will:
1. Parse `opts.Model` to extract provider + actual model
2. Update `opts.Model` to the actual model name
3. Delegate to the resolved provider

## Files to Create/Modify

| File | Action | Lines (est.) |
|------|--------|--------------|
| `internal/provider/registry.go` | Create | ~20 |
| `internal/provider/router.go` | Create | ~80 |
| `cmd/api/main.go` | Modify | ~5 lines changed |

## Verification

1. **Unit tests** for router:
   - Model parsing: `openrouter-gpt-5` → correct provider + model
   - Fallback: `gpt-4` → fallback provider
   - Unknown prefix: returns error

2. **Integration test**:
   - Start server with router
   - Send request with prefixed model
   - Verify correct provider receives request

3. **Manual test**:
   ```bash
   curl -X POST localhost:8080/v1/chat/completions \
     -H "Authorization: Bearer $KEY" \
     -d '{"model": "openrouter-gpt-4", "messages": [...]}'
   ```

## Notes

- Router implements `Provider` interface → no changes to proxy handlers
- Explicit registry (no init() magic) → easy to trace and modify
- Fallback provider handles backward compatibility
