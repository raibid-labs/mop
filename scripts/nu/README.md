# MOP Nushell Automation Scripts

Comprehensive automation scripts for managing the Metrics Observability Platform (MOP).

## Prerequisites

- [Nushell](https://www.nushell.sh/) >= 0.80.0
- `kubectl` - Kubernetes CLI
- `tanka` - Jsonnet-based Kubernetes configuration tool
- `helm` - Kubernetes package manager
- `jq` - JSON processor
- `jsonnet` and `jsonnet-bundler` - Jsonnet tools

## Scripts Overview

### 1. setup.nu - Environment Setup
Complete environment initialization and configuration.

**Features:**
- ✅ Prerequisites validation (kubectl, tanka, helm, jq, jsonnet, jb)
- ✅ Kubernetes cluster connectivity testing
- ✅ Tanka environment initialization
- ✅ Jsonnet dependency vendoring
- ✅ Namespace creation
- ✅ CRD installation

**Usage:**
```bash
# Setup development environment
./setup.nu --env dev

# Setup staging without vendoring
./setup.nu --env staging --skip-vendor

# Force reinstall CRDs
./setup.nu --env prod --force
```

**Options:**
- `--env <dev|staging|prod>` - Environment to setup (default: dev)
- `--skip-vendor` - Skip vendoring Jsonnet dependencies
- `--force` - Force reinstall CRDs

---

### 2. deploy.nu - Safe Deployment
Production-ready deployment with validation and rollback support.

**Features:**
- 🔍 Pre-deployment validation checks
- 📊 Interactive diff review
- ⚠️  User confirmation prompts
- ⏳ Progressive rollout monitoring
- 🧪 Post-deployment smoke tests
- 🔄 Automatic rollback on failure

**Usage:**
```bash
# Deploy to development (with confirmation)
./deploy.nu --env dev

# Deploy specific component
./deploy.nu --env staging --component mimir-ingester

# Auto-approve deployment (CI/CD)
./deploy.nu --env dev --auto-approve

# Skip smoke tests
./deploy.nu --env prod --no-smoke-test

# Custom timeout
./deploy.nu --env staging --timeout 900
```

**Options:**
- `--env <environment>` - Target environment (required)
- `--component <name>` - Deploy specific component only
- `--auto-approve` - Skip confirmation prompts
- `--no-smoke-test` - Skip post-deployment tests
- `--timeout <seconds>` - Deployment timeout (default: 600)

**Safety Features:**
- Pre-deployment validation
- Cluster connectivity check
- Configuration validation
- Resource availability check
- Pod health verification
- Service endpoint validation
- Component health monitoring

---

### 3. health-check.nu - System Health Monitoring
Comprehensive health verification for all MOP components.

**Features:**
- 🏥 Pod status and readiness checks
- 📡 Service endpoint validation
- 📊 Metrics endpoint verification
- 🔗 Inter-component connectivity tests
- 💻 Resource utilization monitoring
- 📈 Health report generation
- 👁️  Continuous watch mode

**Usage:**
```bash
# Check all components
./health-check.nu --env dev

# Check specific component
./health-check.nu --env prod --component mimir-ingester

# Export report as JSON
./health-check.nu --env staging --format json --export health-report.json

# Continuous monitoring (watch mode)
./health-check.nu --env dev --watch

# Generate markdown report
./health-check.nu --env prod --format markdown --export report.md
```

**Options:**
- `--env <environment>` - Target environment (required)
- `--component <name>` - Check specific component only
- `--format <table|json|markdown>` - Output format (default: table)
- `--export <path>` - Export report to file
- `--watch` - Continuous monitoring mode

**Health Checks:**
- Pod phase and container status
- Container restart counts
- Service endpoint availability
- Metrics endpoint accessibility (`:8080/metrics`)
- Inter-component connectivity (distributor→ingester, query-frontend→querier)
- Resource usage (CPU, memory)

---

### 4. cost-analysis.nu - Cost Analysis & Optimization
Analyze costs and generate optimization recommendations.

**Features:**
- 💰 Storage cost estimation
- ⚡ Compute cost calculation
- 📈 Ingestion cost analysis
- 📊 Cost breakdown by service
- 🎯 Optimization recommendations
- 📉 Baseline comparison
- 💡 Potential savings estimates

**Usage:**
```bash
# Analyze current costs
./cost-analysis.nu --env prod

# Custom analysis period
./cost-analysis.nu --env prod --period 30d

# Compare to baseline
./cost-analysis.nu --env prod --baseline baseline-2024-01.json

# Export as CSV
./cost-analysis.nu --env staging --format csv --export costs.csv

# Custom Mimir endpoint
./cost-analysis.nu --env dev --mimir-url http://mimir.example.com:8080
```

**Options:**
- `--env <environment>` - Target environment (required)
- `--period <duration>` - Analysis period: 1h, 1d, 7d, 30d (default: 7d)
- `--format <table|json|csv>` - Output format (default: table)
- `--export <path>` - Export report to file
- `--baseline <path>` - Compare to baseline file
- `--mimir-url <url>` - Mimir query endpoint (default: http://localhost:8080)

**Cost Metrics:**
- Active time series count
- Sample ingestion rate
- Query request rate
- Storage utilization
- Ingester instance count
- Storage block count

**Recommendations Include:**
- Data retention policy optimization
- Ingester scaling recommendations
- Service-level trace sampling adjustments
- Adaptive sampling enablement
- Tiered storage strategy suggestions

---

### 5. backup.nu - Configuration Backup
Automated backup of configurations and dashboards.

**Features:**
- 📊 Grafana dashboard export
- 🔌 Grafana datasource backup
- ⚙️  Tanka configuration backup
- ☸️  Kubernetes resource export
- 📦 Compressed archive creation
- ☁️  Cloud storage upload (S3/GCS)
- 🧹 Automatic retention cleanup
- ✅ Backup integrity verification

**Usage:**
```bash
# Basic backup
./backup.nu --env prod

# Custom output directory
./backup.nu --env staging --output /backups

# Upload to S3
./backup.nu --env prod --upload s3://my-bucket/mop-backups

# Upload to GCS
./backup.nu --env prod --upload gs://my-bucket/mop-backups

# Custom retention period
./backup.nu --env dev --retention 60

# With Grafana credentials
./backup.nu --env prod --grafana-url http://grafana.local --grafana-token <token>
```

**Options:**
- `--env <environment>` - Target environment (required)
- `--output <path>` - Output directory (default: backups)
- `--upload <url>` - Cloud storage URL (s3:// or gs://)
- `--retention <days>` - Retention period (default: 30)
- `--grafana-url <url>` - Grafana URL (default: http://localhost:3000)
- `--grafana-token <token>` - Grafana API token (or use GRAFANA_TOKEN env var)

**Backup Contents:**
- Grafana dashboards (JSON)
- Grafana datasources (JSON, credentials sanitized)
- Tanka environments and libraries
- Rendered Kubernetes manifests
- ConfigMaps, Secrets, Services
- Deployments, StatefulSets
- PVCs, Ingresses

**Archive Format:**
```
mop-prod-20240106-143022.tar.gz
├── grafana/
│   ├── dashboards/
│   │   ├── mimir-overview.json
│   │   └── trace-analysis.json
│   └── datasources/
│       ├── mimir.json
│       └── tempo.json
├── tanka/
│   ├── environments/
│   ├── lib/
│   ├── jsonnetfile.json
│   └── rendered/
│       └── prod.yaml
└── kubernetes/
    ├── configmaps.yaml
    ├── deployments.yaml
    └── services.yaml
```

---

### 6. experiment-runner.nu - OBI Experiment Automation
Automated experiment execution and analysis using the Observability-by-Inference framework.

**Features:**
- 🧪 Automated experiment execution
- 📊 Baseline metric collection
- 🚀 Experimental change deployment
- 👁️  Continuous metric monitoring
- 🔍 Statistical analysis
- 📈 Improvement calculation
- 🎯 Automated recommendations
- 🔄 Automatic rollback on degradation
- 📄 Comprehensive report generation

**Usage:**
```bash
# Run experiment from config
./experiment-runner.nu --config experiments/adaptive-sampling.json --env dev

# Custom duration
./experiment-runner.nu --config exp.json --env staging --duration 7200

# Auto-rollback on degradation
./experiment-runner.nu --config exp.json --env prod --auto-rollback

# Export results
./experiment-runner.nu --config exp.json --env dev --export results.json

# Extended baseline collection
./experiment-runner.nu --config exp.json --env staging --baseline-duration 600
```

**Options:**
- `--config <path>` - Experiment configuration file (required)
- `--env <environment>` - Target environment (default: dev)
- `--duration <seconds>` - Experiment duration (default: 3600)
- `--baseline-duration <seconds>` - Baseline collection period (default: 300)
- `--auto-rollback` - Automatically rollback on metric degradation
- `--export <path>` - Export results to file

**Experiment Configuration Format:**
```json
{
  "name": "Adaptive Sampling Test",
  "description": "Test adaptive sampling impact on cost and quality",
  "changes": [
    {
      "type": "deployment",
      "component": "mimir-distributor",
      "container": "distributor",
      "parameter": "SAMPLE_RATE",
      "value": "0.5"
    }
  ],
  "success_metrics": [
    {
      "name": "ingestion_rate",
      "query": "sum(rate(mimir_distributor_samples_in_total[5m]))",
      "direction": "lower",
      "threshold": 10000
    },
    {
      "name": "query_latency_p95",
      "query": "histogram_quantile(0.95, rate(mimir_request_duration_seconds_bucket[5m]))",
      "direction": "lower",
      "threshold": 0.5
    }
  ]
}
```

**Change Types:**
- `deployment` - Modify deployment environment variables
- `configmap` - Update ConfigMap values

**Metric Directions:**
- `lower` - Lower is better (latency, cost, errors)
- `higher` - Higher is better (throughput, availability)

**Analysis Recommendations:**
- `adopt` - Score ≥ 0.8, clear improvement
- `investigate` - Score ≥ 0.5, inconclusive results
- `rollback` - Score < 0.5, degradation detected

---

## Common Workflows

### Initial Setup
```bash
# 1. Setup environment
./setup.nu --env dev

# 2. Deploy components
./deploy.nu --env dev

# 3. Verify health
./health-check.nu --env dev
```

### Production Deployment
```bash
# 1. Deploy to staging first
./deploy.nu --env staging

# 2. Run health checks
./health-check.nu --env staging

# 3. Create backup before prod deployment
./backup.nu --env prod --upload s3://backups/mop

# 4. Deploy to production
./deploy.nu --env prod

# 5. Monitor health continuously
./health-check.nu --env prod --watch
```

### Cost Optimization
```bash
# 1. Analyze current costs
./cost-analysis.nu --env prod --export baseline.json

# 2. Run experiment with optimizations
./experiment-runner.nu --config optimize-sampling.json --env dev

# 3. Compare results
./cost-analysis.nu --env prod --baseline baseline.json

# 4. Deploy if successful
./deploy.nu --env prod --component mimir-distributor
```

### Disaster Recovery
```bash
# 1. Create comprehensive backup
./backup.nu --env prod --upload s3://dr-backups/mop

# 2. If recovery needed, restore from backup
# (Manual restoration from backup archive)

# 3. Verify health after restoration
./health-check.nu --env prod --format json --export health-report.json
```

---

## Environment Variables

### Grafana Authentication
```bash
export GRAFANA_TOKEN="your-api-token"
./backup.nu --env prod
```

### Custom Kubernetes Context
```bash
export KUBECONFIG=/path/to/kubeconfig
./deploy.nu --env prod
```

### AWS Credentials (for S3 upload)
```bash
export AWS_ACCESS_KEY_ID="your-key"
export AWS_SECRET_ACCESS_KEY="your-secret"
./backup.nu --env prod --upload s3://bucket/path
```

### GCP Credentials (for GCS upload)
```bash
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/credentials.json"
./backup.nu --env prod --upload gs://bucket/path
```

---

## Nushell Features Used

These scripts leverage Nushell's powerful features:

- **Structured Data**: All data is typed and structured
- **Pipelines**: Clean data transformation with `|`
- **Error Handling**: Robust `try`/`catch` blocks
- **Type Safety**: Strong typing for function parameters
- **Tables**: Beautiful table formatting with `| table -e`
- **JSON Support**: Native JSON parsing with `from json` / `to json`
- **YAML Support**: Native YAML parsing with `from yaml` / `to yaml`
- **Date/Time**: Built-in date manipulation
- **Math Operations**: Native math functions
- **HTTP Requests**: Built-in HTTP client
- **ANSI Colors**: Rich terminal output with color support

---

## Troubleshooting

### Script Permissions
```bash
chmod +x scripts/nu/*.nu
```

### Missing Tools
```bash
# Install Nushell
brew install nushell

# Install Kubernetes tools
brew install kubectl tanka helm

# Install Jsonnet tools
brew install jsonnet jsonnet-bundler

# Install utilities
brew install jq
```

### Port Forward Issues
```bash
# Check existing port forwards
ps aux | grep port-forward

# Kill existing port forwards
pkill -f "port-forward.*mimir"

# Manually setup port forward
kubectl port-forward -n mop-prod svc/mimir-query-frontend 8080:8080
```

### Grafana Connection
```bash
# Test Grafana connectivity
curl -H "Authorization: Bearer $GRAFANA_TOKEN" http://localhost:3000/api/health

# Generate API token in Grafana
# Settings → API Keys → Add API Key
```

---

## Best Practices

1. **Always run health checks after deployment**
   ```bash
   ./deploy.nu --env prod && ./health-check.nu --env prod
   ```

2. **Create backups before major changes**
   ```bash
   ./backup.nu --env prod --upload s3://backups/mop
   ```

3. **Test in dev/staging first**
   ```bash
   ./deploy.nu --env dev
   ./health-check.nu --env dev
   ./deploy.nu --env staging
   ./deploy.nu --env prod
   ```

4. **Use experiments for risky changes**
   ```bash
   ./experiment-runner.nu --config change.json --env dev --auto-rollback
   ```

5. **Monitor costs regularly**
   ```bash
   # Weekly cost analysis
   ./cost-analysis.nu --env prod --export "costs-$(date +%Y%m%d).json"
   ```

---

## Contributing

When adding new scripts:

1. Follow the existing structure and naming conventions
2. Include comprehensive error handling
3. Add detailed comments and documentation
4. Use Nushell idioms (structured data, pipelines)
5. Provide helpful output with ANSI colors
6. Include usage examples in comments

---

## License

Part of the MOP (Metrics Observability Platform) project.
