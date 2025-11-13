# MOP Component Libraries

Reusable Jsonnet libraries for deploying the Managed Observability Platform on Kubernetes.

## Quick Start

```jsonnet
// Import libraries
local alloy = import 'alloy.libsonnet';
local tempo = import 'tempo.libsonnet';
local mimir = import 'mimir.libsonnet';
local loki = import 'loki.libsonnet';
local grafana = import 'grafana.libsonnet';
local obi = import 'obi.libsonnet';
local config = import 'config.libsonnet';

// Select environment
local env = config.environments.dev;

// Deploy full stack
{
  obi: obi.new(env),
  alloy: alloy.new(env),
  tempo: tempo.new(env),
  mimir: mimir.new(env),
  loki: loki.new(env),
  grafana: grafana.new(env),
}
```

## Components

| Library | Description | Ports | Storage |
|---------|-------------|-------|---------|
| `alloy.libsonnet` | OpenTelemetry Collector | 4317 (gRPC), 4318 (HTTP), 12345 (UI) | 10Gi |
| `obi.libsonnet` | eBPF instrumentation (DaemonSet) | 8888 (metrics), 13133 (health) | N/A |
| `tempo.libsonnet` | Distributed tracing | 4317, 4318, 3200, 9095 | 50Gi |
| `mimir.libsonnet` | Metrics storage | 9009 (HTTP), 9095 (gRPC) | 100Gi |
| `loki.libsonnet` | Log aggregation | 3100 (HTTP), 9095 (gRPC) | 50Gi |
| `grafana.libsonnet` | Visualization | 3000 (HTTP) | - |

## Configuration

All components use `config.libsonnet` for:
- Component versions
- Environment-specific settings (dev/staging/production)
- Resource limits and replicas
- Storage configuration

### Environments

```jsonnet
local config = import 'config.libsonnet';

// Development - minimal resources, filesystem storage
config.environments.dev

// Staging - moderate resources, S3 storage
config.environments.staging

// Production - full resources, S3 storage, HA
config.environments.production
```

## Examples

### Full Observability Stack

```bash
jsonnet lib/examples/full-stack.jsonnet
```

Deploys: OBI + Alloy + Tempo + Mimir + Loki + Grafana

### Minimal Tracing Setup

```bash
jsonnet lib/examples/minimal.jsonnet
```

Deploys: Alloy + Tempo + Grafana

## Validation

Run the validation script to test all libraries:

```bash
./tests/validate-libraries.sh
```

## Resource Counts

### Full Stack (dev environment)
- **ConfigMaps**: 11
- **Services**: 16
- **StatefulSets**: 6
- **Deployments**: 3
- **DaemonSets**: 1 (OBI)
- **RBAC**: 3 resources (ServiceAccount, ClusterRole, ClusterRoleBinding)

### Minimal Stack (dev environment)
- **ConfigMaps**: 6
- **Services**: 7
- **StatefulSets**: 2
- **Deployments**: 3

## Documentation

See [docs/development/component-libraries.md](/Users/beengud/raibid-labs/mop/docs/development/component-libraries.md) for:
- Detailed library documentation
- Configuration options
- Usage patterns
- Troubleshooting

## Library Structure

```
lib/
├── config.libsonnet           # Central configuration
├── alloy.libsonnet            # OpenTelemetry Collector
├── obi.libsonnet              # eBPF instrumentation
├── tempo.libsonnet            # Distributed tracing
├── mimir.libsonnet            # Metrics storage
├── loki.libsonnet             # Log aggregation
├── grafana.libsonnet          # Visualization
├── kubernetes/                # Kubernetes helpers
│   ├── namespace.libsonnet
│   ├── rbac.libsonnet
│   ├── network.libsonnet
│   └── storage.libsonnet
└── examples/                  # Usage examples
    ├── full-stack.jsonnet     # Complete observability stack
    └── minimal.jsonnet        # Minimal tracing setup
```

## Usage with Tanka

```bash
# Create new environment
tk env add environments/dev \
  --namespace=observability-dev \
  --server=https://kubernetes.default.svc

# Show resources
tk show environments/dev

# Diff changes
tk diff environments/dev

# Apply to cluster
tk apply environments/dev
```

## Testing

```bash
# Validate Jsonnet syntax
jsonnet lib/examples/full-stack.jsonnet > /dev/null

# Generate YAML
jsonnet lib/examples/full-stack.jsonnet | yq eval -P

# Dry-run against Kubernetes
jsonnet lib/examples/full-stack.jsonnet | kubectl apply --dry-run=client -f -

# Run full validation suite
./tests/validate-libraries.sh
```

## Component Versions

Configured in `config.libsonnet`:

```jsonnet
versions:: {
  obi: '0.1.0',
  alloy: '1.0.0',
  tempo: '2.3.1',
  mimir: '5.3.0',
  loki: '5.41.0',
  grafana: '10.2.3',
}
```

## Features

- **Environment-driven**: Dev/staging/production configurations
- **Modular**: Use individual components or full stack
- **Validated**: All libraries tested and validated
- **Documented**: Comprehensive documentation and examples
- **Correlation**: Tempo ↔ Loki ↔ Mimir integration built-in
- **eBPF**: Automatic instrumentation with OBI

## Support

- **Documentation**: [docs/development/component-libraries.md](/Users/beengud/raibid-labs/mop/docs/development/component-libraries.md)
- **Validation**: `./tests/validate-libraries.sh`
- **Examples**: `lib/examples/`
