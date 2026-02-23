# Code Style & Conventions

## File Size
- Target: under 120 lines
- Maximum: 150 lines (never exceed)
- Split into smaller modules if approaching limits

## Organization
- One file, one clear responsibility
- Group related functions into dedicated modules
- Prefer editing existing files over creating new ones

## Go Conventions
- Handlers are methods on `handler.Repo` struct
- Shared dependencies injected via `Repo` (cache, provider)
- Use Go standard library; avoid frameworks
- Table-driven tests preferred; mock external HTTP calls

## Commits
- Conventional commits: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

## Streaming Rules (CRITICAL)
1. `http.Transport.DisableCompression` MUST be `true`
2. `text/event-stream` responses MUST call `Flusher.Flush()` after each write
3. Never buffer full responses or accumulate SSE chunks
4. Client context MUST propagate end-to-end
5. No retries or background goroutines in request path
