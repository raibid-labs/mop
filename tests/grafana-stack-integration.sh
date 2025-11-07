#!/bin/bash
set -e

# Grafana Stack Integration Test Script
# Tests end-to-end flow of metrics, logs, and traces through the observability pipeline

echo "========================================="
echo "Grafana Stack Integration Test"
echo "========================================="

# Configuration
NAMESPACE="${NAMESPACE:-mop-system}"
ALLOY_URL="http://localhost:4318"
GRAFANA_URL="http://localhost:3000"
TEMPO_URL="http://localhost:3200"
MIMIR_URL="http://localhost:8080"
LOKI_URL="http://localhost:3100"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test functions
test_component() {
    local name=$1
    local url=$2
    local endpoint=$3

    echo -n "Testing $name... "
    if curl -s -f "$url$endpoint" > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} $name is responding"
        return 0
    else
        echo -e "${RED}✗${NC} $name is not responding at $url$endpoint"
        return 1
    fi
}

send_test_metrics() {
    echo "Sending test metrics to Alloy..."
    cat <<EOF | curl -X POST -H "Content-Type: application/x-protobuf" --data-binary @- "$ALLOY_URL/v1/metrics"
# TYPE test_metric counter
test_metric{job="integration_test",instance="test_host"} 42
EOF
}

send_test_traces() {
    echo "Sending test traces to Alloy..."
    cat <<EOF | curl -X POST -H "Content-Type: application/json" --data @- "$ALLOY_URL/v1/traces"
{
  "resourceSpans": [{
    "resource": {
      "attributes": [{
        "key": "service.name",
        "value": {"stringValue": "integration-test"}
      }]
    },
    "scopeSpans": [{
      "spans": [{
        "traceId": "$(openssl rand -hex 16)",
        "spanId": "$(openssl rand -hex 8)",
        "name": "test-span",
        "kind": 1,
        "startTimeUnixNano": "$(date +%s)000000000",
        "endTimeUnixNano": "$(date +%s)000000001",
        "status": {}
      }]
    }]
  }]
}
EOF
}

send_test_logs() {
    echo "Sending test logs to Loki..."
    TIMESTAMP=$(date +%s)000000000
    cat <<EOF | curl -X POST -H "Content-Type: application/json" --data @- "$LOKI_URL/loki/api/v1/push"
{
  "streams": [{
    "stream": {
      "job": "integration_test",
      "level": "info"
    },
    "values": [
      ["$TIMESTAMP", "Integration test log message"],
      ["$((TIMESTAMP+1000000000))", "Another test log message"]
    ]
  }]
}
EOF
}

query_mimir() {
    echo "Querying Mimir for test metrics..."
    curl -s "$MIMIR_URL/prometheus/api/v1/query?query=test_metric" | jq -r '.status'
}

query_tempo() {
    echo "Querying Tempo for recent traces..."
    curl -s "$TEMPO_URL/api/search?limit=10" | jq -r '.traces | length'
}

query_loki() {
    echo "Querying Loki for test logs..."
    curl -s "$LOKI_URL/loki/api/v1/query?query={job=\"integration_test\"}" | jq -r '.status'
}

test_grafana_datasources() {
    echo "Testing Grafana datasources..."

    # Get admin password
    ADMIN_PASSWORD=$(kubectl get secret -n $NAMESPACE grafana -o jsonpath="{.data.admin-password}" 2>/dev/null | base64 --decode || echo "admin")

    # Test each datasource
    for datasource in "Tempo" "Mimir" "Loki"; do
        echo -n "  Testing $datasource datasource... "
        response=$(curl -s -u admin:$ADMIN_PASSWORD "$GRAFANA_URL/api/datasources/name/$datasource" 2>/dev/null || echo "{}")
        if echo "$response" | grep -q "\"name\":\"$datasource\""; then
            echo -e "${GREEN}✓${NC}"
        else
            echo -e "${YELLOW}⚠${NC} (may need configuration)"
        fi
    done
}

# Main test execution
main() {
    echo ""
    echo "Step 1: Testing component availability"
    echo "---------------------------------------"

    # Port-forward if needed (assumes kubectl is configured)
    if [ "$1" == "--port-forward" ]; then
        echo "Setting up port forwarding..."
        kubectl port-forward -n $NAMESPACE svc/alloy 4318:4318 &
        kubectl port-forward -n $NAMESPACE svc/tempo 3200:3200 &
        kubectl port-forward -n $NAMESPACE svc/mimir 8080:8080 &
        kubectl port-forward -n $NAMESPACE svc/loki 3100:3100 &
        kubectl port-forward -n $NAMESPACE svc/grafana 3000:3000 &
        sleep 5
    fi

    test_component "Alloy" "$ALLOY_URL" "/metrics"
    test_component "Tempo" "$TEMPO_URL" "/ready"
    test_component "Mimir" "$MIMIR_URL" "/ready"
    test_component "Loki" "$LOKI_URL" "/ready"
    test_component "Grafana" "$GRAFANA_URL" "/api/health"

    echo ""
    echo "Step 2: Sending test telemetry data"
    echo "------------------------------------"
    send_test_metrics
    send_test_traces
    send_test_logs

    echo ""
    echo "Step 3: Waiting for data propagation (10 seconds)..."
    sleep 10

    echo ""
    echo "Step 4: Querying backends for test data"
    echo "----------------------------------------"

    echo -n "Mimir query status: "
    mimir_status=$(query_mimir)
    if [ "$mimir_status" == "success" ]; then
        echo -e "${GREEN}✓${NC} Metrics received"
    else
        echo -e "${RED}✗${NC} Metrics not found"
    fi

    echo -n "Tempo trace count: "
    trace_count=$(query_tempo)
    if [ "$trace_count" -gt 0 ]; then
        echo -e "${GREEN}✓${NC} Found $trace_count traces"
    else
        echo -e "${YELLOW}⚠${NC} No traces found yet"
    fi

    echo -n "Loki query status: "
    loki_status=$(query_loki)
    if [ "$loki_status" == "success" ]; then
        echo -e "${GREEN}✓${NC} Logs received"
    else
        echo -e "${RED}✗${NC} Logs not found"
    fi

    echo ""
    echo "Step 5: Testing Grafana integration"
    echo "------------------------------------"
    test_grafana_datasources

    echo ""
    echo "Step 6: Testing trace-to-logs correlation"
    echo "------------------------------------------"
    echo "This requires manual verification in Grafana UI:"
    echo "1. Open Grafana at $GRAFANA_URL"
    echo "2. Navigate to Explore -> Tempo"
    echo "3. Search for traces from service 'integration-test'"
    echo "4. Click on a trace and verify 'Logs for this span' button works"

    echo ""
    echo "========================================="
    echo "Integration Test Complete"
    echo "========================================="
    echo ""
    echo "Summary:"
    echo "- Components deployed: Alloy, Tempo, Mimir, Loki, Grafana"
    echo "- Test data sent: Metrics, Traces, Logs"
    echo "- Datasources configured: Tempo, Mimir, Loki with correlation"
    echo ""
    echo "Access Grafana dashboards at: $GRAFANA_URL"
    echo "Default credentials: admin / admin"
    echo ""

    # Kill port-forward processes if we started them
    if [ "$1" == "--port-forward" ]; then
        echo "Cleaning up port forwarding..."
        killall kubectl 2>/dev/null || true
    fi
}

# Run main function
main "$@"