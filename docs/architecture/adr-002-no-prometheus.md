# ADR-002: Use Mimir Instead of Prometheus

## Status

**ACCEPTED**

## Context

Traditional observability stacks use Prometheus for metrics collection and storage. However, Prometheus has scalability limitations for large, multi-tenant, cloud-native environments.

Options evaluated:
1. **Prometheus**: Industry-standard, simple, widely adopted
2. **Mimir**: Horizontally scalable, Prometheus-compatible, object storage
3. **Thanos**: Prometheus HA with object storage
4. **Cortex**: Mimir's predecessor (now deprecated in favor of Mimir)
5. **Victoria Metrics**: Fast, cost-efficient, Prometheus-compatible

## Decision

**Use Mimir as the metrics backend. No standalone Prometheus.**

## Rationale

### Why Mimir?

**1. Horizontal Scalability**
- Prometheus: Single-instance, limited by local disk
- Mimir: Distributed architecture, scales to billions of active series
- Components: Distributor, Ingester, Querier, Compactor, Store Gateway

**2. Cost-Efficient Storage**
- Prometheus: Local SSD required (expensive at scale)
- Mimir: Object storage (S3, GCS, Azure) - 10x cheaper
- Long-term retention without breaking the bank

**3. Multi-Tenancy**
- Prometheus: Single-tenant only
- Mimir: Native multi-tenancy with isolated namespaces
- Critical for shared platforms

**4. Prometheus Compatibility**
- Full PromQL support
- Remote write API compatible
- Existing Prometheus dashboards work unchanged
- Grafana datasource: same as Prometheus

**5. High Availability**
- Prometheus: Requires external HA setup (Thanos/federation)
- Mimir: Built-in HA via replication factor
- No single point of failure

**6. Better Retention**
- Prometheus: Retention limited by disk size
- Mimir: Unlimited retention in object storage
- Tiered storage (hot/warm/cold)

### Why Not Prometheus?

**Scale Limitations:**
- Max ~10M active series per instance
- Query performance degrades with large datasets
- Manual sharding required for scale
- Complex federation topologies

**Operational Complexity:**
- High cardinality = out of memory
- Retention = expensive local storage
- HA = complex external systems (Thanos)
- Disaster recovery = custom backup solutions

### Why Not Alternatives?

**Thanos:**
- More complex than Mimir (more components)
- Prometheus still required (adds another layer)
- Slower query performance than Mimir
- More operational overhead

**Victoria Metrics:**
- Less mature ecosystem
- Smaller community
- Not CNCF project (Mimir is)
- Fewer integrations

## Architecture

```
┌──────────────┐
│ OBI + Alloy  │
└──────┬───────┘
       │ Remote Write (Prometheus protocol)
       │
┌──────▼────────────────────────────────────┐
│              Mimir Cluster                │
├───────────────────────────────────────────┤
│                                           │
│  [Distributor] → [Ingester] → [Store]   │
│        ↓              ↓          ↓        │
│    Replication    WAL Cache   Blocks     │
│        ↓              ↓          ↓        │
│  [Querier] ← [Query Frontend]            │
│        ↓                                  │
│  [Grafana] ← PromQL API                  │
│                                           │
│  Backend: [S3 / GCS / Azure Blob]        │
└───────────────────────────────────────────┘
```

## Components

### Mimir Services

1. **Distributor**: Receives remote write, validates, replicates
2. **Ingester**: Buffers recent data, writes blocks to object storage
3. **Querier**: Executes PromQL queries
4. **Query Frontend**: Query coordination and caching
5. **Compactor**: Merges and downsamples blocks
6. **Store Gateway**: Queries long-term storage
7. **Ruler**: Evaluates recording and alerting rules
8. **Alertmanager**: Handles alert routing

### Storage Tiers

- **Recent (< 12h)**: Ingester memory + WAL
- **Recent (12h-24h)**: Ingester blocks in object storage
- **Long-term (> 24h)**: Compacted blocks in object storage

## Configuration Example

```yaml
# Mimir configuration
multitenancy_enabled: true

limits:
  ingestion_rate: 10000
  ingestion_burst_size: 200000
  max_global_series_per_user: 10000000

blocks_storage:
  backend: s3
  s3:
    endpoint: s3.amazonaws.com
    bucket_name: mop-metrics
  tsdb:
    retention_period: 30d  # In memory
    block_ranges_period:
      - 2h   # Recent blocks
      - 12h  # Medium blocks
      - 24h  # Compacted blocks

compactor:
  compaction_interval: 30m
  retention_enabled: true
  retention_delete_delay: 12h

query_frontend:
  cache_results: true
  results_cache:
    backend: memcached
```

## Migration Path

Users familiar with Prometheus can use Mimir without changes:

**Prometheus Remote Write:**
```yaml
# Alloy sends to Mimir exactly like Prometheus
prometheus.remote_write "mimir" {
  endpoint {
    url = "http://mimir:9009/api/v1/push"
  }
}
```

**Grafana Datasource:**
```yaml
# Same configuration as Prometheus
datasources:
  - name: Mimir
    type: prometheus
    url: http://mimir:9009/prometheus
    jsonData:
      timeInterval: 15s
```

## Performance Comparison

| Metric | Prometheus | Mimir |
|--------|-----------|-------|
| Max Active Series | 10M | Unlimited (billions) |
| Query Latency (p99) | 500ms | 300ms (cached) |
| Ingestion Rate | 1M samples/sec | 10M+ samples/sec |
| Storage Cost (1TB) | $100/month (SSD) | $10/month (S3) |
| HA Setup | Complex (Thanos) | Native |
| Retention | Disk-limited | Unlimited |

## Consequences

### Positive
- Unlimited scale for metrics
- Cost-efficient long-term retention
- Built-in HA and multi-tenancy
- Prometheus compatibility (no migration pain)
- Cloud-native architecture

### Negative
- More complex than single Prometheus instance
- Requires object storage
- Additional operational knowledge needed
- Multiple services to monitor (vs one Prometheus)

### Mitigation
- Start with simple Mimir deployment (monolithic mode)
- Graduate to microservices mode as scale increases
- Use Helm charts for easy deployment
- Leverage Grafana dashboards for Mimir monitoring

## Cost Analysis

**Scenario: 1M active series, 30-day retention**

| Solution | Storage | Compute | Total/Month |
|----------|---------|---------|-------------|
| Prometheus (SSD) | $100 | $50 | **$150** |
| Prometheus + Thanos | $20 (S3) + $100 (SSD) | $80 | **$200** |
| Mimir | $20 (S3) | $60 | **$80** |

**Mimir is 47% cheaper than Prometheus, 60% cheaper than Thanos.**

## References

- [Mimir Documentation](https://grafana.com/docs/mimir/)
- [Mimir Architecture](https://grafana.com/docs/mimir/latest/references/architecture/)
- [Scaling Prometheus with Mimir](https://grafana.com/blog/2022/03/29/scaling-prometheus-with-grafana-mimir/)
- [Prometheus vs Mimir Comparison](https://grafana.com/blog/2022/04/25/prometheus-and-grafana-mimir/)

## Related Decisions

- ADR-003: OBI as Primary Instrumentation (generates metrics for Mimir)
- ADR-001: Alloy Operator (Alloy sends metrics to Mimir)

---

**Date**: 2025-01-06
**Author**: MOP Architecture Team
**Reviewers**: Platform Engineering, SRE, FinOps
