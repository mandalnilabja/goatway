# Implementation Plan: Project Structure Refactor

This plan outlines the steps to refactor the `goatway` project to a modular, provider-based architecture as described in `docs/STRUCTURE_RECOMMENDATION.md`.

## Phase 1: Directory Structure Setup
- [ ] Create `internal/provider`
- [ ] Create `internal/config`
- [ ] Create `internal/transport/http/handler`
- [ ] Create `internal/transport/http/middleware`
- [ ] Create `internal/app`
- [ ] Create `internal/domain`
- [ ] Create `pkg/logger`

## Phase 2: Provider Implementation
- [ ] Create `internal/provider/provider.go`: Define the `Provider` interface.
- [ ] Create `internal/provider/openrouter.go`: Implement `OpenRouterProvider` by extracting logic from `internal/handlers/proxy.go`.
    - Ensure `PrepareRequest` handles headers correctly.
    - Ensure `ProxyRequest` maintains streaming correctness.

## Phase 3: Configuration
- [ ] Create `internal/config/config.go`: Implement environment variable loading.
    - Support `SERVER_ADDR`, `LLM_PROVIDER`, `OPENROUTER_API_KEY`, etc.

## Phase 4: Handler Refactoring
- [ ] Create `internal/transport/http/handler/repo.go`: Define `Repo` struct with `Cache` and `Provider`.
- [ ] Create `internal/transport/http/handler/health.go`: Move `HealthCheck` and `Home` handlers.
- [ ] Create `internal/transport/http/handler/cache.go`: Move `GetCachedData` handler.
- [ ] Create `internal/transport/http/handler/proxy.go`: Refactor `OpenAIProxy` to use `h.Provider.ProxyRequest`.

## Phase 5: Application Wiring
- [ ] Create `internal/app/router.go`: Setup `http.ServeMux` with the new handlers.
- [ ] Create `internal/app/server.go`: Encapsulate server configuration and startup.
- [ ] Update `cmd/api/main.go`:
    - Load config.
    - Initialize cache.
    - Initialize provider based on config.
    - Initialize repo.
    - Initialize router and server.
    - Start server.

## Phase 6: Cleanup and Verification
- [ ] Remove `internal/handlers` directory.
- [ ] Run `go mod tidy`.
- [ ] Run `go build ./...` to ensure no compilation errors.
- [ ] Verify `AGENTS.md` compliance.
