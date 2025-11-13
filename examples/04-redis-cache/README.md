# Redis Cache Example - OBI Instrumentation

This example demonstrates automatic Redis instrumentation using OBI (OpenTelemetry-Based Instrumentation) eBPF. It showcases zero-code observability for Redis caching patterns including cache-aside, pub/sub invalidation, and pipeline operations.

## Features

- **Cache-Aside Pattern**: Automatic caching with TTL management
- **Redis Pub/Sub**: Cache invalidation across multiple instances
- **Pipeline Operations**: Bulk get/set operations for performance
- **Cache Statistics**: Hit/miss tracking and performance metrics
- **OBI eBPF Instrumentation**: Zero-code automatic tracing of all Redis operations
- **Multi-Instance Support**: Horizontal scaling with shared cache

## Architecture

```
┌─────────────────┐
│   HTTP Client   │
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Gin HTTP API   │◄──── OBI eBPF (HTTP tracing)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Cache Layer    │◄──── OBI eBPF (Redis tracing)
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│  Redis Server   │
│  + Pub/Sub      │
└─────────────────┘
```

## Cache Patterns Demonstrated

### 1. Cache-Aside (Lazy Loading)

```go
// Try cache first
item, err := cache.Get(ctx, key)
if item != nil {
    return item // Cache hit
}

// Cache miss - fetch from upstream
item = fetchFromUpstream(id)

// Store in cache for next time
cache.Set(ctx, key, item, ttl)
```

### 2. Pub/Sub Invalidation

```go
// On update, invalidate cache across all instances
pubsub.InvalidateKey(ctx, "item:123")

// On delete, invalidate by pattern
pubsub.InvalidatePattern(ctx, "item:*")
```

### 3. Pipeline Operations

```go
// Bulk get for better performance
keys := []string{"item:1", "item:2", "item:3"}
items := cache.GetMultiple(ctx, keys)

// Bulk set
cache.SetMultiple(ctx, itemsMap, ttl)
```

## OBI Instrumentation

OBI automatically captures:

- **All Redis Commands**: GET, SET, DEL, SCAN, PUBLISH, SUBSCRIBE
- **Command Duration**: Latency histograms for performance analysis
- **Cache Hit/Miss Ratio**: Automatic detection based on GET responses
- **Pub/Sub Messages**: Invalidation events and message flow
- **Pipeline Operations**: Batch operation tracing
- **Connection Pool**: Connection usage and health

### What OBI Captures (Zero Code Required)

1. **Redis Protocol Operations**
   - Command name (GET, SET, DEL, etc.)
   - Key patterns
   - Response status (hit/miss for GET)
   - Payload size

2. **Performance Metrics**
   - Operation latency (p50, p95, p99)
   - Throughput (ops/sec)
   - Error rates

3. **Distributed Traces**
   - End-to-end request flow
   - Redis operation within HTTP request context
   - Multi-span traces across services

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/stats` | Cache statistics (hit rate, total keys) |
| POST | `/stats/reset` | Reset cache statistics |
| GET | `/items/:id` | Get item (cache-aside pattern) |
| POST | `/items/batch` | Get multiple items (pipeline) |
| POST | `/items` | Create item |
| PUT | `/items/:id` | Update item (with invalidation) |
| DELETE | `/items/:id` | Delete item (with invalidation) |
| POST | `/cache/invalidate` | Manual cache invalidation |

## Running Locally

### Prerequisites

- Go 1.21+
- Redis 7.2+
- Docker (optional)

### Start Redis

```bash
docker run -d --name redis -p 6379:6379 redis:7.2-alpine
```

### Run the Application

```bash
cd examples/04-redis-cache
go mod download
go run cmd/server/main.go
```

The server will start on `http://localhost:8080`.

### Test the API

```bash
# Health check
curl http://localhost:8080/health

# Get item (cache miss first time)
curl http://localhost:8080/items/1
# Headers: X-Cache-Status: MISS

# Get item again (cache hit)
curl http://localhost:8080/items/1
# Headers: X-Cache-Status: HIT

# Get multiple items
curl -X POST http://localhost:8080/items/batch \
  -H "Content-Type: application/json" \
  -d '{"ids": ["1", "2", "3"]}'

# Create item
curl -X POST http://localhost:8080/items \
  -H "Content-Type: application/json" \
  -d '{
    "id": "100",
    "name": "New Product",
    "price": 99.99,
    "category": "Electronics",
    "stock": 10
  }'

# Update item (triggers invalidation)
curl -X PUT http://localhost:8080/items/100 \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Product", "price": 149.99}'

# Delete item (triggers invalidation)
curl -X DELETE http://localhost:8080/items/100

# Get cache statistics
curl http://localhost:8080/stats

# Manual cache invalidation
curl -X POST http://localhost:8080/cache/invalidate \
  -H "Content-Type: application/json" \
  -d '{"key": "item:1"}'

# Invalidate by pattern
curl -X POST http://localhost:8080/cache/invalidate \
  -H "Content-Type: application/json" \
  -d '{"pattern": "item:*"}'
```

## Running with Docker

### Build Image

```bash
cd examples/04-redis-cache
docker build -t redis-cache-example .
```

### Run with Docker Compose

```bash
docker-compose up -d
```

## Deploying to Kubernetes

### Deploy Redis and Application

```bash
kubectl apply -f deployments/examples/04-redis-cache/redis.yaml
kubectl apply -f deployments/examples/04-redis-cache/app.yaml
```

### Verify Deployment

```bash
kubectl get pods -n mop-examples
kubectl logs -n mop-examples -l app=redis-cache-example -f
```

### Port Forward

```bash
kubectl port-forward -n mop-examples svc/redis-cache-app 8080:80
```

## Testing

### Run Unit Tests

```bash
cd examples/04-redis-cache
go test -v ./internal/cache/...
```

### Run Integration Tests

Requires Redis running on `localhost:6379`:

```bash
go test -v ./tests/...
```

### Run with Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Observability with OBI

### View Traces in Jaeger

1. Deploy OBI collector (see main README)
2. Run the application with OBI enabled
3. Open Jaeger UI: `http://localhost:16686`
4. Search for service: `redis-cache-app`
5. View traces with Redis operations

### Grafana Dashboard

Import the dashboard from `lib/grafana/dashboards/examples/redis-cache-dashboard.json`:

**Panels Include:**
- Redis Operations/sec
- Cache Hit Rate (gauge)
- Total Cached Keys
- Redis Command Latency (p95, p99)
- Operations by Command (GET, SET, DEL, etc.)
- Cache Hits vs Misses
- Pub/Sub Messages
- Redis Memory Usage

### Prometheus Metrics

OBI automatically exports these metrics:

```promql
# Total Redis operations
obi_redis_commands_total{command="GET"}

# Command duration histogram
obi_redis_command_duration_seconds_bucket

# Cache hit rate
sum(rate(obi_redis_commands_total{command="GET",status="hit"}[5m])) /
sum(rate(obi_redis_commands_total{command="GET"}[5m]))

# Pub/sub messages
obi_redis_pubsub_messages_total{channel="cache:invalidate"}

# Total keys
obi_redis_keys_total

# Memory usage
obi_redis_memory_usage_bytes
```

## Performance Considerations

### Cache TTL Strategy

- Default TTL: 5 minutes
- Adjust based on data volatility
- Use shorter TTL for frequently changing data
- Use longer TTL for static reference data

### Pipeline Performance

- Use `GetMultiple`/`SetMultiple` for batch operations
- Reduces network round trips
- 10x faster for bulk operations

### Memory Management

- Redis configured with `maxmemory-policy allkeys-lru`
- Evicts least recently used keys when memory full
- Monitor memory usage in Grafana

### Pub/Sub Considerations

- Pub/sub is fire-and-forget (no guaranteed delivery)
- Combine with TTL for eventual consistency
- Use for cache invalidation hints, not critical updates

## Troubleshooting

### High Cache Miss Rate

1. Check TTL configuration (may be too short)
2. Review memory limits (eviction may be aggressive)
3. Analyze access patterns in Grafana
4. Consider cache warming for popular items

### Slow Redis Operations

1. Check network latency (should be <1ms)
2. Review command distribution (avoid expensive commands)
3. Monitor Redis CPU usage
4. Consider Redis cluster for scaling

### Pub/Sub Not Working

1. Verify Redis connection
2. Check subscription in logs
3. Test with Redis CLI: `SUBSCRIBE cache:invalidate`
4. Ensure multiple instances connected to same Redis

## OBI eBPF Implementation Details

### How OBI Captures Redis Traffic

OBI uses eBPF (extended Berkeley Packet Filter) to trace Redis operations at the kernel level:

1. **Socket-Level Tracing**: Intercepts Redis protocol (RESP) on TCP sockets
2. **Command Parsing**: Extracts command name, keys, and arguments
3. **Timing Instrumentation**: Measures round-trip time for each operation
4. **Context Propagation**: Links Redis operations to parent HTTP traces

### No Code Changes Required

The application uses standard `go-redis/v9` library with **zero instrumentation code**:

```go
// Standard go-redis - OBI captures automatically
client := redis.NewClient(&redis.Options{
    Addr: "redis:6379",
})

// All operations traced by OBI
client.Get(ctx, "key")
client.Set(ctx, "key", value, ttl)
client.Publish(ctx, "channel", message)
```

### What This Means for Production

- **No Performance Overhead**: <1% CPU impact from eBPF
- **No Library Updates**: Works with any Redis client
- **Complete Visibility**: Every operation traced automatically
- **Zero Maintenance**: No SDK updates or version conflicts

## Learn More

- [OBI eBPF Documentation](../../docs/obi-instrumentation.md)
- [Redis Instrumentation Guide](../../docs/examples/redis-instrumentation.md)
- [Cache Patterns Best Practices](../../docs/examples/cache-patterns.md)
- [Load Testing Guide](../../docs/examples/load-testing-guide.md)

## License

Apache 2.0
