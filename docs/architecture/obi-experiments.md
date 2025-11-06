# OBI Experiments and Examples

This document outlines proposed experiments to explore OpenTelemetry Backend Initiative (OBI) capabilities within the MOP platform.

## Experiment 1: Adaptive Tail-Based Sampling with SLO Integration

### Objective
Dynamically adjust sampling rates based on service-level objective (SLO) breaches to balance cost and observability.

### Hypothesis
We can reduce trace storage costs by 90% while maintaining 100% visibility during incidents by:
- Sampling at 10% during normal operations
- Automatically increasing to 50-100% when SLOs are breached
- Returning to baseline after a cooldown period

### Implementation

**Components:**
- OBI: Captures all traces initially
- Alloy: Applies adaptive sampling logic
- Prometheus/Mimir: Exposes SLO metrics
- Tempo: Stores sampled traces

**Alloy Configuration:**
```yaml
otelcol.processor.tail_sampling "adaptive" {
  # Default policy: 10% sampling
  policies = [
    {
      name   = "baseline"
      type   = "probabilistic"
      config = {
        sampling_percentage = 10
      }
    },
  ]

  # SLO breach detection
  decision_wait = "10s"

  # Query Mimir for SLO metrics
  # If p95 latency > 500ms OR error_rate > 1%
  # Switch to 50% sampling for 30 minutes
}

# Custom processor to query Mimir
otelcol.processor.transform "slo_check" {
  metric_statements {
    context = "resource"
    statements = [
      # Query: rate(http_requests_total{status=~"5.."}[5m]) > 0.01
      # If true, set attribute "slo_breach" = true
      "set(attributes[\"slo_breach\"], mimir_query(\"rate(http_requests_total{status=~'5..'}[5m]) > 0.01\"))",
    ]
  }
}
```

**Metrics to Track:**
- Cost savings: Trace volume reduction
- Coverage: Percentage of errors captured
- Detection latency: Time to increase sampling
- False positives: Unnecessary sampling increases

### Expected Results

| Metric | Baseline | Target | Actual |
|--------|----------|--------|--------|
| Cost reduction | 0% | 90% | _TBD_ |
| Error capture rate | 100% | 100% | _TBD_ |
| Normal sampling | 100% | 10% | _TBD_ |
| Incident sampling | 100% | 50-100% | _TBD_ |

### Success Criteria
- Cost reduction > 80%
- Error capture rate = 100%
- Sampling increase latency < 30s
- No missed incidents

---

## Experiment 2: Network-Level Service Dependency Discovery

### Objective
Automatically generate service dependency maps by analyzing network traffic without requiring application instrumentation.

### Hypothesis
OBI's eBPF network probes can identify service-to-service communication patterns and build dependency graphs in real-time.

### Implementation

**OBI Configuration:**
```yaml
# Enable network-level instrumentation
beyla:
  network_events:
    enabled: true
    protocols:
      - http
      - grpc
      - tcp

  export:
    attributes:
      service.name: "{kubernetes.pod.labels.app}"
      service.namespace: "{kubernetes.namespace}"
      peer.service: "{destination.service.name}"
```

**Analysis Pipeline:**
1. OBI captures network events with source/destination
2. Alloy enriches with Kubernetes metadata
3. Custom processor builds adjacency matrix
4. Export to Grafana for visualization

**Grafana Dashboard:**
- Node graph panel showing service dependencies
- Edge labels: Request rate, error rate, latency
- Auto-refresh every 5 minutes

### Metrics to Track
- Dependency accuracy: Compare to known architecture
- Discovery latency: Time to detect new service
- False positives: Spurious dependencies
- Coverage: Percentage of actual dependencies found

### Expected Results

| Metric | Target |
|--------|--------|
| Accuracy | > 95% |
| Discovery latency | < 5 min |
| False positive rate | < 5% |
| Coverage | > 90% |

### Use Cases
1. **Migration Planning**: Identify all dependencies before moving service
2. **Blast Radius Analysis**: Understand impact of service failures
3. **Architecture Validation**: Verify actual vs. intended dependencies
4. **Security**: Detect unexpected communication patterns

---

## Experiment 3: Database Query Performance Profiling

### Objective
Identify slow database queries across all services without database-side instrumentation or query log parsing.

### Hypothesis
OBI can capture SQL queries at the network layer, correlate with application traces, and identify performance bottlenecks.

### Implementation

**OBI Configuration:**
```yaml
beyla:
  discovery:
    services:
      - k8s_namespace: "production"
        protocols:
          - sql  # PostgreSQL, MySQL, etc.

  instrumentation:
    sql:
      capture_queries: true
      sanitize_queries: true  # Remove sensitive values
      max_query_length: 1024
```

**Analysis:**
1. OBI captures SQL queries with timing
2. Correlate queries with distributed traces (trace_id)
3. Aggregate by query pattern (sanitized)
4. Alert on queries exceeding threshold

**Alloy Processing:**
```yaml
otelcol.processor.transform "sql_analysis" {
  trace_statements = [
    # Extract query from span attributes
    "set(attributes[\"db.statement.normalized\"], normalize_sql(attributes[\"db.statement\"]))",

    # Calculate p95 duration per query pattern
    "set(attributes[\"db.query.p95\"], percentile(attributes[\"db.statement.normalized\"], 95))",
  ]
}
```

**Grafana Dashboard:**
- Top 10 slowest queries (p95 duration)
- Query count and frequency
- Affected services
- Trace examples for investigation

### Metrics to Track
- Query capture rate: Percentage of queries instrumented
- Sanitization accuracy: No PII leakage
- Correlation accuracy: Trace-to-query linkage
- False negatives: Missed slow queries

### Expected Results

| Metric | Target |
|--------|--------|
| Capture rate | > 99% |
| Correlation accuracy | > 95% |
| PII leakage | 0% |
| Detection latency | < 1 min |

### Use Cases
1. **Performance Optimization**: Find N+1 queries, missing indexes
2. **Capacity Planning**: Identify high-volume queries
3. **Security**: Detect SQL injection attempts
4. **Cost Attribution**: Query costs per service/team

---

## Experiment 4: Cost-Optimized Multi-Region Observability

### Objective
Deploy a multi-region observability architecture that minimizes data transfer costs while maintaining global visibility.

### Hypothesis
Regional Tempo instances with centralized Mimir metrics can reduce costs by 70-80% by keeping high-volume traces local while aggregating low-volume metrics globally.

### Architecture

```
┌─────────────────────────────────────────────┐
│              Global Region (us-east-1)      │
├─────────────────────────────────────────────┤
│  [Mimir] ← Aggregated Metrics               │
│  [Grafana] ← Global Dashboards              │
│  [Loki] ← Critical Logs Only                │
└──────▲───────────────────────────────────────┘
       │ Metrics only (low bandwidth)
       │
   ┌───┴────────┬────────────┬─────────────┐
   │            │            │             │
┌──▼────────┐ ┌─▼──────────┐ ┌─▼───────────┐
│ us-east-1 │ │ eu-west-1  │ │ ap-south-1  │
├───────────┤ ├────────────┤ ├─────────────┤
│ [OBI]     │ │ [OBI]      │ │ [OBI]       │
│ [Alloy]   │ │ [Alloy]    │ │ [Alloy]     │
│ [Tempo]   │ │ [Tempo]    │ │ [Tempo]     │ ← Regional
│ [Loki]    │ │ [Loki]     │ │ [Loki]      │ ← Regional
└───────────┘ └────────────┘ └─────────────┘
   Traces stay local         Traces stay local
```

### Cost Analysis

**Traditional (All data centralized):**
```
Data Transfer: 10TB/month × 3 regions × $0.09/GB = $2,700/month
Storage (Tempo): 30TB × $0.023/GB = $690/month
Total: $3,390/month
```

**Regional (Traces local, metrics central):**
```
Data Transfer: 100GB metrics × 3 regions × $0.09/GB = $27/month
Storage (Tempo): 30TB × $0.023/GB = $690/month (same, but regional)
Total: $717/month
```

**Savings: 79% ($2,673/month)**

### Implementation

**Alloy Configuration (Regional):**
```yaml
# Regional Alloy exports traces locally, metrics globally
otelcol.exporter.otlp "tempo_local" {
  client {
    endpoint = "tempo.local:4317"  # Same region
  }
}

otelcol.exporter.prometheusremotewrite "mimir_global" {
  endpoint {
    url = "https://mimir.global.mop.io/api/v1/push"
  }
}

# Route by signal type
otelcol.processor.routing "by_signal" {
  default_exporters = ["tempo_local"]
  table = [
    {
      statement = "route() where signal == 'traces'",
      exporters = ["tempo_local"],
    },
    {
      statement = "route() where signal == 'metrics'",
      exporters = ["mimir_global"],
    },
  ]
}
```

### Metrics to Track
- Data transfer volume per region
- Query latency (global vs. regional)
- Cost savings actual vs. projected
- User satisfaction with global view

### Expected Results

| Metric | Baseline | Target |
|--------|----------|--------|
| Data transfer cost | $2,700 | < $500 |
| Total cost | $3,390 | < $1,000 |
| Savings | 0% | > 70% |
| Global query latency | N/A | < 500ms |

---

## Experiment 5: Canary Deployment Automated Rollback

### Objective
Use OBI-generated metrics to automatically rollback canary deployments when error rates or latency exceed thresholds.

### Hypothesis
OBI provides reliable, zero-code metrics that can drive automated quality gates for progressive delivery.

### Implementation

**Stack:**
- OBI: Capture canary metrics
- Alloy: Route canary metrics with labels
- Mimir: Store metrics
- Argo Rollouts: Progressive delivery
- Custom controller: Query metrics and trigger rollback

**Argo Rollouts Analysis Template:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: obi-canary-metrics
spec:
  metrics:
    - name: error-rate
      interval: 30s
      successCondition: result < 0.01  # < 1% errors
      failureLimit: 3
      provider:
        prometheus:
          address: http://mimir:9009
          query: |
            sum(rate(http_server_request_count{
              deployment="my-service",
              status=~"5..",
              version="{{args.canary-version}}"
            }[5m]))
            /
            sum(rate(http_server_request_count{
              deployment="my-service",
              version="{{args.canary-version}}"
            }[5m]))

    - name: latency-p95
      interval: 30s
      successCondition: result < 500  # < 500ms
      failureLimit: 3
      provider:
        prometheus:
          address: http://mimir:9009
          query: |
            histogram_quantile(0.95,
              sum(rate(http_server_request_duration_bucket{
                deployment="my-service",
                version="{{args.canary-version}}"
              }[5m])) by (le)
            )
```

**Rollout Configuration:**
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: my-service
spec:
  replicas: 10
  strategy:
    canary:
      steps:
        - setWeight: 10  # 10% canary
        - pause: {duration: 5m}
        - analysis:
            templates:
              - templateName: obi-canary-metrics
        - setWeight: 50  # If passed, 50% canary
        - pause: {duration: 5m}
        - analysis:
            templates:
              - templateName: obi-canary-metrics
        - setWeight: 100  # Full rollout

      # Automatic rollback on failure
      abortScaleDownDelaySeconds: 30
```

### Metrics to Track
- False positives: Rollback when canary was fine
- False negatives: No rollback when canary was bad
- Detection latency: Time to detect bad canary
- Rollback latency: Time to restore stable version

### Expected Results

| Metric | Target |
|--------|--------|
| False positive rate | < 5% |
| False negative rate | < 1% |
| Detection latency | < 2 min |
| Rollback latency | < 1 min |

### Use Cases
1. **Risk Mitigation**: Catch canary issues before full rollout
2. **Reduced MTTR**: Automatic rollback vs. manual intervention
3. **Confidence**: Deploy more frequently with safety net
4. **Metrics-Driven**: Objective quality gates, not subjective

---

## Experiment Comparison

| Experiment | Impact | Complexity | Time to Value |
|-----------|--------|------------|---------------|
| Adaptive Sampling | 🟢 High (cost) | 🟡 Medium | 2 weeks |
| Service Discovery | 🟢 High (architecture) | 🟢 Low | 1 week |
| SQL Profiling | 🟢 High (performance) | 🟡 Medium | 2 weeks |
| Multi-Region | 🟢 High (cost) | 🔴 High | 4 weeks |
| Canary Rollback | 🟢 High (reliability) | 🟡 Medium | 3 weeks |

## Implementation Roadmap

### Phase 1: Foundation (Weeks 1-2)
- Deploy base MOP stack (OBI + Grafana + Alloy + Tempo/Mimir/Loki)
- Instrument demo application
- Validate data flow

### Phase 2: Quick Wins (Weeks 3-4)
- **Experiment 2**: Service Discovery (low complexity, high value)
- **Experiment 1**: Adaptive Sampling (immediate cost savings)

### Phase 3: Deep Dives (Weeks 5-8)
- **Experiment 3**: SQL Profiling (engineering efficiency)
- **Experiment 5**: Canary Rollback (reliability improvement)

### Phase 4: Scale (Weeks 9-12)
- **Experiment 4**: Multi-Region (cost optimization at scale)
- Documentation and best practices

## Success Metrics

### Technical
- All 5 experiments completed
- Cost reduction: > 60% overall
- Detection latency: < 1 minute for incidents
- False positive rate: < 5%

### Business
- Faster MTTR: < 5 minutes (vs. 30+ minutes baseline)
- More frequent deployments: 2x increase
- Developer satisfaction: > 8/10
- Reduced on-call burden: 50% fewer alerts

## Resources

- [OBI Documentation](https://opentelemetry.io/blog/2025/obi-announcing-first-release/)
- [Argo Rollouts](https://argoproj.github.io/argo-rollouts/)
- [Alloy Sampling](https://grafana.com/docs/alloy/latest/reference/components/otelcol.processor.tail_sampling/)
- [Cost Optimization Guide](cost-optimization.md)

---

**Status**: Proposed
**Next Steps**: Review with engineering team, prioritize experiments, assign workstreams
