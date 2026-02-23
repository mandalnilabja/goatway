# Goatway Project Overview

## Purpose
Goatway is a lightweight, streaming-safe, OpenAI-compatible HTTP proxy for routing requests to different LLM providers. Written in Go with minimal dependencies.

## Tech Stack
- **Language**: Go
- **Database**: SQLite (with encryption for API keys)
- **Web**: Standard library http.ServeMux

## Architecture
```
cmd/api/main.go          # Entry point
internal/
  config/                # Environment-based configuration
  provider/              # Provider interface + implementations (OpenRouter)
  app/                   # Router and server
  storage/               # Storage interface + SQLite implementation
  tokenizer/             # Token counting
  transport/http/
    handler/             # HTTP handlers (admin, proxy, webui, infra)
    middleware/          # CORS, RequestID, Auth, Logging
  types/                 # OpenAI-compatible type definitions
```

## Request Flow
`Client → Router → Handler (Repo) → Provider.ProxyRequest → Upstream LLM → Streaming Response`

## Critical Constraints
- Streaming correctness > features
- Low latency > abstraction
- No global state outside initialization
- Filter hop-by-hop headers; forward all others verbatim
