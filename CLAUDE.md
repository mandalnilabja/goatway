# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Goatway is a lightweight, streaming-safe, OpenAI-compatible HTTP proxy for routing requests to different LLM providers. Written in Go with minimal dependencies.

## Build and Development Commands

```bash
make build        # Build binary to bin/goatway
make run          # Run the server (go run)
make test         # Run all tests
make fmt          # Format code with goimports
make fmt-check    # Check formatting without modifying
make lint         # Run golangci-lint
make tools        # Install dev tools (goimports, golangci-lint)
make clean        # Remove build artifacts
```

## Environment Variables

- `SERVER_ADDR` - Server address (default: `:8080`)
- `LLM_PROVIDER` - Provider to use: `openrouter` (default), future: `openai`, `azure`
- `OPENROUTER_API_KEY` - OpenRouter API key

## Architecture

```
cmd/api/main.go          # Entry point: wires config → provider → handlers → router → server
internal/
  config/                # Environment-based configuration
  provider/
    provider.go          # Provider interface
    openrouter/          # OpenRouter implementation with streaming proxy
  app/
    router.go            # Route registration (http.ServeMux)
    server.go            # HTTP server wrapper
  storage/
    storage.go           # Storage interface
    sqlite/              # SQLite implementation
    models/              # Data models
    encryption/          # AES encryption for API keys
  tokenizer/             # Token counting for requests
  transport/http/
    handler/
      handler.go         # Handler repository (composes domain handlers)
      admin/             # Admin API handlers (credentials, apikeys, usage)
      proxy/             # OpenAI-compatible proxy endpoints
      webui/             # Web UI handlers
      infra/             # Health and cache handlers
      shared/            # Shared utilities
    middleware/          # CORS, RequestID, Auth, Logging
  types/                 # OpenAI-compatible type definitions
```

**Request Flow**: `Client → Router → Handler (Repo) → Provider.ProxyRequest → Upstream LLM → Streaming Response`

## Critical Streaming Rules

Any code touching the proxy path MUST follow these rules:

1. `http.Transport.DisableCompression` MUST be `true` (prevents gzip breaking SSE)
2. `text/event-stream` responses MUST call `Flusher.Flush()` after each write
3. Never buffer full responses or accumulate SSE chunks
4. Client context MUST propagate end-to-end
5. No retries or background goroutines in request path

## Code Conventions

- Handlers are methods on `handler.Repo` struct
- Shared dependencies injected via `Repo` (cache, provider)
- Provider interface in `provider/provider.go` - add new providers by implementing it
- Use Go standard library; avoid frameworks
- Conventional commits: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

## Code Guidelines

When writing or modifying code in this repository, follow these practices:

### File Size & Organization
- Keep all code files under 200 lines. If a file exceeds this, split it into smaller modules.
- One file should have one clear responsibility. Avoid mixing unrelated functionality.

### Functions
- Write small, focused functions that do one thing well.
- Extract reusable logic into separate utility functions.
- Prefer multiple small functions over one large function with many responsibilities.

### Modularity
- Group related functions into dedicated modules/files.
- Use clear imports rather than cramming everything into one file.
- When adding new features, create new files rather than bloating existing ones.

### Splitting Guidelines
- Router files: Split by resource/domain if they grow large.
- Service files: One service class/module per domain entity.
- Components: Extract sub-components when a component does multiple distinct things.
- Utility functions: Group by purpose


## Key Constraints (from AGENTS.md)

- Streaming correctness > features
- Low latency > abstraction
- No global state outside initialization
- No cross-package imports from internal packages
- Filter hop-by-hop headers; forward all others verbatim
- Table-driven tests preferred; mock external HTTP calls
