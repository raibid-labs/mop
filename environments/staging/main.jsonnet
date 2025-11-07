// Staging environment for MOP
// Uses moderate resources and S3 storage

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';
local obi = import '../../lib/obi.libsonnet';

{
  _config:: config.environments.staging,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    environment: 'staging',
    'mop.io/version': config.version,
  }),

  // RBAC configuration for all components
  rbac: rbac.new(self._config.namespace),

  // Storage classes for staging (use standard storage)
  storage: storage.new()['dev-storage'],

  // Network policies for component isolation
  network: network.new(self._config.namespace),

  // OBI eBPF instrumentation (Workstream 2)
  obi: obi.new(self._config),

  // TODO: Add remaining component deployments
  // Components will be added in subsequent workstreams:
  // - Alloy StatefulSet (Workstream 2)
  // - Tempo StatefulSet (Workstream 3)
  // - Mimir StatefulSet (Workstream 3)
  // - Loki StatefulSet (Workstream 3)
  // - Grafana Deployment (Workstream 3)
}
