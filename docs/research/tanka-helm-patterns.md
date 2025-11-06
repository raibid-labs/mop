# Tanka + Helm Integration Patterns: Research Findings

## Executive Summary

This document contains research findings on using Grafana Tanka with Helm charts, specifically focused on patterns applicable to deploying Grafana observability stack components (Loki, Mimir, Tempo, Prometheus, Grafana).

**Note**: The specific `gudo11y/mop-core` repository was not found in search results. This research provides industry-standard patterns from Grafana Labs and the Tanka community.

---

## 1. Directory Structure Best Practices

### Standard Tanka Project Layout

```
project-root/
├── jsonnetfile.json          # Direct dependency declarations
├── jsonnetfile.lock.json     # Locked versions for reproducibility
├── tkrc.yaml                 # Tanka root identifier (optional)
├── chartfile.yaml            # Helm chart dependencies (optional)
│
├── environments/             # Deployment targets
│   ├── dev/
│   │   ├── main.jsonnet     # Entry point for dev environment
│   │   └── spec.json        # Cluster config (API server, namespace)
│   ├── staging/
│   │   ├── main.jsonnet
│   │   └── spec.json
│   └── production/
│       ├── main.jsonnet
│       └── spec.json
│
├── lib/                      # Project-local reusable libraries
│   ├── k.libsonnet          # Kubernetes helpers (auto-generated)
│   ├── grafana/             # Custom Grafana configs
│   │   └── dashboards.libsonnet
│   ├── loki/                # Loki-specific helpers
│   │   └── config.libsonnet
│   └── common.libsonnet     # Shared utilities
│
├── vendor/                   # External dependencies (managed by jb)
│   └── github.com/
│       ├── grafana/
│       │   └── jsonnet-libs/
│       │       ├── tanka-util/
│       │       └── ksonnet-util/
│       └── jsonnet-libs/
│           └── k8s-libsonnet/
│
└── charts/                   # Vendored Helm charts
    ├── grafana/
    ├── prometheus/
    └── loki-stack/
```

### Environment spec.json Example

```json
{
  "apiVersion": "tanka.dev/v1alpha1",
  "kind": "Environment",
  "metadata": {
    "name": "environments/production"
  },
  "spec": {
    "apiServer": "https://kubernetes.production.example.com",
    "namespace": "monitoring",
    "resourceDefaults": {
      "labels": {
        "environment": "production",
        "managed-by": "tanka"
      }
    },
    "expectVersions": {
      "kubernetes": "1.28.0"
    }
  }
}
```

---

## 2. Helm Chart Integration Patterns

### Pattern A: Direct Helm Template Integration

**Use Case**: Simple chart consumption with value overrides

```jsonnet
// environments/production/main.jsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);

{
  // Load Grafana Helm chart with basic customization
  grafana: helm.template('grafana', '../../charts/grafana', {
    namespace: 'monitoring',
    values: {
      persistence: {
        enabled: true,
        size: '10Gi',
        storageClassName: 'fast-ssd',
      },
      adminPassword: 'changeme',
      plugins: [
        'grafana-clock-panel',
        'grafana-simple-json-datasource',
      ],
      datasources: {
        'datasources.yaml': {
          apiVersion: 1,
          datasources: [
            {
              name: 'Prometheus',
              type: 'prometheus',
              url: 'http://prometheus:9090',
              isDefault: true,
            },
            {
              name: 'Loki',
              type: 'loki',
              url: 'http://loki:3100',
            },
          ],
        },
      },
    },
  }),
}
```

### Pattern B: Helm + Deep Jsonnet Merging

**Use Case**: Override fields not exposed in values.yaml

```jsonnet
// environments/production/main.jsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);
local k = import 'github.com/grafana/jsonnet-libs/ksonnet-util/kausal.libsonnet';

{
  // Load chart and deeply merge custom modifications
  grafana: helm.template('grafana', '../../charts/grafana', {
    namespace: 'monitoring',
    values: {
      persistence: { enabled: true, size: '10Gi' },
    },
  }) + {
    // Add custom annotations not in values.yaml
    deployment_grafana+: {
      spec+: {
        template+: {
          metadata+: {
            annotations+: {
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '3000',
              'vault.hashicorp.com/agent-inject': 'true',
              'vault.hashicorp.com/role': 'grafana',
            },
          },
        },
      },
    },

    // Modify service to add custom labels
    service_grafana+: {
      metadata+: {
        labels+: {
          'monitoring.grafana.com/scrape': 'true',
        },
      },
    },

    // Add init container not possible via values
    deployment_grafana+: {
      spec+: {
        template+: {
          spec+: {
            initContainers+: [
              k.core.v1.container.new('wait-for-postgres', 'busybox:1.36')
              + k.core.v1.container.withCommand([
                'sh',
                '-c',
                'until nc -z postgres 5432; do sleep 2; done',
              ]),
            ],
          },
        },
      },
    },
  },
}
```

### Pattern C: Wrapped Library Pattern

**Use Case**: Create reusable abstractions for teams

```jsonnet
// lib/grafana/grafana.libsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);

{
  new(config={}):: {
    local defaults = {
      namespace: 'monitoring',
      replicas: 1,
      persistence: { enabled: true, size: '10Gi' },
      adminPassword: 'changeme',
      datasources: [],
      dashboards: [],
      plugins: [],
    },

    local cfg = defaults + config,

    _config:: cfg,

    // Generate Helm resources
    _helm: helm.template('grafana', '../../charts/grafana', {
      namespace: cfg.namespace,
      values: {
        replicas: cfg.replicas,
        persistence: cfg.persistence,
        adminPassword: cfg.adminPassword,
        plugins: cfg.plugins,
        datasources: if std.length(cfg.datasources) > 0 then {
          'datasources.yaml': {
            apiVersion: 1,
            datasources: cfg.datasources,
          },
        } else {},
      },
    }),

    // Expose components for further customization
    deployment: self._helm.deployment_grafana,
    service: self._helm.service_grafana,
    configmap: self._helm.configmap_grafana,

    // Helper method to add datasource
    withDatasource(name, type, url, isDefault=false):: self + {
      _config+:: {
        datasources+: [{
          name: name,
          type: type,
          url: url,
          isDefault: isDefault,
        }],
      },
    },

    // Helper method to add plugin
    withPlugin(plugin):: self + {
      _config+:: {
        plugins+: [plugin],
      },
    },
  },
}
```

Usage:

```jsonnet
// environments/production/main.jsonnet
local grafana = import '../../lib/grafana/grafana.libsonnet';

{
  grafana: grafana.new({
    namespace: 'monitoring',
    replicas: 2,
    persistence: { enabled: true, size: '20Gi' },
  })
  .withDatasource('Prometheus', 'prometheus', 'http://prometheus:9090', true)
  .withDatasource('Loki', 'loki', 'http://loki:3100')
  .withPlugin('grafana-clock-panel')
  + {
    // Additional customization
    deployment+: {
      spec+: { template+: { spec+: {
        securityContext: { runAsUser: 472, fsGroup: 472 },
      }}},
    },
  },
}
```

---

## 3. Grafana Stack Configuration Examples

### Complete Monitoring Stack

```jsonnet
// environments/production/main.jsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);
local k = import 'github.com/grafana/jsonnet-libs/ksonnet-util/kausal.libsonnet';

// Common configuration
local config = {
  namespace: 'monitoring',
  storageClass: 'fast-ssd',
  domain: 'monitoring.example.com',

  retention: {
    metrics: '30d',
    logs: '7d',
    traces: '2d',
  },
};

{
  // Prometheus
  prometheus: helm.template('prometheus', '../../charts/prometheus', {
    namespace: config.namespace,
    values: {
      server: {
        retention: config.retention.metrics,
        persistentVolume: {
          enabled: true,
          size: '100Gi',
          storageClass: config.storageClass,
        },
        resources: {
          requests: { cpu: '500m', memory: '2Gi' },
          limits: { cpu: '2000m', memory: '4Gi' },
        },
      },
      alertmanager: {
        enabled: true,
        persistentVolume: {
          enabled: true,
          size: '10Gi',
        },
      },
    },
  }),

  // Loki
  loki: helm.template('loki', '../../charts/loki-stack', {
    namespace: config.namespace,
    values: {
      loki: {
        config: {
          auth_enabled: false,
          server: { http_listen_port: 3100 },
          ingester: {
            lifecycler: {
              ring: {
                kvstore: { store: 'inmemory' },
                replication_factor: 1,
              },
            },
            chunk_idle_period: '5m',
            chunk_retain_period: '30s',
          },
          schema_config: {
            configs: [{
              from: '2024-01-01',
              store: 's3',
              object_store: 's3',
              schema: 'v12',
              index: {
                prefix: 'loki_index_',
                period: '24h',
              },
            }],
          },
          storage_config: {
            aws: {
              s3: 's3://us-east-1/loki-data',
              s3forcepathstyle: true,
            },
          },
          limits_config: {
            retention_period: config.retention.logs,
            ingestion_rate_mb: 10,
            ingestion_burst_size_mb: 20,
          },
        },
        persistence: {
          enabled: true,
          size: '50Gi',
          storageClassName: config.storageClass,
        },
      },
      promtail: {
        enabled: true,
        config: {
          clients: [{
            url: 'http://loki:3100/loki/api/v1/push',
          }],
        },
      },
    },
  }),

  // Tempo
  tempo: helm.template('tempo', '../../charts/tempo', {
    namespace: config.namespace,
    values: {
      tempo: {
        retention: config.retention.traces,
        storage: {
          trace: {
            backend: 's3',
            s3: {
              bucket: 'tempo-traces',
              endpoint: 's3.amazonaws.com',
            },
          },
        },
      },
      persistence: {
        enabled: true,
        size: '30Gi',
        storageClassName: config.storageClass,
      },
    },
  }),

  // Grafana
  grafana: helm.template('grafana', '../../charts/grafana', {
    namespace: config.namespace,
    values: {
      persistence: {
        enabled: true,
        size: '10Gi',
        storageClassName: config.storageClass,
      },
      ingress: {
        enabled: true,
        hosts: [config.domain],
        tls: [{
          secretName: 'grafana-tls',
          hosts: [config.domain],
        }],
      },
      datasources: {
        'datasources.yaml': {
          apiVersion: 1,
          datasources: [
            {
              name: 'Prometheus',
              type: 'prometheus',
              url: 'http://prometheus-server',
              isDefault: true,
              jsonData: { timeInterval: '30s' },
            },
            {
              name: 'Loki',
              type: 'loki',
              url: 'http://loki:3100',
              jsonData: { maxLines: 1000 },
            },
            {
              name: 'Tempo',
              type: 'tempo',
              url: 'http://tempo:3200',
              jsonData: {
                tracesToLogs: {
                  datasourceUid: 'loki',
                },
              },
            },
          ],
        },
      },
    },
  }),
}
```

---

## 4. Helm Chart Management

### Using tk tool charts

```bash
# Initialize chartfile
cd environments/production
tk tool charts init

# Add Helm repositories
tk tool charts add-repo grafana https://grafana.github.io/helm-charts
tk tool charts add-repo prometheus-community https://prometheus-community.github.io/helm-charts

# Add specific charts
tk tool charts add grafana/grafana@7.0.0
tk tool charts add grafana/loki-stack@2.9.11
tk tool charts add grafana/tempo@1.6.1
tk tool charts add prometheus-community/prometheus@25.3.1

# Vendor all charts locally
tk tool charts vendor
```

This creates `chartfile.yaml`:

```yaml
# chartfile.yaml
version: 1
requires:
  - chart: grafana/grafana
    version: 7.0.0
  - chart: grafana/loki-stack
    version: 2.9.11
  - chart: grafana/tempo
    version: 1.6.1
  - chart: prometheus-community/prometheus
    version: 25.3.1

repositories:
  - name: grafana
    url: https://grafana.github.io/helm-charts
  - name: prometheus-community
    url: https://prometheus-community.github.io/helm-charts
```

---

## 5. Advanced Patterns

### Multi-Environment with Shared Config

```jsonnet
// lib/common.libsonnet
{
  new(env):: {
    local envDefaults = {
      dev: {
        replicas: 1,
        resources: { requests: { cpu: '100m', memory: '256Mi' } },
        storageSize: '10Gi',
        retention: '7d',
      },
      staging: {
        replicas: 2,
        resources: { requests: { cpu: '500m', memory: '1Gi' } },
        storageSize: '50Gi',
        retention: '14d',
      },
      production: {
        replicas: 3,
        resources: { requests: { cpu: '2000m', memory: '4Gi' } },
        storageSize: '200Gi',
        retention: '30d',
      },
    },

    config: envDefaults[env],
    namespace: 'monitoring-' + env,
    domain: env + '.monitoring.example.com',
  },
}
```

```jsonnet
// environments/production/main.jsonnet
local common = import '../../lib/common.libsonnet';
local grafana = import '../../lib/grafana/grafana.libsonnet';

local env = common.new('production');

{
  grafana: grafana.new({
    namespace: env.namespace,
    replicas: env.config.replicas,
    persistence: { enabled: true, size: env.config.storageSize },
  }),
}
```

### Using jsonnet-libs k8s-libsonnet

```jsonnet
local k = import 'github.com/grafana/jsonnet-libs/ksonnet-util/kausal.libsonnet';

{
  local deployment = k.apps.v1.deployment,
  local container = k.core.v1.container,
  local port = k.core.v1.containerPort,
  local service = k.core.v1.service,
  local configMap = k.core.v1.configMap,
  local volumeMount = k.core.v1.volumeMount,
  local volume = k.core.v1.volume,

  // Custom application deployment
  myapp: {
    configmap: configMap.new('myapp-config')
    + configMap.withData({
      'config.yaml': std.manifestYamlDoc({
        server: { port: 8080 },
        logging: { level: 'info' },
      }),
    }),

    deployment: deployment.new('myapp', 3, [
      container.new('myapp', 'myapp:v1.0.0')
      + container.withPorts([port.new('http', 8080)])
      + container.withVolumeMounts([
        volumeMount.new('config', '/etc/myapp'),
      ])
      + container.withResources({
        requests: { cpu: '100m', memory: '128Mi' },
        limits: { cpu: '500m', memory: '512Mi' },
      }),
    ])
    + deployment.mixin.spec.template.spec.withVolumes([
      volume.fromConfigMap('config', 'myapp-config'),
    ]),

    service: k.util.serviceFor(self.deployment)
    + service.mixin.spec.withType('ClusterIP'),
  },
}
```

---

## 6. Best Practices Summary

### DO ✅

1. **Vendor charts locally** - Keep charts in repository for reproducibility
2. **Use wrapper libraries** - Abstract Helm complexity for consumers
3. **Deep merge for customization** - Override any field without forking charts
4. **Lock versions** - Always use `jsonnetfile.lock.json` and `chartfile.yaml`
5. **Environment-specific configs** - Use `_config::` pattern for parameterization
6. **Validate before apply** - Use `tk show` to preview changes
7. **Use tk tool charts** - Automate chart vendoring
8. **Structure by component** - Organize lib/ by service/component
9. **Test environments** - Always test in dev before production
10. **Document overrides** - Comment why deep merges are needed

### DON'T ❌

1. **Don't use remote charts** - Always vendor locally
2. **Don't skip .new(std.thisFile)** - Required for proper path resolution
3. **Don't hardcode values** - Use `_config::` for reusability
4. **Don't ignore lock files** - They ensure reproducibility
5. **Don't fork charts unnecessarily** - Use Jsonnet merging instead
6. **Don't mix Helm and plain manifests carelessly** - Use consistent patterns
7. **Don't skip validation** - Always `tk diff` before apply
8. **Don't commit vendor/** without locks - Unstable across machines
9. **Don't use complex shell templating** - Let Jsonnet handle logic
10. **Don't bypass Tanka** - Use `tk apply`, not `kubectl apply`

---

## 7. Common Issues & Solutions

| Issue | Cause | Solution |
|-------|-------|----------|
| `opts.calledFrom unset` | Missing `.new(std.thisFile)` | Add to helm import: `helm.new(std.thisFile)` |
| Chart not found | Wrong relative path | Verify chart location from calling file |
| Resource name conflicts | Default nameFormat | Use `nameFormat: '{{ .Release.Name }}-{{ .Chart.Name }}-{{ .Template.Name }}'` |
| Helm not found | Missing binary | Install Helm or set `TANKA_HELM_PATH` |
| Import path errors | Wrong vendor structure | Run `jb install` to fix vendor/ |
| Schema validation fails | API version mismatch | Set `kubeVersion` in helm.template() |
| Deep merge not working | Incorrect object path | Use `tk eval` to inspect structure |
| Charts out of sync | No version locking | Use `chartfile.yaml` with versions |

---

## 8. Reference Repositories

- **Grafana Tanka**: https://github.com/grafana/tanka
- **Grafana Jsonnet Libs**: https://github.com/grafana/jsonnet-libs
- **k8s-libsonnet**: https://github.com/jsonnet-libs/k8s-libsonnet
- **Mimir Operations**: https://github.com/grafana/mimir/tree/main/operations/mimir
- **Loki Production Ksonnet**: https://github.com/grafana/loki/tree/main/production/ksonnet
- **TNS Demo (Full Stack)**: https://github.com/grafana/tns
- **Helm-Tanka Plugin**: https://github.com/Duologic/helm-tanka

---

## 9. Next Steps for Implementation

1. **Initialize Tanka project**:
   ```bash
   tk init --k8s=1.28
   ```

2. **Install dependencies**:
   ```bash
   jb install github.com/grafana/jsonnet-libs/tanka-util
   jb install github.com/grafana/jsonnet-libs/ksonnet-util
   ```

3. **Setup Helm charts**:
   ```bash
   tk tool charts init
   tk tool charts add-repo grafana https://grafana.github.io/helm-charts
   tk tool charts add grafana/grafana@7.0.0
   tk tool charts vendor
   ```

4. **Create environment**:
   ```bash
   tk env add environments/dev \
     --namespace=monitoring \
     --server-from-context=$(kubectl config current-context)
   ```

5. **Build configuration** using patterns from this document

6. **Preview and apply**:
   ```bash
   tk diff environments/dev
   tk apply environments/dev
   ```
