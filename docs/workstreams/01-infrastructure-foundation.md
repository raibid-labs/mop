# Workstream 1: Infrastructure Foundation

## Status
🟢 Completed

## Overview
Establish the foundational Kubernetes infrastructure for the MOP (Metrics, Observability, Practices) platform. This includes setting up namespaces, installing base dependencies (Tanka, jsonnet-bundler, Helm), configuring storage classes, and establishing RBAC policies for secure multi-tenant operations.

## Objectives
- [ ] Create and configure Kubernetes namespaces with proper isolation
- [ ] Install and configure Tanka, jsonnet-bundler, and Helm tooling
- [ ] Set up storage classes for persistent volume claims
- [ ] Implement RBAC policies and service accounts
- [ ] Establish network policies for namespace isolation
- [ ] Configure resource quotas and limit ranges

## Agent Assignment
**Suggested Agent Type**: `backend-dev`, `system-architect`
**Skill Requirements**: Kubernetes administration, YAML/Jsonnet, RBAC design, storage architecture

## Dependencies
- Kubernetes cluster must be accessible (v1.24+)
- kubectl configured with cluster admin access
- Local development environment with Tanka, jb, and Helm installed

## Tasks

### Task 1.1: Namespace Creation and Configuration
**Description**: Create dedicated namespaces for observability components with proper labels and annotations.

**Deliverables**:
- Namespace manifests for `mop-system`, `mop-traces`, `mop-metrics`, `mop-logs`
- Resource quotas and limit ranges
- Network policies for inter-namespace communication

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/k8s/base/namespaces/mop-system.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/namespaces/mop-traces.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/namespaces/mop-metrics.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/namespaces/mop-logs.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/network-policies/isolation.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/resource-quotas/quotas.yaml`

**Validation**:
```bash
# Verify namespaces exist
kubectl get namespaces | grep mop-

# Check resource quotas
kubectl get resourcequota -n mop-system
kubectl describe quota -n mop-system

# Verify network policies
kubectl get networkpolicies -n mop-system
kubectl describe networkpolicy -n mop-system
```

### Task 1.2: Tanka and Jsonnet Setup
**Description**: Initialize Tanka project structure and configure jsonnet-bundler for library management.

**Deliverables**:
- Tanka project initialized with environments
- Base jsonnet libraries installed
- Vendor management configured
- Documentation for local development

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tanka/jsonnetfile.json`
- `/Users/beengud/raibid-labs/mop/tanka/jsonnetfile.lock.json`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/config.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/environments/default/main.jsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/environments/default/spec.json`

**Validation**:
```bash
# Verify Tanka installation
tk --version

# Check jsonnet-bundler
jb --version

# Validate Tanka project
cd /Users/beengud/raibid-labs/mop/tanka
tk show environments/default

# Verify vendor libraries
ls -la vendor/
```

### Task 1.3: Storage Class Configuration
**Description**: Define and deploy storage classes for persistent volumes needed by observability components.

**Deliverables**:
- StorageClass definitions for fast SSD and standard storage
- PersistentVolumeClaim templates
- Volume snapshot classes
- Storage capacity documentation

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/k8s/base/storage/storage-class-fast.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/storage/storage-class-standard.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/storage/volume-snapshot-class.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/storage/pvc-templates.yaml`

**Validation**:
```bash
# List storage classes
kubectl get storageclass

# Check default storage class
kubectl get storageclass -o=jsonpath='{.items[?(@.metadata.annotations.storageclass\.kubernetes\.io/is-default-class=="true")].metadata.name}'

# Test PVC creation
kubectl apply -f /Users/beengud/raibid-labs/mop/k8s/base/storage/pvc-templates.yaml -n mop-system
kubectl get pvc -n mop-system
kubectl delete -f /Users/beengud/raibid-labs/mop/k8s/base/storage/pvc-templates.yaml -n mop-system
```

### Task 1.4: RBAC Configuration
**Description**: Implement Role-Based Access Control with service accounts, roles, and bindings for secure operations.

**Deliverables**:
- Service accounts for each observability component
- Cluster roles and role bindings
- Namespace-scoped roles
- Security context constraints
- RBAC audit documentation

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/k8s/base/rbac/service-accounts.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/rbac/cluster-roles.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/rbac/role-bindings.yaml`
- `/Users/beengud/raibid-labs/mop/k8s/base/rbac/namespace-roles.yaml`
- `/Users/beengud/raibid-labs/mop/docs/rbac-policy.md`

**Validation**:
```bash
# List service accounts
kubectl get serviceaccounts -n mop-system

# Check cluster roles
kubectl get clusterroles | grep mop-

# Verify role bindings
kubectl get rolebindings -n mop-system
kubectl get clusterrolebindings | grep mop-

# Test RBAC permissions
kubectl auth can-i list pods --as=system:serviceaccount:mop-system:obi-collector -n mop-system
```

### Task 1.5: Helm Repository Configuration
**Description**: Configure Helm repositories for Grafana stack and other third-party charts.

**Deliverables**:
- Helm repository list
- Repository credentials (if private)
- Chart version pinning strategy
- Helm values templates

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/helm/repositories.yaml`
- `/Users/beengud/raibid-labs/mop/helm/Chart.yaml`
- `/Users/beengud/raibid-labs/mop/helm/values.yaml`
- `/Users/beengud/raibid-labs/mop/docs/helm-workflow.md`

**Validation**:
```bash
# Add Grafana Helm repo
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Search for charts
helm search repo grafana/tempo
helm search repo grafana/mimir
helm search repo grafana/loki

# List repositories
helm repo list
```

### Task 1.6: Base Infrastructure Testing
**Description**: Create automated tests to validate infrastructure foundation before proceeding with component deployment.

**Deliverables**:
- Infrastructure validation script
- Namespace connectivity tests
- Storage provisioning tests
- RBAC validation tests
- CI integration for infrastructure checks

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tests/infrastructure/validate-namespaces.sh`
- `/Users/beengud/raibid-labs/mop/tests/infrastructure/validate-storage.sh`
- `/Users/beengud/raibid-labs/mop/tests/infrastructure/validate-rbac.sh`
- `/Users/beengud/raibid-labs/mop/tests/infrastructure/validate-network.sh`
- `/Users/beengud/raibid-labs/mop/.github/workflows/infrastructure-tests.yml`

**Validation**:
```bash
# Run all infrastructure tests
cd /Users/beengud/raibid-labs/mop/tests/infrastructure
./validate-namespaces.sh
./validate-storage.sh
./validate-rbac.sh
./validate-network.sh

# Check exit codes
echo "All tests passed: $?"
```

## Definition of Done
- [ ] All namespaces created and labeled correctly
- [ ] Tanka project initialized with working environments
- [ ] Storage classes deployed and tested
- [ ] RBAC policies implemented and validated
- [ ] Helm repositories configured
- [ ] Network policies enforcing namespace isolation
- [ ] Resource quotas and limits applied
- [ ] All infrastructure validation tests passing
- [ ] Documentation complete with architecture diagrams
- [ ] Code reviewed by at least one team member

## Agent Coordination Hooks
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-1-infrastructure-foundation"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-1"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/k8s/base/namespaces/mop-system.yaml" --memory-key "swarm/mop/ws-1/namespace-config"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/tanka/jsonnetfile.json" --memory-key "swarm/mop/ws-1/tanka-setup"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/k8s/base/rbac/service-accounts.yaml" --memory-key "swarm/mop/ws-1/rbac-config"
npx claude-flow@alpha hooks notify --message "Infrastructure foundation tasks completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-1-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

## Estimated Effort
**Duration**: 3-5 days
**Complexity**: Medium

## References
- [Kubernetes RBAC Documentation](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Tanka Documentation](https://tanka.dev/)
- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes Storage Classes](https://kubernetes.io/docs/concepts/storage/storage-classes/)
- [Network Policies](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

## Notes
- Ensure kubectl context is set to the correct cluster before running any commands
- Storage class names may vary by cloud provider (e.g., gp3 on AWS, pd-ssd on GCP)
- Consider using kustomize alongside Tanka for base/overlay pattern
- RBAC policies should follow principle of least privilege
- Document any cluster-specific configuration requirements
- Consider implementing admission controllers for policy enforcement
- Plan for disaster recovery and backup strategies for persistent volumes
- Namespace naming convention: `mop-<component>` for consistency
