// Grafana Loki log aggregation configuration
{
  new(config):: {
    local loki = self,
    _config:: config {
      loki+: {
        name: 'loki',
        namespace: config.namespace,
        image: 'grafana/loki:2.9.2',
        replicas: 1,
        resources: {
          requests: {
            memory: '256Mi',
            cpu: '100m',
          },
          limits: {
            memory: '512Mi',
            cpu: '500m',
          },
        },
        storage: {
          backend: 's3',
          s3: {
            bucketnames: 'loki-logs',
            endpoint: 'minio.' + config.namespace + '.svc.cluster.local:9000',
            access_key_id: 'loki',
            secret_access_key: 'loki123',
            insecure: true,
            s3forcepathstyle: true,
          },
        },
      },
    },

    statefulSet: {
      apiVersion: 'apps/v1',
      kind: 'StatefulSet',
      metadata: {
        name: loki._config.loki.name,
        namespace: loki._config.loki.namespace,
        labels: {
          app: loki._config.loki.name,
          component: 'logs',
        },
      },
      spec: {
        replicas: loki._config.loki.replicas,
        serviceName: loki._config.loki.name,
        selector: {
          matchLabels: {
            app: loki._config.loki.name,
            component: 'logs',
          },
        },
        template: {
          metadata: {
            labels: {
              app: loki._config.loki.name,
              component: 'logs',
            },
          },
          spec: {
            containers: [{
              name: 'loki',
              image: loki._config.loki.image,
              args: [
                '-config.file=/etc/loki/config.yaml',
                '-target=all',
              ],
              ports: [
                { name: 'http', containerPort: 3100, protocol: 'TCP' },
                { name: 'grpc', containerPort: 9095, protocol: 'TCP' },
              ],
              resources: loki._config.loki.resources,
              volumeMounts: [
                {
                  name: 'config',
                  mountPath: '/etc/loki',
                },
                {
                  name: 'data',
                  mountPath: '/var/loki',
                },
              ],
            }],
            volumes: [{
              name: 'config',
              configMap: {
                name: loki._config.loki.name + '-config',
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
        name: loki._config.loki.name + '-config',
        namespace: loki._config.loki.namespace,
      },
      data: {
        'config.yaml': std.manifestYamlDoc({
          auth_enabled: false,
          server: {
            http_listen_port: 3100,
            grpc_listen_port: 9095,
            log_level: 'info',
          },
          common: {
            instance_addr: '127.0.0.1',
            path_prefix: '/var/loki',
            storage: {
              s3: loki._config.loki.storage.s3,
            },
            replication_factor: 1,
            ring: {
              kvstore: {
                store: 'inmemory',
              },
            },
          },
          schema_config: {
            configs: [{
              from: '2023-01-01',
              store: 'tsdb',
              object_store: 's3',
              schema: 'v12',
              index: {
                prefix: 'index_',
                period: '24h',
              },
            }],
          },
          ingester: {
            wal: {
              enabled: true,
              dir: '/var/loki/wal',
            },
            lifecycler: {
              address: '127.0.0.1',
              ring: {
                kvstore: {
                  store: 'inmemory',
                },
                replication_factor: 1,
              },
            },
            chunk_idle_period: '5m',
            chunk_retain_period: '30s',
            max_transfer_retries: 0,
            chunk_target_size: 1572864,
            chunk_encoding: 'snappy',
          },
          storage_config: {
            tsdb_shipper: {
              active_index_directory: '/var/loki/tsdb-index',
              cache_location: '/var/loki/tsdb-cache',
              shared_store: 's3',
            },
            aws: loki._config.loki.storage.s3,
          },
          limits_config: {
            enforce_metric_name: false,
            reject_old_samples: true,
            reject_old_samples_max_age: '168h',
            max_entries_limit_per_query: 5000,
            ingestion_rate_mb: 10,
            ingestion_burst_size_mb: 20,
            max_query_parallelism: 100,
            per_stream_rate_limit: '3MB',
            per_stream_rate_limit_burst: '15MB',
            retention_period: '168h',
          },
          chunk_store_config: {
            max_look_back_period: '0s',
          },
          table_manager: {
            retention_deletes_enabled: true,
            retention_period: '168h',
          },
          query_range: {
            results_cache: {
              cache: {
                embedded_cache: {
                  enabled: true,
                  max_size_mb: 100,
                },
              },
            },
          },
          frontend: {
            compress_responses: true,
            log_queries_longer_than: '10s',
          },
          compactor: {
            working_directory: '/var/loki/compactor',
            shared_store: 's3',
            compaction_interval: '10m',
            retention_enabled: true,
            retention_delete_delay: '2h',
            retention_delete_worker_count: 150,
          },
          ruler: {
            storage: {
              'type': 'local',
              'local': {
                directory: '/var/loki/rules',
              },
            },
            rule_path: '/var/loki/rules-temp',
          },
        }),
      },
    },

    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: loki._config.loki.name,
        namespace: loki._config.loki.namespace,
        labels: {
          app: loki._config.loki.name,
          component: 'logs',
        },
      },
      spec: {
        selector: {
          app: loki._config.loki.name,
          component: 'logs',
        },
        ports: [
          { name: 'http', port: 3100, targetPort: 3100, protocol: 'TCP' },
          { name: 'grpc', port: 9095, targetPort: 9095, protocol: 'TCP' },
        ],
      },
    },

    distributorService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: loki._config.loki.name + '-distributor',
        namespace: loki._config.loki.namespace,
        labels: {
          app: loki._config.loki.name,
          component: 'distributor',
        },
      },
      spec: {
        selector: {
          app: loki._config.loki.name,
          component: 'logs',
        },
        ports: [
          { name: 'http', port: 3100, targetPort: 3100, protocol: 'TCP' },
        ],
      },
    },

    all: [
      loki.configMap,
      loki.statefulSet,
      loki.service,
      loki.distributorService,
    ],
  },
}