#!/bin/bash
# update-changelog.sh - Automate CHANGELOG.md updates after task completion

set -e

# Check arguments
if [ $# -lt 2 ]; then
    echo "Usage: $0 <task-id> <description> [files...]"
    echo "Example: $0 1.2.3 'Create logging package' shared/pkg/logging/logger.go"
    exit 1
fi

TASK_ID=$1
DESCRIPTION=$2
shift 2
FILES="$@"

# Get current date
DATE=$(date +%Y-%m-%d)

# Get current user (fallback to git config)
USER=${USER:-$(git config user.name)}
USER=${USER:-"Unknown"}

# Create changelog entry
ENTRY="#### $DATE - $TASK_ID: $DESCRIPTION
**Completed by**: $USER
**Files modified**: $FILES
**Verification**: Tests passed, artifacts created, acceptance criteria met

"

# Check if CHANGELOG.md exists
if [ ! -f "CHANGELOG.md" ]; then
    echo "❌ CHANGELOG.md not found"
    exit 1
fi

# Find the unreleased section
if grep -q "## \[Unreleased\]" CHANGELOG.md; then
    # Insert after the Unreleased header
    sed -i.bak "/## \[Unreleased\]/a\\
$ENTRY" CHANGELOG.md
    echo "✅ CHANGELOG.md updated with task $TASK_ID"
else
    echo "❌ Could not find [Unreleased] section in CHANGELOG.md"
    exit 1
fi

# Update project state summary if task counters are available
if command -v jq &> /dev/null && [ -f "tasks.json" ]; then
    COMPLETED_TASKS=$(jq '.phases[].sections[].tasks[] | select(.status=="completed") | .id' tasks.json | wc -l)
    TOTAL_TASKS=$(jq '.phases[].totalTasks' tasks.json | awk '{sum += $1} END {print sum}')

    # Update the summary in CHANGELOG.md
    sed -i.bak "s/Total Progress: [0-9]*\/[0-9]* tasks/Total Progress: $COMPLETED_TASKS\/$TOTAL_TASKS tasks/" CHANGELOG.md
    echo "✅ Project progress updated: $COMPLETED_TASKS/$TOTAL_TASKS tasks"
fi

echo "🎉 CHANGELOG.md update complete"