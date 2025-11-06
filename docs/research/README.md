# MOP Research: Tanka + Helm + Grafana Stack

## Overview

This directory contains comprehensive research findings on implementing the MOP (Monitoring Operations Platform) using Grafana Tanka, Helm charts, and Jsonnet for deploying and managing the Grafana observability stack.

## Research Documents

### 1. [Tanka + Helm Integration Patterns](./tanka-helm-patterns.md)

**Purpose**: Complete guide to integrating Tanka with Helm charts

**Key Topics**:
- Directory structure best practices
- Helm chart integration patterns (direct, deep merge, wrapper library)
- Chart management with `tk tool charts`
- Multi-environment configuration strategies
- Common issues and solutions
- Best practices (DO/DON'T lists)

**Use this when**: Learning how to work with Tanka and Helm together, setting up project structure

### 2. [Grafana Stack Concrete Examples](./grafana-stack-examples.md)

**Purpose**: Production-ready configuration examples for the complete Grafana observability stack

**Key Topics**:
- Complete project structure
- Shared configuration library
- Loki microservices configuration
- Mimir distributed deployment
- Tempo with distributed tracing
- Grafana with integrated datasources
- Deployment and testing procedures
- Troubleshooting guides

**Use this when**: Implementing actual Grafana stack components, need working code examples

### 3. [Architecture Decision Guide](./architecture-decision-guide.md)

**Purpose**: Strategic decisions and architectural patterns for the MOP project

**Key Topics**:
- Architecture Decision Records (ADRs)
- Directory structure decision matrix
- Component integration patterns
- Configuration management strategies
- Testing and deployment strategies
- Operational patterns
- Migration roadmap
- Decision checklist

**Use this when**: Making architectural decisions, planning implementation strategy

## Quick Start

### For First-Time Implementers

1. **Read**: [Architecture Decision Guide](./architecture-decision-guide.md) - Section 12 (Decision Checklist)
2. **Read**: [Tanka + Helm Patterns](./tanka-helm-patterns.md) - Sections 1-2 (Structure & Integration)
3. **Read**: [Grafana Stack Examples](./grafana-stack-examples.md) - Section 1 (Project Structure)
4. **Implement**: Start with dev environment, single component (Grafana)

### For Experienced Tanka Users

1. **Review**: [Architecture Decision Guide](./architecture-decision-guide.md) - ADRs for context
2. **Copy**: [Grafana Stack Examples](./grafana-stack-examples.md) - Production configurations
3. **Adapt**: Modify for your specific requirements
4. **Deploy**: Use deployment strategies from Architecture Guide

### For Troubleshooting

1. **Check**: [Tanka + Helm Patterns](./tanka-helm-patterns.md) - Section 7 (Common Issues)
2. **Review**: [Grafana Stack Examples](./grafana-stack-examples.md) - Section 11 (Troubleshooting)
3. **Verify**: Configuration against best practices in patterns document

## Key Findings Summary

### ✅ Recommended Approach

1. **Tool Stack**:
   - Tanka 0.25+ for orchestration
   - Jsonnet for configuration
   - Helm charts vendored locally
   - k8s-libsonnet for Kubernetes resources

2. **Project Structure**:
   ```
   mop/
   ├── environments/     # Environment-specific configs
   ├── lib/              # Reusable libraries
   ├── charts/           # Vendored Helm charts
   └── vendor/           # External Jsonnet dependencies
   ```

3. **Integration Pattern**:
   - Wrap Helm charts in Jsonnet libraries
   - Use deep merging for customization
   - Centralize configuration in `lib/config.libsonnet`
   - Environment-specific overrides in `environments/*/main.jsonnet`

4. **Component Configuration**:
   - Loki: Microservices mode with S3 backend
   - Mimir: Distributed mode with S3 blocks storage
   - Tempo: Distributed tracing with S3 traces
   - Grafana: Integrated with all datasources

5. **Deployment Strategy**:
   - Dev: All-at-once deployment
   - Staging: Component-by-component
   - Production: Canary or blue-green

### ❌ What NOT to Do

1. Don't use remote Helm charts (always vendor locally)
2. Don't skip `.new(std.thisFile)` in helm initialization
3. Don't fork Helm charts (use Jsonnet deep merging instead)
4. Don't hardcode values (use configuration libraries)
5. Don't commit secrets (use external secret management)
6. Don't skip testing (`tk diff` before `tk apply`)

### 🎯 Critical Success Factors

1. **Team Knowledge**: Ensure team understands Jsonnet basics
2. **Environment Parity**: Keep dev/staging/prod configurations similar
3. **Testing**: Always test in lower environments first
4. **Documentation**: Document customizations and overrides
5. **Version Control**: Lock all dependencies (charts, jsonnet libs)
6. **Monitoring**: Implement drift detection and alerting

## Research Methodology

This research was conducted by:

1. **Official Documentation Review**:
   - Grafana Tanka documentation (tanka.dev)
   - Grafana Labs blog posts and tutorials
   - Helm integration guides

2. **Code Analysis**:
   - Grafana jsonnet-libs repository patterns
   - Mimir operations Jsonnet files
   - Loki production ksonnet configurations
   - TNS (observability demo) reference implementation

3. **Best Practices Extraction**:
   - Community patterns and examples
   - Production deployments learnings
   - Integration challenges and solutions

4. **Pattern Synthesis**:
   - Distilled common patterns across sources
   - Created reusable abstractions
   - Documented decision frameworks

## Notable Limitations

### gudo11y/mop-core Repository

The specific `gudo11y/mop-core` repository mentioned in the original request was not found in:
- GitHub search results
- Web search engines
- Grafana Labs official repositories
- Community examples

This could indicate:
- Private/internal repository
- Typo in username or repository name
- Recently deleted or renamed repository
- Very new project with limited online presence

**Recommendation**: If this repository exists and contains relevant patterns, please provide:
- Correct repository URL
- Access credentials (if private)
- Specific files or patterns to review

The patterns and examples in this research are based on **official Grafana Labs implementations** and **community best practices**, which should be applicable regardless of the specific reference repository.

## Tools and Dependencies

### Required

```bash
# Tanka
brew install tanka
# or
go install github.com/grafana/tanka/cmd/tk@latest

# Jsonnet Bundler
brew install jsonnet-bundler
# or
go install github.com/jsonnet-bundler/jsonnet-bundler/cmd/jb@latest

# Helm (for tk tool charts)
brew install helm
# or
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash

# Kubectl (for deployment)
brew install kubectl
```

### Optional but Recommended

```bash
# Jsonnet formatter
brew install jsonnet
# or
go install github.com/google/go-jsonnet/cmd/jsonnetfmt@latest

# Kubernetes schema validation
brew install kubeval
# or
go install github.com/instrumenta/kubeval@latest

# Alternative: kubeconform
brew install kubeconform
# or
go install github.com/yannh/kubeconform/cmd/kubeconform@latest
```

## Next Steps

### Immediate Actions (Week 1)

1. **Initialize Project**:
   ```bash
   cd /Users/beengud/raibid-labs/mop
   tk init --k8s=1.28
   ```

2. **Setup Dependencies**:
   ```bash
   jb install github.com/grafana/jsonnet-libs/tanka-util
   jb install github.com/grafana/jsonnet-libs/ksonnet-util
   ```

3. **Create Structure**:
   ```bash
   mkdir -p lib/{config,components,utils}
   mkdir -p environments/{dev,staging,production}
   ```

4. **Vendor Charts**:
   ```bash
   tk tool charts init
   tk tool charts add-repo grafana https://grafana.github.io/helm-charts
   # Add specific charts as needed
   ```

### Short Term (Week 2-4)

1. Implement dev environment with Grafana
2. Add Loki with basic configuration
3. Add Prometheus or Mimir
4. Configure datasources and test

### Medium Term (Week 5-8)

1. Create staging environment
2. Add Tempo for tracing
3. Implement production configuration
4. Setup CI/CD pipeline

### Long Term (Month 3+)

1. Optimize resource usage
2. Implement advanced patterns
3. Add monitoring and alerting
4. Document learnings and patterns

## Support and Resources

### Official Documentation

- **Tanka**: https://tanka.dev
- **Grafana Jsonnet Libs**: https://github.com/grafana/jsonnet-libs
- **k8s-libsonnet**: https://github.com/jsonnet-libs/k8s-libsonnet
- **Helm**: https://helm.sh/docs

### Example Repositories

- **TNS Demo**: https://github.com/grafana/tns (complete observability stack)
- **Mimir Operations**: https://github.com/grafana/mimir/tree/main/operations/mimir
- **Loki Production**: https://github.com/grafana/loki/tree/main/production/ksonnet

### Community

- **Grafana Community**: https://community.grafana.com
- **Tanka Discussions**: https://github.com/grafana/tanka/discussions
- **Cloud Native Slack**: #tanka channel

## Contributing to This Research

If you find additional patterns, examples, or corrections:

1. Add findings to appropriate document
2. Update this README with new sections
3. Add sources and references
4. Share learnings with the team

## Changelog

- **2024-06-11**: Initial research compilation
  - Tanka + Helm integration patterns
  - Grafana stack concrete examples
  - Architecture decision guide
  - Summary and quick start guide

---

**Researcher**: Claude (Research Agent)
**Date**: June 11, 2024
**Status**: ✅ Complete

**Note**: Research conducted without access to `gudo11y/mop-core` repository. All patterns based on official Grafana Labs implementations and community best practices.
