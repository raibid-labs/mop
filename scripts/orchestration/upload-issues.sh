#!/bin/bash
# Upload all issue markdown files to GitHub

set -e

ISSUES_DIR="docs/issues"

if [ ! -d "$ISSUES_DIR" ]; then
  echo "Error: $ISSUES_DIR not found"
  exit 1
fi

echo "Uploading issues to GitHub..."

# Counter
CREATED=0
FAILED=0

# Iterate through all workstream directories
for ws_dir in "$ISSUES_DIR"/ws-ref-*/; do
  ws_name=$(basename "$ws_dir")
  echo "Processing $ws_name..."

  # Iterate through all issue files
  for issue_file in "$ws_dir"/*.md; do
    if [ ! -f "$issue_file" ]; then
      continue
    fi

    issue_name=$(basename "$issue_file" .md)
    echo "  Creating issue $issue_name..."

    # Extract title from first line (# Title)
    title=$(head -1 "$issue_file" | sed 's/^# //')

    # Extract labels from file (look for ## Labels section or infer from ws)
    labels="${ws_name/ws-/workstream:ws-},status:new"

    # Add priority label if found in file
    if grep -q "priority:critical" "$issue_file"; then
      labels="$labels,priority:critical"
    elif grep -q "priority:high" "$issue_file"; then
      labels="$labels,priority:high"
    elif grep -q "priority:medium" "$issue_file"; then
      labels="$labels,priority:medium"
    fi

    # Create issue
    if gh issue create \
      --title "$title" \
      --body-file "$issue_file" \
      --label "$labels" 2>&1 | tee /tmp/gh-issue-create.log; then
      CREATED=$((CREATED + 1))
      # Extract issue number from output
      issue_num=$(grep -oP '#\K\d+' /tmp/gh-issue-create.log | head -1)
      echo "    ✓ Created issue #$issue_num"
    else
      FAILED=$((FAILED + 1))
      echo "    ✗ Failed to create issue"
    fi

    # Rate limit: sleep between issues
    sleep 1
  done
done

echo ""
echo "Summary:"
echo "  Created: $CREATED issues"
echo "  Failed: $FAILED issues"

if [ $CREATED -gt 0 ]; then
  echo ""
  echo "Next steps:"
  echo "  1. Review issues at: https://github.com/$(gh repo view --json nameWithOwner -q .nameWithOwner)/issues"
  echo "  2. Answer any clarifying questions"
  echo "  3. Orchestrator will automatically spawn agents for ready issues"
fi
