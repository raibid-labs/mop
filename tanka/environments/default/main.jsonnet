// Default environment for MOP
// Single environment for all deployments

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';

local rbacResources = rbac.new(config.namespace);
local networkPolicies = network.new(config.namespace);

{
  _config:: config,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    'mop.io/version': config.version,
  }),

  // Storage configuration
  storage: storage.new()['standard-storage'],

} + rbacResources + networkPolicies

  // TODO: Add component deployments
  // Components will be added in subsequent workstreams:
  // - OBI DaemonSet (Workstream 2)
  // - Alloy StatefulSet (Workstream 2)
  // - Tempo StatefulSet (Workstream 3)
  // - Mimir StatefulSet (Workstream 3)
  // - Loki StatefulSet (Workstream 3)
  // - Grafana Deployment (Workstream 3)
