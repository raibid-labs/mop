#!/bin/bash
# Validation script for MOP Jsonnet component libraries
set -e

echo "=== MOP Component Libraries Validation ==="
echo

# Check jsonnet installation
if ! command -v jsonnet &> /dev/null; then
    echo "ERROR: jsonnet is not installed"
    echo "Install with: brew install jsonnet"
    exit 1
fi

# Test each library individually
echo "Testing individual libraries..."
for lib in alloy obi tempo mimir loki grafana; do
    echo -n "  - ${lib}.libsonnet: "
    if jsonnet -e "(import 'lib/${lib}.libsonnet').new((import 'lib/config.libsonnet').environments.dev)" > /dev/null 2>&1; then
        echo "✓"
    else
        echo "✗ FAILED"
        jsonnet -e "(import 'lib/${lib}.libsonnet').new((import 'lib/config.libsonnet').environments.dev)"
        exit 1
    fi
done
echo

# Test examples
echo "Testing examples..."
echo -n "  - full-stack.jsonnet: "
if jsonnet lib/examples/full-stack.jsonnet > /dev/null 2>&1; then
    echo "✓"
    resources=$(jsonnet lib/examples/full-stack.jsonnet | jq -r '[.. | objects | select(.kind) | .kind] | unique | length')
    echo "    Generated ${resources} unique Kubernetes resource types"
else
    echo "✗ FAILED"
    jsonnet lib/examples/full-stack.jsonnet
    exit 1
fi

echo -n "  - minimal.jsonnet: "
if jsonnet lib/examples/minimal.jsonnet > /dev/null 2>&1; then
    echo "✓"
    resources=$(jsonnet lib/examples/minimal.jsonnet | jq -r '[.. | objects | select(.kind) | .kind] | unique | length')
    echo "    Generated ${resources} unique Kubernetes resource types"
else
    echo "✗ FAILED"
    jsonnet lib/examples/minimal.jsonnet
    exit 1
fi
echo

# Test all environments
echo "Testing environment configurations..."
for env in dev staging production; do
    echo -n "  - ${env} environment: "
    if jsonnet -e "(import 'lib/config.libsonnet').environments.${env}" > /dev/null 2>&1; then
        echo "✓"
    else
        echo "✗ FAILED"
        jsonnet -e "(import 'lib/config.libsonnet').environments.${env}"
        exit 1
    fi
done
echo

# Count total resources
echo "Resource Summary:"
echo "  Full Stack:"
jsonnet lib/examples/full-stack.jsonnet | jq -r '
  [.. | objects | select(.kind)] |
  group_by(.kind) |
  map({kind: .[0].kind, count: length}) |
  .[] |
  "    - \(.kind): \(.count)"
'

echo
echo "  Minimal Stack:"
jsonnet lib/examples/minimal.jsonnet | jq -r '
  [.. | objects | select(.kind)] |
  group_by(.kind) |
  map({kind: .[0].kind, count: length}) |
  .[] |
  "    - \(.kind): \(.count)"
'

echo
echo "=== All Validations Passed ✓ ==="
