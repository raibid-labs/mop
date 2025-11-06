# MOP - Managed Observability Platform

A reference implementation for a modern observability stack using OpenTelemetry Backend Initiative (OBI), Grafana, and cloud-native components.

## 🎯 Project Overview

MOP provides a production-ready observability platform featuring:

- **OpenTelemetry Backend Initiative (OBI)**: Zero-code, eBPF-based instrumentation with <1% CPU overhead
- **Grafana Stack**: Unified visualization and alerting
- **Grafana Alloy**: Advanced telemetry pipeline with sampling and routing
- **Tempo**: Distributed tracing backend with cost-efficient object storage
- **Mimir**: Long-term metrics storage (Prometheus-compatible, no Prometheus)
- **Loki**: Log aggregation with trace correlation
- **Tanka**: Infrastructure as code with Jsonnet + Helm

## 🏗️ Architecture

```
┌─────────────────┐
│   Application   │
│   (Any Lang)    │
└────────┬────────┘
         │
    ╔════▼════════════════════════════╗
    ║  OBI (eBPF Instrumentation)    ║
    ║  - HTTP/gRPC/SQL/Redis/Kafka   ║
    ║  - <1% CPU overhead            ║
    ╚════╤════════════════════════════╝
         │ OTLP
    ╔════▼═══════════════════════════╗
    ║  Grafana Alloy                 ║
    ║  - Sampling & Routing          ║
    ║  - Cost Optimization           ║
    ╚════╤═══════════════╤═══════════╝
         │               │
    ╔════▼═════╗    ╔════▼═════╗
    ║  Tempo   ║    ║  Mimir   ║
    ║ (Traces) ║    ║ (Metrics)║
    ╚══════════╝    ╚══════════╝
         │               │
    ╔════▼═══════════════▼═══════════╗
    ║        Loki (Logs)             ║
    ╚════╤═══════════════════════════╝
         │
    ╔════▼═══════════════════════════╗
    ║  Grafana (Visualization)       ║
    ║  - Stateless, Auth Disabled    ║
    ╚════════════════════════════════╝
```

## 🚀 Quick Start

```bash
# Install dependencies
just install

# Initialize Tanka
just init

# Deploy to dev environment
just deploy dev

# View logs
just logs alloy

# Access Grafana
just grafana-port-forward
open http://localhost:3000
```

## 📁 Repository Structure

```
mop/
├── docs/                      # Documentation
│   ├── architecture/          # Architecture Decision Records (ADRs)
│   ├── workstreams/           # Parallel workstream issues
│   ├── agents/                # Agent coordination configs
│   └── research/              # Research findings
├── environments/              # Tanka environments
│   ├── dev/                   # Development environment
│   ├── staging/               # Staging environment
│   └── production/            # Production environment
├── lib/                       # Jsonnet libraries
│   ├── config.libsonnet       # Centralized configuration
│   ├── alloy.libsonnet        # Alloy configuration
│   ├── obi.libsonnet          # OBI DaemonSet configuration
│   ├── tempo.libsonnet        # Tempo distributed tracing
│   ├── mimir.libsonnet        # Mimir metrics storage
│   ├── loki.libsonnet         # Loki log aggregation
│   └── grafana.libsonnet      # Grafana dashboards
├── charts/                    # Vendored Helm charts
├── vendor/                    # Jsonnet dependencies
├── scripts/                   # Automation scripts
│   └── nu/                    # Nushell scripts
├── tests/                     # Integration tests
├── Tiltfile                   # Local development with Tilt
├── justfile                   # Common commands
└── tanka.yaml                 # Tanka configuration
```

## 🛠️ Technology Stack

| Component | Purpose | Why No Prometheus? |
|-----------|---------|-------------------|
| **OBI** | eBPF instrumentation | Zero-code, universal coverage |
| **Grafana Alloy** | Telemetry pipeline | Advanced sampling & routing |
| **Tempo** | Distributed tracing | Cost-efficient, object storage |
| **Mimir** | Metrics storage | **Prometheus-compatible API, better for scale** |
| **Loki** | Log aggregation | Trace-log correlation |
| **Grafana** | Visualization | Unified observability UX |
| **Tanka** | Infrastructure as Code | Jsonnet + Helm flexibility |

**Why Mimir instead of Prometheus?**
- Horizontally scalable (Prometheus is single-instance)
- Object storage backend (cheaper than local disks)
- Multi-tenancy built-in
- Better retention policies
- Still exposes Prometheus-compatible API for querying

## 🧪 OBI Experiments

See [`docs/architecture/obi-experiments.md`](docs/architecture/obi-experiments.md) for detailed experiment proposals:

1. **Adaptive Tail-Based Sampling**: Dynamic sampling based on SLO breaches (90% cost reduction)
2. **Network Service Discovery**: Auto-generate dependency graphs from traffic
3. **Database Query Profiling**: Identify slow SQL without instrumentation
4. **Multi-Region Cost Optimization**: Regional traces, global metrics (79% cost reduction)
5. **Canary Automated Rollback**: OBI metrics drive Argo Rollouts quality gates

## 📋 Parallel Workstreams

This project is organized into parallel workstreams that can be worked on concurrently:

- [Workstream 1: Infrastructure Foundation](docs/workstreams/01-infrastructure-foundation.md)
- [Workstream 2: OBI Integration](docs/workstreams/02-obi-integration.md)
- [Workstream 3: Grafana Stack](docs/workstreams/03-grafana-stack.md)
- [Workstream 4: Tanka Configuration](docs/workstreams/04-tanka-configuration.md)
- [Workstream 5: Development Tools](docs/workstreams/05-development-tools.md)
- [Workstream 6: OBI Experiments](docs/workstreams/06-obi-experiments.md)

## 🤖 Agent Coordination

See [`docs/agents/coordination.md`](docs/agents/coordination.md) for agent roles and collaboration patterns.

## 🔧 Development

### Prerequisites

- Kubernetes cluster (kind, minikube, or cloud)
- Tanka (`brew install tanka`)
- jsonnet-bundler (`brew install jsonnet-bundler`)
- Tilt (`brew install tilt`)
- just (`brew install just`)
- nushell (`brew install nushell`)

### Local Development Workflow

```bash
# 1. Start local Kubernetes cluster
just cluster-up

# 2. Start Tilt (hot reload)
tilt up

# 3. Make changes to Jsonnet files
# Tilt automatically reloads

# 4. Run tests
just test

# 5. Apply to dev environment
just deploy dev
```

## 📖 Documentation

- [Architecture Overview](docs/architecture/README.md)
- [Alloy Operator Decision](docs/architecture/adr-001-alloy-operator.md)
- [OBI Integration Patterns](docs/architecture/obi-integration.md)
- [Tanka Best Practices](docs/research/tanka-helm-patterns.md)
- [Cost Optimization Guide](docs/architecture/cost-optimization.md)

## 🎓 Learning Resources

- [OBI Comprehensive Research](docs/research/obi-comprehensive-research.md)
- [Grafana Stack Examples](docs/research/grafana-stack-examples.md)
- [Tanka Helm Patterns](docs/research/tanka-helm-patterns.md)

## 📊 Monitoring & Alerting

Default dashboards are provisioned automatically:

- **OBI Overview**: eBPF instrumentation health
- **Alloy Pipeline**: Sampling rates, throughput, errors
- **Tempo**: Trace ingestion, query latency
- **Mimir**: Metrics cardinality, ingestion rate
- **Loki**: Log volume, query performance
- **SLO Dashboard**: Service-level objectives tracking

## 🔐 Security

- Grafana: Stateless deployment, auth disabled (for internal use)
- OBI: Read-only eBPF probes, no data modification
- Secrets: Managed via Kubernetes Secrets (not in git)
- Network policies: Least-privilege access

## 🤝 Contributing

1. Create a workstream issue in `docs/workstreams/`
2. Use agent coordination patterns from `docs/agents/`
3. Follow Tanka best practices
4. Ensure tests pass
5. Update documentation

## 📝 License

MIT License - see LICENSE file

## 🙋 Support

- Issues: File in GitHub Issues with workstream label
- Docs: See `docs/` directory
- Examples: See `docs/research/` for detailed guides

---

**Status**: 🏗️ Initial Setup Phase

**Next Steps**: See [Workstream 1: Infrastructure Foundation](docs/workstreams/01-infrastructure-foundation.md)
