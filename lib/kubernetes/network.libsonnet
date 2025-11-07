{
  // NetworkPolicy creation helper
  networkPolicy(name, namespace, podSelector, policyTypes=['Ingress', 'Egress'], ingress=[], egress=[], labels={}):: {
    apiVersion: 'networking.k8s.io/v1',
    kind: 'NetworkPolicy',
    metadata: {
      name: name,
      namespace: namespace,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'network',
      },
    },
    spec: {
      podSelector: podSelector,
      policyTypes: policyTypes,
      [if std.length(ingress) > 0 then 'ingress']: ingress,
      [if std.length(egress) > 0 then 'egress']: egress,
    },
  },

  // Helper to create port specifications
  port(protocol, port, endPort=null):: {
    protocol: protocol,
    port: port,
    [if endPort != null then 'endPort']: endPort,
  },

  // Helper to create peer specifications
  peer(podSelector={}, namespaceSelector={}, ipBlock=null):: {
    [if std.length(podSelector) > 0 then 'podSelector']: podSelector,
    [if std.length(namespaceSelector) > 0 then 'namespaceSelector']: namespaceSelector,
    [if ipBlock != null then 'ipBlock']: ipBlock,
  },

  // Pre-configured network policies for MOP components
  new(namespace):: {
    // Default deny all ingress/egress for namespace
    'default-deny': self.networkPolicy(
      'default-deny-all',
      namespace,
      {},  // Empty selector applies to all pods
      ['Ingress', 'Egress'],
      [],  // No ingress allowed
      [
        // Allow DNS resolution
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
            podSelector: {
              matchLabels: {
                'k8s-app': 'kube-dns',
              },
            },
          }],
          ports: [
            self.port('TCP', 53),
            self.port('UDP', 53),
          ],
        },
      ],
      {'policy': 'default-deny'}
    ),

    // OBI Collector network policy (needs to scrape all namespaces)
    'obi-network': self.networkPolicy(
      'obi-collector',
      namespace,
      {
        matchLabels: {
          app: 'obi',
        },
      },
      ['Ingress', 'Egress'],
      [
        // Allow ingress from Grafana
        {
          from: [self.peer({matchLabels: {app: 'grafana'}})],
          ports: [self.port('TCP', 9090)],
        },
      ],
      [
        // Allow egress to all pods for metrics scraping
        {
          to: [self.peer()],
          ports: [self.port('TCP', 9090)],
        },
        // Allow DNS
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
          }],
          ports: [self.port('UDP', 53)],
        },
      ],
      {'app': 'obi'}
    ),

    // Alloy network policy
    'alloy-network': self.networkPolicy(
      'alloy',
      namespace,
      {
        matchLabels: {
          app: 'alloy',
        },
      },
      ['Ingress', 'Egress'],
      [
        // Allow ingress from pods for metrics/logs/traces
        {
          from: [self.peer()],
          ports: [
            self.port('TCP', 4317),  // OTLP gRPC
            self.port('TCP', 4318),  // OTLP HTTP
            self.port('TCP', 9411),  // Zipkin
            self.port('TCP', 14268), // Jaeger
          ],
        },
      ],
      [
        // Allow egress to Tempo, Mimir, Loki
        {
          to: [
            self.peer({matchLabels: {app: 'tempo'}}),
            self.peer({matchLabels: {app: 'mimir'}}),
            self.peer({matchLabels: {app: 'loki'}}),
          ],
          ports: [
            self.port('TCP', 9095),  // Tempo
            self.port('TCP', 9009),  // Mimir
            self.port('TCP', 3100),  // Loki
          ],
        },
        // Allow DNS
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
          }],
          ports: [self.port('UDP', 53)],
        },
      ],
      {'app': 'alloy'}
    ),

    // Tempo network policy
    'tempo-network': self.networkPolicy(
      'tempo',
      namespace,
      {
        matchLabels: {
          app: 'tempo',
        },
      },
      ['Ingress', 'Egress'],
      [
        // Allow ingress from Alloy
        {
          from: [self.peer({matchLabels: {app: 'alloy'}})],
          ports: [self.port('TCP', 9095)],
        },
        // Allow ingress from Grafana
        {
          from: [self.peer({matchLabels: {app: 'grafana'}})],
          ports: [self.port('TCP', 3200)],
        },
      ],
      [
        // Allow DNS
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
          }],
          ports: [self.port('UDP', 53)],
        },
      ],
      {'app': 'tempo'}
    ),

    // Mimir network policy
    'mimir-network': self.networkPolicy(
      'mimir',
      namespace,
      {
        matchLabels: {
          app: 'mimir',
        },
      },
      ['Ingress', 'Egress'],
      [
        // Allow ingress from Alloy
        {
          from: [self.peer({matchLabels: {app: 'alloy'}})],
          ports: [self.port('TCP', 9009)],
        },
        // Allow ingress from Grafana
        {
          from: [self.peer({matchLabels: {app: 'grafana'}})],
          ports: [self.port('TCP', 9009)],
        },
      ],
      [
        // Allow DNS
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
          }],
          ports: [self.port('UDP', 53)],
        },
      ],
      {'app': 'mimir'}
    ),

    // Loki network policy
    'loki-network': self.networkPolicy(
      'loki',
      namespace,
      {
        matchLabels: {
          app: 'loki',
        },
      },
      ['Ingress', 'Egress'],
      [
        // Allow ingress from Alloy
        {
          from: [self.peer({matchLabels: {app: 'alloy'}})],
          ports: [self.port('TCP', 3100)],
        },
        // Allow ingress from Grafana
        {
          from: [self.peer({matchLabels: {app: 'grafana'}})],
          ports: [self.port('TCP', 3100)],
        },
      ],
      [
        // Allow DNS
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
          }],
          ports: [self.port('UDP', 53)],
        },
      ],
      {'app': 'loki'}
    ),

    // Grafana network policy
    'grafana-network': self.networkPolicy(
      'grafana',
      namespace,
      {
        matchLabels: {
          app: 'grafana',
        },
      },
      ['Ingress', 'Egress'],
      [
        // Allow ingress from anywhere for UI access
        {
          from: [],  // Empty means from anywhere
          ports: [self.port('TCP', 3000)],
        },
      ],
      [
        // Allow egress to Tempo, Mimir, Loki, OBI
        {
          to: [
            self.peer({matchLabels: {app: 'tempo'}}),
            self.peer({matchLabels: {app: 'mimir'}}),
            self.peer({matchLabels: {app: 'loki'}}),
            self.peer({matchLabels: {app: 'obi'}}),
          ],
          ports: [
            self.port('TCP', 3200),  // Tempo
            self.port('TCP', 9009),  // Mimir
            self.port('TCP', 3100),  // Loki
            self.port('TCP', 9090),  // OBI
          ],
        },
        // Allow DNS
        {
          to: [{
            namespaceSelector: {
              matchLabels: {
                'kubernetes.io/metadata.name': 'kube-system',
              },
            },
          }],
          ports: [self.port('UDP', 53)],
        },
      ],
      {'app': 'grafana'}
    ),
  },
}