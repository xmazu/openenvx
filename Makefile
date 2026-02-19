BINARY_NAME=openenvx
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Default build for current platform
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Clean build artifacts
clean:
	rm -rf dist/
	rm -f $(BINARY_NAME)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Build for all platforms
build-all: build-macos build-linux

# macOS builds
build-macos:
	mkdir -p dist
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 .

# Linux builds
build-linux:
	mkdir -p dist
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 .

# Install locally
install:
	go install $(LDFLAGS) .

# Development run
dev:
	go run . init

# Format code
fmt:
	go fmt ./...

# Full CI pipeline
ci: fmt test build

.PHONY: build clean test test-coverage build-all build-macos build-linux install dev fmt ci