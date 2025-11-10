# MOP Reference Implementations - Launch Instructions

**For**: New Claude session on any machine
**Status**: All setup complete, ready to launch agents
**Date**: 2025-11-09

---

## 🎯 Current State

Everything is ready to launch parallel development agents:

✅ **Setup Complete**:
- 53 GitHub issues created and labeled
- 14 GitHub labels configured
- 3 GitHub Actions workflows active
- 5 orchestration scripts deployed
- All documentation complete

⏭️ **Next Step**: Launch development agents (instructions below)

---

## 📍 Quick Verification

Before launching, verify the setup:

```bash
cd ~/raibid-labs/mop

# Check you're on latest main branch
git pull origin main

# Verify GitHub issues exist (should show 53 issues)
gh issue list --limit 5

# Verify labels exist (should show 14 labels)
gh label list | grep -E "(workstream|status|ready|priority)" | wc -l
```

**Expected**:
- Git pull shows "Already up to date" or pulls latest
- Issues list shows REF-XX-XXX titled issues
- Label count shows 14

---

## 🚀 Launch Option 1: Manual (Recommended for Control)

Launch agents manually using Claude Code's Task tool. This gives you direct control over execution.

### Wave 1: Foundation (Launch All 3 in Parallel)

**IMPORTANT**: Send as a **single message** with all 3 Task calls to run in parallel.

```
Launch Wave 1 agents in parallel:

Task("HTTP API Developer", "Complete WS-REF-01 workstream (7 issues total). Read all issue files in docs/issues/ws-ref-01/ for detailed specifications. Implement HTTP REST API example application using Go 1.21+, Gin framework, Docker multi-stage builds, and OBI eBPF instrumentation for automatic tracing. Follow TDD workflow. Create feature branch ref-01-implementation. Submit PR when complete. Reference: docs/planning/reference-implementations-plan.md for context.", "golang-pro")

Task("gRPC Developer", "Complete WS-REF-02 workstream (8 issues total). Read all issue files in docs/issues/ws-ref-02/ for detailed specifications. Implement gRPC service example with Protocol Buffers, Go 1.21+, Docker multi-stage builds, and OBI eBPF instrumentation for automatic tracing. Follow TDD workflow. Create feature branch ref-02-implementation. Submit PR when complete. Reference: docs/planning/reference-implementations-plan.md for context.", "golang-pro")

Task("Redis Developer", "Complete WS-REF-04 workstream (8 issues total). Read all issue files in docs/issues/ws-ref-04/ for detailed specifications. Implement Redis caching patterns example using Go 1.21+, go-redis client, Docker multi-stage builds, and OBI eBPF instrumentation for automatic cache operation tracing. Follow TDD workflow. Create feature branch ref-04-implementation. Submit PR when complete. Reference: docs/planning/reference-implementations-plan.md for context.", "golang-pro")
```

**Duration**: ~6-8 hours for Wave 1 to complete

---

### Wave 2: Data Services (After Wave 1 Completes)

Launch after Wave 1 PRs are merged:

```
Launch Wave 2 agents in parallel:

Task("SQL Developer", "Complete WS-REF-03 workstream (8 issues total). Read all issue files in docs/issues/ws-ref-03/ for detailed specifications. Implement SQL application example using Go 1.21+, PostgreSQL, sqlx or GORM, Docker Compose with database, and OBI eBPF instrumentation for automatic query tracing. Follow TDD workflow. Create feature branch ref-03-implementation. Submit PR when complete. Reference: docs/planning/reference-implementations-plan.md and examples/01-http-api for Go patterns.", "golang-pro")

Task("Kafka Developer", "Complete WS-REF-05 workstream (8 issues total). Read all issue files in docs/issues/ws-ref-05/ for detailed specifications. Implement Kafka streaming example using Go 1.21+, kafka-go or sarama client, Docker Compose with Kafka/Zookeeper, and OBI eBPF instrumentation for automatic message tracing. Follow TDD workflow. Create feature branch ref-05-implementation. Submit PR when complete. Reference: docs/planning/reference-implementations-plan.md and examples/01-http-api for Go patterns.", "golang-pro")
```

**Duration**: ~8-10 hours for Wave 2 to complete

---

### Wave 3: Validation (After Waves 1 & 2 Complete)

Launch after all application examples are complete:

```
Launch Wave 3 agents sequentially:

Task("Load Test Engineer", "Complete WS-REF-06 workstream (7 issues total). Read all issue files in docs/issues/ws-ref-06/ for detailed specifications. Implement load generators for all 5 protocol examples (HTTP, gRPC, SQL, Redis, Kafka) using appropriate tools (hey, ghz, pgbench, redis-benchmark, kafka-producer-perf-test). Create Docker Compose orchestration for realistic traffic generation. Include documentation. Create feature branch ref-06-implementation. Submit PR when complete.", "test-automator")

Task("Documentation Engineer", "Complete WS-REF-07 workstream (7 issues total). Read all issue files in docs/issues/ws-ref-07/ for detailed specifications. Create protocol-specific Grafana dashboards (6 total: HTTP, gRPC, SQL, Redis, Kafka, Multi-Protocol). Write comprehensive documentation for each example. Create getting started guide. Document OBI instrumentation patterns. Create feature branch ref-07-implementation. Submit PR when complete.", "docs-architect")
```

**Duration**: ~10-12 hours for Wave 3 to complete

---

## 🤖 Launch Option 2: Automatic (GitHub Actions)

GitHub Actions workflows will automatically orchestrate agents if you prefer hands-off approach.

### How Automatic Works

1. **Workflow triggers**: Already active on issue/comment/PR events
2. **Issue processing**: Checks for clarifying questions
3. **Label management**: Adds `ready:work` or `waiting:answers`
4. **Agent spawning**: Posts spawn trigger comments
5. **PR management**: Auto-assigns next issue on PR merge

### Start Automatic Orchestration

GitHub Actions will start processing automatically once issues have labels. To manually trigger:

```bash
# Trigger issue event workflow for all issues
gh api repos/raibid-labs/mop/dispatches -f event_type=orchestrator-init
```

**Note**: Automatic mode requires an orchestrator agent that polls for spawn triggers. The workflows are in place, but you'll need to run the orchestrator polling loop.

---

## 📋 Agent Instructions Summary

Each agent receives:
- **Issue files**: Full specifications in `docs/issues/ws-ref-XX/`
- **Planning doc**: Context in `docs/planning/reference-implementations-plan.md`
- **Workflow**: TDD workflow (tests first, then implementation)
- **Branching**: Feature branch `ref-XX-implementation`
- **Deliverable**: PR with working code, tests, docs, Dockerfile

---

## 🔍 Monitoring Progress

### Check Agent Status

```bash
# View active issues
gh issue list --label "ready:work"

# View completed PRs
gh pr list --state merged

# View workflow runs
gh run list --limit 10
```

### Expected Artifacts

After each workstream completes:

**WS-REF-01 (HTTP)**:
- `examples/01-http-api/` directory
- Go HTTP REST API with Gin
- Dockerfile and README
- Tests with >80% coverage

**WS-REF-02 (gRPC)**:
- `examples/02-grpc-service/` directory
- Go gRPC service with protobuf
- Dockerfile and README
- Tests with >80% coverage

**WS-REF-03 (SQL)**:
- `examples/03-sql-app/` directory
- Go application with PostgreSQL
- Docker Compose with database
- Tests with >80% coverage

**WS-REF-04 (Redis)**:
- `examples/04-redis-cache/` directory
- Go application with Redis
- Docker Compose with Redis
- Tests with >80% coverage

**WS-REF-05 (Kafka)**:
- `examples/05-kafka-streaming/` directory
- Go producer/consumer with Kafka
- Docker Compose with Kafka
- Tests with >80% coverage

**WS-REF-06 (Load Generators)**:
- `examples/load-generators/` directory
- Load test scripts for all protocols
- Docker Compose orchestration
- Documentation

**WS-REF-07 (Dashboards)**:
- `dashboards/` directory (6 Grafana dashboards)
- `docs/guides/` directory (comprehensive docs)
- Getting started guide
- OBI instrumentation patterns guide

---

## ⏱️ Timeline

- **Wave 1**: 6-8 hours (HTTP, gRPC, Redis in parallel)
- **Wave 2**: 8-10 hours (SQL, Kafka in parallel)
- **Wave 3**: 10-12 hours (Load tests, Dashboards sequential)
- **Total**: 24-30 hours elapsed (vs 50-60 if sequential)

---

## 🐛 Troubleshooting

### If agents can't find issue files:

```bash
# Verify issue files exist
ls -la docs/issues/ws-ref-01/
# Should show: REF-01-001.md through REF-01-007.md
```

### If agents ask about clarifying questions:

All issues have clarifying questions with sensible defaults. Agents should:
1. Use the default recommendations in issues
2. Proceed without blocking on answers
3. Document assumptions in PR descriptions

### If you need to answer clarifying questions:

Comment on specific GitHub issues with format:
```
A1: [Your answer to question 1]
A2: [Your answer to question 2]
```

GitHub Actions will detect answers and update labels automatically.

---

## 📚 Reference Documentation

- **This file**: Quick launch instructions
- **`docs/ORCHESTRATION-READY.md`**: Detailed orchestration guide
- **`docs/STATUS.md`**: Current project status
- **`docs/planning/reference-implementations-plan.md`**: Complete technical plan (549 lines)
- **`docs/issues/ws-ref-XX/`**: Individual issue specifications (53 files)

---

## 🎯 Quick Start Checklist

1. ✅ Verify you're in `~/raibid-labs/mop` directory
2. ✅ Verify git is up to date: `git pull origin main`
3. ✅ Verify 53 issues exist: `gh issue list --limit 5`
4. ✅ Verify 14 labels exist: `gh label list | grep workstream`
5. ✅ Choose launch option (Manual recommended)
6. ✅ Copy/paste Task commands into Claude Code
7. ✅ Monitor progress via GitHub issues/PRs

---

## 🚀 Ready to Launch

**Recommended**: Use Manual Launch Option 1 with Wave 1 agents.

Copy the Wave 1 Task commands above and paste into Claude Code as a single message to launch all 3 agents in parallel.

**Timeline**: Wave 1 will take ~6-8 hours. Return after Wave 1 completes to launch Wave 2.

---

**Last Updated**: 2025-11-09
**Setup Status**: ✅ Complete
**Ready to Execute**: Yes

---

## 💡 Tips for New Claude Session

When starting fresh:
1. Read this file first (you're doing it!)
2. Verify checklist items above
3. Don't try to understand full history - documentation is self-contained
4. Launch Wave 1 agents using commands above
5. Monitor progress and launch subsequent waves when ready

All context needed is in:
- This file (launch instructions)
- `docs/planning/reference-implementations-plan.md` (technical details)
- `docs/issues/ws-ref-XX/*.md` (detailed specifications)

**You don't need to read previous conversation history - everything is documented!**
