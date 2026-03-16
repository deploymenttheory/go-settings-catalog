.PHONY: all build test clean install lint fmt tidy update-catalog help

all: build

# Build main CLI
build:
	@echo "Building mobileconfig-to-terraform CLI..."
	@mkdir -p bin
	@go build -o bin/mobileconfig-to-terraform ./cmd/mobileconfig-to-terraform
	@echo "✓ Binary created: bin/mobileconfig-to-terraform"

# Run tests
test:
	@echo "Running tests..."
	@go test -v ./catalog ./converter

# Run tests with coverage
test-coverage:
	@echo "Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./catalog ./converter
	@go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Update catalog data (maintenance)
update-catalog:
	@echo "Updating catalog data..."
	@go run ./scripts/fetch-catalog --tenant=$(TENANT_ID) --client=$(CLIENT_ID) --secret=$(CLIENT_SECRET) --platform=macOS
	@echo "✓ Catalog updated. Rebuild binary with 'make build'"

# Install library
install:
	@echo "Installing library..."
	@go install ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -rf bin/
	@rm -f coverage.out coverage.html
	@go clean

# Run linter
lint:
	@echo "Running linter..."
	@golangci-lint run

# Format code
fmt:
	@echo "Formatting code..."
	@go fmt ./...

# Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	@go mod tidy

# Run all checks (fmt, lint, test)
check: fmt lint test
	@echo "All checks passed!"

# Show help
help:
	@echo "Available targets:"
	@echo "  all            - Run tests and build (default)"
	@echo "  build          - Build all packages"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  examples       - Build example binaries"
	@echo "  install        - Install library"
	@echo "  clean          - Clean build artifacts"
	@echo "  lint           - Run linter"
	@echo "  fmt            - Format code"
	@echo "  tidy           - Tidy dependencies"
	@echo "  check          - Run fmt, lint, and test"
	@echo "  help           - Show this help message"
