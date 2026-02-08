# Variables
BINARY_NAME=goatway
MAIN_FILE=./cmd/api/main.go
TOOLS_DIR=./bin/tools
DIST_DIR=./dist

# Version info (override with: make build VERSION=1.0.0)
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u '+%Y-%m-%dT%H:%M:%SZ')

# Linker flags to inject version info
LDFLAGS=-ldflags "-s -w -X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

GOIMPORTS=$(TOOLS_DIR)/goimports
GOLANGCI_LINT=$(TOOLS_DIR)/golangci-lint

# Default command
all: test build

# -------- TOOLS --------

tools: $(GOIMPORTS) $(GOLANGCI_LINT)

$(GOIMPORTS):
	@echo "Installing goimports..."
	mkdir -p $(TOOLS_DIR)
	GOBIN=$(abspath $(TOOLS_DIR)) go install golang.org/x/tools/cmd/goimports@v0.37.0

$(GOLANGCI_LINT):
	@echo "Installing golangci-lint..."
	mkdir -p $(TOOLS_DIR)
	GOBIN=$(abspath $(TOOLS_DIR)) go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5

# -------- BUILD --------

build:
	@echo "Building Goatway $(VERSION)..."
	mkdir -p bin
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) $(MAIN_FILE)

run:
	@echo "Running Goatway..."
	go run $(MAIN_FILE)

# Install locally via go install
install:
	@echo "Installing Goatway $(VERSION)..."
	go install $(LDFLAGS) $(MAIN_FILE)

# Build for all platforms
build-all: clean-dist
	@echo "Building Goatway $(VERSION) for all platforms..."
	mkdir -p $(DIST_DIR)
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_FILE)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_FILE)
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_FILE)
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_FILE)
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_FILE)
	@echo "Binaries built in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

# Create release archives
release: build-all
	@echo "Creating release archives..."
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-darwin-amd64.tar.gz $(BINARY_NAME)-darwin-amd64
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-darwin-arm64.tar.gz $(BINARY_NAME)-darwin-arm64
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-linux-amd64.tar.gz $(BINARY_NAME)-linux-amd64
	cd $(DIST_DIR) && tar -czf $(BINARY_NAME)-linux-arm64.tar.gz $(BINARY_NAME)-linux-arm64
	cd $(DIST_DIR) && zip -q $(BINARY_NAME)-windows-amd64.zip $(BINARY_NAME)-windows-amd64.exe
	@echo "Release archives created:"
	@ls -la $(DIST_DIR)/*.tar.gz $(DIST_DIR)/*.zip

test:
	@echo "Running tests..."
	go test -v ./...

# -------- QUALITY --------

fmt: tools
	@echo "Formatting code..."
	$(GOIMPORTS) -w .

fmt-check: tools
	@echo "Checking formatting..."
	test -z "$$($(GOIMPORTS) -l .)"

lint: tools
	@echo "Running linters..."
	$(GOLANGCI_LINT) run

# -------- CLEAN --------

clean:
	@echo "Cleaning up..."
	go clean
	rm -rf bin dist

clean-dist:
	@rm -rf $(DIST_DIR)

.PHONY: all build run test fmt fmt-check lint clean clean-dist tools install build-all release
