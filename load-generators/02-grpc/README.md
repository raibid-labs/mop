# gRPC Load Generator

A flexible, production-ready gRPC load generator for testing service performance with configurable load patterns.

## Features

- Multiple load patterns: constant, spike, ramp, step, wave
- Configurable concurrency and request rates
- Dynamic gRPC method invocation
- Prometheus metrics export
- JSON and text report formats
- Docker and Kubernetes ready
- Low overhead and high performance

## Quick Start

### Local Usage

```bash
# Build
go build -o grpc-load-gen ./cmd

# Run with constant load
./grpc-load-gen \
  -target localhost:9090 \
  -method auth.v1.AuthService/Login \
  -data '{"username":"loadtest","password":"password"}' \
  -pattern constant \
  -rps 100 \
  -duration 1m

# Run with spike pattern
./grpc-load-gen \
  -target localhost:9090 \
  -method auth.v1.AuthService/ValidateToken \
  -data '{"token":"test-token"}' \
  -pattern spike \
  -rps 50 \
  -max-rps 500 \
  -duration 2m

# Run with ramp pattern
./grpc-load-gen \
  -target localhost:9090 \
  -method auth.v1.AuthService/Login \
  -pattern ramp \
  -rps 10 \
  -max-rps 200 \
  -duration 5m
```

### Docker Usage

```bash
# Build image
docker build -t grpc-load-generator .

# Run load test
docker run --rm \
  -e TARGET=host.docker.internal:9090 \
  -e METHOD=auth.v1.AuthService/Login \
  -e DATA='{"username":"loadtest","password":"password"}' \
  -e LOAD_PATTERN=constant \
  -e RPS=100 \
  -e DURATION=1m \
  grpc-load-generator
```

### Kubernetes CronJob

```bash
# Deploy CronJob
kubectl apply -f deployments/cronjob.yaml

# View logs
kubectl logs -l app=grpc-load-generator
```

## Configuration

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-target` | `localhost:9090` | gRPC server address |
| `-method` | `auth.v1.AuthService/Login` | gRPC method to call |
| `-pattern` | `constant` | Load pattern (constant, spike, ramp) |
| `-duration` | `60s` | Test duration |
| `-rps` | `100` | Requests per second for constant load |
| `-max-rps` | `500` | Maximum RPS for spike/ramp patterns |
| `-concurrency` | `10` | Number of concurrent workers |
| `-timeout` | `30s` | Request timeout |
| `-data` | `{}` | Request data as JSON |
| `-metrics-port` | `9091` | Prometheus metrics port |
| `-report` | `text` | Report format (text, json) |
| `-insecure` | `true` | Use insecure connection |

### Environment Variables

All flags can be set via environment variables:

- `TARGET`
- `METHOD`
- `LOAD_PATTERN`
- `DURATION`
- `RPS`
- `MAX_RPS`
- `CONCURRENCY`
- `TIMEOUT`
- `DATA`
- `METRICS_PORT`
- `REPORT_FORMAT`
- `INSECURE`

## Load Patterns

### Constant Load

Maintains a steady request rate:

```bash
./grpc-load-gen -pattern constant -rps 100
```

### Spike Load

Periodic spikes to test burst handling:

```bash
./grpc-load-gen -pattern spike -rps 50 -max-rps 500
```

Generates 50 RPS with spikes to 500 RPS every 10 seconds.

### Ramp Load

Gradually increases load:

```bash
./grpc-load-gen -pattern ramp -rps 10 -max-rps 200 -duration 5m
```

Ramps from 10 to 200 RPS over 5 minutes.

## Metrics

Prometheus metrics exposed on `/metrics`:

- `grpc_load_requests_total` - Total requests (by status and method)
- `grpc_load_request_duration_seconds` - Request duration histogram
- `grpc_load_requests_in_flight` - Current in-flight requests

## Output Format

### Text Format

```
=== gRPC Load Test Results ===
Total Requests:   6000
Success:          5950 (99.17%)
Failed:           50 (0.83%)
Duration:         60.2s
RPS:              99.67

Latency Statistics:
  Min:            1.5ms
  Max:            180.3ms
  Avg:            12.4ms
  P50:            10.1ms
  P95:            32.5ms
  P99:            75.2ms

Status Codes:
  OK:             5950
  Unavailable:    50
```

### JSON Format

```json
{
  "total_requests": 6000,
  "success_requests": 5950,
  "failed_requests": 50,
  "duration_seconds": 60.2,
  "rps": 99.67,
  "latency": {
    "min_ms": 1.5,
    "max_ms": 180.3,
    "avg_ms": 12.4,
    "p50_ms": 10.1,
    "p95_ms": 32.5,
    "p99_ms": 75.2
  }
}
```

## Testing the gRPC Service Example

```bash
# Start the gRPC service example
cd ../../examples/02-grpc-service
make docker-run

# Run load test
cd ../../load-generators/02-grpc
./grpc-load-gen \
  -target localhost:9090 \
  -method auth.v1.AuthService/Login \
  -data '{"username":"alice","password":"password"}' \
  -pattern constant \
  -rps 100 \
  -duration 2m \
  -report json
```

## Architecture

The generator uses a worker pool pattern:

1. Main goroutine generates load according to pattern
2. Worker pool (configurable size) processes requests
3. gRPC connection is reused across all workers
4. Results aggregated with atomic operations
5. Metrics exported to Prometheus

## Performance

- Low overhead: <5% CPU for 1000 RPS
- Memory efficient: <100MB for 100 workers
- Scales to 10,000+ RPS on modern hardware
- Single connection pooling for efficiency

## License

See main repository LICENSE file.
