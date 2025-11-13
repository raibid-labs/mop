// Minimal observability stack example
// Demonstrates deploying just Alloy + Tempo for distributed tracing
local alloy = import '../alloy.libsonnet';
local tempo = import '../tempo.libsonnet';
local grafana = import '../grafana.libsonnet';
local config = import '../config.libsonnet';

local environment = config.environments.dev;

{
  // Alloy for OTLP ingestion
  alloy: alloy.new(environment),

  // Tempo for trace storage
  tempo: tempo.new(environment),

  // Grafana for visualization
  grafana: grafana.new(environment),
}
