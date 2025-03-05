.PHONY: build test lint help clean

BINARY_NAME=anilist-mal-sync
LINT_VERSION=v1.64.6
DOCKER_LINT_CMD=docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:$(LINT_VERSION)

.DEFAULT_GOAL := help

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/main.go

# Run tests
test:
	go test ./... -v

# Run linter using Docker
lint:
	@echo "Running golangci-lint $(LINT_VERSION) in Docker..."
	$(DOCKER_LINT_CMD) golangci-lint run --new

# Clean build artifacts, temporary files and test cache
clean:
	@echo "Cleaning build artifacts and temporary files..."
	@rm -f $(BINARY_NAME)
	@go clean -testcache
	@echo "Cleanup complete!"

# Show help
help:
	@echo "Available commands:"
	@echo "  build     - Build the application"
	@echo "  test      - Run tests"
	@echo "  lint      - Run linter using Docker (golangci-lint $(LINT_VERSION))"
	@echo "  clean     - Remove build artifacts, temporary files and clean test cache"
	@echo "  help      - Show this help message"
