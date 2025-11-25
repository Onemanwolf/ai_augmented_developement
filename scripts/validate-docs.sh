#!/bin/bash
# validate-docs.sh - Cross-document consistency validation

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ERRORS=0
WARNINGS=0

echo "Validating document consistency..."
echo "=================================="
echo ""

# Check all required documents exist
echo "Checking required documents..."
DOCS=("Requirements.md" "Plan.md" "Task.md" "Guidelines.md" "CHANGELOG.md" "Setup.md" "tasks.json")

for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        echo -e "${GREEN}[OK]${NC} $doc exists"
    else
        echo -e "${RED}[ERROR]${NC} $doc missing"
        ERRORS=$((ERRORS + 1))
    fi
done

echo ""

# Check task counts (requires jq)
if command -v jq &> /dev/null && [ -f "tasks.json" ]; then
    echo "Checking task counts..."

    # Get total from tasks.json metadata
    TASKS_JSON_META=$(jq -r '.metadata.totalTasks' tasks.json)

    # Calculate actual total from phases
    TASKS_JSON_ACTUAL=$(jq '[.phases[].totalTasks] | add' tasks.json)

    if [ "$TASKS_JSON_META" != "$TASKS_JSON_ACTUAL" ]; then
        echo -e "${RED}[ERROR]${NC} tasks.json internal mismatch: metadata=$TASKS_JSON_META, phases sum=$TASKS_JSON_ACTUAL"
        ERRORS=$((ERRORS + 1))
    else
        echo -e "${GREEN}[OK]${NC} tasks.json task count: $TASKS_JSON_META"
    fi

    # Check CHANGELOG.md project summary
    if [ -f "CHANGELOG.md" ]; then
        CHANGELOG_TOTAL=$(grep -oE '[0-9]+/[0-9]+ tasks' CHANGELOG.md | head -1 | grep -oE '/[0-9]+' | tr -d '/')
        if [ -n "$CHANGELOG_TOTAL" ] && [ "$CHANGELOG_TOTAL" != "$TASKS_JSON_META" ]; then
            echo -e "${YELLOW}[WARNING]${NC} CHANGELOG.md total ($CHANGELOG_TOTAL) differs from tasks.json ($TASKS_JSON_META)"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi

    # Check Task.md summary table
    if [ -f "Task.md" ]; then
        TASK_MD_TOTAL=$(grep -E '^\| \*\*Total\*\*' Task.md | grep -oE '[0-9]+' | head -1)
        if [ -n "$TASK_MD_TOTAL" ] && [ "$TASK_MD_TOTAL" != "$TASKS_JSON_META" ]; then
            echo -e "${RED}[ERROR]${NC} Task.md total ($TASK_MD_TOTAL) differs from tasks.json ($TASKS_JSON_META)"
            ERRORS=$((ERRORS + 1))
        else
            echo -e "${GREEN}[OK]${NC} Task.md total matches tasks.json"
        fi
    fi
else
    echo -e "${YELLOW}[SKIP]${NC} Task count validation skipped (jq not installed or tasks.json missing)"
fi

echo ""

# Check cross-references in document headers
echo "Checking cross-references..."

# Check that documents reference each other appropriately
for doc in "Requirements.md" "Plan.md" "Task.md" "Guidelines.md"; do
    if [ -f "$doc" ]; then
        if grep -q "Document Cross-References" "$doc" || grep -q "Cross-References" "$doc"; then
            echo -e "${GREEN}[OK]${NC} $doc has cross-reference section"
        else
            echo -e "${YELLOW}[WARNING]${NC} $doc missing cross-reference section"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
done

echo ""

# Check for placeholder dates
echo "Checking for placeholder values..."
for doc in "${DOCS[@]}"; do
    if [ -f "$doc" ]; then
        # Check for XX-XX or YYYY placeholders (but not TBD in version history tables)
        REAL_PLACEHOLDERS=$(grep -E "XX-XX|YYYY-MM" "$doc" 2>/dev/null | wc -l | xargs)
        if [ "$REAL_PLACEHOLDERS" -gt 0 ]; then
            echo -e "${YELLOW}[WARNING]${NC} $doc contains $REAL_PLACEHOLDERS date placeholder(s)"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
done

echo ""

# Check scripts are executable
echo "Checking script permissions..."
for script in scripts/*.sh; do
    if [ -f "$script" ]; then
        if [ -x "$script" ]; then
            echo -e "${GREEN}[OK]${NC} $script is executable"
        else
            echo -e "${YELLOW}[WARNING]${NC} $script is not executable"
            WARNINGS=$((WARNINGS + 1))
        fi
    fi
done

echo ""

# Validate JSON syntax
echo "Checking JSON syntax..."
if [ -f "tasks.json" ]; then
    if command -v jq &> /dev/null; then
        if jq empty tasks.json 2>/dev/null; then
            echo -e "${GREEN}[OK]${NC} tasks.json is valid JSON"
        else
            echo -e "${RED}[ERROR]${NC} tasks.json has invalid JSON syntax"
            ERRORS=$((ERRORS + 1))
        fi
    else
        if command -v python3 &> /dev/null; then
            if python3 -m json.tool tasks.json > /dev/null 2>&1; then
                echo -e "${GREEN}[OK]${NC} tasks.json is valid JSON"
            else
                echo -e "${RED}[ERROR]${NC} tasks.json has invalid JSON syntax"
                ERRORS=$((ERRORS + 1))
            fi
        else
            echo -e "${YELLOW}[SKIP]${NC} JSON validation skipped (jq/python3 not available)"
        fi
    fi
fi

echo ""
echo "=================================="

if [ $ERRORS -gt 0 ]; then
    echo -e "${RED}Validation FAILED: $ERRORS error(s), $WARNINGS warning(s)${NC}"
    exit 1
elif [ $WARNINGS -gt 0 ]; then
    echo -e "${YELLOW}Validation PASSED with $WARNINGS warning(s)${NC}"
    exit 0
else
    echo -e "${GREEN}All validations passed!${NC}"
    exit 0
fi
