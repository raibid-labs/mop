#!/bin/bash
# Find and assign next ready issue based on priority

set -e

# Fetch all issues with ready:work label
READY_ISSUES=$(gh issue list --label "ready:work" --json number,title,labels,createdAt --limit 100)

if [ "$READY_ISSUES" == "[]" ]; then
  echo "No ready issues found"
  exit 0
fi

# Priority order: critical > high > medium > low
PRIORITIES=("priority:critical" "priority:high" "priority:medium" "priority:low")

NEXT_ISSUE=""
for priority in "${PRIORITIES[@]}"; do
  # Find oldest issue with this priority
  NEXT_ISSUE=$(echo "$READY_ISSUES" | jq -r ".[] | select(.labels[].name == \"$priority\") | .number" | head -1)

  if [ -n "$NEXT_ISSUE" ]; then
    echo "Found issue #$NEXT_ISSUE with $priority"
    break
  fi
done

if [ -z "$NEXT_ISSUE" ]; then
  # No priority label found, take oldest ready issue
  NEXT_ISSUE=$(echo "$READY_ISSUES" | jq -r '.[0].number')
  echo "Found issue #$NEXT_ISSUE (no priority label)"
fi

if [ -n "$NEXT_ISSUE" ]; then
  echo "Spawning agent for issue #$NEXT_ISSUE"
  bash "$(dirname "$0")/spawn-agent-comment.sh" "$NEXT_ISSUE"
else
  echo "No issues to assign"
fi
