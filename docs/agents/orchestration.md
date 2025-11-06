# Orchestration Guide

This document defines event-driven orchestration patterns for the MOP project, enabling automated agent spawning, health monitoring, and adaptive coordination.

---

## Overview

Orchestration in the MOP project uses event-driven patterns to:
- Automatically spawn agents based on triggers
- Monitor agent health and performance
- Adapt coordination topology dynamically
- Provide real-time dashboards
- Enable Q&A workflows with humans

---

## Event-Driven Architecture

### Event Types

```yaml
Event Categories:
  - file_events: File creation, modification, deletion
  - agent_events: Agent spawn, complete, error, health
  - task_events: Task start, progress, complete, fail
  - system_events: Resource limits, bottlenecks, errors
  - user_events: Questions, approvals, feedback
```

### Event Flow

```
Trigger → Event → Orchestrator → Decision → Action → Outcome → Dashboard
```

**Example:**
```
File Created (obi-experiment.yaml)
  ↓
Event: "new_experiment_config"
  ↓
Orchestrator: Check if experiment needs deployment
  ↓
Decision: YES - Deploy experiment
  ↓
Action: Spawn obi-specialist agent
  ↓
Outcome: Experiment deployed
  ↓
Dashboard: Update status "Experiment Active"
```

---

## Spawn Trigger Detection

### Automatic Agent Spawning

The orchestrator watches for triggers that require agent spawning:

#### Trigger 1: New File Requires Processing

**Pattern:**
```yaml
trigger:
  type: file_created
  path_pattern: "environments/*/observability/obi-*.yaml"
  action: spawn_agent
  agent_type: obi-specialist
  task: "Deploy OBI configuration"
```

**Implementation:**
```javascript
// Orchestrator watches for file events
mcp__claude-flow__watch_files {
  patterns: [
    "environments/*/observability/obi-*.yaml",
    "environments/*/kubernetes/*.yaml",
    "experiments/*.yaml"
  ],
  on_event: "check_spawn_trigger"
}

// When file created, evaluate if agent needed
function checkSpawnTrigger(event) {
  if (event.type === "file_created" && event.path.includes("obi-")) {
    // Spawn OBI specialist to process new config
    mcp__claude-flow__agent_spawn {
      type: "obi-specialist",
      task: `Deploy OBI configuration from ${event.path}`,
      priority: "high",
      auto_start: true
    }
  }
}
```

#### Trigger 2: Task Requires Expertise

**Pattern:**
```yaml
trigger:
  type: task_created
  requires_skill: "kubernetes"
  complexity: "high"
  action: spawn_agent
  agent_type: kubernetes-architect
```

**Implementation:**
```javascript
// Task orchestrator analyzes new tasks
mcp__claude-flow__task_orchestrate {
  on_task_created: (task) => {
    const complexity = analyzeComplexity(task);
    const requiredSkills = identifySkills(task);

    if (complexity === "high" && requiredSkills.includes("kubernetes")) {
      // Spawn architect for complex K8s tasks
      mcp__claude-flow__agent_spawn {
        type: "kubernetes-architect",
        task: task.description,
        model: "opus"  // Use Opus for complex decisions
      }
    }
  }
}
```

#### Trigger 3: Dependency Chain Activation

**Pattern:**
```yaml
trigger:
  type: dependency_ready
  wait_for: "swarm/platform/namespace-ready"
  action: spawn_dependent_agents
  agents:
    - obi-specialist
    - grafana-specialist
    - alloy-specialist
```

**Implementation:**
```javascript
// Memory watcher triggers on dependency completion
mcp__claude-flow__memory_watch {
  key: "swarm/platform/namespace-ready",
  on_update: (value) => {
    if (value.status === "complete") {
      // Spawn entire observability team
      [
        "obi-specialist",
        "grafana-specialist",
        "alloy-specialist"
      ].forEach(agentType => {
        Task(`${agentType}`, `Deploy observability stack...`, agentType)
      });
    }
  }
}
```

#### Trigger 4: Error Recovery

**Pattern:**
```yaml
trigger:
  type: agent_error
  error_type: "deployment_failed"
  action: spawn_troubleshooter
  agent_type: devops-troubleshooter
```

**Implementation:**
```javascript
// Error handler spawns troubleshooter
mcp__claude-flow__agent_monitor {
  on_error: (agent, error) => {
    if (error.type === "deployment_failed") {
      mcp__claude-flow__agent_spawn {
        type: "devops-troubleshooter",
        task: `Debug failed deployment for ${agent.id}: ${error.message}`,
        priority: "urgent",
        context: {
          failed_agent: agent.id,
          error_details: error,
          logs: agent.logs
        }
      }
    }
  }
}
```

---

## Agent Health Monitoring

### Health Metrics

Each agent reports health metrics:

```javascript
mcp__claude-flow__agent_health_report {
  agent_id: "obi-specialist-001",
  metrics: {
    status: "healthy",           // healthy | degraded | unhealthy
    cpu_usage: 45,                // percentage
    memory_usage: 62,             // percentage
    task_progress: 75,            // percentage
    error_count: 0,
    last_activity: Date.now(),
    response_time: 1250           // milliseconds
  }
}
```

### Health Check Intervals

```yaml
health_checks:
  interval: 30s
  timeout: 10s
  failure_threshold: 3
  success_threshold: 1
```

### Health States

#### Healthy
- Response time < 2s
- Error count = 0
- Making progress
- Resource usage < 80%

**Action:** Continue normal operation

#### Degraded
- Response time 2-5s
- Error count 1-2
- Slow progress
- Resource usage 80-95%

**Actions:**
- Increase monitoring frequency
- Notify coordinator
- Consider task reassignment

#### Unhealthy
- Response time > 5s or no response
- Error count > 2
- No progress for >10 minutes
- Resource usage > 95%

**Actions:**
- Spawn troubleshooter agent
- Reassign critical tasks
- Restart agent if possible
- Escalate to human

### Monitoring Implementation

```javascript
// Continuous health monitoring
mcp__claude-flow__swarm_monitor {
  session_id: "swarm-mop-platform",
  check_interval: 30,  // seconds
  on_health_change: (agent, old_state, new_state) => {
    if (new_state === "degraded") {
      console.warn(`Agent ${agent.id} is degraded`);
      // Notify coordinator
      mcp__claude-flow__notify {
        target: "coordinator",
        message: `Agent ${agent.id} performance degraded`,
        severity: "warning"
      }
    }

    if (new_state === "unhealthy") {
      console.error(`Agent ${agent.id} is unhealthy`);
      // Spawn troubleshooter
      Task("DevOps Troubleshooter", `
        Debug unhealthy agent ${agent.id}:
        - Check resource usage
        - Review error logs
        - Identify bottlenecks
        - Recommend remediation
      `, "devops-troubleshooter")
    }
  }
}
```

---

## State Transitions

### Agent State Machine

```
SPAWNING → INITIALIZING → AVAILABLE → ASSIGNED → ACTIVE → COMPLETE
                                ↓           ↓         ↓
                            PAUSED ← → DEGRADED → UNHEALTHY → FAILED
```

#### State Definitions

**SPAWNING**
- Agent creation in progress
- Resources being allocated
- Context being loaded

**INITIALIZING**
- Executing pre-task hooks
- Loading dependencies
- Checking prerequisites

**AVAILABLE**
- Ready to receive tasks
- Waiting in agent pool
- Consuming minimal resources

**ASSIGNED**
- Task allocated
- Preparing to execute
- Locking resources

**ACTIVE**
- Working on task
- Making progress
- Reporting metrics

**PAUSED**
- Temporarily suspended
- Waiting for dependency
- Can resume quickly

**DEGRADED**
- Performance issues
- Increased error rate
- Still functional

**UNHEALTHY**
- Severe issues
- Not making progress
- Needs intervention

**COMPLETE**
- Task finished successfully
- Resources released
- Output delivered

**FAILED**
- Task failed unrecoverably
- Resources released
- Error reported

### Transition Triggers

```yaml
Transitions:
  SPAWNING → INITIALIZING:
    trigger: agent_created

  INITIALIZING → AVAILABLE:
    trigger: pre_task_hook_complete

  AVAILABLE → ASSIGNED:
    trigger: task_assigned

  ASSIGNED → ACTIVE:
    trigger: task_started

  ACTIVE → COMPLETE:
    trigger: post_task_hook_complete

  ACTIVE → PAUSED:
    trigger: dependency_not_ready

  PAUSED → ACTIVE:
    trigger: dependency_ready

  ACTIVE → DEGRADED:
    trigger: performance_threshold_exceeded

  DEGRADED → ACTIVE:
    trigger: performance_recovered

  DEGRADED → UNHEALTHY:
    trigger: degradation_timeout

  ACTIVE → FAILED:
    trigger: unrecoverable_error
```

### State Transition Handlers

```javascript
mcp__claude-flow__state_machine {
  agent_id: "obi-specialist-001",
  on_transition: (from_state, to_state, context) => {
    console.log(`Agent transition: ${from_state} → ${to_state}`);

    // Handle specific transitions
    if (to_state === "DEGRADED") {
      // Increase monitoring
      increaseMonitoringFrequency(context.agent_id);

      // Notify coordinator
      notifyCoordinator({
        agent: context.agent_id,
        issue: "performance_degraded",
        metrics: context.metrics
      });
    }

    if (to_state === "UNHEALTHY") {
      // Spawn troubleshooter
      spawnTroubleshooter(context.agent_id);

      // Consider task reassignment
      if (context.task.priority === "critical") {
        reassignTask(context.task);
      }
    }

    if (to_state === "COMPLETE") {
      // Release resources
      releaseResources(context.agent_id);

      // Trigger dependent tasks
      triggerDependentTasks(context.task.id);

      // Update dashboard
      updateDashboard({
        agent: context.agent_id,
        status: "complete",
        output: context.output
      });
    }
  }
}
```

---

## Dashboard Updates

### Real-Time Dashboard

The orchestrator maintains a real-time dashboard showing:

```
╔══════════════════════════════════════════════════════════════╗
║              MOP Platform Deployment Dashboard                ║
╠══════════════════════════════════════════════════════════════╣
║ Session: swarm-mop-platform                                  ║
║ Started: 2025-11-06 14:30:00                                 ║
║ Elapsed: 1h 15m                                              ║
║ Progress: ████████████████░░░░░░░░░░ 67%                     ║
╠══════════════════════════════════════════════════════════════╣
║ Platform Team (6 agents)                          75% [████] ║
║   ✅ platform-engineer-dev           COMPLETE    100%        ║
║   ✅ platform-engineer-staging       COMPLETE    100%        ║
║   🔄 platform-engineer-prod          ACTIVE       85%        ║
║   🔄 terraform-specialist            ACTIVE       60%        ║
║   ⏳ production-validator            PENDING       0%        ║
║   📋 kubernetes-architect            REVIEWING   100%        ║
╠══════════════════════════════════════════════════════════════╣
║ Observability Team (5 agents)                    58% [███░] ║
║   🔄 obi-specialist                  ACTIVE       45%        ║
║   🔄 grafana-specialist-1            ACTIVE       65%        ║
║   🔄 grafana-specialist-2            ACTIVE       55%        ║
║   ⏳ alloy-specialist                PENDING       0%        ║
║   🔄 experiment-designer             ACTIVE       80%        ║
╠══════════════════════════════════════════════════════════════╣
║ DevOps Team (4 agents)                            83% [████] ║
║   ✅ devops-automation               COMPLETE    100%        ║
║   🔄 cicd-engineer                   ACTIVE       90%        ║
║   🔄 tester                          ACTIVE       75%        ║
║   🔄 devops-troubleshooter           ACTIVE       65%        ║
╠══════════════════════════════════════════════════════════════╣
║ Recent Events:                                                ║
║   14:45 - platform-engineer-dev completed namespace creation  ║
║   15:10 - obi-specialist started experiment deployment        ║
║   15:22 - grafana-specialist-1 deployed Tempo                 ║
║   15:35 - tester achieved 80% test coverage                   ║
║   15:42 - terraform-specialist completed staging provision    ║
╠══════════════════════════════════════════════════════════════╣
║ Issues: 0 critical, 1 warning, 0 info                         ║
║   ⚠️  alloy-specialist waiting for obi-deployed dependency    ║
╠══════════════════════════════════════════════════════════════╣
║ Estimated Completion: 2025-11-06 16:00:00 (45 minutes)       ║
╚══════════════════════════════════════════════════════════════╝
```

### Dashboard Implementation

```javascript
// Initialize dashboard
mcp__claude-flow__dashboard_init {
  session_id: "swarm-mop-platform",
  refresh_interval: 10,  // seconds
  display_mode: "terminal"  // or "web"
}

// Update dashboard on events
mcp__claude-flow__dashboard_update {
  session_id: "swarm-mop-platform",
  event_type: "agent_progress",
  data: {
    agent_id: "obi-specialist-001",
    status: "ACTIVE",
    progress: 45,
    current_task: "Deploying OBI experiments"
  }
}

// Add event to dashboard
mcp__claude-flow__dashboard_event {
  session_id: "swarm-mop-platform",
  timestamp: Date.now(),
  message: "OBI specialist started experiment deployment",
  severity: "info"
}
```

### Dashboard Sections

#### 1. Overview
- Session ID and metadata
- Overall progress percentage
- Elapsed time
- Estimated completion

#### 2. Team Status
- Team name and size
- Team progress percentage
- Individual agent status
- Current task for each agent

#### 3. Recent Events
- Chronological event log
- Completion notifications
- Error alerts
- Dependency notifications

#### 4. Issues
- Critical issues (immediate action required)
- Warnings (attention needed)
- Info (FYI only)

#### 5. Metrics
- Total agents
- Active/complete/pending counts
- Resource usage
- Estimated completion time

---

## Q&A Workflow

### Human-in-the-Loop Decision Making

For critical decisions, the orchestrator can pause and ask humans:

#### Question Types

**1. Approval Required**
```javascript
mcp__claude-flow__ask_human {
  session_id: "swarm-mop-platform",
  question: "Ready to deploy to production?",
  question_type: "approval",
  context: {
    environment: "production",
    changes: [
      "OBI experiments with eBPF probes",
      "Grafana stack (Tempo, Mimir, Loki)",
      "Alloy data pipelines"
    ],
    validation_results: {
      tests_passing: true,
      security_audit: "PASS",
      resource_limits: "OK"
    }
  },
  timeout: 3600,  // 1 hour
  required: true
}
```

**2. Choice Selection**
```javascript
mcp__claude-flow__ask_human {
  question: "Which storage backend for Tempo?",
  question_type: "choice",
  options: [
    { value: "s3", label: "AWS S3", description: "Scalable, cost-effective" },
    { value: "gcs", label: "Google Cloud Storage", description: "Better for GCP" },
    { value: "local", label: "Local Storage", description: "Dev only" }
  ],
  default: "s3",
  timeout: 1800
}
```

**3. Configuration Input**
```javascript
mcp__claude-flow__ask_human {
  question: "Enter Grafana admin password:",
  question_type: "input",
  input_type: "password",
  validation: {
    min_length: 12,
    require_special: true
  },
  required: true
}
```

**4. Problem Resolution**
```javascript
mcp__claude-flow__ask_human {
  question: "OBI deployment failed in prod. How to proceed?",
  question_type: "choice",
  context: {
    error: "Insufficient CPU quota",
    current_quota: "50 CPU",
    required: "75 CPU"
  },
  options: [
    { value: "increase_quota", label: "Increase CPU quota to 75" },
    { value: "reduce_replicas", label: "Reduce OBI replicas from 3 to 2" },
    { value: "skip_prod", label: "Skip prod deployment for now" },
    { value: "manual_fix", label: "I'll fix manually" }
  ],
  on_answer: (answer) => {
    if (answer === "increase_quota") {
      // Update resource quotas
      Task("Platform Engineer", "Increase prod CPU quota to 75", "platform-engineer")
    } else if (answer === "reduce_replicas") {
      // Modify deployment
      Task("OBI Specialist", "Deploy OBI with 2 replicas", "obi-specialist")
    }
  }
}
```

### Q&A Implementation

```javascript
// Orchestrator pauses for human input
function askHumanForApproval(context) {
  console.log("⏸️  Pausing workflow for human approval");

  const response = mcp__claude-flow__ask_human({
    question: context.question,
    question_type: "approval",
    context: context.data,
    timeout: 3600
  });

  if (response.approved) {
    console.log("✅ Approved - continuing workflow");
    continueWorkflow(context);
  } else {
    console.log("❌ Rejected - stopping workflow");
    stopWorkflow(context, response.reason);
  }
}

// Example: Prod deployment approval
if (environment === "production" && !context.auto_deploy) {
  askHumanForApproval({
    question: "Deploy to production?",
    data: {
      changes: getChangesSummary(),
      tests: getTestResults(),
      security: getSecurityAudit()
    }
  });
}
```

---

## Adaptive Coordination

### Dynamic Topology Adjustment

The orchestrator can change coordination topology based on workload:

```javascript
mcp__claude-flow__adapt_topology {
  session_id: "swarm-mop-platform",
  trigger: "bottleneck_detected",
  current_topology: "hierarchical",
  proposed_topology: "mesh",
  reason: "Too much coordinator overhead, teams working independently",
  on_adapt: () => {
    console.log("Switching from hierarchical to mesh coordination");
    // Reconfigure agent communication
    agents.forEach(agent => {
      agent.enablePeerCommunication();
      agent.disableHierarchicalReporting();
    });
  }
}
```

### Adaptive Triggers

**1. Bottleneck Detection**
- One agent blocking many others
- Action: Add more agents or redistribute work

**2. Resource Constraints**
- CPU/memory usage > 90%
- Action: Pause non-critical agents

**3. Dependency Deadlock**
- Circular dependencies detected
- Action: Break cycle or reorganize tasks

**4. Performance Degradation**
- Overall progress < expected
- Action: Increase parallelism or optimize tasks

### Adaptation Strategies

```javascript
mcp__claude-flow__adaptation_strategies {
  strategies: [
    {
      name: "scale_up",
      trigger: "progress_slow",
      action: "spawn_additional_agents",
      max_agents: 20
    },
    {
      name: "scale_down",
      trigger: "resource_constrained",
      action: "pause_non_critical_agents",
      keep_critical: ["platform-engineer", "obi-specialist"]
    },
    {
      name: "redistribute",
      trigger: "agent_overloaded",
      action: "move_tasks_to_available_agents"
    },
    {
      name: "change_topology",
      trigger: "coordination_overhead_high",
      action: "switch_topology",
      from: "hierarchical",
      to: "mesh"
    }
  ]
}
```

---

## Neural Pattern Training

### Learning from Successful Workflows

The orchestrator trains neural patterns from successful workflows:

```javascript
mcp__claude-flow__neural_train {
  session_id: "swarm-mop-platform",
  pattern_type: "workflow",
  input_data: {
    task: "deploy_complete_platform",
    team_composition: {
      platform: 6,
      observability: 5,
      devops: 4
    },
    execution_time: 7200,  // 2 hours
    success: true
  },
  output_prediction: {
    optimal_team_size: 15,
    estimated_duration: 7200,
    recommended_topology: "hierarchical",
    critical_path: ["platform", "observability", "testing"]
  }
}

// Later, use trained model for predictions
mcp__claude-flow__neural_predict {
  model: "workflow_optimizer",
  input: {
    task: "deploy_complete_platform",
    constraints: {
      max_time: 10800,  // 3 hours
      budget: "medium"
    }
  },
  on_prediction: (result) => {
    console.log(`Recommended team size: ${result.team_size}`);
    console.log(`Estimated duration: ${result.duration / 60} minutes`);
    console.log(`Topology: ${result.topology}`);
  }
}
```

---

## Orchestration Best Practices

### 1. Event-Driven Design
- Use events for loose coupling
- React to state changes
- Enable async workflows
- Support dynamic adaptation

### 2. Health Monitoring
- Monitor all agents continuously
- Define clear health thresholds
- Auto-remediate when possible
- Escalate to humans when needed

### 3. Dashboard Visibility
- Show real-time progress
- Display recent events
- Highlight issues prominently
- Provide ETA estimates

### 4. Human-in-the-Loop
- Ask for approval on critical changes
- Provide context with questions
- Set reasonable timeouts
- Have fallback decisions

### 5. Adaptive Coordination
- Monitor coordination overhead
- Detect bottlenecks early
- Adjust topology dynamically
- Learn from successful workflows

### 6. Neural Training
- Train on successful workflows
- Predict optimal configurations
- Continuously improve
- Use predictions for planning

---

## Example: Complete Orchestration

### Scenario: Deploy MOP Platform with Full Orchestration

```javascript
// 1. Initialize orchestration
mcp__claude-flow__orchestrate_session_start {
  session_id: "swarm-mop-platform",
  project: "Multi-cluster Observability Platform",
  goal: "Deploy complete MOP infrastructure",
  topology: "adaptive",  // Will adapt based on workload
  max_agents: 20,
  auto_spawn: true,       // Enable automatic agent spawning
  monitoring_interval: 30,
  dashboard_enabled: true
}

// 2. Define spawn triggers
mcp__claude-flow__orchestrate_triggers {
  triggers: [
    {
      name: "platform_namespace_ready",
      type: "memory_key",
      key: "swarm/platform/namespace-ready",
      action: "spawn_observability_team"
    },
    {
      name: "deployment_error",
      type: "agent_error",
      error_pattern: "deployment_failed",
      action: "spawn_troubleshooter"
    },
    {
      name: "prod_deployment",
      type: "environment_deploy",
      environment: "production",
      action: "ask_human_approval"
    }
  ]
}

// 3. Launch initial wave (platform team)
Task("Kubernetes Architect", "Design and lead platform deployment", "kubernetes-architect")
Task("Platform Engineer - Dev", "Configure dev environment", "platform-engineer")
Task("Platform Engineer - Staging", "Configure staging environment", "platform-engineer")
Task("Platform Engineer - Prod", "Configure prod environment", "platform-engineer")

// 4. Orchestrator automatically:
//    - Monitors health of all agents
//    - Updates dashboard every 30 seconds
//    - Spawns observability team when namespace ready
//    - Spawns troubleshooter on errors
//    - Asks for approval before prod deploy
//    - Adapts topology if bottlenecks detected
//    - Trains neural patterns on success

// 5. Human receives notification when approval needed
// Dashboard shows:
//   ⏸️  Workflow paused - awaiting approval for production deployment
//   📋 Review changes and test results before approving

// 6. After approval, deployment continues
// Dashboard updates in real-time with progress

// 7. On completion, orchestrator:
//    - Generates summary report
//    - Trains neural model
//    - Saves session state
//    - Provides metrics and insights
```

---

## Summary

Effective orchestration requires:

1. ✅ **Event-driven architecture** - React to changes, don't poll
2. ✅ **Automatic agent spawning** - Spawn based on triggers, not manually
3. ✅ **Continuous health monitoring** - Detect issues early
4. ✅ **Real-time dashboards** - Provide visibility to humans
5. ✅ **Human-in-the-loop** - Ask for critical decisions
6. ✅ **Adaptive coordination** - Adjust topology dynamically
7. ✅ **Neural learning** - Improve from experience

Follow these patterns to create self-managing, adaptive agent swarms that achieve optimal performance while maintaining human oversight for critical decisions.
