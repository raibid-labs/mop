# Workstream 4: Tanka Configuration

## Status
🔴 Not Started

## Overview
Establish a comprehensive Tanka configuration framework for managing Kubernetes resources using Jsonnet. This includes project initialization, library management, Helm integration patterns, environment-specific configurations (dev/staging/prod), vendor management, and CI/CD integration for automated deployments.

## Objectives
- [ ] Initialize Tanka project structure with best practices
- [ ] Configure jsonnet-bundler for library dependency management
- [ ] Integrate Helm charts with Tanka using jsonnet-helm
- [ ] Create environment-specific configurations (dev, staging, prod)
- [ ] Implement reusable Jsonnet libraries for common patterns
- [ ] Set up vendor directory management and updates
- [ ] Document Tanka workflows and conventions

## Agent Assignment
**Suggested Agent Type**: `backend-dev`, `system-architect`, `reviewer`
**Skill Requirements**: Jsonnet language, Tanka, Kubernetes YAML, Helm, declarative configuration, GitOps practices

## Dependencies
- Workstream 1 must complete Tanka installation and base setup
- Git repository initialized for version control
- Kubernetes cluster access for validation
- Basic understanding of OBI, Tempo, Mimir, Loki deployment patterns

## Tasks

### Task 4.1: Tanka Project Structure
**Description**: Initialize Tanka project with proper directory structure, configuration files, and base environments.

**Deliverables**:
- Tanka project initialized with `tk init`
- Directory structure for environments, libraries, and vendors
- Base configuration in `tkrc.yaml`
- Git ignore patterns for generated files
- Documentation for project layout

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tanka/tkrc.yaml`
- `/Users/beengud/raibid-labs/mop/tanka/.gitignore`
- `/Users/beengud/raibid-labs/mop/tanka/README.md`
- `/Users/beengud/raibid-labs/mop/tanka/lib/.gitkeep`
- `/Users/beengud/raibid-labs/mop/tanka/vendor/.gitkeep`
- `/Users/beengud/raibid-labs/mop/docs/tanka-project-structure.md`

**Validation**:
```bash
# Initialize Tanka project
cd /Users/beengud/raibid-labs/mop/tanka
tk init

# Verify project structure
tree -L 2 /Users/beengud/raibid-labs/mop/tanka

# Check Tanka configuration
cat tkrc.yaml

# Validate Tanka environment
tk env list

# Test basic Jsonnet evaluation
echo '{ hello: "world" }' > test.jsonnet
jsonnet test.jsonnet
rm test.jsonnet
```

### Task 4.2: Jsonnet Library Management
**Description**: Set up jsonnet-bundler for managing external libraries and create custom libraries for MOP components.

**Deliverables**:
- `jsonnetfile.json` with dependencies
- Custom Jsonnet libraries for OBI, Grafana stack
- Helper functions for common Kubernetes patterns
- Library documentation and examples
- Vendor directory with pinned versions

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tanka/jsonnetfile.json`
- `/Users/beengud/raibid-labs/mop/tanka/jsonnetfile.lock.json`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/obi.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/grafana.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/k8s-helpers.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/config.libsonnet`
- `/Users/beengud/raibid-labs/mop/docs/jsonnet-libraries.md`

**Validation**:
```bash
# Install jsonnet-bundler dependencies
cd /Users/beengud/raibid-labs/mop/tanka
jb install

# Update dependencies
jb update

# List installed packages
jb list

# Verify vendor directory
ls -la vendor/

# Test library imports
jsonnet -J vendor -J lib -e 'local k = import "k.libsonnet"; k'
jsonnet -J vendor -J lib -e 'local obi = import "mop/obi.libsonnet"; obi'

# Check for import errors
find lib -name "*.libsonnet" -exec jsonnet -J vendor -J lib {} \; > /dev/null
```

### Task 4.3: Helm Integration
**Description**: Integrate Helm charts with Tanka using `tanka-util/helm` for managing Grafana stack deployments.

**Deliverables**:
- Helm chart integration for Tempo
- Helm chart integration for Mimir
- Helm chart integration for Loki
- Helm chart integration for Grafana
- Value overrides in Jsonnet
- Helm repository configuration

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/helm/tempo.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/helm/mimir.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/helm/loki.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/helm/grafana.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/helm/common.libsonnet`
- `/Users/beengud/raibid-labs/mop/docs/tanka-helm-integration.md`

**Validation**:
```bash
# Add Grafana Helm repository
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

# Test Helm chart templating with Tanka
cd /Users/beengud/raibid-labs/mop/tanka
tk show environments/dev --dangerous-allow-redirect | grep -A 10 "kind: Deployment"

# Verify Helm values are applied
tk show environments/dev --dangerous-allow-redirect | grep -A 5 "image:"

# Validate generated YAML
tk show environments/dev --dangerous-allow-redirect | kubectl apply --dry-run=client -f -

# Check for Helm chart versions
helm search repo grafana/tempo --versions | head -5
```

### Task 4.4: Environment Configuration
**Description**: Create environment-specific configurations for dev, staging, and production with appropriate resource limits and settings.

**Deliverables**:
- Dev environment with minimal resources
- Staging environment mirroring production
- Production environment with HA configuration
- Environment-specific secrets management
- Namespace mapping per environment
- Environment promotion strategy

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tanka/environments/dev/main.jsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/environments/dev/spec.json`
- `/Users/beengud/raibid-labs/mop/tanka/environments/staging/main.jsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/environments/staging/spec.json`
- `/Users/beengud/raibid-labs/mop/tanka/environments/prod/main.jsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/environments/prod/spec.json`
- `/Users/beengud/raibid-labs/mop/docs/environment-management.md`

**Validation**:
```bash
# List all environments
tk env list

# Show dev environment
tk show environments/dev

# Show staging environment
tk show environments/staging

# Show production environment
tk show environments/prod

# Compare environments
diff <(tk show environments/dev) <(tk show environments/staging)

# Validate environment specs
cat environments/dev/spec.json | jq
cat environments/staging/spec.json | jq
cat environments/prod/spec.json | jq

# Check for environment-specific differences
tk diff environments/dev
tk diff environments/staging
```

### Task 4.5: Reusable Jsonnet Patterns
**Description**: Create reusable Jsonnet libraries for common Kubernetes patterns like DaemonSets, ConfigMaps, Secrets, and RBAC.

**Deliverables**:
- DaemonSet generator function
- ConfigMap and Secret helpers
- RBAC resource generators
- Service and Ingress templates
- Resource limit/request patterns
- Label and annotation standards

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/patterns/daemonset.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/patterns/configmap.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/patterns/secret.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/patterns/rbac.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/patterns/service.libsonnet`
- `/Users/beengud/raibid-labs/mop/tanka/lib/mop/patterns/ingress.libsonnet`
- `/Users/beengud/raibid-labs/mop/docs/jsonnet-patterns.md`

**Validation**:
```bash
# Test DaemonSet generator
cd /Users/beengud/raibid-labs/mop/tanka
jsonnet -J vendor -J lib -e 'local ds = import "mop/patterns/daemonset.libsonnet"; ds.new("test", "nginx:latest")'

# Test ConfigMap generator
jsonnet -J vendor -J lib -e 'local cm = import "mop/patterns/configmap.libsonnet"; cm.new("test", {key: "value"})'

# Test RBAC generator
jsonnet -J vendor -J lib -e 'local rbac = import "mop/patterns/rbac.libsonnet"; rbac.serviceAccount("test")'

# Validate all patterns compile
find lib/mop/patterns -name "*.libsonnet" -exec jsonnet -J vendor -J lib {} \;

# Check for common errors
jsonnet-lint lib/mop/patterns/*.libsonnet
```

### Task 4.6: Vendor Management
**Description**: Establish vendor management practices for updating and maintaining external Jsonnet libraries.

**Deliverables**:
- Vendor update automation script
- Version pinning strategy
- Security scanning for dependencies
- Changelog tracking for updates
- Rollback procedures
- Vendor audit documentation

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/scripts/update-vendors.sh`
- `/Users/beengud/raibid-labs/mop/scripts/audit-vendors.sh`
- `/Users/beengud/raibid-labs/mop/tanka/.jb-pin`
- `/Users/beengud/raibid-labs/mop/docs/vendor-management.md`
- `/Users/beengud/raibid-labs/mop/VENDOR_CHANGELOG.md`

**Validation**:
```bash
# Update vendors
cd /Users/beengud/raibid-labs/mop
./scripts/update-vendors.sh

# Audit vendors
./scripts/audit-vendors.sh

# Check for outdated packages
cd tanka
jb list | while read pkg; do
  echo "Checking $pkg..."
  # Compare with latest version
done

# Verify lock file
cat jsonnetfile.lock.json | jq

# Test after vendor update
tk show environments/dev > /dev/null && echo "Vendor update successful"
```

### Task 4.7: CI/CD Integration
**Description**: Integrate Tanka with CI/CD pipelines for automated validation, diff generation, and deployment.

**Deliverables**:
- GitHub Actions workflow for Tanka validation
- Automated diff generation on PRs
- Deployment automation for environments
- Drift detection and alerting
- Rollback automation
- Deployment status reporting

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/.github/workflows/tanka-validate.yml`
- `/Users/beengud/raibid-labs/mop/.github/workflows/tanka-deploy.yml`
- `/Users/beengud/raibid-labs/mop/.github/workflows/tanka-diff.yml`
- `/Users/beengud/raibid-labs/mop/scripts/tanka-deploy.sh`
- `/Users/beengud/raibid-labs/mop/scripts/tanka-drift-detect.sh`
- `/Users/beengud/raibid-labs/mop/docs/tanka-cicd.md`

**Validation**:
```bash
# Validate Tanka locally (simulates CI)
cd /Users/beengud/raibid-labs/mop/tanka
tk show environments/dev > /dev/null && echo "✓ Dev environment valid"
tk show environments/staging > /dev/null && echo "✓ Staging environment valid"
tk show environments/prod > /dev/null && echo "✓ Prod environment valid"

# Generate diff
tk diff environments/dev

# Dry-run deployment
tk apply environments/dev --dry-run

# Check GitHub Actions syntax
cd /Users/beengud/raibid-labs/mop
actionlint .github/workflows/tanka-*.yml || echo "actionlint not installed, skipping"

# Test deployment script
./scripts/tanka-deploy.sh dev --dry-run
```

## Definition of Done
- [ ] Tanka project initialized with proper structure
- [ ] Jsonnet libraries installed and documented
- [ ] Helm charts integrated with Tanka
- [ ] All three environments (dev, staging, prod) configured
- [ ] Reusable Jsonnet patterns created and tested
- [ ] Vendor management automation implemented
- [ ] CI/CD workflows created and tested
- [ ] All environments validate successfully with `tk show`
- [ ] Diff generation working correctly
- [ ] Documentation complete with examples
- [ ] Code reviewed by at least one team member

## Agent Coordination Hooks
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-4-tanka-configuration"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-4"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/tanka/jsonnetfile.json" --memory-key "swarm/mop/ws-4/jsonnet-deps"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/tanka/lib/mop/obi.libsonnet" --memory-key "swarm/mop/ws-4/obi-lib"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/tanka/environments/dev/main.jsonnet" --memory-key "swarm/mop/ws-4/dev-env"
npx claude-flow@alpha hooks notify --message "Tanka configuration completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-4-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

## Estimated Effort
**Duration**: 5-7 days
**Complexity**: Medium-High

## References
- [Tanka Documentation](https://tanka.dev/)
- [Jsonnet Language Reference](https://jsonnet.org/ref/language.html)
- [jsonnet-bundler Documentation](https://github.com/jsonnet-bundler/jsonnet-bundler)
- [Kubernetes API Reference](https://kubernetes.io/docs/reference/kubernetes-api/)
- [Grafana Tanka Examples](https://github.com/grafana/tanka/tree/main/examples)
- [Tanka Best Practices](https://tanka.dev/tutorial/abstraction)

## Notes
- Jsonnet can be difficult to debug - use `jsonnet fmt` and `jsonnet-lint` regularly
- Keep Jsonnet functions pure and side-effect free for predictability
- Use `local` variables extensively to avoid recomputation
- Tanka's `tk show` is invaluable for debugging generated YAML
- Consider using `tk prune` to remove orphaned resources
- Jsonnet evaluation can be slow for large configs - optimize imports
- Use `--dangerous-allow-redirect` carefully in CI/CD (validates against cluster)
- Pin vendor versions in `jsonnetfile.lock.json` for reproducibility
- Helm integration via Tanka is powerful but adds complexity
- Test Jsonnet changes locally before pushing to CI
- Use `tk env set` to update environment API server addresses
- Consider using `tk export` for GitOps workflows (e.g., ArgoCD)
- Document Jsonnet idioms and patterns for team consistency
- Tanka's diff output is more readable than `kubectl diff`
- Use `tk tool charts` to manage Helm chart versions
