// Staging environment for MOP
// Uses moderate resources and S3 storage

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';

local rbacResources = rbac.new(config.environments.staging.namespace);
local networkPolicies = network.new(config.environments.staging.namespace);

{
  _config:: config.environments.staging,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    environment: 'staging',
    'mop.io/version': config.version,
  }),

  // Storage classes for staging (use standard storage)
  storage: storage.new()['dev-storage'],

} + rbacResources + networkPolicies

  // TODO: Add component deployments
  // Components will be added in subsequent workstreams:
  // - OBI DaemonSet (Workstream 2)
  // - Alloy StatefulSet (Workstream 2)
  // - Tempo StatefulSet (Workstream 3)
  // - Mimir StatefulSet (Workstream 3)
  // - Loki StatefulSet (Workstream 3)
  // - Grafana Deployment (Workstream 3)
