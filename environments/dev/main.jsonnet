// Development environment for MOP
// Uses minimal resources and local storage

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';

local rbacResources = rbac.new(config.environments.dev.namespace);
local networkPolicies = network.new(config.environments.dev.namespace);

{
  _config:: config.environments.dev,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    environment: 'dev',
    'mop.io/version': config.version,
  }),

  // Storage classes for development
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
