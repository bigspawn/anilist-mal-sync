.PHONY: build test lint help clean

BINARY_NAME=anilist-mal-sync
LINT_VERSION=1.64.6
LINT_BINARY=bin/golangci-lint

.DEFAULT_GOAL := help

# Build the application
build:
	go build -o $(BINARY_NAME) ./cmd/main.go

# Run tests
test:
	go test ./... -v

# Install and run linter
lint:
	@mkdir -p bin
	@[ -f $(LINT_BINARY) ] curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s v$(LINT_VERSION)
	$(LINT_BINARY) run

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
	@echo "  lint      - Run linter (installs golangci-lint $(LINT_VERSION) if needed)"
	@echo "  clean     - Remove build artifacts, temporary files and clean test cache"
	@echo "  help      - Show this help message"
