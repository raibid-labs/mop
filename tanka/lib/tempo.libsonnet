// Grafana Tempo distributed tracing configuration
{
  new(config):: {
    local tempo = self,
    _config:: config {
      tempo+: {
        name: 'tempo',
        namespace: config.namespace,
        image: 'grafana/tempo:2.3.0',
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
            bucket: 'tempo-traces',
            endpoint: 'minio.' + config.namespace + '.svc.cluster.local:9000',
            access_key: 'tempo',
            secret_key: 'tempo123',
            insecure: true,
          },
        },
      },
    },

    statefulSet: {
      apiVersion: 'apps/v1',
      kind: 'StatefulSet',
      metadata: {
        name: tempo._config.tempo.name,
        namespace: tempo._config.tempo.namespace,
        labels: {
          app: tempo._config.tempo.name,
          component: 'tracing',
        },
      },
      spec: {
        replicas: tempo._config.tempo.replicas,
        serviceName: tempo._config.tempo.name,
        selector: {
          matchLabels: {
            app: tempo._config.tempo.name,
            component: 'tracing',
          },
        },
        template: {
          metadata: {
            labels: {
              app: tempo._config.tempo.name,
              component: 'tracing',
            },
          },
          spec: {
            containers: [{
              name: 'tempo',
              image: tempo._config.tempo.image,
              args: [
                '-config.file=/etc/tempo/config.yaml',
              ],
              ports: [
                { name: 'otlp-grpc', containerPort: 4317, protocol: 'TCP' },
                { name: 'tempo-grpc', containerPort: 9095, protocol: 'TCP' },
                { name: 'tempo-http', containerPort: 3200, protocol: 'TCP' },
              ],
              resources: tempo._config.tempo.resources,
              volumeMounts: [
                {
                  name: 'config',
                  mountPath: '/etc/tempo',
                },
                {
                  name: 'data',
                  mountPath: '/var/tempo',
                },
              ],
            }],
            volumes: [{
              name: 'config',
              configMap: {
                name: tempo._config.tempo.name + '-config',
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
        name: tempo._config.tempo.name + '-config',
        namespace: tempo._config.tempo.namespace,
      },
      data: {
        'config.yaml': std.manifestYamlDoc({
          server: {
            http_listen_port: 3200,
            grpc_listen_port: 9095,
          },
          distributor: {
            receivers: {
              otlp: {
                protocols: {
                  grpc: {
                    endpoint: '0.0.0.0:4317',
                  },
                },
              },
              jaeger: {
                protocols: {
                  grpc: {
                    endpoint: '0.0.0.0:14250',
                  },
                  thrift_binary: {
                    endpoint: '0.0.0.0:6832',
                  },
                  thrift_compact: {
                    endpoint: '0.0.0.0:6831',
                  },
                },
              },
              zipkin: {
                endpoint: '0.0.0.0:9411',
              },
            },
          },
          ingester: {
            max_block_duration: '5m',
          },
          compactor: {
            compaction: {
              block_retention: '168h',
            },
          },
          storage: {
            trace: {
              backend: tempo._config.tempo.storage.backend,
            } + (
              if tempo._config.tempo.storage.backend == 's3' then {
                s3: tempo._config.tempo.storage.s3,
              } else {}
            ) + {
              'local': {
                path: '/var/tempo/traces',
              },
              wal: {
                path: '/var/tempo/wal',
              },
            },
          },
          querier: {
            frontend_worker: {
              frontend_address: tempo._config.tempo.name + ':9095',
            },
          },
          overrides: {
            max_traces_per_user: 100000,
            max_search_duration: '0s',
          },
        }),
      },
    },

    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: tempo._config.tempo.name,
        namespace: tempo._config.tempo.namespace,
        labels: {
          app: tempo._config.tempo.name,
          component: 'tracing',
        },
      },
      spec: {
        selector: {
          app: tempo._config.tempo.name,
          component: 'tracing',
        },
        ports: [
          { name: 'otlp-grpc', port: 4317, targetPort: 4317, protocol: 'TCP' },
          { name: 'tempo-grpc', port: 9095, targetPort: 9095, protocol: 'TCP' },
          { name: 'tempo-http', port: 3200, targetPort: 3200, protocol: 'TCP' },
        ],
      },
    },

    distributorService: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: tempo._config.tempo.name + '-distributor',
        namespace: tempo._config.tempo.namespace,
        labels: {
          app: tempo._config.tempo.name,
          component: 'distributor',
        },
      },
      spec: {
        selector: {
          app: tempo._config.tempo.name,
          component: 'tracing',
        },
        ports: [
          { name: 'otlp-grpc', port: 4317, targetPort: 4317, protocol: 'TCP' },
        ],
      },
    },

    all: [
      tempo.configMap,
      tempo.statefulSet,
      tempo.service,
      tempo.distributorService,
    ],
  },
}