.PHONY: help build test test-unit test-integration test-coverage run clean docker-build docker-up docker-down lint

# Variables
APP_NAME := redis-cache
DOCKER_IMAGE := raibid-labs/redis-cache-example
DOCKER_TAG := latest

help: ## Show this help
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

build: ## Build the application
	@echo "Building $(APP_NAME)..."
	go build -o $(APP_NAME) ./cmd/server

run: ## Run the application locally
	@echo "Starting $(APP_NAME)..."
	@echo "Make sure Redis is running on localhost:6379"
	go run ./cmd/server/main.go

test: test-unit test-integration ## Run all tests

test-unit: ## Run unit tests
	@echo "Running unit tests..."
	go test -v -race -timeout 30s ./internal/...

test-integration: ## Run integration tests (requires Redis on localhost:6379)
	@echo "Running integration tests..."
	@echo "Make sure Redis is running on localhost:6379"
	go test -v -race -timeout 60s ./tests/...

test-coverage: ## Run tests with coverage report
	@echo "Running tests with coverage..."
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out
	@echo ""
	@echo "Total coverage:"
	@go tool cover -func=coverage.out | grep total | awk '{print $$3}'
	@echo ""
	@echo "Generate HTML report with: make coverage-html"

coverage-html: test-coverage ## Generate HTML coverage report
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

lint: ## Run linters
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -f $(APP_NAME)
	rm -f coverage.out coverage.html
	go clean -cache -testcache

docker-build: ## Build Docker image
	@echo "Building Docker image..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-up: ## Start services with docker-compose
	@echo "Starting services..."
	docker-compose up -d
	@echo "Waiting for services to be healthy..."
	@sleep 5
	@docker-compose ps

docker-down: ## Stop services with docker-compose
	@echo "Stopping services..."
	docker-compose down

docker-logs: ## Show docker-compose logs
	docker-compose logs -f

deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod verify

tidy: ## Tidy go.mod
	@echo "Tidying go.mod..."
	go mod tidy

fmt: ## Format code
	@echo "Formatting code..."
	go fmt ./...

vet: ## Run go vet
	@echo "Running go vet..."
	go vet ./...

check: fmt vet lint test ## Run all checks (format, vet, lint, test)

.DEFAULT_GOAL := help
