# Kubernetes Infrastructure Setup

## Overview
This document describes the foundational Kubernetes infrastructure for the MOP (Metrics, Observability, Practices) platform. The infrastructure is defined using Jsonnet and managed with Tanka for GitOps-based deployments.

## Architecture

### Directory Structure
```
mop/
├── lib/
│   ├── kubernetes/           # Kubernetes resource libraries
│   │   ├── namespace.libsonnet
│   │   ├── rbac.libsonnet
│   │   ├── storage.libsonnet
│   │   └── network.libsonnet
│   └── config.libsonnet      # Central configuration
├── environments/
│   ├── dev/                  # Development environment
│   ├── staging/              # Staging environment
│   └── production/          # Production environment
└── vendor/                  # Vendored jsonnet dependencies
```

## Components

### 1. Namespaces
- **Purpose**: Logical isolation of observability components
- **Configuration**: Defined in `lib/kubernetes/namespace.libsonnet`
- **Labels**:
  - `mop.io/managed: true` - Indicates Tanka management
  - `mop.io/component: infrastructure` - Component type
  - `mop.io/version: 1.0.0` - Platform version
  - `environment: <env>` - Environment identifier

### 2. RBAC (Role-Based Access Control)
- **Purpose**: Fine-grained access control for components
- **Configuration**: Defined in `lib/kubernetes/rbac.libsonnet`
- **Components**:
  - **ServiceAccounts**: Individual identity for each component
  - **ClusterRoles**: Cluster-wide permissions
  - **Roles**: Namespace-scoped permissions
  - **RoleBindings**: Associate roles with service accounts

#### Service Accounts Created:
- `obi-collector` - eBPF instrumentation (privileged)
- `alloy` - Metrics/logs/traces collector
- `tempo` - Distributed tracing backend
- `mimir` - Metrics storage
- `loki` - Log aggregation
- `grafana` - Visualization dashboard

#### Permissions:
- **OBI**: Cluster-wide read access to nodes, pods, services (required for eBPF)
- **Alloy**: Cluster-wide read access for scraping metrics
- **Tempo/Mimir/Loki**: Namespace-scoped access to config and storage
- **Grafana**: Read-only access to configurations

### 3. Storage Classes
- **Purpose**: Define storage tiers for persistent data
- **Configuration**: Defined in `lib/kubernetes/storage.libsonnet`
- **Classes**:
  - **mop-standard**: Standard performance storage for dev/staging
  - **mop-fast-ssd**: High-performance SSD for production

#### Storage Configuration by Environment:
| Environment | Storage Class | Provisioner | Reclaim Policy |
|------------|--------------|-------------|----------------|
| Development | mop-standard | kubernetes.io/gce-pd | Delete |
| Staging | mop-standard | kubernetes.io/gce-pd | Delete |
| Production | mop-fast-ssd | kubernetes.io/gce-pd | Retain |

**Note**: Provisioner configuration varies by cloud provider:
- GKE: `kubernetes.io/gce-pd`
- EKS: `ebs.csi.aws.com`
- AKS: `disk.csi.azure.com`

### 4. Network Policies
- **Purpose**: Microsegmentation and traffic control
- **Configuration**: Defined in `lib/kubernetes/network.libsonnet`
- **Policies**:
  - **default-deny-all**: Deny all ingress/egress by default
  - **Component-specific policies**: Allow only required communication

#### Network Flow Matrix:
| From | To | Port | Protocol | Purpose |
|------|-----|------|----------|---------|
| Grafana | Tempo | 3200 | TCP | Query traces |
| Grafana | Mimir | 9009 | TCP | Query metrics |
| Grafana | Loki | 3100 | TCP | Query logs |
| Grafana | OBI | 9090 | TCP | Query eBPF metrics |
| Alloy | Tempo | 9095 | TCP | Send traces |
| Alloy | Mimir | 9009 | TCP | Send metrics |
| Alloy | Loki | 3100 | TCP | Send logs |
| * | Alloy | 4317/4318 | TCP | OTLP ingestion |
| All | kube-dns | 53 | UDP/TCP | DNS resolution |

## Environment Configuration

### Development
- **Namespace**: `observability-dev`
- **Resources**: Minimal (testing and development)
- **Storage**: Local filesystem
- **Replicas**: Single instance per component

### Staging
- **Namespace**: `observability-staging`
- **Resources**: Moderate
- **Storage**: S3-compatible object storage
- **Replicas**: 2-3 instances per component

### Production
- **Namespace**: `observability`
- **Resources**: Full allocation
- **Storage**: S3 with regional replication
- **Replicas**: 3+ instances with HA

## Deployment

### Prerequisites
1. Install required tools:
```bash
brew install tanka jsonnet-bundler
```

2. Configure kubectl context:
```bash
kubectl config use-context <cluster-context>
```

### Deploy Infrastructure

#### Development Environment:
```bash
cd environments/dev
tk apply main.jsonnet
```

#### Staging Environment:
```bash
cd environments/staging
tk apply main.jsonnet
```

#### Production Environment:
```bash
cd environments/production
tk apply main.jsonnet --dangerous-auto-approve=false
```

### Validation

Verify namespace creation:
```bash
kubectl get namespaces | grep observability
```

Verify RBAC:
```bash
kubectl get serviceaccounts -n observability-dev
kubectl get clusterroles | grep mop
kubectl get rolebindings -n observability-dev
```

Verify storage classes:
```bash
kubectl get storageclasses | grep mop
```

Verify network policies:
```bash
kubectl get networkpolicies -n observability-dev
```

## Maintenance

### Updating Configuration
1. Modify the appropriate jsonnet library in `lib/kubernetes/`
2. Validate changes: `tk eval main.jsonnet`
3. Review diff: `tk diff main.jsonnet`
4. Apply changes: `tk apply main.jsonnet`

### Adding New Components
1. Update RBAC in `lib/kubernetes/rbac.libsonnet`
2. Add network policy in `lib/kubernetes/network.libsonnet`
3. Configure storage if needed in `lib/kubernetes/storage.libsonnet`
4. Update environment configurations

### Troubleshooting

Common issues and solutions:

**Issue**: Tanka evaluation fails
```bash
# Check jsonnet syntax
jsonnet --version
tk eval main.jsonnet 2>&1 | head -20
```

**Issue**: RBAC permission denied
```bash
# Test permissions
kubectl auth can-i list pods --as=system:serviceaccount:<namespace>:<sa-name>
```

**Issue**: Network connectivity problems
```bash
# Check network policies
kubectl describe networkpolicy <policy-name> -n <namespace>
```

## Security Considerations

1. **Principle of Least Privilege**: Each component has minimal required permissions
2. **Network Segmentation**: Default deny with explicit allow rules
3. **Storage Encryption**: Enable encryption at rest for production
4. **Audit Logging**: Enable Kubernetes audit logging for RBAC events
5. **Secret Management**: Use External Secrets Operator or Sealed Secrets

## Next Steps

With the infrastructure foundation in place, the following workstreams can proceed:
- **Workstream 2**: Deploy OBI and Alloy collectors
- **Workstream 3**: Deploy Tempo, Mimir, Loki storage backends
- **Workstream 4**: Deploy Grafana and configure dashboards

## References
- [Tanka Documentation](https://tanka.dev/)
- [Jsonnet Language Guide](https://jsonnet.org/learning/tutorial.html)
- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)