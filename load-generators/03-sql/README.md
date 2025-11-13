# SQL Load Generator

A flexible PostgreSQL load generator for testing database performance with realistic workloads.

## Features

- Multiple load patterns: constant, spike, ramp
- Configurable transaction rates (TPS)
- Read/write/mixed workloads
- Automatic schema initialization
- Prometheus metrics export
- Connection pooling
- Realistic query patterns

## Quick Start

```bash
# Build
go build -o sql-load-gen ./cmd

# Run with constant load
./sql-load-gen \
  -db-host localhost \
  -db-port 5432 \
  -db-name orders \
  -pattern constant \
  -tps 100 \
  -duration 1m \
  -query-type mixed
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-db-host` | `localhost` | Database host |
| `-db-port` | `5432` | Database port |
| `-db-user` | `postgres` | Database user |
| `-db-password` | `postgres` | Database password |
| `-db-name` | `orders` | Database name |
| `-pattern` | `constant` | Load pattern |
| `-tps` | `100` | Transactions per second |
| `-max-tps` | `500` | Maximum TPS |
| `-clients` | `10` | Number of concurrent clients |
| `-query-type` | `mixed` | Query type: read, write, mixed |
| `-duration` | `60s` | Test duration |
| `-metrics-port` | `9092` | Prometheus metrics port |
| `-report` | `text` | Report format |

## Query Types

- **read**: SELECT queries (70% by ID, 20% recent, 10% aggregates)
- **write**: INSERT/UPDATE operations (70% INSERT, 30% UPDATE)
- **mixed**: 70% read, 30% write

## Metrics

- `sql_load_transactions_total` - Total transactions by status and type
- `sql_load_query_duration_seconds` - Query duration histogram
- `sql_load_active_connections` - Active database connections

## License

See main repository LICENSE file.
