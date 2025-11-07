// Development environment for MOP
// Uses minimal resources and local storage

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';

{
  _config:: config.environments.dev,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    environment: 'dev',
    'mop.io/version': config.version,
  }),

  // RBAC configuration for all components
  rbac: rbac.new(self._config.namespace),

  // Storage classes for development
  storage: storage.new()['dev-storage'],

  // Network policies for component isolation
  network: network.new(self._config.namespace),

  // TODO: Add component deployments
  // Components will be added in subsequent workstreams:
  // - OBI DaemonSet (Workstream 2)
  // - Alloy StatefulSet (Workstream 2)
  // - Tempo StatefulSet (Workstream 3)
  // - Mimir StatefulSet (Workstream 3)
  // - Loki StatefulSet (Workstream 3)
  // - Grafana Deployment (Workstream 3)
}
