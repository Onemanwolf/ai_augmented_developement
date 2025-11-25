#!/bin/bash

# Security Scanning Script
# Runs multiple security checks on the codebase

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/reports/security"

mkdir -p "${REPORT_DIR}"

echo "Running security scans..."
echo ""

vulnerabilities_found=0
issues_found=0

# Check if tools are available
check_tool() {
    if ! command -v "$1" &> /dev/null; then
        echo "⚠️  $1 not installed, skipping..."
        return 1
    fi
    return 0
}

# Run govulncheck
run_govulncheck() {
    echo "🔍 Running govulncheck..."

    for dir in shared services/order services/payment services/fulfillment; do
        full_path="${PROJECT_ROOT}/${dir}"
        if [[ -d "$full_path" ]]; then
            echo "  Scanning ${dir}..."
            cd "$full_path"

            output=$(govulncheck ./... 2>&1 || true)
            echo "$output" > "${REPORT_DIR}/${dir//\//-}-govulncheck.txt"

            vuln_count=$(echo "$output" | grep -c "Vulnerability" || echo "0")
            if [[ "$vuln_count" -gt 0 ]]; then
                echo "    ❌ Found ${vuln_count} vulnerabilities"
                vulnerabilities_found=$((vulnerabilities_found + vuln_count))
            else
                echo "    ✅ No vulnerabilities found"
            fi
        fi
    done
}

# Run gosec
run_gosec() {
    echo ""
    echo "🔍 Running gosec..."

    for dir in shared services/order services/payment services/fulfillment; do
        full_path="${PROJECT_ROOT}/${dir}"
        if [[ -d "$full_path" ]]; then
            echo "  Scanning ${dir}..."
            cd "$full_path"

            output=$(gosec -fmt json ./... 2>&1 || true)
            echo "$output" > "${REPORT_DIR}/${dir//\//-}-gosec.json"

            issue_count=$(echo "$output" | jq '.Issues | length' 2>/dev/null || echo "0")
            if [[ "$issue_count" -gt 0 ]]; then
                echo "    ⚠️  Found ${issue_count} security issues"
                issues_found=$((issues_found + issue_count))
            else
                echo "    ✅ No security issues found"
            fi
        fi
    done
}

# Check for secrets in code
run_secrets_scan() {
    echo ""
    echo "🔍 Checking for hardcoded secrets..."

    # Simple pattern matching for common secrets
    local patterns=(
        "password\s*=\s*['\"][^'\"]+['\"]"
        "api_key\s*=\s*['\"][^'\"]+['\"]"
        "secret\s*=\s*['\"][^'\"]+['\"]"
        "token\s*=\s*['\"][^'\"]+['\"]"
        "AWS_ACCESS_KEY"
        "PRIVATE_KEY"
    )

    local secrets_found=0

    for pattern in "${patterns[@]}"; do
        matches=$(grep -r -i -E "$pattern" "${PROJECT_ROOT}" \
            --include="*.go" \
            --exclude-dir=".git" \
            --exclude-dir="vendor" \
            2>/dev/null | grep -v "_test.go" | grep -v "example" || true)

        if [[ -n "$matches" ]]; then
            echo "  ⚠️  Potential secrets found matching: $pattern"
            secrets_found=$((secrets_found + 1))
        fi
    done

    if [[ $secrets_found -eq 0 ]]; then
        echo "  ✅ No hardcoded secrets detected"
    else
        issues_found=$((issues_found + secrets_found))
    fi
}

# Check dependencies for known vulnerabilities
run_dependency_check() {
    echo ""
    echo "🔍 Checking dependencies..."

    for dir in shared services/order services/payment services/fulfillment; do
        full_path="${PROJECT_ROOT}/${dir}"
        if [[ -d "$full_path" && -f "${full_path}/go.mod" ]]; then
            echo "  Checking ${dir}..."
            cd "$full_path"

            # Check for outdated dependencies
            go list -m -u all 2>/dev/null | grep '\[' > "${REPORT_DIR}/${dir//\//-}-outdated.txt" || true

            outdated_count=$(wc -l < "${REPORT_DIR}/${dir//\//-}-outdated.txt" | tr -d ' ')
            if [[ "$outdated_count" -gt 0 ]]; then
                echo "    ⚠️  ${outdated_count} outdated dependencies"
            else
                echo "    ✅ All dependencies up to date"
            fi
        fi
    done
}

# Generate summary report
generate_report() {
    local report_file="${REPORT_DIR}/security-report.json"
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

    cat > "$report_file" << EOF
{
  "timestamp": "${timestamp}",
  "vulnerabilities": ${vulnerabilities_found},
  "security_issues": ${issues_found},
  "status": "$([ $((vulnerabilities_found + issues_found)) -eq 0 ] && echo "PASSED" || echo "FAILED")",
  "scans_run": ["govulncheck", "gosec", "secrets", "dependencies"]
}
EOF

    echo ""
    echo "Report saved to: ${report_file}"
}

main() {
    echo "========================================="
    echo "Security Scan Report"
    echo "========================================="
    echo ""

    if check_tool govulncheck; then
        run_govulncheck
    fi

    if check_tool gosec; then
        run_gosec
    fi

    run_secrets_scan
    run_dependency_check

    generate_report

    echo ""
    echo "========================================="
    echo "Summary"
    echo "========================================="
    echo "Vulnerabilities: ${vulnerabilities_found}"
    echo "Security Issues: ${issues_found}"
    echo ""

    if [[ $((vulnerabilities_found + issues_found)) -gt 0 ]]; then
        echo "❌ Security scan FAILED"
        exit 1
    else
        echo "✅ Security scan PASSED"
        exit 0
    fi
}

main "$@"
