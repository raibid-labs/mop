#!/bin/bash
# Post spawn trigger comment for orchestrator to detect

set -e

ISSUE_NUMBER="${1:-$ISSUE_NUMBER}"

if [ -z "$ISSUE_NUMBER" ]; then
  echo "Error: ISSUE_NUMBER not provided"
  exit 1
fi

# Fetch issue details
ISSUE_JSON=$(gh issue view "$ISSUE_NUMBER" --json title,labels,body)
ISSUE_TITLE=$(echo "$ISSUE_JSON" | jq -r '.title')
ISSUE_LABELS=$(echo "$ISSUE_JSON" | jq -r '.labels[].name' | tr '\n' ',' | sed 's/,$//')

# Extract issue ID from title (e.g., "REF-01-001: Title" -> "REF-01-001")
ISSUE_ID=$(echo "$ISSUE_TITLE" | grep -oP '^[A-Z]+-\d+-\d+' || echo "UNKNOWN")

# Determine agent type based on workstream
AGENT_TYPE="coder"  # Default
if echo "$ISSUE_LABELS" | grep -q "ws-ref-01\|ws-ref-02\|ws-ref-03\|ws-ref-04\|ws-ref-05"; then
  AGENT_TYPE="golang-pro"
elif echo "$ISSUE_LABELS" | grep -q "ws-ref-06"; then
  AGENT_TYPE="test-automator"
elif echo "$ISSUE_LABELS" | grep -q "ws-ref-07"; then
  AGENT_TYPE="docs-architect"
fi

# Current timestamp
TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Post spawn trigger comment
gh issue comment "$ISSUE_NUMBER" --body "$(cat <<EOF
🤖 **ORCHESTRATOR-SPAWN-AGENT**

**Issue**: #${ISSUE_NUMBER}
**Issue ID**: ${ISSUE_ID}
**Type**: ${AGENT_TYPE}
**Status**: ready
**Timestamp**: ${TIMESTAMP}

**Agent Instructions:**
1. Review this issue and all comments thoroughly
2. Follow TDD workflow (tests first, then implementation)
3. Create feature branch: \`${ISSUE_ID,,}-implementation\`
4. Commit frequently with clear messages
5. Submit PR when complete, referencing this issue

---
<!-- ORCHESTRATOR-STATE
{
  "issue": ${ISSUE_NUMBER},
  "issue_id": "${ISSUE_ID}",
  "agent_type": "${AGENT_TYPE}",
  "status": "ready",
  "spawned_at": "${TIMESTAMP}"
}
-->
EOF
)"

echo "Spawn trigger comment posted for issue #${ISSUE_NUMBER}"
