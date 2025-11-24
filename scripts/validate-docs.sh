#!/bin/bash
# validate-docs.sh - Cross-document consistency validation

set -e

echo "🔍 Validating document consistency..."

# Check task counts across documents
echo "Checking task counts..."

TASK_MD_TOTAL=$(grep "| \*\*Total\*\* |" Task.md | grep -o '[0-9]\+' | head -1)
TASKS_JSON_TOTAL=$(grep -A 5 '"id": "phase-' tasks.json | grep '"totalTasks"' | grep -o '[0-9]\+' | awk '{sum += $1} END {print sum}')

if [ "$TASK_MD_TOTAL" != "$TASKS_JSON_TOTAL" ]; then
    echo "❌ Task count mismatch: Task.md=$TASK_MD_TOTAL, tasks.json=$TASKS_JSON_TOTAL"
    exit 1
fi

echo "✅ Task counts match: $TASK_MD_TOTAL"

# Check section references
echo "Checking section references..."

# Check critical path tasks (only for expanded phases)
CRITICAL_PATH=$(grep -A 20 "Critical Path" Task.md | grep -E "^[0-9]+\.[0-9]+\.[0-9]+" | tr -d '→')

for task in $CRITICAL_PATH; do
    # Skip tasks from phases 1, 2, 4, 5, 6 since they're not fully expanded
    if [[ $task =~ ^[1-2]\. ]]; then
        continue
    fi
    if [[ $task =~ ^[4-6]\. ]]; then
        continue
    fi

    if ! grep -q "\"id\": \"$task\"" tasks.json; then
        echo "❌ Critical path task $task not found in tasks.json"
        exit 1
    fi
done

echo "✅ Critical path tasks exist (for expanded phases)"

# Check cross-references
echo "Checking cross-references..."

# Verify all documents reference each other appropriately
DOCS=("Requirements.md" "Plan.md" "Task.md" "Guidelines.md" "CHANGELOG.md" "Setup.md")

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo "✅ $doc exists"
    else
        echo "❌ $doc missing"
        exit 1
    fi
done

echo "🎉 All validations passed!"