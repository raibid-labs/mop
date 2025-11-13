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
    'default-deny': $.networkPolicy(
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
            $.port('TCP', 53),
            $.port('UDP', 53),
          ],
        },
      ],
      {'policy': 'default-deny'}
    ),

    // OBI Collector network policy (needs to scrape all namespaces)
    'obi-network': $.networkPolicy(
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
          from: [$.peer({matchLabels: {app: 'grafana'}})],
          ports: [$.port('TCP', 9090)],
        },
      ],
      [
        // Allow egress to all pods for metrics scraping
        {
          to: [$.peer()],
          ports: [$.port('TCP', 9090)],
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
          ports: [$.port('UDP', 53)],
        },
      ],
      {'app': 'obi'}
    ),

    // Alloy network policy
    'alloy-network': $.networkPolicy(
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
          from: [$.peer()],
          ports: [
            $.port('TCP', 4317),  // OTLP gRPC
            $.port('TCP', 4318),  // OTLP HTTP
            $.port('TCP', 9411),  // Zipkin
            $.port('TCP', 14268), // Jaeger
          ],
        },
      ],
      [
        // Allow egress to Tempo, Mimir, Loki
        {
          to: [
            $.peer({matchLabels: {app: 'tempo'}}),
            $.peer({matchLabels: {app: 'mimir'}}),
            $.peer({matchLabels: {app: 'loki'}}),
          ],
          ports: [
            $.port('TCP', 9095),  // Tempo
            $.port('TCP', 9009),  // Mimir
            $.port('TCP', 3100),  // Loki
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
          ports: [$.port('UDP', 53)],
        },
      ],
      {'app': 'alloy'}
    ),

    // Tempo network policy
    'tempo-network': $.networkPolicy(
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
          from: [$.peer({matchLabels: {app: 'alloy'}})],
          ports: [$.port('TCP', 9095)],
        },
        // Allow ingress from Grafana
        {
          from: [$.peer({matchLabels: {app: 'grafana'}})],
          ports: [$.port('TCP', 3200)],
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
          ports: [$.port('UDP', 53)],
        },
      ],
      {'app': 'tempo'}
    ),

    // Mimir network policy
    'mimir-network': $.networkPolicy(
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
          from: [$.peer({matchLabels: {app: 'alloy'}})],
          ports: [$.port('TCP', 9009)],
        },
        // Allow ingress from Grafana
        {
          from: [$.peer({matchLabels: {app: 'grafana'}})],
          ports: [$.port('TCP', 9009)],
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
          ports: [$.port('UDP', 53)],
        },
      ],
      {'app': 'mimir'}
    ),

    // Loki network policy
    'loki-network': $.networkPolicy(
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
          from: [$.peer({matchLabels: {app: 'alloy'}})],
          ports: [$.port('TCP', 3100)],
        },
        // Allow ingress from Grafana
        {
          from: [$.peer({matchLabels: {app: 'grafana'}})],
          ports: [$.port('TCP', 3100)],
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
          ports: [$.port('UDP', 53)],
        },
      ],
      {'app': 'loki'}
    ),

    // Grafana network policy
    'grafana-network': $.networkPolicy(
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
          ports: [$.port('TCP', 3000)],
        },
      ],
      [
        // Allow egress to Tempo, Mimir, Loki, OBI
        {
          to: [
            $.peer({matchLabels: {app: 'tempo'}}),
            $.peer({matchLabels: {app: 'mimir'}}),
            $.peer({matchLabels: {app: 'loki'}}),
            $.peer({matchLabels: {app: 'obi'}}),
          ],
          ports: [
            $.port('TCP', 3200),  // Tempo
            $.port('TCP', 9009),  // Mimir
            $.port('TCP', 3100),  // Loki
            $.port('TCP', 9090),  // OBI
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
          ports: [$.port('UDP', 53)],
        },
      ],
      {'app': 'grafana'}
    ),
  },
}