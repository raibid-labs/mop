.PHONY: proto build test lint clean install-tools docker-build run-server run-client

# Install required tools
install-tools:
	@echo "Installing protoc plugins..."
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate Go code from proto files
proto:
	@echo "Generating Go code from proto files..."
	@mkdir -p gen/go/auth/v1
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/auth/v1/auth.proto

# Build server and client binaries
build: proto
	@echo "Building server and client..."
	go build -o bin/server ./cmd/server
	go build -o bin/client ./cmd/client

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v -cover

# Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not installed, skipping..."; \
	fi

# Build Docker image
docker-build:
	@echo "Building Docker image..."
	docker build -t grpc-auth-service:latest .

# Run server locally
run-server: build
	@echo "Starting gRPC server on :9090..."
	./bin/server

# Run client locally
run-client: build
	@echo "Running gRPC client..."
	./bin/client

# Clean generated files and binaries
clean:
	@echo "Cleaning generated files and binaries..."
	rm -rf bin/ gen/

# Development: watch and rebuild on changes
dev: build
	@echo "Running in development mode..."
	./bin/server
