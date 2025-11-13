.PHONY: help build test run docker-build docker-up docker-down migrate-up migrate-down clean

help:
	@echo "Available commands:"
	@echo "  make build        - Build the application"
	@echo "  make test         - Run tests"
	@echo "  make run          - Run the application locally"
	@echo "  make docker-build - Build Docker image"
	@echo "  make docker-up    - Start services with Docker Compose"
	@echo "  make docker-down  - Stop services with Docker Compose"
	@echo "  make migrate-up   - Run database migrations up"
	@echo "  make migrate-down - Run database migrations down"
	@echo "  make clean        - Clean build artifacts"

build:
	go build -o bin/server ./cmd/server

test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

run:
	go run ./cmd/server

docker-build:
	docker build -t sql-app:latest .

docker-up:
	docker-compose up -d
	@echo "Waiting for database to be ready..."
	@sleep 5
	@echo "Running migrations..."
	@make migrate-up
	@echo "Services started successfully!"

docker-down:
	docker-compose down

migrate-up:
	docker-compose exec postgres psql -U postgres -d orders -c "\i /docker-entrypoint-initdb.d/001_create_customers.up.sql" || true
	docker-compose exec -T postgres psql -U postgres -d orders < migrations/001_create_customers.up.sql || true
	docker-compose exec -T postgres psql -U postgres -d orders < migrations/002_create_orders.up.sql || true

migrate-down:
	docker-compose exec -T postgres psql -U postgres -d orders < migrations/002_create_orders.down.sql || true
	docker-compose exec -T postgres psql -U postgres -d orders < migrations/001_create_customers.down.sql || true

clean:
	rm -rf bin/
	rm -f coverage.out coverage.html
	docker-compose down -v
