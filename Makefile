.PHONY: build test lint fmt check clean install dry-run generate

# Load .env file if exists
-include .env

BINARY_NAME=anilist-mal-sync
LINT_VERSION=v2.10.1
DOCKER_LINT_CMD=docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:$(LINT_VERSION)
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

.DEFAULT_GOAL := help

# Install all development tools
install:
	@echo "🔧 Installing development tools..."
	@echo ""
	@echo "Installing golangci-lint $(LINT_VERSION)..."
	@which golangci-lint > /dev/null 2>&1 || \
		brew install golangci/tap/golangci-lint
	@echo "✓ golangci-lint installed"
	@echo ""
	@echo "Installing gofumpt..."
	@which gofumpt > /dev/null 2>&1 || \
		go install mvdan.cc/gofumpt@latest
	@echo "✓ gofumpt installed"
	@echo ""
	@echo "Installing goimports..."
	@which goimports > /dev/null 2>&1 || \
		go install golang.org/x/tools/cmd/goimports@latest
	@echo "✓ goimports installed"
	@echo ""
	@echo "Installing gci (import organizer)..."
	@which gci > /dev/null 2>&1 || \
		go install github.com/daixiang0/gci@latest
	@echo "✓ gci installed"
	@echo ""
	@echo "Installing govulncheck (vulnerability scanner)..."
	@which govulncheck > /dev/null 2>&1 || \
		go install golang.org/x/vuln/cmd/govulncheck@latest
	@echo "✓ govulncheck installed"
	@echo ""
	@echo "✅ All development tools installed successfully!"
	@echo ""
	@echo "Available commands:"
	@echo "  make build    - Build the application"
	@echo "  make test     - Run tests"
	@echo "  make generate - Generate mocks"
	@echo "  make fmt      - Format code"
	@echo "  make lint     - Run linter"
	@echo "  make clean    - Clean build artifacts"

# Build the application
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Run sync in dry-run mode (reads ANILIST_MAL_SYNC_CONFIG from .env file)
dry-run:
	@if [ -n "$(ANILIST_MAL_SYNC_CONFIG)" ]; then \
		go run . -c "$(ANILIST_MAL_SYNC_CONFIG)" sync -d --verbose --all; \
	else \
		go run . sync -d --verbose --all; \
	fi

# Run tests
test:
	go test ./... -v -race

# Generate mocks using mockgen
generate:
	@echo "🔧 Generating mocks..."
	@go generate ./...
	@echo "✓ Mocks generated"

# Format code with gofumpt
fmt:
	@echo "Formatting code with gofumpt..."
	@gofumpt -l -w .
	@echo "Formatting complete!"

# Run linter using Docker
lint:
	@echo "Running golangci-lint $(LINT_VERSION) in Docker..."
	# $(DOCKER_LINT_CMD) golangci-lint run --new
	golangci-lint run --new

# Run all checks (same as Git hooks: format + imports + lint + vet + test)
check: generate
	@echo "🔍 Running all checks..."
	@echo ""
	@echo "1️⃣  Formatting code with gofumpt..."
	@gofumpt -l -w .
	@echo "✓ Format complete"
	@echo ""
	@echo "2️⃣  Organizing imports with goimports..."
	@goimports -w .
	@echo "✓ Imports organized"
	@echo ""
	@echo "3️⃣  Running go vet..."
	@go vet ./...
	@echo "✓ Vet complete"
	@echo ""
	@echo "4️⃣  Running golangci-lint..."
	@golangci-lint run --timeout=5m
	@echo "✓ Lint complete"
	@echo ""
	@echo "5️⃣  Running tests..."
	@go test ./... -v
	@echo "✓ Tests complete"
	@echo ""
	@echo "✅ All checks passed!"

# Clean build artifacts, temporary files and test cache
clean:
	@echo "Cleaning build artifacts and temporary files..."
	@rm -f $(BINARY_NAME)
	@go clean -testcache
	@echo "Cleanup complete!"

# Show help
help:
	@echo "Available commands:"
	@echo "  install          - Install all development tools (brew + go install)"
	@echo "  build            - Build the application"
	@echo "  dry-run          - Run sync in dry-run mode (reads ANILIST_MAL_SYNC_CONFIG from .env)"
	@echo "  test             - Run tests"
	@echo "  generate         - Generate mocks using mockgen"
	@echo "  fmt              - Format code with gofumpt"
	@echo "  lint             - Run linter (golangci-lint $(LINT_VERSION))"
	@echo "  check            - Run all checks (generate + format + imports + lint + vet + test)"
	@echo "  clean            - Remove build artifacts, temporary files and clean test cache"
	@echo "  help             - Show this help message"
