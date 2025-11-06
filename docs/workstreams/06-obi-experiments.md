# Workstream 6: OBI Experiments

## Status
🔴 Not Started

## Overview
Implement the five comprehensive OBI experiments to validate observability capabilities, data quality, performance characteristics, and operational patterns. This workstream focuses on practical validation of the OBI deployment through controlled experiments, creating validation dashboards, and documenting findings for operational teams.

## Objectives
- [ ] Implement Experiment 1: Baseline Network Observability
- [ ] Implement Experiment 2: Service Mesh Integration
- [ ] Implement Experiment 3: High-Cardinality Metrics
- [ ] Implement Experiment 4: Trace-to-Metrics Correlation
- [ ] Implement Experiment 5: Performance Under Load
- [ ] Create validation dashboards for each experiment
- [ ] Document findings and operational recommendations

## Agent Assignment
**Suggested Agent Type**: `researcher`, `tester`, `perf-analyzer`, `reviewer`
**Skill Requirements**: Observability engineering, distributed tracing, performance testing, data analysis, dashboard creation

## Dependencies
- Workstream 2 must complete OBI integration and deployment
- Workstream 3 must complete Grafana stack deployment
- Sample applications deployed for testing (microservices architecture)
- Load testing tools installed (k6, wrk, or similar)
- Access to Grafana for dashboard creation

## Tasks

### Task 6.1: Experiment 1 - Baseline Network Observability
**Description**: Validate basic OBI network observability capabilities by collecting and analyzing TCP/UDP traffic, connection states, and packet-level metrics.

**Experiment Goals**:
- Verify OBI captures all TCP connections
- Validate UDP traffic visibility
- Confirm packet loss detection
- Test connection state tracking
- Validate OTLP export of network metrics

**Deliverables**:
- Test application deployment (simple client-server)
- Network traffic generation scripts
- Validation queries for Mimir (metrics)
- Validation dashboards showing:
  - TCP connection rates and states
  - UDP packet flow
  - Packet loss and retransmission rates
  - Network throughput per pod
- Experiment report with findings

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/experiments/01-baseline/test-app.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/01-baseline/traffic-generator.sh`
- `/Users/beengud/raibid-labs/mop/experiments/01-baseline/validate.sh`
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/01-baseline-network.json`
- `/Users/beengud/raibid-labs/mop/docs/experiments/01-baseline-report.md`

**Validation**:
```bash
# Deploy test application
kubectl apply -f /Users/beengud/raibid-labs/mop/experiments/01-baseline/test-app.yaml -n mop-experiments

# Generate network traffic
cd /Users/beengud/raibid-labs/mop/experiments/01-baseline
./traffic-generator.sh --duration 5m --connections 100

# Validate OBI captured traffic
./validate.sh

# Query Mimir for metrics
kubectl port-forward -n mop-metrics svc/mimir-query-frontend 9009:9009 &
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=obi_tcp_connections_total"
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=obi_udp_packets_total"
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=obi_packet_loss_ratio"

# Import dashboard to Grafana
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/Users/beengud/raibid-labs/mop/dashboards/experiments/01-baseline-network.json

# Verify dashboard
open "http://localhost:3000/d/obi-exp-01/baseline-network-observability"
```

### Task 6.2: Experiment 2 - Service Mesh Integration
**Description**: Test OBI integration with service mesh (Istio/Linkerd) to compare eBPF-based observability with sidcar proxy telemetry.

**Experiment Goals**:
- Deploy sample application with service mesh
- Compare OBI metrics with Envoy/Linkerd metrics
- Validate data consistency and latency
- Identify gaps or discrepancies
- Test multi-protocol support (HTTP/1.1, HTTP/2, gRPC)

**Deliverables**:
- Service mesh deployment configuration
- Sample microservices application (3-5 services)
- Comparison analysis scripts
- Validation dashboards showing:
  - OBI vs mesh metrics comparison
  - Request rates and error rates
  - Latency distributions (p50, p95, p99)
  - Protocol-specific metrics
- Experiment report with recommendations

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/experiments/02-service-mesh/istio-setup.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/02-service-mesh/sample-app.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/02-service-mesh/compare-metrics.py`
- `/Users/beengud/raibid-labs/mop/experiments/02-service-mesh/validate.sh`
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/02-service-mesh-comparison.json`
- `/Users/beengud/raibid-labs/mop/docs/experiments/02-service-mesh-report.md`

**Validation**:
```bash
# Deploy Istio (or Linkerd)
cd /Users/beengud/raibid-labs/mop/experiments/02-service-mesh
istioctl install --set profile=demo -y
kubectl label namespace mop-experiments istio-injection=enabled

# Deploy sample application
kubectl apply -f sample-app.yaml -n mop-experiments

# Generate multi-protocol traffic
kubectl run load-test --image=fortio/fortio --rm -it -- \
  load -qps 100 -t 5m -c 10 http://frontend.mop-experiments.svc.cluster.local:8080

# Compare metrics
python compare-metrics.py --duration 5m --output comparison-report.html

# Run validation
./validate.sh

# Import comparison dashboard
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/Users/beengud/raibid-labs/mop/dashboards/experiments/02-service-mesh-comparison.json

# View results
open comparison-report.html
open "http://localhost:3000/d/obi-exp-02/service-mesh-comparison"
```

### Task 6.3: Experiment 3 - High-Cardinality Metrics
**Description**: Stress test OBI and Mimir's ability to handle high-cardinality metrics from containerized workloads.

**Experiment Goals**:
- Generate high-cardinality labels (pod IDs, request IDs)
- Test Mimir ingestion and query performance
- Validate OBI's resource usage under high cardinality
- Identify cardinality limits and optimization strategies
- Test time-series compaction and retention

**Deliverables**:
- High-cardinality workload generator
- Cardinality analysis scripts
- Performance benchmarks
- Validation dashboards showing:
  - Active time series count
  - Cardinality distribution
  - Query latency vs cardinality
  - Mimir resource usage
  - OBI overhead metrics
- Experiment report with tuning recommendations

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/experiments/03-high-cardinality/workload-generator.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/03-high-cardinality/cardinality-analysis.py`
- `/Users/beengud/raibid-labs/mop/experiments/03-high-cardinality/benchmark.sh`
- `/Users/beengud/raibid-labs/mop/experiments/03-high-cardinality/validate.sh`
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/03-high-cardinality.json`
- `/Users/beengud/raibid-labs/mop/docs/experiments/03-high-cardinality-report.md`

**Validation**:
```bash
# Deploy high-cardinality workload
cd /Users/beengud/raibid-labs/mop/experiments/03-high-cardinality
kubectl apply -f workload-generator.yaml -n mop-experiments

# Wait for cardinality to build up
sleep 300

# Analyze cardinality
python cardinality-analysis.py --output cardinality-report.json

# Run performance benchmarks
./benchmark.sh --queries 1000 --concurrency 10

# Validate results
./validate.sh

# Check Mimir metrics
kubectl port-forward -n mop-metrics svc/mimir-query-frontend 9009:9009 &
curl -X GET "http://localhost:9009/prometheus/api/v1/label/__name__/values" | jq '. | length'
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=cortex_ingester_active_series"

# Check OBI overhead
kubectl top pod -n mop-system -l app=obi

# Import dashboard
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/Users/beengud/raibid-labs/mop/dashboards/experiments/03-high-cardinality.json

# View analysis
cat cardinality-report.json | jq
open "http://localhost:3000/d/obi-exp-03/high-cardinality-analysis"
```

### Task 6.4: Experiment 4 - Trace-to-Metrics Correlation
**Description**: Validate end-to-end observability by correlating distributed traces with metrics using exemplars and derived fields.

**Experiment Goals**:
- Generate correlated traces and metrics
- Test exemplar support in Mimir
- Validate trace ID propagation
- Test Grafana's trace-to-metrics navigation
- Measure query performance for correlated data

**Deliverables**:
- Instrumented sample application (with OpenTelemetry)
- Trace generation scripts
- Correlation validation scripts
- Validation dashboards showing:
  - Metrics with exemplar links
  - Trace-to-metrics navigation
  - Span duration histograms
  - Error rate correlated with traces
  - RED (Rate, Errors, Duration) metrics
- Experiment report with workflow recommendations

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/experiments/04-trace-correlation/instrumented-app.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/04-trace-correlation/generate-traces.sh`
- `/Users/beengud/raibid-labs/mop/experiments/04-trace-correlation/validate-correlation.py`
- `/Users/beengud/raibid-labs/mop/experiments/04-trace-correlation/validate.sh`
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/04-trace-correlation.json`
- `/Users/beengud/raibid-labs/mop/docs/experiments/04-trace-correlation-report.md`

**Validation**:
```bash
# Deploy instrumented application
cd /Users/beengud/raibid-labs/mop/experiments/04-trace-correlation
kubectl apply -f instrumented-app.yaml -n mop-experiments

# Generate traces with errors
./generate-traces.sh --duration 10m --error-rate 0.05

# Validate trace-to-metrics correlation
python validate-correlation.py --sample-size 100

# Run validation script
./validate.sh

# Query Tempo for traces
kubectl port-forward -n mop-traces svc/tempo-query-frontend 3200:3200 &
curl -X GET "http://localhost:3200/api/search?limit=20" | jq

# Query Mimir for metrics with exemplars
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=http_request_duration_seconds_bucket" | jq '.data.result[0].exemplars'

# Import correlation dashboard
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/Users/beengud/raibid-labs/mop/dashboards/experiments/04-trace-correlation.json

# Test navigation in Grafana
open "http://localhost:3000/d/obi-exp-04/trace-correlation"
# Click on exemplar in metric graph, verify trace opens
```

### Task 6.5: Experiment 5 - Performance Under Load
**Description**: Stress test the entire observability stack under realistic production load to identify bottlenecks and capacity limits.

**Experiment Goals**:
- Simulate production-level traffic (10k+ RPS)
- Test OBI CPU/memory limits
- Test Grafana stack ingestion capacity
- Measure end-to-end latency (collection to visualization)
- Identify scaling requirements
- Test failure scenarios (component restarts)

**Deliverables**:
- Load testing infrastructure
- Performance test scenarios
- Bottleneck analysis reports
- Validation dashboards showing:
  - System throughput and latency
  - Resource utilization (CPU, memory, I/O)
  - Queue depths and backlogs
  - Error rates under load
  - Recovery time from failures
- Experiment report with capacity planning recommendations

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/experiments/05-performance/load-test-app.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/05-performance/k6-load-test.js`
- `/Users/beengud/raibid-labs/mop/experiments/05-performance/chaos-scenarios.yaml`
- `/Users/beengud/raibid-labs/mop/experiments/05-performance/analyze-performance.py`
- `/Users/beengud/raibid-labs/mop/experiments/05-performance/validate.sh`
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/05-performance-load.json`
- `/Users/beengud/raibid-labs/mop/docs/experiments/05-performance-report.md`

**Validation**:
```bash
# Deploy load test application
cd /Users/beengud/raibid-labs/mop/experiments/05-performance
kubectl apply -f load-test-app.yaml -n mop-experiments

# Run k6 load test (ramp up to 10k RPS)
k6 run --vus 100 --duration 15m k6-load-test.js

# Monitor system during load
kubectl top nodes
kubectl top pods -n mop-system
kubectl top pods -n mop-traces
kubectl top pods -n mop-metrics
kubectl top pods -n mop-logs

# Inject chaos (restart OBI pods during load)
kubectl delete pod -n mop-system -l app=obi

# Analyze performance data
python analyze-performance.py --load-test-results results.json --output performance-report.html

# Run validation
./validate.sh

# Check for data loss during chaos
curl -X GET "http://localhost:9009/prometheus/api/v1/query?query=rate(obi_data_points_dropped_total[5m])"

# Import performance dashboard
curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
  -H "Content-Type: application/json" \
  -d @/Users/beengud/raibid-labs/mop/dashboards/experiments/05-performance-load.json

# Review results
open performance-report.html
open "http://localhost:3000/d/obi-exp-05/performance-under-load"
```

### Task 6.6: Experiment Validation Dashboard Suite
**Description**: Create a comprehensive dashboard suite that consolidates all experiment results for easy comparison and analysis.

**Deliverables**:
- Master experiment dashboard
- Comparison views across experiments
- Automated dashboard provisioning
- Dashboard documentation
- Export templates for reporting

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/00-experiment-overview.json`
- `/Users/beengud/raibid-labs/mop/dashboards/experiments/experiments-comparison.json`
- `/Users/beengud/raibid-labs/mop/config/grafana/dashboards-experiments.yaml`
- `/Users/beengud/raibid-labs/mop/docs/experiment-dashboards.md`

**Validation**:
```bash
# Import all experiment dashboards
cd /Users/beengud/raibid-labs/mop/dashboards/experiments
for dashboard in *.json; do
  curl -u admin:$ADMIN_PASSWORD -X POST http://localhost:3000/api/dashboards/db \
    -H "Content-Type: application/json" \
    -d @$dashboard
done

# Verify all dashboards
curl -u admin:$ADMIN_PASSWORD http://localhost:3000/api/search?type=dash-db | jq -r '.[] | select(.title | contains("Experiment"))'

# Open experiment overview
open "http://localhost:3000/d/obi-exp-00/experiment-overview"

# Export dashboard for reporting
curl -u admin:$ADMIN_PASSWORD "http://localhost:3000/api/dashboards/uid/obi-exp-00" | jq > exported-overview.json
```

### Task 6.7: Experiment Documentation and Findings
**Description**: Compile all experiment findings into comprehensive documentation with operational recommendations.

**Deliverables**:
- Individual experiment reports
- Consolidated findings document
- Operational runbooks based on experiments
- Capacity planning guide
- Troubleshooting guide based on experiment learnings

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/docs/experiments/experiment-summary.md`
- `/Users/beengud/raibid-labs/mop/docs/experiments/operational-recommendations.md`
- `/Users/beengud/raibid-labs/mop/docs/experiments/capacity-planning.md`
- `/Users/beengud/raibid-labs/mop/docs/experiments/troubleshooting-guide.md`
- `/Users/beengud/raibid-labs/mop/docs/experiments/lessons-learned.md`

**Validation**:
```bash
# Verify all experiment reports exist
cd /Users/beengud/raibid-labs/mop/docs/experiments
ls -1 0*-report.md | wc -l  # Should be 5

# Check documentation completeness
grep -r "TODO\|FIXME\|TBD" *.md || echo "No TODOs found"

# Validate markdown
markdownlint *.md

# Check for broken links
markdown-link-check *.md

# Generate consolidated PDF report (requires pandoc)
pandoc experiment-summary.md operational-recommendations.md \
  -o mop-experiments-report.pdf \
  --pdf-engine=xelatex \
  --toc

# Review final report
open mop-experiments-report.pdf
```

## Definition of Done
- [ ] All 5 experiments completed successfully
- [ ] Validation dashboards created and tested
- [ ] All experiment reports written with findings
- [ ] Operational recommendations documented
- [ ] Capacity planning guide completed
- [ ] Troubleshooting guide based on experiments
- [ ] No data loss detected during stress tests
- [ ] Performance baselines established
- [ ] Scaling recommendations provided
- [ ] All dashboards imported to Grafana
- [ ] Experiment suite can be re-run for regression testing
- [ ] Code reviewed by at least one team member

## Agent Coordination Hooks
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-6-obi-experiments"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-6"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/experiments/01-baseline/validate.sh" --memory-key "swarm/mop/ws-6/exp-01"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/experiments/02-service-mesh/compare-metrics.py" --memory-key "swarm/mop/ws-6/exp-02"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/experiments/03-high-cardinality/benchmark.sh" --memory-key "swarm/mop/ws-6/exp-03"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/experiments/04-trace-correlation/validate-correlation.py" --memory-key "swarm/mop/ws-6/exp-04"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/experiments/05-performance/k6-load-test.js" --memory-key "swarm/mop/ws-6/exp-05"
npx claude-flow@alpha hooks notify --message "OBI experiments completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-6-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

## Estimated Effort
**Duration**: 8-10 days
**Complexity**: High

## References
- [OBI Documentation](https://github.com/obi-metrics/obi) (placeholder)
- [Grafana Exemplars Documentation](https://grafana.com/docs/grafana/latest/fundamentals/exemplars/)
- [OpenTelemetry Best Practices](https://opentelemetry.io/docs/concepts/observability-primer/)
- [k6 Load Testing](https://k6.io/docs/)
- [Chaos Engineering Principles](https://principlesofchaos.org/)
- [Prometheus Cardinality Best Practices](https://prometheus.io/docs/practices/naming/)

## Notes
- Run experiments in isolated namespace to avoid production impact
- Consider using separate Kubernetes cluster for load testing
- Experiments should be repeatable and automated
- Document any unexpected behaviors or bugs discovered
- Performance baselines are environment-specific
- High-cardinality tests may require increased Mimir resources
- Load tests should gradually ramp up to avoid shocking the system
- Monitor for memory leaks during long-running experiments
- Consider using recorded query results for dashboard demos
- Chaos experiments should have rollback plans
- Correlate experiment findings with vendor documentation
- Share findings with OBI and Grafana communities
- Create regression test suite based on experiments
- Budget extra time for troubleshooting unexpected issues
- Consider creating video demos of experiment results
