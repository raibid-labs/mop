# OBI eBPF Instrumentation Deployment Guide

## Overview

OpenTelemetry Backend Initiative (OBI) provides automatic, zero-code instrumentation for observability through eBPF (extended Berkeley Packet Filter) technology. This guide covers the deployment and management of OBI as part of the MOP (Modern Observability Platform).

## Architecture

OBI runs as a Kubernetes DaemonSet, ensuring one instance per node for comprehensive system-level observability:

```
┌─────────────────────────────────────────────────┐
│                  Kubernetes Node                 │
├─────────────────────────────────────────────────┤
│  ┌─────────────────────────────────────────┐    │
│  │            OBI DaemonSet Pod            │    │
│  │  ┌─────────────────────────────────┐    │    │
│  │  │     eBPF Programs (Kernel)      │    │    │
│  │  │  • Network packet inspection    │    │    │
│  │  │  • System call monitoring       │    │    │
│  │  │  • TCP/UDP connection tracking  │    │    │
│  │  │  • Process attribution          │    │    │
│  │  └─────────────────────────────────┘    │    │
│  │              ↓                           │    │
│  │  ┌─────────────────────────────────┐    │    │
│  │  │    OBI Collector & Processor    │    │    │
│  │  │  • Data collection              │    │    │
│  │  │  • Resource enrichment          │    │    │
│  │  │  • Batching & buffering         │    │    │
│  │  └─────────────────────────────────┘    │    │
│  │              ↓                           │    │
│  │  ┌─────────────────────────────────┐    │    │
│  │  │       OTLP Exporter             │    │    │
│  │  │  → Alloy (gateway)              │    │    │
│  │  │  → Tempo (traces)               │    │    │
│  │  │  → Mimir (metrics)              │    │    │
│  │  │  → Loki (logs)                  │    │    │
│  │  └─────────────────────────────────┘    │    │
│  └─────────────────────────────────────────┘    │
└─────────────────────────────────────────────────┘
```

## Prerequisites

### Kernel Requirements
- Linux kernel version 4.18+ (minimum)
- Linux kernel version 5.8+ (recommended for full features)
- eBPF support enabled in kernel
- BTF (BPF Type Format) support for CO-RE

### Kubernetes Requirements
- Kubernetes 1.19+ cluster
- Nodes must support privileged containers
- DaemonSet deployment permissions
- ClusterRole and ClusterRoleBinding creation permissions

### Verification Commands

Check kernel version:
```bash
kubectl get nodes -o wide
# Or on a specific node:
kubectl debug node/<node-name> -it --image=busybox -- uname -r
```

Verify eBPF support:
```bash
kubectl debug node/<node-name> -it --image=busybox -- ls /sys/fs/bpf
```

## Deployment Methods

### Method 1: Using Jsonnet/Tanka

Deploy OBI using Tanka for each environment:

```bash
# Development environment
cd /Users/beengud/raibid-labs/mop
tk apply environments/dev/

# Staging environment
tk apply environments/staging/

# Production environment
tk apply environments/production/
```

### Method 2: Direct YAML Application

Deploy OBI using kubectl:

```bash
# Create namespaces
kubectl apply -f k8s/obi/namespace.yaml

# Deploy to development
kubectl apply -f k8s/obi/rbac.yaml
kubectl apply -f k8s/obi/configmap.yaml
kubectl apply -f k8s/obi/daemonset.yaml
kubectl apply -f k8s/obi/service.yaml

# For staging/production, modify the namespace in the YAML files
```

### Method 3: Using Helm (Future)

```bash
# Install OBI via Helm chart (when available)
helm repo add mop https://charts.mop.io
helm install obi mop/obi -n observability-dev --values values.yaml
```

## Configuration

### Environment Variables

OBI pods automatically receive:
- `NODE_NAME`: The Kubernetes node name
- `POD_NAME`: The pod instance name
- `POD_NAMESPACE`: The deployment namespace

### ConfigMap Settings

Key configuration in `config.yaml`:

```yaml
exporters:
  otlp:
    endpoint: alloy.observability-dev.svc.cluster.local:4317
    insecure: true  # Use TLS in production
    headers:
      x-service: obi
      x-namespace: observability-dev

receivers:
  ebpf:
    protocols:
      - HTTP
      - gRPC
      - SQL
      - Redis
      - Kafka
    syscalls: true
    network: true
    tcp: true
    udp: true
    sampling_rate: 1.0  # Adjust for high-traffic environments

processors:
  batch:
    timeout: 5s
    send_batch_size: 100
  resource:
    attributes:
      - key: service.name
        value: obi
        action: insert
```

### Resource Limits

Default resource configuration:
- **Requests**: CPU: 100m, Memory: 128Mi
- **Limits**: CPU: 500m, Memory: 512Mi

Adjust based on node capacity and traffic volume.

## Validation

### Quick Validation

Run the validation script:
```bash
./tests/obi-validation.sh
```

### Manual Validation

1. **Check DaemonSet Status**:
```bash
kubectl get daemonset obi -n observability-dev
kubectl describe daemonset obi -n observability-dev
```

2. **Verify Pod Distribution**:
```bash
kubectl get pods -n observability-dev -l app=obi -o wide
```

3. **Check Pod Logs**:
```bash
kubectl logs -n observability-dev -l app=obi --tail=50
```

4. **Test Health Endpoints**:
```bash
# Port-forward to a pod
kubectl port-forward -n observability-dev daemonset/obi 13133:13133

# In another terminal
curl http://localhost:13133/health
curl http://localhost:13133/ready
```

5. **Verify eBPF Programs**:
```bash
kubectl exec -n observability-dev daemonset/obi -- bpftool prog list
kubectl exec -n observability-dev daemonset/obi -- bpftool map list
```

## Troubleshooting

### Common Issues

#### 1. Pods Not Starting

**Symptom**: OBI pods stuck in `Init:0/1` or `ContainerCreating`

**Solution**:
```bash
# Check init container logs
kubectl logs -n observability-dev <pod-name> -c verify-kernel

# Verify kernel compatibility
kubectl get nodes -o custom-columns=NAME:.metadata.name,KERNEL:.status.nodeInfo.kernelVersion
```

#### 2. eBPF Program Load Failures

**Symptom**: Logs show "failed to load eBPF program"

**Solution**:
- Verify kernel version meets requirements
- Check for kernel security restrictions
- Ensure privileged mode is allowed:
```bash
kubectl get psp -A  # Check PodSecurityPolicies
kubectl describe psp  # Look for privileged: true
```

#### 3. OTLP Export Failures

**Symptom**: "failed to export" errors in logs

**Solution**:
```bash
# Verify Alloy is running
kubectl get pods -n observability-dev -l app=alloy

# Test connectivity
kubectl exec -n observability-dev daemonset/obi -- \
  nc -zv alloy.observability-dev.svc.cluster.local 4317
```

#### 4. High Memory Usage

**Symptom**: OBI pods using excessive memory

**Solution**:
- Reduce sampling rate in ConfigMap
- Adjust batch processor settings
- Increase memory limits if justified

### Debug Commands

Enable verbose logging:
```yaml
# In ConfigMap, add to config.yaml:
service:
  telemetry:
    logs:
      level: debug
```

Check eBPF maps usage:
```bash
kubectl exec -n observability-dev daemonset/obi -- \
  bpftool map show | grep -E "key|value|max_entries"
```

Monitor resource usage:
```bash
kubectl top pods -n observability-dev -l app=obi
```

## Performance Tuning

### Sampling Strategies

For high-traffic environments:
```yaml
receivers:
  ebpf:
    sampling_rate: 0.1  # Sample 10% of traffic
    adaptive_sampling:
      enabled: true
      target_tps: 1000  # Target traces per second
```

### Batch Processing

Optimize batching for throughput:
```yaml
processors:
  batch:
    timeout: 10s  # Increase for better compression
    send_batch_size: 500  # Larger batches
    send_batch_max_size: 1000
```

### Protocol Selection

Disable unnecessary protocols:
```yaml
receivers:
  ebpf:
    protocols:
      - HTTP
      - gRPC
      # Comment out unused protocols
      # - SQL
      # - Redis
      # - Kafka
```

## Security Considerations

### Required Capabilities

OBI requires these Linux capabilities:
- `SYS_ADMIN`: Load eBPF programs
- `SYS_RESOURCE`: Override resource limits
- `SYS_PTRACE`: Trace processes
- `NET_ADMIN`: Network operations
- `IPC_LOCK`: Lock memory pages

### Network Policies

Example NetworkPolicy for OBI:
```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: obi-network-policy
  namespace: observability-dev
spec:
  podSelector:
    matchLabels:
      app: obi
  policyTypes:
  - Egress
  egress:
  - to:
    - podSelector:
        matchLabels:
          app: alloy
    ports:
    - protocol: TCP
      port: 4317
```

### Seccomp Profiles

Consider using custom seccomp profiles for additional security:
```yaml
securityContext:
  seccompProfile:
    type: Localhost
    localhostProfile: obi-seccomp.json
```

## Monitoring OBI

### Key Metrics to Monitor

1. **eBPF Program Metrics**:
   - `obi_ebpf_programs_loaded`: Number of loaded programs
   - `obi_ebpf_events_total`: Total events captured
   - `obi_ebpf_drops_total`: Dropped events

2. **Export Metrics**:
   - `obi_otlp_export_success_total`: Successful exports
   - `obi_otlp_export_failed_total`: Failed exports
   - `obi_otlp_export_duration_seconds`: Export latency

3. **Resource Metrics**:
   - CPU usage per pod
   - Memory consumption
   - Network I/O

### Grafana Dashboard

Import the OBI dashboard:
```bash
kubectl create configmap obi-dashboard \
  --from-file=dashboards/obi-health.json \
  -n observability-dev
```

## Maintenance

### Updating OBI

1. Update the image version in `lib/obi.libsonnet`
2. Apply changes:
```bash
tk apply environments/dev/
tk apply environments/staging/
tk apply environments/production/
```

### Rolling Restart

Force a rolling restart:
```bash
kubectl rollout restart daemonset/obi -n observability-dev
kubectl rollout status daemonset/obi -n observability-dev
```

### Backup Configuration

Backup OBI configuration:
```bash
kubectl get configmap obi -n observability-dev -o yaml > obi-config-backup.yaml
```

## Integration Points

### Alloy Gateway

OBI exports all telemetry to Alloy, which then routes to:
- **Tempo**: Distributed traces
- **Mimir**: Metrics and time series
- **Loki**: Logs and events

### Service Discovery

OBI automatically discovers services through Kubernetes API:
- Pod labels
- Service endpoints
- Namespace metadata

### Correlation

OBI enriches data with:
- Trace IDs for correlation
- Kubernetes metadata
- Node and pod information
- Network context

## Best Practices

1. **Start with Development**: Test configurations in dev before production
2. **Monitor Impact**: Watch CPU/memory usage during initial deployment
3. **Gradual Rollout**: Use node selectors for phased deployment
4. **Regular Updates**: Keep OBI image updated for bug fixes and features
5. **Documentation**: Document any custom configurations
6. **Alerting**: Set up alerts for OBI health issues

## Support and References

- [eBPF Documentation](https://ebpf.io/)
- [OpenTelemetry Protocol](https://opentelemetry.io/docs/specs/otlp/)
- [Kubernetes DaemonSets](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
- [BPF Compiler Collection](https://github.com/iovisor/bcc)

## Appendix

### Complete Feature List

OBI provides instrumentation for:
- HTTP/1.1 and HTTP/2 requests
- gRPC method calls
- SQL queries (MySQL, PostgreSQL)
- Redis commands
- Kafka produce/consume operations
- DNS queries
- TCP connection lifecycle
- UDP packet flows
- System calls
- File I/O operations
- Network packet analysis

### Compatibility Matrix

| Kubernetes Version | OBI Version | Kernel Version | Status |
|-------------------|-------------|----------------|---------|
| 1.19+             | latest      | 5.8+           | ✅ Full Support |
| 1.19+             | latest      | 4.18-5.7       | ⚠️ Limited |
| <1.19             | latest      | Any            | ❌ Not Supported |