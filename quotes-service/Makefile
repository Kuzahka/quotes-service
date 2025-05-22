.PHONY: build run test test-coverage lint fmt clean docker-up docker-down docker-build docker-logs help

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt

# Binary name
BINARY_NAME=quotes-service
BINARY_PATH=bin/$(BINARY_NAME)

# Build the application
build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin/
	$(GOBUILD) -o $(BINARY_PATH) cmd/server/main.go
	@echo "Build completed: $(BINARY_PATH)"

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	$(GOCMD) run cmd/server/main.go

# Test the application
test:
	@echo "Running tests..."
	$(GOTEST) ./tests/unit/...

# Test with coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run integration tests
test-integration:
	@echo "Running integration tests..."
	$(GOTEST) -tags=integration -v ./tests/integration/...

# Lint the code
lint:
	@echo "Running linters..."
	@command -v golangci-lint >/dev/null 2>&1 || { echo "golangci-lint not installed. Installing..."; go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; }
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .

# Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.out coverage.html

# Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

# Security audit
audit:
	@echo "Running security audit..."
	@command -v govulncheck >/dev/null 2>&1 || { echo "govulncheck not installed. Installing..."; go install golang.org/x/vuln/cmd/govulncheck@latest; }
	govulncheck ./...

# Docker commands
docker-build:
	@echo "Building Docker image..."
	docker-compose build

docker-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up -d

docker-down:
	@echo "Stopping services..."
	docker-compose down

docker-logs:
	@echo "Showing logs..."
	docker-compose logs -f

docker-restart:
	@echo "Restarting services..."
	docker-compose restart

# Database commands
db-migrate:
	@echo "Running database migrations..."
	docker exec -it quotes-service_postgres_1 psql -U quotes_user -d quotes_db -f /docker-entrypoint-initdb.d/001_create_quotes_table.sql

db-shell:
	@echo "Connecting to database..."
	docker exec -it quotes-service_postgres_1 psql -U quotes_user -d quotes_db

# Development helpers
dev-setup: deps fmt lint test
	@echo "Development setup completed!"

install-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

# Help
help:
	@echo "Available commands:"
	@echo "  build           - Build the application"
	@echo "  run             - Run the application"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  test-integration- Run integration tests"
	@echo "  lint            - Run linters"
	@echo "  fmt             - Format code"
	@echo "  clean           - Clean build artifacts"
	@echo "  deps            - Download dependencies"
	@echo "  audit           - Run security audit"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-up       - Start services"
	@echo "  docker-down     - Stop services"
	@echo "  docker-logs     - Show logs"
	@echo "  db-migrate      - Run database migrations"
	@echo "  db-shell        - Connect to database"
	@echo "  dev-setup       - Setup development environment"
	@echo "  install-tools   - Install development tools"
	@echo "  help            - Show this help"

# Default target
.DEFAULT_GOAL := help