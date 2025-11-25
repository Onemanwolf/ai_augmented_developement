#!/bin/bash

# Load Test Runner Script
# Runs K6 load tests against the e-commerce services

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/reports/load"

# Default configuration
TEST_TYPE=${1:-smoke}
BASE_URL=${BASE_URL:-http://localhost:8080}
ORDER_SERVICE=${ORDER_SERVICE:-$BASE_URL}
PAYMENT_SERVICE=${PAYMENT_SERVICE:-http://localhost:8081}
FULFILLMENT_SERVICE=${FULFILLMENT_SERVICE:-http://localhost:8082}

mkdir -p "${REPORT_DIR}"

echo "========================================="
echo "E-Commerce Load Test Runner"
echo "========================================="
echo ""
echo "Test Type: ${TEST_TYPE}"
echo "Order Service: ${ORDER_SERVICE}"
echo "Payment Service: ${PAYMENT_SERVICE}"
echo "Fulfillment Service: ${FULFILLMENT_SERVICE}"
echo ""

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo "❌ k6 is not installed. Please install it first:"
    echo "   brew install k6  (macOS)"
    echo "   apt install k6   (Ubuntu)"
    echo "   https://k6.io/docs/getting-started/installation/"
    exit 1
fi

# Run the appropriate test scenario
run_test() {
    local scenario=$1
    echo "Running ${scenario} test..."

    k6 run \
        --out json="${REPORT_DIR}/${scenario}-results.json" \
        --tag testid="${scenario}-$(date +%Y%m%d-%H%M%S)" \
        -e BASE_URL="${BASE_URL}" \
        -e ORDER_SERVICE="${ORDER_SERVICE}" \
        -e PAYMENT_SERVICE="${PAYMENT_SERVICE}" \
        -e FULFILLMENT_SERVICE="${FULFILLMENT_SERVICE}" \
        --scenario "${scenario}" \
        "${PROJECT_ROOT}/tests/load/k6-config.js"
}

case $TEST_TYPE in
    smoke)
        run_test "smoke"
        ;;
    load)
        run_test "load"
        ;;
    stress)
        run_test "stress"
        ;;
    spike)
        run_test "spike"
        ;;
    all)
        echo "Running all test scenarios..."
        for scenario in smoke load stress spike; do
            run_test "$scenario"
            echo ""
            sleep 30  # Cool down between tests
        done
        ;;
    *)
        echo "Unknown test type: ${TEST_TYPE}"
        echo "Available types: smoke, load, stress, spike, all"
        exit 1
        ;;
esac

echo ""
echo "========================================="
echo "Load tests completed!"
echo "Results saved to: ${REPORT_DIR}"
echo "========================================="
