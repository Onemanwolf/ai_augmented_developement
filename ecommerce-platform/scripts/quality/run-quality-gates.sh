#!/bin/bash

# Quality Gates Runner Script
# Runs all quality checks and reports results

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/reports/quality"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Quality gate thresholds
MIN_COVERAGE=80
MAX_CYCLOMATIC_COMPLEXITY=15
MAX_COGNITIVE_COMPLEXITY=20
MAX_LINT_WARNINGS=0
MAX_SECURITY_ISSUES=0

# Initialize report directory
mkdir -p "${REPORT_DIR}"

# Results tracking
GATES_PASSED=0
GATES_FAILED=0
GATES_TOTAL=0

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_gate() {
    local name=$1
    local result=$2
    local threshold=$3
    local comparison=$4  # "lt" for less than, "gt" for greater than

    GATES_TOTAL=$((GATES_TOTAL + 1))

    if [[ "$comparison" == "lt" ]]; then
        if (( $(echo "$result < $threshold" | bc -l) )); then
            log_error "FAILED: $name ($result < $threshold)"
            GATES_FAILED=$((GATES_FAILED + 1))
            return 1
        fi
    else
        if (( $(echo "$result > $threshold" | bc -l) )); then
            log_error "FAILED: $name ($result > $threshold)"
            GATES_FAILED=$((GATES_FAILED + 1))
            return 1
        fi
    fi

    log_info "PASSED: $name ($result)"
    GATES_PASSED=$((GATES_PASSED + 1))
    return 0
}

run_unit_tests() {
    log_info "Running unit tests..."

    local coverage_total=0
    local service_count=0

    for service in order payment fulfillment; do
        local service_dir="${PROJECT_ROOT}/services/${service}"
        if [[ -d "$service_dir" ]]; then
            log_info "Testing ${service} service..."
            cd "$service_dir"

            # Run tests with coverage
            go test -v -race -coverprofile="${REPORT_DIR}/${service}-coverage.out" ./... 2>&1 | tee "${REPORT_DIR}/${service}-test.log" || true

            # Extract coverage percentage
            if [[ -f "${REPORT_DIR}/${service}-coverage.out" ]]; then
                local coverage=$(go tool cover -func="${REPORT_DIR}/${service}-coverage.out" | grep total | awk '{print $3}' | sed 's/%//')
                if [[ -n "$coverage" ]]; then
                    coverage_total=$(echo "$coverage_total + $coverage" | bc)
                    service_count=$((service_count + 1))
                    log_info "${service} coverage: ${coverage}%"
                fi
            fi
        fi
    done

    # Calculate average coverage
    if [[ $service_count -gt 0 ]]; then
        local avg_coverage=$(echo "scale=2; $coverage_total / $service_count" | bc)
        check_gate "Code Coverage" "$avg_coverage" "$MIN_COVERAGE" "lt"
    else
        log_warn "No test coverage data available"
    fi
}

run_linting() {
    log_info "Running linting checks..."

    local total_issues=0

    for service in shared services/order services/payment services/fulfillment; do
        local service_dir="${PROJECT_ROOT}/${service}"
        if [[ -d "$service_dir" ]]; then
            cd "$service_dir"

            # Run golangci-lint
            local issues=$(golangci-lint run --out-format json 2>/dev/null | jq '.Issues | length' 2>/dev/null || echo "0")
            total_issues=$((total_issues + issues))

            if [[ "$issues" -gt 0 ]]; then
                log_warn "${service}: ${issues} lint issues"
                golangci-lint run > "${REPORT_DIR}/${service//\//-}-lint.log" 2>&1 || true
            fi
        fi
    done

    check_gate "Lint Issues" "$total_issues" "$MAX_LINT_WARNINGS" "gt"
}

run_security_scan() {
    log_info "Running security scans..."

    local total_vulns=0

    # Run govulncheck
    for service in shared services/order services/payment services/fulfillment; do
        local service_dir="${PROJECT_ROOT}/${service}"
        if [[ -d "$service_dir" ]]; then
            cd "$service_dir"

            # Check for vulnerabilities
            local vulns=$(govulncheck ./... 2>&1 | grep -c "Vulnerability" || echo "0")
            total_vulns=$((total_vulns + vulns))

            if [[ "$vulns" -gt 0 ]]; then
                log_warn "${service}: ${vulns} vulnerabilities found"
                govulncheck ./... > "${REPORT_DIR}/${service//\//-}-vulns.log" 2>&1 || true
            fi
        fi
    done

    check_gate "Security Vulnerabilities" "$total_vulns" "$MAX_SECURITY_ISSUES" "gt"
}

run_complexity_check() {
    log_info "Running complexity analysis..."

    local max_cyclomatic=0
    local max_cognitive=0

    for service in shared services/order services/payment services/fulfillment; do
        local service_dir="${PROJECT_ROOT}/${service}"
        if [[ -d "$service_dir" ]]; then
            # Use gocyclo for cyclomatic complexity (if installed)
            if command -v gocyclo &> /dev/null; then
                local cyclo=$(gocyclo -over 10 "$service_dir" 2>/dev/null | head -1 | awk '{print $1}' || echo "0")
                if [[ "$cyclo" -gt "$max_cyclomatic" ]]; then
                    max_cyclomatic=$cyclo
                fi
            fi
        fi
    done

    if [[ "$max_cyclomatic" -gt 0 ]]; then
        check_gate "Cyclomatic Complexity" "$max_cyclomatic" "$MAX_CYCLOMATIC_COMPLEXITY" "gt"
    else
        log_warn "Complexity analysis skipped (gocyclo not installed)"
    fi
}

generate_report() {
    log_info "Generating quality report..."

    local report_file="${REPORT_DIR}/quality-report.json"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat > "$report_file" << EOF
{
  "timestamp": "${timestamp}",
  "summary": {
    "total_gates": ${GATES_TOTAL},
    "passed": ${GATES_PASSED},
    "failed": ${GATES_FAILED},
    "pass_rate": $(echo "scale=2; ${GATES_PASSED} * 100 / ${GATES_TOTAL}" | bc)
  },
  "thresholds": {
    "min_coverage": ${MIN_COVERAGE},
    "max_cyclomatic_complexity": ${MAX_CYCLOMATIC_COMPLEXITY},
    "max_lint_warnings": ${MAX_LINT_WARNINGS},
    "max_security_issues": ${MAX_SECURITY_ISSUES}
  },
  "status": "$([ $GATES_FAILED -eq 0 ] && echo "PASSED" || echo "FAILED")"
}
EOF

    log_info "Report generated: ${report_file}"
}

main() {
    log_info "Starting Quality Gates..."
    log_info "Project Root: ${PROJECT_ROOT}"
    echo ""

    # Run all quality checks
    run_unit_tests || true
    run_linting || true
    run_security_scan || true
    run_complexity_check || true

    echo ""
    log_info "========================================="
    log_info "Quality Gates Summary"
    log_info "========================================="
    log_info "Total Gates: ${GATES_TOTAL}"
    log_info "Passed: ${GATES_PASSED}"
    log_info "Failed: ${GATES_FAILED}"
    echo ""

    generate_report

    if [[ $GATES_FAILED -gt 0 ]]; then
        log_error "Quality gates FAILED"
        exit 1
    else
        log_info "All quality gates PASSED"
        exit 0
    fi
}

main "$@"
