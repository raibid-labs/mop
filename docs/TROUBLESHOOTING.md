# OBI Troubleshooting Guide

Comprehensive troubleshooting guide for OBI (Observability Infrastructure) across all supported protocols and deployment scenarios.

## Table of Contents

1. [Quick Diagnostics](#quick-diagnostics)
2. [Common Issues](#common-issues)
3. [Protocol-Specific Issues](#protocol-specific-issues)
4. [Performance Issues](#performance-issues)
5. [Data Pipeline Issues](#data-pipeline-issues)
6. [Kubernetes Issues](#kubernetes-issues)
7. [Debug Tools](#debug-tools)
8. [Support Resources](#support-resources)

## Quick Diagnostics

### Health Check Script

Run this script first to identify common issues:

```bash
#!/bin/bash
# obi-health-check.sh

echo "=== OBI Health Check ==="
echo ""

# 1. Check OBI agent pods
echo "1. OBI Agent Status:"
kubectl get pods -n observability -l app=obi-agent
echo ""

# 2. Check agent logs for errors
echo "2. Recent Errors:"
kubectl logs -n observability -l app=obi-agent --tail=50 | grep -i error
echo ""

# 3. Check eBPF programs loaded
echo "3. eBPF Programs:"
kubectl exec -n observability -it $(kubectl get pods -n observability -l app=obi-agent -o jsonpath='{.items[0].metadata.name}') -- bpftool prog list | grep obi
echo ""

# 4. Check metrics endpoint
echo "4. Metrics Endpoint:"
kubectl port-forward -n observability svc/obi-agent 9090:9090 &
PF_PID=$!
sleep 2
curl -s http://localhost:9090/metrics | head -20
kill $PF_PID
echo ""

# 5. Check trace export
echo "5. Trace Export Status:"
kubectl logs -n observability -l app=obi-agent --tail=100 | grep -i tempo
echo ""

# 6. Check instrumented services
echo "6. Instrumented Services:"
kubectl get pods --all-namespaces -l obi.io/instrumentation=enabled
echo ""

echo "=== Health Check Complete ==="
```

### Quick Command Reference

```bash
# View OBI agent status
kubectl get pods -n observability -l app=obi-agent

# Check agent logs
kubectl logs -n observability -l app=obi-agent -f

# Describe agent pod
kubectl describe pod -n observability -l app=obi-agent

# Check agent metrics
kubectl port-forward -n observability svc/obi-agent 9090:9090
curl http://localhost:9090/metrics

# View eBPF programs
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog list

# View eBPF maps
kubectl exec -n observability -it obi-agent-xxx -- bpftool map list

# Check trace data
kubectl exec -n observability -it obi-agent-xxx -- bpftool map dump name traces

# Restart OBI agent
kubectl rollout restart daemonset/obi-agent -n observability
```

## Common Issues

### Issue 1: No Traces Appearing

**Symptoms:**
- Traces not showing up in Grafana/Tempo
- Metrics visible but traces missing
- Empty trace queries

**Diagnosis:**

```bash
# 1. Check OBI agent is running
kubectl get pods -n observability -l app=obi-agent

# 2. Check agent logs for export errors
kubectl logs -n observability -l app=obi-agent | grep -i "tempo\|trace\|export"

# 3. Verify Tempo connectivity
kubectl exec -n observability -it obi-agent-xxx -- nc -zv tempo 4317

# 4. Check trace sampling rate
kubectl get configmap -n observability obi-config -o yaml | grep -A5 tracing

# 5. Verify application has traffic
kubectl logs -l app=http-api --tail=100
```

**Solutions:**

1. **Enable tracing in OBI config:**
```yaml
tracing:
  enabled: true
  sampler:
    type: probabilistic
    rate: 1.0  # 100% sampling
```

2. **Check Tempo endpoint:**
```yaml
exporters:
  tempo:
    enabled: true
    endpoint: http://tempo:4317
    protocol: grpc
```

3. **Verify network policies:**
```bash
# Allow OBI agent to reach Tempo
kubectl get networkpolicies -n observability
```

4. **Check Tempo is receiving data:**
```bash
kubectl logs -n observability -l app=tempo | grep -i received
```

### Issue 2: High CPU Usage

**Symptoms:**
- OBI agent using >1 CPU core
- Node CPU saturation
- Application performance degradation

**Diagnosis:**

```bash
# 1. Check CPU usage
kubectl top pods -n observability -l app=obi-agent

# 2. Check traffic volume
kubectl exec -n observability -it obi-agent-xxx -- cat /proc/net/snmp | grep Tcp:

# 3. Check eBPF overhead
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog show

# 4. Profile OBI agent
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:6060/debug/pprof/profile?seconds=30 > cpu.prof
```

**Solutions:**

1. **Reduce sampling rate:**
```yaml
tracing:
  sampler:
    type: probabilistic
    rate: 0.1  # 10% sampling
```

2. **Disable body capture:**
```yaml
instrumentation:
  http:
    capture_body: false
  sql:
    capture_queries: false
```

3. **Increase export interval:**
```yaml
agent:
  export_interval: 30s  # Default: 15s
```

4. **Adjust resource limits:**
```yaml
resources:
  limits:
    cpu: "1"
    memory: 512Mi
  requests:
    cpu: "200m"
    memory: 128Mi
```

### Issue 3: Memory Leak

**Symptoms:**
- OBI agent memory increasing over time
- Out of memory (OOM) kills
- Pod restarts

**Diagnosis:**

```bash
# 1. Check memory usage over time
kubectl top pods -n observability -l app=obi-agent --sort-by=memory

# 2. Check for OOM kills
kubectl describe pod -n observability obi-agent-xxx | grep -A10 "Last State"

# 3. Check eBPF map sizes
kubectl exec -n observability -it obi-agent-xxx -- bpftool map list

# 4. Memory profile
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:6060/debug/pprof/heap > heap.prof
```

**Solutions:**

1. **Reduce eBPF map sizes:**
```yaml
ebpf:
  map_max_entries: 10000  # Default: 100000
```

2. **Limit buffer sizes:**
```yaml
ebpf:
  buffer_size: 8192  # Ring buffer per CPU
```

3. **Increase memory limits:**
```yaml
resources:
  limits:
    memory: 1Gi
  requests:
    memory: 256Mi
```

4. **Enable memory profiling:**
```yaml
agent:
  profiling:
    enabled: true
    port: 6060
```

### Issue 4: Missing Metrics

**Symptoms:**
- Metrics not appearing in Prometheus
- Empty Grafana dashboards
- Scrape failures

**Diagnosis:**

```bash
# 1. Check metrics endpoint
kubectl port-forward -n observability svc/obi-agent 9090:9090
curl http://localhost:9090/metrics

# 2. Check Prometheus scrape config
kubectl get configmap -n observability prometheus-config -o yaml

# 3. Check Prometheus targets
kubectl port-forward -n observability svc/prometheus 9090:9090
# Navigate to: http://localhost:9090/targets

# 4. Check agent logs for export errors
kubectl logs -n observability -l app=obi-agent | grep -i "prometheus\|metrics"
```

**Solutions:**

1. **Enable metrics export:**
```yaml
exporters:
  prometheus:
    enabled: true
    port: 9090
    path: /metrics
```

2. **Add Prometheus scrape annotation:**
```yaml
metadata:
  annotations:
    prometheus.io/scrape: "true"
    prometheus.io/port: "9090"
    prometheus.io/path: "/metrics"
```

3. **Check ServiceMonitor:**
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: obi-agent
spec:
  selector:
    matchLabels:
      app: obi-agent
  endpoints:
  - port: metrics
    interval: 15s
```

### Issue 5: Incorrect Service Names

**Symptoms:**
- Services showing as "unknown"
- Traces not grouped by service
- Missing service topology

**Diagnosis:**

```bash
# 1. Check Kubernetes metadata
kubectl get pods -l app=http-api -o yaml | grep -A10 metadata

# 2. Check OBI agent service discovery
kubectl logs -n observability -l app=obi-agent | grep -i "service discovery"

# 3. Check trace metadata
kubectl exec -n observability -it obi-agent-xxx -- bpftool map dump name service_map
```

**Solutions:**

1. **Add service label:**
```yaml
metadata:
  labels:
    app: http-api
    obi.io/service.name: "http-api"
```

2. **Configure service name annotation:**
```yaml
metadata:
  annotations:
    obi.io/service.name: "product-catalog-api"
```

3. **Enable Kubernetes integration:**
```yaml
kubernetes:
  enabled: true
  use_pod_labels: true
```

## Protocol-Specific Issues

### HTTP Issues

#### Issue: HTTP/2 Requests Not Captured

**Diagnosis:**
```bash
# Check HTTP version
kubectl logs -l app=http-api | grep "HTTP/2"

# Check OBI HTTP/2 support
kubectl logs -n observability -l app=obi-agent | grep -i "http2"
```

**Solution:**
```yaml
instrumentation:
  http:
    enabled: true
    http2_enabled: true
```

#### Issue: Large Response Bodies Truncated

**Diagnosis:**
```bash
# Check max body size setting
kubectl get configmap -n observability obi-config -o yaml | grep max_body_size
```

**Solution:**
```yaml
instrumentation:
  http:
    max_body_size: 10240  # 10KB
```

### gRPC Issues

#### Issue: gRPC Streaming Not Traced

**Diagnosis:**
```bash
# Check for streaming calls
kubectl logs -l app=grpc-service | grep -i stream

# Check OBI streaming support
kubectl logs -n observability -l app=obi-agent | grep -i "grpc.*stream"
```

**Solution:**
```yaml
instrumentation:
  grpc:
    enabled: true
    capture_streaming: true
```

#### Issue: Missing gRPC Metadata

**Diagnosis:**
```bash
# Check metadata capture
kubectl get configmap -n observability obi-config -o yaml | grep capture_metadata
```

**Solution:**
```yaml
instrumentation:
  grpc:
    capture_metadata: true
    metadata_whitelist:
      - user-agent
      - authorization
```

### SQL Issues

#### Issue: Queries Not Captured

**Diagnosis:**
```bash
# Check SQL instrumentation
kubectl logs -n observability -l app=obi-agent | grep -i "sql"

# Verify database connections
kubectl exec -l app=sql-app -- netstat -an | grep 5432
```

**Solution:**
```yaml
instrumentation:
  sql:
    enabled: true
    capture_queries: true
    databases:
      - postgres
      - mysql
```

#### Issue: Query Parameters Missing

**Diagnosis:**
```bash
# Check parameter capture
kubectl get configmap -n observability obi-config -o yaml | grep capture_parameters
```

**Solution:**
```yaml
instrumentation:
  sql:
    capture_parameters: true
    max_param_length: 100
```

### Redis Issues

#### Issue: High Volume Missing Commands

**Diagnosis:**
```bash
# Check Redis command rate
kubectl exec -l app=redis-cache -- redis-cli INFO stats | grep total_commands

# Check OBI buffer overflow
kubectl logs -n observability -l app=obi-agent | grep -i "buffer.*full"
```

**Solution:**
```yaml
ebpf:
  buffer_size: 16384  # Increase buffer size
instrumentation:
  redis:
    sampling_rate: 0.1  # Sample 10% of commands
```

### Kafka Issues

#### Issue: Consumer Lag Not Reported

**Diagnosis:**
```bash
# Check Kafka metrics
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:9090/metrics | grep kafka_consumer_lag

# Check consumer group
kubectl exec -l app=kafka-broker -- kafka-consumer-groups.sh --bootstrap-server localhost:9092 --describe --group my-group
```

**Solution:**
```yaml
instrumentation:
  kafka:
    enabled: true
    consumer_groups:
      - my-group
    lag_monitoring: true
```

## Performance Issues

### Issue: High Latency

**Symptoms:**
- Application latency increased after OBI deployment
- p99 latency spike
- Timeout errors

**Diagnosis:**

```bash
# 1. Compare latency before/after OBI
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:9090/metrics | grep http_request_duration_seconds

# 2. Check eBPF overhead
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog show | grep run_time

# 3. Profile application
kubectl exec -l app=http-api -- curl http://localhost:6060/debug/pprof/profile?seconds=30 > app.prof
```

**Solutions:**

1. **Disable body capture:**
```yaml
instrumentation:
  http:
    capture_body: false
```

2. **Reduce sampling:**
```yaml
tracing:
  sampler:
    rate: 0.1
```

3. **Optimize eBPF programs:**
```yaml
ebpf:
  optimization_level: 3
```

### Issue: Low Throughput

**Symptoms:**
- Reduced requests per second
- Increased queue times
- Connection timeouts

**Diagnosis:**

```bash
# 1. Check throughput metrics
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:9090/metrics | grep http_requests_total

# 2. Check connection states
kubectl exec -l app=http-api -- ss -s

# 3. Check for connection limits
kubectl exec -n observability -it obi-agent-xxx -- bpftool map dump name connection_map | wc -l
```

**Solutions:**

1. **Increase connection limits:**
```yaml
ebpf:
  map_max_entries: 50000
```

2. **Disable unnecessary capture:**
```yaml
instrumentation:
  http:
    capture_headers: false
```

3. **Use async export:**
```yaml
exporters:
  async_export: true
  batch_size: 1000
```

## Data Pipeline Issues

### Issue: Data Loss

**Symptoms:**
- Missing traces in time series
- Gaps in metrics
- Incomplete traces

**Diagnosis:**

```bash
# 1. Check buffer overflows
kubectl logs -n observability -l app=obi-agent | grep -i "overflow\|dropped\|lost"

# 2. Check export queue
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:9090/metrics | grep export_queue

# 3. Check backend capacity
kubectl logs -n observability -l app=tempo | grep -i "rejected\|throttled"
```

**Solutions:**

1. **Increase buffer sizes:**
```yaml
ebpf:
  buffer_size: 16384
  map_max_entries: 50000
```

2. **Increase export workers:**
```yaml
exporters:
  workers: 4
  queue_size: 10000
```

3. **Scale backends:**
```bash
kubectl scale deployment/tempo -n observability --replicas=3
```

### Issue: Export Failures

**Symptoms:**
- Data not reaching backends
- Export errors in logs
- Queue buildup

**Diagnosis:**

```bash
# 1. Check export errors
kubectl logs -n observability -l app=obi-agent | grep -i "export.*error\|failed to export"

# 2. Check backend connectivity
kubectl exec -n observability -it obi-agent-xxx -- nc -zv tempo 4317

# 3. Check authentication
kubectl get secret -n observability obi-credentials
```

**Solutions:**

1. **Fix endpoint configuration:**
```yaml
exporters:
  tempo:
    endpoint: http://tempo.observability.svc.cluster.local:4317
```

2. **Add authentication:**
```yaml
exporters:
  tempo:
    auth:
      type: bearer
      token_file: /var/run/secrets/tempo-token
```

3. **Retry configuration:**
```yaml
exporters:
  retry:
    enabled: true
    max_attempts: 3
    backoff: exponential
```

## Kubernetes Issues

### Issue: DaemonSet Not Starting

**Symptoms:**
- OBI agent pods not running
- CrashLoopBackOff
- ImagePullBackOff

**Diagnosis:**

```bash
# 1. Check pod status
kubectl get pods -n observability -l app=obi-agent

# 2. Describe pod
kubectl describe pod -n observability obi-agent-xxx

# 3. Check events
kubectl get events -n observability --sort-by='.lastTimestamp'

# 4. Check logs
kubectl logs -n observability obi-agent-xxx
```

**Solutions:**

1. **Fix RBAC permissions:**
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: obi-agent
rules:
- apiGroups: [""]
  resources: ["pods", "nodes"]
  verbs: ["get", "list", "watch"]
```

2. **Fix security context:**
```yaml
securityContext:
  privileged: true
  capabilities:
    add:
      - SYS_ADMIN
      - NET_ADMIN
      - BPF
```

3. **Check node compatibility:**
```bash
# Verify kernel version (>=5.8 required)
kubectl get nodes -o custom-columns=NAME:.metadata.name,KERNEL:.status.nodeInfo.kernelVersion
```

### Issue: eBPF Programs Not Loading

**Symptoms:**
- "Failed to load eBPF program" errors
- No instrumentation working
- Permission denied errors

**Diagnosis:**

```bash
# 1. Check kernel config
kubectl exec -n observability -it obi-agent-xxx -- zgrep CONFIG_BPF /proc/config.gz

# 2. Check BPF filesystem
kubectl exec -n observability -it obi-agent-xxx -- mount | grep bpf

# 3. Check capabilities
kubectl exec -n observability -it obi-agent-xxx -- capsh --print
```

**Solutions:**

1. **Enable kernel features:**
```bash
# On node (may require reboot)
sudo modprobe bpf
sudo mount -t bpf bpf /sys/fs/bpf
```

2. **Update security policy:**
```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: obi-agent
spec:
  privileged: true
  allowedCapabilities:
    - SYS_ADMIN
    - NET_ADMIN
    - BPF
```

## Debug Tools

### OBI CLI

```bash
# Install OBI CLI
kubectl exec -n observability -it obi-agent-xxx -- /opt/obi/bin/obi-cli

# View active traces
obi-cli traces list

# View metrics
obi-cli metrics query 'http_requests_total'

# View eBPF stats
obi-cli ebpf stats

# View service map
obi-cli services map
```

### bpftool

```bash
# List loaded programs
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog list

# Show program details
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog show id 123

# Dump eBPF map
kubectl exec -n observability -it obi-agent-xxx -- bpftool map dump name traces

# View program stats
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog show id 123 --json | jq .run_time_ns
```

### tcpdump

```bash
# Capture HTTP traffic
kubectl exec -n observability -it obi-agent-xxx -- tcpdump -i any -n -A 'tcp port 80'

# Capture gRPC traffic
kubectl exec -n observability -it obi-agent-xxx -- tcpdump -i any -n -X 'tcp port 50051'

# Save to file
kubectl exec -n observability -it obi-agent-xxx -- tcpdump -i any -w /tmp/capture.pcap
```

### Prometheus Queries

```promql
# Check instrumentation coverage
count by (service) (up{job="obi-instrumented"})

# Check error rates
sum by (service) (rate(http_requests_total{status_code=~"5.."}[5m]))

# Check latency distribution
histogram_quantile(0.95, sum by (le, service) (rate(http_request_duration_seconds_bucket[5m])))

# Check OBI agent health
up{job="obi-agent"}
rate(obi_events_processed_total[5m])
obi_export_errors_total
```

## Support Resources

### Documentation

- [OBI Instrumentation Guide](OBI-INSTRUMENTATION-GUIDE.md)
- [Load Testing Guide](LOAD-TESTING.md)
- [Best Practices](BEST-PRACTICES.md)
- [Protocol Guides](examples/)

### Community

- **GitHub Issues**: https://github.com/obi/obi/issues
- **Slack Community**: https://obi-community.slack.com
- **Stack Overflow**: Tag `obi-observability`

### Commercial Support

- **Email**: support@obi.io
- **Support Portal**: https://support.obi.io
- **Emergency Hotline**: Available for Enterprise customers

### Reporting Bugs

When reporting issues, please include:

1. **Environment:**
   - Kubernetes version
   - Kernel version
   - OBI agent version
   - Application details

2. **Logs:**
```bash
kubectl logs -n observability -l app=obi-agent --tail=500 > obi-agent.log
kubectl describe pod -n observability obi-agent-xxx > obi-agent-describe.txt
```

3. **Configuration:**
```bash
kubectl get configmap -n observability obi-config -o yaml > obi-config.yaml
```

4. **Metrics:**
```bash
kubectl exec -n observability -it obi-agent-xxx -- curl http://localhost:9090/metrics > obi-metrics.txt
```

5. **Reproduction steps**

6. **Expected vs actual behavior**

---

## Quick Reference Card

### Common Commands

```bash
# Health check
kubectl get pods -n observability -l app=obi-agent

# View logs
kubectl logs -n observability -l app=obi-agent -f

# Restart agent
kubectl rollout restart daemonset/obi-agent -n observability

# Check metrics
kubectl port-forward -n observability svc/obi-agent 9090:9090

# View traces
kubectl port-forward -n observability svc/grafana 3000:3000

# Debug eBPF
kubectl exec -n observability -it obi-agent-xxx -- bpftool prog list
```

### Emergency Procedures

**Complete system failure:**
```bash
# 1. Disable OBI
kubectl scale daemonset/obi-agent -n observability --replicas=0

# 2. Verify applications recover
kubectl get pods --all-namespaces

# 3. Check logs
kubectl logs -n observability -l app=obi-agent --tail=1000 > emergency.log

# 4. Contact support with logs
```

**High load scenario:**
```bash
# 1. Reduce sampling
kubectl patch configmap/obi-config -n observability --type merge -p '{"data":{"config.yaml":"tracing:\n  sampler:\n    rate: 0.01"}}'

# 2. Restart agents
kubectl rollout restart daemonset/obi-agent -n observability

# 3. Monitor impact
kubectl top pods -n observability
```
