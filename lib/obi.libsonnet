// OpenTelemetry Backend Initiative (OBI) eBPF instrumentation
// Provides automatic, zero-code instrumentation for network and system observability

{
  new(config):: {
    local obi = self,

    // Configuration for OBI
    _config:: {
      namespace: config.namespace,
      name: 'obi',
      image: 'otel/opentelemetry-ebpf:latest',
      replicas: 1,  // DaemonSet, so one per node

      // Resource limits
      resources: {
        requests: {
          cpu: '100m',
          memory: '128Mi',
        },
        limits: {
          cpu: '500m',
          memory: '512Mi',
        },
      },

      // OTLP export endpoints
      otlp: {
        endpoint: 'alloy.%s.svc.cluster.local:4317' % config.namespace,
        insecure: true,  // Within cluster communication
      },

      // eBPF configuration
      ebpf: {
        protocols: ['HTTP', 'gRPC', 'SQL', 'Redis', 'Kafka'],
        syscalls: true,
        network: true,
        tcp: true,
        udp: true,
      },
    },

    // ConfigMap for OBI configuration
    configMap: {
      apiVersion: 'v1',
      kind: 'ConfigMap',
      metadata: {
        name: obi._config.name,
        namespace: obi._config.namespace,
        labels: {
          app: obi._config.name,
          component: 'ebpf-instrumentation',
        },
      },
      data: {
        'config.yaml': std.manifestYamlDoc({
          exporters: {
            otlp: {
              endpoint: obi._config.otlp.endpoint,
              insecure: obi._config.otlp.insecure,
              headers: {
                'x-service': 'obi',
                'x-namespace': obi._config.namespace,
              },
            },
          },
          processors: {
            batch: {
              timeout: '5s',
              send_batch_size: 100,
            },
            resource: {
              attributes: [
                {
                  key: 'service.name',
                  value: 'obi',
                  action: 'insert',
                },
                {
                  key: 'service.namespace',
                  value: obi._config.namespace,
                  action: 'insert',
                },
                {
                  key: 'telemetry.sdk.name',
                  value: 'obi-ebpf',
                  action: 'insert',
                },
              ],
            },
          },
          receivers: {
            ebpf: {
              protocols: obi._config.ebpf.protocols,
              syscalls: obi._config.ebpf.syscalls,
              network: obi._config.ebpf.network,
              tcp: obi._config.ebpf.tcp,
              udp: obi._config.ebpf.udp,
              sampling_rate: 1.0,
            },
          },
          service: {
            pipelines: {
              traces: {
                receivers: ['ebpf'],
                processors: ['batch', 'resource'],
                exporters: ['otlp'],
              },
              metrics: {
                receivers: ['ebpf'],
                processors: ['batch', 'resource'],
                exporters: ['otlp'],
              },
              logs: {
                receivers: ['ebpf'],
                processors: ['batch', 'resource'],
                exporters: ['otlp'],
              },
            },
          },
        }),
      },
    },

    // ServiceAccount for OBI
    serviceAccount: {
      apiVersion: 'v1',
      kind: 'ServiceAccount',
      metadata: {
        name: obi._config.name,
        namespace: obi._config.namespace,
        labels: {
          app: obi._config.name,
        },
      },
    },

    // ClusterRole for OBI (needs privileged access for eBPF)
    clusterRole: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRole',
      metadata: {
        name: obi._config.name,
        labels: {
          app: obi._config.name,
        },
      },
      rules: [
        {
          apiGroups: [''],
          resources: ['nodes', 'nodes/proxy'],
          verbs: ['get', 'list'],
        },
        {
          apiGroups: [''],
          resources: ['pods', 'endpoints', 'services'],
          verbs: ['get', 'list', 'watch'],
        },
      ],
    },

    // ClusterRoleBinding for OBI
    clusterRoleBinding: {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'ClusterRoleBinding',
      metadata: {
        name: obi._config.name,
        labels: {
          app: obi._config.name,
        },
      },
      roleRef: {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: obi._config.name,
      },
      subjects: [
        {
          kind: 'ServiceAccount',
          name: obi._config.name,
          namespace: obi._config.namespace,
        },
      ],
    },

    // DaemonSet for OBI
    daemonSet: {
      apiVersion: 'apps/v1',
      kind: 'DaemonSet',
      metadata: {
        name: obi._config.name,
        namespace: obi._config.namespace,
        labels: {
          app: obi._config.name,
          component: 'ebpf-instrumentation',
        },
      },
      spec: {
        selector: {
          matchLabels: {
            app: obi._config.name,
          },
        },
        template: {
          metadata: {
            labels: {
              app: obi._config.name,
              component: 'ebpf-instrumentation',
            },
            annotations: {
              'prometheus.io/scrape': 'true',
              'prometheus.io/port': '8888',
              'prometheus.io/path': '/metrics',
            },
          },
          spec: {
            serviceAccountName: obi._config.name,
            hostNetwork: true,
            hostPID: true,
            dnsPolicy: 'ClusterFirstWithHostNet',

            // Init container to verify kernel compatibility
            initContainers: [
              {
                name: 'verify-kernel',
                image: 'busybox:latest',
                command: ['sh', '-c'],
                args: [|||
                  kernel_version=$(uname -r | cut -d. -f1,2)
                  major=$(echo $kernel_version | cut -d. -f1)
                  minor=$(echo $kernel_version | cut -d. -f2)

                  if [ "$major" -lt 4 ] || ([ "$major" -eq 4 ] && [ "$minor" -lt 18 ]); then
                    echo "Kernel version $kernel_version is too old. Minimum required: 4.18"
                    exit 1
                  fi

                  echo "Kernel version $kernel_version is compatible"
                |||],
              },
            ],

            containers: [
              {
                name: obi._config.name,
                image: obi._config.image,
                imagePullPolicy: 'IfNotPresent',

                securityContext: {
                  privileged: true,
                  capabilities: {
                    add: [
                      'SYS_ADMIN',
                      'SYS_RESOURCE',
                      'SYS_PTRACE',
                      'NET_ADMIN',
                      'IPC_LOCK',
                    ],
                  },
                },

                volumeMounts: [
                  {
                    name: 'config',
                    mountPath: '/etc/obi',
                  },
                  {
                    name: 'sys',
                    mountPath: '/sys',
                    readOnly: true,
                  },
                  {
                    name: 'cgroup',
                    mountPath: '/sys/fs/cgroup',
                    readOnly: true,
                  },
                  {
                    name: 'debugfs',
                    mountPath: '/sys/kernel/debug',
                  },
                  {
                    name: 'bpf',
                    mountPath: '/sys/fs/bpf',
                  },
                ],

                env: [
                  {
                    name: 'NODE_NAME',
                    valueFrom: {
                      fieldRef: {
                        fieldPath: 'spec.nodeName',
                      },
                    },
                  },
                  {
                    name: 'POD_NAME',
                    valueFrom: {
                      fieldRef: {
                        fieldPath: 'metadata.name',
                      },
                    },
                  },
                  {
                    name: 'POD_NAMESPACE',
                    valueFrom: {
                      fieldRef: {
                        fieldPath: 'metadata.namespace',
                      },
                    },
                  },
                ],

                args: [
                  '--config=/etc/obi/config.yaml',
                  '--metrics-address=:8888',
                  '--health-check-address=:13133',
                ],

                ports: [
                  {
                    containerPort: 8888,
                    name: 'metrics',
                    protocol: 'TCP',
                  },
                  {
                    containerPort: 13133,
                    name: 'health',
                    protocol: 'TCP',
                  },
                ],

                livenessProbe: {
                  httpGet: {
                    path: '/health',
                    port: 13133,
                  },
                  initialDelaySeconds: 30,
                  periodSeconds: 30,
                  timeoutSeconds: 5,
                  failureThreshold: 3,
                },

                readinessProbe: {
                  httpGet: {
                    path: '/ready',
                    port: 13133,
                  },
                  initialDelaySeconds: 10,
                  periodSeconds: 10,
                  timeoutSeconds: 5,
                  failureThreshold: 3,
                },

                resources: obi._config.resources,
              },
            ],

            volumes: [
              {
                name: 'config',
                configMap: {
                  name: obi._config.name,
                },
              },
              {
                name: 'sys',
                hostPath: {
                  path: '/sys',
                },
              },
              {
                name: 'cgroup',
                hostPath: {
                  path: '/sys/fs/cgroup',
                },
              },
              {
                name: 'debugfs',
                hostPath: {
                  path: '/sys/kernel/debug',
                },
              },
              {
                name: 'bpf',
                hostPath: {
                  path: '/sys/fs/bpf',
                },
              },
            ],

            tolerations: [
              {
                effect: 'NoSchedule',
                operator: 'Exists',
              },
              {
                effect: 'NoExecute',
                operator: 'Exists',
              },
            ],
          },
        },
      },
    },

    // Service for OBI metrics
    service: {
      apiVersion: 'v1',
      kind: 'Service',
      metadata: {
        name: obi._config.name + '-metrics',
        namespace: obi._config.namespace,
        labels: {
          app: obi._config.name,
          component: 'metrics',
        },
      },
      spec: {
        type: 'ClusterIP',
        clusterIP: 'None',  // Headless service for DaemonSet
        selector: {
          app: obi._config.name,
        },
        ports: [
          {
            name: 'metrics',
            port: 8888,
            targetPort: 8888,
            protocol: 'TCP',
          },
        ],
      },
    },
  },
}