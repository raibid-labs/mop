# Workstream 7: Reference Implementations - OBI Protocol Showcase

## Status
🟡 Planning Phase

## Overview
Create comprehensive reference implementations demonstrating OBI's automatic eBPF instrumentation across all supported protocols (HTTP/HTTPS, gRPC, SQL, Redis, Kafka). Each implementation provides realistic, production-grade examples with load generators, Grafana dashboards, and troubleshooting scenarios to showcase zero-code observability.

## Objectives
- [ ] Build reference applications for all 5 OBI-supported protocols
- [ ] Create realistic traffic patterns with load generators
- [ ] Demonstrate zero-code instrumentation (no SDK/agent installation)
- [ ] Provide protocol-specific Grafana dashboards
- [ ] Include troubleshooting scenarios (slow queries, errors, resource exhaustion)
- [ ] Document OBI instrumentation patterns and telemetry correlation
- [ ] Enable self-service exploration of OBI capabilities

## Strategic Value
**Impact**: 🟢 High - Demonstrates OBI capabilities, accelerates adoption, provides testing harness
**Complexity**: 🟡 Medium - Multiple languages/frameworks, but independent workstreams
**Timeline**: 6-8 weeks (2 weeks foundation + 4 weeks parallel development + 2 weeks integration)

---

## Technology Stack Decisions

### Protocol Implementation Choices

| Protocol | Language | Framework | Rationale |
|----------|----------|-----------|-----------|
| **HTTP/REST** | Go | Echo | Fast, popular, generates diverse HTTP patterns |
| **gRPC** | Go | grpc-go | Native gRPC support, high performance |
| **SQL** | Python | FastAPI + SQLAlchemy | Common ORM patterns, connection pooling |
| **Redis** | Node.js | Express + ioredis | Common caching patterns, session storage |
| **Kafka** | Java | Spring Boot + Kafka | Enterprise standard, rich Kafka ecosystem |

### Supporting Technologies

- **Load Generation**: k6 (programmable, Grafana-native)
- **Database**: PostgreSQL 16 (advanced query features)
- **Message Broker**: Apache Kafka 3.6
- **Caching**: Redis 7.2
- **Container Base Images**: Distroless (minimal attack surface)
- **Kubernetes**: Deployment, Service, ConfigMap, HPA

---

## Directory Structure

```
mop/
├── examples/                              # Reference implementations
│   ├── 00-shared/                         # Shared utilities
│   │   ├── k6/                            # Load testing scripts
│   │   │   ├── http-load.js
│   │   │   ├── grpc-load.js
│   │   │   ├── sql-load.js
│   │   │   ├── redis-load.js
│   │   │   └── kafka-load.js
│   │   ├── scenarios/                     # Troubleshooting scenarios
│   │   │   ├── slow-query.yaml
│   │   │   ├── error-spike.yaml
│   │   │   ├── connection-leak.yaml
│   │   │   └── resource-exhaustion.yaml
│   │   └── lib/                           # Shared libraries
│   │       ├── tracing-helpers.go
│   │       └── metrics-helpers.go
│   │
│   ├── 01-http-rest-api/                  # HTTP/HTTPS example
│   │   ├── cmd/                           # Application entrypoint
│   │   │   └── server/
│   │   │       └── main.go
│   │   ├── internal/                      # Internal packages
│   │   │   ├── handlers/
│   │   │   ├── middleware/
│   │   │   └── models/
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   ├── go.sum
│   │   └── README.md
│   │
│   ├── 02-grpc-microservice/              # gRPC example
│   │   ├── cmd/
│   │   │   ├── server/main.go
│   │   │   └── client/main.go             # Sample gRPC client
│   │   ├── internal/
│   │   │   ├── service/
│   │   │   └── interceptors/
│   │   ├── proto/                         # Protobuf definitions
│   │   │   ├── user.proto
│   │   │   └── product.proto
│   │   ├── Dockerfile
│   │   ├── go.mod
│   │   └── README.md
│   │
│   ├── 03-sql-application/                # SQL (PostgreSQL) example
│   │   ├── app/
│   │   │   ├── main.py
│   │   │   ├── models.py                  # SQLAlchemy models
│   │   │   ├── routes.py
│   │   │   └── database.py                # Connection pooling
│   │   ├── migrations/                    # Alembic migrations
│   │   │   └── versions/
│   │   ├── Dockerfile
│   │   ├── requirements.txt
│   │   └── README.md
│   │
│   ├── 04-redis-cache/                    # Redis example
│   │   ├── src/
│   │   │   ├── index.js                   # Express app
│   │   │   ├── routes/
│   │   │   ├── cache/                     # Caching layer
│   │   │   └── session/                   # Session management
│   │   ├── Dockerfile
│   │   ├── package.json
│   │   └── README.md
│   │
│   └── 05-kafka-streaming/                # Kafka example
│       ├── src/
│       │   └── main/
│       │       └── java/com/mop/kafka/
│       │           ├── KafkaApplication.java
│       │           ├── producer/
│       │           │   └── EventProducer.java
│       │           ├── consumer/
│       │           │   └── EventConsumer.java
│       │           └── config/
│       ├── Dockerfile
│       ├── pom.xml
│       └── README.md
│
├── deployments/examples/                  # Kubernetes manifests
│   ├── http-rest-api/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   ├── configmap.yaml
│   │   ├── hpa.yaml                       # Horizontal Pod Autoscaler
│   │   └── kustomization.yaml
│   ├── grpc-microservice/
│   ├── sql-application/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── postgres-statefulset.yaml      # PostgreSQL database
│   ├── redis-cache/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── redis-statefulset.yaml         # Redis instance
│   ├── kafka-streaming/
│   │   ├── deployment.yaml
│   │   ├── service.yaml
│   │   └── kafka-cluster.yaml             # Kafka + Zookeeper
│   └── load-generators/
│       ├── k6-http-load.yaml
│       ├── k6-grpc-load.yaml
│       ├── k6-sql-load.yaml
│       ├── k6-redis-load.yaml
│       └── k6-kafka-load.yaml
│
├── dashboards/examples/                   # Grafana dashboards
│   ├── http-rest-api-dashboard.json
│   ├── grpc-microservice-dashboard.json
│   ├── sql-application-dashboard.json
│   ├── redis-cache-dashboard.json
│   └── kafka-streaming-dashboard.json
│
└── docs/examples/                         # Documentation
    ├── README.md                          # Overview and index
    ├── http-rest-api.md                   # HTTP example guide
    ├── grpc-microservice.md               # gRPC example guide
    ├── sql-application.md                 # SQL example guide
    ├── redis-cache.md                     # Redis example guide
    ├── kafka-streaming.md                 # Kafka example guide
    ├── obi-instrumentation-patterns.md    # OBI patterns across protocols
    ├── troubleshooting-scenarios.md       # How to use trouble scenarios
    └── load-testing-guide.md              # k6 load testing guide
```

---

## Parallel Workstream Organization

### Level 1: Parallel Workstreams (Independent Execution)

```
┌─────────────────────────────────────────────────────────────────┐
│                    Foundation Workstream                        │
│  (Must Complete First - Week 1-2)                              │
│  - Shared libraries, k6 scripts, documentation templates       │
└─────────────────────────────────────────────────────────────────┘
                              ↓
        ┌─────────────────────┬─────────────────────┬─────────────────────┐
        ↓                     ↓                     ↓                     ↓
┌────────────────┐  ┌────────────────┐  ┌────────────────┐  ┌────────────────┐
│ HTTP Workstream│  │ gRPC Workstream│  │ SQL Workstream │  │ Data Workstream│
│  (Week 3-4)    │  │  (Week 3-4)    │  │  (Week 3-4)    │  │  (Week 3-4)    │
│                │  │                │  │                │  │  - Redis       │
│  - REST API    │  │  - Microservice│  │  - FastAPI     │  │  - Kafka       │
│  - Echo        │  │  - grpc-go     │  │  - PostgreSQL  │  │                │
│  - HTTP/2      │  │  - Protobuf    │  │  - SQLAlchemy  │  │                │
└────────────────┘  └────────────────┘  └────────────────┘  └────────────────┘
        │                     │                     │                     │
        └─────────────────────┴─────────────────────┴─────────────────────┘
                                      ↓
                        ┌─────────────────────────┐
                        │ Integration Workstream  │
                        │  (Week 5-6)             │
                        │  - E2E tests            │
                        │  - Multi-protocol flows │
                        │  - Documentation polish │
                        └─────────────────────────┘
```

**Rationale**:
- Foundation creates shared utilities (k6 scripts, helpers)
- 4 parallel workstreams execute independently (no blocking dependencies)
- Redis + Kafka grouped as "Data Workstream" (similar patterns)
- Integration phase validates cross-protocol scenarios

---

## Granular Issue Breakdown

### Foundation Workstream (WS-7.0)

**Duration**: 2 weeks | **Agent**: `system-architect`, `planner`

#### Issue 7.0.1: Shared Load Testing Framework
**Description**: Create reusable k6 load testing scripts for all protocols.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/00-shared/k6/http-load.js`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/k6/grpc-load.js`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/k6/sql-load.js`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/k6/redis-load.js`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/k6/kafka-load.js`
- `/Users/beengud/raibid-labs/mop/docs/examples/load-testing-guide.md`

**Acceptance Criteria**:
- [ ] k6 scripts parameterized via environment variables
- [ ] Configurable load profiles (constant, ramp, spike, stress)
- [ ] Custom metrics tracked (business transactions, not just requests)
- [ ] HTML report generation
- [ ] Tested against sample endpoints

#### Issue 7.0.2: Troubleshooting Scenario Library
**Description**: Define reproducible troubleshooting scenarios.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/00-shared/scenarios/slow-query.yaml`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/scenarios/error-spike.yaml`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/scenarios/connection-leak.yaml`
- `/Users/beengud/raibid-labs/mop/examples/00-shared/scenarios/resource-exhaustion.yaml`
- `/Users/beengud/raibid-labs/mop/docs/examples/troubleshooting-scenarios.md`

**Acceptance Criteria**:
- [ ] Scenarios trigger via ConfigMap changes (no code deploy)
- [ ] Observable via OBI telemetry (traces show root cause)
- [ ] Documented resolution steps
- [ ] Include "smoking gun" metrics to look for

#### Issue 7.0.3: Documentation Templates
**Description**: Create standardized documentation templates.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/docs/examples/README.md` (master index)
- `/Users/beengud/raibid-labs/mop/docs/examples/_TEMPLATE.md` (example guide template)
- `/Users/beengud/raibid-labs/mop/docs/examples/obi-instrumentation-patterns.md`

**Acceptance Criteria**:
- [ ] Template includes: Overview, Architecture, OBI Telemetry, Deployment, Troubleshooting
- [ ] OBI patterns documented for each protocol
- [ ] Links to relevant OBI configuration

---

### HTTP/REST Workstream (WS-7.1)

**Duration**: 2 weeks | **Agent**: `backend-dev` (Go specialist)

#### Issue 7.1.1: Go Echo REST API Implementation
**Description**: Build REST API with Echo framework demonstrating common patterns.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/01-http-rest-api/` (full application)
- Endpoints: CRUD operations, file uploads, authentication, rate limiting
- Middleware: Logging, recovery, CORS, request ID propagation
- `/Users/beengud/raibid-labs/mop/examples/01-http-rest-api/README.md`

**Acceptance Criteria**:
- [ ] 10+ RESTful endpoints (GET, POST, PUT, DELETE)
- [ ] JSON request/response handling
- [ ] Error handling with proper HTTP status codes
- [ ] Context propagation for tracing
- [ ] Unit tests with >80% coverage

#### Issue 7.1.2: HTTP Kubernetes Deployment
**Description**: Create production-ready Kubernetes manifests.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/deployments/examples/http-rest-api/` (all manifests)
- Deployment with resource limits, health checks, rolling updates
- Service (ClusterIP and optional Ingress)
- ConfigMap for environment variables
- HPA (Horizontal Pod Autoscaler)

**Acceptance Criteria**:
- [ ] Liveness and readiness probes configured
- [ ] Resource requests/limits set
- [ ] HPA scales 1-10 replicas on CPU >70%
- [ ] Deployment tested on dev cluster

#### Issue 7.1.3: HTTP Load Generator
**Description**: k6 load generator for HTTP endpoints.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/deployments/examples/load-generators/k6-http-load.yaml`
- Kubernetes Job manifest for k6
- Configurable via ConfigMap (RPS, duration, virtual users)

**Acceptance Criteria**:
- [ ] Tests all CRUD endpoints
- [ ] Generates realistic traffic distribution
- [ ] Exports results to Prometheus
- [ ] Runs as Kubernetes Job

#### Issue 7.1.4: HTTP Grafana Dashboard
**Description**: Protocol-specific dashboard visualizing OBI telemetry.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/dashboards/examples/http-rest-api-dashboard.json`

**Panels**:
- Request rate (by endpoint, status code)
- Latency percentiles (p50, p90, p95, p99)
- Error rate
- Top slowest endpoints
- HTTP method distribution
- Request/response size distribution
- Trace exemplars (click to view in Tempo)

**Acceptance Criteria**:
- [ ] All metrics sourced from OBI (not application instrumentation)
- [ ] Trace exemplars link to Tempo
- [ ] Variables for filtering by namespace, pod, endpoint

#### Issue 7.1.5: HTTP Integration Test
**Description**: End-to-end test validating OBI instrumentation.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/tests/examples/http-rest-api-test.sh`

**Test Steps**:
1. Deploy HTTP application
2. Generate traffic
3. Query Tempo for traces
4. Query Mimir for metrics
5. Validate trace structure (HTTP spans, attributes)
6. Validate metrics (http_server_requests_total, http_server_request_duration)

**Acceptance Criteria**:
- [ ] Test runs in CI/CD pipeline
- [ ] Validates trace presence and structure
- [ ] Validates metrics presence and labels
- [ ] Fails if telemetry missing

#### Issue 7.1.6: HTTP Documentation
**Description**: Complete guide for HTTP example.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/docs/examples/http-rest-api.md`

**Sections**:
- Architecture overview
- OBI instrumentation points
- Deployment instructions
- Running load tests
- Troubleshooting scenarios
- Expected telemetry (sample traces, metrics)

**Acceptance Criteria**:
- [ ] Follows documentation template
- [ ] Includes architecture diagram
- [ ] Screenshots of Grafana dashboard
- [ ] Tested by following instructions end-to-end

---

### gRPC Workstream (WS-7.2)

**Duration**: 2 weeks | **Agent**: `backend-dev` (Go/Protobuf specialist)

#### Issue 7.2.1: Go gRPC Microservice Implementation
**Description**: Build gRPC service with bidirectional streaming.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/02-grpc-microservice/` (full application)
- Services: Unary RPC, server streaming, client streaming, bidirectional streaming
- Protobuf definitions for User, Product domains
- Interceptors for logging, recovery
- Sample gRPC client for testing

**Acceptance Criteria**:
- [ ] 4+ gRPC methods (covering all RPC types)
- [ ] Protobuf definitions compiled
- [ ] gRPC health check service
- [ ] Context propagation for distributed tracing
- [ ] Unit tests for RPC handlers

#### Issue 7.2.2: gRPC Kubernetes Deployment
**Deliverables**: Similar to 7.1.2 but for gRPC (headless Service for discovery)

#### Issue 7.2.3: gRPC Load Generator
**Description**: k6 load generator with grpc-js extension.

**Deliverables**:
- k6 script with gRPC client
- Tests unary and streaming RPCs

#### Issue 7.2.4: gRPC Grafana Dashboard
**Panels**:
- RPC rate (by method, status)
- Latency percentiles
- gRPC status code distribution
- Streaming connection count
- Message size distribution

#### Issue 7.2.5: gRPC Integration Test
**Deliverables**: Similar to 7.1.5 but validates gRPC spans

#### Issue 7.2.6: gRPC Documentation
**Deliverables**: `/Users/beengud/raibid-labs/mop/docs/examples/grpc-microservice.md`

---

### SQL Workstream (WS-7.3)

**Duration**: 2 weeks | **Agent**: `backend-dev` (Python specialist)

#### Issue 7.3.1: Python FastAPI + SQLAlchemy Application
**Description**: Build REST API with PostgreSQL backend.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/03-sql-application/` (full application)
- FastAPI endpoints (CRUD for users, posts, comments)
- SQLAlchemy models with relationships
- Alembic migrations
- Connection pooling configuration
- Includes slow query scenario (missing index)

**Acceptance Criteria**:
- [ ] 8+ endpoints hitting PostgreSQL
- [ ] Models: User (1:many) -> Post (1:many) -> Comment
- [ ] N+1 query scenario for troubleshooting
- [ ] Connection pool metrics exposed
- [ ] Pytest tests with database fixtures

#### Issue 7.3.2: SQL Kubernetes Deployment
**Deliverables**:
- Application deployment
- PostgreSQL StatefulSet with persistent volume
- Database initialization Job (schema migration)

#### Issue 7.3.3: SQL Load Generator
**Description**: k6 script hitting CRUD endpoints to generate SQL queries.

**Deliverables**:
- Traffic pattern: 70% reads, 30% writes
- Triggers N+1 query scenario
- Includes slow query scenario

#### Issue 7.3.4: SQL Grafana Dashboard
**Panels**:
- Database query rate (by operation: SELECT, INSERT, UPDATE, DELETE)
- Query latency percentiles
- Slow query log (queries >100ms)
- Connection pool utilization
- Database table access patterns
- Top 10 slowest queries

#### Issue 7.3.5: SQL Integration Test
**Deliverables**: Validates SQL spans in traces, db.statement attribute

#### Issue 7.3.6: SQL Documentation
**Deliverables**: `/Users/beengud/raibid-labs/mop/docs/examples/sql-application.md`

---

### Data Workstream (WS-7.4: Redis + Kafka)

**Duration**: 2 weeks | **Agents**: `backend-dev` (Node.js), `backend-dev` (Java)

#### Issue 7.4.1: Node.js Express + Redis Application
**Description**: REST API with Redis caching and session management.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/04-redis-cache/` (full application)
- Endpoints: User sessions, rate limiting, caching layer
- Redis patterns: GET/SET, INCR, EXPIRE, pub/sub
- Cache miss/hit scenario

**Acceptance Criteria**:
- [ ] 6+ endpoints using Redis
- [ ] Cache-aside pattern implementation
- [ ] Session storage with TTL
- [ ] Rate limiting with sliding window
- [ ] Jest tests with Redis mock

#### Issue 7.4.2: Redis Kubernetes Deployment
**Deliverables**:
- Application deployment
- Redis StatefulSet (single instance for simplicity)

#### Issue 7.4.3: Redis Load Generator
**Description**: k6 script generating cache hits/misses.

#### Issue 7.4.4: Redis Grafana Dashboard
**Panels**:
- Redis operation rate (by command)
- Command latency
- Cache hit ratio
- Key distribution
- Connection count

#### Issue 7.4.5: Java Spring Boot + Kafka Application
**Description**: Event streaming with producer/consumer.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/05-kafka-streaming/` (full application)
- Producer: Publishes events to topics
- Consumer: Consumes events with consumer groups
- Topics: user-events, order-events
- Includes DLQ (Dead Letter Queue) pattern

**Acceptance Criteria**:
- [ ] 3+ topics with different partition counts
- [ ] Producer with batching, compression
- [ ] Consumer with offset management
- [ ] Error handling with retry logic
- [ ] JUnit tests with embedded Kafka

#### Issue 7.4.6: Kafka Kubernetes Deployment
**Deliverables**:
- Application deployment
- Kafka cluster (3 brokers via StatefulSet)
- Zookeeper ensemble (3 nodes)

#### Issue 7.4.7: Kafka Load Generator
**Description**: k6 script producing events to Kafka.

#### Issue 7.4.8: Kafka Grafana Dashboard
**Panels**:
- Producer throughput (messages/sec)
- Consumer lag
- Topic partition distribution
- Broker performance
- Message size distribution

#### Issue 7.4.9: Redis Integration Test
**Deliverables**: Validates Redis spans

#### Issue 7.4.10: Kafka Integration Test
**Deliverables**: Validates Kafka producer/consumer spans

#### Issue 7.4.11: Redis Documentation
**Deliverables**: `/Users/beengud/raibid-labs/mop/docs/examples/redis-cache.md`

#### Issue 7.4.12: Kafka Documentation
**Deliverables**: `/Users/beengud/raibid-labs/mop/docs/examples/kafka-streaming.md`

---

### Integration Workstream (WS-7.5)

**Duration**: 2 weeks | **Agent**: `tester`, `system-architect`

#### Issue 7.5.1: Multi-Protocol End-to-End Flow
**Description**: Create scenario combining all protocols.

**Example Flow**:
1. HTTP request to REST API
2. REST API calls gRPC service
3. gRPC service queries PostgreSQL
4. gRPC service checks Redis cache
5. gRPC service publishes Kafka event

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/examples/06-multi-protocol-flow/` (orchestration)
- Integration test validating distributed trace across all protocols

**Acceptance Criteria**:
- [ ] Single trace spans all 5 protocols
- [ ] Trace correctly correlates all services
- [ ] Demonstrates OBI's protocol coverage

#### Issue 7.5.2: Comprehensive Troubleshooting Guide
**Description**: Walkthrough of common issues using examples.

**Deliverables**:
- `/Users/beengud/raibid-labs/mop/docs/examples/troubleshooting-scenarios.md`

**Scenarios**:
1. Slow SQL query (missing index)
2. HTTP 5xx error spike
3. Redis cache stampede
4. Kafka consumer lag
5. gRPC deadline exceeded

**Acceptance Criteria**:
- [ ] Each scenario has step-by-step diagnosis using OBI telemetry
- [ ] Screenshots from Grafana
- [ ] Resolution documented

#### Issue 7.5.3: Documentation Polish
**Description**: Review all documentation for consistency.

**Deliverables**:
- Updated `/Users/beengud/raibid-labs/mop/docs/examples/README.md` (index)
- Consistent formatting across all guides
- Working links

#### Issue 7.5.4: CI/CD Pipeline
**Description**: Automate building, testing, deploying examples.

**Deliverables**:
- GitHub Actions workflow
- Builds all Docker images
- Runs integration tests
- Deploys to dev environment

---

## Detailed Application Specifications

### HTTP REST API (Go + Echo)

**Endpoints**:
- `GET /api/v1/users` - List users (pagination)
- `GET /api/v1/users/:id` - Get user by ID
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user
- `POST /api/v1/users/:id/avatar` - Upload user avatar (multipart)
- `GET /api/v1/health` - Health check
- `GET /api/v1/metrics` - Prometheus metrics (app-level)
- `POST /api/v1/auth/login` - Authentication (JWT)
- `POST /api/v1/auth/refresh` - Token refresh

**Middleware**:
- Request ID injection
- Structured logging
- Panic recovery
- CORS
- Rate limiting (per IP)
- JWT validation

**Data Storage**: In-memory map (no database for simplicity)

**OBI Telemetry Generated**:
- HTTP server spans (`http.method`, `http.route`, `http.status_code`)
- Metrics: `http_server_requests_total`, `http_server_request_duration_seconds`

---

### gRPC Microservice (Go + grpc-go)

**Services**:

```protobuf
service UserService {
  rpc GetUser(GetUserRequest) returns (User);                   // Unary
  rpc ListUsers(ListUsersRequest) returns (stream User);        // Server streaming
  rpc CreateUsers(stream User) returns (CreateUsersResponse);   // Client streaming
  rpc Chat(stream ChatMessage) returns (stream ChatMessage);    // Bidirectional
}

service ProductService {
  rpc GetProduct(GetProductRequest) returns (Product);
  rpc SearchProducts(SearchRequest) returns (stream Product);
}
```

**Features**:
- gRPC health check service
- gRPC reflection (for debugging)
- Interceptors for logging, metrics, recovery
- TLS support (optional)

**OBI Telemetry Generated**:
- gRPC spans (`rpc.service`, `rpc.method`, `rpc.grpc.status_code`)
- Metrics: `grpc_server_requests_total`, `grpc_server_request_duration_seconds`

---

### SQL Application (Python + FastAPI + PostgreSQL)

**Endpoints**:
- `GET /api/v1/posts` - List posts (with pagination, filtering)
- `GET /api/v1/posts/:id` - Get post with comments (N+1 query scenario)
- `POST /api/v1/posts` - Create post
- `PUT /api/v1/posts/:id` - Update post
- `DELETE /api/v1/posts/:id` - Delete post
- `GET /api/v1/users/:id/posts` - Get user's posts
- `POST /api/v1/posts/:id/comments` - Add comment to post

**Database Schema**:
```sql
CREATE TABLE users (
  id SERIAL PRIMARY KEY,
  username VARCHAR(50) UNIQUE NOT NULL,
  email VARCHAR(100) UNIQUE NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE posts (
  id SERIAL PRIMARY KEY,
  user_id INTEGER REFERENCES users(id),
  title VARCHAR(200) NOT NULL,
  content TEXT,
  created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE comments (
  id SERIAL PRIMARY KEY,
  post_id INTEGER REFERENCES posts(id),
  user_id INTEGER REFERENCES users(id),
  content TEXT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
);

-- Missing index on posts.user_id for slow query scenario
```

**SQLAlchemy Configuration**:
- Connection pooling (pool size: 20)
- Statement timeout: 30s
- Echo mode for logging (dev only)

**OBI Telemetry Generated**:
- Database spans (`db.system=postgresql`, `db.statement`, `db.operation`)
- Metrics: `db_query_duration_seconds`, `db_query_total`

---

### Redis Cache (Node.js + Express)

**Endpoints**:
- `POST /api/v1/sessions` - Create session (Redis SET with TTL)
- `GET /api/v1/sessions/:id` - Get session (Redis GET)
- `DELETE /api/v1/sessions/:id` - Delete session (Redis DEL)
- `GET /api/v1/products/:id` - Get product (cache-aside pattern)
- `POST /api/v1/rate-limit/check` - Check rate limit (Redis INCR)
- `GET /api/v1/leaderboard` - Get leaderboard (Redis ZRANGE)

**Redis Patterns Demonstrated**:
- Cache-aside (check cache, fallback to DB, populate cache)
- Write-through cache
- TTL-based expiration
- Pub/Sub for invalidation
- Sorted sets for leaderboard

**OBI Telemetry Generated**:
- Redis spans (`db.system=redis`, `db.operation=GET/SET/DEL`)
- Metrics: `redis_command_duration_seconds`, `redis_command_total`

---

### Kafka Streaming (Java + Spring Boot)

**Topics**:
- `user-events` (partitions: 3)
- `order-events` (partitions: 5)
- `notification-events` (partitions: 1)
- `dlq-events` (Dead Letter Queue)

**Producer Configuration**:
- Compression: snappy
- Batching: linger.ms=10
- Idempotence: true
- Retries: 3

**Consumer Configuration**:
- Consumer group: `mop-consumer-group`
- Auto-offset reset: earliest
- Max poll records: 500
- Session timeout: 30s

**Event Types**:
```java
UserCreatedEvent {
  userId: String
  username: String
  email: String
  timestamp: Instant
}

OrderPlacedEvent {
  orderId: String
  userId: String
  items: List<OrderItem>
  totalAmount: BigDecimal
  timestamp: Instant
}
```

**OBI Telemetry Generated**:
- Kafka producer spans (`messaging.system=kafka`, `messaging.destination=topic`)
- Kafka consumer spans
- Metrics: `kafka_producer_requests_total`, `kafka_consumer_lag`

---

## Load Testing Specifications

### k6 Load Profiles

**Constant Load**:
```javascript
export let options = {
  stages: [
    { duration: '5m', target: 100 }, // Ramp to 100 VUs
    { duration: '10m', target: 100 }, // Stay at 100 VUs
    { duration: '5m', target: 0 }, // Ramp down
  ],
};
```

**Spike Test**:
```javascript
export let options = {
  stages: [
    { duration: '1m', target: 100 },
    { duration: '10s', target: 1000 }, // Spike!
    { duration: '1m', target: 100 },
  ],
};
```

**Stress Test**:
```javascript
export let options = {
  stages: [
    { duration: '2m', target: 100 },
    { duration: '5m', target: 500 },
    { duration: '2m', target: 1000 },
    { duration: '5m', target: 0 },
  ],
};
```

### k6 Custom Metrics

```javascript
import { Counter, Trend } from 'k6/metrics';

let businessTransactionCount = new Counter('business_transactions');
let orderProcessingTime = new Trend('order_processing_duration');

export default function() {
  // Scenario: Place order
  let order = createOrder();
  let start = Date.now();
  let res = http.post('http://api/orders', JSON.stringify(order));

  if (res.status === 201) {
    businessTransactionCount.add(1);
    orderProcessingTime.add(Date.now() - start);
  }
}
```

---

## Grafana Dashboard Specifications

### Shared Panel Templates

**Request Rate Panel** (All protocols):
```json
{
  "title": "Request Rate",
  "targets": [
    {
      "expr": "sum(rate(http_server_requests_total[5m])) by (method, route, status_code)",
      "legendFormat": "{{method}} {{route}} {{status_code}}"
    }
  ],
  "type": "graph"
}
```

**Latency Percentiles Panel**:
```json
{
  "title": "Latency Percentiles",
  "targets": [
    {
      "expr": "histogram_quantile(0.50, sum(rate(http_server_request_duration_bucket[5m])) by (le))",
      "legendFormat": "p50"
    },
    {
      "expr": "histogram_quantile(0.90, sum(rate(http_server_request_duration_bucket[5m])) by (le))",
      "legendFormat": "p90"
    },
    {
      "expr": "histogram_quantile(0.95, sum(rate(http_server_request_duration_bucket[5m])) by (le))",
      "legendFormat": "p95"
    },
    {
      "expr": "histogram_quantile(0.99, sum(rate(http_server_request_duration_bucket[5m])) by (le))",
      "legendFormat": "p99"
    }
  ]
}
```

**Trace Exemplars Panel**:
```json
{
  "title": "Slowest Requests (with traces)",
  "targets": [
    {
      "expr": "topk(10, http_server_request_duration_seconds)",
      "exemplar": true  // Links to Tempo traces
    }
  ],
  "type": "table"
}
```

### Protocol-Specific Panels

**SQL Dashboard**:
- Query breakdown by operation (SELECT vs INSERT vs UPDATE)
- Slow query log (queries >100ms with db.statement)
- Connection pool saturation
- Table-level access patterns

**Kafka Dashboard**:
- Producer throughput by topic
- Consumer lag by partition
- Message size distribution
- Broker health metrics (from OBI, not JMX)

---

## Integration Points & Dependencies

### Cross-Workstream Dependencies

```
Foundation (WS-7.0)
    ├── Creates k6 scripts → Used by WS-7.1, 7.2, 7.3, 7.4
    ├── Creates scenarios → Used by all workstreams
    └── Creates templates → Used by all workstreams

HTTP (WS-7.1) ────────┐
gRPC (WS-7.2) ────────┤
SQL (WS-7.3) ─────────┼──→ Integration (WS-7.5)
Data (WS-7.4) ────────┘
```

**Blocking Dependencies**:
- WS-7.1, 7.2, 7.3, 7.4 MUST wait for WS-7.0 (k6 scripts, templates)
- WS-7.5 MUST wait for WS-7.1, 7.2, 7.3, 7.4 (all apps deployed)

**No Blocking**: WS-7.1, 7.2, 7.3, 7.4 can run fully in parallel

### External Dependencies

- OBI DaemonSet deployed (Workstream 2)
- Grafana stack operational (Workstream 3)
- Kubernetes cluster available
- Docker registry accessible

---

## Timeline & Resource Allocation

### Phase 1: Foundation (Weeks 1-2)
**Agents**: 2 agents
- `system-architect`: Designs shared utilities, scenarios
- `planner`: Creates documentation templates, OBI pattern guide

**Milestones**:
- Week 1: k6 scripts completed
- Week 2: Scenarios and templates completed

### Phase 2: Parallel Development (Weeks 3-4)
**Agents**: 4 agents (one per workstream)
- `backend-dev` (Go): HTTP + gRPC workstreams
- `backend-dev` (Python): SQL workstream
- `backend-dev` (Node.js + Java): Data workstream (Redis + Kafka)

**Milestones**:
- Week 3: All applications implemented and containerized
- Week 4: Kubernetes deployments, dashboards, documentation complete

### Phase 3: Integration & Testing (Weeks 5-6)
**Agents**: 2 agents
- `tester`: Integration tests, multi-protocol flows
- `system-architect`: Documentation polish, CI/CD pipeline

**Milestones**:
- Week 5: Multi-protocol flow working, integration tests passing
- Week 6: CI/CD pipeline operational, documentation complete

### Total Timeline: 6 weeks

### Resource Requirements

**Human Resources**:
- 2 agents (weeks 1-2)
- 4 agents (weeks 3-4)
- 2 agents (weeks 5-6)
- **Total Agent-Weeks**: 2×2 + 4×2 + 2×2 = 16 agent-weeks

**Infrastructure**:
- Kubernetes cluster (3 nodes minimum)
- Docker registry
- PostgreSQL (1 instance)
- Redis (1 instance)
- Kafka (3 brokers + 3 Zookeeper)
- Storage: ~50GB for databases and Kafka logs

---

## Agent Assignment Matrix

| Workstream | Primary Agent | Skills Required | Duration |
|------------|---------------|-----------------|----------|
| WS-7.0 Foundation | `system-architect` | Architecture, k6, documentation | 2 weeks |
| WS-7.1 HTTP | `backend-dev` (Go) | Go, Echo, HTTP/2, REST | 2 weeks |
| WS-7.2 gRPC | `backend-dev` (Go) | Go, gRPC, Protobuf | 2 weeks |
| WS-7.3 SQL | `backend-dev` (Python) | Python, FastAPI, SQLAlchemy, PostgreSQL | 2 weeks |
| WS-7.4 Data | `backend-dev` (Multi) | Node.js, Java, Redis, Kafka | 2 weeks |
| WS-7.5 Integration | `tester` + `system-architect` | Testing, CI/CD, documentation | 2 weeks |

---

## Success Metrics

### Technical Metrics
- [ ] All 5 protocol applications deployed and operational
- [ ] Zero application-side instrumentation (OBI only)
- [ ] Telemetry coverage: 100% of protocols instrumented
- [ ] Load tests generate realistic traffic (>1000 RPS combined)
- [ ] Integration tests pass in CI/CD
- [ ] Dashboard accuracy: All panels populated with OBI data

### Quality Metrics
- [ ] Code coverage: >80% for all applications
- [ ] Documentation completeness: 100% (all templates filled)
- [ ] Zero critical security vulnerabilities (container scans)
- [ ] Performance: Applications handle 10x baseline load without degradation

### Business Metrics
- [ ] Adoption: 5+ teams explore examples within 30 days
- [ ] Feedback: >8/10 satisfaction score on documentation clarity
- [ ] Time-to-value: New user can deploy all examples in <1 hour
- [ ] Troubleshooting effectiveness: Users resolve 80% of issues using guides

---

## Risk Assessment & Mitigation

### Risk 1: eBPF Instrumentation Gaps
**Risk**: OBI may not capture all expected telemetry for newer protocol versions.
**Probability**: Medium
**Impact**: High
**Mitigation**:
- Test with latest OBI version
- Document known gaps
- Fallback to manual spans if critical data missing
- Contribute fixes to OBI upstream

### Risk 2: Multi-Language Coordination
**Risk**: 5 different languages/frameworks increases coordination complexity.
**Probability**: Medium
**Impact**: Medium
**Mitigation**:
- Clear API contracts between services
- Shared OpenAPI/Protobuf definitions
- Standardized logging/error formats
- Weekly sync between agents

### Risk 3: Kubernetes Resource Constraints
**Risk**: Running all examples + infrastructure exceeds cluster capacity.
**Probability**: Low
**Impact**: Medium
**Mitigation**:
- Resource limits on all deployments
- Namespaces for isolation
- Optional: Deploy examples on-demand, not all at once
- Autoscaling for bursty workloads

### Risk 4: Documentation Drift
**Risk**: Code evolves but documentation not updated.
**Probability**: High
**Impact**: Low
**Mitigation**:
- Documentation as code (checked into repo)
- CI/CD validates code examples
- Quarterly documentation review
- Automated link checking

---

## Definition of Done

### Overall Workstream
- [ ] All 5 protocol applications deployed to dev cluster
- [ ] Zero manual instrumentation (only OBI)
- [ ] All load generators operational
- [ ] All Grafana dashboards published
- [ ] All integration tests passing in CI/CD
- [ ] Documentation complete with screenshots
- [ ] Troubleshooting scenarios validated
- [ ] Code reviewed and merged to main branch
- [ ] Demo presentation recorded (video walkthrough)

### Per-Application Checklist
- [ ] Application builds and runs locally
- [ ] Dockerfile optimized (multi-stage build, distroless base)
- [ ] Kubernetes manifests valid (linting passed)
- [ ] Health checks respond correctly
- [ ] Load test runs successfully
- [ ] OBI telemetry visible in Grafana
- [ ] Integration test passes
- [ ] README.md complete with examples
- [ ] Code coverage >80%
- [ ] Security scan passed (no critical vulnerabilities)

---

## Agent Coordination Hooks

### Foundation Workstream (WS-7.0)
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-7.0-foundation"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-7.0"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/00-shared/k6/http-load.js" --memory-key "swarm/mop/ws-7.0/k6-http"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/docs/examples/obi-instrumentation-patterns.md" --memory-key "swarm/mop/ws-7.0/obi-patterns"
npx claude-flow@alpha hooks notify --message "Foundation shared utilities completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-7.0-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

### HTTP Workstream (WS-7.1)
```bash
npx claude-flow@alpha hooks pre-task --description "workstream-7.1-http-rest-api"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-7.1"

npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/01-http-rest-api/cmd/server/main.go" --memory-key "swarm/mop/ws-7.1/http-app"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/dashboards/examples/http-rest-api-dashboard.json" --memory-key "swarm/mop/ws-7.1/dashboard"
npx claude-flow@alpha hooks notify --message "HTTP REST API implementation completed"

npx claude-flow@alpha hooks post-task --task-id "ws-7.1-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

### gRPC Workstream (WS-7.2)
```bash
npx claude-flow@alpha hooks pre-task --description "workstream-7.2-grpc-microservice"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-7.2"

npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/02-grpc-microservice/proto/user.proto" --memory-key "swarm/mop/ws-7.2/protobuf"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/02-grpc-microservice/cmd/server/main.go" --memory-key "swarm/mop/ws-7.2/grpc-app"
npx claude-flow@alpha hooks notify --message "gRPC microservice implementation completed"

npx claude-flow@alpha hooks post-task --task-id "ws-7.2-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

### SQL Workstream (WS-7.3)
```bash
npx claude-flow@alpha hooks pre-task --description "workstream-7.3-sql-application"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-7.3"

npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/03-sql-application/app/models.py" --memory-key "swarm/mop/ws-7.3/sqlalchemy-models"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/03-sql-application/migrations/versions/001_initial.py" --memory-key "swarm/mop/ws-7.3/migrations"
npx claude-flow@alpha hooks notify --message "SQL application with PostgreSQL completed"

npx claude-flow@alpha hooks post-task --task-id "ws-7.3-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

### Data Workstream (WS-7.4)
```bash
npx claude-flow@alpha hooks pre-task --description "workstream-7.4-data-redis-kafka"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-7.4"

npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/04-redis-cache/src/index.js" --memory-key "swarm/mop/ws-7.4/redis-app"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/examples/05-kafka-streaming/src/main/java/com/mop/kafka/KafkaApplication.java" --memory-key "swarm/mop/ws-7.4/kafka-app"
npx claude-flow@alpha hooks notify --message "Redis and Kafka applications completed"

npx claude-flow@alpha hooks post-task --task-id "ws-7.4-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

### Integration Workstream (WS-7.5)
```bash
npx claude-flow@alpha hooks pre-task --description "workstream-7.5-integration"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-7.5"

npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/tests/examples/multi-protocol-flow-test.sh" --memory-key "swarm/mop/ws-7.5/e2e-test"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/docs/examples/troubleshooting-scenarios.md" --memory-key "swarm/mop/ws-7.5/troubleshooting"
npx claude-flow@alpha hooks notify --message "Integration testing and documentation completed"

npx claude-flow@alpha hooks post-task --task-id "ws-7.5-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

---

## References & Resources

### OBI Documentation
- [OBI Announcement](https://opentelemetry.io/blog/2025/obi-announcing-first-release/)
- [OBI GitHub](https://github.com/open-telemetry/opentelemetry-beyla)
- [eBPF Introduction](https://ebpf.io/what-is-ebpf/)

### Load Testing
- [k6 Documentation](https://k6.io/docs/)
- [k6 gRPC Extension](https://k6.io/docs/javascript-api/k6-net-grpc/)
- [k6 Best Practices](https://k6.io/docs/testing-guides/load-testing/)

### Grafana Dashboards
- [Grafana Dashboard Best Practices](https://grafana.com/docs/grafana/latest/dashboards/build-dashboards/best-practices/)
- [Trace Exemplars](https://grafana.com/docs/tempo/latest/metrics-generator/exemplars/)

### Protocols
- [gRPC Best Practices](https://grpc.io/docs/guides/performance/)
- [PostgreSQL Performance](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Redis Best Practices](https://redis.io/topics/memory-optimization)
- [Kafka Best Practices](https://docs.confluent.io/platform/current/kafka/deployment.html)

---

## Next Steps

1. **Review this plan** with stakeholders
2. **Assign agents** to each workstream
3. **Provision infrastructure** (Kubernetes cluster, registries)
4. **Kick off WS-7.0** (Foundation) immediately
5. **Schedule weekly syncs** for coordination
6. **Track progress** via GitHub Issues (one issue per task)
7. **Update timeline** based on actual velocity

---

**Prepared by**: `planner` agent
**Date**: 2025-11-08
**Version**: 1.0
