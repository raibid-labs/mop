# Agent Coordination Guide

## Overview

This guide defines the coordination patterns for multi-agent development in the MOP (Multi-cluster Observability Platform) project. Based on raibid-labs patterns, this system enables parallel workstreams with hook-based coordination and clear ownership boundaries.

## Core Coordination Principles

### 1. Directory Ownership Model

Each agent or team owns specific directories and files, preventing conflicts during parallel execution:

```
Platform Team:
├── environments/*/infrastructure/    # Platform engineers
├── environments/*/kubernetes/        # Kubernetes specialists
├── lib/tanka/                       # Tanka library maintainers
└── scripts/infra/                   # Infrastructure automation

Observability Team:
├── environments/*/observability/    # OBI, Grafana, Tempo, Mimir, Loki
├── charts/obi-*                     # OBI Helm charts
├── lib/alloy/                       # Alloy pipeline library
└── experiments/                     # OBI experiments

DevOps Team:
├── Tiltfile                         # Local development
├── justfile                         # Task automation
├── .github/workflows/               # CI/CD pipelines
└── scripts/automation/              # Build and deploy scripts
```

### 2. Agent State Machine

All agents follow a standard lifecycle:

```
AVAILABLE → ASSIGNED → ACTIVE → COMPLETE
    ↓          ↓          ↓          ↓
  Ready    Task Given  Working   Delivered
```

**State Transitions:**
- `AVAILABLE`: Agent registered and ready for tasks
- `ASSIGNED`: Task allocated, preparing to execute
- `ACTIVE`: Actively working on assigned task
- `COMPLETE`: Task finished, output delivered

### 3. Hook-Based Coordination Protocol

Every agent MUST execute coordination hooks at specific points:

#### Pre-Task Hook
```bash
# Execute BEFORE starting any work
npx claude-flow@alpha hooks pre-task \
  --description "Implement OBI experiment configuration" \
  --agent-id "obi-specialist-001" \
  --session-id "swarm-mop-platform"
```

**Purpose:**
- Validates agent readiness
- Checks for dependency completion
- Locks resources/files
- Restores previous context

#### Post-Edit Hook
```bash
# Execute AFTER each significant file change
npx claude-flow@alpha hooks post-edit \
  --file "environments/prod/observability/obi-config.yaml" \
  --memory-key "swarm/obi/prod-config" \
  --agent-id "obi-specialist-001"
```

**Purpose:**
- Notifies other agents of changes
- Updates shared memory
- Trains neural patterns
- Triggers dependent tasks

#### Post-Task Hook
```bash
# Execute AFTER completing task
npx claude-flow@alpha hooks post-task \
  --task-id "implement-obi-experiment" \
  --status "complete" \
  --output-files "environments/prod/observability/experiments/latency-test.yaml" \
  --agent-id "obi-specialist-001"
```

**Purpose:**
- Marks task complete
- Releases resources
- Updates metrics
- Generates summary

### 4. Session Management

#### Session Start
```bash
npx claude-flow@alpha hooks session-start \
  --session-id "swarm-mop-platform" \
  --topology "mesh" \
  --max-agents 12
```

#### Session Restore
```bash
# Restore context from previous session
npx claude-flow@alpha hooks session-restore \
  --session-id "swarm-mop-platform" \
  --restore-memory true
```

#### Session End
```bash
npx claude-flow@alpha hooks session-end \
  --session-id "swarm-mop-platform" \
  --export-metrics true \
  --save-state true
```

## Agent Roles and Responsibilities

### Platform Engineers
**Primary Focus:** Infrastructure, Kubernetes, cluster management

**Responsibilities:**
- Kubernetes cluster configuration
- Infrastructure as code (Tanka/Jsonnet)
- Network policies and service meshes
- Resource quotas and limits
- Cluster upgrades and maintenance

**File Ownership:**
- `environments/*/kubernetes/`
- `lib/tanka/`
- `scripts/cluster-management/`

### OBI Specialists
**Primary Focus:** Observability Backend Interface configuration

**Responsibilities:**
- OBI experiment design and implementation
- eBPF probe configuration
- Metrics collection setup
- Experiment scheduling and lifecycle
- Performance tuning

**File Ownership:**
- `charts/obi-*/`
- `environments/*/observability/obi-*.yaml`
- `experiments/`

### Grafana Specialists
**Primary Focus:** Grafana stack (Tempo, Mimir, Loki, Grafana)

**Responsibilities:**
- Tempo trace storage configuration
- Mimir metrics storage setup
- Loki log aggregation
- Grafana dashboard creation
- Data source configuration

**File Ownership:**
- `environments/*/observability/tempo/`
- `environments/*/observability/mimir/`
- `environments/*/observability/loki/`
- `environments/*/observability/grafana/`
- `dashboards/`

### Alloy Specialists
**Primary Focus:** Grafana Alloy pipeline configuration

**Responsibilities:**
- Alloy receiver configuration
- Pipeline processing logic
- Data transformation rules
- Export target configuration
- Pipeline optimization

**File Ownership:**
- `lib/alloy/`
- `environments/*/observability/alloy-config.river`

### DevOps Automation Engineers
**Primary Focus:** CI/CD, local development, automation

**Responsibilities:**
- Tiltfile maintenance (local dev)
- Justfile task definitions
- GitHub Actions workflows
- Build and deploy automation
- Testing infrastructure

**File Ownership:**
- `Tiltfile`
- `justfile`
- `.github/workflows/`
- `scripts/automation/`

## Communication Protocols

### 1. Memory-Based Communication

Agents share state through structured memory keys:

```javascript
// Store information for other agents
mcp__claude-flow__memory_usage {
  action: "store",
  key: "swarm/platform/cluster-version",
  namespace: "coordination",
  value: JSON.stringify({
    version: "1.29.0",
    provider: "kind",
    nodes: 3,
    updated_by: "platform-engineer-001",
    timestamp: Date.now()
  })
}

// Retrieve information from other agents
mcp__claude-flow__memory_usage {
  action: "retrieve",
  key: "swarm/observability/grafana-endpoints",
  namespace: "coordination"
}
```

### 2. File-Based Handoffs

When one agent completes work that another depends on:

**Step 1: Complete work and notify**
```bash
npx claude-flow@alpha hooks post-edit \
  --file "environments/dev/kubernetes/namespace.yaml" \
  --memory-key "swarm/platform/namespace-ready" \
  --notify "namespace-created"
```

**Step 2: Dependent agent checks readiness**
```bash
npx claude-flow@alpha hooks pre-task \
  --requires "namespace-created" \
  --wait-for-dependency true
```

### 3. Dashboard Status Updates

Agents update a central dashboard for human visibility:

```bash
npx claude-flow@alpha hooks notify \
  --message "OBI experiment 'latency-analysis' deployed to prod" \
  --level "info" \
  --dashboard true
```

## Parallel Execution Patterns

### Pattern 1: Independent Workstreams

When tasks have no dependencies, launch all agents simultaneously:

```javascript
// Single message with all parallel tasks
Task("Platform Engineer", "Configure Kubernetes namespaces and RBAC", "platform-engineer")
Task("OBI Specialist", "Design latency experiment with eBPF probes", "obi-specialist")
Task("Grafana Specialist", "Setup Tempo trace storage", "grafana-specialist")
Task("Alloy Specialist", "Configure OTLP receiver pipeline", "alloy-specialist")
Task("DevOps Engineer", "Update Tiltfile for local Tempo", "devops-automation")
```

### Pattern 2: Dependency Chain

When tasks depend on each other, use memory checks:

```javascript
// Wave 1: Foundation
Task("Platform Engineer", "Create base infrastructure. Store cluster info in memory.", "platform-engineer")

// Wave 2: Observability (depends on Wave 1)
// These agents check memory for cluster readiness before starting
Task("OBI Specialist", "Deploy OBI after cluster ready. Check memory: swarm/platform/cluster-ready", "obi-specialist")
Task("Grafana Specialist", "Deploy Grafana stack after cluster ready", "grafana-specialist")

// Wave 3: Configuration (depends on Wave 2)
Task("Alloy Specialist", "Configure pipelines after OBI deployed. Check memory: swarm/obi/deployed", "alloy-specialist")
Task("Experiment Designer", "Create experiments after Grafana ready", "experiment-designer")
```

### Pattern 3: Team-Based Parallelism

Organize agents into teams with internal coordination:

```javascript
// Platform Team (parallel within team)
Task("Kubernetes Architect", "Design cluster architecture. Lead platform team.", "kubernetes-architect")
Task("Platform Engineer 1", "Implement dev environment", "platform-engineer")
Task("Platform Engineer 2", "Implement staging environment", "platform-engineer")
Task("Platform Engineer 3", "Implement prod environment", "platform-engineer")

// Observability Team (parallel within team)
Task("OBI Specialist", "Lead observability team. Design experiment framework.", "obi-specialist")
Task("Grafana Specialist 1", "Setup Tempo and Mimir", "grafana-specialist")
Task("Grafana Specialist 2", "Setup Loki and Grafana", "grafana-specialist")
Task("Alloy Specialist", "Configure all pipeline stages", "alloy-specialist")
```

## File Ownership Matrix

### Exclusive Ownership (No Conflicts)

| Directory/File | Owner Agent | Access Level |
|----------------|-------------|--------------|
| `environments/*/kubernetes/` | platform-engineer | Exclusive Write |
| `environments/*/observability/obi-*.yaml` | obi-specialist | Exclusive Write |
| `environments/*/observability/tempo/` | grafana-specialist | Exclusive Write |
| `environments/*/observability/mimir/` | grafana-specialist | Exclusive Write |
| `lib/alloy/` | alloy-specialist | Exclusive Write |
| `Tiltfile` | devops-automation | Exclusive Write |

### Shared Ownership (Coordination Required)

| Directory/File | Owners | Coordination Method |
|----------------|--------|---------------------|
| `README.md` | All agents | Post-edit hook required |
| `environments/*/values.yaml` | Platform + Observability | Memory locks |
| `scripts/deploy.sh` | Platform + DevOps | Version control |

## Conflict Resolution

### 1. File Locking

Before modifying shared files:

```bash
npx claude-flow@alpha hooks lock-file \
  --file "environments/prod/values.yaml" \
  --agent-id "platform-engineer-001" \
  --timeout 300
```

### 2. Merge Coordination

When conflicts detected:

```bash
npx claude-flow@alpha hooks resolve-conflict \
  --file "environments/prod/values.yaml" \
  --agents "platform-engineer-001,obi-specialist-002" \
  --strategy "merge"
```

### 3. Priority Rules

1. **Infrastructure First**: Platform changes take precedence
2. **Environment Isolation**: Prod changes require approval
3. **Backward Compatibility**: Never break existing deployments
4. **Testing Required**: All changes must pass validation

## Health Monitoring

### Agent Health Checks

```bash
# Self-report health status
npx claude-flow@alpha hooks agent-health \
  --agent-id "obi-specialist-001" \
  --status "healthy" \
  --cpu-usage 45 \
  --memory-usage 62
```

### Swarm Health Dashboard

```bash
# View overall swarm status
npx claude-flow@alpha hooks swarm-status \
  --session-id "swarm-mop-platform" \
  --detailed true
```

## Best Practices

### 1. Always Use Hooks
- **Never skip hooks** - they enable coordination
- Execute in correct order: pre-task → post-edit → post-task
- Include meaningful descriptions and context

### 2. Clear Ownership
- **One owner per file** when possible
- Document shared ownership explicitly
- Use memory to coordinate shared access

### 3. Atomic Commits
- Complete logical units of work
- Test before marking complete
- Update documentation with code

### 4. Memory as Source of Truth
- Store all coordination state in memory
- Use structured keys: `swarm/{team}/{resource}`
- Include timestamps and agent IDs

### 5. Graceful Degradation
- Handle dependency failures
- Provide meaningful error messages
- Enable retry mechanisms

## Example Coordination Workflow

### Scenario: Deploy New OBI Experiment

**Step 1: Platform Foundation**
```javascript
Task("Platform Engineer", `
1. Ensure namespace exists: observability
2. Check resource quotas
3. Store namespace status in memory: swarm/platform/namespace-ready
4. Execute hooks at each step
`, "platform-engineer")
```

**Step 2: OBI Deployment (waits for namespace)**
```javascript
Task("OBI Specialist", `
1. Pre-task: Check memory for swarm/platform/namespace-ready
2. Design experiment: latency-analysis
3. Create experiment YAML
4. Post-edit hook after each file
5. Store experiment config in memory: swarm/obi/latency-experiment
6. Post-task: Mark complete
`, "obi-specialist")
```

**Step 3: Grafana Dashboard (waits for OBI)**
```javascript
Task("Grafana Specialist", `
1. Pre-task: Check memory for swarm/obi/latency-experiment
2. Create dashboard for latency metrics
3. Configure data sources
4. Post-edit hook for dashboard JSON
5. Post-task: Mark complete
`, "grafana-specialist")
```

**Step 4: Pipeline Configuration (waits for both)**
```javascript
Task("Alloy Specialist", `
1. Pre-task: Check memory for experiment and dashboard
2. Configure Alloy pipeline for experiment metrics
3. Add OTLP receiver for traces
4. Post-edit hook for pipeline config
5. Post-task: Mark complete with endpoint URLs
`, "alloy-specialist")
```

## Troubleshooting

### Issue: Agent Blocked Waiting for Dependency

**Diagnosis:**
```bash
npx claude-flow@alpha hooks task-status \
  --task-id "deploy-obi-experiment" \
  --show-dependencies true
```

**Solution:**
- Check if dependency task completed
- Verify memory key exists
- Consider manual unblock if dependency failed

### Issue: File Conflict

**Diagnosis:**
```bash
npx claude-flow@alpha hooks list-locks \
  --file "environments/prod/values.yaml"
```

**Solution:**
- Identify lock holder
- Coordinate merge strategy
- Use conflict resolution hook

### Issue: Agent Unhealthy

**Diagnosis:**
```bash
npx claude-flow@alpha hooks agent-health \
  --agent-id "obi-specialist-001" \
  --history true
```

**Solution:**
- Review error logs
- Restart agent if necessary
- Reassign tasks to healthy agents

## Summary

Effective agent coordination requires:
1. ✅ Clear ownership boundaries
2. ✅ Consistent hook execution
3. ✅ Memory-based communication
4. ✅ Dependency management
5. ✅ Health monitoring
6. ✅ Conflict resolution

Follow these patterns to achieve **2.8-4.4x speed improvements** through parallel execution while maintaining code quality and preventing conflicts.
