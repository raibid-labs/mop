# MOP Component Libraries

This document explains the Tanka/Jsonnet component libraries for the Managed Observability Platform (MOP).

## Overview

The MOP component libraries provide reusable Jsonnet modules for deploying the complete observability stack on Kubernetes. Each component is designed to work independently or as part of the full stack.

## Library Structure

```
lib/
├── config.libsonnet         # Central configuration
├── alloy.libsonnet          # Grafana Alloy (OTLP collector)
├── obi.libsonnet            # eBPF-based instrumentation
├── tempo.libsonnet          # Distributed tracing
├── mimir.libsonnet          # Metrics storage
├── loki.libsonnet           # Log aggregation
├── grafana.libsonnet        # Visualization
└── examples/
    ├── full-stack.jsonnet   # Complete stack deployment
    └── minimal.jsonnet      # Minimal tracing-only setup
```

## Component Libraries

### 1. Alloy (`lib/alloy.libsonnet`)

**Purpose**: OpenTelemetry Collector for receiving and routing telemetry data.

**Resources Created**:
- ConfigMap with Alloy configuration
- StatefulSet for Alloy instances
- Service for OTLP endpoints

**Ports**:
- 4317: OTLP gRPC
- 4318: OTLP HTTP
- 12345: Alloy admin UI

**Usage**:
```jsonnet
local alloy = import 'lib/alloy.libsonnet';
local config = import 'lib/config.libsonnet';

{
  alloy: alloy.new(config.environments.dev),
}
```

### 2. OBI (`lib/obi.libsonnet`)

**Purpose**: eBPF-based automatic instrumentation for protocol-level observability.

**Resources Created**:
- ServiceAccount with cluster permissions
- ClusterRole and ClusterRoleBinding
- ConfigMap for OBI configuration
- DaemonSet (runs on every node)

**Supported Protocols**:
- HTTP/HTTPS
- gRPC
- SQL (PostgreSQL, MySQL)
- Redis
- Kafka

**Usage**:
```jsonnet
local obi = import 'lib/obi.libsonnet';
local config = import 'lib/config.libsonnet';

{
  obi: obi.new(config.environments.dev),
}
```

**Note**: OBI requires privileged access for eBPF operations.

### 3. Tempo (`lib/tempo.libsonnet`)

**Purpose**: Distributed tracing backend for storing and querying traces.

**Resources Created**:
- ConfigMap with Tempo configuration
- StatefulSet for Tempo instances
- Service for HTTP/gRPC access

**Ports**:
- 3200: HTTP API
- 9095: gRPC
- 4317: OTLP gRPC receiver
- 4318: OTLP HTTP receiver

**Configuration**:
- Trace retention: 30 days (configurable via `config.libsonnet`)
- Storage: Filesystem (dev) or S3 (staging/prod)
- Max trace size: 5MB

**Usage**:
```jsonnet
local tempo = import 'lib/tempo.libsonnet';
local config = import 'lib/config.libsonnet';

{
  tempo: tempo.new(config.environments.production),
}
```

### 4. Mimir (`lib/mimir.libsonnet`)

**Purpose**: Long-term metrics storage with Prometheus compatibility.

**Resources Created**:
- ConfigMap with Mimir configuration
- StatefulSet for Mimir instances
- Service for HTTP/gRPC access

**Ports**:
- 9009: HTTP API
- 9095: gRPC

**Configuration**:
- Metrics retention: 365 days
- Storage: Filesystem (dev) or S3 (staging/prod)
- Replication factor: min(3, replicas)

**Usage**:
```jsonnet
local mimir = import 'lib/mimir.libsonnet';
local config = import 'lib/config.libsonnet';

{
  mimir: mimir.new(config.environments.staging),
}
```

### 5. Loki (`lib/loki.libsonnet`)

**Purpose**: Log aggregation and querying with labels.

**Resources Created**:
- ConfigMap with Loki configuration
- StatefulSet for Loki instances
- Service for HTTP/gRPC access

**Ports**:
- 3100: HTTP API
- 9095: gRPC

**Configuration**:
- Log retention: 30 days
- Storage: Filesystem (dev) or S3 (staging/prod)
- Compression: Enabled

**Usage**:
```jsonnet
local loki = import 'lib/loki.libsonnet';
local config = import 'lib/config.libsonnet';

{
  loki: loki.new(config.environments.dev),
}
```

### 6. Grafana (`lib/grafana.libsonnet`)

**Purpose**: Visualization and dashboards for all telemetry data.

**Resources Created**:
- ConfigMap for Grafana configuration
- ConfigMap for datasource provisioning
- Deployment for Grafana instances
- Service (LoadBalancer type by default)

**Ports**:
- 3000: HTTP UI

**Pre-configured Datasources**:
- Tempo (with logs correlation)
- Mimir (Prometheus-compatible)
- Loki (with trace correlation)

**Usage**:
```jsonnet
local grafana = import 'lib/grafana.libsonnet';
local config = import 'lib/config.libsonnet';

{
  grafana: grafana.new(config.environments.dev),
}
```

## Configuration

All components reference `lib/config.libsonnet` for environment-specific settings:

```jsonnet
{
  environments:: {
    dev: { /* dev config */ },
    staging: { /* staging config */ },
    production: { /* production config */ },
  },
  versions:: { /* component versions */ },
  // ... more config
}
```

### Environment Configuration

Each environment defines:
- **namespace**: Kubernetes namespace
- **domain**: Base domain for services
- **replicas**: Number of replicas per component
- **resources**: CPU/memory requests and limits
- **storage**: Storage class and backend type

### Customizing Configuration

To customize component versions:

```jsonnet
// lib/config.libsonnet
{
  versions:: {
    alloy: '1.1.0',  // Update version here
    tempo: '2.4.0',
    // ...
  },
}
```

To adjust resource limits:

```jsonnet
// lib/config.libsonnet
{
  environments:: {
    dev: {
      resources: {
        tempo: {
          requests: { cpu: '200m', memory: '1Gi' },
          limits: { cpu: '2', memory: '4Gi' },
        },
      },
    },
  },
}
```

## Examples

### Full Stack Deployment

Deploy the complete observability platform:

```jsonnet
// environments/dev/main.jsonnet
local alloy = import '../../lib/alloy.libsonnet';
local obi = import '../../lib/obi.libsonnet';
local tempo = import '../../lib/tempo.libsonnet';
local mimir = import '../../lib/mimir.libsonnet';
local loki = import '../../lib/loki.libsonnet';
local grafana = import '../../lib/grafana.libsonnet';
local config = import '../../lib/config.libsonnet';

local env = config.environments.dev;

{
  obi: obi.new(env),
  alloy: alloy.new(env),
  tempo: tempo.new(env),
  mimir: mimir.new(env),
  loki: loki.new(env),
  grafana: grafana.new(env),
}
```

### Minimal Tracing Setup

Deploy just tracing components:

```jsonnet
local alloy = import 'lib/alloy.libsonnet';
local tempo = import 'lib/tempo.libsonnet';
local grafana = import 'lib/grafana.libsonnet';
local config = import 'lib/config.libsonnet';

local env = config.environments.dev;

{
  alloy: alloy.new(env),
  tempo: tempo.new(env),
  grafana: grafana.new(env),
}
```

### Environment-Specific Deployment

Use different environments:

```jsonnet
// For production
local config = import 'lib/config.libsonnet';
local env = config.environments.production;

// Components will use production resource limits and S3 storage
{
  tempo: (import 'lib/tempo.libsonnet').new(env),
  mimir: (import 'lib/mimir.libsonnet').new(env),
  loki: (import 'lib/loki.libsonnet').new(env),
}
```

## Testing and Validation

### Validate Jsonnet

```bash
# Test that libraries compile
jsonnet lib/examples/full-stack.jsonnet

# Validate specific component
jsonnet -e 'local alloy = import "lib/alloy.libsonnet"; local config = import "lib/config.libsonnet"; alloy.new(config.environments.dev)'
```

### Generate Kubernetes YAML

```bash
# Generate YAML for review
jsonnet lib/examples/full-stack.jsonnet | yq eval -P

# Count resources
jsonnet lib/examples/full-stack.jsonnet | yq eval '.[] | .[] | .kind' | sort | uniq -c
```

### Dry-Run Deployment

```bash
# Test against Kubernetes API
jsonnet lib/examples/full-stack.jsonnet | kubectl apply --dry-run=client -f -

# Server-side validation
jsonnet lib/examples/full-stack.jsonnet | kubectl apply --dry-run=server -f -
```

## Library Design Patterns

### 1. Function-Based Interface

Each library exports a `new(config)` function:

```jsonnet
{
  new(envConfig):: {
    // Returns Kubernetes resources
  },
}
```

### 2. Environment-Driven Configuration

All libraries accept environment configuration:

```jsonnet
local env = config.environments.dev;
alloy.new(env)  // Uses dev resources and settings
```

### 3. Conditional Resource Generation

Resources adapt to environment:

```jsonnet
storage: {
  [if envConfig.storage.type == 'filesystem' then 'local']: { /* ... */ },
  [if envConfig.storage.type == 's3' then 's3']: { /* ... */ },
}
```

### 4. Common Labels

All resources include common labels:

```jsonnet
labels: config.commonLabels + { component: 'tempo' }
```

## Best Practices

1. **Always import config.libsonnet**: Ensures consistent versioning and configuration
2. **Use environment configs**: Don't hardcode resource limits or replicas
3. **Test locally first**: Validate Jsonnet before deploying
4. **Version control**: Commit generated YAML or use Tanka's GitOps export
5. **Document changes**: Update this file when adding new features

## Troubleshooting

### Import Errors

```bash
# Error: import "lib/config.libsonnet" not found
# Solution: Ensure you're in the project root or use -J flag
jsonnet -J /Users/beengud/raibid-labs/mop lib/examples/full-stack.jsonnet
```

### Missing Fields

```bash
# Error: field does not exist: storage
# Solution: Check config.libsonnet for required fields
jsonnet -e 'import "lib/config.libsonnet"' | yq eval '.environments.dev'
```

### Resource Validation

```bash
# Test specific component
jsonnet -e '(import "lib/tempo.libsonnet").new((import "lib/config.libsonnet").environments.dev)'
```

## Next Steps

- **Tanka Integration**: Use these libraries with Tanka environments
- **Helm Charts**: Consider wrapping in Helm for easier deployment
- **Custom Dashboards**: Add Grafana dashboard provisioning
- **Alerting**: Integrate with Grafana Alerting or Prometheus Alertmanager
- **Multi-Cluster**: Extend for federated observability

## References

- [Jsonnet Language Reference](https://jsonnet.org/)
- [Tanka Documentation](https://tanka.dev/)
- [Grafana Alloy](https://grafana.com/docs/alloy/)
- [Grafana Tempo](https://grafana.com/docs/tempo/)
- [Grafana Mimir](https://grafana.com/docs/mimir/)
- [Grafana Loki](https://grafana.com/docs/loki/)
- [Grafana OSS](https://grafana.com/grafana/)
