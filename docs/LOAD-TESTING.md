# Load Testing Guide

Comprehensive guide for load testing OBI-instrumented applications to validate performance, scalability, and overhead measurements.

## Table of Contents

1. [Overview](#overview)
2. [Load Testing Tools](#load-testing-tools)
3. [Test Scenarios](#test-scenarios)
4. [Protocol-Specific Tests](#protocol-specific-tests)
5. [Baseline Measurements](#baseline-measurements)
6. [Performance Analysis](#performance-analysis)
7. [Scaling Tests](#scaling-tests)
8. [Best Practices](#best-practices)

## Overview

Load testing OBI-instrumented applications serves multiple purposes:

1. **Validate Zero-Overhead Claims**: Measure actual overhead of OBI instrumentation
2. **Capacity Planning**: Determine maximum throughput and resource requirements
3. **Identify Bottlenecks**: Find performance limitations in application or infrastructure
4. **SLO Validation**: Verify service level objectives under load
5. **Regression Testing**: Detect performance degradation between versions

### Key Metrics to Monitor

- **Throughput**: Requests/queries per second
- **Latency**: p50, p95, p99, p999 response times
- **Error Rate**: Percentage of failed requests
- **CPU Usage**: Application and OBI agent CPU consumption
- **Memory Usage**: Application and OBI agent memory footprint
- **Network I/O**: Bandwidth utilization
- **OBI Overhead**: Difference with and without OBI

## Load Testing Tools

### 1. k6 (Recommended)

Modern load testing tool with excellent Kubernetes support.

**Installation:**
```bash
# Install k6
brew install k6  # macOS
# or
wget https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz
tar -xzf k6-v0.47.0-linux-amd64.tar.gz
sudo mv k6 /usr/local/bin/
```

**Basic HTTP Test:**
```javascript
// http-load-test.js
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp up to 100 users
    { duration: '5m', target: 100 },   // Stay at 100 users
    { duration: '2m', target: 200 },   // Ramp up to 200 users
    { duration: '5m', target: 200 },   // Stay at 200 users
    { duration: '2m', target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],  // 95% of requests < 500ms
    http_req_failed: ['rate<0.01'],    // Error rate < 1%
  },
};

export default function () {
  const res = http.get('http://http-api:8080/products');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });
  sleep(1);
}
```

**Run Test:**
```bash
k6 run http-load-test.js
```

### 2. Hey

Simple HTTP load generator.

**Installation:**
```bash
go install github.com/rakyll/hey@latest
```

**Usage:**
```bash
# 10,000 requests, 50 concurrent workers
hey -n 10000 -c 50 http://http-api:8080/products

# 60 second test, 100 concurrent workers
hey -z 60s -c 100 http://http-api:8080/products

# POST request with JSON body
hey -m POST -n 1000 -c 10 \
  -H "Content-Type: application/json" \
  -d '{"name":"Product","price":99.99}' \
  http://http-api:8080/products
```

### 3. ghz (gRPC Load Testing)

Load testing tool specifically for gRPC.

**Installation:**
```bash
# macOS
brew install ghz

# Linux
wget https://github.com/bojand/ghz/releases/download/v0.117.0/ghz-linux-x86_64.tar.gz
tar -xzf ghz-linux-x86_64.tar.gz
sudo mv ghz /usr/local/bin/
```

**Usage:**
```bash
# Basic gRPC test
ghz --insecure \
  --proto ./proto/users.proto \
  --call users.UserService.GetUser \
  -d '{"id":"123"}' \
  -n 10000 -c 50 \
  grpc-service:50051

# With load profile
ghz --insecure \
  --proto ./proto/users.proto \
  --call users.UserService.GetUser \
  -d '{"id":"123"}' \
  --load-schedule=step \
  --load-start=10 \
  --load-step=10 \
  --load-end=100 \
  --load-step-duration=30s \
  grpc-service:50051
```

### 4. Redis Benchmark

Built-in Redis performance testing tool.

**Usage:**
```bash
# Basic benchmark
redis-benchmark -h redis-cache -p 6379 -n 100000 -c 50

# Specific commands
redis-benchmark -h redis-cache -p 6379 -t set,get -n 100000 -c 50

# Pipeline mode (faster)
redis-benchmark -h redis-cache -p 6379 -n 1000000 -c 50 -P 16
```

### 5. Kafka Performance Tools

Kafka includes built-in performance testing tools.

**Producer Performance Test:**
```bash
kafka-producer-perf-test \
  --topic orders \
  --num-records 1000000 \
  --record-size 1000 \
  --throughput 10000 \
  --producer-props bootstrap.servers=kafka:9092
```

**Consumer Performance Test:**
```bash
kafka-consumer-perf-test \
  --topic orders \
  --messages 1000000 \
  --threads 4 \
  --broker-list kafka:9092
```

## Test Scenarios

### Scenario 1: Baseline Performance (No OBI)

Establish baseline metrics without OBI instrumentation.

**Steps:**

1. **Disable OBI agent:**
```bash
kubectl scale daemonset/obi-agent -n observability --replicas=0
```

2. **Wait for DNS cache to clear:**
```bash
sleep 60
```

3. **Run load test:**
```bash
k6 run --out json=baseline.json baseline-test.js
```

4. **Collect metrics:**
```bash
# CPU usage
kubectl top pods -l app=http-api > baseline-cpu.txt

# Memory usage
kubectl exec -it http-api-xxx -- cat /proc/meminfo > baseline-mem.txt

# Application metrics
curl http://http-api:8080/metrics > baseline-metrics.txt
```

### Scenario 2: OBI Instrumented Performance

Measure performance with OBI instrumentation enabled.

**Steps:**

1. **Enable OBI agent:**
```bash
kubectl scale daemonset/obi-agent -n observability --replicas=1
kubectl rollout status daemonset/obi-agent -n observability
```

2. **Verify instrumentation:**
```bash
kubectl logs -n observability -l app=obi-agent | grep "Instrumented.*http-api"
```

3. **Run same load test:**
```bash
k6 run --out json=instrumented.json baseline-test.js
```

4. **Collect metrics:**
```bash
# Application CPU/memory
kubectl top pods -l app=http-api > instrumented-cpu.txt

# OBI agent CPU/memory
kubectl top pods -n observability -l app=obi-agent > obi-cpu.txt

# Application metrics
curl http://http-api:8080/metrics > instrumented-metrics.txt

# OBI metrics
curl http://obi-agent:9090/metrics > obi-metrics.txt
```

### Scenario 3: Stress Test

Push system to limits to find breaking points.

**k6 Stress Test:**
```javascript
// stress-test.js
export const options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '2m', target: 500 },
    { duration: '2m', target: 1000 },
    { duration: '5m', target: 1000 },
    { duration: '2m', target: 0 },
  ],
};
```

### Scenario 4: Spike Test

Sudden traffic increase to test elasticity.

**k6 Spike Test:**
```javascript
// spike-test.js
export const options = {
  stages: [
    { duration: '10s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '10s', target: 1000 },  // Spike!
    { duration: '3m', target: 1000 },
    { duration: '10s', target: 100 },
    { duration: '1m', target: 100 },
    { duration: '10s', target: 0 },
  ],
};
```

### Scenario 5: Soak Test

Long-running test to detect memory leaks.

**k6 Soak Test:**
```javascript
// soak-test.js
export const options = {
  stages: [
    { duration: '5m', target: 100 },
    { duration: '4h', target: 100 },  // Soak for 4 hours
    { duration: '5m', target: 0 },
  ],
};
```

## Protocol-Specific Tests

### HTTP API Load Test

**Complete k6 HTTP Test:**
```javascript
// http-complete-test.js
import http from 'k6/http';
import { check, group, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 100 },
    { duration: '2m', target: 200 },
    { duration: '5m', target: 200 },
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

const BASE_URL = 'http://http-api:8080';

export default function () {
  group('Product API', function () {
    // List products
    let res = http.get(`${BASE_URL}/products?limit=10`);
    check(res, {
      'list status 200': (r) => r.status === 200,
      'list has products': (r) => JSON.parse(r.body).products.length > 0,
    });

    // Get specific product
    res = http.get(`${BASE_URL}/products/550e8400-e29b-41d4-a716-446655440000`);
    check(res, {
      'get status 200': (r) => r.status === 200,
    });

    // Create product
    const payload = JSON.stringify({
      name: 'Test Product',
      description: 'Load test product',
      price: 99.99,
      stock: 100,
    });

    res = http.post(`${BASE_URL}/products`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });
    check(res, {
      'create status 201': (r) => r.status === 201,
    });

    // Search products
    res = http.get(`${BASE_URL}/search?q=test`);
    check(res, {
      'search status 200': (r) => r.status === 200,
    });
  });

  sleep(1);
}
```

### gRPC Service Load Test

**ghz Configuration:**
```javascript
// grpc-load-test.json
{
  "proto": "./proto/users.proto",
  "call": "users.UserService.GetUser",
  "total": 10000,
  "concurrency": 50,
  "data": {
    "id": "{{randomString 10}}"
  },
  "metadata": {
    "trace-id": "{{uuid}}"
  }
}
```

**Run:**
```bash
ghz --config=grpc-load-test.json grpc-service:50051
```

### SQL Database Load Test

**pgbench (PostgreSQL):**
```bash
# Initialize test database
pgbench -i -s 50 -h postgres -U admin testdb

# Run test
pgbench -c 10 -j 2 -t 10000 -h postgres -U admin testdb

# Custom script
cat > sql-test.sql <<EOF
\set id random(1, 100000)
SELECT * FROM users WHERE id = :id;
EOF

pgbench -c 50 -j 4 -t 10000 -f sql-test.sql -h postgres -U admin testdb
```

### Redis Cache Load Test

**redis-benchmark Advanced:**
```bash
# Mixed workload
redis-benchmark -h redis-cache -p 6379 \
  -t set,get,incr,lpush,rpush,lpop,rpop,sadd,hset,spop,zadd,zpopmin,lrange,mset \
  -n 1000000 -c 50 -q

# Realistic pattern
redis-benchmark -h redis-cache -p 6379 \
  -r 100000 \
  -n 1000000 \
  -t get,set \
  -c 50 \
  --csv > redis-results.csv
```

### Kafka Streaming Load Test

**Producer Load:**
```bash
# High throughput test
kafka-producer-perf-test \
  --topic orders \
  --num-records 10000000 \
  --record-size 1000 \
  --throughput -1 \
  --producer-props \
    bootstrap.servers=kafka:9092 \
    batch.size=16384 \
    linger.ms=10 \
    compression.type=lz4

# Latency test
kafka-producer-perf-test \
  --topic orders \
  --num-records 100000 \
  --record-size 100 \
  --throughput 1000 \
  --producer-props \
    bootstrap.servers=kafka:9092 \
    acks=all \
    batch.size=1
```

## Baseline Measurements

### Expected Performance Metrics

Based on standard hardware (4 CPU, 8GB RAM):

#### HTTP API

| Metric | Without OBI | With OBI | Overhead |
|--------|------------|----------|----------|
| **Throughput** | 10,000 rps | 9,950 rps | -0.5% |
| **Latency p50** | 5ms | 5.02ms | +0.4% |
| **Latency p95** | 15ms | 15.1ms | +0.7% |
| **Latency p99** | 50ms | 50.2ms | +0.4% |
| **CPU Usage** | 25% | 25.5% | +0.5% |
| **Memory** | 100MB | 105MB | +5MB |

#### gRPC Service

| Metric | Without OBI | With OBI | Overhead |
|--------|------------|----------|----------|
| **Throughput** | 15,000 rps | 14,925 rps | -0.5% |
| **Latency p50** | 3ms | 3.01ms | +0.3% |
| **Latency p95** | 10ms | 10.05ms | +0.5% |
| **CPU Usage** | 20% | 20.4% | +0.4% |

#### SQL Database

| Metric | Without OBI | With OBI | Overhead |
|--------|------------|----------|----------|
| **QPS** | 5,000 | 4,975 | -0.5% |
| **Latency p50** | 10ms | 10.05ms | +0.5% |
| **Latency p95** | 50ms | 50.3ms | +0.6% |
| **CPU Usage** | 30% | 30.5% | +0.5% |

#### Redis Cache

| Metric | Without OBI | With OBI | Overhead |
|--------|------------|----------|----------|
| **OPS** | 100,000 | 99,500 | -0.5% |
| **Latency p50** | 0.5ms | 0.502ms | +0.4% |
| **CPU Usage** | 15% | 15.3% | +0.3% |

#### Kafka Streaming

| Metric | Without OBI | With OBI | Overhead |
|--------|------------|----------|----------|
| **Msgs/s** | 50,000 | 49,800 | -0.4% |
| **Latency p50** | 2ms | 2.01ms | +0.5% |
| **CPU Usage** | 25% | 25.5% | +0.5% |

### Collecting Baseline Data

**Automated Script:**
```bash
#!/bin/bash
# collect-baseline.sh

set -e

# Configuration
APP_NAME=${1:-http-api}
DURATION=${2:-300}  # 5 minutes
OUTPUT_DIR="./results/$(date +%Y%m%d-%H%M%S)"

mkdir -p "$OUTPUT_DIR"

echo "Collecting baseline for $APP_NAME (${DURATION}s)..."

# Start monitoring in background
kubectl top pods -l app=$APP_NAME --containers > "$OUTPUT_DIR/resources-start.txt"

# Start metrics collection
kubectl exec -l app=$APP_NAME -- curl -s http://localhost:8080/metrics > "$OUTPUT_DIR/metrics-start.txt"

# Run load test
echo "Running load test..."
k6 run --duration ${DURATION}s --vus 100 \
  --out json="$OUTPUT_DIR/k6-results.json" \
  load-test.js

# End monitoring
kubectl top pods -l app=$APP_NAME --containers > "$OUTPUT_DIR/resources-end.txt"
kubectl exec -l app=$APP_NAME -- curl -s http://localhost:8080/metrics > "$OUTPUT_DIR/metrics-end.txt"

# If OBI is enabled, collect OBI metrics
if kubectl get pods -n observability -l app=obi-agent | grep -q Running; then
  kubectl top pods -n observability -l app=obi-agent > "$OUTPUT_DIR/obi-resources.txt"
  kubectl exec -n observability -it $(kubectl get pods -n observability -l app=obi-agent -o jsonpath='{.items[0].metadata.name}') -- curl -s http://localhost:9090/metrics > "$OUTPUT_DIR/obi-metrics.txt"
fi

echo "Results saved to $OUTPUT_DIR"
```

## Performance Analysis

### Analyzing k6 Results

**Generate HTML Report:**
```bash
# Install k6-reporter
npm install -g k6-to-html

# Generate report
k6-to-html k6-results.json
```

**Extract Key Metrics:**
```bash
# Parse JSON results
jq '.metrics | {
  "http_req_duration": .http_req_duration,
  "http_req_failed": .http_req_failed,
  "http_reqs": .http_reqs,
  "vus": .vus
}' k6-results.json
```

### Comparing Results

**Compare Script:**
```bash
#!/bin/bash
# compare-results.sh

BASELINE="$1"
INSTRUMENTED="$2"

echo "=== Performance Comparison ==="
echo ""

# Extract p95 latency
baseline_p95=$(jq '.metrics.http_req_duration.values["p(95)"]' "$BASELINE")
instrumented_p95=$(jq '.metrics.http_req_duration.values["p(95)"]' "$INSTRUMENTED")
overhead_p95=$(echo "scale=2; ($instrumented_p95 - $baseline_p95) / $baseline_p95 * 100" | bc)

echo "Latency p95:"
echo "  Baseline: ${baseline_p95}ms"
echo "  Instrumented: ${instrumented_p95}ms"
echo "  Overhead: ${overhead_p95}%"
echo ""

# Extract throughput
baseline_rps=$(jq '.metrics.http_reqs.values.rate' "$BASELINE")
instrumented_rps=$(jq '.metrics.http_reqs.values.rate' "$INSTRUMENTED")
overhead_rps=$(echo "scale=2; ($instrumented_rps - $baseline_rps) / $baseline_rps * 100" | bc)

echo "Throughput:"
echo "  Baseline: ${baseline_rps} rps"
echo "  Instrumented: ${instrumented_rps} rps"
echo "  Overhead: ${overhead_rps}%"
echo ""

# Extract error rate
baseline_errors=$(jq '.metrics.http_req_failed.values.rate' "$BASELINE")
instrumented_errors=$(jq '.metrics.http_req_failed.values.rate' "$INSTRUMENTED")

echo "Error Rate:"
echo "  Baseline: ${baseline_errors}"
echo "  Instrumented: ${instrumented_errors}"
```

### Visualizing Results

**Grafana Dashboard Query:**
```promql
# Compare latencies
histogram_quantile(0.95,
  sum by (le) (rate(http_request_duration_seconds_bucket[5m]))
)

# Compare throughput
sum(rate(http_requests_total[5m]))

# Compare error rates
sum(rate(http_requests_total{status_code=~"5.."}[5m])) /
sum(rate(http_requests_total[5m])) * 100
```

## Scaling Tests

### Horizontal Pod Autoscaling (HPA) Test

**Setup HPA:**
```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: http-api
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: http-api
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

**Test Scaling:**
```bash
# Monitor HPA
watch kubectl get hpa http-api

# Start load test
k6 run --vus 500 --duration 10m scaling-test.js

# Observe scaling
kubectl get pods -l app=http-api -w
```

### Multi-Protocol Concurrent Load

**Test all protocols simultaneously:**
```bash
#!/bin/bash
# multi-protocol-load.sh

# Start HTTP load
k6 run --vus 100 --duration 10m http-test.js &
HTTP_PID=$!

# Start gRPC load
ghz --insecure --duration 10m --rps 1000 \
  --proto ./proto/users.proto \
  --call users.UserService.GetUser \
  grpc-service:50051 &
GRPC_PID=$!

# Start Redis load
redis-benchmark -h redis-cache -p 6379 -n 1000000 -c 50 &
REDIS_PID=$!

# Start Kafka load
kafka-producer-perf-test \
  --topic orders \
  --num-records 1000000 \
  --throughput 10000 \
  --record-size 1000 \
  --producer-props bootstrap.servers=kafka:9092 &
KAFKA_PID=$!

# Wait for all tests
wait $HTTP_PID $GRPC_PID $REDIS_PID $KAFKA_PID

echo "All load tests completed"
```

## Best Practices

### 1. Establish Baselines First

Always measure performance without OBI before enabling instrumentation:

```bash
# 1. Deploy application without OBI
kubectl scale daemonset/obi-agent -n observability --replicas=0

# 2. Run baseline tests
./run-baseline-tests.sh

# 3. Enable OBI
kubectl scale daemonset/obi-agent -n observability --replicas=1

# 4. Run instrumented tests
./run-instrumented-tests.sh

# 5. Compare results
./compare-results.sh baseline/ instrumented/
```

### 2. Use Realistic Load Patterns

Model real-world traffic patterns:

```javascript
export const options = {
  scenarios: {
    // Business hours (high traffic)
    business_hours: {
      executor: 'ramping-vus',
      startTime: '0s',
      stages: [
        { duration: '1m', target: 200 },
        { duration: '8h', target: 200 },
        { duration: '1m', target: 0 },
      ],
    },
    // Night time (low traffic)
    night_time: {
      executor: 'constant-vus',
      vus: 20,
      duration: '14h',
      startTime: '10h',
    },
  },
};
```

### 3. Monitor System Resources

Watch for resource saturation:

```bash
# CPU, memory, network
kubectl top nodes
kubectl top pods --all-namespaces

# Disk I/O
kubectl exec -it pod-name -- iostat -x 1

# Network connections
kubectl exec -it pod-name -- ss -s
```

### 4. Test Failure Scenarios

Include error injection:

```javascript
export default function () {
  // 90% success, 10% simulated errors
  if (Math.random() < 0.9) {
    http.get('http://http-api:8080/products');
  } else {
    http.get('http://http-api:8080/error');
  }
}
```

### 5. Run Long Duration Tests

Soak tests reveal memory leaks:

```bash
# 24-hour soak test
k6 run --vus 100 --duration 24h soak-test.js
```

### 6. Document Everything

Create a test report template:

```markdown
# Load Test Report

## Test Details
- Date: 2024-01-15
- Duration: 10 minutes
- Target: http-api
- OBI Version: v1.2.3

## Configuration
- VUs: 100
- RPS Target: 10,000
- Test Type: Steady state

## Results

### Baseline (No OBI)
- Throughput: 10,000 rps
- Latency p95: 15ms
- Error rate: 0.01%
- CPU: 25%
- Memory: 100MB

### Instrumented (With OBI)
- Throughput: 9,950 rps (-0.5%)
- Latency p95: 15.1ms (+0.7%)
- Error rate: 0.01% (same)
- CPU: 25.5% (+0.5%)
- Memory: 105MB (+5MB)

### OBI Agent Resources
- CPU: 0.2 cores
- Memory: 50MB

## Conclusion
OBI overhead is minimal (<1%) and acceptable for production use.
```

### 7. Automate Testing

CI/CD integration:

```yaml
# .github/workflows/load-test.yml
name: Load Test

on:
  pull_request:
    branches: [ main ]

jobs:
  load-test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3

    - name: Setup k6
      run: |
        curl https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz -L | tar xvz
        sudo mv k6 /usr/local/bin

    - name: Run load test
      run: k6 run load-test.js

    - name: Check thresholds
      run: |
        if grep -q "✓" k6-output.txt; then
          echo "Load test passed"
        else
          echo "Load test failed"
          exit 1
        fi
```

## Summary Checklist

- [ ] Establish baseline performance without OBI
- [ ] Test with OBI instrumentation enabled
- [ ] Measure overhead (< 1% target)
- [ ] Test all supported protocols
- [ ] Run stress tests to find limits
- [ ] Run soak tests for memory leaks
- [ ] Test autoscaling behavior
- [ ] Document all results
- [ ] Compare with SLO targets
- [ ] Validate in production-like environment

## Related Documentation

- [OBI Instrumentation Guide](OBI-INSTRUMENTATION-GUIDE.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [Best Practices](BEST-PRACTICES.md)
- [Example Applications](../examples/)

## Support

For questions or issues with load testing:
- GitHub Issues: https://github.com/obi/obi/issues
- Slack: https://obi-community.slack.com
- Email: support@obi.io
