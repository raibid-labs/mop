{
  // ServiceAccount creation
  serviceAccount(name, namespace, labels={}):: {
    apiVersion: 'v1',
    kind: 'ServiceAccount',
    metadata: {
      name: name,
      namespace: namespace,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'rbac',
      },
    },
  },

  // ClusterRole creation
  clusterRole(name, rules, labels={}):: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'ClusterRole',
    metadata: {
      name: name,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'rbac',
      },
    },
    rules: rules,
  },

  // RoleBinding creation
  roleBinding(name, namespace, roleRef, subjects, labels={}):: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'RoleBinding',
    metadata: {
      name: name,
      namespace: namespace,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'rbac',
      },
    },
    roleRef: roleRef,
    subjects: subjects,
  },

  // ClusterRoleBinding creation
  clusterRoleBinding(name, roleRef, subjects, labels={}):: {
    apiVersion: 'rbac.authorization.k8s.io/v1',
    kind: 'ClusterRoleBinding',
    metadata: {
      name: name,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'rbac',
      },
    },
    roleRef: roleRef,
    subjects: subjects,
  },

  // Pre-configured RBAC for MOP components
  new(namespace):: {
    // ServiceAccounts
    'obi-sa': $.serviceAccount('obi-collector', namespace, {'app': 'obi'}),
    'alloy-sa': $.serviceAccount('alloy', namespace, {'app': 'alloy'}),
    'tempo-sa': $.serviceAccount('tempo', namespace, {'app': 'tempo'}),
    'mimir-sa': $.serviceAccount('mimir', namespace, {'app': 'mimir'}),
    'loki-sa': $.serviceAccount('loki', namespace, {'app': 'loki'}),
    'grafana-sa': $.serviceAccount('grafana', namespace, {'app': 'grafana'}),

    // OBI ClusterRole (privileged for eBPF operations)
    'obi-cr': $.clusterRole('obi-collector', [
      {
        apiGroups: [''],
        resources: ['nodes', 'pods', 'services', 'endpoints', 'configmaps'],
        verbs: ['get', 'list', 'watch'],
      },
      {
        apiGroups: ['apps'],
        resources: ['deployments', 'daemonsets', 'statefulsets'],
        verbs: ['get', 'list', 'watch'],
      },
      {
        apiGroups: [''],
        resources: ['nodes/proxy'],
        verbs: ['get'],
      },
    ], {'app': 'obi'}),

    // Alloy ClusterRole (for scraping metrics)
    'alloy-cr': $.clusterRole('alloy', [
      {
        apiGroups: [''],
        resources: ['nodes', 'nodes/metrics', 'pods', 'services', 'endpoints'],
        verbs: ['get', 'list', 'watch'],
      },
      {
        apiGroups: [''],
        resources: ['configmaps'],
        verbs: ['get', 'list', 'watch', 'create', 'update', 'patch'],
      },
    ], {'app': 'alloy'}),

    // Tempo Role (namespace-scoped)
    'tempo-role': {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'Role',
      metadata: {
        name: 'tempo',
        namespace: namespace,
        labels: {
          'app': 'tempo',
          'mop.io/managed': 'true',
        },
      },
      rules: [
        {
          apiGroups: [''],
          resources: ['configmaps', 'secrets'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch'],
        },
      ],
    },

    // Mimir Role (namespace-scoped)
    'mimir-role': {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'Role',
      metadata: {
        name: 'mimir',
        namespace: namespace,
        labels: {
          'app': 'mimir',
          'mop.io/managed': 'true',
        },
      },
      rules: [
        {
          apiGroups: [''],
          resources: ['configmaps', 'secrets', 'persistentvolumeclaims'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch'],
        },
      ],
    },

    // Loki Role (namespace-scoped)
    'loki-role': {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'Role',
      metadata: {
        name: 'loki',
        namespace: namespace,
        labels: {
          'app': 'loki',
          'mop.io/managed': 'true',
        },
      },
      rules: [
        {
          apiGroups: [''],
          resources: ['configmaps', 'secrets', 'persistentvolumeclaims'],
          verbs: ['get', 'list', 'watch', 'create', 'update', 'patch'],
        },
      ],
    },

    // Grafana Role (namespace-scoped)
    'grafana-role': {
      apiVersion: 'rbac.authorization.k8s.io/v1',
      kind: 'Role',
      metadata: {
        name: 'grafana',
        namespace: namespace,
        labels: {
          'app': 'grafana',
          'mop.io/managed': 'true',
        },
      },
      rules: [
        {
          apiGroups: [''],
          resources: ['configmaps', 'secrets'],
          verbs: ['get', 'list', 'watch'],
        },
      ],
    },

    // ClusterRoleBindings
    'obi-crb': $.clusterRoleBinding(
      'obi-collector',
      {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: 'obi-collector',
      },
      [{
        kind: 'ServiceAccount',
        name: 'obi-collector',
        namespace: namespace,
      }],
      {'app': 'obi'}
    ),

    'alloy-crb': $.clusterRoleBinding(
      'alloy',
      {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'ClusterRole',
        name: 'alloy',
      },
      [{
        kind: 'ServiceAccount',
        name: 'alloy',
        namespace: namespace,
      }],
      {'app': 'alloy'}
    ),

    // RoleBindings
    'tempo-rb': $.roleBinding(
      'tempo',
      namespace,
      {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'Role',
        name: 'tempo',
      },
      [{
        kind: 'ServiceAccount',
        name: 'tempo',
        namespace: namespace,
      }],
      {'app': 'tempo'}
    ),

    'mimir-rb': $.roleBinding(
      'mimir',
      namespace,
      {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'Role',
        name: 'mimir',
      },
      [{
        kind: 'ServiceAccount',
        name: 'mimir',
        namespace: namespace,
      }],
      {'app': 'mimir'}
    ),

    'loki-rb': $.roleBinding(
      'loki',
      namespace,
      {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'Role',
        name: 'loki',
      },
      [{
        kind: 'ServiceAccount',
        name: 'loki',
        namespace: namespace,
      }],
      {'app': 'loki'}
    ),

    'grafana-rb': $.roleBinding(
      'grafana',
      namespace,
      {
        apiGroup: 'rbac.authorization.k8s.io',
        kind: 'Role',
        name: 'grafana',
      },
      [{
        kind: 'ServiceAccount',
        name: 'grafana',
        namespace: namespace,
      }],
      {'app': 'grafana'}
    ),
  },
}