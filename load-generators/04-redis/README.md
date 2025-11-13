# Redis Load Generator

A high-performance Redis load generator for testing cache performance with realistic workloads.

## Features

- Multiple load patterns: constant, spike, ramp
- Configurable operation rates (OPS)
- GET/SET/mixed operations
- Connection pooling
- Prometheus metrics export
- Low latency overhead

## Quick Start

```bash
# Build
go build -o redis-load-gen ./cmd

# Run with constant load
./redis-load-gen \
  -redis-addr localhost:6379 \
  -pattern constant \
  -rps 1000 \
  -duration 1m \
  -op-type mixed
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-redis-addr` | `localhost:6379` | Redis address |
| `-redis-password` | `""` | Redis password |
| `-redis-db` | `0` | Redis database |
| `-pattern` | `constant` | Load pattern |
| `-rps` | `1000` | Operations per second |
| `-max-rps` | `5000` | Maximum OPS |
| `-clients` | `10` | Number of concurrent clients |
| `-op-type` | `mixed` | Operation type: get, set, mixed |
| `-key-size` | `16` | Key size in bytes |
| `-value-size` | `256` | Value size in bytes |
| `-duration` | `60s` | Test duration |
| `-metrics-port` | `9093` | Prometheus metrics port |

## Operation Types

- **get**: Only GET operations
- **set**: Only SET operations
- **mixed**: 80% GET, 20% SET (realistic cache workload)

## Metrics

- `redis_load_operations_total` - Total operations by status and type
- `redis_load_operation_duration_seconds` - Operation duration histogram
- `redis_load_active_connections` - Active connections

## License

See main repository LICENSE file.
