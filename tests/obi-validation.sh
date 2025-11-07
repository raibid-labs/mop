#!/bin/bash

# OBI eBPF Instrumentation Validation Script
# Validates OBI deployment across all environments

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

print_info() {
    echo -e "ℹ️  $1"
}

# Function to validate OBI in a specific namespace
validate_obi_deployment() {
    local namespace=$1
    echo ""
    echo "========================================="
    echo "Validating OBI in namespace: $namespace"
    echo "========================================="

    # Check if namespace exists
    if kubectl get namespace "$namespace" >/dev/null 2>&1; then
        print_success "Namespace $namespace exists"
    else
        print_error "Namespace $namespace not found"
        return 1
    fi

    # Check DaemonSet status
    echo ""
    echo "Checking OBI DaemonSet..."
    if kubectl get daemonset obi -n "$namespace" >/dev/null 2>&1; then
        print_success "OBI DaemonSet found"

        # Get DaemonSet details
        desired=$(kubectl get daemonset obi -n "$namespace" -o jsonpath='{.status.desiredNumberScheduled}')
        ready=$(kubectl get daemonset obi -n "$namespace" -o jsonpath='{.status.numberReady}')

        if [ "$desired" = "$ready" ] && [ "$ready" -gt 0 ]; then
            print_success "All OBI pods are ready ($ready/$desired)"
        else
            print_warning "OBI pods not fully ready ($ready/$desired)"
        fi
    else
        print_error "OBI DaemonSet not found"
        return 1
    fi

    # Check individual pods
    echo ""
    echo "Checking OBI pods..."
    pod_count=$(kubectl get pods -n "$namespace" -l app=obi --no-headers 2>/dev/null | wc -l)
    if [ "$pod_count" -gt 0 ]; then
        print_success "Found $pod_count OBI pod(s)"

        # Show pod status
        kubectl get pods -n "$namespace" -l app=obi -o wide

        # Check if all pods are running
        running_count=$(kubectl get pods -n "$namespace" -l app=obi --field-selector=status.phase=Running --no-headers 2>/dev/null | wc -l)
        if [ "$running_count" = "$pod_count" ]; then
            print_success "All OBI pods are running"
        else
            print_warning "Some OBI pods are not running ($running_count/$pod_count)"
        fi
    else
        print_error "No OBI pods found"
    fi

    # Check ConfigMap
    echo ""
    echo "Checking OBI ConfigMap..."
    if kubectl get configmap obi -n "$namespace" >/dev/null 2>&1; then
        print_success "OBI ConfigMap found"

        # Verify OTLP endpoint configuration
        endpoint=$(kubectl get configmap obi -n "$namespace" -o jsonpath='{.data.config\.yaml}' | grep -oP 'endpoint:\s*\K[^\s]+' | head -1)
        if [ -n "$endpoint" ]; then
            print_info "OTLP endpoint configured: $endpoint"
        else
            print_warning "OTLP endpoint not found in ConfigMap"
        fi
    else
        print_error "OBI ConfigMap not found"
    fi

    # Check RBAC
    echo ""
    echo "Checking OBI RBAC..."
    if kubectl get serviceaccount obi -n "$namespace" >/dev/null 2>&1; then
        print_success "OBI ServiceAccount found"
    else
        print_error "OBI ServiceAccount not found"
    fi

    if kubectl get clusterrole obi >/dev/null 2>&1; then
        print_success "OBI ClusterRole found"
    else
        print_error "OBI ClusterRole not found"
    fi

    if kubectl get clusterrolebinding obi >/dev/null 2>&1; then
        print_success "OBI ClusterRoleBinding found"
    else
        print_error "OBI ClusterRoleBinding not found"
    fi

    # Check Service
    echo ""
    echo "Checking OBI Service..."
    if kubectl get service obi-metrics -n "$namespace" >/dev/null 2>&1; then
        print_success "OBI metrics service found"
        kubectl get service obi-metrics -n "$namespace"
    else
        print_warning "OBI metrics service not found (optional)"
    fi

    # Check pod logs for errors
    echo ""
    echo "Checking OBI pod logs for errors..."
    for pod in $(kubectl get pods -n "$namespace" -l app=obi -o name); do
        pod_name=$(echo "$pod" | cut -d'/' -f2)
        error_count=$(kubectl logs "$pod" -n "$namespace" --tail=100 2>/dev/null | grep -iE "error|fatal|panic" | wc -l)
        if [ "$error_count" -eq 0 ]; then
            print_success "No errors in pod $pod_name logs"
        else
            print_warning "Found $error_count error(s) in pod $pod_name logs"
            echo "Recent errors:"
            kubectl logs "$pod" -n "$namespace" --tail=100 2>/dev/null | grep -iE "error|fatal|panic" | head -5
        fi
    done

    # Test health endpoint if pod is running
    echo ""
    echo "Testing OBI health endpoints..."
    for pod in $(kubectl get pods -n "$namespace" -l app=obi --field-selector=status.phase=Running -o name); do
        pod_name=$(echo "$pod" | cut -d'/' -f2)

        # Test health endpoint
        health_status=$(kubectl exec -n "$namespace" "$pod_name" -- wget -q -O - http://localhost:13133/health 2>/dev/null || echo "failed")
        if [ "$health_status" != "failed" ]; then
            print_success "Health check passed for pod $pod_name"
        else
            print_warning "Health check failed for pod $pod_name"
        fi

        # Test ready endpoint
        ready_status=$(kubectl exec -n "$namespace" "$pod_name" -- wget -q -O - http://localhost:13133/ready 2>/dev/null || echo "failed")
        if [ "$ready_status" != "failed" ]; then
            print_success "Readiness check passed for pod $pod_name"
        else
            print_warning "Readiness check failed for pod $pod_name"
        fi

        break  # Test only one pod
    done

    echo ""
    print_info "OBI validation completed for $namespace"
    return 0
}

# Main execution
echo "================================================"
echo "      OBI eBPF Instrumentation Validator"
echo "================================================"
echo ""
print_info "Starting OBI deployment validation..."

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    print_error "kubectl command not found. Please install kubectl."
    exit 1
fi

# Check cluster connectivity
if ! kubectl cluster-info >/dev/null 2>&1; then
    print_error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

print_success "Connected to Kubernetes cluster"

# Validate OBI in all environments
environments=("observability-dev" "observability-staging" "observability-production")
failed_envs=()

for env in "${environments[@]}"; do
    if ! validate_obi_deployment "$env"; then
        failed_envs+=("$env")
    fi
done

# Summary
echo ""
echo "========================================="
echo "             VALIDATION SUMMARY"
echo "========================================="

if [ ${#failed_envs[@]} -eq 0 ]; then
    print_success "All OBI deployments validated successfully!"
    exit 0
else
    print_error "OBI validation failed for the following environments:"
    for env in "${failed_envs[@]}"; do
        echo "  - $env"
    done
    exit 1
fi