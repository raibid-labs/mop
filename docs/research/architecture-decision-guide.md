# Tanka Architecture & Decision Guide for MOP Project

## Executive Summary

This document provides architectural patterns, decision frameworks, and implementation strategies for the MOP (Monitoring Operations Platform) project using Tanka, Jsonnet, and Helm.

---

## 1. Architecture Decision Records (ADR)

### ADR-001: Use Tanka for Kubernetes Configuration Management

**Status**: Recommended

**Context**:
- Need to deploy complex Grafana observability stack
- Require environment-specific configurations (dev, staging, production)
- Want to leverage existing Helm charts without their limitations
- Need programmatic configuration with type safety

**Decision**:
Use Grafana Tanka with Jsonnet as the primary configuration management tool, consuming Helm charts where appropriate.

**Consequences**:
- ✅ Deep merging capabilities for customization beyond Helm values
- ✅ Type-safe, programmatic configuration
- ✅ Reusable libraries and abstractions
- ✅ Better suited for complex, interdependent services
- ✅ Native Grafana Labs support for their stack
- ⚠️ Learning curve for Jsonnet syntax
- ⚠️ Smaller community compared to Helm alone
- ⚠️ Need to maintain both Jsonnet and Helm chart versions

### ADR-002: Vendor Helm Charts Locally

**Status**: Adopted

**Context**:
- Need hermetic, reproducible builds
- Want to ensure exact versions across environments
- Require offline capability for air-gapped deployments

**Decision**:
Use `tk tool charts` to vendor all Helm charts into the repository under `charts/` directory.

**Consequences**:
- ✅ Reproducible builds across all environments
- ✅ No dependency on external chart repositories at runtime
- ✅ Version control of exact chart contents
- ✅ Faster CI/CD pipelines (no chart download step)
- ⚠️ Larger repository size
- ⚠️ Need process for chart updates
- ⚠️ Must track chart versions in chartfile.yaml

### ADR-003: Use Wrapped Library Pattern for Components

**Status**: Recommended

**Context**:
- Multiple environments with similar but different configurations
- Want to hide Helm complexity from end users
- Need consistent patterns across components

**Decision**:
Create wrapper libraries in `lib/` for each major component (Loki, Mimir, Tempo, Grafana) that expose simple interfaces.

**Consequences**:
- ✅ Consistent API across all components
- ✅ Easier for teams to consume
- ✅ Centralized defaults and best practices
- ✅ Simpler environment configurations
- ⚠️ Additional abstraction layer to maintain
- ⚠️ May need to expose advanced options as needed

### ADR-004: Environment Configuration Strategy

**Status**: Adopted

**Context**:
- Need to support dev, staging, and production environments
- Each environment has different resource requirements
- Want to minimize duplication while allowing flexibility

**Decision**:
Use a centralized `lib/config.libsonnet` with environment-specific defaults, overridden in individual environment `main.jsonnet` files.

**Consequences**:
- ✅ Single source of truth for environment differences
- ✅ Easy to compare environments
- ✅ Reduced duplication
- ✅ Type-safe environment selection
- ⚠️ All environments defined in one place (could be large)
- ⚠️ Need discipline to avoid environment-specific hacks

### ADR-005: Storage Backend Strategy

**Status**: Recommended

**Context**:
- Loki, Mimir, and Tempo all support object storage
- Need cost-effective long-term storage
- Want to separate compute from storage

**Decision**:
Use S3-compatible object storage for all components (Loki chunks/indexes, Mimir blocks, Tempo traces).

**Consequences**:
- ✅ Cost-effective for large data volumes
- ✅ Scales independently of compute
- ✅ Works across cloud providers (S3, GCS, MinIO)
- ✅ Durability and availability of cloud storage
- ⚠️ Requires S3 credentials management
- ⚠️ Network latency considerations
- ⚠️ Need proper bucket lifecycle policies

---

## 2. Directory Structure Decision Matrix

### Option A: Monolithic Structure (Recommended for MOP)

```
mop/
├── environments/
│   ├── dev/
│   ├── staging/
│   └── production/
├── lib/
│   ├── config.libsonnet
│   ├── loki.libsonnet
│   ├── mimir.libsonnet
│   ├── tempo.libsonnet
│   └── grafana.libsonnet
├── charts/
└── vendor/
```

**Pros**:
- Simple to understand
- All code in one repository
- Easy to ensure consistency
- Single deployment pipeline

**Cons**:
- Can become large over time
- All components version together

**Best for**: Single team, tightly coupled stack

### Option B: Component-Based Structure

```
mop/
├── environments/
├── components/
│   ├── loki/
│   │   ├── lib/
│   │   └── charts/
│   ├── mimir/
│   │   ├── lib/
│   │   └── charts/
│   ├── tempo/
│   │   ├── lib/
│   │   └── charts/
│   └── grafana/
│       ├── lib/
│       └── charts/
├── lib/
│   └── common.libsonnet
└── vendor/
```

**Pros**:
- Clear component boundaries
- Can version components independently
- Easier for multiple teams

**Cons**:
- More complex structure
- Potential for duplication
- Harder to ensure consistency

**Best for**: Multiple teams, independent component releases

### Option C: Multi-Repository

```
mop-core/          (shared libraries)
mop-loki/          (Loki deployment)
mop-mimir/         (Mimir deployment)
mop-tempo/         (Tempo deployment)
mop-grafana/       (Grafana deployment)
mop-environments/  (environment configurations)
```

**Pros**:
- Maximum separation of concerns
- Independent versioning and deployment
- Clear ownership boundaries

**Cons**:
- Coordination overhead
- Dependency management complexity
- Cross-repository changes difficult

**Best for**: Large organizations, separate teams per component

**Recommendation for MOP**: Start with Option A (Monolithic), migrate to Option B if team/scaling requires it.

---

## 3. Component Integration Patterns

### Pattern 1: Tightly Coupled (Recommended for Grafana Stack)

```jsonnet
// All components deployed together
{
  loki: loki.new(config),
  mimir: mimir.new(config),
  tempo: tempo.new(config),
  grafana: grafana.new(config)
    .withDatasource('Loki', 'loki', 'http://loki:3100')
    .withDatasource('Mimir', 'prometheus', 'http://mimir:8080')
    .withDatasource('Tempo', 'tempo', 'http://tempo:3200'),
}
```

**Use when**:
- Components are interdependent
- Deploy as a unit
- Shared configuration

### Pattern 2: Loosely Coupled

```jsonnet
// Components reference each other via service discovery
{
  loki: loki.new(config),
  mimir: mimir.new(config),
  tempo: tempo.new(config + {
    mimirUrl: 'http://mimir.monitoring.svc:8080',
  }),
  grafana: grafana.new(config + {
    datasourceDiscovery: true,
  }),
}
```

**Use when**:
- Independent deployment schedules
- Optional components
- Multi-cluster setups

### Pattern 3: Service Mesh Integration

```jsonnet
{
  loki: loki.new(config) + {
    deployment+: {
      spec+: { template+: { metadata+: {
        annotations+: {
          'sidecar.istio.io/inject': 'true',
        },
      }}},
    },
  },
  // ... similar for other components
}
```

**Use when**:
- Need mTLS between components
- Advanced traffic management
- Already using service mesh

---

## 4. Configuration Management Strategies

### Strategy A: Centralized Configuration (Recommended)

```jsonnet
// lib/config.libsonnet - single source of truth
{
  new(env):: {
    environment: env,
    namespace: 'monitoring',

    // All configuration here
    loki: { /* config */ },
    mimir: { /* config */ },
    tempo: { /* config */ },
    grafana: { /* config */ },
  },
}
```

**Pros**:
- Easy to see all environment differences
- Consistent patterns
- Type-safe

**Cons**:
- Can become large
- All components see all config

### Strategy B: Distributed Configuration

```jsonnet
// lib/loki/config.libsonnet
// lib/mimir/config.libsonnet
// lib/tempo/config.libsonnet
// Each component manages its own config
```

**Pros**:
- Component isolation
- Smaller files
- Clearer ownership

**Cons**:
- Harder to ensure consistency
- Potential duplication
- More files to maintain

### Strategy C: Layered Configuration

```jsonnet
// lib/config/base.libsonnet - common to all
// lib/config/observability.libsonnet - observability stack
// lib/config/environments.libsonnet - env overrides
```

**Pros**:
- Good separation of concerns
- Composable
- Flexible

**Cons**:
- More complex
- Need clear guidelines
- Override precedence rules

**Recommendation**: Start with Strategy A, refactor to Strategy C if complexity grows.

---

## 5. Helm Chart Integration Approaches

### Approach 1: Direct Template (Simple Components)

```jsonnet
local helm = tanka.helm.new(std.thisFile);

{
  simple: helm.template('name', './charts/name', {
    values: { /* basic overrides */ },
  }),
}
```

**Use for**:
- Simple charts with good defaults
- Minimal customization needed
- Standard deployments

### Approach 2: Template + Deep Merge (Complex Components)

```jsonnet
{
  complex: helm.template('name', './charts/name', {
    values: { /* values.yaml overrides */ },
  }) + {
    // Deep merge for fields not in values.yaml
    deployment+: { spec+: { template+: { metadata+: {
      annotations+: { 'custom': 'value' },
    }}}},
  },
}
```

**Use for**:
- Charts missing needed configuration
- Adding Kubernetes features not exposed
- Platform-specific requirements

### Approach 3: Wrapper Library (Reusable Components)

```jsonnet
// lib/component.libsonnet
{
  new(config):: {
    _config:: config,
    _helm: helm.template(/* ... */),

    deployment: self._helm.deployment_name,
    service: self._helm.service_name,

    // Helper methods
    withAnnotation(k, v):: self + {
      deployment+: { metadata+: { annotations+: { [k]: v } } },
    },
  },
}
```

**Use for**:
- Components used across environments
- Team-wide patterns
- Complex customization logic

---

## 6. Jsonnet Library Organization

### Small Project (<10 components)

```
lib/
├── k.libsonnet          (k8s helpers)
├── config.libsonnet     (all config)
└── helpers.libsonnet    (utility functions)
```

### Medium Project (10-30 components)

```
lib/
├── k.libsonnet
├── config/
│   ├── base.libsonnet
│   └── environments.libsonnet
├── components/
│   ├── loki.libsonnet
│   ├── mimir.libsonnet
│   └── ...
└── utils/
    ├── helpers.libsonnet
    └── mixins.libsonnet
```

### Large Project (>30 components)

```
lib/
├── k.libsonnet
├── config/
│   ├── base.libsonnet
│   ├── environments/
│   │   ├── dev.libsonnet
│   │   ├── staging.libsonnet
│   │   └── production.libsonnet
│   └── defaults/
│       └── observability.libsonnet
├── components/
│   ├── observability/
│   │   ├── loki/
│   │   │   ├── config.libsonnet
│   │   │   ├── deployment.libsonnet
│   │   │   └── service.libsonnet
│   │   ├── mimir/
│   │   └── tempo/
│   └── platform/
└── utils/
    ├── k8s.libsonnet
    ├── helm.libsonnet
    └── validation.libsonnet
```

---

## 7. Testing Strategy

### Level 1: Syntax Validation

```bash
# Jsonnet syntax check
jsonnetfmt --test lib/**/*.libsonnet

# Tanka evaluation test
tk eval environments/dev
tk eval environments/staging
tk eval environments/production
```

### Level 2: Schema Validation

```bash
# Kubernetes schema validation
tk export environments/dev /tmp/manifests
kubeval /tmp/manifests/*.yaml

# Or use kubeconform
kubeconform -summary /tmp/manifests/*.yaml
```

### Level 3: Diff Testing

```bash
# Compare against running cluster
tk diff environments/production

# Compare environments
diff <(tk show environments/dev) <(tk show environments/staging)
```

### Level 4: Integration Testing

```bash
# Deploy to test cluster
tk apply environments/dev --force

# Run smoke tests
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=300s
curl -f http://grafana.dev.local/api/health
```

### Level 5: Canary Testing

```jsonnet
// environments/production-canary/main.jsonnet
local production = import '../production/main.jsonnet';

production + {
  grafana+: {
    deployment+: {
      metadata+: { name: 'grafana-canary' },
      spec+: { replicas: 1 },
    },
  },
}
```

---

## 8. Deployment Strategies

### Strategy 1: All-at-Once (Development)

```bash
tk apply environments/dev
```

**Pros**: Fast, simple
**Cons**: Higher risk, potential downtime

### Strategy 2: Component-by-Component (Staging)

```bash
tk apply environments/staging --target=loki
# Verify
tk apply environments/staging --target=mimir
# Verify
tk apply environments/staging --target=tempo
# Verify
tk apply environments/staging --target=grafana
```

**Pros**: Controlled, easier rollback
**Cons**: Slower, more manual

### Strategy 3: Blue-Green (Production)

```jsonnet
// Deploy new version alongside old
{
  'loki-blue': loki.new(config),
  'loki-green': loki.new(config + { version: 'new' }),

  'loki-service': {
    // Switch traffic via selector
    spec+: { selector: { version: 'green' } },
  },
}
```

**Pros**: Zero downtime, easy rollback
**Cons**: Requires double resources temporarily

### Strategy 4: Canary (Production)

```jsonnet
{
  'loki-stable': loki.new(config + { replicas: 9 }),
  'loki-canary': loki.new(config + {
    replicas: 1,
    version: 'new',
  }),
}
```

**Pros**: Gradual rollout, real production testing
**Cons**: Complex routing, monitoring required

---

## 9. Operational Patterns

### Pattern: Configuration Drift Detection

```bash
#!/bin/bash
# check-drift.sh

# Get current cluster state
kubectl get all -n monitoring -o yaml > current-state.yaml

# Get desired state from Tanka
tk show environments/production > desired-state.yaml

# Compare
diff current-state.yaml desired-state.yaml
```

### Pattern: Secret Management

```jsonnet
// DO NOT commit secrets in jsonnet
local secrets = import 'secrets.jsonnet';  // git-ignored

// DO use external secret managers
{
  grafana: {
    envFrom: [{
      secretRef: { name: 'grafana-secrets' },
    }],
  },
}

// DO use sealed secrets or external secrets operator
```

### Pattern: Disaster Recovery

```bash
# Backup configuration
git tag -a "production-$(date +%Y%m%d)" -m "Production state"
git push origin --tags

# Backup state
kubectl get all -n monitoring -o yaml > backup-$(date +%Y%m%d).yaml

# Restore
git checkout production-20240101
tk apply environments/production
```

### Pattern: Multi-Cluster Management

```
environments/
├── us-west-2/
│   ├── dev/
│   ├── staging/
│   └── production/
├── eu-central-1/
│   ├── dev/
│   ├── staging/
│   └── production/
└── ap-southeast-1/
    ├── dev/
    ├── staging/
    └── production/
```

```jsonnet
// lib/clusters.libsonnet
{
  'us-west-2': { region: 'us-west-2', s3Endpoint: '...' },
  'eu-central-1': { region: 'eu-central-1', s3Endpoint: '...' },
}
```

---

## 10. Performance Optimization

### Optimization 1: Parallel Evaluation

```bash
# Evaluate environments in parallel
tk eval environments/dev &
tk eval environments/staging &
tk eval environments/production &
wait
```

### Optimization 2: Caching

```jsonnet
// Cache expensive computations
local cachedResult = std.native('cache')(
  'expensive-key',
  function() expensiveComputation()
);
```

### Optimization 3: Lazy Evaluation

```jsonnet
// Don't evaluate unless needed
local conditionalComponent =
  if config.featureEnabled then
    import 'expensive.libsonnet'
  else
    {};
```

---

## 11. Migration Path

### Phase 1: Setup (Week 1)
1. Initialize Tanka project
2. Setup basic directory structure
3. Install jsonnet-bundler dependencies
4. Configure dev environment

### Phase 2: Single Component (Week 2)
1. Start with Grafana (simplest)
2. Create wrapper library
3. Deploy to dev
4. Validate and iterate

### Phase 3: Add Observability Stack (Week 3-4)
1. Add Loki
2. Add Mimir or Prometheus
3. Add Tempo
4. Configure datasources in Grafana

### Phase 4: Multi-Environment (Week 5)
1. Create staging environment
2. Refactor common configuration
3. Test promotion flow
4. Document differences

### Phase 5: Production (Week 6)
1. Create production environment
2. Add proper resource sizing
3. Configure HA and persistence
4. Setup monitoring and alerting

### Phase 6: Optimization (Ongoing)
1. Refactor based on learnings
2. Add CI/CD automation
3. Implement advanced patterns
4. Document and share knowledge

---

## 12. Decision Checklist

Before implementing, answer these questions:

- [ ] What is the team's Jsonnet experience level?
- [ ] How many environments need to be supported?
- [ ] Are components deployed together or independently?
- [ ] What is the change frequency for each component?
- [ ] Is there a service mesh in place?
- [ ] What are the secret management requirements?
- [ ] Are there multi-cluster requirements?
- [ ] What is the rollback strategy?
- [ ] How will configuration drift be detected?
- [ ] What are the disaster recovery requirements?

---

## 13. Recommended Stack for MOP

Based on research and best practices:

```yaml
Foundation:
  - Tanka 0.25+
  - Jsonnet 0.20+
  - Kubernetes 1.28+
  - Helm 3.13+

Core Components:
  - Grafana 10.2+
  - Loki 2.9+ (microservices mode)
  - Mimir 2.10+ (distributed mode)
  - Tempo 2.3+

Storage:
  - S3-compatible object storage
  - Fast SSDs for write path
  - Standard storage for general use

Structure:
  - Monolithic repository (Option A)
  - Centralized configuration (Strategy A)
  - Wrapped libraries (Approach 3)
  - Component-by-component deployment

CI/CD:
  - GitOps with ArgoCD
  - Automated testing
  - Canary deployments to production
  - Automated rollback on failure
```

---

## Conclusion

The MOP project should:

1. **Start simple** with monolithic structure
2. **Use wrapped libraries** for reusability
3. **Centralize configuration** for consistency
4. **Vendor Helm charts** for reproducibility
5. **Test thoroughly** before production
6. **Deploy incrementally** to reduce risk
7. **Monitor continuously** for drift
8. **Document extensively** for team knowledge

This approach balances flexibility, maintainability, and operational safety while leveraging the strengths of Tanka, Jsonnet, and Helm.
