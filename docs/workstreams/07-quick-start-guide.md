# Reference Implementations - Quick Start Guide

## 5-Minute Overview

### What is This?
A collection of 5 production-grade applications demonstrating OBI's automatic eBPF instrumentation across all supported protocols. **No code changes required** - OBI automatically captures traces, metrics, and logs.

### What Can You Do With This?
1. **Learn OBI**: See zero-code instrumentation in action
2. **Test Scenarios**: Run pre-built troubleshooting scenarios
3. **Benchmark**: Load test with included k6 scripts
4. **Visualize**: Explore protocol-specific Grafana dashboards
5. **Reference**: Use as templates for your own services

---

## Quick Deploy (All Examples)

```bash
# Prerequisites: Kubernetes cluster, OBI deployed, Grafana stack running

# 1. Clone repository
cd /Users/beengud/raibid-labs/mop

# 2. Deploy all examples
kubectl apply -k deployments/examples/http-rest-api/
kubectl apply -k deployments/examples/grpc-microservice/
kubectl apply -k deployments/examples/sql-application/
kubectl apply -k deployments/examples/redis-cache/
kubectl apply -k deployments/examples/kafka-streaming/

# 3. Wait for deployments
kubectl wait --for=condition=available --timeout=300s \
  deployment/http-api \
  deployment/grpc-service \
  deployment/sql-app \
  deployment/redis-app \
  deployment/kafka-app

# 4. Port-forward services
kubectl port-forward svc/http-api 8080:80 &
kubectl port-forward svc/grpc-service 9090:9090 &
kubectl port-forward svc/sql-app 8000:8000 &
kubectl port-forward svc/redis-app 3000:3000 &

# 5. Import dashboards to Grafana
curl -X POST http://grafana:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @dashboards/examples/http-rest-api-dashboard.json

# 6. Run load tests (optional)
kubectl apply -f deployments/examples/load-generators/k6-http-load.yaml
```

---

## Protocol-Specific Quick Starts

### HTTP REST API (Go + Echo)

**Deploy**:
```bash
kubectl apply -k deployments/examples/http-rest-api/
kubectl port-forward svc/http-api 8080:80
```

**Test**:
```bash
# Create user
curl -X POST http://localhost:8080/api/v1/users \
  -H "Content-Type: application/json" \
  -d '{"username": "alice", "email": "alice@example.com"}'

# List users
curl http://localhost:8080/api/v1/users

# Get user
curl http://localhost:8080/api/v1/users/1
```

**View Traces**: Grafana → Explore → Tempo → Query: `{service.name="http-api"}`

**Dashboard**: Grafana → Dashboards → "HTTP REST API Overview"

**Scenarios**:
```bash
# Trigger slow response
kubectl apply -f examples/00-shared/scenarios/slow-query.yaml

# Trigger error spike
kubectl apply -f examples/00-shared/scenarios/error-spike.yaml
```

---

### gRPC Microservice (Go + grpc-go)

**Deploy**:
```bash
kubectl apply -k deployments/examples/grpc-microservice/
kubectl port-forward svc/grpc-service 9090:9090
```

**Test**:
```bash
# Install grpcurl
brew install grpcurl

# List services
grpcurl -plaintext localhost:9090 list

# Call GetUser RPC
grpcurl -plaintext -d '{"user_id": "123"}' \
  localhost:9090 example.UserService/GetUser

# Call ListUsers (streaming)
grpcurl -plaintext -d '{"limit": 10}' \
  localhost:9090 example.UserService/ListUsers
```

**View Traces**: Filter by `{rpc.service="example.UserService"}`

**Dashboard**: "gRPC Microservice Overview"

---

### SQL Application (Python + FastAPI + PostgreSQL)

**Deploy**:
```bash
kubectl apply -k deployments/examples/sql-application/
kubectl port-forward svc/sql-app 8000:8000
```

**Test**:
```bash
# Create post
curl -X POST http://localhost:8000/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{"title": "Hello World", "content": "This is a test post"}'

# Get post (triggers N+1 query scenario)
curl http://localhost:8000/api/v1/posts/1

# List posts
curl http://localhost:8000/api/v1/posts?limit=10&offset=0
```

**View Slow Queries**:
Grafana → Dashboard → "SQL Application" → "Slowest Queries" panel

**Database Access**:
```bash
kubectl exec -it statefulset/postgres -- psql -U postgres -d mop
\dt  # List tables
SELECT * FROM posts LIMIT 5;
```

---

### Redis Cache (Node.js + Express)

**Deploy**:
```bash
kubectl apply -k deployments/examples/redis-cache/
kubectl port-forward svc/redis-app 3000:3000
```

**Test**:
```bash
# Create session
curl -X POST http://localhost:3000/api/v1/sessions \
  -H "Content-Type: application/json" \
  -d '{"user_id": "user123", "data": {"role": "admin"}}'

# Get session (cache hit)
curl http://localhost:3000/api/v1/sessions/user123

# Get product (cache-aside pattern)
curl http://localhost:3000/api/v1/products/prod-456

# Check rate limit
curl -X POST http://localhost:3000/api/v1/rate-limit/check \
  -d '{"user_id": "user123"}'
```

**View Cache Metrics**:
Dashboard → "Redis Cache" → "Cache Hit Ratio" panel

---

### Kafka Streaming (Java + Spring Boot)

**Deploy**:
```bash
kubectl apply -k deployments/examples/kafka-streaming/
kubectl port-forward svc/kafka-app 8081:8080
```

**Test**:
```bash
# Produce user event
curl -X POST http://localhost:8081/api/v1/events/user \
  -H "Content-Type: application/json" \
  -d '{"event_type": "user_created", "user_id": "u123"}'

# Produce order event
curl -X POST http://localhost:8081/api/v1/events/order \
  -H "Content-Type: application/json" \
  -d '{"event_type": "order_placed", "order_id": "o456"}'

# Check consumer lag
curl http://localhost:8081/api/v1/metrics/consumer-lag
```

**View Kafka Metrics**:
Dashboard → "Kafka Streaming" → "Consumer Lag" panel

**Kafka Access**:
```bash
kubectl exec -it kafka-0 -- kafka-console-consumer \
  --bootstrap-server localhost:9092 \
  --topic user-events \
  --from-beginning
```

---

## Load Testing Quick Start

All examples include pre-configured k6 load tests.

### Run HTTP Load Test
```bash
kubectl apply -f deployments/examples/load-generators/k6-http-load.yaml

# Watch progress
kubectl logs -f job/k6-http-load

# View results in Grafana
# Dashboard → "HTTP REST API" → time range: last 10 minutes
```

### Customize Load Profile
```bash
# Edit ConfigMap
kubectl edit configmap k6-http-config

# Change these values:
# - VUS: Virtual users (concurrent connections)
# - DURATION: Test duration
# - RPS: Requests per second

# Restart load test
kubectl delete job k6-http-load
kubectl apply -f deployments/examples/load-generators/k6-http-load.yaml
```

---

## Troubleshooting Scenarios

### Scenario 1: Slow SQL Query (Missing Index)

**Trigger**:
```bash
kubectl apply -f examples/00-shared/scenarios/slow-query.yaml
```

**What Happens**:
- SQL application starts executing slow queries (>500ms)
- Queries scan full table instead of using index

**Diagnosis Steps**:
1. Open Grafana → "SQL Application" dashboard
2. Check "Slowest Queries" panel
3. Click trace exemplar → Open in Tempo
4. Inspect `db.statement` attribute
5. Notice query lacks WHERE clause optimization

**Resolution**:
```sql
-- Connect to PostgreSQL
kubectl exec -it statefulset/postgres -- psql -U postgres -d mop

-- Add missing index
CREATE INDEX idx_posts_user_id ON posts(user_id);

-- Verify performance improvement
EXPLAIN ANALYZE SELECT * FROM posts WHERE user_id = 123;
```

**Expected Outcome**: Query time drops from 500ms → 5ms

---

### Scenario 2: HTTP Error Spike

**Trigger**:
```bash
kubectl apply -f examples/00-shared/scenarios/error-spike.yaml
```

**What Happens**:
- HTTP API starts returning 500 errors at 20% rate
- Error rate exceeds SLO threshold

**Diagnosis Steps**:
1. Grafana → "HTTP REST API" dashboard
2. "Error Rate" panel shows spike
3. Filter traces by `{http.status_code=500}`
4. Inspect error traces for stack trace

**Resolution**: ConfigMap will auto-revert after 5 minutes

---

### Scenario 3: Redis Connection Leak

**Trigger**:
```bash
kubectl apply -f examples/00-shared/scenarios/connection-leak.yaml
```

**What Happens**:
- Redis app stops closing connections
- Connection pool exhausted
- New requests fail with "connection timeout"

**Diagnosis Steps**:
1. Grafana → "Redis Cache" dashboard
2. "Connection Count" panel shows constant growth
3. Check application logs for connection errors

**Resolution**: Restart application to reset connection pool

---

### Scenario 4: Kafka Consumer Lag

**Trigger**:
```bash
kubectl apply -f examples/00-shared/scenarios/consumer-lag.yaml
```

**What Happens**:
- Kafka producer rate increases to 10,000 msg/sec
- Consumer can only process 1,000 msg/sec
- Lag grows continuously

**Diagnosis Steps**:
1. Grafana → "Kafka Streaming" dashboard
2. "Consumer Lag" panel shows growth
3. Compare producer vs consumer throughput

**Resolution**:
```bash
# Scale consumer instances
kubectl scale deployment kafka-app --replicas=5

# Verify lag decreases
curl http://localhost:8081/api/v1/metrics/consumer-lag
```

---

## Expected OBI Telemetry

### What OBI Automatically Captures (No Code Changes)

**HTTP**:
```
Spans:
  - http.method: GET, POST, PUT, DELETE
  - http.route: /api/v1/users/:id
  - http.status_code: 200, 404, 500
  - http.request_content_length: bytes
  - http.response_content_length: bytes

Metrics:
  - http_server_requests_total{method, route, status_code}
  - http_server_request_duration_seconds{method, route}
```

**gRPC**:
```
Spans:
  - rpc.system: grpc
  - rpc.service: example.UserService
  - rpc.method: GetUser, ListUsers
  - rpc.grpc.status_code: OK, INVALID_ARGUMENT, DEADLINE_EXCEEDED

Metrics:
  - grpc_server_requests_total{service, method, status}
  - grpc_server_request_duration_seconds{service, method}
```

**SQL**:
```
Spans:
  - db.system: postgresql
  - db.name: mop
  - db.statement: SELECT * FROM users WHERE id = $1
  - db.operation: SELECT, INSERT, UPDATE, DELETE
  - db.sql.table: users, posts, comments

Metrics:
  - db_query_duration_seconds{operation, table}
  - db_query_total{operation, table, status}
```

**Redis**:
```
Spans:
  - db.system: redis
  - db.operation: GET, SET, DEL, INCR, ZADD
  - db.redis.key: session:user123

Metrics:
  - redis_command_duration_seconds{operation}
  - redis_command_total{operation, status}
```

**Kafka**:
```
Spans:
  - messaging.system: kafka
  - messaging.destination: user-events
  - messaging.operation: publish, receive
  - messaging.kafka.partition: 0, 1, 2

Metrics:
  - kafka_producer_requests_total{topic}
  - kafka_consumer_lag{topic, partition}
```

---

## Grafana Dashboard Navigation

### Global Dashboard Variables

All dashboards support these filters:
- `$namespace`: Kubernetes namespace
- `$pod`: Specific pod instance
- `$time_range`: Time window for queries

### Common Panels

**Request Rate**:
- Shows operations per second
- Grouped by status code/method
- Useful for capacity planning

**Latency Percentiles**:
- p50, p90, p95, p99 latency
- Helps identify tail latency issues
- Correlates with SLO targets

**Error Rate**:
- Percentage of failed requests
- Broken down by error type
- Alert threshold visualization

**Trace Exemplars**:
- Click any metric data point → View trace in Tempo
- Shows end-to-end distributed trace
- Includes all protocol hops

---

## Integration Test Validation

Each example includes an integration test that validates OBI telemetry.

### Run All Tests
```bash
cd /Users/beengud/raibid-labs/mop
./tests/examples/run-all-tests.sh
```

### Run Single Test
```bash
./tests/examples/http-rest-api-test.sh
```

**What Tests Verify**:
1. Application is healthy (HTTP 200 on /health)
2. Traces exist in Tempo
3. Trace structure is correct (expected span count, attributes)
4. Metrics exist in Mimir
5. Metrics have correct labels
6. Dashboard queries return data

---

## CI/CD Pipeline

All examples are automatically built, tested, and deployed via GitHub Actions.

### Pipeline Stages

1. **Build**: Compile applications, run unit tests
2. **Containerize**: Build Docker images, scan for vulnerabilities
3. **Integration Test**: Deploy to test cluster, run validation
4. **Push**: Push images to registry
5. **Deploy**: Deploy to dev environment
6. **Smoke Test**: Run basic sanity checks

### Trigger Manually
```bash
# GitHub Actions
gh workflow run "Reference Implementations CI/CD"

# Or via git tag
git tag v1.0.0-http-api
git push origin v1.0.0-http-api
```

---

## Common Issues & Solutions

### Issue: Traces Not Appearing in Tempo

**Symptoms**:
- Dashboard panels empty
- Tempo search returns no results

**Diagnosis**:
```bash
# Check OBI DaemonSet
kubectl get pods -n mop-system -l app=obi

# Check OBI logs
kubectl logs -n mop-system -l app=obi | grep OTLP

# Verify Tempo is receiving data
kubectl port-forward -n mop-traces svc/tempo 3200:3200
curl http://localhost:3200/api/search
```

**Solution**:
- Ensure OBI DaemonSet is running on all nodes
- Verify OTLP endpoint configuration
- Check network policies allow OBI → Tempo

---

### Issue: Metrics Not in Mimir

**Symptoms**:
- Grafana dashboards show "No data"
- Prometheus queries return empty

**Diagnosis**:
```bash
# Check Mimir ingestion
kubectl logs -n mop-metrics -l app=mimir-distributor

# Verify Alloy is scraping OBI
kubectl logs -n mop-system -l app=alloy | grep metrics
```

**Solution**:
- Verify OBI exports metrics to Alloy
- Check Alloy → Mimir export configuration
- Ensure metric names match dashboard queries

---

### Issue: Application Won't Start

**Symptoms**:
- Pod in `CrashLoopBackOff`
- Liveness probe failing

**Diagnosis**:
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

**Common Causes**:
1. **Database not ready**: Check PostgreSQL/Redis/Kafka StatefulSet
2. **ConfigMap missing**: Verify ConfigMaps exist
3. **Resource limits**: Pod OOMKilled, increase memory limit
4. **Image pull error**: Check registry credentials

---

## Next Steps

1. **Explore Examples**: Deploy one protocol at a time
2. **Run Load Tests**: See OBI under realistic load
3. **Trigger Scenarios**: Learn troubleshooting workflows
4. **Customize**: Modify applications for your use case
5. **Contribute**: Add new scenarios, improve dashboards

---

## Documentation Index

- [Full Workstream Plan](/Users/beengud/raibid-labs/mop/docs/workstreams/07-reference-implementations.md)
- [Executive Summary](/Users/beengud/raibid-labs/mop/docs/workstreams/07-reference-implementations-summary.md)
- [HTTP REST API Guide](/Users/beengud/raibid-labs/mop/docs/examples/http-rest-api.md)
- [gRPC Microservice Guide](/Users/beengud/raibid-labs/mop/docs/examples/grpc-microservice.md)
- [SQL Application Guide](/Users/beengud/raibid-labs/mop/docs/examples/sql-application.md)
- [Redis Cache Guide](/Users/beengud/raibid-labs/mop/docs/examples/redis-cache.md)
- [Kafka Streaming Guide](/Users/beengud/raibid-labs/mop/docs/examples/kafka-streaming.md)
- [OBI Instrumentation Patterns](/Users/beengud/raibid-labs/mop/docs/examples/obi-instrumentation-patterns.md)
- [Troubleshooting Scenarios](/Users/beengud/raibid-labs/mop/docs/examples/troubleshooting-scenarios.md)
- [Load Testing Guide](/Users/beengud/raibid-labs/mop/docs/examples/load-testing-guide.md)

---

**Quick Links**:
- [OBI Documentation](https://opentelemetry.io/blog/2025/obi-announcing-first-release/)
- [k6 Load Testing](https://k6.io/docs/)
- [Grafana Dashboards](https://grafana.com/docs/grafana/latest/dashboards/)

**Support**: File issues in GitHub with `workstream-7` label

---

**Last Updated**: 2025-11-08
