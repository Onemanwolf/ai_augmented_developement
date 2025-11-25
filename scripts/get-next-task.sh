#!/bin/bash
# get-next-task.sh - Find the next unblocked task to work on

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
TASKS_FILE="$PROJECT_ROOT/tasks.json"

# Color codes
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

# Check if jq is installed
if ! command -v jq &> /dev/null; then
    echo "Error: jq is required but not installed."
    echo "Install with: brew install jq"
    exit 1
fi

# Check if tasks.json exists
if [ ! -f "$TASKS_FILE" ]; then
    echo "Error: tasks.json not found at $TASKS_FILE"
    exit 1
fi

echo "Analyzing tasks from tasks.json..."
echo "=================================="
echo ""

# Get all completed task IDs
COMPLETED_TASKS=$(jq -r '
    [.phases[].sections[].tasks[] | select(.status == "completed") | .id] | @json
' "$TASKS_FILE")

# Find next pending task with all dependencies completed
NEXT_TASK=$(jq -r --argjson completed "$COMPLETED_TASKS" '
    .phases[].sections[].tasks[] |
    select(.status == "pending") |
    select(
        (.dependencies | length == 0) or
        (all(.dependencies[]; . as $dep | $completed | contains([$dep])))
    ) |
    .id
' "$TASKS_FILE" | head -1)

if [ -z "$NEXT_TASK" ]; then
    # Check if there are any in_progress tasks
    IN_PROGRESS=$(jq -r '
        .phases[].sections[].tasks[] | select(.status == "in_progress") | .id
    ' "$TASKS_FILE" | head -1)

    if [ -n "$IN_PROGRESS" ]; then
        echo -e "${YELLOW}Task in progress:${NC} $IN_PROGRESS"
        echo ""
        jq -r --arg id "$IN_PROGRESS" '
            .phases[].sections[].tasks[] | select(.id == $id) |
            "Description: \(.description)\nType: \(.type)\nAgent Role: \(.recommendedAgentRole)"
        ' "$TASKS_FILE"
    else
        echo "No pending tasks with satisfied dependencies found."
        echo "All tasks may be completed or blocked."
    fi
    exit 0
fi

echo -e "${GREEN}Next available task:${NC} $NEXT_TASK"
echo ""

# Get task details
jq -r --arg id "$NEXT_TASK" '
    .phases[].sections[].tasks[] | select(.id == $id) |
    "Description: \(.description)
Type: \(.type)
Recommended Agent: \(.recommendedAgentRole)
Dependencies: \(if .dependencies | length > 0 then .dependencies | join(", ") else "None" end)

Definition of Done:
\(.definitionOfDone | map("  - " + .) | join("\n"))

Acceptance Criteria:
\(.acceptanceCriteria | map("  - " + .) | join("\n"))

Expected Artifacts:
\(.expectedArtifacts | map("  - " + .) | join("\n"))
"
' "$TASKS_FILE"

echo ""
echo -e "${CYAN}To start working on this task:${NC}"
echo "  git checkout -b feature/task-$NEXT_TASK"
echo ""
echo -e "${CYAN}When complete, commit with:${NC}"
echo "  git commit -m \"[TASK-$NEXT_TASK] <description>\""

# Show guidelines reference if available
GUIDELINES_REF=$(jq -r --arg id "$NEXT_TASK" '
    .phases[].sections[].tasks[] | select(.id == $id) | .guidelinesRef // empty
' "$TASKS_FILE")

if [ -n "$GUIDELINES_REF" ]; then
    echo ""
    echo -e "${CYAN}Guidelines Reference:${NC} $GUIDELINES_REF"
fi
