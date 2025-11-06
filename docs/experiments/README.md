# MOP Experiment Configurations

This directory contains experiment configurations for the Observability-by-Inference (OBI) framework.

## Available Experiments

### 1. Adaptive Sampling (`adaptive-sampling.json`)
**Objective**: Reduce ingestion costs while maintaining query quality

**Changes**:
- Reduce sample rate to 50% on mimir-distributor

**Expected Impact**:
- 30-50% cost reduction
- Minimal (<5%) query impact
- 30-40% storage savings

**When to Run**:
- Anytime - low risk
- Best during normal traffic periods to get representative data

**Duration**: 1-2 hours minimum

---

### 2. Compaction Tuning (`compaction-tuning.json`)
**Objective**: Optimize storage compaction for better CPU efficiency

**Changes**:
- Increase compaction interval from 15m to 30m

**Expected Impact**:
- 20-30% CPU reduction on compactors
- Temporary increase in block count
- Minimal (<5%) query impact

**When to Run**:
- During low-query periods
- Monitor storage capacity

**Duration**: 2-4 hours to see full compaction cycle impact

---

### 3. Ingester Scaling (`ingester-scaling.json`)
**Objective**: Test reduced ingester count during low-traffic periods

**Changes**:
- Reduce ingester replicas from 6 to 3

**Expected Impact**:
- 50% cost reduction
- Increased resource utilization (within safe limits)
- No ingestion impact

**When to Run**:
- **Only during confirmed low-traffic periods**
- Not recommended for production initially

**Duration**: 1 hour minimum, 4 hours recommended

---

## Running Experiments

### Basic Usage
```bash
# Run experiment in development
cd /Users/beengud/raibid-labs/mop
./scripts/nu/experiment-runner.nu --config docs/experiments/adaptive-sampling.json --env dev

# Run with auto-rollback on degradation
./scripts/nu/experiment-runner.nu \
  --config docs/experiments/compaction-tuning.json \
  --env staging \
  --auto-rollback

# Extended experiment with custom duration
./scripts/nu/experiment-runner.nu \
  --config docs/experiments/ingester-scaling.json \
  --env dev \
  --duration 7200 \
  --baseline-duration 600
```

### Recommended Workflow

1. **Review Configuration**
   ```bash
   cat docs/experiments/adaptive-sampling.json | from json
   ```

2. **Test in Development**
   ```bash
   ./scripts/nu/experiment-runner.nu \
     --config docs/experiments/adaptive-sampling.json \
     --env dev \
     --export results/dev-test.json
   ```

3. **Validate in Staging**
   ```bash
   ./scripts/nu/experiment-runner.nu \
     --config docs/experiments/adaptive-sampling.json \
     --env staging \
     --auto-rollback \
     --export results/staging-test.json
   ```

4. **Deploy to Production** (if successful)
   ```bash
   # Run as experiment first with auto-rollback
   ./scripts/nu/experiment-runner.nu \
     --config docs/experiments/adaptive-sampling.json \
     --env prod \
     --auto-rollback \
     --export results/prod-test.json

   # If successful, apply permanently
   ./scripts/nu/deploy.nu --env prod --component mimir-distributor
   ```

---

## Experiment Configuration Format

```json
{
  "name": "Experiment Name",
  "description": "What this experiment tests",
  "author": "Your Name",
  "version": "1.0.0",

  "changes": [
    {
      "type": "deployment | configmap",
      "component": "kubernetes-resource-name",
      "container": "container-name",
      "parameter": "ENV_VAR or config key",
      "value": "new value",
      "description": "What this change does"
    }
  ],

  "success_metrics": [
    {
      "name": "metric_name",
      "query": "PromQL query",
      "direction": "lower | higher",
      "threshold": 100,
      "description": "What this metric measures"
    }
  ],

  "expected_outcomes": {
    "primary_goal": "expected result",
    "secondary_goal": "expected result"
  },

  "rollback_criteria": {
    "metric_name": "condition"
  },

  "notes": [
    "Important considerations",
    "Warnings and precautions"
  ]
}
```

### Change Types

**`deployment`**
- Modifies deployment environment variables
- Triggers rolling update
- Can change: resource limits, env vars, replicas

**`configmap`**
- Updates ConfigMap values
- May require pod restart
- Can change: configuration parameters

### Metric Directions

**`lower`** - Lower values are better
- Latency metrics
- Error rates
- Cost metrics
- Resource usage (when optimizing)

**`higher`** - Higher values are better
- Throughput metrics
- Availability metrics
- Quality metrics
- Efficiency metrics

### Success Criteria

Experiments are evaluated with a scoring system:
- **Score ≥ 0.8**: `adopt` - Clear improvement, safe to deploy
- **Score ≥ 0.5**: `investigate` - Mixed results, needs analysis
- **Score < 0.5**: `rollback` - Degradation detected, revert changes

Score is calculated based on:
- Did metrics improve in the expected direction?
- Were thresholds met?
- Percentage improvement vs baseline

---

## Creating Custom Experiments

### 1. Start with Template
```bash
cp docs/experiments/adaptive-sampling.json docs/experiments/my-experiment.json
```

### 2. Modify Configuration
- Update name and description
- Define changes to apply
- Specify success metrics with PromQL queries
- Set appropriate thresholds
- Document expected outcomes

### 3. Validate Configuration
```bash
# Check JSON syntax
cat docs/experiments/my-experiment.json | from json

# Test in development first
./scripts/nu/experiment-runner.nu \
  --config docs/experiments/my-experiment.json \
  --env dev
```

### 4. Document Results
```bash
# Export results for analysis
./scripts/nu/experiment-runner.nu \
  --config docs/experiments/my-experiment.json \
  --env dev \
  --export results/my-experiment-$(date +%Y%m%d).json
```

---

## Best Practices

### Before Running Experiments

1. **Understand the Change**
   - Know what you're modifying
   - Understand potential impact
   - Have rollback plan ready

2. **Set Appropriate Duration**
   - Minimum 1 hour for meaningful data
   - Longer for subtle changes (2-4 hours)
   - Consider traffic patterns

3. **Choose Right Environment**
   - Start in development
   - Validate in staging
   - Production only after success

### During Experiments

1. **Monitor Actively**
   - Watch experiment output
   - Check dashboards
   - Review logs if issues arise

2. **Document Observations**
   - Note unexpected behavior
   - Record metric changes
   - Capture timestamps

### After Experiments

1. **Analyze Results**
   - Review all metrics
   - Compare to baseline
   - Check recommendations

2. **Make Decision**
   - Adopt if clearly beneficial
   - Investigate if inconclusive
   - Rollback if degraded

3. **Document Learnings**
   - Update experiment notes
   - Share with team
   - Archive results

---

## Safety Guidelines

### Development Environment
- ✅ Safe to run any experiment
- ✅ Can test aggressive changes
- ✅ Good for learning

### Staging Environment
- ⚠️  Should mirror production
- ⚠️  Use for validation
- ⚠️  Enable auto-rollback

### Production Environment
- ⛔ Only run after dev/staging success
- ⛔ **Always** use auto-rollback
- ⛔ Monitor closely
- ⛔ Have incident response ready
- ⛔ Schedule during low-traffic if possible

### High-Risk Changes
These require extra caution:
- Reducing replica counts
- Modifying critical paths (ingesters, distributors)
- Changing compaction settings
- Adjusting retention policies

### Low-Risk Changes
Generally safe to test:
- Query optimization
- Sampling rate adjustments
- Cache tuning
- Read path modifications

---

## Troubleshooting

### Experiment Fails to Start
```bash
# Check configuration syntax
cat docs/experiments/my-experiment.json | from json

# Verify environment exists
./scripts/nu/health-check.nu --env dev

# Check component exists
kubectl get deployment -n mop-dev mimir-distributor
```

### Metrics Not Collecting
```bash
# Verify Mimir connectivity
kubectl port-forward -n mop-dev svc/mimir-query-frontend 8080:8080

# Test query manually
curl "http://localhost:8080/prometheus/api/v1/query?query=up"
```

### Rollback Issues
```bash
# Manual rollback
kubectl rollout undo deployment/mimir-distributor -n mop-dev

# Check rollout status
kubectl rollout status deployment/mimir-distributor -n mop-dev
```

---

## Example Results Analysis

After running an experiment, you'll get results like:

```
📈 Experiment Results

Experiment: Adaptive Sampling Test
Description: Test adaptive sampling impact on cost and quality

Summary:
┌─────────────────────┬──────────┬────────────┬─────────┬────────┐
│ metric              │ baseline │ experiment │ change  │ status │
├─────────────────────┼──────────┼────────────┼─────────┼────────┤
│ ingestion_rate      │ 15234.50 │ 7823.20    │ -48.65% │ ✓      │
│ query_latency_p95   │ 0.32     │ 0.34       │ +6.25%  │ ⚠️      │
│ query_accuracy      │ 245.80   │ 238.90     │ -2.81%  │ ✓      │
│ storage_usage       │ 1234567  │ 753421     │ -38.96% │ ✓      │
└─────────────────────┴──────────┴────────────┴─────────┴────────┘

Overall Score: 0.88
Recommendation: ADOPT - Changes show clear improvement

💡 Analysis:
- ✅ Ingestion rate reduced by 48.65% (cost savings!)
- ⚠️  Query latency increased by 6.25% (within acceptable range)
- ✅ Query accuracy minimally impacted (-2.81%)
- ✅ Storage usage reduced by 38.96%

Decision: Safe to deploy to production
```

---

## Contributing

To add new experiments:

1. Create configuration file in this directory
2. Test thoroughly in development
3. Document expected outcomes
4. Add to this README
5. Share results with team

---

## Resources

- [OBI Framework Documentation](../README.md)
- [MOP Architecture Overview](../architecture.md)
- [Experiment Runner Script](../../scripts/nu/experiment-runner.nu)
- [Health Check Script](../../scripts/nu/health-check.nu)
- [Cost Analysis Script](../../scripts/nu/cost-analysis.nu)
