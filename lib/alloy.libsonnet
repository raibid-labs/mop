// Grafana Alloy (OpenTelemetry Collector) configuration
{
  new(config):: {
    local alloy = self,
    _config:: config {
      alloy+: {
        name: 'alloy',
        namespace: config.namespace,
        image: 'grafana/alloy:v1.0.0',
        replicas: 2,
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
        exporters: {
          tempo: 'tempo-distributor.' + config.namespace + '.svc.cluster.local:4317',
          mimir: 'mimir-distributor.' + config.namespace + '.svc.cluster.local:8080',
          loki: 'loki-distributor.' + config.namespace + '.svc.cluster.local:3100',
        },
      },
    },

    deployment: {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: alloy._config.alloy.name,
        namespace: alloy._config.alloy.namespace,
        labels: {
          app: alloy._config.alloy.name,
          component: 'collector',
        },
      },
      spec: {
        replicas: alloy._config.alloy.replicas,
        selector: {
          matchLabels: {
            app: alloy._config.alloy.name,
            component: 'collector',
          },
        },
        template: {
          metadata: {
            labels: {
              app: alloy._config.alloy.name,
              component: 'collector',
            },
          },
          spec: {
            containers: [{
              name: 'alloy',
              image: alloy._config.alloy.image,
              args: [
                'run',
                '/etc/alloy/config.river',
                '--server.http.listen-addr=0.0.0.0:12345',
              ],
              ports: [
                { name: 'otlp-grpc', containerPort: 4317, protocol: 'TCP' },
                { name: 'otlp-http', containerPort: 4318, protocol: 'TCP' },
                { name: 'metrics', containerPort: 12345, protocol: 'TCP' },
              ],
              resources: alloy._config.alloy.resources,
              volumeMounts: [{
                name: 'config',
                mountPath: '/etc/alloy',
              }],
            }],
            volumes: [{
              name: 'config',
              configMap: {
                name: alloy._config.alloy.name + '-config',
              },
            }],
          },
        },
      },
    },

    configMap: {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: alloy._config.alloy.name + '-config',
        namespace: alloy._config.alloy.namespace,
      },
      data: {
        'config.river': |||
          // OTLP receiver for traces, metrics, and logs
          otelcol.receiver.otlp "default" {
            grpc {
              endpoint = "0.0.0.0:4317"
            }
            http {
              endpoint = "0.0.0.0:4318"
            }

            output {
              metrics = [otelcol.processor.batch.default.input]
              logs    = [otelcol.processor.batch.default.input]
              traces  = [otelcol.processor.batch.default.input]
            }
          }

          // Batch processor to optimize exports
          otelcol.processor.batch "default" {
            timeout = "5s"
            send_batch_size = 1000

            output {
              metrics = [otelcol.exporter.prometheus.mimir.input]
              logs    = [otelcol.exporter.loki.default.input]
              traces  = [otelcol.exporter.otlp.tempo.input]
            }
          }

          // Export metrics to Mimir
          otelcol.exporter.prometheus "mimir" {
            forward_to = [prometheus.remote_write.mimir.receiver]
          }

          prometheus.remote_write "mimir" {
            endpoint {
              url = "http://%(mimir)s/api/v1/push"
            }
          }

          // Export logs to Loki
          otelcol.exporter.loki "default" {
            forward_to = [loki.write.default.receiver]
          }

          loki.write "default" {
            endpoint {
              url = "http://%(loki)s/loki/api/v1/push"
            }
          }

          // Export traces to Tempo
          otelcol.exporter.otlp "tempo" {
            client {
              endpoint = "%(tempo)s"
              tls {
                insecure = true
              }
            }
          }
        ||| % alloy._config.alloy.exporters,
      },
    },

    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: alloy._config.alloy.name,
        namespace: alloy._config.alloy.namespace,
        labels: {
          app: alloy._config.alloy.name,
          component: 'collector',
        },
      },
      spec: {
        selector: {
          app: alloy._config.alloy.name,
          component: 'collector',
        },
        ports: [
          { name: 'otlp-grpc', port: 4317, targetPort: 4317, protocol: 'TCP' },
          { name: 'otlp-http', port: 4318, targetPort: 4318, protocol: 'TCP' },
          { name: 'metrics', port: 12345, targetPort: 12345, protocol: 'TCP' },
        ],
      },
    },

    all: [
      alloy.configMap,
      alloy.deployment,
      alloy.service,
    ],
  },
}