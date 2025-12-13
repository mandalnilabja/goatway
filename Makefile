# Variables
BINARY_NAME=goatway
MAIN_FILE=./cmd/api/main.go

# Default command (runs when you type just 'make')
all: test build

# 1. Build the binary
build:
	@echo "Building Goatway..."
	# Creates a 'bin' folder and puts the executable there
	mkdir -p bin
	go build -o bin/$(BINARY_NAME) $(MAIN_FILE)

# 2. Run the application
run:
	@echo "Running Goatway..."
	go run $(MAIN_FILE)

# 3. Test the application
test:
	@echo "Running tests..."
	# ./... tells Go to test the current directory and all subdirectories
	go test -v ./...

# 4. Clean up binaries
clean:
	@echo "Cleaning up..."
	go clean
	rm -rf bin

# Marks these commands as not associated with actual files
.PHONY: all build run test clean