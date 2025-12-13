# Goatway Project Structure Proposal

## Current Structure
```
goatway/
├── cmd/
│   └── api/
│       └── main.go
├── internal/
│   └── handlers/
│       ├── handlers.go
│       └── proxy.go
├── go.mod
├── go.sum
└── docs/
```

## Proposed Structure

```
goatway/
├── cmd/
│   └── api/
│       └── main.go              # Entry point only
│
├── internal/
│   ├── app/
│   │   ├── server.go            # Server initialization and lifecycle
│   │   └── router.go            # Route definitions
│   │
│   ├── config/
│   │   └── config.go            # Configuration loading (env vars)
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── health.go    # Health check handler
│   │       │   ├── cache.go     # Cache demo handler
│   │       │   └── proxy.go     # Main proxy handler
│   │       └── middleware/
│   │           └── logging.go   # Request logging (optional)
│   │
│   ├── provider/
│   │   ├── provider.go          # Provider interface
│   │   ├── openrouter.go        # OpenRouter implementation
│   │   ├── openai.go            # OpenAI implementation (future)
│   │   └── azure.go             # Azure OpenAI (future)
│   │
│   └── domain/
│       └── errors.go            # Domain-specific errors
│
├── pkg/                         # Public utilities (if needed)
│   └── logger/
│       └── logger.go
│
├── docs/
│   └── PROJECT_STRUCTURE.md     # This file
│
├── scripts/
│   └── dev.sh                   # Development helper scripts
│
├── go.mod
├── go.sum
├── Makefile
├── README.md
├── AGENTS.md
├── CONTRIBUTING.md
├── ROADMAP.md
└── LICENSE
```

## Design Decisions

### 1. Provider Architecture
**Location**: `internal/provider/`

Each LLM provider gets its own file implementing a common interface:

```go
// internal/provider/provider.go
type Provider interface {
    Name() string
    ProxyRequest(ctx context.Context, req *http.Request) (*http.Response, error)
    PrepareRequest(ctx context.Context, req *http.Request) error
}
```

**Benefits**:
- Easy to add new providers
- Each provider encapsulates its specific logic
- Maintains streaming compatibility
- No breaking changes to existing code

### 2. Repository Pattern (Maintained)
**Location**: `internal/transport/http/handler/`

Handlers remain as methods on the repository struct:
```go
type Repo struct {
    Cache    *ristretto.Cache[string, any]
    Provider provider.Provider
}
```

This satisfies AGENTS.md requirement: "HTTP handlers as methods on repository struct"

### 3. Configuration
**Location**: `internal/config/`

Simple environment-based configuration:
```go
type Config struct {
    ServerAddr  string
    Provider    string // "openrouter", "openai", etc.
    APIKey      string
    LogLevel    string
}
```

### 4. Transport Layer
**Location**: `internal/transport/http/`

Separates HTTP concerns:
- **handler/**: Business logic handlers
- **middleware/**: Optional cross-cutting concerns (logging, metrics)

### 5. Application Layer
**Location**: `internal/app/`

- **server.go**: Server initialization, graceful shutdown
- **router.go**: Route registration (maps URLs to handlers)

## Migration Path

### Phase 1: Structural Reorganization (Non-Breaking)
1. Create new directory structure
2. Move existing code to new locations
3. Update imports
4. Verify streaming still works

### Phase 2: Provider Abstraction
1. Create provider interface
2. Extract OpenRouter logic to `internal/provider/openrouter.go`
3. Update proxy handler to use provider interface
4. Add tests

### Phase 3: Additional Providers (Future)
1. Implement OpenAI provider
2. Implement Azure OpenAI provider
3. Add provider selection via config

## Compliance with AGENTS.md

✅ **Entry point logic in main command file only**
- `cmd/api/main.go` only initializes and wires dependencies

✅ **HTTP handlers as methods on repository struct**
- All handlers in `internal/transport/http/handler/` are methods on `Repo`

✅ **Shared dependencies in repository struct**
- Cache, Provider, and other dependencies injected into `Repo`

✅ **No global state outside initialization**
- All state managed through dependency injection

✅ **No cross-package imports from internal packages**
- Internal packages don't import each other's internals

✅ **Streaming correctness preserved**
- Provider interface designed to maintain streaming semantics
- No buffering or transformation of responses

✅ **Prefer standard library**
- Only external dependency: ristretto (already in use)
- No frameworks or middleware stacks added

## File Organization Principles

1. **Domain-driven**: Group by feature/domain (provider, transport)
2. **Layered**: Separate concerns (transport, domain, infrastructure)
3. **Testable**: Each layer can be tested independently
4. **Minimal**: No unnecessary abstractions
5. **Explicit**: Clear naming and structure

## Example: Adding a New Provider

```go
// internal/provider/anthropic.go
package provider

import (
    "context"
    "net/http"
)

type AnthropicProvider struct {
    APIKey  string
    BaseURL string
}

func NewAnthropicProvider(apiKey string) *AnthropicProvider {
    return &AnthropicProvider{
        APIKey:  apiKey,
        BaseURL: "https://api.anthropic.com/v1",
    }
}

func (p *AnthropicProvider) Name() string {
    return "anthropic"
}

func (p *AnthropicProvider) ProxyRequest(ctx context.Context, req *http.Request) (*http.Response, error) {
    // Streaming-safe proxy logic here
    // Similar to current proxy.go implementation
    return nil, nil
}

func (p *AnthropicProvider) PrepareRequest(ctx context.Context, req *http.Request) error {
    // Set Anthropic-specific headers
    req.Header.Set("x-api-key", p.APIKey)
    req.Header.Set("anthropic-version", "2023-06-01")
    return nil
}
```

Then in `cmd/api/main.go`:
```go
// Select provider based on config
var provider provider.Provider
switch cfg.Provider {
case "openrouter":
    provider = provider.NewOpenRouterProvider(cfg.APIKey)
case "anthropic":
    provider = provider.NewAnthropicProvider(cfg.APIKey)
default:
    provider = provider.NewOpenRouterProvider(cfg.APIKey)
}

repo := handler.NewRepo(cache, provider)
```

## Benefits of This Structure

1. **Scalability**: Easy to add new providers without changing core logic
2. **Maintainability**: Clear separation of concerns
3. **Testability**: Each layer can be mocked and tested
4. **AGENTS.md Compliant**: Respects all architectural constraints
5. **Standard Layout**: Follows Go community best practices
6. **Migration Friendly**: Can be adopted incrementally

## Next Steps

1. Get approval for this structure
2. Create implementation plan
3. Execute migration in small, testable steps
4. Add comprehensive tests at each step
5. Document any deviations or learnings