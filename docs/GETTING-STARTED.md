# Getting Started with OBI

Quick start guide to get up and running with OBI (Observability Infrastructure) in under 30 minutes.

## What is OBI?

OBI provides **zero-code observability** for your applications using eBPF technology:

- **No SDK required** - Applications run unmodified
- **No code changes** - Pure business logic
- **Automatic instrumentation** - HTTP, gRPC, SQL, Redis, Kafka
- **Distributed tracing** - Complete request flows
- **Metrics** - Request rates, latencies, error rates
- **Minimal overhead** - < 1% CPU, < 50MB memory

## Prerequisites

- Kubernetes cluster (v1.21+)
- kubectl configured
- Linux kernel 5.8+ (for eBPF support)
- Helm 3.x (optional)

**Verify Prerequisites:**
```bash
# Check Kubernetes version
kubectl version --short

# Check kernel version
kubectl get nodes -o custom-columns=NAME:.metadata.name,KERNEL:.status.nodeInfo.kernelVersion

# Check eBPF support
kubectl run test-bpf --rm -it --restart=Never --image=ubuntu -- bash -c "apt update && apt install -y linux-headers-generic && modprobe bpf"
```

## Quick Start (5 Minutes)

### 1. Deploy Observability Stack

Deploy Prometheus, Tempo, Grafana, and Loki:

```bash
# Clone repository
git clone https://github.com/raibid-labs/mop.git
cd mop

# Create namespace
kubectl create namespace observability

# Deploy stack
kubectl apply -f deployments/observability/
```

**Verify deployment:**
```bash
kubectl get pods -n observability

# Expected output:
# NAME                          READY   STATUS    RESTARTS   AGE
# prometheus-0                  1/1     Running   0          2m
# tempo-0                       1/1     Running   0          2m
# grafana-xxx                   1/1     Running   0          2m
# loki-0                        1/1     Running   0          2m
```

### 2. Deploy OBI Agent

Deploy OBI agent as a DaemonSet:

```bash
# Deploy OBI agent
kubectl apply -f deployments/obi/

# Verify deployment
kubectl get daemonset -n observability obi-agent
kubectl get pods -n observability -l app=obi-agent
```

**Check agent logs:**
```bash
kubectl logs -n observability -l app=obi-agent --tail=50

# Expected output:
# INFO: OBI agent started
# INFO: eBPF programs loaded
# INFO: Instrumentation enabled
```

### 3. Deploy Example Application

Deploy the HTTP API example:

```bash
# Deploy HTTP API example
kubectl apply -f deployments/examples/01-http-api/

# Verify deployment
kubectl get pods -l app=http-api

# Port forward
kubectl port-forward svc/http-api 8080:80
```

### 4. Generate Traffic

Send requests to generate traces:

```bash
# Get products
curl http://localhost:8080/products

# Create product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget","price":29.99,"stock":100}'

# Search products
curl http://localhost:8080/search?q=widget
```

### 5. View Dashboards

Access Grafana to view traces and metrics:

```bash
# Port forward Grafana
kubectl port-forward -n observability svc/grafana 3000:3000

# Open browser
open http://localhost:3000

# Default credentials:
# Username: admin
# Password: admin
```

**Navigate to:**
- **Explore → Tempo** - View distributed traces
- **Dashboards → HTTP API** - View application metrics

## Detailed Setup

### Step 1: Deploy Observability Backends

#### Option A: Using kubectl

```bash
# Create namespace
kubectl create namespace observability

# Deploy Prometheus
kubectl apply -f deployments/observability/prometheus/

# Deploy Tempo
kubectl apply -f deployments/observability/tempo/

# Deploy Grafana
kubectl apply -f deployments/observability/grafana/

# Deploy Loki
kubectl apply -f deployments/observability/loki/
```

#### Option B: Using Helm

```bash
# Add Helm repositories
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install Prometheus
helm install prometheus prometheus-community/kube-prometheus-stack \
  --namespace observability \
  --create-namespace

# Install Tempo
helm install tempo grafana/tempo \
  --namespace observability

# Install Loki
helm install loki grafana/loki-stack \
  --namespace observability
```

### Step 2: Configure OBI Agent

Create OBI configuration:

```yaml
# obi-config.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: obi-config
  namespace: observability
data:
  config.yaml: |
    agent:
      name: obi-agent
      log_level: info
      export_interval: 15s

    instrumentation:
      http:
        enabled: true
        capture_headers: true
        capture_body: false

      grpc:
        enabled: true
        capture_metadata: true

      sql:
        enabled: true
        capture_queries: true
        slow_query_threshold: 100ms

      redis:
        enabled: true
        capture_commands: true

      kafka:
        enabled: true
        capture_headers: true

    tracing:
      enabled: true
      sampler:
        type: probabilistic
        rate: 1.0

    exporters:
      prometheus:
        enabled: true
        port: 9090

      tempo:
        enabled: true
        endpoint: http://tempo:4317
        protocol: grpc

      loki:
        enabled: true
        endpoint: http://loki:3100
```

Apply configuration:
```bash
kubectl apply -f obi-config.yaml
```

### Step 3: Deploy OBI Agent

```yaml
# obi-daemonset.yaml
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
      serviceAccountName: obi-agent
      containers:
      - name: obi-agent
        image: obi/agent:latest
        securityContext:
          privileged: true
          capabilities:
            add:
              - SYS_ADMIN
              - NET_ADMIN
              - BPF
        env:
        - name: NODE_NAME
          valueFrom:
            fieldRef:
              fieldPath: spec.nodeName
        volumeMounts:
        - name: config
          mountPath: /etc/obi
        - name: sys
          mountPath: /sys
          readOnly: true
        - name: debugfs
          mountPath: /sys/kernel/debug
        resources:
          requests:
            cpu: 200m
            memory: 256Mi
          limits:
            cpu: 1000m
            memory: 512Mi
      volumes:
      - name: config
        configMap:
          name: obi-config
      - name: sys
        hostPath:
          path: /sys
      - name: debugfs
        hostPath:
          path: /sys/kernel/debug
---
apiVersion: v1
kind: ServiceAccount
metadata:
  name: obi-agent
  namespace: observability
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: obi-agent
rules:
- apiGroups: [""]
  resources: ["pods", "nodes", "services"]
  verbs: ["get", "list", "watch"]
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

Deploy:
```bash
kubectl apply -f obi-daemonset.yaml
```

### Step 4: Deploy Example Applications

#### HTTP API Example

```bash
cd examples/01-http-api

# Build image
make docker-build

# Deploy
kubectl apply -f ../../deployments/examples/01-http-api/

# Verify
kubectl get pods -l app=http-api
kubectl logs -l app=http-api
```

#### gRPC Service Example

```bash
cd examples/02-grpc-service

make docker-build
kubectl apply -f ../../deployments/examples/02-grpc-service/
```

#### SQL App Example

```bash
cd examples/03-sql-app

make docker-build
kubectl apply -f ../../deployments/examples/03-sql-app/
```

### Step 5: Import Grafana Dashboards

```bash
# Port forward Grafana
kubectl port-forward -n observability svc/grafana 3000:3000 &

# Wait for port forward
sleep 5

# Import dashboards
for dashboard in lib/grafana/dashboards/examples/*.json; do
  curl -X POST http://admin:admin@localhost:3000/api/dashboards/db \
    -H "Content-Type: application/json" \
    -d @"$dashboard"
done
```

Or manually import via UI:
1. Navigate to **Dashboards → Import**
2. Upload JSON files from `lib/grafana/dashboards/examples/`
3. Select Prometheus and Tempo data sources

## Verify Installation

### Check OBI Agent

```bash
# Agent status
kubectl get pods -n observability -l app=obi-agent

# Agent logs
kubectl logs -n observability -l app=obi-agent --tail=100

# Check eBPF programs
kubectl exec -n observability -it $(kubectl get pods -n observability -l app=obi-agent -o jsonpath='{.items[0].metadata.name}') -- bpftool prog list | grep obi

# Check metrics
kubectl port-forward -n observability svc/obi-agent 9090:9090 &
curl http://localhost:9090/metrics | grep obi_
```

### Check Application Instrumentation

```bash
# Generate traffic
kubectl port-forward svc/http-api 8080:80 &
for i in {1..100}; do curl http://localhost:8080/products; done

# Check traces in Tempo
kubectl port-forward -n observability svc/grafana 3000:3000 &
# Navigate to: Explore → Tempo → Service: http-api

# Check metrics in Prometheus
kubectl port-forward -n observability svc/prometheus 9090:9090 &
# Navigate to: http://localhost:9090
# Query: http_requests_total{service="http-api"}
```

### Verify Data Flow

```bash
# Check OBI agent is exporting
kubectl logs -n observability -l app=obi-agent | grep -i export

# Check Tempo is receiving traces
kubectl logs -n observability -l app=tempo | grep -i received

# Check Prometheus is scraping metrics
kubectl logs -n observability -l app=prometheus | grep -i scrape
```

## Common Issues

### Issue: OBI Agent Not Starting

**Solution:**
```bash
# Check kernel version (must be >= 5.8)
kubectl get nodes -o custom-columns=NAME:.metadata.name,KERNEL:.status.nodeInfo.kernelVersion

# Check permissions
kubectl describe pod -n observability obi-agent-xxx | grep -A10 "Security Context"

# Check logs
kubectl logs -n observability obi-agent-xxx
```

### Issue: No Traces Appearing

**Solution:**
```bash
# Check sampling rate
kubectl get configmap -n observability obi-config -o yaml | grep rate

# Check Tempo connectivity
kubectl exec -n observability -it obi-agent-xxx -- nc -zv tempo 4317

# Check application has traffic
kubectl logs -l app=http-api
```

### Issue: High CPU Usage

**Solution:**
```bash
# Reduce sampling
kubectl patch configmap/obi-config -n observability --type merge -p '{"data":{"config.yaml":"tracing:\n  sampler:\n    rate: 0.1"}}'

# Restart agents
kubectl rollout restart daemonset/obi-agent -n observability
```

## Next Steps

Now that you have OBI running:

1. **Explore Dashboards** - View the [pre-built dashboards](../lib/grafana/dashboards/examples/)
2. **Read Documentation** - Deep dive into [OBI Instrumentation](OBI-INSTRUMENTATION-GUIDE.md)
3. **Load Testing** - Follow the [Load Testing Guide](LOAD-TESTING.md)
4. **Best Practices** - Review [Best Practices](BEST-PRACTICES.md)
5. **Protocol Guides** - Learn about [protocol-specific instrumentation](examples/)

## Example Queries

### Prometheus Queries

```promql
# Request rate
sum(rate(http_requests_total[5m]))

# Latency p95
histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))

# Error rate
sum(rate(http_requests_total{status_code=~"5.."}[5m])) /
sum(rate(http_requests_total[5m])) * 100

# Top endpoints
topk(10, sum by (endpoint) (rate(http_requests_total[5m])))
```

### Tempo Queries

```
# Find traces by service
{service.name="http-api"}

# Find slow traces
{service.name="http-api" && duration > 1s}

# Find error traces
{service.name="http-api" && status.code=STATUS_CODE_ERROR}

# Find traces with specific tag
{service.name="http-api" && http.method="POST"}
```

## Resources

- **Documentation**: Full docs in [`docs/`](.)
- **Examples**: Working examples in [`examples/`](../examples/)
- **Dashboards**: Grafana dashboards in [`lib/grafana/dashboards/`](../lib/grafana/dashboards/)
- **Deployments**: Kubernetes manifests in [`deployments/`](../deployments/)

## Support

- **GitHub Issues**: https://github.com/raibid-labs/mop/issues
- **Documentation**: See [docs/](.) for detailed guides
- **Examples**: See [examples/](../examples/) for working code

## Quick Reference

```bash
# Check OBI status
kubectl get pods -n observability -l app=obi-agent

# View logs
kubectl logs -n observability -l app=obi-agent -f

# Restart OBI
kubectl rollout restart daemonset/obi-agent -n observability

# Access Grafana
kubectl port-forward -n observability svc/grafana 3000:3000

# Access Prometheus
kubectl port-forward -n observability svc/prometheus 9090:9090

# Generate test traffic
kubectl port-forward svc/http-api 8080:80 &
for i in {1..1000}; do curl http://localhost:8080/products; done
```

---

**Congratulations!** You now have OBI running with zero-code observability for your applications.
