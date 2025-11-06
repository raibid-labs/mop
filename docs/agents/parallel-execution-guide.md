# Parallel Execution Guide

This guide provides detailed instructions for executing multiple agents in parallel to achieve 2.8-4.4x speed improvements in the MOP project.

---

## Core Principle: The Golden Rule

> **"1 MESSAGE = ALL RELATED OPERATIONS"**

This is the most important principle for parallel execution. Spawning agents, batching operations, and coordinating work must ALL happen in a single message to achieve maximum parallelism.

---

## Prerequisites

### 1. Directory Structure

Ensure clean directory ownership to prevent conflicts:

```bash
# Create all necessary directories upfront
mkdir -p /Users/beengud/raibid-labs/mop/{environments,charts,lib,scripts,tests,docs}
mkdir -p /Users/beengud/raibid-labs/mop/environments/{dev,staging,prod}/{kubernetes,observability,infrastructure}
mkdir -p /Users/beengud/raibid-labs/mop/lib/{tanka,alloy}
mkdir -p /Users/beengud/raibid-labs/mop/charts/{obi-operator,obi-experiments}
```

### 2. Directory Ownership Matrix

Define clear ownership BEFORE launching agents:

| Directory | Owner Agent(s) | Access Pattern |
|-----------|----------------|----------------|
| `environments/dev/kubernetes/` | platform-engineer-dev | Exclusive |
| `environments/staging/kubernetes/` | platform-engineer-staging | Exclusive |
| `environments/prod/kubernetes/` | platform-engineer-prod | Exclusive |
| `environments/*/observability/obi-*` | obi-specialist | Exclusive |
| `environments/*/observability/tempo/` | grafana-specialist-1 | Exclusive |
| `environments/*/observability/mimir/` | grafana-specialist-1 | Exclusive |
| `environments/*/observability/loki/` | grafana-specialist-2 | Exclusive |
| `environments/*/observability/grafana/` | grafana-specialist-2 | Exclusive |
| `lib/alloy/` | alloy-specialist | Exclusive |
| `lib/tanka/` | platform-engineer | Shared (read-mostly) |
| `charts/obi-*` | obi-specialist | Exclusive |
| `Tiltfile` | devops-automation | Exclusive |
| `justfile` | devops-automation | Exclusive |
| `.github/workflows/` | cicd-engineer | Exclusive |
| `tests/` | tester | Exclusive |
| `docs/` | All agents | Shared (coordination required) |

### 3. Dependency Graph

Map dependencies BEFORE launching to determine execution waves:

```
Wave 1 (Independent):
├── Platform Team → environments/*/kubernetes/
├── DevOps Team → Tiltfile, justfile, .github/workflows/
└── Documentation → docs/ (can start anytime)

Wave 2 (Depends on Wave 1 Platform):
├── Observability Team → environments/*/observability/
└── (Waits for: swarm/platform/namespace-ready)

Wave 3 (Depends on Wave 2):
├── Integration Testing → tests/integration/
└── (Waits for: swarm/observability/deployed)
```

---

## Execution Patterns

### Pattern 1: Pure Parallel (No Dependencies)

**Use Case:** Independent workstreams with no shared files or dependencies

**Example:** Platform engineers working on different environments

```javascript
// ✅ CORRECT: All agents launched in ONE message
[Single Message]:
  Task("Platform Engineer - Dev", `
    Implement dev environment Kubernetes configuration:
    1. Pre-task hook: npx claude-flow@alpha hooks pre-task --description "Configure dev k8s"
    2. Create namespace: observability-dev
    3. Configure RBAC with developer permissions
    4. Set resource quotas: 50 CPU, 128Gi memory
    5. Deploy network policies for pod isolation
    6. Post-edit hook after each file creation
    7. Post-task hook: npx claude-flow@alpha hooks post-task --task-id "dev-k8s-config"
    8. Store completion: swarm/platform/dev-ready

    Files to create:
    - environments/dev/kubernetes/namespace.yaml
    - environments/dev/kubernetes/rbac.yaml
    - environments/dev/kubernetes/resourcequota.yaml
    - environments/dev/kubernetes/networkpolicy.yaml
  `, "platform-engineer")

  Task("Platform Engineer - Staging", `
    Implement staging environment Kubernetes configuration:
    1. Pre-task hook: npx claude-flow@alpha hooks pre-task --description "Configure staging k8s"
    2. Create namespace: observability-staging
    3. Configure RBAC with stricter permissions than dev
    4. Set resource quotas: 75 CPU, 192Gi memory
    5. Deploy network policies
    6. Post-edit hook after each file creation
    7. Post-task hook: npx claude-flow@alpha hooks post-task --task-id "staging-k8s-config"
    8. Store completion: swarm/platform/staging-ready

    Files to create:
    - environments/staging/kubernetes/namespace.yaml
    - environments/staging/kubernetes/rbac.yaml
    - environments/staging/kubernetes/resourcequota.yaml
    - environments/staging/kubernetes/networkpolicy.yaml
  `, "platform-engineer")

  Task("Platform Engineer - Prod", `
    Implement production environment Kubernetes configuration:
    1. Pre-task hook: npx claude-flow@alpha hooks pre-task --description "Configure prod k8s"
    2. Create namespace: observability-prod
    3. Configure RBAC with strict production permissions
    4. Set resource quotas: 100 CPU, 256Gi memory
    5. Deploy comprehensive network policies with zero-trust model
    6. Enable audit logging
    7. Post-edit hook after each file creation
    8. Post-task hook: npx claude-flow@alpha hooks post-task --task-id "prod-k8s-config"
    9. Store completion: swarm/platform/prod-ready

    Files to create:
    - environments/prod/kubernetes/namespace.yaml
    - environments/prod/kubernetes/rbac.yaml
    - environments/prod/kubernetes/resourcequota.yaml
    - environments/prod/kubernetes/networkpolicy.yaml
    - environments/prod/kubernetes/auditpolicy.yaml
  `, "platform-engineer")

  // Batch ALL todos in ONE call
  TodoWrite { todos: [
    {id: "1", content: "Configure dev environment", status: "in_progress", activeForm: "Configuring dev environment", priority: "high"},
    {id: "2", content: "Configure staging environment", status: "in_progress", activeForm: "Configuring staging environment", priority: "high"},
    {id: "3", content: "Configure prod environment", status: "in_progress", activeForm: "Configuring prod environment", priority: "high"},
    {id: "4", content: "Review all configurations", status: "pending", activeForm: "Reviewing configurations", priority: "medium"},
    {id: "5", content: "Deploy to clusters", status: "pending", activeForm: "Deploying to clusters", priority: "medium"}
  ]}
```

**Result:** 3 agents complete in ~45 minutes (vs ~2 hours sequential)

---

### Pattern 2: Sequential Waves (With Dependencies)

**Use Case:** Workstreams with clear dependencies

**Example:** Platform → Observability → Testing

```javascript
// ✅ CORRECT: Launch Wave 1 in one message, Wave 2 in another (after Wave 1 completes)

[Message 1 - Wave 1: Platform Foundation]:
  Task("Platform Engineer - Dev", "Configure dev k8s...", "platform-engineer")
  Task("Platform Engineer - Staging", "Configure staging k8s...", "platform-engineer")
  Task("Platform Engineer - Prod", "Configure prod k8s...", "platform-engineer")
  Task("Terraform Specialist", "Provision cloud resources...", "terraform-specialist")

  TodoWrite { todos: [...Wave 1 todos...] }

// Wait for Wave 1 completion, then launch Wave 2

[Message 2 - Wave 2: Observability Stack]:
  Task("OBI Specialist", `
    Deploy OBI experiment framework:
    1. Pre-task hook: Check memory for swarm/platform/prod-ready (MUST exist)
    2. If not ready, wait and check again
    3. Design OBI experiment framework...
  `, "obi-specialist")

  Task("Grafana Specialist 1", `
    Deploy Tempo and Mimir:
    1. Pre-task hook: Check memory for swarm/platform/prod-ready (MUST exist)
    2. If not ready, wait and check again
    3. Configure Tempo with S3 backend...
  `, "grafana-specialist")

  Task("Grafana Specialist 2", `
    Deploy Loki and Grafana:
    1. Pre-task hook: Check memory for swarm/platform/prod-ready (MUST exist)
    2. If not ready, wait and check again
    3. Configure Loki with S3 backend...
  `, "grafana-specialist")

  TodoWrite { todos: [...Wave 2 todos...] }
```

**Result:** Wave 1 + Wave 2 complete in ~1.5 hours (vs ~4 hours sequential)

---

### Pattern 3: Partial Dependencies (Smart Parallelism)

**Use Case:** Some work can start before full completion of dependencies

**Example:** Observability can start once namespace exists, doesn't need full platform completion

```javascript
// ✅ CORRECT: Launch all agents together, use granular dependency checks

[Single Message - Smart Parallel Launch]:
  // Platform agents start immediately
  Task("Platform Engineer - Prod", `
    1. Create namespace FIRST (highest priority)
    2. Store completion immediately: swarm/platform/namespace-ready
    3. Then continue with RBAC, quotas, etc.
  `, "platform-engineer")

  // Observability agents check for PARTIAL completion
  Task("OBI Specialist", `
    1. Pre-task hook: Loop until swarm/platform/namespace-ready exists
    2. Once namespace ready, start OBI deployment (don't wait for full platform)
    3. Continue with experiment design...
  `, "obi-specialist")

  Task("Grafana Specialist", `
    1. Pre-task hook: Loop until swarm/platform/namespace-ready exists
    2. Once namespace ready, start Grafana deployment
    3. Continue with Tempo/Mimir/Loki...
  `, "grafana-specialist")

  // DevOps can run fully parallel (no dependencies)
  Task("DevOps Automation", `
    1. Pre-task hook: No dependencies, start immediately
    2. Create Tiltfile with all services...
  `, "devops-automation")
```

**Result:** Maximum parallelism, complete in ~1 hour (vs ~4 hours sequential)

**Key Insight:** Don't over-serialize! Use granular memory keys for partial completion.

---

## Memory-Based Coordination

### Memory Key Conventions

Use structured keys for coordination:

```
swarm/
├── platform/
│   ├── dev-ready              # Dev environment complete
│   ├── staging-ready          # Staging environment complete
│   ├── prod-ready             # Prod environment complete
│   ├── namespace-ready        # Namespace created (partial completion)
│   ├── terraform-outputs      # Cloud resource info
│   └── cluster-info           # Cluster connection details
├── observability/
│   ├── obi-deployed           # OBI experiments deployed
│   ├── grafana-deployed       # Grafana stack deployed
│   ├── tempo-endpoint         # Tempo URL
│   ├── mimir-endpoint         # Mimir URL
│   ├── loki-endpoint          # Loki URL
│   ├── grafana-url            # Grafana dashboard URL
│   └── alloy-config           # Alloy pipeline status
├── devops/
│   ├── ci-ready               # CI/CD pipelines ready
│   ├── dev-env-ready          # Local dev environment ready
│   └── tests-passing          # Test suite status
└── shared/
    ├── dependencies           # Cross-team dependencies
    └── decisions              # Shared architectural decisions
```

### Storing Information

```javascript
// Agent stores completion status
mcp__claude-flow__memory_usage {
  action: "store",
  key: "swarm/platform/dev-ready",
  namespace: "coordination",
  value: JSON.stringify({
    status: "complete",
    agent: "platform-engineer-dev-001",
    timestamp: Date.now(),
    namespace: "observability-dev",
    resourceQuota: { cpu: "50", memory: "128Gi" },
    files: [
      "environments/dev/kubernetes/namespace.yaml",
      "environments/dev/kubernetes/rbac.yaml"
    ]
  })
}
```

### Retrieving Information

```javascript
// Dependent agent checks for readiness
mcp__claude-flow__memory_usage {
  action: "retrieve",
  key: "swarm/platform/dev-ready",
  namespace: "coordination"
}

// In agent task instructions:
// "1. Pre-task: Check memory for swarm/platform/dev-ready"
// "2. If not exists, wait 30 seconds and check again (max 10 attempts)"
// "3. If exists, proceed with deployment"
```

---

## Hook Execution Protocol

### Pre-Task Hook

Execute BEFORE starting any work:

```bash
npx claude-flow@alpha hooks pre-task \
  --description "Deploy OBI experiments to production" \
  --agent-id "obi-specialist-001" \
  --session-id "swarm-mop-platform" \
  --requires "platform-infrastructure-ready"
```

**What it does:**
- Validates agent is ready
- Checks dependencies in memory
- Locks resources if needed
- Restores previous session context
- Returns: proceed or wait

### Post-Edit Hook

Execute AFTER each significant file change:

```bash
npx claude-flow@alpha hooks post-edit \
  --file "environments/prod/observability/obi-config.yaml" \
  --memory-key "swarm/obi/prod-config" \
  --agent-id "obi-specialist-001" \
  --notify "obi-config-updated"
```

**What it does:**
- Stores file change in memory
- Notifies dependent agents
- Trains neural patterns on the edit
- Updates dashboards
- Triggers dependent tasks if configured

### Post-Task Hook

Execute AFTER completing entire task:

```bash
npx claude-flow@alpha hooks post-task \
  --task-id "deploy-obi-experiments" \
  --status "complete" \
  --agent-id "obi-specialist-001" \
  --output-files "environments/prod/observability/experiments/latency-test.yaml" \
  --summary "Deployed 3 OBI experiments: latency, errors, throughput"
```

**What it does:**
- Marks task as complete
- Releases locked resources
- Updates metrics and analytics
- Generates task summary
- Notifies coordinator

---

## Conflict Resolution

### Prevention Strategies

1. **Clear Ownership**: Each file owned by exactly one agent
2. **Directory Isolation**: Agents work in separate directories
3. **Memory Locks**: Lock shared files before editing
4. **Granular Updates**: Small, atomic changes

### Detecting Conflicts

```bash
# Agent checks for file locks before editing
npx claude-flow@alpha hooks check-lock \
  --file "environments/prod/values.yaml" \
  --agent-id "platform-engineer-001"
```

### Resolving Conflicts

```bash
# If conflict detected, request resolution
npx claude-flow@alpha hooks resolve-conflict \
  --file "environments/prod/values.yaml" \
  --agents "platform-engineer-001,obi-specialist-002" \
  --strategy "merge" \
  --priority "platform-engineer"
```

**Resolution Strategies:**
- `merge`: Merge both changes (if possible)
- `priority`: Use changes from higher-priority agent
- `latest`: Use most recent changes
- `manual`: Flag for human resolution

---

## Real-World Example: Complete Platform Deployment

### Scenario
Deploy complete MOP platform with:
- 3 environments (dev, staging, prod)
- OBI experiments
- Grafana stack
- Alloy pipelines
- CI/CD
- Tests

### Step 1: Preparation (Sequential)

```javascript
[Message 1 - Preparation]:
  // Create directory structure
  Bash "mkdir -p /Users/beengud/raibid-labs/mop/environments/{dev,staging,prod}/{kubernetes,observability,infrastructure}"
  Bash "mkdir -p /Users/beengud/raibid-labs/mop/{charts,lib,scripts,tests,docs}/..."

  // Initialize coordination
  mcp__claude-flow__swarm_init { topology: "hierarchical", maxAgents: 15 }
```

### Step 2: Launch All Agents (Parallel - ONE MESSAGE)

```javascript
[Message 2 - Full Parallel Launch]:
  // Coordinator
  Task("Hierarchical Coordinator", `
    Coordinate full platform deployment:
    1. Monitor all team progress
    2. Resolve dependencies
    3. Handle conflicts
    4. Report status every 30 minutes
    5. Coordinate final integration
  `, "hierarchical-coordinator")

  // Platform Team (6 agents - parallel)
  Task("Kubernetes Architect", `
    Lead platform team:
    1. Design cluster architecture
    2. Define standards and conventions
    3. Review all platform changes
    4. Mentor platform engineers
    5. Store architecture in memory: swarm/platform/architecture
  `, "kubernetes-architect")

  Task("Platform Engineer - Dev", `
    Configure dev environment (full instructions from Pattern 1)
    Store completion: swarm/platform/dev-ready
  `, "platform-engineer")

  Task("Platform Engineer - Staging", `
    Configure staging environment
    Store completion: swarm/platform/staging-ready
  `, "platform-engineer")

  Task("Platform Engineer - Prod", `
    Configure prod environment
    Store completion: swarm/platform/prod-ready
  `, "platform-engineer")

  Task("Terraform Specialist", `
    Provision cloud infrastructure:
    1. Create Terraform modules for K8s clusters
    2. Configure networking and security groups
    3. Apply: dev → staging → prod
    4. Store outputs: swarm/platform/terraform-outputs
  `, "terraform-specialist")

  Task("Production Validator", `
    Validate all environments:
    1. Wait for all environments ready in memory
    2. Check security: RBAC, network policies
    3. Validate resources: quotas, limits
    4. Check HA: replicas, PDBs
    5. Approve or flag issues
  `, "production-validator")

  // Observability Team (5 agents - parallel, depends on platform namespace)
  Task("OBI Specialist", `
    Lead observability team and deploy OBI:
    1. Pre-task: Wait for swarm/platform/namespace-ready (any env)
    2. Design experiment framework
    3. Create OBI operator deployment
    4. Deploy experiments: latency, errors, throughput
    5. Store completion: swarm/observability/obi-deployed
  `, "obi-specialist")

  Task("Grafana Specialist 1", `
    Deploy Tempo and Mimir:
    1. Pre-task: Wait for swarm/platform/namespace-ready
    2. Deploy Tempo with S3 backend
    3. Deploy Mimir with S3 backend
    4. Store endpoints: swarm/observability/tempo-endpoint, swarm/observability/mimir-endpoint
  `, "grafana-specialist")

  Task("Grafana Specialist 2", `
    Deploy Loki and Grafana:
    1. Pre-task: Wait for swarm/platform/namespace-ready
    2. Deploy Loki with S3 backend
    3. Deploy Grafana with data sources
    4. Create dashboards: OBI Overview, System Health
    5. Store URL: swarm/observability/grafana-url
  `, "grafana-specialist")

  Task("Alloy Specialist", `
    Configure Alloy pipelines:
    1. Pre-task: Wait for swarm/observability/obi-deployed
    2. Configure OTLP receiver
    3. Set up pipelines: metrics, traces, logs
    4. Deploy to all environments
    5. Store status: swarm/observability/alloy-config
  `, "alloy-specialist")

  Task("Experiment Designer", `
    Design observability experiments:
    1. Design experiment methodology
    2. Define metrics and KPIs
    3. Create experiment specs
    4. Document analysis plans
    5. Store designs: swarm/observability/experiment-specs
  `, "experiment-designer")

  // DevOps Team (4 agents - fully parallel, no dependencies)
  Task("DevOps Automation", `
    Setup local development:
    1. Create Tiltfile with all services
    2. Create Justfile with common tasks
    3. Configure hot reload
    4. Set up port forwards
    5. Store config: swarm/devops/dev-env-ready
  `, "devops-automation")

  Task("CI/CD Engineer", `
    Build CI/CD pipelines:
    1. Create GitHub Actions for PR validation
    2. Create deployment workflows
    3. Configure auto-deploy to dev
    4. Add manual approval for prod
    5. Store status: swarm/devops/ci-ready
  `, "cicd-engineer")

  Task("Tester", `
    Create test suite:
    1. Unit tests for config validation
    2. Integration tests for full stack
    3. Load tests for experiments
    4. Achieve 80%+ coverage
    5. Store results: swarm/devops/tests-passing
  `, "tester")

  Task("DevOps Troubleshooter", `
    Monitor and debug:
    1. Set up monitoring for all agents
    2. Watch for errors and conflicts
    3. Debug issues as they arise
    4. Provide support to other agents
    5. Log all issues and resolutions
  `, "devops-troubleshooter")

  // Batch ALL todos in ONE call
  TodoWrite { todos: [
    // Platform todos (6)
    {id: "1", content: "Design platform architecture", status: "in_progress", activeForm: "Designing platform architecture", priority: "high"},
    {id: "2", content: "Configure dev environment", status: "in_progress", activeForm: "Configuring dev environment", priority: "high"},
    {id: "3", content: "Configure staging environment", status: "in_progress", activeForm: "Configuring staging environment", priority: "high"},
    {id: "4", content: "Configure prod environment", status: "in_progress", activeForm: "Configuring prod environment", priority: "high"},
    {id: "5", content: "Provision cloud infrastructure", status: "in_progress", activeForm: "Provisioning cloud infrastructure", priority: "high"},
    {id: "6", content: "Validate production readiness", status: "pending", activeForm: "Validating production readiness", priority: "high"},

    // Observability todos (5)
    {id: "7", content: "Deploy OBI experiments", status: "in_progress", activeForm: "Deploying OBI experiments", priority: "high"},
    {id: "8", content: "Deploy Tempo and Mimir", status: "in_progress", activeForm: "Deploying Tempo and Mimir", priority: "high"},
    {id: "9", content: "Deploy Loki and Grafana", status: "in_progress", activeForm: "Deploying Loki and Grafana", priority: "high"},
    {id: "10", content: "Configure Alloy pipelines", status: "pending", activeForm: "Configuring Alloy pipelines", priority: "medium"},
    {id: "11", content: "Design experiments", status: "in_progress", activeForm: "Designing experiments", priority: "medium"},

    // DevOps todos (4)
    {id: "12", content: "Setup local dev environment", status: "in_progress", activeForm: "Setting up local dev environment", priority: "medium"},
    {id: "13", content: "Build CI/CD pipelines", status: "in_progress", activeForm: "Building CI/CD pipelines", priority: "medium"},
    {id: "14", content: "Create test suite", status: "in_progress", activeForm: "Creating test suite", priority: "medium"},
    {id: "15", content: "Monitor and debug", status: "in_progress", activeForm: "Monitoring and debugging", priority: "low"}
  ]}
```

### Step 3: Results

**Timeline:**
- **Minute 0**: All 15 agents launch simultaneously
- **Minute 15**: Platform namespaces created, observability agents start deploying
- **Minute 45**: Platform complete, observability in progress, DevOps ~80% done
- **Minute 90**: Observability complete, starting integration tests
- **Minute 120**: Full platform deployed, tested, validated

**Total Time: 2 hours** (vs 8-12 hours sequential)

**Speed Improvement: 4-6x**

---

## Common Pitfalls

### ❌ WRONG: Sequential Agent Spawning

```javascript
// DON'T DO THIS - Multiple messages
Message 1: Task("Agent 1", ...)
Message 2: Task("Agent 2", ...)
Message 3: Task("Agent 3", ...)
// This breaks parallel coordination!
```

### ❌ WRONG: Incomplete Dependency Checks

```javascript
Task("OBI Specialist", `
  1. Deploy OBI immediately  // Missing dependency check!
  2. Hope platform is ready
`)
// This will fail if platform not ready!
```

### ❌ WRONG: Shared File Ownership

```javascript
Task("Agent 1", "Edit environments/prod/values.yaml", ...)
Task("Agent 2", "Edit environments/prod/values.yaml", ...)
// This creates conflicts!
```

### ❌ WRONG: Missing Hook Execution

```javascript
Task("Agent", `
  1. Create file
  2. Deploy
  // Missing all hooks!
`)
// No coordination = chaos
```

---

## Performance Optimization

### 1. Minimize Cross-Team Dependencies

**Bad:**
```
Platform → Observability → Testing → Validation
(Fully sequential)
```

**Good:**
```
Platform → Observability
    ↓
  DevOps (parallel)
    ↓
  Testing (parallel)
```

### 2. Use Granular Memory Keys

**Bad:**
```
swarm/platform/complete  // All or nothing
```

**Good:**
```
swarm/platform/namespace-ready  // Partial completion
swarm/platform/rbac-ready
swarm/platform/network-ready
swarm/platform/complete
```

### 3. Batch All Operations

**Bad:**
```javascript
Message 1: TodoWrite { todos: [todo1] }
Message 2: TodoWrite { todos: [todo2] }
Message 3: Task("Agent 1", ...)
```

**Good:**
```javascript
Message 1:
  TodoWrite { todos: [todo1, todo2, ..., todo15] }
  Task("Agent 1", ...)
  Task("Agent 2", ...)
  // ... all tasks
```

### 4. Right-Size Teams

| Team Size | Coordination Overhead | Optimal Use Case |
|-----------|----------------------|------------------|
| 2-3 agents | ~5% | Simple tasks |
| 4-6 agents | ~10% | Standard workstreams |
| 7-10 agents | ~15% | Complex projects |
| 11-15 agents | ~20% | Full platform |
| 16+ agents | ~30%+ | Avoid if possible |

---

## Monitoring Progress

### Dashboard View

```bash
# Check overall swarm status
npx claude-flow@alpha hooks swarm-status \
  --session-id "swarm-mop-platform" \
  --detailed true
```

**Output:**
```
Swarm Status: ACTIVE
Agents: 15 total, 12 active, 3 complete
Progress: 67%

Platform Team (6 agents):
  ✅ platform-engineer-dev-001: COMPLETE
  ✅ platform-engineer-staging-001: COMPLETE
  🔄 platform-engineer-prod-001: ACTIVE (85%)
  🔄 terraform-specialist-001: ACTIVE (60%)
  ⏳ production-validator-001: PENDING
  📋 kubernetes-architect-001: REVIEWING

Observability Team (5 agents):
  🔄 obi-specialist-001: ACTIVE (40%)
  🔄 grafana-specialist-001: ACTIVE (55%)
  🔄 grafana-specialist-002: ACTIVE (45%)
  ⏳ alloy-specialist-001: PENDING (waiting: obi-deployed)
  🔄 experiment-designer-001: ACTIVE (70%)

DevOps Team (4 agents):
  ✅ devops-automation-001: COMPLETE
  🔄 cicd-engineer-001: ACTIVE (80%)
  🔄 tester-001: ACTIVE (60%)
  🔄 devops-troubleshooter-001: ACTIVE (monitoring)
```

### Individual Agent Status

```bash
# Check specific agent
npx claude-flow@alpha hooks agent-status \
  --agent-id "obi-specialist-001"
```

---

## Troubleshooting

### Issue: Agent Blocked on Dependency

**Symptoms:**
- Agent status: PENDING
- No progress for >10 minutes

**Diagnosis:**
```bash
npx claude-flow@alpha hooks task-status \
  --task-id "deploy-obi" \
  --show-dependencies true
```

**Solution:**
1. Check if dependency task completed
2. Verify memory key exists
3. Check for errors in dependency task
4. Consider manual unblock if dependency failed but not critical

### Issue: File Conflict

**Symptoms:**
- Multiple agents trying to edit same file
- Git merge conflicts

**Diagnosis:**
```bash
npx claude-flow@alpha hooks list-locks \
  --file "environments/prod/values.yaml"
```

**Solution:**
1. Identify conflicting agents
2. Determine priority (platform > observability > devops)
3. Coordinate merge strategy
4. Use conflict resolution hook

### Issue: Slow Progress

**Symptoms:**
- Agents taking longer than expected
- Overall progress < 50% after 1 hour

**Diagnosis:**
```bash
npx claude-flow@alpha hooks swarm-metrics \
  --session-id "swarm-mop-platform"
```

**Solution:**
1. Check for bottlenecks (one agent blocking many others)
2. Consider breaking up large tasks
3. Add more agents to slow workstreams
4. Optimize dependencies (make more parallel)

---

## Best Practices Summary

### ✅ DO

1. **Launch all agents in ONE message** (Golden Rule)
2. **Define clear directory ownership** before launching
3. **Map dependencies** and use memory for coordination
4. **Execute hooks consistently** (pre-task, post-edit, post-task)
5. **Use granular memory keys** for partial completion
6. **Batch all operations** (todos, file ops, bash commands)
7. **Monitor progress** with dashboard
8. **Right-size teams** (4-8 agents optimal)

### ❌ DON'T

1. **Don't spawn agents in multiple messages** (breaks parallelism)
2. **Don't share file ownership** without coordination
3. **Don't skip dependency checks** (will cause failures)
4. **Don't skip hooks** (breaks coordination)
5. **Don't over-serialize** (use partial dependencies)
6. **Don't ignore conflicts** (resolve immediately)
7. **Don't create teams >15 agents** (coordination overhead too high)

---

## Expected Results

Following this guide, you should achieve:

- ✅ **2.8-4.4x speed improvement** vs sequential execution
- ✅ **80%+ parallel execution** (vs 0-20% ad-hoc)
- ✅ **Zero file conflicts** (clear ownership)
- ✅ **Smooth coordination** (memory + hooks)
- ✅ **Predictable timelines** (dependency-aware)
- ✅ **High quality output** (parallel doesn't mean rushed)

**Example Timeline:**
- Small task (4 agents): 30-45 min (vs 2 hours)
- Medium task (8 agents): 1-1.5 hours (vs 4 hours)
- Large task (15 agents): 2-3 hours (vs 8-12 hours)

---

## Next Steps

1. Review `team-compositions.md` for standard team patterns
2. Study `agent-definitions.md` for agent capabilities
3. Check `coordination.md` for coordination protocols
4. Try a small parallel deployment (4 agents) first
5. Scale up to larger teams once comfortable
6. Monitor metrics and optimize

**Remember:** Parallel execution is powerful, but requires discipline. Follow the Golden Rule: "1 MESSAGE = ALL RELATED OPERATIONS" and you'll achieve dramatic speed improvements.
