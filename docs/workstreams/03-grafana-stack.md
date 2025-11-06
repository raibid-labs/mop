# Workstream 3: Grafana Stack Deployment

## Status
🔴 Not Started

## Overview
Deploy the complete Grafana observability stack including Tempo (distributed tracing), Mimir (metrics), Loki (logs), and Grafana (visualization) with proper configuration, datasources, and pre-built dashboards. This provides a unified interface for querying and visualizing telemetry data collected by OBI and other instrumentation.

## Objectives
- [ ] Deploy Tempo for distributed tracing with S3-compatible storage backend
- [ ] Deploy Mimir for long-term metrics storage with high cardinality support
- [ ] Deploy Loki for log aggregation and querying
- [ ] Deploy Grafana with OAuth integration and RBAC
- [ ] Configure datasources for Tempo, Mimir, Loki, and Prometheus
- [ ] Import and customize pre-built dashboards
- [ ] Set up alerting and notification channels

## Agent Assignment
**Suggested Agent Type**: `backend-dev`, `system-architect`, `reviewer`
**Skill Requirements**: Kubernetes deployments, Grafana administration, time-series databases, distributed systems, observability best practices

## Dependencies
- Workstream 1 must complete storage class and RBAC configuration
- Object storage bucket for Tempo, Mimir, Loki (S3/GCS/MinIO)
- PostgreSQL or similar database for Grafana metadata
- Ingress controller configured for external access
- TLS certificates for secure communication

## Tasks

### Task 3.1: Tempo Deployment
**Description**: Deploy Grafana Tempo for distributed tracing with scalable backend storage and query capabilities.

**Deliverables**:
- Tempo StatefulSet or deployment via Helm
- Object storage configuration for trace data
- Tempo query frontend and distributor
- Ingester and compactor components
- OTLP receiver configuration
- Jaeger and Zipkin compatibility layer

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/helm/tempo/values.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/tempo/tempo-config.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/tempo/storage-secret.yaml`
- `/Users/beengud/raibid-labs/mop/tanka/lib/tempo/main.libsonnet`
- `/Users/beengud/raibid-labs/mop/docs/tempo-architecture.md`

**Validation**:
```bash
# Deploy Tempo via Helm
helm install tempo grafana/tempo -f /Users/beengud/raibid-labs/mop/helm/tempo/values.yaml -n mop-traces

# Verify deployment
kubectl get pods -n mop-traces -l app.kubernetes.io/name=tempo

# Check Tempo services
kubectl get svc -n mop-traces

# Test OTLP receiver
kubectl port-forward -n mop-traces svc/tempo 4317:4317 &
grpcurl -plaintext -d '{"resourceSpans":[]}' localhost:4317 opentelemetry.proto.collector.trace.v1.TraceService/Export

# Query Tempo API
kubectl port-forward -n mop-traces svc/tempo-query-frontend 3200:3200 &
curl http://localhost:3200/api/search?limit=10

# Verify object storage integration
kubectl logs -n mop-traces -l app.kubernetes.io/component=compactor | grep -i "uploaded"
```

### Task 3.2: Mimir Deployment
**Description**: Deploy Grafana Mimir for scalable, long-term metrics storage with high cardinality support.

**Deliverables**:
- Mimir microservices deployment (distributor, ingester, querier, compactor)
- Object storage configuration for metrics data
- Remote write configuration for Prometheus compatibility
- Query frontend with caching
- Tenant isolation configuration
- Compaction and retention policies

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/helm/mimir/values.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/mimir/mimir-config.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/mimir/storage-secret.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/mimir/limits.yaml`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mimir/main.libsonnet`

**Validation**:
```bash
# Deploy Mimir via Helm
helm install mimir grafana/mimir-distributed -f /Users/beengud/raibid-labs/mop/helm/mimir/values.yaml -n mop-metrics

# Verify all components
kubectl get pods -n mop-metrics -l app.kubernetes.io/name=mimir

# Check distributor
kubectl port-forward -n mop-metrics svc/mimir-distributor 8080:8080 &
curl http://localhost:8080/ready

# Test remote write
curl -X POST http://localhost:8080/api/v1/push -H "Content-Type: application/x-protobuf" --data-binary @sample-metrics.pb

# Query Mimir
kubectl port-forward -n mop-metrics svc/mimir-query-frontend 9009:9009 &
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=up"

# Verify compaction
kubectl logs -n mop-metrics -l app.kubernetes.io/component=compactor | grep -i "compacted"
```

### Task 3.3: Loki Deployment
**Description**: Deploy Grafana Loki for log aggregation, indexing, and querying with LogQL support.

**Deliverables**:
- Loki microservices deployment (distributor, ingester, querier)
- Object storage configuration for log data
- Index configuration for efficient querying
- Retention and compaction policies
- Promtail or Fluentd integration for log collection
- LogQL query optimization

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/helm/loki/values.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/loki/loki-config.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/loki/storage-secret.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/loki/promtail-daemonset.yaml`
- `/Users/beengud/raibid-labs/mop/tanka/lib/loki/main.libsonnet`

**Validation**:
```bash
# Deploy Loki via Helm
helm install loki grafana/loki -f /Users/beengud/raibid-labs/mop/helm/loki/values.yaml -n mop-logs

# Verify deployment
kubectl get pods -n mop-logs -l app.kubernetes.io/name=loki

# Check distributor
kubectl port-forward -n mop-logs svc/loki-distributor 3100:3100 &
curl http://localhost:3100/ready

# Send test logs
curl -X POST http://localhost:3100/loki/api/v1/push -H "Content-Type: application/json" --data '{"streams": [{"stream": {"app": "test"}, "values": [["1234567890000000000", "test log line"]]}]}'

# Query logs with LogQL
curl -X GET "http://localhost:3100/loki/api/v1/query?query=%7Bapp%3D%22test%22%7D"

# Verify Promtail collection
kubectl logs -n mop-logs -l app=promtail | grep -i "sent"
```

### Task 3.4: Grafana Deployment
**Description**: Deploy Grafana with OAuth integration, RBAC, and pre-configured datasources for Tempo, Mimir, and Loki.

**Deliverables**:
- Grafana deployment with persistent storage
- OAuth/OIDC configuration for authentication
- RBAC roles and permissions
- Datasource provisioning for Tempo, Mimir, Loki
- Default organization and team setup
- Dashboard provisioning
- Plugin management

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/helm/grafana/values.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/grafana/grafana-config.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/grafana/datasources.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/grafana/oauth-secret.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/grafana/ingress.yaml`

**Validation**:
```bash
# Deploy Grafana via Helm
helm install grafana grafana/grafana -f /Users/beengud/raibid-labs/mop/helm/grafana/values.yaml -n mop-system

# Verify deployment
kubectl get pods -n mop-system -l app.kubernetes.io/name=grafana

# Get admin password
kubectl get secret -n mop-system grafana -o jsonpath="{.data.admin-password}" | base64 --decode

# Port forward to access UI
kubectl port-forward -n mop-system svc/grafana 3000:80

# Test datasource connectivity
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/datasources
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/datasources/proxy/1/api/v1/query?query=up

# Verify OAuth login
curl -I http://localhost:3000/login
```

### Task 3.5: Datasource Configuration
**Description**: Configure and validate all datasources in Grafana for querying Tempo, Mimir, Loki, and Prometheus.

**Deliverables**:
- Tempo datasource with trace-to-metrics correlation
- Mimir datasource with exemplars support
- Loki datasource with derived fields
- Prometheus datasource (if separate)
- Datasource permissions and access control
- Query optimization settings

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/config/grafana/datasources/tempo.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/datasources/mimir.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/datasources/loki.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/datasources/prometheus.yaml`
- `/Users/beengud/raibid-labs/mop/tests/grafana/test-datasources.sh`

**Validation**:
```bash
# List all datasources
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/datasources | jq

# Test Tempo datasource
curl -u admin:$ADMIN_PASSWORD "http://localhost:3000/api/datasources/proxy/tempo/api/search?limit=10"

# Test Mimir datasource
curl -u admin:$ADMIN_PASSWORD "http://localhost:3000/api/datasources/proxy/mimir/api/v1/query?query=up"

# Test Loki datasource
curl -u admin:$ADMIN_PASSWORD "http://localhost:3000/api/datasources/proxy/loki/loki/api/v1/label"

# Run automated tests
/Users/beengud/raibid-labs/mop/tests/grafana/test-datasources.sh
```

### Task 3.6: Dashboard Provisioning
**Description**: Import, customize, and provision pre-built dashboards for OBI, Kubernetes, and application monitoring.

**Deliverables**:
- OBI network observability dashboard
- Kubernetes cluster overview dashboard
- Node and pod resource dashboard
- Distributed tracing dashboard
- Log analysis dashboard
- Alert overview dashboard
- Custom dashboard templates

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/dashboards/obi-network.json`
- `/Users/beengud/raibid-labs/mop/dashboards/kubernetes-cluster.json`
- `/Users/beengud/raibid-labs/mop/dashboards/node-resources.json`
- `/Users/beengud/raibid-labs/mop/dashboards/distributed-tracing.json`
- `/Users/beengud/raibid-labs/mop/dashboards/log-analysis.json`
- `/Users/beengud/raibid-labs/mop/config/grafana/dashboards.yaml`

**Validation**:
```bash
# List all dashboards
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/search?type=dash-db | jq

# Import dashboard via API
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/Users/beengud/raibid-labs/mop/dashboards/obi-network.json

# Export dashboard
curl -u admin:$ADMIN_PASSWORD "http://localhost:3000/api/dashboards/uid/obi-network" | jq > exported-dashboard.json

# Verify dashboard provisioning
kubectl logs -n mop-system -l app.kubernetes.io/name=grafana | grep -i "provisioning"
```

### Task 3.7: Alerting and Notifications
**Description**: Configure Grafana alerting rules and notification channels for proactive monitoring.

**Deliverables**:
- Alert rules for OBI agent failures
- Alert rules for high resource usage
- Alert rules for data ingestion failures
- Notification channels (Slack, PagerDuty, email)
- Alert grouping and routing policies
- Silence and inhibition rules

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/config/grafana/alerting/obi-alerts.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/alerting/resource-alerts.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/alerting/ingestion-alerts.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/notification-channels.yaml`
- `/Users/beengud/raibid-labs/mop/config/grafana/alert-routing.yaml`

**Validation**:
```bash
# List alert rules
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/v1/provisioning/alert-rules | jq

# Test alert notification
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/alerts/test \
  -H "Content-Type: application/json" \
  -d '{"name":"test-alert","message":"Test notification"}'

# Check notification channels
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/alert-notifications | jq

# Verify alert firing
curl -u admin:$ADMIN_PASSWORD "http://localhost:3000/api/alertmanager/grafana/api/v2/alerts"
```

## Definition of Done
- [ ] Tempo deployed and accepting OTLP traces
- [ ] Mimir deployed and accepting Prometheus remote write
- [ ] Loki deployed and ingesting logs from Promtail
- [ ] Grafana deployed with OAuth authentication working
- [ ] All datasources configured and tested
- [ ] Pre-built dashboards imported and functional
- [ ] Alert rules configured and tested
- [ ] Notification channels operational
- [ ] Ingress configured for external access with TLS
- [ ] Object storage backends verified
- [ ] Performance testing completed (query latency, ingestion rate)
- [ ] Documentation complete with screenshots
- [ ] Code reviewed by at least one team member

## Agent Coordination Hooks
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-3-grafana-stack"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-3"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/helm/tempo/values.yaml" --memory-key "swarm/mop/ws-3/tempo-config"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/helm/mimir/values.yaml" --memory-key "swarm/mop/ws-3/mimir-config"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/helm/loki/values.yaml" --memory-key "swarm/mop/ws-3/loki-config"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/helm/grafana/values.yaml" --memory-key "swarm/mop/ws-3/grafana-config"
npx claude-flow@alpha hooks notify --message "Grafana stack deployment completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-3-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

## Estimated Effort
**Duration**: 7-10 days
**Complexity**: High

## References
- [Grafana Tempo Documentation](https://grafana.com/docs/tempo/latest/)
- [Grafana Mimir Documentation](https://grafana.com/docs/mimir/latest/)
- [Grafana Loki Documentation](https://grafana.com/docs/loki/latest/)
- [Grafana Administration](https://grafana.com/docs/grafana/latest/administration/)
- [LogQL Language](https://grafana.com/docs/loki/latest/logql/)
- [PromQL Language](https://prometheus.io/docs/prometheus/latest/querying/basics/)
- [TraceQL Language](https://grafana.com/docs/tempo/latest/traceql/)

## Notes
- Object storage is critical for Tempo, Mimir, and Loki - ensure proper permissions
- Consider using memcached for query caching to improve performance
- Grafana requires persistent storage for dashboards and users if not using OAuth
- Tempo, Mimir, and Loki can be deployed in monolithic or microservices mode
- Microservices mode recommended for production (better scalability and resilience)
- Configure appropriate retention policies to manage storage costs
- Use exemplars in Mimir to link metrics to traces
- Configure derived fields in Loki to link logs to traces
- Grafana Enterprise features (not required) offer additional RBAC and audit logging
- Test disaster recovery procedures for all components
- Monitor object storage costs and implement lifecycle policies
- Consider using Grafana Agent as lightweight alternative to Promtail
- Ingress controller should support WebSocket for Grafana Live features
- TLS certificates can be managed via cert-manager
