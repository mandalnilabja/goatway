# Suggested Commands

## Build & Run
```bash
make build        # Build binary to bin/goatway
make run          # Run the server (go run)
```

## Testing & Quality
```bash
make test         # Run all tests
make fmt          # Format code with goimports
make fmt-check    # Check formatting without modifying
make lint         # Run golangci-lint
```

## Development
```bash
make tools        # Install dev tools (goimports, golangci-lint)
make clean        # Remove build artifacts
```

## Environment Variables
- `SERVER_PORT` - Server address (default: `:8080`)
- `LLM_PROVIDER` - Provider: `openrouter` (default)
- `OPENROUTER_API_KEY` - OpenRouter API key
