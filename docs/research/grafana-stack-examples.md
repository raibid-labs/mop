# Grafana Observability Stack with Tanka: Concrete Examples

## Overview

This document provides production-ready examples for deploying the complete Grafana observability stack using Tanka and Jsonnet, including Loki, Mimir, Tempo, Prometheus, and Grafana.

---

## 1. Project Structure

```
mop/
├── jsonnetfile.json
├── jsonnetfile.lock.json
├── chartfile.yaml
│
├── environments/
│   ├── dev/
│   │   ├── main.jsonnet
│   │   └── spec.json
│   ├── staging/
│   │   ├── main.jsonnet
│   │   └── spec.json
│   └── production/
│       ├── main.jsonnet
│       └── spec.json
│
├── lib/
│   ├── k.libsonnet
│   ├── config.libsonnet         # Shared configuration
│   ├── grafana.libsonnet        # Grafana wrapper
│   ├── loki.libsonnet           # Loki wrapper
│   ├── mimir.libsonnet          # Mimir wrapper
│   ├── tempo.libsonnet          # Tempo wrapper
│   └── prometheus.libsonnet     # Prometheus wrapper
│
├── vendor/                       # Managed by jsonnet-bundler
│   └── github.com/
│       └── grafana/
│           └── jsonnet-libs/
│
└── charts/                       # Vendored Helm charts
    ├── grafana/
    ├── loki/
    ├── mimir-distributed/
    ├── tempo/
    └── prometheus/
```

---

## 2. Shared Configuration Library

```jsonnet
// lib/config.libsonnet
{
  new(environment):: {
    local envConfigs = {
      dev: {
        domain: 'dev.monitoring.local',
        storageClass: 'standard',
        retention: {
          metrics: '7d',
          logs: '3d',
          traces: '24h',
        },
        resources: {
          small: { requests: { cpu: '100m', memory: '256Mi' }, limits: { cpu: '500m', memory: '512Mi' } },
          medium: { requests: { cpu: '500m', memory: '1Gi' }, limits: { cpu: '1000m', memory: '2Gi' } },
          large: { requests: { cpu: '1000m', memory: '2Gi' }, limits: { cpu: '2000m', memory: '4Gi' } },
        },
        replicas: { min: 1, max: 2 },
        storage: {
          prometheus: '20Gi',
          loki: '30Gi',
          mimir: '50Gi',
          tempo: '20Gi',
          grafana: '5Gi',
        },
      },
      staging: {
        domain: 'staging.monitoring.example.com',
        storageClass: 'fast-ssd',
        retention: {
          metrics: '15d',
          logs: '7d',
          traces: '2d',
        },
        resources: {
          small: { requests: { cpu: '250m', memory: '512Mi' }, limits: { cpu: '1000m', memory: '1Gi' } },
          medium: { requests: { cpu: '1000m', memory: '2Gi' }, limits: { cpu: '2000m', memory: '4Gi' } },
          large: { requests: { cpu: '2000m', memory: '4Gi' }, limits: { cpu: '4000m', memory: '8Gi' } },
        },
        replicas: { min: 2, max: 4 },
        storage: {
          prometheus: '100Gi',
          loki: '100Gi',
          mimir: '200Gi',
          tempo: '50Gi',
          grafana: '10Gi',
        },
      },
      production: {
        domain: 'monitoring.example.com',
        storageClass: 'fast-ssd',
        retention: {
          metrics: '30d',
          logs: '14d',
          traces: '7d',
        },
        resources: {
          small: { requests: { cpu: '500m', memory: '1Gi' }, limits: { cpu: '2000m', memory: '2Gi' } },
          medium: { requests: { cpu: '2000m', memory: '4Gi' }, limits: { cpu: '4000m', memory: '8Gi' } },
          large: { requests: { cpu: '4000m', memory: '8Gi' }, limits: { cpu: '8000m', memory: '16Gi' } },
        },
        replicas: { min: 3, max: 10 },
        storage: {
          prometheus: '500Gi',
          loki: '500Gi',
          mimir: '1Ti',
          tempo: '200Gi',
          grafana: '20Gi',
        },
      },
    },

    local cfg = envConfigs[environment],

    environment: environment,
    namespace: 'monitoring',
    domain: cfg.domain,
    storageClass: cfg.storageClass,
    retention: cfg.retention,
    resources: cfg.resources,
    replicas: cfg.replicas,
    storage: cfg.storage,

    // S3 configuration (environment variables or secrets)
    s3: {
      endpoint: 's3.amazonaws.com',
      region: 'us-east-1',
      buckets: {
        loki: 'mop-loki-' + environment,
        mimir: 'mop-mimir-' + environment,
        tempo: 'mop-tempo-' + environment,
      },
    },

    // Common labels
    labels: {
      environment: environment,
      'app.kubernetes.io/managed-by': 'tanka',
      'app.kubernetes.io/part-of': 'grafana-stack',
    },

    // TLS configuration
    tls: {
      enabled: if environment == 'production' then true else false,
      secretName: 'monitoring-tls',
    },
  },
}
```

---

## 3. Loki Configuration

```jsonnet
// lib/loki.libsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);

{
  new(config):: {
    local lokiConfig = {
      auth_enabled: false,

      server: {
        http_listen_port: 3100,
        grpc_listen_port: 9095,
        log_level: 'info',
      },

      common: {
        path_prefix: '/loki',
        replication_factor: config.replicas.min,
        storage: {
          s3: {
            endpoint: config.s3.endpoint,
            region: config.s3.region,
            bucketnames: config.s3.buckets.loki,
            s3forcepathstyle: false,
          },
        },
      },

      schema_config: {
        configs: [
          {
            from: '2024-01-01',
            store: 'tsdb',
            object_store: 's3',
            schema: 'v13',
            index: {
              prefix: 'loki_index_',
              period: '24h',
            },
          },
        ],
      },

      limits_config: {
        retention_period: config.retention.logs,
        ingestion_rate_strategy: 'global',
        ingestion_rate_mb: 10,
        ingestion_burst_size_mb: 20,
        max_query_length: '721h',  // 30 days
        max_query_parallelism: 16,
        max_streams_per_user: 10000,
        max_global_streams_per_user: 50000,
        reject_old_samples: true,
        reject_old_samples_max_age: '168h',
        split_queries_by_interval: '15m',
      },

      compactor: {
        working_directory: '/loki/compactor',
        shared_store: 's3',
        compaction_interval: '10m',
        retention_enabled: true,
        retention_delete_delay: '2h',
        retention_delete_worker_count: 150,
      },

      query_range: {
        results_cache: {
          cache: {
            embedded_cache: {
              enabled: true,
              max_size_mb: 500,
            },
          },
        },
      },

      ruler: {
        storage: {
          type: 's3',
          s3: {
            bucketnames: config.s3.buckets.loki + '-ruler',
          },
        },
        rule_path: '/tmp/loki/rules',
        alertmanager_url: 'http://alertmanager:9093',
        enable_api: true,
        enable_alertmanager_v2: true,
      },

      frontend: {
        compress_responses: true,
        max_outstanding_per_tenant: 2048,
      },

      ingester: {
        lifecycler: {
          ring: {
            kvstore: {
              store: 'memberlist',
            },
            replication_factor: config.replicas.min,
          },
        },
        chunk_idle_period: '30m',
        chunk_block_size: 262144,
        chunk_encoding: 'snappy',
        chunk_retain_period: '1m',
        max_chunk_age: '1h',
        wal: {
          enabled: true,
          dir: '/loki/wal',
        },
      },
    },

    loki: helm.template('loki', '../charts/loki', {
      namespace: config.namespace,
      values: {
        loki: {
          image: {
            repository: 'grafana/loki',
            tag: '2.9.3',
          },
          config: std.manifestYamlDoc(lokiConfig),
          structuredConfig: lokiConfig,
          persistence: {
            enabled: true,
            size: config.storage.loki,
            storageClassName: config.storageClass,
          },
        },

        // Gateway (nginx)
        gateway: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.small,
          ingress: {
            enabled: true,
            hosts: [{
              host: 'loki.' + config.domain,
              paths: [{ path: '/', pathType: 'Prefix' }],
            }],
            tls: if config.tls.enabled then [{
              secretName: config.tls.secretName,
              hosts: ['loki.' + config.domain],
            }] else [],
          },
        },

        // Write path (distributor, ingester)
        write: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
          persistence: {
            size: config.storage.loki,
            storageClass: config.storageClass,
          },
        },

        // Read path (query-frontend, querier)
        read: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Backend (compactor, index-gateway, ruler)
        backend: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
          persistence: {
            size: config.storage.loki,
            storageClass: config.storageClass,
          },
        },

        // Monitoring
        monitoring: {
          serviceMonitor: {
            enabled: true,
            labels: config.labels,
          },
          selfMonitoring: {
            enabled: true,
            grafanaAgent: {
              installOperator: false,
            },
          },
        },
      },
    }),

    // Add custom annotations for cost tracking
    result: self.loki + {
      deployment_loki_write+: {
        metadata+: {
          annotations+: {
            'cost-center': 'observability',
            'data-classification': 'internal',
          },
        },
      },
    },
  },
}
```

---

## 4. Mimir Configuration

```jsonnet
// lib/mimir.libsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);

{
  new(config):: {
    mimir: helm.template('mimir', '../charts/mimir-distributed', {
      namespace: config.namespace,
      values: {
        global: {
          clusterDomain: 'cluster.local',
        },

        mimir: {
          structuredConfig: {
            multitenancy_enabled: false,

            server: {
              log_level: 'info',
              http_listen_port: 8080,
              grpc_listen_port: 9095,
            },

            common: {
              storage: {
                backend: 's3',
                s3: {
                  endpoint: config.s3.endpoint,
                  region: config.s3.region,
                  bucket_name: config.s3.buckets.mimir,
                },
              },
            },

            blocks_storage: {
              backend: 's3',
              s3: {
                endpoint: config.s3.endpoint,
                region: config.s3.region,
                bucket_name: config.s3.buckets.mimir + '-blocks',
              },
              tsdb: {
                dir: '/data/tsdb',
                retention_period: config.retention.metrics,
              },
            },

            compactor: {
              compaction_interval: '30m',
              deletion_delay: '2h',
              max_opening_blocks_concurrency: 4,
              max_closing_blocks_concurrency: 2,
              symbols_flushers_concurrency: 4,
              data_dir: '/data/compactor',
            },

            distributor: {
              ring: {
                kvstore: { store: 'memberlist' },
              },
            },

            ingester: {
              ring: {
                kvstore: { store: 'memberlist' },
                replication_factor: config.replicas.min,
              },
            },

            ruler_storage: {
              backend: 's3',
              s3: {
                bucket_name: config.s3.buckets.mimir + '-ruler',
              },
            },

            alertmanager_storage: {
              backend: 's3',
              s3: {
                bucket_name: config.s3.buckets.mimir + '-alertmanager',
              },
            },

            limits: {
              max_label_names_per_series: 30,
              max_global_series_per_user: 1500000,
              max_global_series_per_metric: 300000,
              ingestion_rate: 100000,
              ingestion_burst_size: 200000,
            },
          },
        },

        // Distributor
        distributor: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Ingester
        ingester: {
          replicas: config.replicas.min * 2,  // More ingesters for better distribution
          resources: config.resources.large,
          persistentVolume: {
            enabled: true,
            size: config.storage.mimir,
            storageClass: config.storageClass,
          },
          zoneAwareReplication: {
            enabled: config.environment == 'production',
          },
        },

        // Querier
        querier: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Query Frontend
        query_frontend: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Query Scheduler
        query_scheduler: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.small,
        },

        // Store Gateway
        store_gateway: {
          replicas: config.replicas.min,
          resources: config.resources.large,
          persistentVolume: {
            enabled: true,
            size: config.storage.mimir,
            storageClass: config.storageClass,
          },
          zoneAwareReplication: {
            enabled: config.environment == 'production',
          },
        },

        // Compactor
        compactor: {
          replicas: 1,
          resources: config.resources.large,
          persistentVolume: {
            enabled: true,
            size: config.storage.mimir,
            storageClass: config.storageClass,
          },
        },

        // Ruler
        ruler: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Alertmanager
        alertmanager: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.small,
          persistentVolume: {
            enabled: true,
            size: '10Gi',
            storageClass: config.storageClass,
          },
        },

        // Nginx gateway
        nginx: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.small,
          ingress: {
            enabled: true,
            hosts: [{
              host: 'mimir.' + config.domain,
              paths: [{ path: '/', pathType: 'Prefix' }],
            }],
            tls: if config.tls.enabled then [{
              secretName: config.tls.secretName,
              hosts: ['mimir.' + config.domain],
            }] else [],
          },
        },

        // Memcached for caching
        memcached: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Monitoring
        serviceMonitor: {
          enabled: true,
          labels: config.labels,
        },
      },
    }),
  },
}
```

---

## 5. Tempo Configuration

```jsonnet
// lib/tempo.libsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);

{
  new(config):: {
    tempo: helm.template('tempo', '../charts/tempo', {
      namespace: config.namespace,
      values: {
        tempo: {
          repository: 'grafana/tempo',
          tag: '2.3.1',

          retention: config.retention.traces,

          metricsGenerator: {
            enabled: true,
            remoteWriteUrl: 'http://mimir-nginx/api/v1/push',
          },

          storage: {
            trace: {
              backend: 's3',
              s3: {
                bucket: config.s3.buckets.tempo,
                endpoint: config.s3.endpoint + ':443',
                region: config.s3.region,
                forcepathstyle: false,
              },
              wal: {
                path: '/var/tempo/wal',
              },
              pool: {
                max_workers: 100,
                queue_depth: 10000,
              },
            },
          },

          receivers: {
            jaeger: {
              protocols: {
                grpc: { endpoint: '0.0.0.0:14250' },
                thrift_binary: { endpoint: '0.0.0.0:6832' },
                thrift_compact: { endpoint: '0.0.0.0:6831' },
                thrift_http: { endpoint: '0.0.0.0:14268' },
              },
            },
            zipkin: { endpoint: '0.0.0.0:9411' },
            otlp: {
              protocols: {
                grpc: { endpoint: '0.0.0.0:4317' },
                http: { endpoint: '0.0.0.0:4318' },
              },
            },
            opencensus: null,
          },

          overrides: {
            metrics_generator_processors: ['service-graphs', 'span-metrics'],
          },
        },

        // Distributor
        distributor: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Ingester
        ingester: {
          replicas: config.replicas.min,
          resources: config.resources.large,
          persistence: {
            enabled: true,
            size: config.storage.tempo,
            storageClass: config.storageClass,
          },
        },

        // Querier
        querier: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Query Frontend
        queryFrontend: {
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Compactor
        compactor: {
          replicas: 1,
          resources: config.resources.large,
          persistence: {
            enabled: true,
            size: config.storage.tempo,
            storageClass: config.storageClass,
          },
        },

        // Metrics Generator
        metricsGenerator: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.medium,
        },

        // Gateway
        gateway: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.small,
          ingress: {
            enabled: true,
            hosts: [{
              host: 'tempo.' + config.domain,
              paths: [{ path: '/', pathType: 'Prefix' }],
            }],
            tls: if config.tls.enabled then [{
              secretName: config.tls.secretName,
              hosts: ['tempo.' + config.domain],
            }] else [],
          },
        },

        // Monitoring
        serviceMonitor: {
          enabled: true,
          labels: config.labels,
        },

        // Memcached
        memcached: {
          enabled: true,
          replicas: config.replicas.min,
          resources: config.resources.small,
        },
      },
    }),
  },
}
```

---

## 6. Grafana Configuration

```jsonnet
// lib/grafana.libsonnet
local tanka = import 'github.com/grafana/jsonnet-libs/tanka-util/main.libsonnet';
local helm = tanka.helm.new(std.thisFile);

{
  new(config):: {
    grafana: helm.template('grafana', '../charts/grafana', {
      namespace: config.namespace,
      values: {
        replicas: config.replicas.min,

        image: {
          repository: 'grafana/grafana',
          tag: '10.2.2',
        },

        resources: config.resources.medium,

        persistence: {
          enabled: true,
          size: config.storage.grafana,
          storageClass: config.storageClass,
        },

        // Admin credentials (use secrets in production)
        adminUser: 'admin',
        adminPassword: 'changeme',

        // Grafana configuration
        'grafana.ini': {
          server: {
            root_url: 'https://' + config.domain,
            domain: config.domain,
          },

          security: {
            admin_user: 'admin',
            cookie_secure: config.tls.enabled,
            strict_transport_security: config.tls.enabled,
          },

          auth: {
            disable_login_form: false,
            oauth_auto_login: false,
          },

          'auth.anonymous': {
            enabled: false,
          },

          analytics: {
            reporting_enabled: false,
            check_for_updates: false,
          },

          snapshots: {
            external_enabled: false,
          },

          users: {
            allow_sign_up: false,
            auto_assign_org: true,
            auto_assign_org_role: 'Viewer',
          },

          log: {
            mode: 'console',
            level: 'info',
          },

          database: {
            type: 'postgres',
            host: 'postgresql:5432',
            name: 'grafana',
            user: 'grafana',
            password: '$__env{GF_DATABASE_PASSWORD}',
          },
        },

        // Datasources
        datasources: {
          'datasources.yaml': {
            apiVersion: 1,
            datasources: [
              {
                name: 'Mimir',
                type: 'prometheus',
                url: 'http://mimir-nginx.' + config.namespace + '.svc.cluster.local/prometheus',
                access: 'proxy',
                isDefault: true,
                jsonData: {
                  timeInterval: '30s',
                  httpMethod: 'POST',
                },
                editable: false,
              },
              {
                name: 'Loki',
                type: 'loki',
                url: 'http://loki-gateway.' + config.namespace + '.svc.cluster.local',
                access: 'proxy',
                jsonData: {
                  maxLines: 1000,
                  derivedFields: [
                    {
                      datasourceUid: 'tempo',
                      matcherRegex: '"trace_id":"(\\w+)"',
                      name: 'TraceID',
                      url: '$${__value.raw}',
                    },
                  ],
                },
                editable: false,
              },
              {
                name: 'Tempo',
                type: 'tempo',
                url: 'http://tempo-gateway.' + config.namespace + '.svc.cluster.local',
                access: 'proxy',
                jsonData: {
                  httpMethod: 'GET',
                  tracesToLogs: {
                    datasourceUid: 'loki',
                    tags: ['job', 'instance', 'pod', 'namespace'],
                    mappedTags: [{ key: 'service.name', value: 'service' }],
                    mapTagNamesEnabled: true,
                    spanStartTimeShift: '-1h',
                    spanEndTimeShift: '1h',
                    filterByTraceID: true,
                    filterBySpanID: false,
                  },
                  tracesToMetrics: {
                    datasourceUid: 'mimir',
                    tags: [{ key: 'service.name', value: 'service' }],
                    queries: [
                      {
                        name: 'Sample query',
                        query: 'sum(rate(tempo_spanmetrics_latency_bucket{$__tags}[5m]))',
                      },
                    ],
                  },
                  serviceMap: {
                    datasourceUid: 'mimir',
                  },
                  nodeGraph: {
                    enabled: true,
                  },
                  search: {
                    hide: false,
                  },
                  lokiSearch: {
                    datasourceUid: 'loki',
                  },
                },
                editable: false,
              },
            ],
          },
        },

        // Dashboard providers
        dashboardProviders: {
          'dashboardproviders.yaml': {
            apiVersion: 1,
            providers: [
              {
                name: 'default',
                orgId: 1,
                folder: '',
                type: 'file',
                disableDeletion: false,
                updateIntervalSeconds: 30,
                allowUiUpdates: true,
                options: {
                  path: '/var/lib/grafana/dashboards/default',
                },
              },
              {
                name: 'observability',
                orgId: 1,
                folder: 'Observability Stack',
                type: 'file',
                disableDeletion: false,
                options: {
                  path: '/var/lib/grafana/dashboards/observability',
                },
              },
            ],
          },
        },

        // Pre-installed dashboards
        dashboards: {
          default: {},
          observability: {
            'loki-overview': {
              gnetId: 13639,
              revision: 2,
              datasource: 'Loki',
            },
            'mimir-overview': {
              gnetId: 19125,
              revision: 1,
              datasource: 'Mimir',
            },
            'tempo-overview': {
              gnetId: 16369,
              revision: 1,
              datasource: 'Tempo',
            },
          },
        },

        // Plugins
        plugins: [
          'grafana-clock-panel',
          'grafana-piechart-panel',
          'grafana-worldmap-panel',
        ],

        // Environment variables
        env: {
          GF_EXPLORE_ENABLED: 'true',
          GF_PANELS_DISABLE_SANITIZE_HTML: 'true',
          GF_LOG_FILTERS: 'rendering:debug',
        },

        // Ingress
        ingress: {
          enabled: true,
          hosts: [config.domain],
          path: '/',
          pathType: 'Prefix',
          tls: if config.tls.enabled then [{
            secretName: config.tls.secretName,
            hosts: [config.domain],
          }] else [],
          annotations: {
            'kubernetes.io/ingress.class': 'nginx',
            'cert-manager.io/cluster-issuer': 'letsencrypt-prod',
            'nginx.ingress.kubernetes.io/force-ssl-redirect': 'true',
          },
        },

        // Service Monitor
        serviceMonitor: {
          enabled: true,
          labels: config.labels,
        },

        // RBAC
        rbac: {
          create: true,
          pspEnabled: false,
        },

        serviceAccount: {
          create: true,
          name: 'grafana',
        },
      },
    }),
  },
}
```

---

## 7. Production Environment Example

```jsonnet
// environments/production/main.jsonnet
local config = import '../../lib/config.libsonnet';
local grafana = import '../../lib/grafana.libsonnet';
local loki = import '../../lib/loki.libsonnet';
local mimir = import '../../lib/mimir.libsonnet';
local tempo = import '../../lib/tempo.libsonnet';

local env = config.new('production');

{
  _config:: env,

  // Deploy complete stack
  loki: loki.new(env).result,
  mimir: mimir.new(env).mimir,
  tempo: tempo.new(env).tempo,
  grafana: grafana.new(env).grafana,

  // Additional namespace resources
  namespace: {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: {
      name: env.namespace,
      labels: env.labels,
    },
  },

  // Storage class (if custom)
  storageClass: {
    apiVersion: 'storage.k8s.io/v1',
    kind: 'StorageClass',
    metadata: {
      name: env.storageClass,
    },
    provisioner: 'kubernetes.io/aws-ebs',
    parameters: {
      type: 'gp3',
      iops: '3000',
      throughput: '125',
    },
    allowVolumeExpansion: true,
    reclaimPolicy: 'Retain',
  },
}
```

```json
// environments/production/spec.json
{
  "apiVersion": "tanka.dev/v1alpha1",
  "kind": "Environment",
  "metadata": {
    "name": "environments/production"
  },
  "spec": {
    "apiServer": "https://prod-k8s.example.com:6443",
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

## 8. Deployment Commands

```bash
# Initialize project
tk init --k8s=1.28
cd environments/production

# Install dependencies
jb install github.com/grafana/jsonnet-libs/tanka-util
jb install github.com/grafana/jsonnet-libs/ksonnet-util

# Setup Helm charts
tk tool charts init
tk tool charts add-repo grafana https://grafana.github.io/helm-charts
tk tool charts add grafana/grafana@7.0.0
tk tool charts add grafana/loki@5.41.4
tk tool charts add grafana/mimir-distributed@5.1.3
tk tool charts add grafana/tempo@1.7.1
tk tool charts vendor

# Preview changes
tk diff environments/production

# Show full manifests
tk show environments/production

# Apply to cluster
tk apply environments/production

# Apply specific component
tk apply environments/production --target=grafana

# Check status
kubectl get pods -n monitoring
kubectl get svc -n monitoring
kubectl get ingress -n monitoring

# View logs
kubectl logs -n monitoring -l app.kubernetes.io/name=grafana --tail=100
kubectl logs -n monitoring -l app.kubernetes.io/name=loki-write --tail=100
kubectl logs -n monitoring -l app.kubernetes.io/name=mimir-ingester --tail=100

# Delete environment
tk delete environments/production
```

---

## 9. Testing & Validation

```bash
# Test Loki
curl -H "Content-Type: application/json" \
  -XPOST https://loki.monitoring.example.com/loki/api/v1/push \
  --data-raw '{"streams": [{ "stream": { "foo": "bar" }, "values": [ [ "1640000000000000000", "test log line" ] ] }]}'

# Query Loki
curl -G -s "https://loki.monitoring.example.com/loki/api/v1/query" \
  --data-urlencode 'query={foo="bar"}'

# Test Mimir (Prometheus remote write)
curl -X POST https://mimir.monitoring.example.com/api/v1/push \
  -H "Content-Type: application/x-protobuf" \
  --data-binary @metrics.pb

# Query Mimir
curl -G https://mimir.monitoring.example.com/prometheus/api/v1/query \
  --data-urlencode 'query=up'

# Test Tempo (send trace)
curl -X POST https://tempo.monitoring.example.com/v1/traces \
  -H "Content-Type: application/json" \
  -d @trace.json

# Access Grafana
open https://monitoring.example.com

# Port-forward for local testing
kubectl port-forward -n monitoring svc/grafana 3000:80
kubectl port-forward -n monitoring svc/loki-gateway 3100:80
kubectl port-forward -n monitoring svc/mimir-nginx 8080:80
kubectl port-forward -n monitoring svc/tempo-gateway 3200:80
```

---

## 10. Monitoring & Operations

```bash
# View metrics
kubectl top pods -n monitoring
kubectl top nodes

# Check resource usage
kubectl describe node | grep -A 5 "Allocated resources"

# Scale components
kubectl scale deployment -n monitoring grafana --replicas=3
kubectl scale statefulset -n monitoring loki-write --replicas=4

# Update configuration
tk apply environments/production --diff

# Rollback
kubectl rollout undo deployment/grafana -n monitoring
kubectl rollout status deployment/grafana -n monitoring

# Backup
kubectl get all -n monitoring -o yaml > monitoring-backup.yaml

# Check logs for errors
kubectl logs -n monitoring -l app.kubernetes.io/name=loki-write | grep ERROR
kubectl logs -n monitoring -l app.kubernetes.io/name=mimir-ingester | grep ERROR
```

---

## 11. Troubleshooting Common Issues

### Loki not receiving logs

```bash
# Check promtail
kubectl logs -n monitoring -l app.kubernetes.io/name=promtail

# Verify service connectivity
kubectl exec -n monitoring -it promtail-xxxxx -- wget -O- http://loki-gateway/ready

# Check ingester ring
curl https://loki.monitoring.example.com/ring
```

### Mimir ingestion issues

```bash
# Check distributor logs
kubectl logs -n monitoring -l app.kubernetes.io/name=mimir-distributor

# Verify S3 access
kubectl exec -n monitoring -it mimir-ingester-0 -- aws s3 ls s3://mop-mimir-production/

# Check memberlist
kubectl logs -n monitoring mimir-ingester-0 | grep memberlist
```

### Tempo trace ingestion

```bash
# Test receivers
kubectl port-forward -n monitoring svc/tempo-distributor 4317:4317
grpcurl -plaintext localhost:4317 list

# Check ingester
kubectl logs -n monitoring -l app.kubernetes.io/name=tempo-ingester
```

---

## Summary

This configuration provides:

- **Complete observability stack** with Loki, Mimir, Tempo, and Grafana
- **Production-ready** with proper sizing, persistence, and HA
- **Environment-specific** configurations (dev, staging, production)
- **Modular architecture** with reusable libraries
- **Integrated dashboards** and datasources
- **Proper resource management** and scaling
- **Security** with TLS, RBAC, and secrets
- **Monitoring** with ServiceMonitors

All components are deeply integrated with trace-to-logs, logs-to-metrics, and metrics-to-traces correlation.
