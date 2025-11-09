# MOP Reference Implementations Plan

## Overview

Add reference implementations showcasing OBI eBPF automatic instrumentation for all supported protocols. Each example demonstrates zero-code observability with realistic application patterns.

## Objectives

1. **Demonstrate OBI Capabilities**: Show automatic instrumentation for HTTP, gRPC, SQL, Redis, Kafka
2. **Zero-Code Instrumentation**: No SDK or library changes required
3. **Realistic Patterns**: Production-like applications with common patterns
4. **Complete Observability**: Traces, metrics, logs automatically collected
5. **Troubleshooting Scenarios**: Include common issues (slow queries, errors, high latency)
6. **Load Testing**: Generate realistic traffic patterns
7. **Grafana Dashboards**: Protocol-specific visualizations

---

## Technology Stack

| Protocol | Technology Choice | Rationale |
|----------|------------------|-----------|
| **HTTP/REST** | Go (Gin framework) | Fast, simple, great HTTP support |
| **gRPC** | Go (google.golang.org/grpc) | Native gRPC, consistent with HTTP example |
| **SQL** | Go + PostgreSQL (pgx driver) | Realistic queries, migrations, connection pooling |
| **Redis** | Go + Redis (go-redis) | Common caching patterns, pub/sub |
| **Kafka** | Go + Kafka (confluent-kafka-go) | Producer/consumer patterns, partitions |

**Why Go?**:
- Single language reduces complexity
- Excellent performance
- Great libraries for all protocols
- Easy container builds
- Strong typing

---

## Directory Structure

```
mop/
├── examples/
│   ├── 01-http-api/           # HTTP REST API
│   │   ├── cmd/
│   │   │   └── server/
│   │   ├── internal/
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── models/
│   │   ├── tests/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   ├── 02-grpc-service/       # gRPC microservice
│   │   ├── cmd/
│   │   │   ├── server/
│   │   │   └── client/
│   │   ├── internal/
│   │   │   └── service/
│   │   ├── proto/
│   │   ├── tests/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   ├── 03-sql-app/            # Database-backed application
│   │   ├── cmd/
│   │   │   └── server/
│   │   ├── internal/
│   │   │   ├── db/
│   │   │   ├── repository/
│   │   │   └── models/
│   │   ├── migrations/
│   │   ├── tests/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   ├── 04-redis-cache/        # Redis caching layer
│   │   ├── cmd/
│   │   │   └── server/
│   │   ├── internal/
│   │   │   ├── cache/
│   │   │   └── handlers/
│   │   ├── tests/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   ├── 05-kafka-streaming/    # Kafka event streaming
│   │   ├── cmd/
│   │   │   ├── producer/
│   │   │   └── consumer/
│   │   ├── internal/
│   │   │   ├── events/
│   │   │   └── handlers/
│   │   ├── tests/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   └── load-generators/       # Traffic generators
│       ├── http-load/
│       ├── grpc-load/
│       ├── sql-load/
│       ├── redis-load/
│       └── kafka-load/
├── deployments/examples/      # Kubernetes manifests
│   ├── 01-http-api/
│   ├── 02-grpc-service/
│   ├── 03-sql-app/
│   ├── 04-redis-cache/
│   └── 05-kafka-streaming/
├── lib/grafana/dashboards/examples/  # Protocol dashboards
│   ├── http-api-dashboard.json
│   ├── grpc-service-dashboard.json
│   ├── sql-app-dashboard.json
│   ├── redis-cache-dashboard.json
│   └── kafka-streaming-dashboard.json
└── docs/examples/
    ├── http-instrumentation.md
    ├── grpc-instrumentation.md
    ├── sql-instrumentation.md
    ├── redis-instrumentation.md
    ├── kafka-instrumentation.md
    └── load-testing-guide.md
```

---

## Workstream Organization

### Phase 1: Foundation & HTTP (Parallel Possible)

**WS-REF-01: HTTP REST API** (Independent)
- Priority: Critical
- Duration: 6-8 hours
- Can start immediately
- No dependencies

**WS-REF-02: gRPC Service** (Independent)
- Priority: High
- Duration: 6-8 hours
- Can start immediately
- No dependencies

### Phase 2: Data Services (Parallel Possible)

**WS-REF-03: SQL Application** (Depends on HTTP for patterns)
- Priority: High
- Duration: 8-10 hours
- Can start after WS-REF-01 (optional dependency)
- Uses similar Go patterns

**WS-REF-04: Redis Cache** (Independent)
- Priority: Medium
- Duration: 6-8 hours
- Can start immediately or after WS-REF-01
- No hard dependencies

### Phase 3: Event Streaming (After Phase 1-2)

**WS-REF-05: Kafka Streaming** (Depends on patterns)
- Priority: Medium
- Duration: 8-10 hours
- Better after WS-REF-01 (for Go patterns)
- No hard technical dependencies

### Phase 4: Load Testing & Validation (After Phase 1-3)

**WS-REF-06: Load Generators** (Depends on all apps)
- Priority: High
- Duration: 6-8 hours
- Requires all applications complete
- Creates traffic for validation

**WS-REF-07: Dashboards & Documentation** (Depends on all apps)
- Priority: High
- Duration: 4-6 hours
- Requires all applications deployed
- Creates unified observability experience

---

## Detailed Workstream Breakdown

### WS-REF-01: HTTP REST API

**Application**: Product catalog service with REST endpoints

**Features**:
- CRUD operations for products
- Search with pagination
- Rate limiting
- Error scenarios (404, 500, timeout)
- Slow endpoint simulation (>1s latency)

**Endpoints**:
```
GET    /products       - List products (paginated)
GET    /products/:id   - Get product by ID
POST   /products       - Create product
PUT    /products/:id   - Update product
DELETE /products/:id   - Delete product
GET    /search         - Search products
GET    /health         - Health check
GET    /slow           - Slow endpoint (testing)
GET    /error          - Error endpoint (testing)
```

**Issues**:
1. **REF-01-001**: Project setup (Go module, directory structure, Dockerfile)
2. **REF-01-002**: Implement HTTP handlers with Gin framework
3. **REF-01-003**: Add middleware (logging, metrics, error handling)
4. **REF-01-004**: Create Kubernetes deployment manifests
5. **REF-01-005**: Add integration tests
6. **REF-01-006**: Create Grafana dashboard for HTTP metrics
7. **REF-01-007**: Write documentation and OBI instrumentation guide

---

### WS-REF-02: gRPC Service

**Application**: User authentication service with gRPC

**Features**:
- User login/logout
- Token validation
- Session management
- Streaming endpoint (server-side)
- Error handling with gRPC status codes

**RPCs**:
```proto
service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc ValidateToken(ValidateRequest) returns (ValidateResponse);
  rpc RefreshToken(RefreshRequest) returns (RefreshResponse);
  rpc StreamEvents(EventsRequest) returns (stream Event); // Server streaming
}
```

**Issues**:
1. **REF-02-001**: Project setup and protobuf definitions
2. **REF-02-002**: Implement gRPC server and handlers
3. **REF-02-003**: Create gRPC client for testing
4. **REF-02-004**: Add interceptors (auth, logging, tracing)
5. **REF-02-005**: Create Kubernetes deployment manifests
6. **REF-02-006**: Add integration tests with gRPC client
7. **REF-02-007**: Create Grafana dashboard for gRPC metrics
8. **REF-02-008**: Write documentation and OBI instrumentation guide

---

### WS-REF-03: SQL Application

**Application**: Order management system with PostgreSQL

**Features**:
- Order CRUD operations
- Complex queries (joins, aggregations)
- Database migrations
- Connection pooling
- Slow query simulation (N+1 problem)
- Transaction management

**Database Schema**:
```sql
orders (id, customer_id, status, total, created_at)
order_items (id, order_id, product_id, quantity, price)
customers (id, name, email, created_at)
```

**Issues**:
1. **REF-03-001**: Project setup with PostgreSQL integration
2. **REF-03-002**: Create database migrations (golang-migrate)
3. **REF-03-003**: Implement repository pattern with pgx
4. **REF-03-004**: Add HTTP handlers for order operations
5. **REF-03-005**: Create Kubernetes deployment (app + PostgreSQL)
6. **REF-03-006**: Add integration tests with test database
7. **REF-03-007**: Create Grafana dashboard for SQL metrics
8. **REF-03-008**: Write documentation on slow query detection

---

### WS-REF-04: Redis Cache

**Application**: API gateway with Redis caching

**Features**:
- Cache-aside pattern
- Cache invalidation
- TTL management
- Redis pub/sub for cache updates
- Cache hit/miss metrics
- Fallback to upstream on cache miss

**Redis Operations**:
- GET/SET/DEL for caching
- Pub/Sub for invalidation
- EXPIRE for TTL
- Pipeline for bulk operations

**Issues**:
1. **REF-04-001**: Project setup with Redis integration
2. **REF-04-002**: Implement cache layer with go-redis
3. **REF-04-003**: Add pub/sub for cache invalidation
4. **REF-04-004**: Create HTTP handlers with caching
5. **REF-04-005**: Create Kubernetes deployment (app + Redis)
6. **REF-04-006**: Add integration tests with test Redis
7. **REF-04-007**: Create Grafana dashboard for Redis metrics
8. **REF-04-008**: Write documentation on cache patterns

---

### WS-REF-05: Kafka Streaming

**Application**: Event processing pipeline with Kafka

**Features**:
- Producer: Order events (created, updated, cancelled)
- Consumer: Email notifications, analytics
- Consumer groups
- Partition handling
- Error handling and retries
- Dead letter queue pattern

**Topics**:
```
orders.created   - New order events
orders.updated   - Order updates
orders.cancelled - Cancellations
notifications    - Email notifications
analytics        - Analytics events
dlq              - Dead letter queue
```

**Issues**:
1. **REF-05-001**: Project setup with Kafka integration
2. **REF-05-002**: Implement Kafka producer for order events
3. **REF-05-003**: Implement Kafka consumers (notifications, analytics)
4. **REF-05-004**: Add consumer group coordination
5. **REF-05-005**: Create Kubernetes deployment (app + Kafka)
6. **REF-05-006**: Add integration tests with test Kafka
7. **REF-05-007**: Create Grafana dashboard for Kafka metrics
8. **REF-05-008**: Write documentation on event streaming patterns

---

### WS-REF-06: Load Generators

**Purpose**: Generate realistic traffic for all applications

**Load Patterns**:
- **Constant Load**: Steady RPS for baseline testing
- **Spike Load**: Sudden traffic increase
- **Gradual Load**: Slowly increasing RPS
- **Mixed Scenarios**: Combination of success/error cases

**Tools**: Custom Go programs using:
- `net/http` for HTTP load
- gRPC client for gRPC load
- PostgreSQL client for SQL load
- Redis client for Redis load
- Kafka producer for Kafka load

**Issues**:
1. **REF-06-001**: HTTP load generator with configurable patterns
2. **REF-06-002**: gRPC load generator with streaming support
3. **REF-06-003**: SQL load generator with query patterns
4. **REF-06-004**: Redis load generator with cache patterns
5. **REF-06-005**: Kafka load generator with event patterns
6. **REF-06-006**: Kubernetes CronJobs for scheduled load tests
7. **REF-06-007**: Load testing documentation and runbooks

---

### WS-REF-07: Dashboards & Documentation

**Purpose**: Unified observability experience across all protocols

**Dashboards**:
1. **Overview Dashboard**: All protocols at a glance
2. **HTTP Dashboard**: Request rates, latency, errors, endpoints
3. **gRPC Dashboard**: RPC calls, streaming, status codes
4. **SQL Dashboard**: Query duration, N+1 queries, slow queries
5. **Redis Dashboard**: Cache hit/miss, operations, pub/sub
6. **Kafka Dashboard**: Producer/consumer metrics, lag, partitions

**Documentation**:
1. **OBI Instrumentation Guide**: How OBI captures each protocol
2. **Troubleshooting Guide**: Using OBI to debug issues
3. **Dashboard Guide**: Understanding each dashboard
4. **Load Testing Guide**: Running and interpreting load tests
5. **Best Practices**: Recommendations for production

**Issues**:
1. **REF-07-001**: Create overview dashboard
2. **REF-07-002**: Create protocol-specific dashboards (HTTP, gRPC, SQL, Redis, Kafka)
3. **REF-07-003**: Write OBI instrumentation guide for each protocol
4. **REF-07-004**: Write troubleshooting guide with scenarios
5. **REF-07-005**: Write load testing guide
6. **REF-07-006**: Create demo video/screencast (optional)
7. **REF-07-007**: Compile best practices documentation

---

## Parallel Execution Strategy

### Wave 1: Foundation (All Parallel)
**Duration**: 6-8 hours
- WS-REF-01: HTTP REST API ✅ Independent
- WS-REF-02: gRPC Service ✅ Independent
- WS-REF-04: Redis Cache ✅ Independent (can reference HTTP patterns)

### Wave 2: Data Services (Parallel)
**Duration**: 8-10 hours
- WS-REF-03: SQL Application (after WS-REF-01 for patterns)
- WS-REF-05: Kafka Streaming (after WS-REF-01 for patterns)

### Wave 3: Validation (Sequential dependencies)
**Duration**: 10-12 hours
- WS-REF-06: Load Generators (after all apps complete)
- WS-REF-07: Dashboards & Docs (after load testing)

**Total Timeline**:
- **Parallel Execution**: ~24-30 hours
- **Sequential Execution**: ~50-60 hours
- **Speedup**: ~2x

---

## Issue Template Structure

Each issue follows this structure:

```markdown
# [ISSUE-ID]: [Title]

## Context
Brief description of what this issue accomplishes.

## Requirements
- [ ] Requirement 1
- [ ] Requirement 2
- [ ] Requirement 3

## Acceptance Criteria
1. Criterion 1
2. Criterion 2
3. Criterion 3

## Clarifying Questions

**Q1: [Question]**
- Context: [Why this matters]
- Options: [Choices available]
- Default: [Suggested default]

**Q2: [Question]**
...

## Technical Notes
- File paths to create/modify
- Dependencies to install
- Testing approach
- Integration points

## Definition of Done
- [ ] Code implemented and tested
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Documentation updated
- [ ] Kubernetes manifests created
- [ ] Code reviewed
- [ ] PR merged

## Related Issues
- Blocks: #XXX
- Blocked by: #XXX
- Related: #XXX
```

---

## Success Metrics

### Technical Metrics
- ✅ All 5 protocol examples implemented
- ✅ Zero code changes for instrumentation (OBI only)
- ✅ < 1% CPU overhead from OBI
- ✅ 100% trace collection
- ✅ All load tests passing
- ✅ All dashboards functional

### Quality Metrics
- ✅ 80%+ test coverage per application
- ✅ All linting passing (golangci-lint)
- ✅ All builds successful
- ✅ Documentation complete for each protocol
- ✅ No security vulnerabilities (gosec)

### Observability Metrics
- ✅ Traces captured for all protocols
- ✅ Metrics generated automatically
- ✅ Logs correlated with traces
- ✅ Dashboards show real-time data
- ✅ Slow queries detected (SQL)
- ✅ Cache patterns visible (Redis)
- ✅ Event flow visible (Kafka)

---

## Risk Mitigation

### Risk 1: Protocol Complexity
**Mitigation**: Start with HTTP (simplest), validate OBI works, then proceed to complex protocols

### Risk 2: Go Version Compatibility
**Mitigation**: Use Go 1.21+ for all examples, test eBPF compatibility

### Risk 3: OBI Coverage Gaps
**Mitigation**: Validate each protocol with OBI in dev environment first

### Risk 4: Load Testing Impact
**Mitigation**: Use separate namespaces, resource limits, and time-limited tests

### Risk 5: Documentation Quality
**Mitigation**: Have each developer write docs as they build, not after

---

## Next Steps

1. **Review & Approve Plan**: Get feedback on workstream organization
2. **Create GitHub Issues**: Generate 50+ granular issues from this plan
3. **Set Up GitHub Actions**: Event-driven orchestration workflows
4. **Launch Orchestrator**: Begin parallel execution with development agents
5. **Monitor Progress**: Track via GitHub Projects board
6. **Validate Integration**: E2E testing across all protocols
7. **Document Learnings**: Capture OBI instrumentation insights

---

**Estimated Completion**: 24-30 hours with 3-5 concurrent agents
**Speedup vs Sequential**: ~2x faster
**Total Issues**: ~50 issues across 7 workstreams

---

**Status**: 🟡 Ready for Review
**Next**: Create GitHub issues and set up orchestration
