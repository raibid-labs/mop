# Workstream 5: Development Tools

## Status
🔴 Not Started

## Overview
Create a comprehensive developer experience toolkit for the MOP platform, including Tiltfile for local development with hot reloading, justfile for common commands, nushell automation scripts, testing frameworks for validation, and CI/CD integration for automated workflows. This workstream focuses on developer productivity and operational excellence.

## Objectives
- [ ] Implement Tiltfile for local Kubernetes development with live reload
- [ ] Create justfile with common operational commands
- [ ] Develop nushell automation scripts for complex workflows
- [ ] Set up testing framework for infrastructure validation
- [ ] Integrate with CI/CD pipelines (GitHub Actions)
- [ ] Implement pre-commit hooks for code quality
- [ ] Create developer documentation and onboarding guides

## Agent Assignment
**Suggested Agent Type**: `backend-dev`, `cicd-engineer`, `reviewer`
**Skill Requirements**: Tilt, Make/Just, shell scripting, Nushell, CI/CD, testing frameworks, developer experience design

## Dependencies
- Workstream 1 must complete Kubernetes setup
- Workstream 4 must complete Tanka configuration
- Docker and Tilt installed locally
- Local Kubernetes cluster (kind, k3d, or minikube)
- Git repository with branch protection rules

## Tasks

### Task 5.1: Tiltfile Development
**Description**: Create a comprehensive Tiltfile for local development with hot reloading, resource visualization, and debugging capabilities.

**Deliverables**:
- Tiltfile with resource definitions for all MOP components
- Docker build configurations with caching
- Live reload for Jsonnet changes
- Port forwarding for local access
- Resource grouping and filtering
- Performance optimization
- Debug mode configuration

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/Tiltfile`
- `/Users/beengud/raibid-labs/mop/tilt_config.json`
- `/Users/beengud/raibid-labs/mop/tilt/extensions.star`
- `/Users/beengud/raibid-labs/mop/tilt/helpers.star`
- `/Users/beengud/raibid-labs/mop/docs/local-development.md`

**Validation**:
```bash
# Start Tilt
cd /Users/beengud/raibid-labs/mop
tilt up

# Verify resources are healthy
tilt get uiresource

# Check logs
tilt logs obi
tilt logs tempo
tilt logs mimir

# Test hot reload (modify Jsonnet file and observe rebuild)
touch tanka/lib/mop/obi.libsonnet

# Access Tilt UI
open http://localhost:10350

# Tear down
tilt down
```

### Task 5.2: Justfile Creation
**Description**: Develop a justfile (modern Make alternative) with common commands for building, testing, deploying, and managing the MOP platform.

**Deliverables**:
- Justfile with categorized recipes
- Build and test commands
- Deployment commands (dev, staging, prod)
- Utility commands (cleanup, logs, shell)
- Documentation generation commands
- Dependency checking commands

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/justfile`
- `/Users/beengud/raibid-labs/mop/docs/justfile-commands.md`

**Validation**:
```bash
# Install just (if not installed)
which just || brew install just

# List all recipes
cd /Users/beengud/raibid-labs/mop
just --list

# Test build commands
just build-obi
just build-all

# Test deployment commands
just deploy-dev
just diff-dev

# Test utility commands
just logs obi
just shell obi
just cleanup-dev

# Run tests
just test
just test-integration

# Generate documentation
just docs
```

### Task 5.3: Nushell Automation Scripts
**Description**: Create Nushell scripts for complex automation tasks like multi-environment deployment, batch operations, and advanced reporting.

**Deliverables**:
- Multi-environment deployment script
- Resource usage analysis script
- Log aggregation and analysis script
- Performance benchmarking script
- Backup and restore automation
- Chaos testing automation

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/scripts/nu/deploy-multi-env.nu`
- `/Users/beengud/raibid-labs/mop/scripts/nu/analyze-resources.nu`
- `/Users/beengud/raibid-labs/mop/scripts/nu/analyze-logs.nu`
- `/Users/beengud/raibid-labs/mop/scripts/nu/benchmark.nu`
- `/Users/beengud/raibid-labs/mop/scripts/nu/backup-restore.nu`
- `/Users/beengud/raibid-labs/mop/scripts/nu/chaos-test.nu`
- `/Users/beengud/raibid-labs/mop/scripts/nu/lib/common.nu`
- `/Users/beengud/raibid-labs/mop/docs/nushell-scripts.md`

**Validation**:
```bash
# Install nushell (if not installed)
which nu || brew install nushell

# Test multi-environment deployment
cd /Users/beengud/raibid-labs/mop
nu scripts/nu/deploy-multi-env.nu --envs [dev staging] --dry-run

# Test resource analysis
nu scripts/nu/analyze-resources.nu --namespace mop-system

# Test log analysis
nu scripts/nu/analyze-logs.nu --component obi --since 1h

# Test benchmarking
nu scripts/nu/benchmark.nu --duration 60s --concurrency 10

# Test backup
nu scripts/nu/backup-restore.nu backup --output /tmp/mop-backup.tar.gz

# Test chaos testing
nu scripts/nu/chaos-test.nu --target obi --duration 5m --dry-run
```

### Task 5.4: Testing Framework
**Description**: Implement a comprehensive testing framework for infrastructure, integration, and end-to-end testing.

**Deliverables**:
- Unit tests for Jsonnet libraries
- Integration tests for component interactions
- End-to-end tests for full stack validation
- Performance tests for load testing
- Chaos tests for resilience testing
- Test reporting and coverage

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/tests/unit/jsonnet_test.jsonnet`
- `/Users/beengud/raibid-labs/mop/tests/integration/obi_test.sh`
- `/Users/beengud/raibid-labs/mop/tests/integration/grafana_stack_test.sh`
- `/Users/beengud/raibid-labs/mop/tests/e2e/full_stack_test.sh`
- `/Users/beengud/raibid-labs/mop/tests/performance/load_test.py`
- `/Users/beengud/raibid-labs/mop/tests/chaos/obi_failure_test.sh`
- `/Users/beengud/raibid-labs/mop/tests/conftest.py`
- `/Users/beengud/raibid-labs/mop/pytest.ini`

**Validation**:
```bash
# Run unit tests
cd /Users/beengud/raibid-labs/mop
jsonnet tests/unit/jsonnet_test.jsonnet

# Run integration tests
./tests/integration/obi_test.sh
./tests/integration/grafana_stack_test.sh

# Run end-to-end tests
./tests/e2e/full_stack_test.sh

# Run performance tests
pytest tests/performance/ -v

# Run chaos tests
./tests/chaos/obi_failure_test.sh

# Generate coverage report
pytest --cov=tests --cov-report=html
open htmlcov/index.html
```

### Task 5.5: CI/CD Pipeline Configuration
**Description**: Set up GitHub Actions workflows for continuous integration, testing, and deployment automation.

**Deliverables**:
- CI workflow for PR validation
- CD workflow for automated deployments
- Scheduled workflows for drift detection
- Security scanning workflows
- Documentation generation workflow
- Notification integration (Slack, email)

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/.github/workflows/ci.yml`
- `/Users/beengud/raibid-labs/mop/.github/workflows/cd.yml`
- `/Users/beengud/raibid-labs/mop/.github/workflows/drift-detection.yml`
- `/Users/beengud/raibid-labs/mop/.github/workflows/security-scan.yml`
- `/Users/beengud/raibid-labs/mop/.github/workflows/docs.yml`
- `/Users/beengud/raibid-labs/mop/docs/cicd-workflows.md`

**Validation**:
```bash
# Validate GitHub Actions syntax
cd /Users/beengud/raibid-labs/mop
actionlint .github/workflows/*.yml || echo "actionlint not installed"

# Test CI workflow locally with act
act pull_request -j test

# Trigger workflow manually (requires gh CLI)
gh workflow run ci.yml

# Check workflow status
gh run list --workflow=ci.yml

# View workflow logs
gh run view --log

# Test deployment workflow in dry-run
gh workflow run cd.yml -f environment=dev -f dry_run=true
```

### Task 5.6: Pre-commit Hooks
**Description**: Implement pre-commit hooks for code quality, linting, security scanning, and automated formatting.

**Deliverables**:
- Pre-commit configuration file
- Jsonnet linting and formatting
- YAML validation
- Shell script linting (shellcheck)
- Security scanning (trivy, gitleaks)
- Custom hooks for MOP-specific checks

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/.pre-commit-config.yaml`
- `/Users/beengud/raibid-labs/mop/.shellcheckrc`
- `/Users/beengud/raibid-labs/mop/scripts/pre-commit/validate-tanka.sh`
- `/Users/beengud/raibid-labs/mop/scripts/pre-commit/check-secrets.sh`
- `/Users/beengud/raibid-labs/mop/docs/pre-commit-hooks.md`

**Validation**:
```bash
# Install pre-commit
pip install pre-commit

# Install hooks
cd /Users/beengud/raibid-labs/mop
pre-commit install

# Run all hooks manually
pre-commit run --all-files

# Test specific hook
pre-commit run jsonnet-fmt --all-files

# Test pre-commit on staged changes
echo "test change" >> README.md
git add README.md
git commit -m "test commit"

# Verify hooks ran
git log -1 --stat
```

### Task 5.7: Developer Documentation
**Description**: Create comprehensive developer documentation, onboarding guides, and troubleshooting resources.

**Deliverables**:
- Developer setup guide
- Local development workflow documentation
- Troubleshooting guide
- Architecture decision records (ADRs)
- Contributing guidelines
- Code style guide

**Files to Create/Modify**:
- `/Users/beengud/raibid-labs/mop/docs/DEVELOPER_SETUP.md`
- `/Users/beengud/raibid-labs/mop/docs/LOCAL_WORKFLOW.md`
- `/Users/beengud/raibid-labs/mop/docs/TROUBLESHOOTING.md`
- `/Users/beengud/raibid-labs/mop/docs/adr/001-tanka-for-k8s-management.md`
- `/Users/beengud/raibid-labs/mop/docs/adr/002-obi-for-observability.md`
- `/Users/beengud/raibid-labs/mop/CONTRIBUTING.md`
- `/Users/beengud/raibid-labs/mop/docs/CODE_STYLE.md`

**Validation**:
```bash
# Validate markdown
cd /Users/beengud/raibid-labs/mop
markdownlint docs/*.md

# Check for broken links
markdown-link-check docs/*.md

# Verify documentation completeness
grep -r "TODO\|FIXME" docs/ || echo "No TODOs found"

# Test setup guide by following steps in new environment
# (manual validation)

# Generate documentation site
mkdocs build
mkdocs serve
open http://127.0.0.1:8000
```

## Definition of Done
- [ ] Tiltfile working with hot reload for all components
- [ ] Justfile with all common commands documented
- [ ] Nushell scripts tested and functional
- [ ] All test categories implemented (unit, integration, e2e, performance, chaos)
- [ ] CI/CD workflows passing and deploying correctly
- [ ] Pre-commit hooks installed and enforcing standards
- [ ] Developer documentation complete and reviewed
- [ ] Onboarding guide tested with new developer
- [ ] All automation scripts have error handling
- [ ] Notifications configured for CI/CD failures
- [ ] Code reviewed by at least one team member

## Agent Coordination Hooks
```bash
# BEFORE Work:
npx claude-flow@alpha hooks pre-task --description "workstream-5-development-tools"
npx claude-flow@alpha hooks session-restore --session-id "swarm-mop-ws-5"

# DURING Work:
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/Tiltfile" --memory-key "swarm/mop/ws-5/tiltfile"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/justfile" --memory-key "swarm/mop/ws-5/justfile"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/scripts/nu/deploy-multi-env.nu" --memory-key "swarm/mop/ws-5/nushell-scripts"
npx claude-flow@alpha hooks post-edit --file "/Users/beengud/raibid-labs/mop/.github/workflows/ci.yml" --memory-key "swarm/mop/ws-5/cicd"
npx claude-flow@alpha hooks notify --message "Development tools setup completed"

# AFTER Work:
npx claude-flow@alpha hooks post-task --task-id "ws-5-complete"
npx claude-flow@alpha hooks session-end --export-metrics true
```

## Estimated Effort
**Duration**: 6-8 days
**Complexity**: Medium-High

## References
- [Tilt Documentation](https://docs.tilt.dev/)
- [Just Documentation](https://github.com/casey/just)
- [Nushell Documentation](https://www.nushell.sh/)
- [GitHub Actions Documentation](https://docs.github.com/en/actions)
- [Pre-commit Documentation](https://pre-commit.com/)
- [Pytest Documentation](https://docs.pytest.org/)
- [Architecture Decision Records](https://adr.github.io/)

## Notes
- Tilt requires Docker and Kubernetes to be running locally
- Just is more user-friendly than Make but less widely adopted
- Nushell provides structured data pipelines unlike bash
- Consider using Taskfile.yml as alternative to justfile
- Pre-commit hooks should be fast (<30s) to not slow down commits
- CI/CD workflows should fail fast to provide quick feedback
- Use matrix builds in GitHub Actions for testing multiple environments
- Document all just recipes with descriptions
- Tiltfile can become complex - modularize with load() and include()
- Consider using act for local GitHub Actions testing
- Chaos tests should be opt-in to prevent accidental destruction
- Performance tests may require dedicated infrastructure
- Developer documentation should be kept up-to-date in CI
- Consider using devcontainers for consistent development environments
- Tilt's live_update is powerful but requires careful configuration
- Use just --set for parameterized recipes
