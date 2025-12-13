# Variables
BINARY_NAME=goatway
MAIN_FILE=./cmd/api/main.go
TOOLS_DIR=./bin/tools

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
	@echo "Building Goatway..."
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) $(MAIN_FILE)

run:
	@echo "Running Goatway..."
	go run $(MAIN_FILE)

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
	rm -rf bin

.PHONY: all build run test fmt fmt-check lint clean tools
