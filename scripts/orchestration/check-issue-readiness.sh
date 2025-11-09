#!/bin/bash
# Check if issue is ready for development (all clarifying questions answered)

set -e

ISSUE_NUMBER="${ISSUE_NUMBER:-$1}"

if [ -z "$ISSUE_NUMBER" ]; then
  echo "Error: ISSUE_NUMBER not provided"
  exit 1
fi

# Fetch issue details
ISSUE_BODY=$(gh issue view "$ISSUE_NUMBER" --json body -q .body)

# Extract "Clarifying Questions" section
QUESTIONS_SECTION=$(echo "$ISSUE_BODY" | sed -n '/## Clarifying Questions/,/## /p' | head -n -1)

if [ -z "$QUESTIONS_SECTION" ]; then
  echo "No clarifying questions found - issue is ready"
  echo '{"ready": true, "unanswered_questions": []}' > /tmp/issue-readiness.json
  exit 0
fi

# Parse questions (supports **Q1:** and plain Q1: formats)
QUESTIONS=$(echo "$QUESTIONS_SECTION" | grep -E '^\*?\*?Q[0-9]+:' | sed 's/\*\*//g')

if [ -z "$QUESTIONS" ]; then
  echo "No questions found - issue is ready"
  echo '{"ready": true, "unanswered_questions": []}' > /tmp/issue-readiness.json
  exit 0
fi

# Fetch all comments
COMMENTS=$(gh issue view "$ISSUE_NUMBER" --json comments -q '.comments[].body' | tr '\n' ' ')

# Check each question for answers
UNANSWERED=()
while IFS= read -r question; do
  Q_NUM=$(echo "$question" | grep -oP 'Q\K[0-9]+')
  Q_TEXT=$(echo "$question" | sed 's/^Q[0-9]*: //')

  # Look for answer patterns in comments
  # Supports: A1:, Answer 1:, Q1: ... A:, numbered list (1., 2.)
  if echo "$COMMENTS" | grep -qE "(A${Q_NUM}:|Answer ${Q_NUM}:|Q${Q_NUM}:.*A:|^${Q_NUM}\.|\bDecision:)"; then
    echo "Question $Q_NUM answered"
  else
    echo "Question $Q_NUM unanswered"
    UNANSWERED+=("{\"id\": \"Q$Q_NUM\", \"question\": \"$Q_TEXT\"}")
  fi
done <<< "$QUESTIONS"

# Generate result JSON
if [ ${#UNANSWERED[@]} -eq 0 ]; then
  echo "All questions answered - issue is ready"
  echo '{"ready": true, "unanswered_questions": []}' > /tmp/issue-readiness.json
else
  echo "Questions remaining - issue not ready"
  UNANSWERED_JSON=$(IFS=,; echo "[${UNANSWERED[*]}]")
  echo "{\"ready\": false, \"unanswered_questions\": $UNANSWERED_JSON}" > /tmp/issue-readiness.json
fi
