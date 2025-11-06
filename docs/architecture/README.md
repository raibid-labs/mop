# Architecture Documentation

This directory contains architecture decisions, system design documents, and integration patterns for the Managed Observability Platform (MOP).

## Overview

MOP is built on OpenTelemetry Backend Initiative (OBI) and the Grafana observability stack, managed via Tanka/Jsonnet infrastructure as code.

## Core Components

### 1. OpenTelemetry Backend Initiative (OBI)
- **Purpose**: eBPF-based instrumentation with zero code changes
- **Overhead**: <1% CPU
- **Coverage**: HTTP, gRPC, SQL, Redis, MongoDB, Kafka, GraphQL, S3
- **Output**: OTLP (OpenTelemetry Protocol)

### 2. Grafana Alloy
- **Purpose**: Telemetry pipeline and routing
- **Capabilities**:
  - Adaptive sampling (tail-based, probabilistic)
  - Dynamic routing based on labels
  - Cost optimization through filtering
  - Multi-destination export

### 3. Tempo
- **Purpose**: Distributed tracing backend
- **Storage**: Object storage (S3, GCS, Azure Blob)
- **Features**:
  - TraceQL query language
  - Metrics generation from traces
  - Cost-efficient retention

### 4. Mimir
- **Purpose**: Long-term metrics storage
- **Architecture**: Horizontally scalable
- **API**: Prometheus-compatible
- **Storage**: Object storage + memcached
- **Why not Prometheus?**: Scale limitations, single-instance

### 5. Loki
- **Purpose**: Log aggregation and querying
- **Features**:
  - Trace-log correlation
  - LogQL query language
  - Label-based indexing (cost-efficient)

### 6. Grafana
- **Purpose**: Unified visualization and alerting
- **Configuration**: Stateless, auth disabled (internal use)
- **Features**:
  - Pre-provisioned datasources
  - Default dashboards
  - SLO tracking

## Architecture Diagrams

### Data Flow

```
[App] → [OBI eBPF] → OTLP → [Alloy] ──┬→ [Tempo] (traces)
                                       ├→ [Mimir] (metrics)
                                       └→ [Loki] (logs)
                                            ↓
                                        [Grafana]
```

### Deployment Topology

```
┌─────────────────────────────────────┐
│          Kubernetes Cluster         │
├─────────────────────────────────────┤
│                                     │
│  ┌───────────────┐                 │
│  │ OBI DaemonSet │ (every node)    │
│  └───────┬───────┘                 │
│          │ OTLP                    │
│  ┌───────▼───────────┐             │
│  │ Alloy StatefulSet │ (3 replicas)│
│  └───────┬───────────┘             │
│          │                         │
│     ┌────┼────┬─────┐              │
│     │    │    │     │              │
│  ┌──▼──┐┌▼──┐┌▼───┐┌▼────┐        │
│  │Tempo││Mimir│Loki││Grafana│       │
│  │(3r) ││(3r)││(3r)││(2r)   │       │
│  └─────┘└────┘└────┘└──────┘       │
│                                     │
│  [Object Storage: S3/GCS/Azure]    │
└─────────────────────────────────────┘
```

## Architecture Decision Records (ADRs)

- [ADR-001: Alloy Operator vs Standalone](adr-001-alloy-operator.md)
- [ADR-002: No Prometheus, Use Mimir](adr-002-no-prometheus.md)
- [ADR-003: OBI as Primary Instrumentation](adr-003-obi-instrumentation.md)
- [ADR-004: Tanka for Infrastructure as Code](adr-004-tanka-iac.md)

## Integration Patterns

- [OBI Integration Patterns](obi-integration.md)
- [Alloy Sampling Strategies](alloy-sampling.md)
- [Trace-Log-Metric Correlation](correlation-patterns.md)
- [Cost Optimization Guide](cost-optimization.md)

## Experiments

See [OBI Experiments](obi-experiments.md) for proposed experiments:

1. Adaptive Tail-Based Sampling with SLO Integration
2. Network-Level Service Dependency Discovery
3. Database Query Performance Profiling
4. Cost-Optimized Multi-Region Observability
5. Canary Deployment Automated Rollback

## Design Principles

1. **Zero-Code Instrumentation First**: Use OBI for universal coverage
2. **Cost-Conscious**: Sampling, retention, and storage optimization
3. **Cloud-Native**: Kubernetes-native, scalable, resilient
4. **Vendor-Neutral**: Open standards (OTLP, Prometheus API)
5. **GitOps-Ready**: Declarative, version-controlled, reproducible

## Capacity Planning

### Small Deployment (Dev/Test)
- **Scale**: 10-50 services, <1M spans/day
- **Resources**:
  - OBI: 200m CPU, 256Mi RAM (per node)
  - Alloy: 1 CPU, 2Gi RAM
  - Tempo: 2 CPU, 4Gi RAM
  - Mimir: 2 CPU, 4Gi RAM
  - Loki: 2 CPU, 4Gi RAM
  - Grafana: 500m CPU, 1Gi RAM

### Medium Deployment (Staging)
- **Scale**: 50-200 services, 1M-10M spans/day
- **Resources**: Scale each component 3x

### Large Deployment (Production)
- **Scale**: 200+ services, >10M spans/day
- **Resources**: Scale each component 10x, use regional deployments

## Security Considerations

- **OBI**: Read-only eBPF probes, kernel-level isolation
- **Network**: mTLS between components
- **Secrets**: Kubernetes Secrets, external secret management
- **RBAC**: Least-privilege service accounts
- **Audit**: All configuration changes tracked in git

## Monitoring the Monitoring

- **Self-Observability**: Alloy, Tempo, Mimir, Loki emit their own metrics
- **Health Checks**: Kubernetes liveness/readiness probes
- **Alerting**: Critical alerts for data pipeline health
- **SLOs**: 99.9% ingestion availability, <1s query latency

## References

- [OpenTelemetry OBI Documentation](https://opentelemetry.io/blog/2025/obi-announcing-first-release/)
- [Grafana Alloy Documentation](https://grafana.com/docs/alloy/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
- [Mimir Documentation](https://grafana.com/docs/mimir/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Tanka Documentation](https://tanka.dev/)
