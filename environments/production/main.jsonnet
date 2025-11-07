// Production environment for MOP
// Uses full resources and S3 storage with HA

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';
local obi = import '../../lib/obi.libsonnet';

{
  _config:: config.environments.production,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    environment: 'production',
    'mop.io/version': config.version,
  }),

  // RBAC configuration for all components
  rbac: rbac.new(self._config.namespace),

  // Storage classes for production (fast SSD with retention)
  storage: storage.new()['prod-storage'],

  // Network policies for component isolation
  network: network.new(self._config.namespace),

  // OBI eBPF instrumentation (Workstream 2)
  obi: obi.new(self._config),

  // TODO: Add remaining component deployments
  // Components will be added in subsequent workstreams:
  // - Alloy StatefulSet with operator (Workstream 2)
  // - Tempo StatefulSet (Workstream 3)
  // - Mimir StatefulSet (Workstream 3)
  // - Loki StatefulSet (Workstream 3)
  // - Grafana Deployment (Workstream 3)
}
