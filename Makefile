# Makefile for RSS Reader

APP_NAME := rssreader
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
LDFLAGS := -ldflags "-s -w -X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME)"

# Output directory
DIST_DIR := dist

# Go parameters
GOCMD := go
GOBUILD := $(GOCMD) build
GOTEST := $(GOCMD) test
GOMOD := $(GOCMD) mod

.PHONY: all build clean test deps help
.PHONY: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows
.PHONY: build-all docker ui

# Default target
all: build

# Build for current platform
build:
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME) ./cmd/server

# Build for Linux AMD64
build-linux:
	@echo "Building for Linux AMD64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 ./cmd/server

# Build for Linux ARM64
build-linux-arm64:
	@echo "Building for Linux ARM64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-linux-arm64 ./cmd/server

# Build for macOS AMD64 (Intel)
build-darwin:
	@echo "Building for macOS AMD64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-amd64 ./cmd/server

# Build for macOS ARM64 (Apple Silicon)
build-darwin-arm64:
	@echo "Building for macOS ARM64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-darwin-arm64 ./cmd/server

# Build for Windows AMD64
build-windows:
	@echo "Building for Windows AMD64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe ./cmd/server

# Build for Windows ARM64
build-windows-arm64:
	@echo "Building for Windows ARM64..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME)-windows-arm64.exe ./cmd/server

# Build for all platforms
build-all: build-linux build-linux-arm64 build-darwin build-darwin-arm64 build-windows build-windows-arm64
	@echo "All builds completed!"
	@ls -la $(DIST_DIR)/

# Build with CGO (for SQLite FTS5 support)
build-cgo:
	@echo "Building with CGO enabled..."
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 $(GOBUILD) $(LDFLAGS) -o $(DIST_DIR)/$(APP_NAME) ./cmd/server

# Build Docker image
docker:
	docker build -t $(APP_NAME):$(VERSION) .
	docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest

# Build UI
ui:
	@echo "Building UI..."
	cd ui && npm install && npm run build

# Run tests
test:
	$(GOTEST) -v ./...

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf $(DIST_DIR)
	rm -f $(APP_NAME)

# Run locally
run:
	$(GOBUILD) -o $(APP_NAME) ./cmd/server && ./$(APP_NAME)

# Run with hot reload (requires air: go install github.com/cosmtrek/air@latest)
dev:
	air

# Help
help:
	@echo "RSS Reader Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make build              Build for current platform"
	@echo "  make build-cgo          Build with CGO (SQLite FTS5)"
	@echo "  make build-linux        Build for Linux AMD64"
	@echo "  make build-linux-arm64  Build for Linux ARM64"
	@echo "  make build-darwin       Build for macOS Intel"
	@echo "  make build-darwin-arm64 Build for macOS Apple Silicon"
	@echo "  make build-windows      Build for Windows AMD64"
	@echo "  make build-windows-arm64 Build for Windows ARM64"
	@echo "  make build-all          Build for all platforms"
	@echo "  make docker             Build Docker image"
	@echo "  make ui                 Build frontend"
	@echo "  make test               Run tests"
	@echo "  make deps               Download dependencies"
	@echo "  make clean              Clean build artifacts"
	@echo "  make run                Build and run locally"
	@echo "  make dev                Run with hot reload"
	@echo "  make help               Show this help"
