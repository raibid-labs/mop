// Production environment for MOP
// Uses full resources and S3 storage with HA

local config = import '../../lib/config.libsonnet';
local namespace = import '../../lib/kubernetes/namespace.libsonnet';
local rbac = import '../../lib/kubernetes/rbac.libsonnet';
local storage = import '../../lib/kubernetes/storage.libsonnet';
local network = import '../../lib/kubernetes/network.libsonnet';

local rbacResources = rbac.new(config.environments.production.namespace);
local networkPolicies = network.new(config.environments.production.namespace);

{
  _config:: config.environments.production,

  // Create namespace with proper labels
  namespace: namespace.new(self._config.namespace, {
    environment: 'production',
    'mop.io/version': config.version,
  }),

  // Storage classes for production (fast SSD with retention)
  storage: storage.new()['prod-storage'],

} + rbacResources + networkPolicies

  // TODO: Add component deployments
  // Components will be added in subsequent workstreams:
  // - OBI DaemonSet (Workstream 2)
  // - Alloy StatefulSet with operator (Workstream 2)
  // - Tempo StatefulSet (Workstream 3)
  // - Mimir StatefulSet (Workstream 3)
  // - Loki StatefulSet (Workstream 3)
  // - Grafana Deployment (Workstream 3)
