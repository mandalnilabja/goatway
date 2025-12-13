# Structure Comparison: Current vs Proposed

## Current Structure (As-Is)

```
goatway/
├── cmd/
│   └── api/
│       └── main.go                    # Everything mixed: init, routes, server
│
├── internal/
│   └── handlers/
│       ├── handlers.go                # Health, cache handlers
│       └── proxy.go                   # OpenRouter proxy (hardcoded)
│
├── .gitignore
├── AGENTS.md
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
├── README.md
└── ROADMAP.md
```

### Current Issues
1. ❌ Hardcoded OpenRouter URL in [`proxy.go:12`](internal/handlers/proxy.go:12)
2. ❌ No provider abstraction for adding new LLM providers
3. ❌ All initialization logic in main.go
4. ❌ No configuration management
5. ❌ Mixed concerns in handlers package

## Proposed Structure (To-Be)

```
goatway/
├── cmd/
│   └── api/
│       └── main.go                    # Minimal: wire dependencies & start
│
├── internal/
│   ├── app/
│   │   ├── server.go                  # Server setup & lifecycle
│   │   └── router.go                  # Route definitions
│   │
│   ├── config/
│   │   └── config.go                  # Env-based configuration
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── health.go          # Health check
│   │       │   ├── cache.go           # Cache demo
│   │       │   └── proxy.go           # Provider-agnostic proxy
│   │       └── middleware/
│   │           └── logging.go         # Optional request logging
│   │
│   ├── provider/
│   │   ├── provider.go                # Provider interface
│   │   ├── openrouter.go              # OpenRouter impl
│   │   ├── openai.go                  # OpenAI impl (future)
│   │   └── azure.go                   # Azure impl (future)
│   │
│   └── domain/
│       └── errors.go                  # Domain errors
│
├── pkg/                               # Reusable packages (if needed)
│   └── logger/
│       └── logger.go
│
├── docs/
│   ├── PROJECT_STRUCTURE.md           # This document
│   └── STRUCTURE_COMPARISON.md        # Structure comparison
│
├── scripts/
│   └── dev.sh                         # Development helpers
│
├── .gitignore
├── AGENTS.md
├── CONTRIBUTING.md
├── go.mod
├── go.sum
├── LICENSE
├── Makefile
├── README.md
└── ROADMAP.md
```

### Proposed Benefits
1. ✅ Provider abstraction: Easy to add OpenAI, Azure, Anthropic, etc.
2. ✅ Separation of concerns: Transport, domain, infrastructure
3. ✅ Configuration management: Environment-based settings
4. ✅ Maintainable: Clear boundaries between layers
5. ✅ Testable: Each layer can be tested independently
6. ✅ AGENTS.md compliant: No streaming risks, follows all constraints

## Side-by-Side Comparison

| Aspect | Current | Proposed |
|--------|---------|----------|
| **Provider Support** | Hardcoded OpenRouter only | Interface-based, multi-provider |
| **Configuration** | Hardcoded in source | Environment variables |
| **Handler Organization** | Single package | Separated by concern |
| **Testability** | Difficult to mock | Easy to test each layer |
| **Adding New Provider** | Modify proxy.go directly | Create new provider file |
| **Streaming Safety** | ✅ Safe | ✅ Safe (preserved) |
| **AGENTS.md Compliance** | ✅ Compliant | ✅ Compliant |
| **Code Readability** | Mixed concerns | Clear separation |

## Key Files Transformation

### 1. main.go
**Before:**
```go
// Everything in one place
func main() {
    cache, _ := ristretto.NewCache(...)
    h := handlers.NewRepo(cache)
    mux := http.NewServeMux()
    mux.HandleFunc("GET /", h.Home)
    mux.HandleFunc("GET /api/health", h.HealthCheck)
    mux.HandleFunc("POST /v1/chat/completions", h.OpenAIProxy)
    srv := &http.Server{Addr: ":8080", Handler: mux, ...}
    srv.ListenAndServe()
}
```

**After:**
```go
// Minimal wiring
func main() {
    cfg := config.Load()
    cache, _ := ristretto.NewCache(...)
    provider := selectProvider(cfg)
    repo := handler.NewRepo(cache, provider)
    router := app.NewRouter(repo)
    server := app.NewServer(cfg, router)
    server.Start()
}
```

### 2. Provider Implementation
**Before:** Hardcoded in [`proxy.go`](internal/handlers/proxy.go)
```go
targetURL := "https://openrouter.ai/api/v1/chat/completions"
upstreamReq.Header.Set("HTTP-Referer", "https://github.com/mandalnilabja/goatway")
upstreamReq.Header.Set("X-Title", "Goatway Proxy")
```

**After:** Abstracted in `internal/provider/openrouter.go`
```go
type OpenRouterProvider struct {
    BaseURL string
    APIKey  string
}

func (p *OpenRouterProvider) ProxyRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Streaming-safe proxy logic
}
```

### 3. Handler Repository
**Before:**
```go
type Repo struct {
    Cache *ristretto.Cache[string, any]
}
```

**After:**
```go
type Repo struct {
    Cache    *ristretto.Cache[string, any]
    Provider provider.Provider
}
```

## Migration Strategy

### Phase 1: Create New Structure (No Breaking Changes)
```bash
# Create directories
mkdir -p internal/{app,config,transport/http/{handler,middleware},provider,domain}
mkdir -p pkg/logger docs scripts
```

### Phase 2: Move Existing Code
```bash
# Move handlers
mv internal/handlers/handlers.go internal/transport/http/handler/cache.go
mv internal/handlers/proxy.go internal/transport/http/handler/proxy.go
```

### Phase 3: Extract Provider Logic
- Create [`provider.go`](internal/provider/provider.go) interface
- Create [`openrouter.go`](internal/provider/openrouter.go) implementation
- Update [`proxy.go`](internal/transport/http/handler/proxy.go) to use interface

### Phase 4: Add Configuration
- Create [`config.go`](internal/config/config.go)
- Load from environment variables
- Update main.go to use config

### Phase 5: Organize App Layer
- Create [`server.go`](internal/app/server.go) for server lifecycle
- Create [`router.go`](internal/app/router.go) for route registration
- Simplify main.go to wire components

## Testing the Migration

After each phase:
1. ✅ Run `go build ./...`
2. ✅ Run `go test ./...`
3. ✅ Test streaming endpoint manually
4. ✅ Verify health check still works
5. ✅ Check memory usage (should be same)

## Compatibility Matrix

| Feature | Current | Proposed | Status |
|---------|---------|----------|--------|
| Streaming | ✅ Works | ✅ Preserved | No Risk |
| Caching | ✅ Works | ✅ Preserved | No Risk |
| Health Check | ✅ Works | ✅ Preserved | No Risk |
| OpenRouter | ✅ Works | ✅ Preserved | No Risk |
| Dependencies | Minimal | Minimal | No Change |
| Performance | Good | Same | No Impact |

## Recommended Next Steps

1. **Get Approval**: Review this structure with team/stakeholders
2. **Create Branch**: `git checkout -b refactor/project-structure`
3. **Phase 1**: Create directory structure
4. **Phase 2**: Move files, update imports
5. **Phase 3**: Extract provider abstraction
6. **Phase 4**: Add configuration layer
7. **Phase 5**: Organize app layer
8. **Test**: Comprehensive testing at each step
9. **Document**: Update README with new structure
10. **Merge**: After all tests pass

## Rollback Plan

If issues arise during migration:
1. Each phase is a separate commit
2. Can revert to any previous working state
3. Keep old branch until fully validated
4. Have backup of working version

## Success Criteria

- ✅ All existing functionality preserved
- ✅ Streaming still works perfectly
- ✅ Can add new provider in <30 minutes
- ✅ Tests pass
- ✅ Documentation updated
- ✅ AGENTS.md compliance maintained
- ✅ No new dependencies added
- ✅ Performance not degraded