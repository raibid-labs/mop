// Full observability stack example
// This demonstrates how to use all MOP component libraries together
local alloy = import '../alloy.libsonnet';
local obi = import '../obi.libsonnet';
local tempo = import '../tempo.libsonnet';
local mimir = import '../mimir.libsonnet';
local loki = import '../loki.libsonnet';
local grafana = import '../grafana.libsonnet';
local config = import '../config.libsonnet';

// Select environment (dev, staging, or production)
local environment = config.environments.dev;

// Generate all components for the selected environment
{
  // OBI - eBPF-based instrumentation (DaemonSet)
  obi: obi.new(environment),

  // Alloy - OpenTelemetry Collector
  alloy: alloy.new(environment),

  // Tempo - Distributed Tracing
  tempo: tempo.new(environment),

  // Mimir - Metrics Storage
  mimir: mimir.new(environment),

  // Loki - Log Aggregation
  loki: loki.new(environment),

  // Grafana - Visualization
  grafana: grafana.new(environment),
}
