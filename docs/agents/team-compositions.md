# Team Compositions

This document defines standard team compositions for the MOP project based on raibid-labs patterns. Teams are organized by workstream with clear roles, model distribution, and coordination strategies.

---

## Core Principles

### Team Design Guidelines

1. **Size**: 4-8 agents per team for optimal coordination
2. **Leadership**: 1 Opus agent as team lead for strategic decisions
3. **Workers**: 3-7 Sonnet agents for implementation
4. **Specialization**: Clear domain boundaries between teams
5. **Coordination**: Mesh within teams, hierarchical between teams

### Model Distribution Rationale

- **Opus (15-20%)**: Strategic thinking, architecture, complex decisions
- **Sonnet (80-85%)**: Implementation, analysis, most development work
- **Haiku (0-5%)**: Simple tasks only (rarely needed for MOP)

### Workstream Assignment

Each team owns a specific workstream with:
- Dedicated directory ownership
- Clear boundaries
- Minimal cross-team dependencies
- Parallel execution capability

---

## Standard Team Compositions

### 1. Platform Team (6-8 agents)

**Mission**: Kubernetes infrastructure, cluster management, base platform

**Composition:**

```yaml
Team: Platform
Size: 6-8 agents
Model Distribution: 1-2 Opus, 5-6 Sonnet
Coordination: Mesh within team, reports to hierarchical-coordinator
```

**Agents:**

1. **kubernetes-architect** (Opus) - Team Lead
   - Role: Architecture decisions, technical leadership
   - Responsibilities: Design cluster architecture, set standards, mentor team
   - Time allocation: 30% design, 40% review, 30% coordination

2. **cloud-architect** (Opus) - Strategic advisor
   - Role: Cloud strategy, cost optimization
   - Responsibilities: Multi-cloud design, vendor decisions, cost analysis
   - Time allocation: 40% strategy, 30% optimization, 30% compliance

3. **platform-engineer-1** (Sonnet) - Dev environment
   - Role: Dev cluster implementation
   - Responsibilities: Dev environment Kubernetes configs, dev Tanka setup
   - File ownership: `environments/dev/kubernetes/`, `environments/dev/infrastructure/`

4. **platform-engineer-2** (Sonnet) - Staging environment
   - Role: Staging cluster implementation
   - Responsibilities: Staging environment configs, staging Tanka
   - File ownership: `environments/staging/kubernetes/`, `environments/staging/infrastructure/`

5. **platform-engineer-3** (Sonnet) - Production environment
   - Role: Production cluster implementation
   - Responsibilities: Production configs, production Tanka, high-availability setup
   - File ownership: `environments/prod/kubernetes/`, `environments/prod/infrastructure/`

6. **terraform-specialist** (Sonnet)
   - Role: Infrastructure as code
   - Responsibilities: Terraform modules, state management, cloud resources
   - File ownership: `infrastructure/terraform/`

7. **devops-troubleshooter** (Sonnet) - Optional
   - Role: Platform debugging and incident response
   - Responsibilities: Debug cluster issues, performance tuning, incident handling

8. **production-validator** (Sonnet) - Optional
   - Role: Production readiness validation
   - Responsibilities: Validate prod changes, security audits, compliance checks

**Team Coordination:**

```javascript
// Initialize platform team
Task("Kubernetes Architect", `
Lead platform team deployment:
1. Initialize mesh coordination for platform team
2. Design cluster architecture for dev/staging/prod
3. Define naming conventions and standards
4. Assign tasks to platform engineers by environment
5. Review all platform changes before merge
6. Store architecture decisions in memory: swarm/platform/architecture
`, "kubernetes-architect")

// Parallel environment implementation
Task("Platform Engineer 1 - Dev", `
Implement dev environment:
1. Pre-task: Get architecture from memory
2. Create namespace: observability-dev
3. Configure RBAC and network policies
4. Set resource quotas: 50 CPU, 128Gi memory
5. Deploy service mesh
6. Post-edit hooks for each file
7. Store completion in memory: swarm/platform/dev-ready
`, "platform-engineer")

Task("Platform Engineer 2 - Staging", `
Implement staging environment:
1. Pre-task: Get architecture from memory
2. Create namespace: observability-staging
3. Configure RBAC and network policies
4. Set resource quotas: 75 CPU, 192Gi memory
5. Deploy service mesh
6. Store completion in memory: swarm/platform/staging-ready
`, "platform-engineer")

Task("Platform Engineer 3 - Prod", `
Implement production environment:
1. Pre-task: Get architecture and compliance requirements
2. Create namespace: observability-prod
3. Configure strict RBAC and network policies
4. Set resource quotas: 100 CPU, 256Gi memory
5. Deploy service mesh with mTLS
6. Enable audit logging
7. Store completion in memory: swarm/platform/prod-ready
`, "platform-engineer")

// Terraform infrastructure
Task("Terraform Specialist", `
Provision cloud infrastructure:
1. Create Terraform modules for Kubernetes clusters
2. Configure networking and security groups
3. Set up storage classes and persistent volumes
4. Apply Terraform in dev → staging → prod order
5. Store outputs in memory: swarm/platform/terraform-outputs
`, "terraform-specialist")
```

**Deliverables:**
- Kubernetes clusters in all environments
- Tanka library for infrastructure
- Network policies and RBAC
- Resource quotas and limits
- Documentation: architecture, runbooks

---

### 2. Observability Team (5-7 agents)

**Mission**: OBI experiments, Grafana stack, telemetry pipelines

**Composition:**

```yaml
Team: Observability
Size: 5-7 agents
Model Distribution: 1 Opus (optional), 5-6 Sonnet
Coordination: Mesh within team, reports to hierarchical-coordinator
```

**Agents:**

1. **obi-specialist** (Sonnet) - Team Lead
   - Role: OBI experiment design and implementation
   - Responsibilities: Lead observability team, design experiments, eBPF configuration
   - File ownership: `charts/obi-*/`, `experiments/`
   - Time allocation: 20% leadership, 50% implementation, 30% review

2. **grafana-specialist-1** (Sonnet) - Metrics & Traces
   - Role: Tempo and Mimir configuration
   - Responsibilities: Trace storage, metrics storage, long-term retention
   - File ownership: `environments/*/observability/tempo/`, `environments/*/observability/mimir/`

3. **grafana-specialist-2** (Sonnet) - Logs & Dashboards
   - Role: Loki and Grafana configuration
   - Responsibilities: Log aggregation, dashboard creation, data sources
   - File ownership: `environments/*/observability/loki/`, `environments/*/observability/grafana/`, `dashboards/`

4. **alloy-specialist** (Sonnet)
   - Role: Grafana Alloy pipeline configuration
   - Responsibilities: OTLP receivers, data transformation, export targets
   - File ownership: `lib/alloy/`, `environments/*/observability/alloy-config.river`

5. **experiment-designer** (Sonnet)
   - Role: Experiment methodology and analysis
   - Responsibilities: Design experiments, define metrics, analyze results
   - File ownership: `experiments/`, `docs/experiments/`, `analysis/`

6. **performance-engineer** (Sonnet) - Optional
   - Role: Performance analysis and optimization
   - Responsibilities: Benchmark experiments, optimize resource usage, capacity planning

7. **system-architect** (Opus) - Optional, Strategic advisor
   - Role: Observability architecture
   - Responsibilities: Design observability strategy, vendor decisions, long-term planning

**Team Coordination:**

```javascript
// Initialize observability team
Task("OBI Specialist - Team Lead", `
Lead observability team deployment:
1. Pre-task: Wait for platform team completion (swarm/platform/*-ready)
2. Design OBI experiment framework
3. Define metrics taxonomy and labels
4. Assign tasks: Grafana specialists (parallel), Alloy specialist, Experiment designer
5. Review all observability configurations
6. Store framework in memory: swarm/observability/framework
`, "obi-specialist")

// Parallel Grafana stack deployment
Task("Grafana Specialist 1 - Metrics & Traces", `
Deploy Tempo and Mimir:
1. Pre-task: Check platform readiness and storage availability
2. Configure Tempo with S3 backend for traces
3. Set trace retention: 7 days
4. Configure Mimir with S3 backend for metrics
5. Set metrics retention: 30 days
6. Post-edit hooks for each component
7. Store endpoints in memory: swarm/grafana/tempo-endpoint, swarm/grafana/mimir-endpoint
`, "grafana-specialist")

Task("Grafana Specialist 2 - Logs & Dashboards", `
Deploy Loki and Grafana:
1. Pre-task: Check platform readiness
2. Configure Loki with S3 backend for logs
3. Set log retention: 14 days
4. Deploy Grafana with authentication
5. Configure data sources: Tempo, Mimir, Loki
6. Create dashboards: OBI Overview, Experiment Details, System Health
7. Store Grafana URL in memory: swarm/grafana/grafana-url
`, "grafana-specialist")

Task("Alloy Specialist", `
Configure Alloy pipelines:
1. Pre-task: Wait for OBI and Grafana endpoints in memory
2. Configure OTLP receiver on port 4317
3. Set up metrics pipeline: OBI → transformation → Mimir
4. Set up traces pipeline: OBI → sampling → Tempo
5. Set up logs pipeline: cluster → filtering → Loki
6. Add labels: cluster, namespace, experiment
7. Deploy Alloy to all environments
8. Post-edit hooks, store pipeline status
`, "alloy-specialist")

Task("Experiment Designer", `
Design initial experiments:
1. Pre-task: Review OBI framework from memory
2. Design experiment: HTTP latency analysis
3. Define metrics: p50, p95, p99 latency by endpoint
4. Create experiment YAML with eBPF probes
5. Design experiment: Error rate tracking
6. Define metrics: error_rate, error_count by service
7. Document experiment methodology
8. Store experiment specs in memory
`, "experiment-designer")

// Optional performance analysis
Task("Performance Engineer", `
Benchmark and optimize:
1. Test different sampling rates: 0.1%, 1%, 10%
2. Measure resource overhead of experiments
3. Identify bottlenecks in pipeline
4. Recommend optimal configurations
5. Create performance dashboard
6. Document findings
`, "performance-engineer")
```

**Deliverables:**
- OBI experiment framework
- Grafana stack deployment (Tempo, Mimir, Loki, Grafana)
- Alloy data pipelines
- Initial experiments (latency, errors)
- Dashboards and alerts
- Documentation: experiment guide, runbooks

---

### 3. DevOps Team (4-5 agents)

**Mission**: CI/CD, local development, automation, testing

**Composition:**

```yaml
Team: DevOps
Size: 4-5 agents
Model Distribution: 4-5 Sonnet (no Opus needed for tactical work)
Coordination: Mesh within team
```

**Agents:**

1. **devops-automation** (Sonnet) - Team Lead
   - Role: Development workflow and automation
   - Responsibilities: Tiltfile, Justfile, local dev environment
   - File ownership: `Tiltfile`, `justfile`, `scripts/automation/`

2. **cicd-engineer** (Sonnet)
   - Role: CI/CD pipelines
   - Responsibilities: GitHub Actions, build automation, deployment pipelines
   - File ownership: `.github/workflows/`, `scripts/ci/`

3. **tester** (Sonnet)
   - Role: Testing and validation
   - Responsibilities: Unit tests, integration tests, load tests
   - File ownership: `tests/`

4. **production-validator** (Sonnet)
   - Role: Production readiness
   - Responsibilities: Validate prod changes, security audits, compliance
   - Time allocation: 60% validation, 40% auditing

5. **devops-troubleshooter** (Sonnet) - Optional
   - Role: Incident response and debugging
   - Responsibilities: Debug production issues, performance tuning, on-call

**Team Coordination:**

```javascript
// DevOps team can work mostly in parallel
Task("DevOps Automation - Team Lead", `
Set up local development environment:
1. Create Tiltfile with all services: OBI, Grafana, Alloy
2. Configure hot reload for configs
3. Set up port forwards: Grafana (3000), Tempo (3200)
4. Create Justfile targets: dev-up, dev-down, dev-logs, dev-test
5. Add scripts for common tasks
6. Test full environment startup
7. Document development workflow
8. Store dev environment info in memory
`, "devops-automation")

Task("CI/CD Engineer", `
Build CI/CD pipelines:
1. Create GitHub Actions workflow for PR validation
2. Add jobs: lint, test, build, validate-k8s
3. Create deployment workflow: dev → staging → prod
4. Add auto-deploy to dev on merge to main
5. Require manual approval for prod deploys
6. Configure secrets and credentials
7. Test pipelines end-to-end
8. Document CI/CD process
`, "cicd-engineer")

Task("Tester", `
Create comprehensive test suite:
1. Write unit tests for OBI experiment validation
2. Create integration test: deploy experiment to test cluster
3. Write load test: 1000 req/s, verify metrics collected
4. Test Grafana dashboard rendering
5. Test Alloy pipeline data flow
6. Validate all environments: dev, staging, prod
7. Achieve 80%+ test coverage
8. Document testing strategy
`, "tester")

Task("Production Validator", `
Validate production readiness:
1. Check security: RBAC, network policies, secrets management
2. Validate resource limits: CPU, memory, storage
3. Check high availability: replicas, pod disruption budgets
4. Audit configurations for best practices
5. Validate backup and disaster recovery
6. Check compliance requirements
7. Approve or flag production changes
8. Document validation criteria
`, "production-validator")
```

**Deliverables:**
- Local development environment (Tilt)
- CI/CD pipelines (GitHub Actions)
- Comprehensive test suite
- Production validation criteria
- Documentation: dev setup, CI/CD guide, testing guide

---

### 4. Full Platform Team (12-15 agents)

**Mission**: Complete platform deployment with all workstreams

**Composition:**

```yaml
Team: Full Platform
Size: 12-15 agents
Model Distribution: 2-3 Opus, 10-12 Sonnet
Coordination: Hierarchical (coordinator → team leads → workers)
```

**Structure:**
- 1 hierarchical-coordinator (Opus)
- Platform Team (6-8 agents)
- Observability Team (5-7 agents)
- DevOps Team (4-5 agents)

**Coordination:**

```javascript
// Top-level coordination
Task("Hierarchical Coordinator", `
Coordinate full platform deployment:
1. Initialize swarm with hierarchical topology, max 15 agents
2. Spawn team leads: Kubernetes Architect, OBI Specialist, DevOps Automation
3. Define workstreams:
   - Platform: Infrastructure and Kubernetes
   - Observability: OBI, Grafana, experiments
   - DevOps: CI/CD, testing, validation
4. Set dependencies:
   - Observability depends on Platform
   - DevOps can run parallel to others
5. Monitor progress across all teams
6. Resolve inter-team dependencies and conflicts
7. Generate status reports every 30 minutes
8. Coordinate final integration testing
`, "hierarchical-coordinator")

// Team leads coordinate their teams (as shown in previous sections)
// All team members work in parallel within their workstreams
```

**Execution Waves:**

**Wave 1: Platform Foundation (Parallel)**
- All platform engineers work simultaneously on different environments
- Terraform specialist provisions cloud resources
- Target: Complete in 45-60 minutes

**Wave 2: Observability Stack (Parallel, depends on Wave 1)**
- All Grafana specialists deploy stack components simultaneously
- OBI specialist designs experiments
- Alloy specialist configures pipelines
- Target: Complete in 30-45 minutes

**Wave 3: DevOps & Validation (Parallel, can start with Wave 1)**
- DevOps automation sets up Tiltfile
- CI/CD engineer builds pipelines
- Tester creates test suites
- Target: Complete in 30-45 minutes

**Wave 4: Integration (Sequential)**
- Production validator checks all environments
- Tester runs integration tests
- All teams coordinate for final validation
- Target: Complete in 15-30 minutes

**Total Time: 2-3 hours** (vs 8-12 hours sequential)

---

## Team Coordination Patterns

### Pattern 1: Mesh Coordination (Within Teams)

**Best for:** Teams of 4-8 agents working on related tasks

**Characteristics:**
- Peer-to-peer communication
- Shared memory for coordination
- No strict hierarchy
- Fast decision making

**Implementation:**
```javascript
// Each team member has equal status
mcp__claude-flow__swarm_init { topology: "mesh", maxAgents: 8 }

// Agents coordinate through memory
mcp__claude-flow__memory_usage {
  action: "store",
  key: "swarm/team/shared-decision",
  value: JSON.stringify({ decision: "use S3 for storage" })
}
```

### Pattern 2: Hierarchical Coordination (Between Teams)

**Best for:** Multiple teams (12+ agents) with dependencies

**Characteristics:**
- Clear leadership chain
- Team leads coordinate with coordinator
- Structured decision flow
- Scalable to 50+ agents

**Implementation:**
```javascript
// Top-level coordinator
mcp__claude-flow__swarm_init { topology: "hierarchical", maxAgents: 15 }

// Coordinator spawns team leads
mcp__claude-flow__agent_spawn { type: "kubernetes-architect", role: "team-lead" }
mcp__claude-flow__agent_spawn { type: "obi-specialist", role: "team-lead" }

// Team leads spawn their team members
// Coordinator manages dependencies between teams
```

### Pattern 3: Adaptive Coordination (Dynamic)

**Best for:** Complex projects with changing requirements

**Characteristics:**
- Topology changes based on needs
- Agents join/leave dynamically
- Self-organizing teams
- Handles uncertainty well

**Implementation:**
```javascript
mcp__claude-flow__swarm_init { topology: "adaptive", maxAgents: 20 }

// System adapts coordination based on:
// - Task complexity
// - Agent availability
// - Dependency patterns
// - Performance metrics
```

---

## Team Size Guidelines

### Small Team (4-5 agents)
- **Use case:** Single workstream, focused task
- **Example:** Deploy OBI to dev environment
- **Coordination:** Mesh
- **Time to complete:** 30-60 minutes

### Medium Team (6-10 agents)
- **Use case:** Multiple related workstreams
- **Example:** Full observability stack deployment
- **Coordination:** Mesh with optional lead
- **Time to complete:** 1-2 hours

### Large Team (12-20 agents)
- **Use case:** Complete platform with all workstreams
- **Example:** Multi-cluster MOP deployment
- **Coordination:** Hierarchical
- **Time to complete:** 2-4 hours

### Very Large Team (20+ agents)
- **Use case:** Multi-project or complex migration
- **Example:** Migrate entire infrastructure to new platform
- **Coordination:** Hierarchical + Adaptive
- **Time to complete:** 4-8 hours

---

## Model Distribution Strategies

### Budget-Conscious (Minimize Opus Usage)

```yaml
Opus: 1 agent (5-10%)
Sonnet: 9-19 agents (90-95%)
Strategy: Single Opus as overall architect/coordinator
Cost: $$
```

**Best for:**
- Cost-sensitive projects
- Implementation-heavy tasks
- Clear requirements

### Balanced (Standard)

```yaml
Opus: 2-3 agents (15-20%)
Sonnet: 12-18 agents (80-85%)
Strategy: Opus for team leads and key architects
Cost: $$$
```

**Best for:**
- Most MOP projects
- Balance of strategy and implementation
- Standard complexity

### Quality-Focused (Maximize Expertise)

```yaml
Opus: 4-5 agents (25-30%)
Sonnet: 10-15 agents (70-75%)
Strategy: Opus for all leads and complex decisions
Cost: $$$$
```

**Best for:**
- High-stakes projects
- Complex architecture
- Quality-critical work

---

## Dependency Management

### Independent Teams (Parallel)

**DevOps + Platform** (can run simultaneously)
- DevOps doesn't depend on Platform for initial work
- Both can work on their domains independently

### Sequential Dependencies

**Platform → Observability**
- Observability team MUST wait for Platform completion
- Use memory checks: `swarm/platform/dev-ready`, etc.

**Observability → Testing**
- Integration tests MUST wait for Observability deployment
- Use memory checks: `swarm/observability/deployed`

### Partial Dependencies

**Platform (90% complete) → Observability (start)**
- Observability can start when namespace is ready
- Don't wait for complete Platform finish
- Use granular memory keys: `swarm/platform/namespace-ready`

---

## Performance Metrics

### Expected Speed Improvements

| Team Size | Sequential Time | Parallel Time | Speedup |
|-----------|-----------------|---------------|---------|
| 4 agents  | 2 hours        | 45 minutes    | 2.7x    |
| 8 agents  | 4 hours        | 1.5 hours     | 2.7x    |
| 12 agents | 6 hours        | 2 hours       | 3.0x    |
| 15 agents | 8 hours        | 2.5 hours     | 3.2x    |

### Token Usage Optimization

- **Mesh coordination**: 10-15% token overhead
- **Hierarchical coordination**: 20-25% token overhead
- **Memory usage**: 5% token overhead
- **Hook execution**: 8% token overhead

**Net result**: Despite overhead, 2.8-4.4x faster execution

---

## Best Practices

### 1. Team Composition
- Start with small teams (4-5 agents)
- Scale up only when needed
- Keep related work in same team
- Use Opus strategically (leadership roles)

### 2. Coordination
- Mesh within teams (fast decisions)
- Hierarchical between teams (clear dependencies)
- Use memory for all coordination
- Execute hooks consistently

### 3. Parallel Execution
- Identify independent workstreams
- Launch all agents in ONE message
- Use memory for dependency checks
- Monitor progress centrally

### 4. Model Distribution
- 1 Opus per team as lead (small teams)
- 2-3 Opus for large deployments (team leads + coordinator)
- Sonnet for all implementation work
- Avoid Haiku (not needed for MOP complexity)

### 5. Dependency Management
- Map dependencies before starting
- Use granular memory keys
- Don't over-serialize (partial dependencies OK)
- Have fallback strategies

---

## Example: Complete Platform Deployment

### Launch Command (Single Message)

```javascript
// ONE message spawns entire 15-agent team
Task("Hierarchical Coordinator", "Coordinate full deployment...", "hierarchical-coordinator")

// Platform Team (6 agents - parallel)
Task("Kubernetes Architect", "Lead platform team...", "kubernetes-architect")
Task("Platform Engineer 1", "Implement dev environment...", "platform-engineer")
Task("Platform Engineer 2", "Implement staging environment...", "platform-engineer")
Task("Platform Engineer 3", "Implement prod environment...", "platform-engineer")
Task("Terraform Specialist", "Provision cloud resources...", "terraform-specialist")
Task("Production Validator", "Validate all environments...", "production-validator")

// Observability Team (5 agents - parallel, depends on platform)
Task("OBI Specialist", "Lead observability team...", "obi-specialist")
Task("Grafana Specialist 1", "Deploy Tempo and Mimir...", "grafana-specialist")
Task("Grafana Specialist 2", "Deploy Loki and Grafana...", "grafana-specialist")
Task("Alloy Specialist", "Configure pipelines...", "alloy-specialist")
Task("Experiment Designer", "Design experiments...", "experiment-designer")

// DevOps Team (4 agents - parallel, independent)
Task("DevOps Automation", "Setup local dev environment...", "devops-automation")
Task("CI/CD Engineer", "Build CI/CD pipelines...", "cicd-engineer")
Task("Tester", "Create test suite...", "tester")
Task("DevOps Troubleshooter", "Monitor and debug...", "devops-troubleshooter")

// All agents coordinate through hooks and memory
// Complete deployment in 2-3 hours (vs 8-12 hours sequential)
```

---

## Summary

Effective team composition requires:
1. ✅ Right-sized teams (4-8 agents per team)
2. ✅ Strategic Opus usage (15-20% for leads)
3. ✅ Clear workstream boundaries
4. ✅ Appropriate coordination (mesh vs hierarchical)
5. ✅ Parallel execution where possible
6. ✅ Dependency management through memory
7. ✅ Consistent hook execution

Follow these patterns to achieve **2.8-4.4x speed improvements** while maintaining quality and coordination.
