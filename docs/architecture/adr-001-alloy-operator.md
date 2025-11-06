# ADR-001: Use Alloy Operator for Production, Standalone for Dev

## Status

**ACCEPTED**

## Context

Grafana Alloy can be deployed in two ways:
1. **Alloy Operator**: Kubernetes operator that manages Alloy instances declaratively
2. **Standalone Helm**: Direct Helm chart deployment of Alloy

We need to decide which approach to use for different environments.

## Decision

**Use Alloy Operator for production, standalone Helm for dev/testing.**

### Production: Alloy Operator

**Rationale:**
- Declarative management via CRDs (Custom Resource Definitions)
- Auto-configuration based on cluster topology
- Built-in high availability and failover
- Easier multi-environment management
- Better GitOps integration (ArgoCD, Flux)
- Operator handles upgrades and scaling

**Example CRD:**
```yaml
apiVersion: monitoring.grafana.com/v1alpha1
kind: GrafanaAgent
metadata:
  name: alloy-production
spec:
  mode: flow
  config:
    receivers:
      otlp:
        protocols:
          grpc:
            endpoint: 0.0.0.0:4317
    processors:
      tail_sampling:
        policies:
          - name: errors
            type: status_code
            status_code: {status_codes: [ERROR]}
    exporters:
      otlp:
        endpoint: tempo:4317
```

### Dev/Testing: Standalone Helm

**Rationale:**
- Simpler setup, fewer moving parts
- Direct control over configuration
- Faster iteration cycles
- Easier to debug and troubleshoot
- No operator dependency

**Example Helm values:**
```yaml
alloy:
  configMap:
    content: |
      otelcol.receiver.otlp "default" {
        grpc {
          endpoint = "0.0.0.0:4317"
        }
      }
      otelcol.exporter.otlp "tempo" {
        client {
          endpoint = "tempo:4317"
        }
      }
```

## Deployment Modes by Environment

| Environment | Deployment | Why |
|-------------|-----------|-----|
| **Local (kind/minikube)** | Standalone Helm | Fast iteration, simple debugging |
| **Dev** | Standalone Helm | Easier experimentation |
| **Staging** | Alloy Operator | Test production topology |
| **Production** | Alloy Operator | GitOps, HA, auto-configuration |

## Consequences

### Positive
- Best tool for each environment
- Production benefits from operator automation
- Dev retains flexibility and simplicity
- Easier onboarding (start with Helm, graduate to operator)

### Negative
- Two deployment mechanisms to maintain
- Different configuration formats (Helm values vs CRDs)
- Potential drift between environments

### Mitigation
- Use Tanka to abstract differences
- Maintain parallel configs with shared base
- CI/CD validates both paths
- Documentation for both approaches

## Implementation

### Dev Environment
```bash
# Tanka environments/dev/main.jsonnet
local helm = import 'helm-util/helm.libsonnet';

{
  alloy: helm.template('alloy', '../../charts/alloy', {
    namespace: 'observability',
    values: {
      // Helm values for standalone deployment
    }
  })
}
```

### Production Environment
```bash
# Tanka environments/production/main.jsonnet
{
  alloyOperator: {
    apiVersion: 'monitoring.grafana.com/v1alpha1',
    kind: 'GrafanaAgent',
    metadata: {
      name: 'alloy-production',
      namespace: 'observability',
    },
    spec: {
      // Operator CRD spec
    }
  }
}
```

## References

- [Grafana Alloy Operator Documentation](https://grafana.com/docs/alloy/latest/operator/)
- [Helm Chart Documentation](https://grafana.com/docs/alloy/latest/setup/install/helm/)
- [Tanka Best Practices](../research/tanka-helm-patterns.md)

## Related Decisions

- ADR-004: Tanka for Infrastructure as Code (enables this hybrid approach)

---

**Date**: 2025-01-06
**Author**: MOP Architecture Team
**Reviewers**: Platform Engineering, SRE
