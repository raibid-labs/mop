# OpenTelemetry Backend Initiative (OBI) - Comprehensive Research Report

**Research Date:** November 6, 2025
**Status:** Alpha Release (November 3, 2025)

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [What is OBI?](#what-is-obi)
3. [First Release Details](#first-release-details)
4. [Architecture & Technical Design](#architecture--technical-design)
5. [Integration with Grafana Stack](#integration-with-grafana-stack)
6. [Deployment Strategies](#deployment-strategies)
7. [Key Use Cases](#key-use-cases)
8. [Experimental Implementations](#experimental-implementations)
9. [Helm Charts & Kubernetes](#helm-charts--kubernetes)
10. [Comparison with Traditional Backends](#comparison-with-traditional-backends)
11. [Recommendations](#recommendations)

---

## Executive Summary

**OpenTelemetry eBPF Instrumentation (OBI)** represents a paradigm shift in observability, providing zero-code instrumentation through kernel-level eBPF technology. Released as alpha on November 3, 2025, OBI originated from Grafana Beyla and was donated to the OpenTelemetry project to accelerate community-driven development.

**Key Highlights:**
- **Zero Application Impact**: No code changes, restarts, or configuration modifications required
- **Minimal Overhead**: Less than 1% CPU usage, far lower than traditional SDKs
- **Language Agnostic**: Works with Java, .NET, Go, Python, Ruby, Node.js, and more
- **Protocol Coverage**: HTTP/S, HTTP/2, gRPC, SQL, Redis, MongoDB, Kafka, GraphQL, S3
- **Production Ready**: Donated by Grafana Labs with 19,000+ lines of production-tested code

---

## What is OBI?

### Definition

OpenTelemetry eBPF Instrumentation (OBI) is an **out-of-process auto-instrumentation tool** that uses eBPF (extended Berkeley Packet Filter) to capture telemetry at the kernel level, providing metrics and distributed traces without modifying application code.

### Core Concept

Unlike traditional OpenTelemetry instrumentation that operates at the library level within the application process, OBI operates at the **protocol level** in the kernel. This fundamental difference enables:

1. **Zero-Code Instrumentation**: Fully automatic capture without any application changes
2. **Universal Language Support**: Works with any language that makes system calls
3. **Consistent Telemetry**: Same telemetry format across all languages and frameworks
4. **Minimal Performance Impact**: Kernel-level efficiency with < 1% CPU overhead

### Origins

- **Developed By**: Grafana Labs (originally named "Beyla")
- **Donated To**: OpenTelemetry Project (May 2025)
- **First Release**: November 3, 2025 (Alpha)
- **Primary Contributors**: Nikola Grcevski (Grafana Labs), Tyler Yahn (Splunk), Coralogix (19K+ LOC)

---

## First Release Details

### Release Information

- **Date**: November 3, 2025
- **Status**: Alpha Release
- **Announcement**: [OpenTelemetry Blog](https://opentelemetry.io/blog/2025/obi-announcing-first-release/)
- **Authors**: Nikola Grcevski (Grafana Labs), Tyler Yahn (Splunk)

### Key Features in v1.0 Alpha

#### Supported Protocols
- **Web**: HTTP/HTTPS, HTTP/2, gRPC
- **Databases**: SQL (PostgreSQL, MySQL), Redis, MongoDB
- **Message Queues**: Kafka
- **APIs**: GraphQL, REST
- **Cloud Services**: AWS S3, Elasticsearch/OpenSearch

#### Telemetry Capabilities
- **Metrics**: RED (Request rate, Error rate, Duration) metrics
- **Traces**: Distributed tracing with automatic context propagation
- **Export**: OTLP (OpenTelemetry Protocol) to any OTLP-compatible backend

#### Performance Characteristics
- **CPU Overhead**: < 1% in production workloads
- **Memory Footprint**: Minimal compared to in-process SDKs (10-50x reduction)
- **Latency Impact**: Zero application latency impact (kernel-level operation)

### Current Limitations

The alpha release has documented constraints:

1. **Reactive Programming**: Limited support for reactive frameworks (Project Reactor, RxJava)
2. **Java Virtual Threads**: Not yet supported (preview feature in Java 21+)
3. **Complex Thread Pools**: Limited support for advanced threading patterns
4. **Generic Telemetry**: Provides protocol-level data, not application-specific custom attributes

**Recommendation**: Use OBI alongside traditional instrumentation for complementary coverage.

---

## Architecture & Technical Design

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      User Space                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           OBI Agent (User Space)                      │   │
│  │  - Reads eBPF maps                                    │   │
│  │  - Processes & enriches telemetry                     │   │
│  │  - Adds K8s metadata (pod, namespace, deployment)    │   │
│  │  - Exports via OTLP                                   │   │
│  └──────────────────────────────────────────────────────┘   │
│                           ▲                                  │
│                           │ eBPF Maps                        │
└───────────────────────────┼──────────────────────────────────┘
                            │
┌───────────────────────────┼──────────────────────────────────┐
│                    Kernel Space                              │
│  ┌──────────────────────────────────────────────────────┐   │
│  │           eBPF Probes (Kernel)                        │   │
│  │  - kprobes: kernel function tracing                   │   │
│  │  - uprobes: user-space function tracing               │   │
│  │  - tracepoints: static kernel instrumentation         │   │
│  │  - Network hooks: socket operations                   │   │
│  │  - Captures: syscalls, network packets, I/O           │   │
│  └──────────────────────────────────────────────────────┘   │
│                           ▲                                  │
└───────────────────────────┼──────────────────────────────────┘
                            │
                     Application Processes
```

### Components

#### 1. Kernel-Space eBPF Probes

**Function**: Capture raw network, process, and application telemetry data

**Types of Probes**:
- **kprobes**: Dynamic kernel function tracing
- **uprobes**: User-space application function tracing
- **tracepoints**: Static kernel instrumentation points
- **socket hooks**: Network-level packet inspection

**Data Captured**:
- HTTP request/response metadata (method, path, status, headers)
- SQL queries (sanitized for security)
- gRPC calls and responses
- Network socket operations (connect, send, recv)
- Process context (PID, TID, namespace)

**Storage**: Captured data stored in highly efficient **eBPF maps** (kernel memory)

#### 2. User-Space OBI Agent

**Function**: Process and export telemetry from eBPF maps

**Responsibilities**:
1. **Read eBPF Maps**: Continuously poll kernel maps for telemetry data
2. **Process Data**: Parse protocol-specific data (HTTP, gRPC, SQL, etc.)
3. **Enrich Metadata**: Add Kubernetes context (pod, namespace, labels, annotations)
4. **Build Traces**: Construct distributed trace spans with context propagation
5. **Export Telemetry**: Send metrics and traces via OTLP to backends

**Context Propagation Methods**:
- **Network-level**: Extract trace context from HTTP headers (traceparent, b3)
- **Memory-level**: Share trace context between processes via shared memory

### How It Works: HTTP Request Example

```
1. Application makes HTTP request
   ↓
2. eBPF probe attached to socket send() syscall captures:
   - Request headers (including traceparent)
   - HTTP method, path
   - Timestamp (start)
   ↓
3. Data stored in eBPF map (kernel memory)
   ↓
4. eBPF probe on socket recv() captures response:
   - Status code
   - Response headers
   - Timestamp (end)
   ↓
5. User-space agent reads eBPF map:
   - Calculates duration (end - start)
   - Extracts trace context
   - Enriches with K8s metadata
   ↓
6. Agent creates OTLP span and exports to backend
```

### Comparison: Traditional vs eBPF Instrumentation

| Aspect | Traditional SDK | OBI (eBPF) |
|--------|----------------|------------|
| **Installation** | Add SDK to application | Deploy agent (DaemonSet) |
| **Code Changes** | Required (import, configure) | None |
| **Restart Required** | Yes | No |
| **Language Support** | Per-language SDKs | Universal (kernel-level) |
| **Performance Impact** | 5-15% CPU, high memory | < 1% CPU, minimal memory |
| **Custom Attributes** | Full support | Limited (protocol-level only) |
| **Protocol Coverage** | Library-dependent | Protocol-level (HTTP, gRPC, SQL) |
| **Maintenance** | Update each service | Centralized agent updates |

---

## Integration with Grafana Stack

### LGTM Stack Overview

**LGTM** = **L**oki + **G**rafana + **T**empo + **M**imir (+ **A**lloy)

The Grafana observability stack provides a complete, open-source solution for logs, metrics, traces, and profiling.

### Components

#### 1. Grafana Alloy

**Role**: **Unified telemetry aggregation and routing**

**Key Features**:
- OpenTelemetry Collector distribution (vendor-neutral)
- Programmable pipelines with River configuration language
- Supports OpenTelemetry, Prometheus, and custom protocols
- Built-in transformations, filtering, and routing

**Integration with OBI**:
- Alloy receives OTLP data from OBI agent
- Routes traces to Tempo
- Routes metrics to Mimir
- Routes logs to Loki
- Performs filtering, sampling, and enrichment

**Configuration Example**:
```hcl
// Receive OTLP from OBI
otelcol.receiver.otlp "obi" {
  grpc {
    endpoint = "0.0.0.0:4317"
  }
  http {
    endpoint = "0.0.0.0:4318"
  }

  output {
    traces  = [otelcol.processor.batch.default.input]
    metrics = [otelcol.processor.batch.default.input]
  }
}

// Batch processing
otelcol.processor.batch "default" {
  output {
    traces  = [otelcol.exporter.otlp.tempo.input]
    metrics = [otelcol.exporter.prometheus.mimir.input]
  }
}

// Export to Tempo
otelcol.exporter.otlp "tempo" {
  client {
    endpoint = "tempo:4317"
  }
}

// Export to Mimir
otelcol.exporter.prometheus "mimir" {
  endpoint = "http://mimir:9009/api/v1/push"
}
```

#### 2. Grafana Tempo

**Role**: **Distributed tracing backend**

**Key Features**:
- Cost-efficient (object storage only, no index)
- Deeply integrated with Grafana, Prometheus, Loki
- Supports Jaeger, Zipkin, OpenTelemetry protocols
- TraceQL query language for powerful trace search

**OBI Integration**:
- Receives trace spans from OBI via Alloy
- Stores traces in object storage (S3, GCS, Azure Blob)
- Provides trace visualization in Grafana
- Correlates traces with metrics (exemplars) and logs

**Benefits for OBI**:
- No trace indexing = low cost at scale
- Query traces by service, operation, duration, tags
- Automatic correlation with RED metrics from OBI

#### 3. Grafana Loki

**Role**: **Log aggregation system**

**Key Features**:
- Horizontally scalable, multi-tenant
- Indexes labels, not log content (cost-efficient)
- Inspired by Prometheus label model
- LogQL query language

**OBI Integration**:
- OBI doesn't capture logs directly (protocol-level only)
- **Correlation via trace context**: Logs enriched with trace_id
- Query logs for a specific trace: `{trace_id="abc123"}`
- Grafana UI shows traces and related logs side-by-side

**Use Case**:
1. OBI captures trace showing 500 error
2. User clicks "Show Logs" in Grafana trace view
3. Loki queries logs with same trace_id
4. Root cause revealed in application logs

#### 4. Grafana Mimir

**Role**: **Long-term Prometheus metrics storage**

**Key Features**:
- Horizontally scalable for millions of metrics
- Compatible with Prometheus (remote_write)
- Multi-tenant with per-tenant limits
- Fast query performance with sharding

**OBI Integration**:
- OBI exports RED metrics (request rate, error rate, duration)
- Alloy forwards metrics to Mimir
- Store metrics for long-term analysis (months/years)
- Visualize in Grafana dashboards

**Metrics from OBI**:
```
# HTTP metrics
http_server_duration_seconds{service="api", method="GET", path="/users", status="200"}
http_server_request_rate{service="api"}
http_server_error_rate{service="api"}

# gRPC metrics
rpc_server_duration_seconds{service="grpc-backend", method="/api.UserService/GetUser"}

# SQL metrics
db_client_duration_seconds{service="api", db_system="postgresql", operation="SELECT"}
```

#### 5. Grafana (Visualization)

**Role**: **Unified observability UI**

**Key Features**:
- Dashboards for metrics, logs, traces, profiles
- Explore view for ad-hoc queries
- Automatic correlation between signals
- Alerting and notification

**OBI + LGTM Workflow**:
1. **Dashboard**: View RED metrics from Mimir (powered by OBI)
2. **Click spike**: Drill down to traces in Tempo
3. **Select slow trace**: See distributed trace spans (OBI-captured)
4. **Click span**: View related logs from Loki
5. **Root cause**: Identify slow database query or external API call

### Architecture Diagram: OBI + LGTM Stack

```
┌─────────────────────────────────────────────────────────────────┐
│                    Kubernetes Cluster                            │
│                                                                   │
│  ┌──────────────┐      ┌──────────────┐      ┌──────────────┐  │
│  │ Application  │      │ Application  │      │ Application  │  │
│  │   Pod 1      │      │   Pod 2      │      │   Pod 3      │  │
│  └──────────────┘      └──────────────┘      └──────────────┘  │
│         │                     │                     │            │
│         │ (eBPF instrumentation at kernel level)    │            │
│         │                     │                     │            │
│  ┌──────┼─────────────────────┼─────────────────────┼──────┐   │
│  │      ▼                     ▼                     ▼       │   │
│  │  ┌──────────────────────────────────────────────────┐   │   │
│  │  │        OBI DaemonSet (one per node)              │   │   │
│  │  │  - eBPF probes capture HTTP, gRPC, SQL, etc.    │   │   │
│  │  │  - Exports OTLP to Alloy                        │   │   │
│  │  └──────────────────────────────────────────────────┘   │   │
│  │                         │                                 │   │
│  │                         │ OTLP (traces, metrics)          │   │
│  │                         ▼                                 │   │
│  │  ┌──────────────────────────────────────────────────┐   │   │
│  │  │        Grafana Alloy (StatefulSet)               │   │   │
│  │  │  - Receives OTLP                                 │   │   │
│  │  │  - Routes traces → Tempo                         │   │   │
│  │  │  - Routes metrics → Mimir                        │   │   │
│  │  │  - Sampling, filtering, enrichment               │   │   │
│  │  └──────────────────────────────────────────────────┘   │   │
│  │              │                      │                     │   │
│  └──────────────┼──────────────────────┼─────────────────────┘   │
│                 │                      │                         │
└─────────────────┼──────────────────────┼─────────────────────────┘
                  │                      │
          Traces  │              Metrics │
                  ▼                      ▼
       ┌─────────────────┐    ┌─────────────────┐
       │  Grafana Tempo  │    │  Grafana Mimir  │
       │  (Traces)       │    │  (Metrics)      │
       │  - Object store │    │  - Long-term    │
       │  - TraceQL      │    │  - PromQL       │
       └─────────────────┘    └─────────────────┘
                  │                      │
                  └──────────┬───────────┘
                             │
                             ▼
                  ┌─────────────────┐
                  │    Grafana      │
                  │  (Visualization)│
                  │  - Dashboards   │
                  │  - Explore      │
                  │  - Alerts       │
                  └─────────────────┘
```

---

## Deployment Strategies

### Alloy Operator vs Standalone Deployment

#### Alloy Operator (Recommended for Production)

**Overview**: Kubernetes operator that manages Alloy lifecycle declaratively

**Key Advantages**:
1. **Declarative Configuration**: Manage via `kind: Alloy` custom resources
2. **Automatic Configuration**: Operator configures Alloy based on high-level settings
3. **Simplified Management**: Single Helm chart deploys operator + Alloy instances
4. **Dynamic Scaling**: Automatically adjusts based on load and requirements
5. **Built-in Best Practices**: Default configurations follow Grafana recommendations

**Use Cases**:
- Production Kubernetes environments
- Multi-tenant deployments requiring isolation
- Dynamic workloads with auto-scaling needs
- Teams preferring declarative GitOps workflows

**Installation**:
```bash
# Add Grafana Helm repo
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install Alloy Operator
helm install alloy-operator grafana/alloy-operator \
  --namespace alloy-system \
  --create-namespace

# Deploy Alloy instance
kubectl apply -f - <<EOF
apiVersion: alloy.grafana.com/v1alpha1
kind: Alloy
metadata:
  name: alloy-main
  namespace: alloy-system
spec:
  mode: deployment
  replicas: 3
  config: |
    otelcol.receiver.otlp "obi" {
      grpc { endpoint = "0.0.0.0:4317" }
      http { endpoint = "0.0.0.0:4318" }
      output {
        traces = [otelcol.processor.batch.default.input]
      }
    }

    otelcol.processor.batch "default" {
      output {
        traces = [otelcol.exporter.otlp.tempo.input]
      }
    }

    otelcol.exporter.otlp "tempo" {
      client {
        endpoint = "tempo:4317"
      }
    }
EOF
```

**Pros**:
- Less operational overhead (operator handles updates)
- Consistent configuration across clusters
- Easier to scale and manage multiple instances
- Built-in health checks and auto-remediation

**Cons**:
- Additional complexity (operator + CRDs)
- Requires cluster-admin permissions to install operator
- Learning curve for Alloy CRD schema

#### Standalone Helm Deployment

**Overview**: Direct Helm chart deployment without operator

**Key Advantages**:
1. **Direct Control**: Full control over Helm chart values
2. **Simpler Architecture**: No operator dependency
3. **Faster Initial Setup**: Single Helm install command
4. **Flexible Deployment Modes**: DaemonSet, Deployment, StatefulSet

**Use Cases**:
- Development and testing environments
- Smaller deployments without complex requirements
- Teams preferring Helm-based workflows
- Environments with limited cluster permissions

**Installation**:
```bash
# Install Alloy standalone
helm install alloy grafana/alloy \
  --namespace alloy \
  --create-namespace \
  --set controller.type=deployment \
  --set controller.replicas=3 \
  --set-file config.content=alloy-config.yaml
```

**Pros**:
- Simpler architecture (no operator)
- Lower permission requirements
- Direct Helm value control
- Easier to understand for Helm users

**Cons**:
- Manual configuration updates required
- No automatic optimization or scaling
- More operational overhead for multi-instance setups

### Deployment Patterns

#### 1. DaemonSet (Node-Level Collection)

**Use Case**: Node-level metrics, pod logs, host monitoring

**Characteristics**:
- One pod per Kubernetes node
- Access to host network and filesystem
- Collects node-level metrics (cAdvisor, kubelet)

**Configuration**:
```yaml
controller:
  type: daemonset
  hostNetwork: true  # Access host network
  hostPID: true      # Access host processes

volumes:
  - name: hostfs
    hostPath:
      path: /
```

**Best For**:
- OBI DaemonSet deployment (kernel-level access required)
- Node exporter metrics
- Log collection from all pods

#### 2. Deployment (Stateless Workload)

**Use Case**: General telemetry aggregation, stateless processing

**Characteristics**:
- Standard deployment with N replicas
- No persistent storage or stable identities
- Horizontal scaling via replica count

**Configuration**:
```yaml
controller:
  type: deployment
  replicas: 3

autoscaling:
  enabled: true
  minReplicas: 3
  maxReplicas: 10
  targetCPUUtilizationPercentage: 70
```

**Best For**:
- OTLP receiver endpoints
- Stateless metric exporters
- Gateway deployments

#### 3. StatefulSet (Stateful Workload)

**Use Case**: Prometheus scraping, clustered mode, persistent WAL

**Characteristics**:
- Stable network identities (alloy-0, alloy-1, etc.)
- Persistent volumes per pod
- Ordered deployment and scaling

**Configuration**:
```yaml
controller:
  type: statefulset
  replicas: 3

persistence:
  enabled: true
  size: 10Gi
  storageClass: fast-ssd

clustering:
  enabled: true
```

**Best For**:
- Prometheus metrics collection (persistent WAL)
- Clustered Alloy instances with data distribution
- Any workload requiring persistent state

### OBI Deployment Pattern

**Recommended Setup**: Two-tier architecture

#### Tier 1: OBI DaemonSet

**Purpose**: Capture telemetry at node level

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
      serviceAccountName: obi-agent
      containers:
      - name: obi
        image: grafana/beyla:latest
        securityContext:
          privileged: true  # Required for eBPF
          capabilities:
            add:
            - SYS_ADMIN
            - SYS_PTRACE
            - NET_ADMIN
        env:
        - name: BEYLA_OPEN_PORT
          value: "8080,8443,9090"  # Ports to instrument
        - name: OTEL_EXPORTER_OTLP_ENDPOINT
          value: "http://alloy-gateway:4317"
        volumeMounts:
        - name: hostfs
          mountPath: /hostfs
          readOnly: true
      volumes:
      - name: hostfs
        hostPath:
          path: /
```

#### Tier 2: Alloy Gateway (StatefulSet or Deployment)

**Purpose**: Aggregate, process, route telemetry

```yaml
# Using Alloy Operator
apiVersion: alloy.grafana.com/v1alpha1
kind: Alloy
metadata:
  name: alloy-gateway
  namespace: observability
spec:
  mode: statefulset
  replicas: 3
  clustering:
    enabled: true
  config: |
    // Receive from OBI
    otelcol.receiver.otlp "obi" {
      grpc { endpoint = "0.0.0.0:4317" }
      http { endpoint = "0.0.0.0:4318" }
      output {
        traces = [otelcol.processor.tail_sampling.default.input]
      }
    }

    // Tail-based sampling for cost optimization
    otelcol.processor.tail_sampling "default" {
      policies = [
        {
          name = "errors"
          type = "status_code"
          status_code = { status_codes = ["ERROR"] }
        },
        {
          name = "slow"
          type = "latency"
          latency = { threshold_ms = 1000 }
        },
        {
          name = "sample"
          type = "probabilistic"
          probabilistic = { sampling_percentage = 10 }
        }
      ]
      output {
        traces = [otelcol.exporter.otlp.tempo.input]
      }
    }

    // Export to Tempo
    otelcol.exporter.otlp "tempo" {
      client {
        endpoint = "tempo-distributor:4317"
        tls { insecure = true }
      }
    }
```

### Decision Matrix: Operator vs Standalone

| Criteria | Operator | Standalone |
|----------|----------|------------|
| **Cluster Size** | Large (100+ nodes) | Small/Medium (< 100 nodes) |
| **Team Experience** | K8s operators, CRDs | Helm charts |
| **Deployment Model** | GitOps, declarative | Imperative, Helm |
| **Complexity Tolerance** | High (automated) | Low (manual) |
| **Multi-Tenancy** | Native support | Manual configuration |
| **Auto-Scaling** | Built-in | Manual HPA |
| **Best Fit** | Production, enterprise | Dev, test, small prod |

---

## Key Use Cases

### 1. Zero-Code Instrumentation for Legacy Applications

**Problem**: Monolithic applications running on outdated runtimes (Java 8, Python 2.7) without OpenTelemetry support

**Solution**: Deploy OBI as DaemonSet to automatically instrument without code changes

**Benefits**:
- No SDK installation or code modifications
- Works with unsupported language versions
- Instant observability for previously "black box" services
- Risk-free deployment (no application restart)

**Example**:
```yaml
# OBI automatically instruments legacy Java 8 app
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: obi-legacy
spec:
  template:
    spec:
      containers:
      - name: obi
        image: grafana/beyla:latest
        env:
        - name: BEYLA_SERVICE_NAME
          value: "legacy-java8-monolith"
        - name: BEYLA_OPEN_PORT
          value: "8080"  # Java app HTTP port
```

**Real-World Impact**:
- **Before OBI**: No traces, manual log analysis, blind to performance issues
- **After OBI**: Full distributed tracing, RED metrics, automatic root cause analysis

### 2. Multi-Protocol Service Observability

**Problem**: Microservices using mixed protocols (HTTP REST, gRPC, SQL, Redis, Kafka) require separate instrumentation

**Solution**: OBI provides unified instrumentation across all protocols

**Supported Protocols**:
- **Web**: HTTP/1.1, HTTP/2, HTTPS, gRPC
- **Databases**: PostgreSQL, MySQL, Redis, MongoDB
- **Message Queues**: Kafka
- **APIs**: REST, GraphQL, gRPC
- **Cloud**: AWS S3, Elasticsearch

**Example Architecture**:
```
API Gateway (HTTP)
  ↓
  → Backend Service (gRPC)
      ↓
      → PostgreSQL (SQL)
      → Redis (cache)
      → Kafka (events)

All instrumented by OBI with zero code changes
```

**Trace Visualization**:
```
Span 1: HTTP POST /api/orders [200ms]
  Span 2: gRPC OrderService.Create [180ms]
    Span 3: SQL INSERT INTO orders [50ms]
    Span 4: Redis SET order:123 [5ms]
    Span 5: Kafka SEND order.created [10ms]
```

### 3. Cost Optimization with Adaptive Sampling

**Problem**: High-volume production systems generate millions of traces, resulting in expensive storage and processing costs

**Solution**: Combine OBI with Alloy's tail-based sampling for intelligent trace retention

**Strategy**:
1. OBI captures all traces at kernel level (no sampling)
2. Alloy gateway performs tail-based sampling
3. Keep 100% of errors and slow requests
4. Sample 10% of normal requests

**Configuration**:
```hcl
otelcol.processor.tail_sampling "cost_optimized" {
  // Decision wait time (buffer traces)
  decision_wait = "10s"

  policies = [
    // Keep all errors (100%)
    {
      name = "errors"
      type = "status_code"
      status_code {
        status_codes = ["ERROR"]
      }
    },

    // Keep slow requests > 1s (100%)
    {
      name = "slow_requests"
      type = "latency"
      latency {
        threshold_ms = 1000
      }
    },

    // Sample 10% of normal requests
    {
      name = "sample_normal"
      type = "probabilistic"
      probabilistic {
        sampling_percentage = 10
      }
    }
  ]
}
```

**Cost Savings**:
- **Before**: 1M traces/day → $5,000/month storage
- **After**: 100K traces/day (90% reduction) → $500/month storage
- **Result**: 90% cost reduction with 100% error visibility

### 4. Multi-Tenancy with Dynamic Routing

**Problem**: SaaS platform with multiple tenants requires isolated telemetry pipelines

**Solution**: Alloy routing connector dynamically routes traces based on tenant ID

**Architecture**:
```
OBI (all tenants)
  ↓
Alloy Gateway (routing by tenant_id)
  ├─ Tenant A → Tempo A
  ├─ Tenant B → Tempo B
  └─ Tenant C → Tempo C
```

**Configuration**:
```hcl
// Add tenant ID from K8s namespace
otelcol.processor.resource "add_tenant" {
  attributes = [
    {
      key = "tenant_id"
      from_attribute = "k8s.namespace.name"
      action = "insert"
    }
  ]
  output {
    traces = [otelcol.connector.routing.tenants.input]
  }
}

// Route by tenant
otelcol.connector.routing "tenants" {
  from_attribute = "tenant_id"
  attribute_source = "resource"

  table = [
    {
      value = "tenant-a"
      pipelines = ["traces/tenant_a"]
    },
    {
      value = "tenant-b"
      pipelines = ["traces/tenant_b"]
    }
  ]

  default_pipelines = ["traces/default"]
}
```

**Benefits**:
- Tenant isolation (no data mixing)
- Per-tenant sampling policies
- Independent backend scaling
- Compliance (data residency requirements)

### 5. Performance Benchmarking & A/B Testing

**Problem**: Need to measure performance impact of code changes or infrastructure updates

**Solution**: Use OBI's minimal overhead for unbiased performance measurement

**Scenario**: A/B test new database connection pool

**Setup**:
```yaml
# Version A: Old connection pool (10 connections)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-v1
  labels:
    version: v1
spec:
  template:
    metadata:
      labels:
        version: v1
---
# Version B: New connection pool (50 connections)
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-v2
  labels:
    version: v2
```

**Analysis with OBI + Grafana**:
```sql
-- TraceQL query in Tempo
{
  resource.service.name = "api"
  && resource.version = "v1"
  && duration > 100ms
}

{
  resource.service.name = "api"
  && resource.version = "v2"
  && duration > 100ms
}
```

**Results Dashboard**:
```
Version A (v1):
  - p50: 150ms
  - p95: 500ms
  - p99: 1200ms
  - Error rate: 0.5%

Version B (v2):
  - p50: 80ms (47% improvement)
  - p95: 250ms (50% improvement)
  - p99: 600ms (50% improvement)
  - Error rate: 0.2%

Winner: Version B (rollout to 100%)
```

**Why OBI is Ideal**:
- Zero instrumentation overhead (< 1% CPU)
- Consistent measurement across versions
- No SDK version differences
- Captures infrastructure-level details

---

## Experimental Implementations

### Experiment 1: Adaptive Tail-Based Sampling with SLO Integration

**Objective**: Dynamically adjust sampling rates based on SLO breaches

**Concept**: Increase sampling when p95 latency exceeds SLO to capture more debug data

**Architecture**:
```
Mimir (metrics)
  ↓ Alert when p95 > SLO
Alertmanager
  ↓ Webhook
Custom Controller
  ↓ Update ConfigMap
Alloy (reloads config)
  ↓ Increase sampling rate
Tempo (captures more traces)
```

**Implementation**:

**Step 1**: Prometheus rule for SLO breach
```yaml
apiVersion: monitoring.coreos.com/v1
kind: PrometheusRule
metadata:
  name: slo-latency-breach
spec:
  groups:
  - name: slo
    rules:
    - alert: HighLatency
      expr: |
        histogram_quantile(0.95,
          rate(http_server_duration_seconds_bucket[5m])
        ) > 1.0  # SLO: p95 < 1s
      for: 5m
      annotations:
        summary: "p95 latency exceeds SLO"
```

**Step 2**: Webhook receiver adjusts sampling
```go
// Custom controller
func handleSLOBreach(w http.ResponseWriter, r *http.Request) {
    // Increase sampling from 10% to 50%
    updateAlloyConfig("sampling_percentage", "50")

    // Auto-revert after 30 minutes
    time.AfterFunc(30*time.Minute, func() {
        updateAlloyConfig("sampling_percentage", "10")
    })
}
```

**Step 3**: Alloy config with variable sampling
```hcl
otelcol.processor.tail_sampling "adaptive" {
  policies = [
    {
      name = "sample"
      type = "probabilistic"
      probabilistic {
        // Dynamically adjusted via ConfigMap
        sampling_percentage = env("SAMPLING_PERCENTAGE")
      }
    }
  ]
}
```

**Expected Results**:
- **Normal operation**: 10% sampling, low cost
- **SLO breach**: 50% sampling for 30 minutes, capture debug data
- **Cost impact**: Temporary 5x increase, acceptable for troubleshooting

**Benefits**:
- Automatic response to issues
- Captures detailed traces when needed most
- Cost-efficient during normal operation

### Experiment 2: Network-Level Service Dependency Discovery

**Objective**: Automatically discover service dependencies from network traffic (no APM required)

**Concept**: OBI captures network-level HTTP calls, build service graph automatically

**Architecture**:
```
OBI DaemonSet
  ↓ Captures all HTTP calls
  ↓ source_service → destination_service
Alloy
  ↓ Aggregates calls into service graph
Tempo
  ↓ Stores dependency data
Grafana
  ↓ Visualizes service graph
```

**Implementation**:

**Step 1**: OBI captures network calls
```yaml
# OBI auto-discovers all HTTP traffic
# No configuration needed - kernel-level capture
```

**Step 2**: Alloy extracts service relationships
```hcl
otelcol.processor.servicegraph "dependencies" {
  // Build service graph from spans
  latency_histogram_buckets = [0.1, 0.5, 1, 2, 5, 10]
  dimensions = ["service", "http.method", "http.status_code"]

  output {
    metrics = [otelcol.exporter.prometheus.mimir.input]
  }
}
```

**Step 3**: Query service graph
```promql
# Outbound calls from api service
sum by (service, destination_service) (
  rate(traces_service_graph_calls_total{service="api"}[5m])
)

# Result:
api → user-service: 100 req/s
api → payment-service: 50 req/s
api → notification-service: 20 req/s
```

**Step 4**: Grafana visualization
```json
{
  "type": "nodeGraph",
  "datasource": "Mimir",
  "targets": [
    {
      "expr": "traces_service_graph_calls_total"
    }
  ]
}
```

**Expected Results**:
- Automatic service dependency map
- No code instrumentation required
- Real-time updates as topology changes
- Historical dependency tracking

**Use Cases**:
- Migration planning (identify dependencies before decommissioning)
- Blast radius analysis (what services depend on X?)
- Architecture documentation (auto-generated diagrams)

### Experiment 3: Database Query Performance Profiling

**Objective**: Identify slow SQL queries without database instrumentation

**Concept**: OBI captures SQL queries at protocol level, analyze performance patterns

**Architecture**:
```
Application (uninstrumented)
  ↓
PostgreSQL (uninstrumented)
  ↑ (SQL traffic captured by OBI eBPF)
OBI DaemonSet
  ↓ SQL spans with query + duration
Tempo
  ↓
Grafana (query analysis)
```

**Implementation**:

**Step 1**: Enable SQL instrumentation in OBI
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: obi-config
data:
  config.yaml: |
    instrumentation:
      sql:
        enabled: true
        # Security: sanitize queries (remove literals)
        sanitize_queries: true
```

**Step 2**: Query slow SQL spans in Tempo
```sql
-- TraceQL: Find slow SQL queries
{
  span.kind = "client"
  && span.db.system = "postgresql"
  && duration > 1s
}
| group by span.db.statement
| count()
| sort by count desc
```

**Step 3**: Grafana dashboard for SQL analysis
```json
{
  "panels": [
    {
      "title": "Slowest SQL Queries (p95)",
      "targets": [
        {
          "datasource": "Tempo",
          "query": "{span.db.system=\"postgresql\"}"
        }
      ],
      "transformations": [
        {
          "id": "groupBy",
          "options": {
            "fields": {
              "db.statement": { "operation": "groupby" },
              "duration": { "operation": "p95" }
            }
          }
        }
      ]
    }
  ]
}
```

**Expected Results**:
```
Top Slow Queries:
1. SELECT * FROM orders WHERE status = ? (p95: 2.5s, count: 1000)
2. SELECT * FROM products JOIN categories ... (p95: 1.8s, count: 500)
3. UPDATE inventory SET quantity = ? ... (p95: 1.2s, count: 300)
```

**Optimization Actions**:
- Query 1: Add index on `orders.status`
- Query 2: Optimize join with covering index
- Query 3: Batch updates to reduce transactions

**Benefits**:
- Zero database instrumentation
- Identifies N+1 queries automatically
- Correlates slow queries with application traces
- Security-conscious (sanitized queries)

### Experiment 4: Cost-Optimized Multi-Region Observability

**Objective**: Reduce cross-region data transfer costs for global deployments

**Concept**: Deploy regional Tempo instances, replicate only aggregated metrics centrally

**Architecture**:
```
Region US-East:
  OBI → Alloy → Tempo (local)
            ↓ Aggregated metrics
Region EU-West:          ↓
  OBI → Alloy → Tempo (local) → Mimir (central)
            ↓ Aggregated metrics ↑
Region AP-South:         ↑
  OBI → Alloy → Tempo (local) ─┘
```

**Implementation**:

**Step 1**: Regional Tempo deployments
```yaml
# US-East Tempo
apiVersion: v1
kind: ConfigMap
metadata:
  name: tempo-us-east
data:
  tempo.yaml: |
    storage:
      trace:
        backend: s3
        s3:
          bucket: tempo-us-east
          region: us-east-1
---
# EU-West Tempo
apiVersion: v1
kind: ConfigMap
metadata:
  name: tempo-eu-west
data:
  tempo.yaml: |
    storage:
      trace:
        backend: s3
        s3:
          bucket: tempo-eu-west
          region: eu-west-1
```

**Step 2**: Regional Alloy with metric aggregation
```hcl
// Regional Alloy configuration
otelcol.receiver.otlp "obi" {
  grpc { endpoint = "0.0.0.0:4317" }
  output {
    traces = [
      // Keep traces local (Tempo regional)
      otelcol.exporter.otlp.tempo_local.input,

      // Send aggregated metrics to central Mimir
      otelcol.processor.metricstransform.aggregate.input
    ]
  }
}

// Local Tempo (no cross-region egress)
otelcol.exporter.otlp "tempo_local" {
  client {
    endpoint = "tempo.local:4317"  // Same region
  }
}

// Aggregate and send metrics centrally
otelcol.processor.metricstransform "aggregate" {
  // Aggregate to RED metrics only
  transforms = [
    {
      include = ".*"
      match_type = "regexp"
      action = "aggregate"
      aggregation = {
        aggregation_type = "histogram"
      }
    }
  ]
  output {
    metrics = [otelcol.exporter.prometheusremotewrite.central.input]
  }
}

// Central Mimir (cross-region, low bandwidth)
otelcol.exporter.prometheusremotewrite "central" {
  endpoint {
    url = "https://mimir-central.global/api/v1/push"
  }
}
```

**Cost Analysis**:

**Scenario**: 1TB traces/month per region, 3 regions

**Option A: Centralized (naive)**
- Cross-region transfer: 3TB x $0.09/GB = $270/month
- Storage: 3TB x $0.023/GB = $69/month
- **Total**: $339/month

**Option B: Regional with aggregated metrics**
- Cross-region transfer: 10GB (metrics only) x $0.09/GB = $0.90/month
- Regional storage: 3 x 1TB x $0.023/GB = $69/month
- **Total**: $69.90/month

**Savings**: **$269.10/month (79% reduction)**

**Trade-offs**:
- Traces stay regional (need VPN for cross-region query)
- Metrics available globally (RED metrics, SLO monitoring)
- Acceptable for most use cases (errors/latency visible globally)

### Experiment 5: Canary Deployment Automated Rollback

**Objective**: Automatically rollback canary deployments based on OBI telemetry

**Concept**: Compare error rates and latency between stable and canary versions, rollback if canary degrades

**Architecture**:
```
OBI (captures both versions)
  ↓
Alloy (routes by version label)
  ↓
Tempo + Mimir
  ↓ Metrics comparison
Argo Rollouts (progressive delivery)
  ↓ Automated rollback decision
```

**Implementation**:

**Step 1**: Argo Rollouts canary deployment
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: api-service
spec:
  replicas: 10
  strategy:
    canary:
      steps:
      - setWeight: 10  # 10% traffic to canary
      - pause: {duration: 5m}  # Collect metrics
      - setWeight: 50
      - pause: {duration: 5m}
      - setWeight: 100

      analysis:
        templates:
        - templateName: obi-error-rate
        - templateName: obi-latency-p95
        startingStep: 1  # Start after 10% traffic
```

**Step 2**: Analysis templates (query OBI metrics)
```yaml
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: obi-error-rate
spec:
  metrics:
  - name: error-rate
    interval: 1m
    successCondition: result < 0.05  # < 5% errors
    failureLimit: 3
    provider:
      prometheus:
        address: http://mimir:9009
        query: |
          sum(rate(http_server_request_total{
            service="api-service",
            version="{{args.canary-hash}}",
            status=~"5.."
          }[5m]))
          /
          sum(rate(http_server_request_total{
            service="api-service",
            version="{{args.canary-hash}}"
          }[5m]))
---
apiVersion: argoproj.io/v1alpha1
kind: AnalysisTemplate
metadata:
  name: obi-latency-p95
spec:
  metrics:
  - name: latency-p95
    interval: 1m
    successCondition: result < 1.0  # p95 < 1s
    failureLimit: 3
    provider:
      prometheus:
        address: http://mimir:9009
        query: |
          histogram_quantile(0.95,
            rate(http_server_duration_seconds_bucket{
              service="api-service",
              version="{{args.canary-hash}}"
            }[5m])
          )
```

**Step 3**: Automated rollback flow
```
1. Deploy canary (10% traffic)
   ↓
2. OBI captures metrics for 5 minutes
   ↓
3. Argo Rollouts queries Mimir:
   - Canary error rate: 8% (FAIL)
   - Stable error rate: 2%
   ↓
4. Analysis fails → Automatic rollback
   ↓
5. Alert sent to team
```

**Expected Results**:
- Automated quality gates based on OBI telemetry
- Zero-code instrumentation (OBI) enables safe deployments
- Faster rollout cycles (confidence in automated rollback)
- Reduced MTTR (mean time to recovery)

**Benefits**:
- Progressive delivery with telemetry validation
- No manual intervention for rollbacks
- Works with any language (OBI is language-agnostic)
- Historical comparison (canary vs stable)

---

## Helm Charts & Kubernetes

### Official Helm Charts

#### 1. OpenTelemetry eBPF Helm Chart

**Repository**: https://artifacthub.io/packages/helm/opentelemetry-helm/opentelemetry-ebpf

**Installation**:
```bash
# Add OpenTelemetry Helm repo
helm repo add opentelemetry-helm https://open-telemetry.github.io/opentelemetry-helm-charts
helm repo update

# Install OBI
helm install obi opentelemetry-helm/opentelemetry-ebpf \
  --namespace observability \
  --create-namespace \
  --set daemonset.enabled=true \
  --set config.otelExporterOtlpEndpoint="http://alloy-gateway:4317"
```

**Key Configuration Options**:
```yaml
# values.yaml
daemonset:
  enabled: true
  hostNetwork: true
  hostPID: true

config:
  # OTLP export endpoint
  otelExporterOtlpEndpoint: "http://alloy-gateway:4317"

  # Ports to instrument
  openPorts: "8080,8443,9090"

  # Services to instrument (regex)
  serviceNamespace: "production"

  # Protocol-specific settings
  instrumentation:
    http:
      enabled: true
    grpc:
      enabled: true
    sql:
      enabled: true
      sanitizeQueries: true
    redis:
      enabled: true
    kafka:
      enabled: true

securityContext:
  privileged: true
  capabilities:
    add:
    - SYS_ADMIN
    - SYS_PTRACE
    - NET_ADMIN
    - BPF

resources:
  requests:
    cpu: 100m
    memory: 128Mi
  limits:
    cpu: 500m
    memory: 512Mi
```

#### 2. Grafana Alloy Operator Helm Chart

**Repository**: https://github.com/grafana/alloy-operator

**Installation**:
```bash
# Add Grafana Helm repo
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Install Alloy Operator
helm install alloy-operator grafana/alloy-operator \
  --namespace alloy-system \
  --create-namespace
```

**Deploy Alloy Instance**:
```yaml
apiVersion: alloy.grafana.com/v1alpha1
kind: Alloy
metadata:
  name: alloy-gateway
  namespace: observability
spec:
  mode: statefulset
  replicas: 3

  clustering:
    enabled: true

  config: |
    // Full Alloy River configuration here
    otelcol.receiver.otlp "obi" {
      grpc { endpoint = "0.0.0.0:4317" }
      http { endpoint = "0.0.0.0:4318" }
      output {
        traces = [otelcol.processor.batch.default.input]
      }
    }
```

#### 3. Grafana LGTM Stack Helm Chart

**Repository**: https://github.com/grafana/lgtm-stack

**Installation** (All-in-one):
```bash
# Install complete LGTM stack
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install lgtm grafana/lgtm-stack \
  --namespace observability \
  --create-namespace \
  --set alloy.enabled=true \
  --set tempo.enabled=true \
  --set mimir.enabled=true \
  --set loki.enabled=true \
  --set grafana.enabled=true
```

**Key Components**:
- Alloy (telemetry pipeline)
- Tempo (traces)
- Mimir (metrics)
- Loki (logs)
- Grafana (visualization)

### Complete OBI + LGTM Deployment

**values.yaml** (combined):
```yaml
# OBI DaemonSet
obi:
  enabled: true
  daemonset:
    hostNetwork: true
    hostPID: true
  config:
    otelExporterOtlpEndpoint: "http://alloy-gateway:4317"
    openPorts: "8080,8443,9090"
    instrumentation:
      http: true
      grpc: true
      sql: true
      redis: true
      kafka: true
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
    limits:
      cpu: 500m
      memory: 512Mi

# Alloy Gateway
alloy:
  enabled: true
  controller:
    type: statefulset
    replicas: 3
  clustering:
    enabled: true
  config:
    content: |
      // OTLP receiver from OBI
      otelcol.receiver.otlp "obi" {
        grpc { endpoint = "0.0.0.0:4317" }
        http { endpoint = "0.0.0.0:4318" }
        output {
          traces = [otelcol.processor.batch.default.input]
          metrics = [otelcol.processor.batch.default.input]
        }
      }

      // Batch processing
      otelcol.processor.batch "default" {
        send_batch_size = 1024
        timeout = 10s
        output {
          traces = [otelcol.processor.tail_sampling.cost_optimized.input]
          metrics = [otelcol.exporter.prometheus.mimir.input]
        }
      }

      // Tail-based sampling (cost optimization)
      otelcol.processor.tail_sampling "cost_optimized" {
        decision_wait = "10s"
        policies = [
          {
            name = "errors"
            type = "status_code"
            status_code { status_codes = ["ERROR"] }
          },
          {
            name = "slow"
            type = "latency"
            latency { threshold_ms = 1000 }
          },
          {
            name = "sample"
            type = "probabilistic"
            probabilistic { sampling_percentage = 10 }
          }
        ]
        output {
          traces = [otelcol.exporter.otlp.tempo.input]
        }
      }

      // Export to Tempo
      otelcol.exporter.otlp "tempo" {
        client {
          endpoint = "tempo-distributor:4317"
          tls { insecure = true }
        }
      }

      // Export to Mimir
      otelcol.exporter.prometheus "mimir" {
        endpoint {
          url = "http://mimir-distributor:9009/api/v1/push"
        }
      }
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

# Tempo (traces)
tempo:
  enabled: true
  replicas: 3
  storage:
    trace:
      backend: s3
      s3:
        bucket: tempo-traces
        region: us-east-1
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

# Mimir (metrics)
mimir:
  enabled: true
  replicas: 3
  storage:
    backend: s3
    s3:
      bucket: mimir-metrics
      region: us-east-1
  resources:
    requests:
      cpu: 1000m
      memory: 2Gi
    limits:
      cpu: 4000m
      memory: 8Gi

# Loki (logs)
loki:
  enabled: true
  replicas: 3
  storage:
    backend: s3
    s3:
      bucket: loki-logs
      region: us-east-1
  resources:
    requests:
      cpu: 500m
      memory: 1Gi
    limits:
      cpu: 2000m
      memory: 4Gi

# Grafana (visualization)
grafana:
  enabled: true
  replicas: 2
  datasources:
    - name: Tempo
      type: tempo
      url: http://tempo-query-frontend:3100
      isDefault: false
    - name: Mimir
      type: prometheus
      url: http://mimir-query-frontend:8080/prometheus
      isDefault: true
    - name: Loki
      type: loki
      url: http://loki-query-frontend:3100
      isDefault: false
  dashboards:
    enabled: true
    providers:
      - name: default
        folder: Observability
        type: file
        disableDeletion: false
        editable: true
        options:
          path: /var/lib/grafana/dashboards
  resources:
    requests:
      cpu: 250m
      memory: 512Mi
    limits:
      cpu: 1000m
      memory: 2Gi
```

**Deploy Complete Stack**:
```bash
# Install OBI + LGTM stack
helm install observability . \
  --namespace observability \
  --create-namespace \
  --values values.yaml \
  --timeout 15m
```

**Verify Deployment**:
```bash
# Check all pods
kubectl get pods -n observability

# Expected output:
NAME                                READY   STATUS    RESTARTS   AGE
obi-agent-xxxxx                    1/1     Running   0          2m
obi-agent-yyyyy                    1/1     Running   0          2m
alloy-gateway-0                    1/1     Running   0          2m
alloy-gateway-1                    1/1     Running   0          2m
alloy-gateway-2                    1/1     Running   0          2m
tempo-distributor-xxxxx            1/1     Running   0          2m
tempo-ingester-0                   1/1     Running   0          2m
tempo-query-frontend-xxxxx         1/1     Running   0          2m
mimir-distributor-xxxxx            1/1     Running   0          2m
mimir-ingester-0                   1/1     Running   0          2m
mimir-query-frontend-xxxxx         1/1     Running   0          2m
loki-distributor-xxxxx             1/1     Running   0          2m
loki-ingester-0                    1/1     Running   0          2m
loki-query-frontend-xxxxx          1/1     Running   0          2m
grafana-xxxxx                      1/1     Running   0          2m

# Test OBI instrumentation
kubectl port-forward -n observability svc/grafana 3000:80

# Open browser: http://localhost:3000
# Login: admin / (check secret)
# Navigate to Explore → Tempo → Search for traces
```

---

## Comparison with Traditional Backends

### Traditional APM Approach

**Architecture**:
```
Application
  ↓ (SDK embedded in process)
OpenTelemetry SDK
  ↓ (in-process sampling)
OTLP Exporter
  ↓
Collector / Backend
```

**Characteristics**:
- **Installation**: Add SDK to each application (dependency)
- **Language Support**: Per-language SDKs (Java, Python, Go, etc.)
- **Performance Impact**: 5-15% CPU, high memory (GC pressure)
- **Instrumentation**: Manual (code changes) or auto-instrumentation (agent)
- **Custom Attributes**: Full support for application-specific metadata
- **Protocol Coverage**: Library-dependent (HTTP, gRPC, DB drivers)
- **Deployment**: Per-application (SDK + config in each service)

### OBI Approach

**Architecture**:
```
Application (unmodified)
  ↓
Linux Kernel
  ↓ (eBPF probes)
OBI Agent (DaemonSet)
  ↓
Collector / Backend
```

**Characteristics**:
- **Installation**: Deploy DaemonSet (no application changes)
- **Language Support**: Universal (kernel-level, language-agnostic)
- **Performance Impact**: < 1% CPU, minimal memory (kernel-space)
- **Instrumentation**: Zero-code (automatic protocol detection)
- **Custom Attributes**: Limited (protocol-level data only)
- **Protocol Coverage**: Protocol-level (HTTP, gRPC, SQL, Redis, Kafka)
- **Deployment**: Centralized (one DaemonSet per node)

### Detailed Comparison Table

| Aspect | Traditional SDK/APM | OBI (eBPF) |
|--------|---------------------|------------|
| **Installation Complexity** | High (per-service) | Low (DaemonSet) |
| **Code Changes Required** | Yes (import SDK) | No |
| **Application Restart** | Yes | No |
| **Language Coverage** | Per-language SDKs | Universal (any language) |
| **Runtime Version Support** | Limited (recent versions) | Universal (Java 8, Python 2.7, etc.) |
| **CPU Overhead** | 5-15% | < 1% |
| **Memory Overhead** | 50-200MB per service | 50-100MB per node |
| **Latency Impact** | 1-5ms per request | 0ms (kernel-level) |
| **Custom Attributes** | Full support | Limited (protocol-level) |
| **Business Logic Tracing** | Yes (spans within methods) | No (HTTP/RPC boundaries only) |
| **Protocol Support** | Library-dependent | Protocol-level (HTTP, gRPC, SQL, etc.) |
| **Sampling Control** | In-process (head-based) | External (tail-based supported) |
| **Deployment Model** | Distributed (per-app) | Centralized (per-node) |
| **Maintenance Overhead** | High (update each service) | Low (update DaemonSet) |
| **Legacy Application Support** | Poor (no SDK for old runtimes) | Excellent (kernel-level) |
| **Security** | SDK vulnerabilities | Kernel-level isolation |
| **Vendor Lock-in** | Possible (proprietary SDKs) | None (OpenTelemetry standard) |
| **Cost** | High (per-app overhead) | Low (shared per-node) |

### When to Use Each Approach

#### Use Traditional SDK When:
1. **Deep Application Insights Needed**: Custom spans, business-logic tracing, detailed context
2. **Supported Runtime**: Modern language versions with good SDK support
3. **Full Control**: Team wants fine-grained sampling and attribute control
4. **Greenfield Projects**: New applications being built from scratch
5. **Performance Acceptable**: 5-15% overhead is acceptable trade-off

**Example**: New microservice architecture, modern stack (Java 21, Python 3.12, Go 1.22)

#### Use OBI When:
1. **Zero-Code Requirement**: Cannot modify application code or add dependencies
2. **Legacy Applications**: Old runtimes without SDK support (Java 8, Python 2.7)
3. **Minimal Overhead Critical**: High-performance applications where 5% CPU matters
4. **Brownfield Projects**: Existing applications without instrumentation
5. **Universal Coverage**: Consistent telemetry across many languages
6. **Cost Optimization**: Minimize per-service overhead

**Example**: Legacy monolith migration, multi-language microservices, cost-sensitive SaaS platform

#### Hybrid Approach (Recommended):
**Combine Both** for complementary coverage:
- **OBI**: Protocol-level RED metrics and distributed tracing (universal, zero-code)
- **Traditional SDK**: Application-specific custom spans and attributes (selective, high-value services)

**Example Architecture**:
```
API Gateway (OBI only)
  ↓
Auth Service (OBI + SDK)
  - OBI: HTTP/gRPC/SQL spans
  - SDK: Custom spans (token validation, user lookup)
  ↓
User Service (OBI + SDK)
  - OBI: HTTP/Redis/SQL spans
  - SDK: Business logic spans (user profile enrichment)
  ↓
Legacy Service (OBI only)
  - OBI: HTTP/SQL spans
  - No SDK (Java 8, cannot modify)
```

**Benefits of Hybrid**:
- Universal coverage (OBI ensures every service is observable)
- Deep insights where needed (SDK for critical paths)
- Cost-optimized (SDK overhead only for high-value services)
- Gradual migration (add SDK incrementally)

---

## Recommendations

### Quick Start Path

**Goal**: Get OBI + LGTM stack running in under 30 minutes

**Prerequisites**:
- Kubernetes cluster (v1.24+)
- kubectl configured
- Helm 3 installed
- 4 CPU, 8GB RAM minimum

**Steps**:

1. **Install Alloy Operator** (5 minutes)
```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update
helm install alloy-operator grafana/alloy-operator \
  --namespace alloy-system \
  --create-namespace
```

2. **Install LGTM Stack** (10 minutes)
```bash
helm install lgtm grafana/lgtm-stack \
  --namespace observability \
  --create-namespace \
  --set tempo.enabled=true \
  --set mimir.enabled=true \
  --set loki.enabled=true \
  --set grafana.enabled=true
```

3. **Install OBI** (5 minutes)
```bash
helm repo add opentelemetry-helm https://open-telemetry.github.io/opentelemetry-helm-charts
helm install obi opentelemetry-helm/opentelemetry-ebpf \
  --namespace observability \
  --set daemonset.enabled=true \
  --set config.otelExporterOtlpEndpoint="http://lgtm-alloy:4317"
```

4. **Access Grafana** (1 minute)
```bash
kubectl port-forward -n observability svc/lgtm-grafana 3000:80
# Open: http://localhost:3000
# Default credentials: admin / (get from secret)
```

5. **Verify Traces** (1 minute)
- Navigate to **Explore** → **Tempo**
- Search for traces (should see OBI-captured data)
- View service graph automatically generated

### Production Readiness Checklist

#### Infrastructure
- [ ] Kubernetes cluster with 3+ nodes (HA)
- [ ] Object storage configured (S3, GCS, Azure Blob)
- [ ] Persistent volumes for WAL (Alloy, Tempo, Mimir)
- [ ] Network policies configured (isolation)
- [ ] Resource quotas and limits defined

#### OBI Configuration
- [ ] DaemonSet deployed on all nodes
- [ ] Security contexts configured (CAP_BPF, CAP_SYS_PTRACE)
- [ ] Ports to instrument defined (8080, 8443, etc.)
- [ ] Protocol instrumentation enabled (HTTP, gRPC, SQL, Redis, Kafka)
- [ ] OTLP endpoint configured (Alloy gateway)
- [ ] Resource limits set (CPU: 500m, Memory: 512Mi per node)

#### Alloy Configuration
- [ ] StatefulSet with 3+ replicas (HA)
- [ ] Clustering enabled (load distribution)
- [ ] Tail-based sampling configured (cost optimization)
- [ ] Batch processing enabled (performance)
- [ ] OTLP receiver configured (from OBI)
- [ ] Exporters configured (Tempo, Mimir)
- [ ] Resource limits set (CPU: 2000m, Memory: 4Gi per replica)

#### Tempo Configuration
- [ ] Distributed mode (ingester, distributor, query-frontend)
- [ ] Object storage backend configured (S3, GCS)
- [ ] Retention policy configured (14-30 days recommended)
- [ ] Compaction enabled (reduce storage costs)
- [ ] Query frontend with caching (performance)
- [ ] Multi-tenancy configured (if required)

#### Mimir Configuration
- [ ] Distributed mode (ingester, distributor, query-frontend)
- [ ] Object storage backend configured (S3, GCS)
- [ ] Long-term retention (90+ days)
- [ ] Compaction and downsampling enabled
- [ ] Query caching configured
- [ ] Alertmanager configured

#### Grafana Configuration
- [ ] Datasources configured (Tempo, Mimir, Loki)
- [ ] Dashboards provisioned (RED metrics, service graph)
- [ ] Alerts configured (SLO breaches, errors)
- [ ] RBAC configured (team access control)
- [ ] SSO/LDAP configured (if required)

#### Monitoring & Alerting
- [ ] OBI agent health checks (DaemonSet status)
- [ ] Alloy pipeline metrics (throughput, errors)
- [ ] Tempo ingestion rate and latency
- [ ] Mimir cardinality and query performance
- [ ] Grafana availability and dashboard load times
- [ ] Alerts for component failures

#### Cost Optimization
- [ ] Tail-based sampling configured (10% baseline)
- [ ] Object storage lifecycle policies (delete old data)
- [ ] Regional deployments (reduce cross-region egress)
- [ ] Metric cardinality limits (prevent explosion)
- [ ] Log aggregation rules (reduce log volume)

### Interesting Experiments to Try

#### 1. Zero-Code Migration Challenge
**Goal**: Instrument 10+ microservices in under 1 hour without code changes

**Steps**:
1. Deploy OBI DaemonSet
2. Label services to instrument: `obi.instrument: "true"`
3. Wait 5 minutes for telemetry to appear
4. Visualize service graph in Grafana

**Success Criteria**:
- All services show RED metrics
- Distributed traces captured end-to-end
- No application restarts or code changes

#### 2. Cost Optimization Challenge
**Goal**: Reduce observability costs by 80% while maintaining error visibility

**Steps**:
1. Measure baseline: 1M traces/day = $5,000/month
2. Implement tail-based sampling:
   - 100% errors
   - 100% slow requests (> 1s)
   - 10% normal requests
3. Measure new cost: 200K traces/day = $1,000/month

**Success Criteria**:
- 80% cost reduction achieved
- 0% error trace loss
- p95 latency still visible

#### 3. Multi-Tenancy Challenge
**Goal**: Deploy SaaS platform with 5 tenants, isolated telemetry pipelines

**Steps**:
1. Deploy OBI (shared)
2. Deploy Alloy with routing connector (by namespace)
3. Deploy 5 Tempo instances (per tenant)
4. Configure Grafana with tenant-specific datasources

**Success Criteria**:
- Each tenant sees only their traces
- No data leakage between tenants
- Single OBI deployment serves all tenants

#### 4. Adaptive Sampling Challenge
**Goal**: Dynamically adjust sampling based on SLO breaches

**Steps**:
1. Define SLO: p95 < 1s
2. Configure Prometheus alert: p95 > 1s for 5 minutes
3. Webhook triggers sampling increase: 10% → 50%
4. Auto-revert after 30 minutes

**Success Criteria**:
- Automatic response to SLO breach (< 1 minute)
- Capture 5x more traces during incident
- Cost impact minimized (temporary spike only)

#### 5. Database Query Profiling Challenge
**Goal**: Identify top 10 slow SQL queries without DB instrumentation

**Steps**:
1. Enable SQL instrumentation in OBI
2. Run production workload for 1 hour
3. Query Tempo for slow SQL spans (duration > 1s)
4. Group by query, aggregate p95 latency
5. Optimize top 3 queries (add indexes, rewrite)

**Success Criteria**:
- Top 10 slow queries identified
- Zero database instrumentation required
- p95 latency improved by 50% after optimization

### Next Steps

1. **Start Small**: Deploy OBI + LGTM stack in dev/staging environment
2. **Validate Telemetry**: Verify traces, metrics, and service graph accuracy
3. **Tune Sampling**: Implement tail-based sampling for cost optimization
4. **Expand Coverage**: Roll out to production incrementally (one service/team at a time)
5. **Train Team**: Provide training on TraceQL, PromQL, and Grafana dashboards
6. **Measure Impact**: Track MTTR, MTTD, and observability cost savings
7. **Iterate**: Continuously improve sampling policies, dashboards, and alerts

---

## Sources & References

### Official Documentation
- **OpenTelemetry OBI**: https://opentelemetry.io/docs/zero-code/obi/
- **OBI First Release**: https://opentelemetry.io/blog/2025/obi-announcing-first-release/
- **Grafana Beyla**: https://grafana.com/oss/beyla-ebpf/
- **Grafana Alloy**: https://grafana.com/docs/alloy/latest/
- **Grafana Tempo**: https://grafana.com/docs/tempo/latest/
- **Grafana Mimir**: https://grafana.com/docs/mimir/latest/
- **Grafana Loki**: https://grafana.com/docs/loki/latest/

### GitHub Repositories
- **OBI**: https://github.com/open-telemetry/opentelemetry-ebpf-instrumentation
- **Alloy Operator**: https://github.com/grafana/alloy-operator
- **OpenTelemetry Helm Charts**: https://github.com/open-telemetry/opentelemetry-helm-charts

### Blog Posts & Articles
- **Beyla Donation to OpenTelemetry**: https://grafana.com/blog/2025/05/07/opentelemetry-ebpf-instrumentation-beyla-donation/
- **Tail Sampling with OpenTelemetry**: https://opentelemetry.io/blog/2022/tail-sampling/
- **Multi-Tenant Observability**: https://aaronbytestream.medium.com/multi-tenant-distributed-tracing-withopentelemetry-86e1cf940d2e
- **Alloy Operator Configuration**: https://grafana.com/blog/2025/06/17/configure-and-customize-kubernetes-monitoring-easier-with-alloy-operator/

### Community Resources
- **OpenTelemetry Community**: https://opentelemetry.io/community/
- **Grafana Labs Community**: https://community.grafana.com/
- **CNCF Slack**: https://slack.cncf.io/ (#opentelemetry, #grafana)

---

**Report Prepared By**: Research Agent (Claude Flow)
**Last Updated**: November 6, 2025
**Version**: 1.0
