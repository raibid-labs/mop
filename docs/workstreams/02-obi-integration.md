# Workstream 2: OBI Integration

## Status
🔴 Not Started

## Overview
Deploy and configure OpenBSD Network Instrumentation (OBI) as a DaemonSet across the Kubernetes cluster for eBPF-based observability. OBI will collect low-level network and system metrics, export them via OTLP to the Grafana stack, and provide deep visibility into cluster operations without instrumentation overhead.

## Objectives
- [ ] Deploy OBI DaemonSet with proper node affinity and tolerations
- [ ] Configure eBPF programs for network and system monitoring
- [ ] Set up OTLP export to Tempo, Mimir, and Loki
- [ ] Validate OBI data collection and export
- [ ] Monitor OBI agent health and performance
- [ ] Implement OBI dashboards for visibility

## Agent Assignment
**Suggested Agent Type**: `backend-dev`, `system-architect`, `perf-analyzer`
**Skill Requirements**: eBPF/BPF, Kubernetes DaemonSets, OTLP protocol, observability architecture, Linux kernel internals

## Dependencies
- Workstream 1 must complete namespace and RBAC configuration
- Kubernetes nodes must support eBPF (kernel 4.18+)
- Grafana stack endpoints must be available (can use temporary endpoints)
- Storage classes configured for OBI state persistence

## Tasks

### Task 2.1: OBI DaemonSet Deployment
**Description**: Create and deploy OBI as a Kubernetes DaemonSet with privileged access for eBPF operations.

**Deliverables**:
- OBI DaemonSet manifest with security context
- ConfigMap for OBI configuration
- Node selector and tolerations for specific node pools
- Resource limits and requests
- Init containers for eBPF program loading

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/k8s/obi/daemonset.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/obi/configmap.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/obi/service.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/obi/rbac.yaml`
- `/Users/beengud/raibid-labs/mop/tanka/lib/obi/main.libsonnet`

**Validation**:
```bash
# Deploy OBI DaemonSet
kubectl apply -f /Users/beengud/raibid-labs/mop/k8s/obi/ -n mop-system

# Verify DaemonSet rollout
kubectl rollout status daemonset/obi -n mop-system

# Check pod status on all nodes
kubectl get pods -n mop-system -l app=obi -o wide

# Verify eBPF programs loaded
kubectl exec -n mop-system daemonset/obi -- bpftool prog list

# Check OBI logs
kubectl logs -n mop-system -l app=obi --tail=50
```

### Task 2.2: eBPF Configuration
**Description**: Configure eBPF programs for network tracing, syscall monitoring, and performance metrics collection.

**Deliverables**:
- eBPF program configuration for network packet inspection
- Syscall tracing configuration
- TCP/UDP connection tracking
- Process-level network attribution
- Kernel probe attachment points
- Safety limits and resource constraints

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/config/obi/ebpf-network.yaml`
- `/Users/beengud/raibid-labs/mop/config/obi/ebpf-syscall.yaml`
- `/Users/beengud/raibid-labs/mop/config/obi/ebpf-tcp.yaml`
- `/Users/beengud/raibid-labs/mop/config/obi/ebpf-limits.yaml`
- `/Users/beengud/raibid-labs/mop/docs/obi-ebpf-programs.md`

**Validation**:
```bash
# Verify eBPF programs are attached
kubectl exec -n mop-system daemonset/obi -- bpftool prog show

# Check eBPF maps
kubectl exec -n mop-system daemonset/obi -- bpftool map list

# Verify network tracing
kubectl exec -n mop-system daemonset/obi -- cat /sys/kernel/debug/tracing/trace_pipe | head -20

# Test syscall monitoring
kubectl exec -n mop-system daemonset/obi -- cat /proc/kallsyms | grep sys_enter

# Check OBI metrics endpoint
kubectl port-forward -n mop-system daemonset/obi 9090:9090 &
curl localhost:9090/metrics | grep obi_
```

### Task 2.3: OTLP Export Configuration
**Description**: Configure OBI to export telemetry data via OTLP to Tempo (traces), Mimir (metrics), and Loki (logs).

**Deliverables**:
- OTLP exporter configuration for traces
- OTLP exporter configuration for metrics
- OTLP exporter configuration for logs
- Resource attributes and semantic conventions
- Export batch configuration and retry logic
- TLS configuration for secure transport

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/config/obi/otlp-traces.yaml`
- `/Users/beengud/raibid-labs/mop/config/obi/otlp-metrics.yaml`
- `/Users/beengud/raibid-labs/mop/config/obi/otlp-logs.yaml`
- `/Users/beengud/raibid-labs/mop/config/obi/otlp-resource-attributes.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/obi/otlp-secret.yaml`

**Validation**:
```bash
# Test OTLP trace export
kubectl exec -n mop-system daemonset/obi -- obi test-export --type traces --endpoint tempo.mop-traces.svc.cluster.local:4317

# Test OTLP metrics export
kubectl exec -n mop-system daemonset/obi -- obi test-export --type metrics --endpoint mimir.mop-metrics.svc.cluster.local:4317

# Test OTLP logs export
kubectl exec -n mop-system daemonset/obi -- obi test-export --type logs --endpoint loki.mop-logs.svc.cluster.local:4317

# Verify TLS certificates
kubectl get secret -n mop-system obi-otlp-certs -o jsonpath='{.data.tls\.crt}' | base64 -d | openssl x509 -text -noout

# Check export metrics
kubectl port-forward -n mop-system daemonset/obi 9090:9090 &
curl localhost:9090/metrics | grep otlp_exporter
```

### Task 2.4: OBI Health Monitoring
**Description**: Implement health checks, readiness probes, and monitoring for OBI agent health and performance.

**Deliverables**:
- Liveness and readiness probe configuration
- OBI self-monitoring metrics
- Alert rules for OBI agent failures
- Performance impact dashboards
- Resource usage tracking
- eBPF program health checks

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/k8s/obi/health-checks.yaml`
- `/Users/beengud/raibid-labs/mop/config/prometheus/obi-alerts.yaml`
- `/Users/beengud/raibid-labs/mop/dashboards/obi-health.json`
- `/Users/beengud/raibid-labs/mop/dashboards/obi-performance.json`
- `/Users/beengud/raibid-labs/mop/tests/obi/health-check.sh`

**Validation**:
```bash
# Check liveness probe
kubectl describe pod -n mop-system -l app=obi | grep Liveness

# Test readiness probe
kubectl describe pod -n mop-system -l app=obi | grep Readiness

# Verify health endpoint
kubectl port-forward -n mop-system daemonset/obi 8080:8080 &
curl localhost:8080/health
curl localhost:8080/ready

# Check resource usage
kubectl top pod -n mop-system -l app=obi

# Run health check tests
/Users/beengud/raibid-labs/mop/tests/obi/health-check.sh
```

### Task 2.5: OBI Data Validation
**Description**: Validate that OBI is correctly collecting and exporting data to the Grafana stack.

**Deliverables**:
- Trace validation queries in Tempo
- Metric validation queries in Mimir
- Log validation queries in Loki
- Data completeness checks
- Latency and throughput tests
- Sample dashboards showing OBI data

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tests/obi/validate-traces.sh`
- `/Users/beengud/raibid-labs/mop/tests/obi/validate-metrics.sh`
- `/Users/beengud/raibid-labs/mop/tests/obi/validate-logs.sh`
- `/Users/beengud/raibid-labs/mop/tests/obi/data-completeness.sh`
- `/Users/beengud/raibid-labs/mop/dashboards/obi-validation.json`

**Validation**:
```bash
# Query Tempo for OBI traces
curl -X GET "http://tempo.mop-traces.svc.cluster.local:3200/api/search?tags=service.name%3Dobi" | jq

# Query Mimir for OBI metrics
curl -X GET "http://mimir.mop-metrics.svc.cluster.local:9009/prometheus/api/v1/query?query=obi_network_packets_total" | jq

# Query Loki for OBI logs
curl -X GET "http://loki.mop-logs.svc.cluster.local:3100/loki/api/v1/query?query=%7Bapp%3D%22obi%22%7D" | jq

# Run validation tests
cd /Users/beengud/raibid-labs/mop/tests/obi
./validate-traces.sh
./validate-metrics.sh
./validate-logs.sh
./data-completeness.sh
```

### Task 2.6: OBI Documentation and Runbooks
**Description**: Create comprehensive documentation for OBI deployment, configuration, troubleshooting, and maintenance.

**Deliverables**:
- OBI architecture documentation
- Configuration reference guide
- Troubleshooting runbook
- Performance tuning guide
- eBPF program documentation
- OTLP export troubleshooting

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/docs/obi-architecture.md`
- `/Users/beengud/raibid-labs/mop/docs/obi-configuration.md`
- `/Users/beengud/raibid-labs/mop/docs/obi-troubleshooting.md`
- `/Users/beengud/raibid-labs/mop/docs/obi-performance-tuning.md`
- `/Users/beengud/raibid-labs/mop/docs/obi-ebpf-deep-dive.md`

**Validation**:
```bash
# Verify documentation completeness
cd /Users/beengud/raibid-labs/mop/docs
grep -r "TODO" obi-*.md || echo "No TODOs found"

# Check for broken links
markdown-link-check obi-*.md

# Verify code examples work
grep -A 10 '```bash' obi-*.md | bash -n
```

## Definition of Done
- [ ] OBI DaemonSet deployed and running on all nodes
- [ ] eBPF programs loaded and collecting data
- [ ] OTLP export configured for traces, metrics, and logs
- [ ] Data appearing in Tempo, Mimir, and Loki
- [ ] Health checks and monitoring operational
- [ ] All validation tests passing
- [ ] Alert rules configured for OBI failures
- [ ] Performance impact within acceptable limits (<5% CPU, <200MB RAM per node)
- [ ] Documentation complete with architecture diagrams
- [ ] Runbooks reviewed and tested
- [ ] Code reviewed by at least one team member

## Agent Coordination Hooks
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-2-obi-integration"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-2"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/k8s/obi/daemonset.yaml" --memory-key "swarm/mop/ws-2/daemonset-config"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/config/obi/ebpf-network.yaml" --memory-key "swarm/mop/ws-2/ebpf-config"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/config/obi/otlp-traces.yaml" --memory-key "swarm/mop/ws-2/otlp-config"
npx claude-flow@alpha hooks notify --message "OBI integration tasks completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-2-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

## Estimated Effort
**Duration**: 5-7 days
**Complexity**: High

## References
- [eBPF Documentation](https://ebpf.io/what-is-ebpf/)
- [OpenTelemetry Protocol (OTLP)](https://opentelemetry.io/docs/specs/otlp/)
- [Kubernetes DaemonSets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
- [BPF Compiler Collection (BCC)](https://github.com/iovisor/bcc)
- [bpftool Documentation](https://man7.org/linux/man-pages/man8/bpftool.8.html)
- [OpenTelemetry Semantic Conventions](https://opentelemetry.io/docs/specs/semconv/)

## Notes
- eBPF requires privileged mode and host PID namespace access
- Kernel version must be 4.18+ (5.8+ recommended for full feature support)
- Consider using CO-RE (Compile Once, Run Everywhere) for eBPF portability
- OBI should gracefully degrade if eBPF features are unavailable
- Monitor OBI's CPU and memory overhead carefully during initial deployment
- Some cloud providers may restrict eBPF capabilities (check CSP documentation)
- Consider using seccomp profiles to restrict OBI's syscall access
- eBPF programs should have built-in safety limits to prevent kernel issues
- Test OBI upgrade process to ensure zero downtime
- Document fallback procedures if eBPF programs fail to load
- Consider implementing sampling for high-traffic environments
- Network security policies may need adjustment for OTLP export
