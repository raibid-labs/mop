# Kafka Load Generator

A high-throughput Kafka load generator for testing streaming performance with realistic workloads.

## Features

- Multiple load patterns: constant, spike, ramp
- Configurable message rates (MPS)
- Compression support: none, gzip, snappy, lz4
- Batching for high throughput
- Prometheus metrics export
- Multiple producer support

## Quick Start

```bash
# Build
go build -o kafka-load-gen ./cmd

# Run with constant load
./kafka-load-gen \
  -brokers localhost:9092 \
  -topic load-test \
  -pattern constant \
  -mps 100 \
  -duration 1m \
  -message-size 1024
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-brokers` | `localhost:9092` | Kafka broker addresses (comma-separated) |
| `-topic` | `load-test` | Kafka topic |
| `-pattern` | `constant` | Load pattern |
| `-mps` | `100` | Messages per second |
| `-max-mps` | `1000` | Maximum MPS |
| `-producers` | `3` | Number of concurrent producers |
| `-message-size` | `1024` | Message size in bytes |
| `-compression` | `none` | Compression: none, gzip, snappy, lz4 |
| `-duration` | `60s` | Test duration |
| `-metrics-port` | `9094` | Prometheus metrics port |

## Compression

- **none**: No compression (fastest)
- **gzip**: Good compression ratio, higher CPU
- **snappy**: Balanced compression and speed
- **lz4**: Fast compression, moderate ratio

## Metrics

- `kafka_load_messages_total` - Total messages by status
- `kafka_load_message_duration_seconds` - Message send duration
- `kafka_load_bytes_total` - Total bytes sent
- `kafka_load_active_producers` - Active producers

## Performance

- Throughput: 10,000+ MPS on modern hardware
- Supports batching for efficiency
- Multiple producers for parallelism

## License

See main repository LICENSE file.
