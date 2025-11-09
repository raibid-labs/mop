#!/bin/bash
# Add labels to all existing issues based on their title patterns

set -e

echo "Adding labels to all issues..."

# Get all issues
ALL_ISSUES=$(gh issue list --json number,title --limit 100 --state all)

# Function to add labels to issues matching a pattern
add_labels_by_pattern() {
  local pattern="$1"
  local labels="$2"

  echo "Processing issues matching: $pattern"
  echo "$ALL_ISSUES" | jq -r ".[] | select(.title | contains(\"$pattern\")) | .number" | while read -r issue_num; do
    if [ -n "$issue_num" ]; then
      echo "  Adding labels to issue #$issue_num"
      gh issue edit "$issue_num" --add-label "$labels" 2>&1 | grep -v "^$" || true
    fi
  done
}

# Add workstream labels to each set of issues
add_labels_by_pattern "REF-01" "workstream:ws-ref-01,status:new"
add_labels_by_pattern "REF-02" "workstream:ws-ref-02,status:new"
add_labels_by_pattern "REF-03" "workstream:ws-ref-03,status:new"
add_labels_by_pattern "REF-04" "workstream:ws-ref-04,status:new"
add_labels_by_pattern "REF-05" "workstream:ws-ref-05,status:new"
add_labels_by_pattern "REF-06" "workstream:ws-ref-06,status:new"
add_labels_by_pattern "REF-07" "workstream:ws-ref-07,status:new"

echo ""
echo "✅ Labels added to all issues"
echo ""
echo "Next: GitHub Actions workflows will automatically:"
echo "  1. Check issues for clarifying questions"
echo "  2. Add 'ready:work' or 'waiting:answers' labels"
echo "  3. Post spawn trigger comments for ready issues"
echo "  4. Launch development agents"
