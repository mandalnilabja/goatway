# Project Structure Recommendation Summary

## Executive Summary

Based on your requirements for a multi-provider Go project following standard layout practices, I recommend the following structure that:

✅ Keeps each LLM provider in a separate file  
✅ Follows Go standard project layout  
✅ Maintains AGENTS.md compliance  
✅ Preserves streaming correctness  
✅ Enables easy provider addition  

## Quick Visual

### Your Request
> "I want to keep every llm provider in a separate file in a directory (llm provider)"

### Solution
```
internal/
├── provider/              ← Your provider directory
│   ├── provider.go        ← Common interface
│   ├── openrouter.go      ← OpenRouter implementation
│   ├── openai.go          ← OpenAI implementation (future)
│   ├── azure.go           ← Azure OpenAI (future)
│   ├── anthropic.go       ← Anthropic (future)
│   └── gemini.go          ← Google Gemini (future)
```

Each provider file is self-contained with:
- Provider-specific configuration
- Provider-specific headers
- Provider-specific URL/endpoint logic
- Streaming-safe proxy implementation

## Recommended Final Structure

```
goatway/
├── cmd/
│   └── api/
│       └── main.go              # Entry point
│
├── internal/
│   ├── app/
│   │   ├── server.go            # Server lifecycle
│   │   └── router.go            # Route registration
│   │
│   ├── config/
│   │   └── config.go            # Configuration management
│   │
│   ├── transport/
│   │   └── http/
│   │       ├── handler/
│   │       │   ├── health.go    # Health endpoint
│   │       │   ├── cache.go     # Cache demo
│   │       │   └── proxy.go     # Provider-agnostic proxy
│   │       └── middleware/
│   │           └── logging.go   # Request logging (optional)
│   │
│   ├── provider/                # ← Your provider directory
│   │   ├── provider.go          # Interface definition
│   │   ├── openrouter.go        # OpenRouter
│   │   ├── openai.go            # OpenAI (future)
│   │   ├── azure.go             # Azure OpenAI (future)
│   │   ├── anthropic.go         # Anthropic (future)
│   │   └── gemini.go            # Google Gemini (future)
│   │
│   └── domain/
│       └── errors.go            # Domain errors
│
├── pkg/                         # Reusable public packages
│   └── logger/
│       └── logger.go
│
├── docs/
│   ├── PROJECT_STRUCTURE.md           # Detailed design
│   ├── STRUCTURE_COMPARISON.md        # Before/after comparison
│   └── STRUCTURE_RECOMMENDATION.md    # This file
│
├── scripts/
│   └── dev.sh                   # Development scripts
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Provider Interface Design

```go
// internal/provider/provider.go
package provider

import (
    "context"
    "net/http"
)

// Provider defines the interface all LLM providers must implement
type Provider interface {
    // Name returns the provider identifier
    Name() string
    
    // BaseURL returns the provider's API endpoint
    BaseURL() string
    
    // PrepareRequest adds provider-specific headers and modifications
    PrepareRequest(ctx context.Context, req *http.Request) error
    
    // ProxyRequest handles the streaming proxy to the provider
    // MUST maintain streaming semantics (no buffering)
    ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error
}
```

## Example Provider Implementation

```go
// internal/provider/openrouter.go
package provider

import (
    "context"
    "net/http"
)

type OpenRouterProvider struct {
    APIKey string
}

func NewOpenRouterProvider(apiKey string) *OpenRouterProvider {
    return &OpenRouterProvider{APIKey: apiKey}
}

func (p *OpenRouterProvider) Name() string {
    return "openrouter"
}

func (p *OpenRouterProvider) BaseURL() string {
    return "https://openrouter.ai/api/v1/chat/completions"
}

func (p *OpenRouterProvider) PrepareRequest(ctx context.Context, req *http.Request) error {
    // OpenRouter-specific headers
    req.Header.Set("Authorization", "Bearer "+p.APIKey)
    req.Header.Set("HTTP-Referer", "https://github.com/mandalnilabja/goatway")
    req.Header.Set("X-Title", "Goatway Proxy")
    return nil
}

func (p *OpenRouterProvider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
    // Streaming-safe proxy implementation
    // (Similar to current proxy.go logic)
    return nil
}
```

## Adding a New Provider (Example: OpenAI)

Just create a new file:

```go
// internal/provider/openai.go
package provider

import (
    "context"
    "net/http"
)

type OpenAIProvider struct {
    APIKey       string
    Organization string // Optional
}

func NewOpenAIProvider(apiKey, org string) *OpenAIProvider {
    return &OpenAIProvider{
        APIKey:       apiKey,
        Organization: org,
    }
}

func (p *OpenAIProvider) Name() string {
    return "openai"
}

func (p *OpenAIProvider) BaseURL() string {
    return "https://api.openai.com/v1/chat/completions"
}

func (p *OpenAIProvider) PrepareRequest(ctx context.Context, req *http.Request) error {
    // OpenAI-specific headers
    req.Header.Set("Authorization", "Bearer "+p.APIKey)
    if p.Organization != "" {
        req.Header.Set("OpenAI-Organization", p.Organization)
    }
    return nil
}

func (p *OpenAIProvider) ProxyRequest(ctx context.Context, w http.ResponseWriter, req *http.Request) error {
    // Streaming-safe proxy implementation
    return nil
}
```

Then in `cmd/api/main.go`:
```go
var provider provider.Provider

switch os.Getenv("LLM_PROVIDER") {
case "openai":
    provider = provider.NewOpenAIProvider(
        os.Getenv("OPENAI_API_KEY"),
        os.Getenv("OPENAI_ORG"),
    )
case "openrouter":
    provider = provider.NewOpenRouterProvider(
        os.Getenv("OPENROUTER_API_KEY"),
    )
default:
    provider = provider.NewOpenRouterProvider(
        os.Getenv("OPENROUTER_API_KEY"),
    )
}
```

## Configuration Example

```bash
# .env file
LLM_PROVIDER=openrouter        # or "openai", "azure", etc.
OPENROUTER_API_KEY=sk-...
SERVER_ADDR=:8080
LOG_LEVEL=info
```

## AGENTS.md Compliance Checklist

✅ **Entry point logic in main command file only**: `cmd/api/main.go` only wires dependencies  
✅ **HTTP handlers as methods on repository struct**: All handlers in `internal/transport/http/handler/`  
✅ **Shared dependencies in repository struct**: Cache and Provider in Repo  
✅ **No global state**: Everything through dependency injection  
✅ **No cross-package imports from internal**: Clean boundaries  
✅ **Streaming correctness**: Provider interface maintains streaming semantics  
✅ **No buffering**: Each provider implements streaming directly  
✅ **Prefer standard library**: No new dependencies added  
✅ **No frameworks**: Pure Go standard library  

## Benefits Over Current Structure

| Benefit | Description |
|---------|-------------|
| **Provider Isolation** | Each provider in its own file with clear boundaries |
| **Easy Addition** | Add new provider = create one new file |
| **Testability** | Mock provider interface for testing |
| **Configuration** | Environment-based provider selection |
| **Maintainability** | Clear separation of concerns |
| **Scalability** | Support unlimited providers without code changes |
| **Zero Risk** | No impact on streaming or existing functionality |

## Migration Effort Estimate

| Phase | Effort | Risk |
|-------|--------|------|
| Create directory structure | 5 min | None |
| Move existing handlers | 10 min | Low |
| Create provider interface | 15 min | Low |
| Extract OpenRouter provider | 20 min | Low |
| Add configuration layer | 15 min | Low |
| Update main.go | 10 min | Low |
| Testing | 30 min | Low |
| **Total** | **~2 hours** | **Low** |

## Next Steps

1. **Review**: Examine the detailed docs:
   - [`PROJECT_STRUCTURE.md`](PROJECT_STRUCTURE.md) - Full design details
   - [`STRUCTURE_COMPARISON.md`](STRUCTURE_COMPARISON.md) - Before/after comparison

2. **Decide**: Approve the structure or request modifications

3. **Implement**: I can help implement the migration in phases:
   - Phase 1: Create directory structure
   - Phase 2: Move existing code
   - Phase 3: Create provider abstraction
   - Phase 4: Add configuration
   - Phase 5: Test everything

4. **Validate**: Ensure streaming still works perfectly

## Questions to Consider

Before proceeding, consider:

1. **Provider Priority**: Which providers do you want to support first?
   - OpenRouter (current) ✅
   - OpenAI
   - Azure OpenAI
   - Anthropic Claude
   - Google Gemini
   - Others?

2. **Configuration**: How do you want to configure providers?
   - Environment variables (recommended) ✅
   - Config file
   - Command-line flags

3. **Migration Timeline**: When do you want to start the migration?
   - Immediately
   - After review
   - Staged over time

## Conclusion

This structure provides:
- ✅ Clear provider separation (your main requirement)
- ✅ Standard Go project layout (your secondary requirement)
- ✅ AGENTS.md compliance (non-negotiable constraint)
- ✅ Easy extensibility (future-proof)
- ✅ Minimal risk (no breaking changes)

The migration can be done incrementally with testing at each step, ensuring zero disruption to existing functionality while setting up for easy multi-provider support.

---

**Ready to proceed?** Let me know if you'd like me to start implementing this structure, or if you have any questions or modifications to the design.