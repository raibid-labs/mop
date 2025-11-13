// Grafana (Visualization) component library
// Provides dashboards and visualization for observability data with Tempo, Mimir, and Loki integration
local config = import 'config.libsonnet';

{
  new(envConfig):: {
    local grafanaConfig = config.grafana,
    local version = config.versions.grafana,

    // ConfigMap for Grafana datasources with trace-to-logs and metrics correlation
    datasourcesConfigMap: {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: 'grafana-datasources',
        namespace: envConfig.namespace,
        labels: config.commonLabels + { component: 'grafana' },
      },
      data: {
        'datasources.yaml': std.manifestYamlDoc({
          apiVersion: 1,
          datasources: [
            {
              name: 'Tempo',
              type: 'tempo',
              access: 'proxy',
              url: 'http://tempo.' + envConfig.namespace + '.svc.cluster.local:3200',
              uid: 'tempo',
              isDefault: false,
              jsonData: {
                tracesToLogsV2: {
                  datasourceUid: 'loki',
                  spanStartTimeShift: '-1h',
                  spanEndTimeShift: '1h',
                  tags: ['job', 'instance', 'pod', 'namespace'],
                  filterByTraceID: false,
                  filterBySpanID: false,
                  customQuery: true,
                  query: '{$${__tags}} |= "$${__span.traceId}"',
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
                tracesToMetrics: {
                  datasourceUid: 'mimir',
                  spanStartTimeShift: '-1h',
                  spanEndTimeShift: '1h',
                  tags: [{ key: 'service.name', value: 'service' }],
                  queries: [
                    {
                      name: 'Rate',
                      query: 'sum(rate(spans_total{$$__tags}[5m]))',
                    },
                    {
                      name: 'Error Rate',
                      query: 'sum(rate(spans_total{$$__tags,status_code="ERROR"}[5m]))',
                    },
                  ],
                },
              },
            },
            {
              name: 'Mimir',
              uid: 'mimir',
              type: 'prometheus',
              access: 'proxy',
              url: 'http://mimir.' + envConfig.namespace + '.svc.cluster.local:8080/prometheus',
              isDefault: true,
              jsonData: {
                exemplarTraceIdDestinations: [
                  {
                    datasourceUid: 'tempo',
                    name: 'trace_id',
                  },
                ],
                httpMethod: 'POST',
              },
            },
            {
              name: 'Loki',
              uid: 'loki',
              type: 'loki',
              access: 'proxy',
              url: 'http://loki.' + envConfig.namespace + '.svc.cluster.local:3100',
              isDefault: false,
              jsonData: {
                maxLines: 5000,
                derivedFields: [
                  {
                    datasourceUid: 'tempo',
                    matcherRegex: 'trace_id=(\\w+)',
                    name: 'TraceID',
                    url: '$${__value.raw}',
                  },
                  {
                    datasourceUid: 'tempo',
                    matcherRegex: 'traceID=(\\w+)',
                    name: 'TraceID',
                    url: '$${__value.raw}',
                  },
                ],
              },
            },
            {
              name: 'Prometheus',
              uid: 'prometheus',
              type: 'prometheus',
              access: 'proxy',
              url: 'http://prometheus.' + envConfig.namespace + '.svc.cluster.local:9090',
              isDefault: false,
            },
          ],
        }),
      },
    },

    // ConfigMap for Grafana configuration
    configMap: {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: 'grafana-config',
        namespace: envConfig.namespace,
        labels: config.commonLabels + { component: 'grafana' },
      },
      data: {
        'grafana.ini': |||
          [server]
          http_port = 3000
          domain = %(domain)s
          root_url = http://%(domain)s

          [database]
          type = sqlite3
          path = /var/lib/grafana/grafana.db

          [auth]
          disable_login_form = %(disable_login)s

          [auth.anonymous]
          enabled = %(anonymous_enabled)s
          org_role = %(anonymous_role)s

          [security]
          admin_user = admin
          admin_password = admin

          [users]
          allow_sign_up = false

          [analytics]
          reporting_enabled = false
          check_for_updates = false

          [log]
          mode = console
          level = info

          [feature_toggles]
          enable = tempoSearch tempoBackendSearch tempoApmTable traceToMetrics
        ||| % {
          domain: envConfig.domain,
          disable_login: grafanaConfig.auth.disable_login_form,
          anonymous_enabled: grafanaConfig.auth.anonymous.enabled,
          anonymous_role: grafanaConfig.auth.anonymous.org_role,
        },
      },
    },

    // Deployment for Grafana
    deployment: {
      apiVersion: 'apps/v1',
      kind: 'Deployment',
      metadata: {
        name: 'grafana',
        namespace: envConfig.namespace,
        labels: config.commonLabels + { component: 'grafana' },
      },
      spec: {
        replicas: envConfig.replicas.grafana,
        selector: { matchLabels: { app: 'grafana' } },
        template: {
          metadata: {
            labels: config.commonLabels + { app: 'grafana', component: 'grafana' },
          },
          spec: {
            containers: [{
              name: 'grafana',
              image: 'grafana/grafana:' + version,
              ports: [
                { containerPort: 3000, name: 'http', protocol: 'TCP' },
              ],
              env: [
                {
                  name: 'GF_PATHS_CONFIG',
                  value: '/etc/grafana/grafana.ini',
                },
                {
                  name: 'GF_PATHS_PROVISIONING',
                  value: '/etc/grafana/provisioning',
                },
              ],
              resources: envConfig.resources.grafana,
              volumeMounts: [
                {
                  name: 'config',
                  mountPath: '/etc/grafana',
                },
                {
                  name: 'datasources',
                  mountPath: '/etc/grafana/provisioning/datasources',
                },
                {
                  name: 'data',
                  mountPath: '/var/lib/grafana',
                },
              ],
              livenessProbe: {
                httpGet: {
                  path: '/api/health',
                  port: 3000,
                },
                initialDelaySeconds: 30,
                periodSeconds: 10,
              },
              readinessProbe: {
                httpGet: {
                  path: '/api/health',
                  port: 3000,
                },
                initialDelaySeconds: 10,
                periodSeconds: 5,
              },
            }],
            volumes: [
              {
                name: 'config',
                configMap: {
                  name: 'grafana-config',
                },
              },
              {
                name: 'datasources',
                configMap: {
                  name: 'grafana-datasources',
                },
              },
              {
                name: 'data',
                emptyDir: {},
              },
            ],
          },
        },
      },
    },

    // Service for Grafana
    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: 'grafana',
        namespace: envConfig.namespace,
        labels: config.commonLabels + { component: 'grafana' },
      },
      spec: {
        selector: { app: 'grafana' },
        ports: [
          { port: 3000, name: 'http', protocol: 'TCP', targetPort: 3000 },
        ],
        type: 'LoadBalancer',  // Change to NodePort or ClusterIP with Ingress in production
      },
    },
  },
}
