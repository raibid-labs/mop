# Load Generators

Comprehensive load testing suite for the Multi-Protocol Observability (MOP) project. This collection provides production-ready load generators for all five protocol examples, enabling realistic traffic generation and performance testing.

## Overview

This directory contains custom-built load generators for:

1. **HTTP** - RESTful API load testing
2. **gRPC** - RPC service load testing
3. **SQL** - Database load testing
4. **Redis** - Cache load testing
5. **Kafka** - Streaming load testing

Each generator supports multiple load patterns, configurable throughput, and exports Prometheus metrics for observability.

## Quick Start

### Docker Compose (Recommended)

Run all examples and load generators together:

```bash
# Start all services and load generators
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

Access monitoring:
- Prometheus: http://localhost:9095
- Grafana: http://localhost:3000 (admin/admin)

### Individual Generators

Each generator can be run standalone:

```bash
# HTTP Load Generator
cd 01-http
go run ./cmd/main.go -target http://localhost:8080/products -rps 100 -duration 1m

# gRPC Load Generator
cd 02-grpc
go run ./cmd/main.go -target localhost:9090 -method auth.v1.AuthService/Login -rps 100

# SQL Load Generator
cd 03-sql
go run ./cmd/main.go -db-host localhost -tps 100 -query-type mixed

# Redis Load Generator
cd 04-redis
go run ./cmd/main.go -redis-addr localhost:6379 -rps 1000 -op-type mixed

# Kafka Load Generator
cd 05-kafka
go run ./cmd/main.go -brokers localhost:9092 -topic load-test -mps 100
```

## Load Patterns

All generators support configurable load patterns:

### Constant Load
Maintains steady request rate:
```bash
-pattern constant -rps 100
```

### Spike Load
Periodic bursts to test resilience:
```bash
-pattern spike -rps 50 -max-rps 500
```
Generates 50 RPS baseline with spikes to 500 RPS every 10 seconds.

### Ramp Load
Gradual increase to find limits:
```bash
-pattern ramp -rps 10 -max-rps 200 -duration 5m
```
Ramps from 10 to 200 RPS over 5 minutes.

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Load Generators                           │
│  ┌────────────┐  ┌────────────┐  ┌────────────────────┐    │
│  │   HTTP     │  │   gRPC     │  │    SQL/Redis/      │    │
│  │ Generator  │  │ Generator  │  │  Kafka Generators  │    │
│  └─────┬──────┘  └─────┬──────┘  └──────────┬─────────┘    │
│        │                │                     │               │
└────────┼────────────────┼─────────────────────┼──────────────┘
         │                │                     │
         │                │                     │
         ▼                ▼                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Target Services                            │
│  ┌────────────┐  ┌────────────┐  ┌────────────────────┐    │
│  │  HTTP API  │  │   gRPC     │  │  PostgreSQL/Redis  │    │
│  │  Service   │  │  Service   │  │  /Kafka Services   │    │
│  └────────────┘  └────────────┘  └────────────────────┘    │
└─────────────────────────────────────────────────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │   OBI eBPF Agent     │
                │  (Auto-Instrument)   │
                └──────────────────────┘
                           │
                           ▼
                ┌──────────────────────┐
                │    Prometheus        │
                │    Grafana           │
                └──────────────────────┘
```

## Features

### Common Features

All generators include:

- **Multiple Load Patterns**: Constant, spike, ramp, step, wave
- **Prometheus Metrics**: Request counts, latencies, error rates
- **Configurable Throughput**: Adjust RPS/TPS/MPS via flags or env vars
- **JSON/Text Reports**: Machine and human-readable output
- **Docker Support**: Pre-built containers for easy deployment
- **Kubernetes Ready**: CronJob manifests included
- **Low Overhead**: <5% CPU impact on test results
- **Graceful Shutdown**: Clean termination with SIGTERM

### Protocol-Specific Features

#### HTTP (01-http)
- Configurable HTTP methods (GET, POST, PUT, DELETE)
- Custom headers and body
- Connection pooling
- Rate limiting testing

#### gRPC (02-grpc)
- Dynamic method invocation
- Request/response validation
- Connection reuse
- TLS support

#### SQL (03-sql)
- Read/write/mixed workloads
- Connection pooling
- Transaction support
- Realistic query patterns

#### Redis (04-redis)
- GET/SET/mixed operations
- Configurable key/value sizes
- Pipeline support
- Multiple database support

#### Kafka (05-kafka)
- Producer load testing
- Compression support (none, gzip, snappy, lz4)
- Batching configuration
- Multiple broker support

## Metrics

Each generator exports Prometheus metrics on a unique port:

| Generator | Port | Metrics Prefix |
|-----------|------|----------------|
| HTTP | 9090 | `http_load_*` |
| gRPC | 9091 | `grpc_load_*` |
| SQL | 9092 | `sql_load_*` |
| Redis | 9093 | `redis_load_*` |
| Kafka | 9094 | `kafka_load_*` |

### Common Metrics

All generators export:
- `*_requests_total` / `*_transactions_total` - Total operations
- `*_duration_seconds` - Operation latency histogram
- `*_in_flight` / `*_active_connections` - Current operations

Access metrics:
```bash
curl http://localhost:9090/metrics  # HTTP generator
curl http://localhost:9091/metrics  # gRPC generator
# etc.
```

## Configuration

### Environment Variables

All generators support configuration via environment variables:

```bash
# HTTP Generator
export TARGET_URL=http://localhost:8080/products
export LOAD_PATTERN=constant
export RPS=100
export DURATION=5m

# gRPC Generator
export TARGET=localhost:9090
export METHOD=auth.v1.AuthService/Login
export DATA='{"username":"test","password":"test"}'

# SQL Generator
export DB_HOST=localhost
export DB_NAME=orders
export TPS=100
export QUERY_TYPE=mixed

# Redis Generator
export REDIS_ADDR=localhost:6379
export RPS=1000
export OP_TYPE=mixed

# Kafka Generator
export KAFKA_BROKERS=localhost:9092
export KAFKA_TOPIC=load-test
export MPS=100
```

### Command-Line Flags

All environment variables have corresponding flags:

```bash
./http-load-gen -target URL -pattern PATTERN -rps N -duration TIME
./grpc-load-gen -target ADDR -method METHOD -data JSON
./sql-load-gen -db-host HOST -tps N -query-type TYPE
./redis-load-gen -redis-addr ADDR -rps N -op-type TYPE
./kafka-load-gen -brokers ADDRS -topic TOPIC -mps N
```

## Deployment

### Kubernetes CronJobs

Deploy as scheduled load tests:

```bash
# Deploy all CronJobs
kubectl apply -f 01-http/deployments/cronjob.yaml
kubectl apply -f 02-grpc/deployments/cronjob.yaml
kubectl apply -f 03-sql/deployments/cronjob.yaml
kubectl apply -f 04-redis/deployments/cronjob.yaml
kubectl apply -f 05-kafka/deployments/cronjob.yaml

# View running jobs
kubectl get cronjobs
kubectl get jobs

# View logs
kubectl logs -l app=http-load-generator
```

CronJobs run every 5 minutes by default. Modify `spec.schedule` to change frequency.

### Docker Containers

Build and run individual containers:

```bash
# Build
cd 01-http
docker build -t http-load-generator .

# Run
docker run --rm \
  -e TARGET_URL=http://host.docker.internal:8080/products \
  -e RPS=100 \
  -e DURATION=1m \
  http-load-generator
```

## Testing Scenarios

### Baseline Performance

Test normal operation:
```bash
docker-compose up -d
# Wait for services to stabilize
sleep 30
# Run load generators for 5 minutes
```

### Stress Testing

Find breaking points:
```bash
# Ramp pattern
-pattern ramp -rps 10 -max-rps 1000 -duration 10m
```

### Burst Testing

Test resilience:
```bash
# Spike pattern
-pattern spike -rps 100 -max-rps 2000
```

### Sustained Load

Long-running tests:
```bash
# Constant pattern for hours
-pattern constant -rps 100 -duration 4h
```

## Monitoring

### Real-Time Monitoring

1. Start all services:
```bash
docker-compose up -d
```

2. Access Grafana: http://localhost:3000
   - Username: admin
   - Password: admin

3. Add Prometheus data source:
   - URL: http://prometheus:9090

4. Import dashboards (optional)

### Metrics Analysis

View live metrics:
```bash
# Watch HTTP generator metrics
watch -n 1 'curl -s http://localhost:9090/metrics | grep http_load'

# View all generator metrics
curl http://localhost:9095/api/v1/targets  # Prometheus targets
```

## Performance Benchmarks

Expected performance on modern hardware (4 cores, 8GB RAM):

| Generator | Max Throughput | Avg Latency | Memory Usage |
|-----------|----------------|-------------|--------------|
| HTTP | 10,000+ RPS | 5-10ms | <100MB |
| gRPC | 10,000+ RPS | 3-8ms | <100MB |
| SQL | 5,000+ TPS | 10-20ms | <100MB |
| Redis | 50,000+ OPS | <1ms | <100MB |
| Kafka | 10,000+ MPS | 5-15ms | <150MB |

## Troubleshooting

### Connection Issues

```bash
# Check services are running
docker-compose ps

# Check service health
curl http://localhost:8080/health  # HTTP API
grpcurl -plaintext localhost:9090 list  # gRPC

# Check network connectivity
docker-compose exec http-load-generator ping http-api
```

### High Failure Rates

- Reduce RPS/TPS/MPS
- Increase timeout values
- Check target service capacity
- Review service logs

### Missing Metrics

```bash
# Verify metrics endpoints
curl http://localhost:9090/metrics  # HTTP
curl http://localhost:9091/metrics  # gRPC

# Check Prometheus scraping
curl http://localhost:9095/api/v1/targets
```

### Memory Issues

- Reduce concurrency/clients
- Decrease test duration
- Limit result collection
- Use batching where available

## Development

### Adding New Patterns

Edit `internal/patterns/patterns.go`:

```go
type CustomLoad struct {
    // your fields
}

func (c *CustomLoad) RPS(second int) int {
    // your logic
    return rps
}
```

### Custom Queries/Operations

Edit `internal/generator/generator.go`:

```go
func (g *Generator) executeCustom(ctx context.Context) error {
    // your custom logic
}
```

### Testing

```bash
# Run unit tests
cd 01-http
go test ./...

# Run integration tests
docker-compose up -d
go test -tags=integration ./...
```

## Related Documentation

- [WS-REF-06 Issues](../docs/issues/ws-ref-06/) - Issue specifications
- [Example Services](../examples/) - Target services
- [OBI Documentation](../docs/obi/) - Observability platform
- [Deployment Guide](../docs/deployment/) - Production deployment

## Support

For issues and questions:
- GitHub Issues: https://github.com/raibid-labs/mop/issues
- Documentation: ../docs/

## License

See main repository LICENSE file.
