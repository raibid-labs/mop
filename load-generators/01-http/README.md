# HTTP Load Generator

A flexible, production-ready HTTP load generator for testing API performance with configurable load patterns.

## Features

- Multiple load patterns: constant, spike, ramp, step, wave
- Configurable concurrency and request rates
- Prometheus metrics export
- JSON and text report formats
- Docker and Kubernetes ready
- Low overhead and high performance

## Quick Start

### Local Usage

```bash
# Build
go build -o http-load-gen ./cmd

# Run with constant load
./http-load-gen \
  -target http://localhost:8080/products \
  -pattern constant \
  -rps 100 \
  -duration 1m

# Run with spike pattern
./http-load-gen \
  -target http://localhost:8080/products \
  -pattern spike \
  -rps 50 \
  -max-rps 500 \
  -duration 2m

# Run with ramp pattern
./http-load-gen \
  -target http://localhost:8080/products \
  -pattern ramp \
  -rps 10 \
  -max-rps 200 \
  -duration 5m
```

### Docker Usage

```bash
# Build image
docker build -t http-load-generator .

# Run load test
docker run --rm \
  -e TARGET_URL=http://host.docker.internal:8080/products \
  -e LOAD_PATTERN=constant \
  -e RPS=100 \
  -e DURATION=1m \
  http-load-generator
```

### Kubernetes CronJob

```bash
# Deploy CronJob
kubectl apply -f deployments/cronjob.yaml

# View logs
kubectl logs -l app=http-load-generator
```

## Configuration

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-target` | `http://localhost:8080` | Target URL to load test |
| `-pattern` | `constant` | Load pattern (constant, spike, ramp) |
| `-duration` | `60s` | Test duration |
| `-rps` | `100` | Requests per second for constant load |
| `-max-rps` | `500` | Maximum RPS for spike/ramp patterns |
| `-concurrency` | `10` | Number of concurrent workers |
| `-timeout` | `30s` | Request timeout |
| `-method` | `GET` | HTTP method |
| `-body` | `""` | Request body |
| `-headers` | `""` | Headers in key:value format |
| `-metrics-port` | `9090` | Prometheus metrics port |
| `-report` | `text` | Report format (text, json) |

### Environment Variables

All flags can be set via environment variables:

- `TARGET_URL`
- `LOAD_PATTERN`
- `DURATION`
- `RPS`
- `MAX_RPS`
- `CONCURRENCY`
- `TIMEOUT`
- `METHOD`
- `BODY`
- `HEADERS`
- `METRICS_PORT`
- `REPORT_FORMAT`

## Load Patterns

### Constant Load

Maintains a steady request rate:

```bash
./http-load-gen -pattern constant -rps 100
```

### Spike Load

Periodic spikes to test burst handling:

```bash
./http-load-gen -pattern spike -rps 50 -max-rps 500
```

Generates 50 RPS with spikes to 500 RPS every 10 seconds.

### Ramp Load

Gradually increases load:

```bash
./http-load-gen -pattern ramp -rps 10 -max-rps 200 -duration 5m
```

Ramps from 10 to 200 RPS over 5 minutes.

## Metrics

Prometheus metrics exposed on `/metrics`:

- `http_load_requests_total` - Total requests (by status and method)
- `http_load_request_duration_seconds` - Request duration histogram
- `http_load_requests_in_flight` - Current in-flight requests

## Output Format

### Text Format

```
=== Load Test Results ===
Total Requests:   6000
Success:          5950 (99.17%)
Failed:           50 (0.83%)
Duration:         60.2s
RPS:              99.67

Latency Statistics:
  Min:            2.1ms
  Max:            245.3ms
  Avg:            15.4ms
  P50:            12.1ms
  P95:            38.5ms
  P99:            89.2ms

Status Codes:
  200:            5950
  500:            50
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
    "min_ms": 2.1,
    "max_ms": 245.3,
    "avg_ms": 15.4,
    "p50_ms": 12.1,
    "p95_ms": 38.5,
    "p99_ms": 89.2
  }
}
```

## Testing the HTTP API Example

```bash
# Start the HTTP API example
cd ../../examples/01-http-api
make docker-run

# Run load test
cd ../../load-generators/01-http
./http-load-gen \
  -target http://localhost:8080/products \
  -pattern constant \
  -rps 100 \
  -duration 2m \
  -report json
```

## Architecture

The generator uses a worker pool pattern:

1. Main goroutine generates load according to pattern
2. Worker pool (configurable size) processes requests
3. Results aggregated with atomic operations
4. Metrics exported to Prometheus

## Performance

- Low overhead: <5% CPU for 1000 RPS
- Memory efficient: <100MB for 100 workers
- Scales to 10,000+ RPS on modern hardware

## License

See main repository LICENSE file.
