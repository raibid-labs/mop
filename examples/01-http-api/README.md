# HTTP REST API Example - Product Catalog

A production-ready HTTP REST API service demonstrating **zero-code observability** with OBI eBPF automatic instrumentation. This example showcases a product catalog service with CRUD operations, built using Go and the Gin framework.

## Overview

This example demonstrates:
- **Zero-Code Instrumentation**: No SDK or library changes required for complete observability
- **Automatic Tracing**: HTTP requests, responses, and latencies captured automatically by OBI
- **Production Patterns**: Rate limiting, structured logging, error handling, health checks
- **Performance**: High-throughput service with minimal overhead (<1% CPU from OBI)
- **Cloud-Native**: Kubernetes-ready with proper health probes and resource management

## Features

- **CRUD Operations**: Create, Read, Update, Delete products
- **Search**: Full-text search across product names and descriptions
- **Pagination**: Efficient list operations with limit/offset
- **Rate Limiting**: Per-IP rate limiting to prevent abuse
- **Health Checks**: Liveness and readiness probes for Kubernetes
- **Testing Endpoints**: Slow endpoint (1-3s latency) and error simulation for testing
- **Middleware**: Logging, recovery, CORS, timeouts, request IDs
- **OBI Ready**: Annotations and labels for automatic instrumentation

## Quick Start

### Prerequisites

- Go 1.21+
- Docker
- kubectl (for Kubernetes deployment)
- OBI agent deployed in cluster

### Local Development

```bash
# Clone the repository
git clone <repository-url>
cd mop/examples/01-http-api

# Install dependencies
go mod download

# Build the application
make build

# Run locally
make run

# The server will start on http://localhost:8080
```

### Run Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run benchmarks
go test -bench=. ./tests/
```

### Build Docker Image

```bash
# Build the Docker image
make docker-build

# Run the container
make docker-run

# The API will be available at http://localhost:8080
```

## API Reference

### Endpoints

#### Health Check
```bash
GET /health

Response: 200 OK
{
  "status": "healthy",
  "uptime": "2h15m30s",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

#### List Products
```bash
GET /products?limit=10&offset=0

Response: 200 OK
{
  "products": [...],
  "total": 100,
  "limit": 10,
  "offset": 0
}
```

#### Get Product
```bash
GET /products/:id

Response: 200 OK
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "Product Name",
  "description": "Product Description",
  "price": 99.99,
  "stock": 100,
  "created_at": "2024-01-15T10:00:00Z",
  "updated_at": "2024-01-15T10:00:00Z"
}
```

#### Create Product
```bash
POST /products
Content-Type: application/json

{
  "name": "New Product",
  "description": "Description",
  "price": 99.99,
  "stock": 100
}

Response: 201 Created
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "New Product",
  ...
}
```

#### Update Product
```bash
PUT /products/:id
Content-Type: application/json

{
  "name": "Updated Product",
  "description": "Updated Description",
  "price": 149.99,
  "stock": 50
}

Response: 200 OK
```

#### Delete Product
```bash
DELETE /products/:id

Response: 200 OK
{
  "message": "Product deleted successfully"
}
```

#### Search Products
```bash
GET /search?q=query&limit=10&offset=0

Response: 200 OK
{
  "products": [...],
  "total": 5,
  "limit": 10,
  "offset": 0
}
```

#### Testing Endpoints

```bash
# Slow endpoint (1-3 second delay)
GET /slow

# Error endpoint (always returns 500)
GET /error
```

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────┐
│                   Client / Load Balancer                │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                  Gin HTTP Router                        │
│  ┌────────────────────────────────────────────────┐   │
│  │  Middleware Chain (Order Matters)              │   │
│  │  1. Request ID → Generate unique ID            │   │
│  │  2. Logger → Structured logging                │   │
│  │  3. Recovery → Panic recovery                  │   │
│  │  4. CORS → Cross-origin headers                │   │
│  │  5. Timeout → Request timeout (30s)            │   │
│  │  6. Rate Limit → Per-IP limiting               │   │
│  │  7. Metrics → Request counting                 │   │
│  └────────────────────────────────────────────────┘   │
└────────────────────┬────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────┐
│                     Handlers                            │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐ │
│  │   Products   │  │    Health    │  │   Testing    │ │
│  │   Handler    │  │   Handler    │  │   Handler    │ │
│  └──────┬───────┘  └──────────────┘  └──────────────┘ │
│         │                                               │
└─────────┼───────────────────────────────────────────────┘
          │
          ▼
┌─────────────────────────────────────────────────────────┐
│                   Memory Store                          │
│  (Thread-safe in-memory map with sync.RWMutex)         │
└─────────────────────────────────────────────────────────┘

              ┌───────────────────────────┐
              │   OBI eBPF Agent          │
              │   (Automatic Tracing)     │
              │   - HTTP Requests         │
              │   - Latency Tracking      │
              │   - Error Detection       │
              │   - Distributed Tracing   │
              └───────────────────────────┘
```

### OBI Instrumentation

OBI automatically captures HTTP traffic **without any code changes**:

1. **Request Initiation**: OBI intercepts socket connections at the kernel level
2. **HTTP Parsing**: Parses HTTP headers, method, path, query parameters
3. **Trace Creation**: Creates distributed trace spans with unique IDs
4. **Response Capture**: Records status codes, response times, errors
5. **Metrics Export**: Sends data to Prometheus, Grafana, and other backends

**Zero Code Changes Required**: This application has no tracing SDKs, no instrumentation libraries, and no manual span creation. Pure business logic.

## Deployment

### Kubernetes Deployment

```bash
# Build and push Docker image
docker build -t your-registry/http-api:latest .
docker push your-registry/http-api:latest

# Update image in deployment manifest
# Edit deployments/examples/01-http-api/deployment.yaml

# Deploy to Kubernetes
kubectl apply -f ../../deployments/examples/01-http-api/

# Verify deployment
kubectl get pods -l app=http-api
kubectl get svc http-api

# Check logs
kubectl logs -l app=http-api -f
```

### Verify OBI Instrumentation

```bash
# Port-forward to the service
kubectl port-forward svc/http-api 8080:80

# Make test requests
curl http://localhost:8080/products
curl http://localhost:8080/health

# Check OBI agent logs for captured traces
kubectl logs -l app=obi-agent | grep http-api

# View traces in Grafana
# Navigate to: Explore → Tempo → Service: http-api
```

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP server port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `LOG_FORMAT` | `json` | Log format (json, console) |
| `RATE_LIMIT_RPS` | `100` | Rate limit per IP (requests/second) |
| `APP_NAME` | `product-catalog` | Application name |
| `ENVIRONMENT` | `demo` | Environment name |

## Performance

### Baseline Performance

- **Throughput**: 10,000+ requests/second
- **Latency**: p50 < 5ms, p95 < 10ms, p99 < 20ms
- **Memory**: ~50MB resident memory
- **CPU**: <10% CPU under load

### OBI Overhead

- **CPU**: <0.5% additional CPU
- **Memory**: <50MB for OBI agent (shared across all services)
- **Latency**: <100μs per request
- **Network**: Minimal (batched exports to backend)

### Load Testing

See load testing results in `docs/PERFORMANCE.md`.

## Development

### Project Structure

```
examples/01-http-api/
├── cmd/
│   └── server/
│       └── main.go              # Application entrypoint
├── internal/
│   ├── handlers/                # HTTP handlers
│   │   ├── products.go          # Product CRUD
│   │   ├── health.go            # Health checks
│   │   └── testing.go           # Testing endpoints
│   ├── middleware/              # HTTP middleware
│   │   ├── logger.go            # Structured logging
│   │   ├── recovery.go          # Panic recovery
│   │   ├── ratelimit.go         # Rate limiting
│   │   ├── cors.go              # CORS headers
│   │   ├── requestid.go         # Request ID generation
│   │   ├── timeout.go           # Request timeouts
│   │   └── metrics.go           # Metrics collection
│   ├── models/                  # Data models
│   │   └── product.go           # Product model
│   └── store/                   # Data storage
│       └── memory.go            # In-memory store
├── tests/                       # Integration tests
│   ├── integration_test.go      # End-to-end tests
│   ├── benchmark_test.go        # Performance benchmarks
│   └── server.go                # Test server helper
├── docs/                        # Documentation
│   ├── API.md                   # API reference
│   ├── ARCHITECTURE.md          # Architecture details
│   ├── DEPLOYMENT.md            # Deployment guide
│   └── TROUBLESHOOTING.md       # Common issues
├── Dockerfile                   # Multi-stage Docker build
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── README.md                    # This file
```

### Makefile Targets

```bash
make build              # Build the binary
make run                # Run locally
make test               # Run all tests
make test-coverage      # Generate coverage report
make lint               # Run linters
make fmt                # Format code
make docker-build       # Build Docker image
make docker-run         # Run Docker container
make clean              # Clean build artifacts
make all                # Run all checks and build
```

## Troubleshooting

See [docs/TROUBLESHOOTING.md](docs/TROUBLESHOOTING.md) for common issues and solutions.

## Related Documentation

- [API Reference](docs/API.md) - Complete API documentation
- [Architecture](docs/ARCHITECTURE.md) - Detailed architecture overview
- [Deployment Guide](docs/DEPLOYMENT.md) - Step-by-step deployment
- [OBI Instrumentation](../../docs/examples/http-instrumentation.md) - How OBI captures HTTP traffic
- [Grafana Dashboard](../../lib/grafana/dashboards/examples/http-api-dashboard.json) - HTTP metrics visualization

## License

See the main MOP repository LICENSE file.

## Contributing

See CONTRIBUTING.md in the main repository.
