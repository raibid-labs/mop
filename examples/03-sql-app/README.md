# SQL Application - Order Management System

A reference implementation demonstrating OBI eBPF automatic SQL instrumentation for PostgreSQL databases. This application showcases zero-code observability for database operations.

## Overview

This order management system demonstrates:
- Automatic SQL query tracing with OBI eBPF
- Connection pool monitoring
- Slow query detection (including N+1 problems)
- Complex query patterns (JOINs, aggregations, transactions)
- Production-ready Go patterns with PostgreSQL

## Features

### Database Operations
- **Customer Management**: Create and retrieve customer records
- **Order Management**: Create orders with multiple items, update status, retrieve with relationships
- **Order Statistics**: Aggregated metrics per customer
- **Complex Queries**: Demonstrates JOINs between orders and customers
- **N+1 Query Simulation**: Intentionally inefficient endpoint for OBI testing

### OBI Instrumentation Points

OBI automatically captures:
- **Query Execution**: All SQL queries (SELECT, INSERT, UPDATE, etc.)
- **Query Duration**: Precise timing for each query
- **Query Type**: Classification (SELECT, INSERT, UPDATE, DELETE)
- **Connection Pool**: Active/idle connection metrics
- **Slow Queries**: Queries exceeding performance thresholds
- **N+1 Patterns**: Multiple sequential queries that could be optimized

## Architecture

```
examples/03-sql-app/
├── cmd/
│   └── server/          # Main application entry point
├── internal/
│   ├── db/              # Database connection pooling
│   ├── handlers/        # HTTP handlers (Gin framework)
│   ├── models/          # Data models
│   └── repository/      # Repository pattern (pgx driver)
├── migrations/          # SQL schema migrations
├── tests/               # Integration tests
├── Dockerfile           # Container build
├── docker-compose.yaml  # Local development setup
└── Makefile            # Build and deployment commands
```

## Technology Stack

- **Language**: Go 1.21+
- **Database**: PostgreSQL 15
- **Driver**: pgx/v5 (native PostgreSQL driver)
- **HTTP Framework**: Gin
- **Containerization**: Docker, Docker Compose
- **Orchestration**: Kubernetes

## Quick Start

### Prerequisites
- Go 1.21 or higher
- Docker and Docker Compose
- PostgreSQL 15 (for local development without Docker)

### Local Development with Docker Compose

```bash
# Start PostgreSQL and application
make docker-up

# The application will be available at http://localhost:8080
# PostgreSQL will be available at localhost:5432
```

### Local Development without Docker

```bash
# Set environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=orders
export DB_SSLMODE=disable

# Run migrations (manual)
psql -U postgres -d orders < migrations/001_create_customers.up.sql
psql -U postgres -d orders < migrations/002_create_orders.up.sql

# Run the application
make run
```

## API Endpoints

### Health Checks
- `GET /health` - Health check with database connectivity
- `GET /ready` - Readiness check

### Customer Management
- `POST /customers` - Create a new customer
- `GET /customers/:id` - Get customer by ID
- `GET /customers` - List customers (paginated)

### Order Management
- `POST /orders` - Create a new order with items
- `GET /orders/:id` - Get order by ID
- `GET /orders/:id?include_customer=true` - Get order with customer details (JOIN)
- `PUT /orders/:id/status` - Update order status
- `GET /customers/:customer_id/orders` - List orders for a customer
- `GET /customers/:customer_id/orders/stats` - Get order statistics (aggregation)
- `GET /customers/:customer_id/orders/slow` - Simulate N+1 query problem

## Example Requests

### Create a Customer
```bash
curl -X POST http://localhost:8080/customers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com"
  }'
```

### Create an Order
```bash
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{
    "customer_id": 1,
    "items": [
      {"product_id": 101, "quantity": 2, "price": 29.99},
      {"product_id": 102, "quantity": 1, "price": 49.99}
    ]
  }'
```

### Get Order with Customer (JOIN Query)
```bash
curl http://localhost:8080/orders/1?include_customer=true
```

### Update Order Status
```bash
curl -X PUT http://localhost:8080/orders/1/status \
  -H "Content-Type: application/json" \
  -d '{"status": "shipped"}'
```

### Get Order Statistics (Aggregation)
```bash
curl http://localhost:8080/customers/1/orders/stats
```

### Trigger Slow Query Pattern
```bash
# This endpoint intentionally uses N+1 queries
curl http://localhost:8080/customers/1/orders/slow
```

## Database Schema

### Customers Table
```sql
CREATE TABLE customers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
```

### Orders Table
```sql
CREATE TABLE orders (
    id BIGSERIAL PRIMARY KEY,
    customer_id BIGINT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (customer_id) REFERENCES customers(id)
);
```

### Order Items Table
```sql
CREATE TABLE order_items (
    id BIGSERIAL PRIMARY KEY,
    order_id BIGINT NOT NULL,
    product_id BIGINT NOT NULL,
    quantity INT NOT NULL CHECK (quantity > 0),
    price DECIMAL(10, 2) NOT NULL CHECK (price >= 0),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (order_id) REFERENCES orders(id)
);
```

## OBI Instrumentation

### Automatic SQL Tracing

OBI eBPF automatically instruments all PostgreSQL queries without code changes:

1. **Simple Queries**
   ```sql
   SELECT id, name, email FROM customers WHERE id = $1
   ```
   - Captured: Query text, duration, result rows

2. **Complex JOINs**
   ```sql
   SELECT o.*, c.name, c.email
   FROM orders o
   INNER JOIN customers c ON o.customer_id = c.id
   WHERE o.id = $1
   ```
   - Captured: Multi-table query, JOIN performance

3. **Aggregations**
   ```sql
   SELECT COUNT(*), SUM(total), AVG(total)
   FROM orders
   WHERE customer_id = $1
   ```
   - Captured: Aggregation query, calculation time

4. **Transactions**
   ```sql
   BEGIN;
   INSERT INTO orders (...) VALUES (...);
   INSERT INTO order_items (...) VALUES (...);
   COMMIT;
   ```
   - Captured: Transaction boundaries, individual statements

### Slow Query Detection

OBI identifies slow queries automatically:

- **N+1 Problem**: The `/customers/:id/orders/slow` endpoint demonstrates this anti-pattern
  - First query: `SELECT * FROM orders WHERE customer_id = $1`
  - N subsequent queries: `SELECT * FROM order_items WHERE order_id = $1` (for each order)
  - OBI captures the pattern and timing

- **Missing Indexes**: Queries without appropriate indexes show up as slow
- **Complex JOINs**: Multi-table queries with performance implications

### Connection Pool Monitoring

OBI tracks PostgreSQL connection pool metrics:
- Active connections
- Idle connections
- Connection wait time
- Pool exhaustion events

## Kubernetes Deployment

### Deploy to Kubernetes
```bash
kubectl apply -k deployments/examples/03-sql-app/

# Verify deployment
kubectl get pods -n sql-app
kubectl get svc -n sql-app
```

### Access the Application
```bash
# Port forward to access locally
kubectl port-forward -n sql-app svc/sql-app 8080:80

# Test the application
curl http://localhost:8080/health
```

### View OBI Metrics
```bash
# Access Grafana dashboard
kubectl port-forward -n observability svc/grafana 3000:80

# Navigate to: SQL Application - OBI Instrumentation dashboard
```

## Testing

### Run Unit and Integration Tests
```bash
# Run all tests
make test

# View coverage report
open coverage.html
```

### Integration Tests
The test suite includes:
- Customer CRUD operations
- Order creation with transactions
- Order status updates
- Customer order statistics
- Health check validation

### Benchmark Tests
```bash
go test -bench=. ./tests/
```

## Grafana Dashboard

The SQL application includes a comprehensive Grafana dashboard showing:

1. **Query Rate**: Queries per second by type (SELECT, INSERT, UPDATE, DELETE)
2. **Query Latency**: p50, p95, p99 percentiles
3. **Query Distribution**: Pie chart of query types
4. **Slowest Queries**: Top 10 queries by duration
5. **Connection Pool**: Active vs idle connections
6. **Error Rate**: SQL errors over time

Dashboard location: `lib/grafana/dashboards/examples/sql-app-dashboard.json`

## Troubleshooting Slow Queries with OBI

### Identifying N+1 Queries

1. Check the Grafana dashboard for repeated query patterns
2. Look for endpoints with high query counts
3. Use OBI traces to see the sequence of queries
4. Example: The `/customers/:id/orders/slow` endpoint shows N+1 pattern

### Optimizing Queries

**Before (N+1 pattern):**
```go
// One query to get orders
orders := GetOrdersByCustomer(customerID)

// N queries to get items for each order
for _, order := range orders {
    items := GetOrderItems(order.ID)  // Separate query each time!
}
```

**After (JOIN or batching):**
```go
// Single query with JOIN or batch loading
ordersWithItems := GetOrdersWithItemsByCustomer(customerID)
```

OBI will show the dramatic improvement in query count and total duration.

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | localhost | PostgreSQL host |
| `DB_PORT` | 5432 | PostgreSQL port |
| `DB_USER` | postgres | Database user |
| `DB_PASSWORD` | postgres | Database password |
| `DB_NAME` | orders | Database name |
| `DB_SSLMODE` | disable | SSL mode (disable/require) |
| `DB_MAX_CONNS` | 25 | Maximum connections in pool |
| `DB_MIN_CONNS` | 5 | Minimum connections in pool |
| `DB_MAX_CONN_LIFETIME` | 3600 | Max connection lifetime (seconds) |
| `DB_MAX_CONN_IDLE_TIME` | 300 | Max connection idle time (seconds) |
| `SERVER_PORT` | 8080 | HTTP server port |

## Performance Considerations

### Connection Pooling
- Default pool size: 25 max, 5 min connections
- Tune based on workload and PostgreSQL max_connections
- OBI shows pool utilization to guide tuning

### Query Optimization
- Use prepared statements (pgx does this automatically)
- Avoid N+1 queries (use JOINs or batch loading)
- Add appropriate indexes (see migration files)
- Monitor slow queries via OBI

### Resource Limits
Kubernetes deployment includes:
- CPU: 100m request, 200m limit
- Memory: 128Mi request, 256Mi limit
- Adjust based on OBI metrics

## Development Workflow

### TDD Approach
1. Write test for new functionality
2. Implement minimal code to pass test
3. Refactor while keeping tests green
4. Run OBI to verify query patterns

### Adding New Endpoints
1. Define model in `internal/models/`
2. Create repository method in `internal/repository/`
3. Add handler in `internal/handlers/`
4. Wire up route in `cmd/server/main.go`
5. Write integration test in `tests/`
6. Verify with OBI instrumentation

## Related Documentation

- [OBI SQL Instrumentation Guide](../../docs/examples/sql-instrumentation.md)
- [Slow Query Detection with OBI](../../docs/examples/sql-instrumentation.md#slow-query-detection)
- [PostgreSQL Performance Tuning](https://www.postgresql.org/docs/current/performance-tips.html)
- [pgx Driver Documentation](https://github.com/jackc/pgx)

## License

Part of the MOP (Multi-protocol Observability Platform) project.
