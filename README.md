# Goatway

Personal lightweight AI gateway acting as OpenAI-compatible endpoint router for different providers. Written in Go.

Why Goatway? Because Go is the GOAT, also I needed it.

## Installation

### From Release Binaries

Download the latest release from the [GitHub Releases](https://github.com/mandalnilabja/goatway/releases) page.

```bash
# Linux (amd64)
curl -L https://github.com/mandalnilabja/goatway/releases/latest/download/goatway_linux_amd64.tar.gz | tar xz
sudo mv goatway /usr/local/bin/

# macOS (Apple Silicon)
curl -L https://github.com/mandalnilabja/goatway/releases/latest/download/goatway_darwin_arm64.tar.gz | tar xz
sudo mv goatway /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/mandalnilabja/goatway/releases/latest/download/goatway_darwin_amd64.tar.gz | tar xz
sudo mv goatway /usr/local/bin/
```

### From Source

#### Prerequisites

- Go **1.23** or later
- `make` (optional but recommended)

#### Clone and Build

```bash
git clone https://github.com/mandalnilabja/goatway.git
cd goatway
make build
```

The binary will be created at `bin/goatway`.

### Install via Go

```bash
go install github.com/mandalnilabja/goatway/cmd/api@latest
```

## Quick Start

```bash
# Run the server
make run

# Or run the binary directly
./bin/goatway
```

On first run, you'll be prompted to set an admin password.

## Configuration

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `SERVER_PORT` | Server bind address | `:8080` |
| `ENABLE_WEB_UI` | Enable web dashboard | `true` |

## API Endpoints

### OpenAI-Compatible Proxy

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/chat/completions` | Chat completions (streaming supported) |
| GET | `/v1/models` | List available models |
| GET | `/v1/models/{model}` | Get model details |

### Admin API

All admin endpoints require Bearer token authentication with the admin password.

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/credentials` | Add provider credentials |
| GET | `/api/admin/credentials` | List credentials |
| POST | `/api/admin/apikeys` | Create client API key |
| GET | `/api/admin/apikeys` | List API keys |
| GET | `/api/admin/usage` | Get usage statistics |
| GET | `/api/admin/logs` | Get request logs |

### Web UI

Access the web dashboard at `http://localhost:8080/web` (requires login with admin password).

## Development

```bash
make build        # Build binary to bin/goatway
make run          # Run the server
make test         # Run all tests
make fmt          # Format code with goimports
make lint         # Run golangci-lint
make tools        # Install dev tools
make clean        # Remove build artifacts
```

### Release

```bash
# 1. Stage and commit your changes
git add -A
git commit -m "your commit message"

# 2. Push commits to main
git push origin main

# 3. Create and push the release tag
make tag V=0.1.3
```

The `make tag` command validates semver format, creates an annotated tag, and pushes it to trigger the GitHub Actions release workflow.

```bash
# Test release locally (creates binaries without publishing)
make release-snapshot
```

## License

MIT License - see [LICENSE](LICENSE) for details.
