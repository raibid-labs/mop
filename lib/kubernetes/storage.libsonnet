{
  // StorageClass creation helper
  storageClass(name, provisioner, parameters={}, volumeBindingMode='WaitForFirstConsumer', reclaimPolicy='Delete', labels={}):: {
    apiVersion: 'storage.k8s.io/v1',
    kind: 'StorageClass',
    metadata: {
      name: name,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'storage',
      },
    },
    provisioner: provisioner,
    parameters: parameters,
    volumeBindingMode: volumeBindingMode,
    reclaimPolicy: reclaimPolicy,
    allowVolumeExpansion: true,
  },

  // PersistentVolumeClaim template
  pvc(name, namespace, storageClass, size, accessModes=['ReadWriteOnce'], labels={}):: {
    apiVersion: 'v1',
    kind: 'PersistentVolumeClaim',
    metadata: {
      name: name,
      namespace: namespace,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'storage',
      },
    },
    spec: {
      storageClassName: storageClass,
      accessModes: accessModes,
      resources: {
        requests: {
          storage: size,
        },
      },
    },
  },

  // Pre-configured storage classes for different environments
  new():: {
    // Development storage class (standard performance)
    'dev-storage': $.storageClass(
      'mop-standard',
      'kubernetes.io/gce-pd',  // GKE example, adjust for your cloud provider
      {
        type: 'pd-standard',
        replication-type: 'none',
      },
      'WaitForFirstConsumer',
      'Delete',
      {'environment': 'dev'}
    ),

    // Production storage class (fast SSD)
    'prod-storage': $.storageClass(
      'mop-fast-ssd',
      'kubernetes.io/gce-pd',  // GKE example, adjust for your cloud provider
      {
        type: 'pd-ssd',
        replication-type: 'regional-pd',
      },
      'WaitForFirstConsumer',
      'Retain',  // Retain for production data safety
      {'environment': 'prod'}
    ),

    // For AWS EKS
    'eks-dev-storage': $.storageClass(
      'mop-standard',
      'ebs.csi.aws.com',
      {
        type: 'gp3',
        fsType: 'ext4',
      },
      'WaitForFirstConsumer',
      'Delete',
      {'environment': 'dev', 'provider': 'aws'}
    ),

    'eks-prod-storage': $.storageClass(
      'mop-fast-ssd',
      'ebs.csi.aws.com',
      {
        type: 'io2',
        iopsPerGB: '50',
        fsType: 'ext4',
      },
      'WaitForFirstConsumer',
      'Retain',
      {'environment': 'prod', 'provider': 'aws'}
    ),

    // For Azure AKS
    'aks-dev-storage': $.storageClass(
      'mop-standard',
      'disk.csi.azure.com',
      {
        skuName: 'StandardSSD_LRS',
      },
      'WaitForFirstConsumer',
      'Delete',
      {'environment': 'dev', 'provider': 'azure'}
    ),

    'aks-prod-storage': $.storageClass(
      'mop-fast-ssd',
      'disk.csi.azure.com',
      {
        skuName: 'Premium_LRS',
      },
      'WaitForFirstConsumer',
      'Retain',
      {'environment': 'prod', 'provider': 'azure'}
    ),

    // Generic local storage (for on-prem or testing)
    'local-storage': $.storageClass(
      'mop-local',
      'kubernetes.io/no-provisioner',
      {},
      'WaitForFirstConsumer',
      'Delete',
      {'environment': 'local'}
    ),
  },

  // PVC templates for components
  pvcTemplates(namespace, storageClass):: {
    'tempo-pvc': $.pvc('tempo-data', namespace, storageClass, '10Gi', ['ReadWriteOnce'], {'app': 'tempo'}),
    'mimir-pvc': $.pvc('mimir-data', namespace, storageClass, '50Gi', ['ReadWriteOnce'], {'app': 'mimir'}),
    'loki-pvc': $.pvc('loki-data', namespace, storageClass, '30Gi', ['ReadWriteOnce'], {'app': 'loki'}),
    'grafana-pvc': $.pvc('grafana-data', namespace, storageClass, '5Gi', ['ReadWriteOnce'], {'app': 'grafana'}),
  },
}