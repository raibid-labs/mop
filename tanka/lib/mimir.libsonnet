// Grafana Mimir metrics storage configuration
{
  new(config):: {
    local mimir = self,
    _config:: config {
      mimir+: {
        name: 'mimir',
        namespace: config.namespace,
        image: 'grafana/mimir:2.10.0',
        replicas: 1,
        resources: {
          requests: {
            memory: '512Mi',
            cpu: '250m',
          },
          limits: {
            memory: '1Gi',
            cpu: '1000m',
          },
        },
        storage: {
          backend: 's3',
          s3: {
            bucket: 'mimir-metrics',
            endpoint: 'minio.' + config.namespace + '.svc.cluster.local:9000',
            access_key: 'mimir',
            secret_key: 'mimir123',
            insecure: true,
          },
        },
      },
    },

    statefulSet: {
      apiVersion: 'apps/v1',
      kind: 'StatefulSet',
      metadata: {
        name: mimir._config.mimir.name,
        namespace: mimir._config.mimir.namespace,
        labels: {
          app: mimir._config.mimir.name,
          component: 'metrics',
        },
      },
      spec: {
        replicas: mimir._config.mimir.replicas,
        serviceName: mimir._config.mimir.name,
        selector: {
          matchLabels: {
            app: mimir._config.mimir.name,
            component: 'metrics',
          },
        },
        template: {
          metadata: {
            labels: {
              app: mimir._config.mimir.name,
              component: 'metrics',
            },
          },
          spec: {
            containers: [{
              name: 'mimir',
              image: mimir._config.mimir.image,
              args: [
                '-config.file=/etc/mimir/config.yaml',
                '-target=all',
              ],
              ports: [
                { name: 'http', containerPort: 8080, protocol: 'TCP' },
                { name: 'grpc', containerPort: 9095, protocol: 'TCP' },
              ],
              resources: mimir._config.mimir.resources,
              volumeMounts: [
                {
                  name: 'config',
                  mountPath: '/etc/mimir',
                },
                {
                  name: 'data',
                  mountPath: '/var/mimir',
                },
              ],
            }],
            volumes: [{
              name: 'config',
              configMap: {
                name: mimir._config.mimir.name + '-config',
              },
            }],
          },
        },
        volumeClaimTemplates: [{
          metadata: {
            name: 'data',
          },
          spec: {
            accessModes: ['ReadWriteOnce'],
            resources: {
              requests: {
                storage: '10Gi',
              },
            },
          },
        }],
      },
    },

    configMap: {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: mimir._config.mimir.name + '-config',
        namespace: mimir._config.mimir.namespace,
      },
      data: {
        'config.yaml': std.manifestYamlDoc({
          multitenancy_enabled: false,
          server: {
            http_listen_port: 8080,
            grpc_listen_port: 9095,
            log_level: 'info',
          },
          common: {
            storage: {
              backend: mimir._config.mimir.storage.backend,
              s3: mimir._config.mimir.storage.s3,
            },
          },
          blocks_storage: {
            backend: mimir._config.mimir.storage.backend,
            s3: mimir._config.mimir.storage.s3,
            tsdb: {
              dir: '/var/mimir/tsdb',
              retention_period: '24h',
            },
            bucket_store: {
              sync_dir: '/var/mimir/bucket-store-sync',
            },
          },
          ingester: {
            ring: {
              instance_addr: '127.0.0.1',
              kvstore: {
                store: 'inmemory',
              },
              replication_factor: 1,
            },
          },
          distributor: {
            ring: {
              instance_addr: '127.0.0.1',
              kvstore: {
                store: 'inmemory',
              },
            },
          },
          querier: {
            query_ingesters_within: '3h',
          },
          query_scheduler: {
            ring: {
              instance_addr: '127.0.0.1',
              kvstore: {
                store: 'inmemory',
              },
            },
          },
          frontend: {
            log_queries_longer_than: '10s',
          },
          compactor: {
            data_dir: '/var/mimir/compactor',
            sharding_ring: {
              instance_addr: '127.0.0.1',
              kvstore: {
                store: 'inmemory',
              },
            },
          },
          store_gateway: {
            sharding_ring: {
              instance_addr: '127.0.0.1',
              kvstore: {
                store: 'inmemory',
              },
              replication_factor: 1,
            },
          },
          limits: {
            max_ingestion_rate: 100000,
            max_series_per_user: 10000000,
            max_global_series_per_user: 10000000,
            max_series_per_metric: 0,
            max_global_series_per_metric: 0,
            ingestion_burst_size: 150000,
            max_cache_freshness: '10m',
            max_query_parallelism: 100,
            max_query_length: '0h',
          },
        }),
      },
    },

    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: mimir._config.mimir.name,
        namespace: mimir._config.mimir.namespace,
        labels: {
          app: mimir._config.mimir.name,
          component: 'metrics',
        },
      },
      spec: {
        selector: {
          app: mimir._config.mimir.name,
          component: 'metrics',
        },
        ports: [
          { name: 'http', port: 8080, targetPort: 8080, protocol: 'TCP' },
          { name: 'grpc', port: 9095, targetPort: 9095, protocol: 'TCP' },
        ],
      },
    },

    distributorService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: mimir._config.mimir.name + '-distributor',
        namespace: mimir._config.mimir.namespace,
        labels: {
          app: mimir._config.mimir.name,
          component: 'distributor',
        },
      },
      spec: {
        selector: {
          app: mimir._config.mimir.name,
          component: 'metrics',
        },
        ports: [
          { name: 'http', port: 8080, targetPort: 8080, protocol: 'TCP' },
        ],
      },
    },

    all: [
      mimir.configMap,
      mimir.statefulSet,
      mimir.service,
      mimir.distributorService,
    ],
  },
}