#!/bin/bash

# Code Coverage Check Script
# Verifies code coverage meets minimum threshold

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"

MIN_COVERAGE=${MIN_COVERAGE:-80}
REPORT_DIR="${PROJECT_ROOT}/reports/coverage"

mkdir -p "${REPORT_DIR}"

echo "Checking code coverage (minimum: ${MIN_COVERAGE}%)..."

coverage_results=()
failed=false

for service in order payment fulfillment; do
    service_dir="${PROJECT_ROOT}/services/${service}"

    if [[ -d "$service_dir" ]]; then
        echo "Testing ${service} service..."
        cd "$service_dir"

        # Run tests with coverage
        go test -coverprofile="${REPORT_DIR}/${service}.out" -covermode=atomic ./... 2>/dev/null || true

        if [[ -f "${REPORT_DIR}/${service}.out" ]]; then
            # Get coverage percentage
            coverage=$(go tool cover -func="${REPORT_DIR}/${service}.out" 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//' || echo "0")

            if [[ -n "$coverage" && "$coverage" != "0" ]]; then
                coverage_results+=("${service}:${coverage}")

                # Generate HTML report
                go tool cover -html="${REPORT_DIR}/${service}.out" -o "${REPORT_DIR}/${service}.html" 2>/dev/null || true

                # Check against threshold
                if (( $(echo "$coverage < $MIN_COVERAGE" | bc -l) )); then
                    echo "❌ ${service}: ${coverage}% (below ${MIN_COVERAGE}%)"
                    failed=true
                else
                    echo "✅ ${service}: ${coverage}%"
                fi
            else
                echo "⚠️  ${service}: No coverage data (no tests?)"
            fi
        fi
    fi
done

# Test shared package
shared_dir="${PROJECT_ROOT}/shared"
if [[ -d "$shared_dir" ]]; then
    echo "Testing shared package..."
    cd "$shared_dir"

    go test -coverprofile="${REPORT_DIR}/shared.out" -covermode=atomic ./... 2>/dev/null || true

    if [[ -f "${REPORT_DIR}/shared.out" ]]; then
        coverage=$(go tool cover -func="${REPORT_DIR}/shared.out" 2>/dev/null | grep total | awk '{print $3}' | sed 's/%//' || echo "0")

        if [[ -n "$coverage" && "$coverage" != "0" ]]; then
            coverage_results+=("shared:${coverage}")
            go tool cover -html="${REPORT_DIR}/shared.out" -o "${REPORT_DIR}/shared.html" 2>/dev/null || true

            if (( $(echo "$coverage < $MIN_COVERAGE" | bc -l) )); then
                echo "❌ shared: ${coverage}% (below ${MIN_COVERAGE}%)"
                failed=true
            else
                echo "✅ shared: ${coverage}%"
            fi
        fi
    fi
fi

echo ""
echo "Coverage Summary:"
echo "================="
for result in "${coverage_results[@]}"; do
    echo "  $result%"
done

if $failed; then
    echo ""
    echo "❌ Coverage check FAILED - some packages below ${MIN_COVERAGE}%"
    exit 1
else
    echo ""
    echo "✅ Coverage check PASSED"
    exit 0
fi
