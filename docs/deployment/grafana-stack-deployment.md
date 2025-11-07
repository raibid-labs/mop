# Grafana Stack Deployment Guide

## Overview
This guide documents the deployment of the complete Grafana observability stack for the MOP (Multi-cluster Observability Platform) project. The stack includes:
- **Grafana Alloy**: OpenTelemetry collector for telemetry ingestion
- **Tempo**: Distributed tracing backend
- **Mimir**: Long-term metrics storage
- **Loki**: Log aggregation system
- **Grafana**: Visualization and dashboard platform

## Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  OBI Agent  │────▶│    Alloy    │────▶│    Tempo    │
└─────────────┘     │  (Collector) │     └─────────────┘
                    │             │     ┌─────────────┐
                    │             │────▶│    Mimir    │
                    │             │     └─────────────┘
                    │             │     ┌─────────────┐
                    └─────────────┘────▶│    Loki     │
                                        └─────────────┘
                            │
                            ▼
                    ┌─────────────┐
                    │   Grafana   │
                    └─────────────┘
```

## Prerequisites

- Kubernetes cluster (1.24+)
- kubectl configured with cluster access
- Tanka (for Jsonnet deployment)
- Storage class available for PersistentVolumeClaims
- MinIO or S3-compatible storage for object storage

## Component Deployment

### 1. Directory Structure
```bash
mop/
├── lib/
│   ├── alloy.libsonnet         # Alloy collector configuration
│   ├── tempo.libsonnet         # Tempo tracing configuration
│   ├── mimir.libsonnet         # Mimir metrics configuration
│   ├── loki.libsonnet          # Loki logs configuration
│   ├── grafana.libsonnet       # Grafana visualization
│   └── grafana/
│       └── dashboards/         # Pre-built dashboards
├── environments/
│   ├── default/
│   │   └── main.jsonnet        # Default environment
│   ├── dev/
│   │   └── main.jsonnet        # Development environment
│   └── prod/
│       └── main.jsonnet        # Production environment
└── tests/
    └── grafana-stack-integration.sh
```

### 2. Environment Configuration

Each environment's `main.jsonnet` should import and configure the stack:

```jsonnet
local alloy = import '../../lib/alloy.libsonnet';
local tempo = import '../../lib/tempo.libsonnet';
local mimir = import '../../lib/mimir.libsonnet';
local loki = import '../../lib/loki.libsonnet';
local grafana = import '../../lib/grafana.libsonnet';

{
  _config:: {
    namespace: 'mop-system',
    domain: 'observability.local',
    replicas: {
      alloy: 2,
      tempo: 1,
      mimir: 1,
      loki: 1,
      grafana: 1,
    },
    resources: {
      alloy: {
        requests: { memory: '256Mi', cpu: '100m' },
        limits: { memory: '512Mi', cpu: '500m' },
      },
      tempo: {
        requests: { memory: '512Mi', cpu: '250m' },
        limits: { memory: '1Gi', cpu: '1000m' },
      },
      mimir: {
        requests: { memory: '512Mi', cpu: '250m' },
        limits: { memory: '1Gi', cpu: '1000m' },
      },
      loki: {
        requests: { memory: '256Mi', cpu: '100m' },
        limits: { memory: '512Mi', cpu: '500m' },
      },
      grafana: {
        requests: { memory: '256Mi', cpu: '100m' },
        limits: { memory: '512Mi', cpu: '500m' },
      },
    },
  },

  alloy: alloy.new($._config).all,
  tempo: tempo.new($._config).all,
  mimir: mimir.new($._config).all,
  loki: loki.new($._config).all,
  grafana: grafana.new($._config).all,
}
```

### 3. Deploy to Kubernetes

Using Tanka:
```bash
# Initialize environment (if not already done)
cd environments/dev
tk init

# Show what will be deployed
tk show .

# Apply to cluster
tk apply .

# For production
cd ../prod
tk apply .
```

Using kubectl with rendered manifests:
```bash
# Render manifests
tk show environments/dev > /tmp/grafana-stack.yaml

# Apply to cluster
kubectl apply -f /tmp/grafana-stack.yaml
```

## Component Details

### Grafana Alloy (OpenTelemetry Collector)
- **Purpose**: Receives and routes telemetry data
- **Endpoints**:
  - OTLP gRPC: `:4317`
  - OTLP HTTP: `:4318`
  - Internal metrics: `:12345`
- **Configuration**: River configuration language
- **Outputs**: Tempo (traces), Mimir (metrics), Loki (logs)

### Tempo (Distributed Tracing)
- **Purpose**: Store and query distributed traces
- **Storage**: S3-compatible object storage
- **Endpoints**:
  - OTLP gRPC: `:4317`
  - Tempo gRPC: `:9095`
  - HTTP API: `:3200`
- **Retention**: 7 days (configurable)
- **Features**: TraceQL query language, service graph, trace-to-metrics

### Mimir (Metrics Storage)
- **Purpose**: Long-term Prometheus metrics storage
- **Storage**: S3-compatible object storage + local disk
- **Endpoints**:
  - HTTP API: `:8080`
  - gRPC: `:9095`
- **Features**:
  - High cardinality support
  - Multi-tenancy (optional)
  - Exemplar support for trace correlation
  - PromQL compatibility

### Loki (Log Aggregation)
- **Purpose**: Horizontally scalable log aggregation
- **Storage**: S3-compatible object storage + local index
- **Endpoints**:
  - HTTP API: `:3100`
  - gRPC: `:9095`
- **Features**:
  - LogQL query language
  - Label-based indexing
  - Trace ID correlation
  - Log-to-metrics generation

### Grafana (Visualization)
- **Purpose**: Dashboard and visualization platform
- **Endpoints**:
  - Web UI: `:3000`
- **Default Credentials**: admin/admin (change immediately)
- **Datasources**:
  - Tempo (traces)
  - Mimir (metrics)
  - Loki (logs)
  - Prometheus (optional)
- **Features**:
  - Trace-to-logs correlation
  - Metrics-to-traces correlation
  - Pre-built dashboards
  - Alert management

## Dashboards

Pre-configured dashboards are available in `/lib/grafana/dashboards/`:

1. **OBI Overview** (`obi-overview.json`)
   - OBI agent status
   - Active agents count
   - Recent traces
   - Error logs
   - Log level distribution

2. **Alloy Pipeline** (`alloy-pipeline.json`)
   - Telemetry ingestion rates
   - Processor throughput
   - Export success rates
   - Resource usage

3. **Tempo Tracing** (`tempo-tracing.json`)
   - Service map
   - Span ingestion rate
   - Query latency
   - Recent traces table

4. **Mimir Metrics** (`mimir-metrics.json`)
   - Active series count
   - Ingestion/query rates
   - Query latency percentiles
   - Storage usage

5. **Loki Logs** (`loki-logs.json`)
   - Log stream count
   - Ingestion rate
   - Log levels over time
   - Recent logs viewer

## Integration Testing

Run the integration test to verify the deployment:

```bash
# Basic test (assumes port-forwarding is set up)
./tests/grafana-stack-integration.sh

# With automatic port-forwarding
./tests/grafana-stack-integration.sh --port-forward
```

The test will:
1. Check component availability
2. Send test metrics, traces, and logs
3. Query each backend for the test data
4. Verify Grafana datasource configuration
5. Provide instructions for manual correlation testing

## Trace-to-Logs-to-Metrics Correlation

The stack is configured with full correlation between signals:

1. **Trace to Logs**:
   - Tempo datasource configured with Loki integration
   - Automatic log queries based on trace ID
   - Span context preserved in logs

2. **Trace to Metrics**:
   - Exemplars in Mimir linked to Tempo traces
   - Service-level metrics derived from spans
   - RED metrics (Rate, Errors, Duration) generation

3. **Logs to Traces**:
   - Derived fields in Loki for trace ID extraction
   - Direct links from log lines to traces
   - Pattern matching for trace correlation

## Monitoring the Stack

### Health Checks
```bash
# Check pod status
kubectl get pods -n mop-system

# Check services
kubectl get svc -n mop-system

# View logs
kubectl logs -n mop-system deployment/alloy
kubectl logs -n mop-system statefulset/tempo
kubectl logs -n mop-system statefulset/mimir
kubectl logs -n mop-system statefulset/loki
kubectl logs -n mop-system deployment/grafana
```

### Key Metrics to Monitor
- **Alloy**: Receiver acceptance rate, exporter success rate
- **Tempo**: Span ingestion rate, query latency, compaction status
- **Mimir**: Sample ingestion rate, query performance, compaction
- **Loki**: Log ingestion rate, query performance, index size
- **Grafana**: Dashboard load times, datasource query errors

## Troubleshooting

### Common Issues

1. **Components not starting**
   - Check PVC binding: `kubectl get pvc -n mop-system`
   - Verify storage class exists: `kubectl get storageclass`

2. **No data in Grafana**
   - Verify datasources: Settings → Data Sources → Test
   - Check component logs for errors
   - Ensure network policies allow communication

3. **High memory usage**
   - Adjust resource limits in environment configuration
   - Enable sampling in Tempo
   - Configure retention policies

4. **Slow queries**
   - Check object storage connectivity
   - Verify cache configuration
   - Review query complexity

### Debug Commands
```bash
# Port-forward for direct access
kubectl port-forward -n mop-system svc/grafana 3000:3000
kubectl port-forward -n mop-system svc/tempo 3200:3200
kubectl port-forward -n mop-system svc/mimir 8080:8080
kubectl port-forward -n mop-system svc/loki 3100:3100

# Test endpoints
curl http://localhost:3200/ready  # Tempo
curl http://localhost:8080/ready  # Mimir
curl http://localhost:3100/ready  # Loki
curl http://localhost:3000/api/health  # Grafana
```

## Security Considerations

1. **Authentication**:
   - Change default Grafana admin password
   - Enable OAuth/OIDC for production
   - Configure RBAC policies

2. **Network Security**:
   - Use NetworkPolicies to restrict traffic
   - Enable TLS for all components
   - Configure ingress with TLS termination

3. **Data Security**:
   - Enable encryption at rest for object storage
   - Configure retention policies for compliance
   - Implement data masking for sensitive logs

## Scaling Considerations

### Horizontal Scaling
- **Alloy**: Increase replicas for higher ingestion
- **Tempo**: Scale ingesters and queriers separately
- **Mimir**: Scale by component (distributor, ingester, querier)
- **Loki**: Scale read and write paths independently
- **Grafana**: Add replicas behind load balancer

### Vertical Scaling
- Monitor resource usage and adjust limits
- Consider dedicated nodes for storage-intensive components
- Use node affinity for performance optimization

## Backup and Recovery

### What to Backup
- Grafana database (dashboards, users, settings)
- Object storage buckets (traces, metrics, logs)
- Configuration files (ConfigMaps)

### Backup Strategy
```bash
# Backup Grafana database
kubectl exec -n mop-system deployment/grafana -- \
  sqlite3 /var/lib/grafana/grafana.db ".backup /tmp/grafana.db"

# Copy to local
kubectl cp mop-system/grafana-xxx:/tmp/grafana.db ./grafana-backup.db

# Backup object storage (example with MinIO)
mc mirror minio/tempo-traces ./backup/tempo-traces
mc mirror minio/mimir-metrics ./backup/mimir-metrics
mc mirror minio/loki-logs ./backup/loki-logs
```

## Maintenance

### Regular Tasks
1. **Weekly**:
   - Review dashboard usage
   - Check storage consumption
   - Monitor query performance

2. **Monthly**:
   - Update component versions
   - Review and optimize queries
   - Clean up unused dashboards

3. **Quarterly**:
   - Security audit
   - Capacity planning review
   - Disaster recovery testing

## Conclusion

The Grafana stack provides a complete observability solution with:
- Unified telemetry collection (Alloy)
- Distributed tracing (Tempo)
- Long-term metrics storage (Mimir)
- Scalable log aggregation (Loki)
- Powerful visualization (Grafana)

The stack is designed for production use with:
- High availability options
- Horizontal scaling capabilities
- Full correlation between signals
- Enterprise-grade security features

For additional support and updates, refer to:
- [Grafana Documentation](https://grafana.com/docs/)
- [Tempo Documentation](https://grafana.com/docs/tempo/)
- [Mimir Documentation](https://grafana.com/docs/mimir/)
- [Loki Documentation](https://grafana.com/docs/loki/)
- [Alloy Documentation](https://grafana.com/docs/alloy/)