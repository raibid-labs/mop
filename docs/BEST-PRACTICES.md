# OBI Best Practices Guide

Comprehensive best practices for deploying, configuring, and operating OBI (Observability Infrastructure) in production environments.

## Table of Contents

1. [Deployment Best Practices](#deployment-best-practices)
2. [Configuration Best Practices](#configuration-best-practices)
3. [Performance Optimization](#performance-optimization)
4. [Security Best Practices](#security-best-practices)
5. [Monitoring and Alerting](#monitoring-and-alerting)
6. [Protocol-Specific Guidance](#protocol-specific-guidance)
7. [Production Operations](#production-operations)
8. [Cost Optimization](#cost-optimization)

## Deployment Best Practices

### 1. Start with a Pilot

Begin with a small subset of applications before full rollout:

**Phase 1: Single Application (1 week)**
```bash
# Deploy OBI to staging environment
kubectl apply -f obi/staging/

# Instrument one non-critical application
kubectl label namespace/staging obi.io/instrumentation=enabled
```

**Phase 2: Namespace Rollout (2 weeks)**
```bash
# Expand to entire staging namespace
kubectl label namespace/staging obi.io/instrumentation=enabled

# Monitor and validate
kubectl top pods -n staging
```

**Phase 3: Production Rollout (Gradual)**
```bash
# Start with canary deployments
kubectl label deployment/app-v2 obi.io/instrumentation=enabled

# Monitor for issues
watch kubectl get pods -l obi.io/instrumentation=enabled
```

### 2. Use DaemonSet for Node Coverage

Deploy OBI agent as a DaemonSet to ensure all nodes are instrumented:

```yaml
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: obi-agent
  namespace: observability
spec:
  selector:
    matchLabels:
      app: obi-agent
  template:
    metadata:
      labels:
        app: obi-agent
    spec:
      hostNetwork: true
      hostPID: true
      priorityClassName: system-node-critical
      tolerations:
      - effect: NoSchedule
        key: node-role.kubernetes.io/control-plane
      - effect: NoSchedule
        key: node-role.kubernetes.io/master
      containers:
      - name: obi-agent
        image: obi/agent:v1.2.3
        securityContext:
          privileged: true
          capabilities:
            add:
              - SYS_ADMIN
              - NET_ADMIN
              - BPF
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 1000m
            memory: 512Mi
```

### 3. Implement Resource Limits

Set appropriate resource limits to prevent resource exhaustion:

```yaml
resources:
  requests:
    cpu: 200m        # Minimum guaranteed
    memory: 256Mi
  limits:
    cpu: 1000m       # Maximum allowed
    memory: 512Mi    # Hard limit
```

**Sizing Guidelines:**

| Cluster Size | CPU Request | CPU Limit | Memory Request | Memory Limit |
|--------------|-------------|-----------|----------------|--------------|
| **Small** (< 10 nodes) | 100m | 500m | 128Mi | 256Mi |
| **Medium** (10-50 nodes) | 200m | 1000m | 256Mi | 512Mi |
| **Large** (50-200 nodes) | 500m | 2000m | 512Mi | 1Gi |
| **Extra Large** (> 200 nodes) | 1000m | 4000m | 1Gi | 2Gi |

### 4. Configure Health Checks

Implement proper liveness and readiness probes:

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### 5. Use Rolling Updates

Configure update strategy for zero-downtime deployments:

```yaml
updateStrategy:
  type: RollingUpdate
  rollingUpdate:
    maxUnavailable: 1
    maxSurge: 0
```

### 6. Implement Node Affinity

Ensure OBI agents run on appropriate nodes:

```yaml
affinity:
  nodeAffinity:
    requiredDuringSchedulingIgnoredDuringExecution:
      nodeSelectorTerms:
      - matchExpressions:
        - key: node.kubernetes.io/instance-type
          operator: NotIn
          values:
          - spot    # Avoid spot instances
          - burstable
```

## Configuration Best Practices

### 1. Start Conservative

Begin with conservative settings and increase as needed:

```yaml
# Initial configuration
agent:
  export_interval: 30s  # Start with longer intervals

tracing:
  sampler:
    type: probabilistic
    rate: 0.1  # 10% sampling initially

instrumentation:
  http:
    capture_body: false  # Disable body capture
    capture_headers: true
    max_body_size: 0

  sql:
    capture_queries: true
    capture_parameters: false  # Disable parameter capture initially
```

### 2. Use Environment-Specific Configurations

Separate configurations per environment:

```
configs/
├── base/
│   └── obi-config.yaml       # Base configuration
├── staging/
│   └── obi-config.yaml       # Staging overrides
└── production/
    └── obi-config.yaml       # Production overrides
```

**Production Configuration:**
```yaml
# production/obi-config.yaml
agent:
  log_level: warn  # Less verbose in production
  export_interval: 15s

tracing:
  sampler:
    type: adaptive  # Use adaptive sampling
    rate: 0.01     # 1% default, scales up for errors

instrumentation:
  http:
    capture_body: false
  sql:
    slow_query_threshold: 100ms  # Only capture slow queries
```

**Staging Configuration:**
```yaml
# staging/obi-config.yaml
agent:
  log_level: debug  # More verbose for testing
  export_interval: 10s

tracing:
  sampler:
    type: probabilistic
    rate: 1.0  # 100% sampling in staging
```

### 3. Implement Gradual Rollout

Use feature flags for gradual configuration changes:

```yaml
instrumentation:
  http:
    enabled: true
    # Gradual rollout using annotations
    rollout_percentage: 10  # Start with 10%

  grpc:
    enabled: true
    rollout_percentage: 50  # 50% of services

  sql:
    enabled: false  # Disabled initially
```

### 4. Configure Sampling Strategies

Choose appropriate sampling for different scenarios:

**Adaptive Sampling (Recommended):**
```yaml
tracing:
  sampler:
    type: adaptive
    base_rate: 0.01      # 1% baseline
    max_rate: 1.0        # 100% for errors
    rules:
      - condition: error_rate > 0.01
        rate: 1.0          # 100% when errors occur
      - condition: latency > 1s
        rate: 1.0          # 100% for slow requests
```

**Rate Limiting Sampling:**
```yaml
tracing:
  sampler:
    type: rate_limiting
    traces_per_second: 100  # Limit total traces
```

**Parent-Based Sampling:**
```yaml
tracing:
  sampler:
    type: parent_based
    root_sampler:
      type: probabilistic
      rate: 0.1
```

### 5. Optimize Export Configuration

Configure efficient data export:

```yaml
exporters:
  tempo:
    enabled: true
    endpoint: http://tempo:4317
    protocol: grpc
    compression: gzip         # Enable compression
    batch_size: 512           # Batch size
    batch_timeout: 5s         # Batch timeout
    max_export_batch_size: 512
    max_queue_size: 2048
    retry:
      enabled: true
      max_attempts: 3
      initial_interval: 1s
      max_interval: 30s
```

## Performance Optimization

### 1. Minimize Overhead

Configure OBI to minimize performance impact:

**Disable Unnecessary Features:**
```yaml
instrumentation:
  http:
    capture_body: false       # Significant overhead reduction
    capture_headers: true
    headers_whitelist:
      - user-agent           # Only capture needed headers
      - content-type

  sql:
    capture_parameters: false # Reduce overhead
    slow_query_threshold: 100ms  # Only slow queries
```

**Optimize eBPF Configuration:**
```yaml
ebpf:
  buffer_size: 8192         # Balance memory vs. loss
  map_max_entries: 10000    # Adjust based on connection count
  events_per_cpu: 1000      # Events buffer per CPU
```

### 2. Use Connection Pooling

Optimize backend connections:

```yaml
exporters:
  tempo:
    connection_pool:
      max_idle_conns: 10
      max_idle_conns_per_host: 5
      idle_conn_timeout: 90s
```

### 3. Enable Compression

Reduce network overhead:

```yaml
exporters:
  tempo:
    compression: gzip    # or 'snappy' for speed
  prometheus:
    compression: true
```

### 4. Batch Exports

Improve efficiency with batching:

```yaml
exporters:
  batch:
    enabled: true
    size: 1000           # Traces per batch
    timeout: 10s         # Max wait time
```

### 5. Monitor OBI Performance

Track OBI agent metrics:

```promql
# OBI agent CPU usage
rate(process_cpu_seconds_total{job="obi-agent"}[5m])

# OBI agent memory usage
process_resident_memory_bytes{job="obi-agent"}

# Event processing rate
rate(obi_events_processed_total[5m])

# Export queue depth
obi_export_queue_size

# Dropped events
rate(obi_events_dropped_total[5m])
```

## Security Best Practices

### 1. Limit Captured Data

Avoid capturing sensitive information:

```yaml
instrumentation:
  http:
    capture_body: false
    headers_blacklist:
      - authorization
      - cookie
      - set-cookie
      - api-key
      - x-api-key

  sql:
    capture_parameters: false  # May contain sensitive data
    query_obfuscation: true    # Obfuscate literals
```

### 2. Implement RBAC

Use Kubernetes RBAC to restrict access:

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: obi-agent
rules:
- apiGroups: [""]
  resources:
    - pods
    - nodes
    - services
  verbs:
    - get
    - list
    - watch
- apiGroups: ["apps"]
  resources:
    - deployments
    - daemonsets
  verbs:
    - get
    - list
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: obi-agent
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: obi-agent
subjects:
- kind: ServiceAccount
  name: obi-agent
  namespace: observability
```

### 3. Use Network Policies

Restrict network access:

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: obi-agent
  namespace: observability
spec:
  podSelector:
    matchLabels:
      app: obi-agent
  policyTypes:
  - Egress
  egress:
  # Allow DNS
  - to:
    - namespaceSelector:
        matchLabels:
          name: kube-system
    ports:
    - protocol: UDP
      port: 53
  # Allow Tempo
  - to:
    - podSelector:
        matchLabels:
          app: tempo
    ports:
    - protocol: TCP
      port: 4317
  # Allow Prometheus
  - to:
    - podSelector:
        matchLabels:
          app: prometheus
    ports:
    - protocol: TCP
      port: 9090
```

### 4. Enable TLS

Use TLS for all backend connections:

```yaml
exporters:
  tempo:
    endpoint: https://tempo:4317
    tls:
      enabled: true
      cert_file: /etc/obi/certs/client.crt
      key_file: /etc/obi/certs/client.key
      ca_file: /etc/obi/certs/ca.crt
      insecure_skip_verify: false
```

### 5. Rotate Credentials

Implement credential rotation:

```yaml
exporters:
  tempo:
    auth:
      type: bearer
      token_file: /var/run/secrets/tempo-token

# Mount token from Secret
volumeMounts:
- name: tempo-token
  mountPath: /var/run/secrets/tempo-token
  readOnly: true
volumes:
- name: tempo-token
  secret:
    secretName: tempo-credentials
```

### 6. Audit Logging

Enable audit logging for compliance:

```yaml
agent:
  audit:
    enabled: true
    log_file: /var/log/obi/audit.log
    events:
      - config_change
      - agent_start
      - agent_stop
      - instrumentation_enable
      - instrumentation_disable
```

## Monitoring and Alerting

### 1. Monitor OBI Agent Health

Key metrics to monitor:

```promql
# Agent availability
up{job="obi-agent"}

# CPU usage
rate(process_cpu_seconds_total{job="obi-agent"}[5m])

# Memory usage
process_resident_memory_bytes{job="obi-agent"}

# Event processing rate
rate(obi_events_processed_total[5m])

# Export success rate
rate(obi_export_success_total[5m]) / rate(obi_export_attempts_total[5m])

# Event drops
rate(obi_events_dropped_total[5m])
```

### 2. Set Up Alerts

**Critical Alerts:**
```yaml
groups:
- name: obi-critical
  interval: 1m
  rules:
  - alert: OBIAgentDown
    expr: up{job="obi-agent"} == 0
    for: 5m
    labels:
      severity: critical
    annotations:
      summary: "OBI agent is down on {{ $labels.instance }}"

  - alert: OBIHighMemoryUsage
    expr: process_resident_memory_bytes{job="obi-agent"} > 1e9
    for: 10m
    labels:
      severity: critical
    annotations:
      summary: "OBI agent memory usage > 1GB"

  - alert: OBIHighCPUUsage
    expr: rate(process_cpu_seconds_total{job="obi-agent"}[5m]) > 1
    for: 15m
    labels:
      severity: critical
    annotations:
      summary: "OBI agent CPU usage > 100%"
```

**Warning Alerts:**
```yaml
- alert: OBIHighDropRate
  expr: rate(obi_events_dropped_total[5m]) > 100
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "OBI dropping events: {{ $value }} events/sec"

- alert: OBIExportFailures
  expr: rate(obi_export_errors_total[5m]) > 0.01
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "OBI export failures detected"

- alert: OBIHighQueueDepth
  expr: obi_export_queue_size > 1000
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "OBI export queue depth high: {{ $value }}"
```

### 3. Dashboard Monitoring

Use pre-built dashboards:

- **OBI Agent Overview**: `/lib/grafana/dashboards/obi-overview.json`
- **Protocol Dashboards**: `/lib/grafana/dashboards/examples/`
- **Multi-Protocol Overview**: `/lib/grafana/dashboards/examples/multi-protocol-overview-dashboard.json`

### 4. Log Aggregation

Collect and analyze OBI logs:

```yaml
# FluentBit configuration
[INPUT]
    Name              tail
    Path              /var/log/obi/*.log
    Tag               obi.*

[FILTER]
    Name              parser
    Match             obi.*
    Parser            json

[OUTPUT]
    Name              loki
    Match             obi.*
    Host              loki
    Port              3100
```

## Protocol-Specific Guidance

### HTTP/HTTPS

**Recommended Configuration:**
```yaml
instrumentation:
  http:
    enabled: true
    capture_headers: true
    capture_body: false        # Disable for performance
    max_body_size: 0
    headers_whitelist:
      - user-agent
      - content-type
      - accept
    slow_request_threshold: 1s
```

**Best Practices:**
- Disable body capture in production
- Whitelist only necessary headers
- Use slow request threshold to focus on problematic requests

### gRPC

**Recommended Configuration:**
```yaml
instrumentation:
  grpc:
    enabled: true
    capture_metadata: true
    capture_messages: false     # Disable for large messages
    capture_streaming: true
```

**Best Practices:**
- Enable streaming capture for debugging
- Disable message capture for large payloads
- Use metadata for trace context

### SQL Databases

**Recommended Configuration:**
```yaml
instrumentation:
  sql:
    enabled: true
    capture_queries: true
    capture_parameters: false   # Avoid PII
    query_obfuscation: true     # Obfuscate literals
    slow_query_threshold: 100ms
    max_query_length: 1024
```

**Best Practices:**
- Enable query obfuscation
- Disable parameter capture (may contain PII)
- Focus on slow queries only
- Set appropriate query length limit

### Redis

**Recommended Configuration:**
```yaml
instrumentation:
  redis:
    enabled: true
    capture_commands: true
    capture_arguments: true
    sampling_rate: 0.1  # Sample 10% for high-volume
```

**Best Practices:**
- Use sampling for high-volume Redis workloads
- Monitor cache hit/miss rates
- Track slow commands

### Kafka

**Recommended Configuration:**
```yaml
instrumentation:
  kafka:
    enabled: true
    capture_headers: true
    capture_key: true
    capture_value_size: true
    max_message_size: 10240
    lag_monitoring: true
    consumer_groups:
      - my-consumer-group
```

**Best Practices:**
- Enable lag monitoring
- Capture message metadata, not full payloads
- Monitor rebalance events
- Track partition distribution

## Production Operations

### 1. Change Management

Follow a structured change process:

**Change Workflow:**
1. **Test in staging** - Validate all changes
2. **Create rollback plan** - Document rollback steps
3. **Gradual rollout** - Deploy to canary first
4. **Monitor metrics** - Watch for anomalies
5. **Validate** - Confirm expected behavior
6. **Document** - Update runbooks

**Example Rollout:**
```bash
# 1. Update staging
kubectl apply -f obi-config-staging.yaml -n observability-staging

# 2. Validate
kubectl logs -n observability-staging -l app=obi-agent --tail=100

# 3. Deploy to canary (10% of production)
kubectl apply -f obi-config-canary.yaml -n observability

# 4. Monitor for 1 hour
watch kubectl top pods -n observability

# 5. Full production rollout
kubectl apply -f obi-config-production.yaml -n observability
```

### 2. Backup and Recovery

Backup critical configurations:

```bash
# Backup script
#!/bin/bash
DATE=$(date +%Y%m%d-%H%M%S)
BACKUP_DIR="./backups/$DATE"

mkdir -p "$BACKUP_DIR"

# Backup ConfigMaps
kubectl get configmap -n observability obi-config -o yaml > "$BACKUP_DIR/obi-config.yaml"

# Backup Secrets
kubectl get secret -n observability obi-credentials -o yaml > "$BACKUP_DIR/obi-credentials.yaml"

# Backup DaemonSet
kubectl get daemonset -n observability obi-agent -o yaml > "$BACKUP_DIR/obi-daemonset.yaml"

echo "Backup saved to $BACKUP_DIR"
```

### 3. Disaster Recovery

Prepare for failure scenarios:

**Scenario 1: OBI Agent Failure**
```bash
# Verify applications continue working
kubectl get pods --all-namespaces

# Restart OBI agents
kubectl rollout restart daemonset/obi-agent -n observability

# Verify recovery
kubectl get pods -n observability -l app=obi-agent
```

**Scenario 2: Complete Observability Stack Failure**
```bash
# Applications continue running (zero-code)
# Observability data is lost during outage

# Restore from backup
kubectl apply -f backups/latest/
```

### 4. Version Upgrades

Safe upgrade procedure:

```bash
# 1. Check release notes
curl https://github.com/obi/obi/releases/latest

# 2. Backup current configuration
./backup-obi-config.sh

# 3. Update staging
kubectl set image daemonset/obi-agent -n observability-staging \
  obi-agent=obi/agent:v1.3.0

# 4. Validate staging
kubectl rollout status daemonset/obi-agent -n observability-staging

# 5. Update production gradually
kubectl set image daemonset/obi-agent -n observability \
  obi-agent=obi/agent:v1.3.0

# 6. Monitor rollout
kubectl rollout status daemonset/obi-agent -n observability

# 7. Rollback if needed
kubectl rollout undo daemonset/obi-agent -n observability
```

## Cost Optimization

### 1. Optimize Sampling

Reduce costs with smart sampling:

```yaml
# Cost-effective sampling
tracing:
  sampler:
    type: adaptive
    base_rate: 0.01      # 1% baseline (99% cost reduction)
    max_rate: 1.0        # 100% for errors
    rules:
      - condition: status_code >= 500
        rate: 1.0          # Always trace errors
      - condition: duration > 1s
        rate: 1.0          # Always trace slow requests
      - condition: is_important_endpoint()
        rate: 0.1          # 10% for important endpoints
```

### 2. Optimize Retention

Configure appropriate retention periods:

```yaml
# Tempo configuration
storage:
  trace:
    backend: s3
    s3:
      bucket: traces
    blocklist_poll: 5m
    retention: 168h  # 7 days (adjust based on needs)

# Prometheus configuration
storage:
  tsdb:
    retention.time: 15d    # 15 days for metrics
    retention.size: 50GB   # Or size-based
```

### 3. Use Tiered Storage

Implement tiered storage strategy:

**Hot Storage (Fast, Expensive):**
- Recent data (last 7 days)
- SSD storage
- Frequently accessed

**Warm Storage (Moderate):**
- 7-30 days old
- Standard storage
- Occasional access

**Cold Storage (Slow, Cheap):**
- Older than 30 days
- Object storage (S3, GCS)
- Rarely accessed

```yaml
storage:
  trace:
    backend: s3
    local_cache:
      enabled: true
      size: 10GB           # Hot cache
    s3:
      bucket: traces
      storage_class: INTELLIGENT_TIERING
```

### 4. Optimize Resource Usage

Right-size OBI agent resources:

```bash
# Analyze actual usage
kubectl top pods -n observability -l app=obi-agent

# Adjust based on 80% utilization target
# CPU: actual_usage / 0.8
# Memory: actual_usage / 0.8
```

### 5. Compress Exports

Enable compression to reduce network costs:

```yaml
exporters:
  tempo:
    compression: gzip  # Reduces size by 70-90%
  prometheus:
    compression: true
```

## Summary Checklist

### Deployment
- [ ] Start with pilot application
- [ ] Use DaemonSet deployment
- [ ] Configure resource limits
- [ ] Implement health checks
- [ ] Use rolling updates

### Configuration
- [ ] Start with conservative settings
- [ ] Use environment-specific configs
- [ ] Implement adaptive sampling
- [ ] Optimize export configuration
- [ ] Enable compression

### Performance
- [ ] Disable body capture
- [ ] Use connection pooling
- [ ] Enable batching
- [ ] Monitor OBI metrics
- [ ] Measure overhead

### Security
- [ ] Limit captured data
- [ ] Implement RBAC
- [ ] Use network policies
- [ ] Enable TLS
- [ ] Rotate credentials

### Operations
- [ ] Set up monitoring
- [ ] Configure alerts
- [ ] Implement backup strategy
- [ ] Document procedures
- [ ] Test disaster recovery

## Related Documentation

- [OBI Instrumentation Guide](OBI-INSTRUMENTATION-GUIDE.md)
- [Troubleshooting Guide](TROUBLESHOOTING.md)
- [Load Testing Guide](LOAD-TESTING.md)
- [Example Applications](../examples/)

## Support

For questions or assistance:
- **Documentation**: https://docs.obi.io
- **GitHub**: https://github.com/obi/obi/issues
- **Slack**: https://obi-community.slack.com
- **Email**: support@obi.io
