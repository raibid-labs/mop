# SQL Instrumentation with OBI eBPF

This guide explains how OBI (Observability via eBPF Instrumentation) automatically captures and traces SQL queries without requiring code changes or SDK integration.

## Overview

OBI uses eBPF (extended Berkeley Packet Filter) to intercept and trace SQL queries at the kernel level, providing complete observability for PostgreSQL (and other database) operations without modifying application code.

## How OBI Captures SQL Queries

### Zero-Code Instrumentation

OBI operates by:

1. **eBPF Probes**: Attaching to system calls and library functions used by database drivers
2. **Query Interception**: Capturing SQL queries as they're sent to the database
3. **Timing Instrumentation**: Recording precise start/end timestamps
4. **Result Tracking**: Monitoring query completion and row counts
5. **Context Propagation**: Linking SQL queries to HTTP/gRPC traces

### Supported Databases

- PostgreSQL (demonstrated in this example)
- MySQL/MariaDB
- MongoDB
- Redis (see redis-cache example)

### Captured Metrics

For each SQL query, OBI captures:

- **Query Text**: Full SQL statement (with parameter placeholders)
- **Query Type**: SELECT, INSERT, UPDATE, DELETE, etc.
- **Duration**: Execution time in milliseconds
- **Rows Affected**: Number of rows returned or modified
- **Connection Info**: Database, user, host
- **Trace Context**: Parent HTTP/gRPC request ID
- **Error Status**: Success/failure and error messages

## SQL Application Example

The order management system in `examples/03-sql-app` demonstrates various SQL patterns that OBI automatically instruments.

### Simple Queries

**Customer Lookup by ID:**
```go
// Repository code
func (r *CustomerRepository) GetByID(ctx context.Context, id int64) (*models.Customer, error) {
    query := `
        SELECT id, name, email, created_at
        FROM customers
        WHERE id = $1
    `

    var customer models.Customer
    err := r.db.QueryRow(ctx, query, id).Scan(&customer.ID, &customer.Name, &customer.Email, &customer.CreatedAt)
    return &customer, err
}
```

**OBI Captures:**
```json
{
  "query": "SELECT id, name, email, created_at FROM customers WHERE id = $1",
  "query_type": "SELECT",
  "duration_ms": 2.3,
  "rows_returned": 1,
  "parameters": ["[1]"],
  "trace_id": "a1b2c3d4e5f6",
  "span_id": "g7h8i9j0",
  "database": "orders",
  "table": "customers"
}
```

### Complex JOINs

**Order with Customer Details:**
```go
func (r *OrderRepository) GetByIDWithCustomer(ctx context.Context, id int64) (*models.OrderWithItems, error) {
    query := `
        SELECT
            o.id, o.customer_id, o.status, o.total, o.created_at, o.updated_at,
            c.id, c.name, c.email, c.created_at
        FROM orders o
        INNER JOIN customers c ON o.customer_id = c.id
        WHERE o.id = $1
    `

    // Query execution...
}
```

**OBI Captures:**
```json
{
  "query": "SELECT o.id, o.customer_id, ... FROM orders o INNER JOIN customers c ...",
  "query_type": "SELECT",
  "duration_ms": 5.7,
  "rows_returned": 1,
  "tables": ["orders", "customers"],
  "join_type": "INNER JOIN",
  "trace_id": "a1b2c3d4e5f6"
}
```

### Aggregation Queries

**Order Statistics:**
```go
func (r *OrderRepository) GetOrderStats(ctx context.Context, customerID int64) (map[string]interface{}, error) {
    query := `
        SELECT
            COUNT(*) as total_orders,
            SUM(total) as total_spent,
            AVG(total) as average_order_value
        FROM orders
        WHERE customer_id = $1
    `

    // Execution...
}
```

**OBI Captures:**
```json
{
  "query": "SELECT COUNT(*), SUM(total), AVG(total) FROM orders WHERE customer_id = $1",
  "query_type": "SELECT",
  "duration_ms": 12.4,
  "rows_returned": 1,
  "aggregations": ["COUNT", "SUM", "AVG"],
  "trace_id": "a1b2c3d4e5f6"
}
```

### Transactions

**Create Order with Items:**
```go
func (r *OrderRepository) Create(ctx context.Context, req *models.CreateOrderRequest) (*models.OrderWithItems, error) {
    tx, err := r.db.Begin(ctx)
    defer tx.Rollback(ctx)

    // Insert order
    err = tx.QueryRow(ctx, `INSERT INTO orders (...) VALUES (...) RETURNING id`, ...).Scan(&orderID)

    // Insert order items
    for _, item := range req.Items {
        tx.QueryRow(ctx, `INSERT INTO order_items (...) VALUES (...)`, ...)
    }

    tx.Commit(ctx)
}
```

**OBI Captures:**
```json
[
  {
    "query": "BEGIN",
    "query_type": "TRANSACTION",
    "duration_ms": 0.1,
    "transaction_id": "tx123",
    "trace_id": "a1b2c3d4e5f6"
  },
  {
    "query": "INSERT INTO orders (...) VALUES (...)",
    "query_type": "INSERT",
    "duration_ms": 3.2,
    "rows_affected": 1,
    "transaction_id": "tx123",
    "trace_id": "a1b2c3d4e5f6"
  },
  {
    "query": "INSERT INTO order_items (...) VALUES (...)",
    "query_type": "INSERT",
    "duration_ms": 1.8,
    "rows_affected": 1,
    "transaction_id": "tx123",
    "trace_id": "a1b2c3d4e5f6"
  },
  {
    "query": "COMMIT",
    "query_type": "TRANSACTION",
    "duration_ms": 2.5,
    "transaction_id": "tx123",
    "trace_id": "a1b2c3d4e5f6"
  }
]
```

## Slow Query Detection

OBI automatically identifies slow queries and common performance anti-patterns.

### N+1 Query Problem

**Problem Code:**
```go
func (r *OrderRepository) SimulateSlowQuery(ctx context.Context, customerID int64) ([]models.OrderWithItems, error) {
    // 1 query to get orders
    orders, err := r.ListByCustomer(ctx, customerID, 100, 0)

    result := make([]models.OrderWithItems, 0, len(orders))
    for _, order := range orders {
        // N queries to get items (one per order!)
        orderWithItems, err := r.GetByID(ctx, order.ID)
        result = append(result, *orderWithItems)
    }

    return result, nil
}
```

**OBI Detection:**
```json
{
  "pattern": "N+1 Query",
  "severity": "high",
  "queries_executed": 51,
  "total_duration_ms": 127.5,
  "recommendation": "Use JOIN or batch loading",
  "trace_id": "a1b2c3d4e5f6",
  "endpoint": "/customers/1/orders/slow",
  "queries": [
    {
      "query": "SELECT * FROM orders WHERE customer_id = $1",
      "duration_ms": 12.5,
      "rows_returned": 50
    },
    {
      "query": "SELECT * FROM order_items WHERE order_id = $1",
      "duration_ms": 2.3,
      "rows_returned": 3,
      "repetitions": 50
    }
  ]
}
```

### Optimized Alternative

**Better Code:**
```go
func (r *OrderRepository) GetOrdersWithItemsEfficient(ctx context.Context, customerID int64) ([]models.OrderWithItems, error) {
    // Single query with JOIN
    query := `
        SELECT
            o.id, o.customer_id, o.status, o.total, o.created_at, o.updated_at,
            oi.id, oi.product_id, oi.quantity, oi.price
        FROM orders o
        LEFT JOIN order_items oi ON o.order_id = oi.order_id
        WHERE o.customer_id = $1
        ORDER BY o.id, oi.id
    `

    // Process results...
}
```

**OBI Shows Improvement:**
```json
{
  "query": "SELECT o.id, ... FROM orders o LEFT JOIN order_items oi ...",
  "duration_ms": 8.7,
  "rows_returned": 150,
  "optimization": "Single query replaces 51 queries",
  "performance_gain": "93% faster (127.5ms → 8.7ms)"
}
```

### Missing Index Detection

OBI identifies queries that could benefit from indexes:

```json
{
  "query": "SELECT * FROM orders WHERE status = $1",
  "duration_ms": 245.3,
  "rows_scanned": 10000,
  "rows_returned": 150,
  "warning": "Full table scan detected",
  "recommendation": "Add index on orders(status)",
  "suggested_index": "CREATE INDEX idx_orders_status ON orders(status)"
}
```

## Connection Pool Monitoring

OBI tracks PostgreSQL connection pool metrics automatically:

### Pool Stats
```json
{
  "pool": {
    "max_conns": 25,
    "min_conns": 5,
    "active_conns": 12,
    "idle_conns": 8,
    "waiting_queries": 0,
    "total_conns_created": 142,
    "total_conns_closed": 122,
    "avg_wait_time_ms": 1.2,
    "max_wait_time_ms": 45.3
  }
}
```

### Pool Exhaustion Detection
```json
{
  "event": "connection_pool_exhausted",
  "severity": "critical",
  "active_conns": 25,
  "waiting_queries": 15,
  "wait_time_ms": 1250,
  "recommendation": "Increase pool size or optimize query patterns",
  "timestamp": "2025-11-11T10:30:45Z"
}
```

## Grafana Dashboard Integration

The SQL Application dashboard visualizes OBI metrics:

### Panels

1. **SQL Query Rate**
   - Queries per second by type (SELECT, INSERT, UPDATE, DELETE)
   - Helps identify traffic patterns

2. **Query Latency Percentiles**
   - p50, p95, p99 latencies
   - Identifies slow query outliers

3. **Query Type Distribution**
   - Pie chart showing read vs write ratio
   - Helps with read replica scaling decisions

4. **Slowest Queries Table**
   - Top 10 queries by duration
   - Direct link to slow query investigation

5. **Connection Pool Stats**
   - Active vs idle connections over time
   - Helps tune pool size

6. **SQL Error Rate**
   - Failed queries per second
   - Alerts on database issues

### Dashboard Queries

Example Prometheus queries used in the dashboard:

```promql
# Query rate by type
rate(obi_sql_queries_total{app="sql-app"}[5m])

# p99 latency
histogram_quantile(0.99, rate(obi_sql_query_duration_ms_bucket{app="sql-app"}[5m]))

# Top slow queries
topk(10, avg by(query) (obi_sql_query_duration_ms{app="sql-app"}))

# Connection pool utilization
obi_sql_connection_pool_active{app="sql-app"} / obi_sql_connection_pool_max{app="sql-app"}
```

## Best Practices

### 1. Use Prepared Statements
The pgx driver automatically uses prepared statements, which OBI tracks:

```go
// pgx automatically prepares this
result := db.QueryRow(ctx, "SELECT * FROM customers WHERE id = $1", id)
```

### 2. Batch Operations
Instead of individual inserts:

```go
// Good: Batch insert
batch := &pgx.Batch{}
for _, item := range items {
    batch.Queue("INSERT INTO items VALUES ($1, $2)", item.ID, item.Name)
}
db.SendBatch(ctx, batch)
```

OBI shows this as a single batch operation with multiple statements.

### 3. Use Indexes Appropriately
Monitor OBI metrics to identify:
- Full table scans
- High-duration queries
- Queries scanning many rows but returning few

### 4. Monitor Transaction Duration
Long-running transactions can cause:
- Lock contention
- Connection pool exhaustion
- Replication lag

OBI tracks transaction duration and identifies problematic transactions.

### 5. Set Connection Pool Limits
Based on OBI metrics:
- Monitor active/idle connection ratio
- Check for pool exhaustion events
- Tune `max_conns` based on PostgreSQL capacity

## Troubleshooting Guide

### High Query Latency

**Symptom**: p99 latency > 100ms

**Investigation**:
1. Check "Slowest Queries" panel in Grafana
2. Identify repeated slow queries
3. Use EXPLAIN ANALYZE in PostgreSQL
4. Add indexes or optimize query

**OBI Helps**:
- Identifies exact slow queries
- Shows query frequency
- Links to HTTP endpoints causing slowness

### N+1 Query Pattern

**Symptom**: High query count per HTTP request

**Investigation**:
1. Check trace details in OBI
2. Look for repeated similar queries
3. Identify the code path
4. Refactor to use JOINs or batch loading

**OBI Helps**:
- Detects N+1 patterns automatically
- Shows query sequence with timing
- Measures improvement after optimization

### Connection Pool Exhaustion

**Symptom**: Queries waiting for connections

**Investigation**:
1. Check connection pool utilization
2. Identify long-running transactions
3. Look for connection leaks
4. Increase pool size or optimize queries

**OBI Helps**:
- Shows pool metrics over time
- Identifies when exhaustion occurs
- Correlates with high-traffic periods

### Database Deadlocks

**Symptom**: Transaction rollbacks, deadlock errors

**Investigation**:
1. Check PostgreSQL logs
2. Review transaction patterns in OBI
3. Identify conflicting transactions
4. Reorder operations or use row-level locking

**OBI Helps**:
- Shows transaction timing and overlap
- Identifies conflicting queries
- Traces requests causing deadlocks

## Advanced Features

### Query Sampling

For high-traffic applications, OBI can sample queries:

```yaml
obi_config:
  sql:
    sampling_rate: 0.1  # Sample 10% of queries
    always_sample_slow_queries: true
    slow_query_threshold_ms: 100
```

### Parameter Redaction

OBI can redact sensitive parameters:

```yaml
obi_config:
  sql:
    redact_parameters: true
    parameter_placeholder: "[REDACTED]"
```

**Before:**
```json
{"query": "SELECT * FROM users WHERE email = $1", "parameters": ["user@example.com"]}
```

**After:**
```json
{"query": "SELECT * FROM users WHERE email = $1", "parameters": ["[REDACTED]"]}
```

### Custom Metrics

Add custom labels to SQL metrics:

```go
// Application code (no changes needed)
func (h *Handler) GetOrder(c *gin.Context) {
    // OBI automatically captures this with endpoint context
    order, err := h.repo.GetByID(c.Request.Context(), orderID)
}
```

OBI automatically tags SQL queries with:
- HTTP endpoint
- HTTP method
- Service name
- Kubernetes pod name

## Performance Impact

OBI's eBPF instrumentation has minimal overhead:

- **CPU**: < 1% additional CPU usage
- **Memory**: ~50MB per instrumented process
- **Latency**: < 0.1ms per query (imperceptible)
- **Network**: Metrics sent asynchronously

Tested with:
- 10,000 queries/second
- Complex JOINs and aggregations
- Multiple concurrent connections

## Comparison with Traditional Instrumentation

| Approach | Code Changes | Performance | Completeness | Deployment |
|----------|--------------|-------------|--------------|------------|
| **OBI eBPF** | None | < 1% overhead | 100% queries | DaemonSet |
| **APM SDK** | Extensive | 2-5% overhead | 80-90% | Per-app |
| **Database Logs** | None | High I/O | 100% | Database |
| **Query Comments** | Manual | Minimal | Partial | Per-query |

## Related Examples

- [HTTP API Example](../../examples/01-http-api/) - Shows HTTP + SQL correlation
- [Redis Cache Example](../../examples/04-redis-cache/) - Demonstrates caching patterns
- [gRPC Service Example](../../examples/02-grpc-service/) - gRPC + SQL integration

## Resources

- [PostgreSQL EXPLAIN](https://www.postgresql.org/docs/current/sql-explain.html)
- [pgx Driver Docs](https://github.com/jackc/pgx)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [N+1 Query Problem](https://stackoverflow.com/questions/97197/what-is-the-n1-selects-problem)
