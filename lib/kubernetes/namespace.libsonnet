{
  new(name, labels={}):: {
    apiVersion: 'v1',
    kind: 'Namespace',
    metadata: {
      name: name,
      labels: labels + {
        'mop.io/managed': 'true',
        'mop.io/component': 'infrastructure',
      },
    },
  },
}