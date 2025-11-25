#!/bin/bash

# Setup Debezium MongoDB Connectors for Outbox Pattern
# This script registers the CDC connectors with Kafka Connect

set -euo pipefail

CONNECT_URL="${KAFKA_CONNECT_URL:-http://kafka-connect:8083}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "Waiting for Kafka Connect to be ready..."
until curl -s "${CONNECT_URL}/connectors" > /dev/null 2>&1; do
    echo "Kafka Connect is not ready yet. Waiting..."
    sleep 5
done

echo "Kafka Connect is ready. Registering connectors..."

# Function to register a connector
register_connector() {
    local connector_file=$1
    local connector_name=$(jq -r '.name' "${connector_file}")

    echo "Registering connector: ${connector_name}"

    # Check if connector already exists
    if curl -s "${CONNECT_URL}/connectors/${connector_name}" > /dev/null 2>&1; then
        echo "Connector ${connector_name} already exists. Updating..."
        curl -X PUT \
            -H "Content-Type: application/json" \
            -d @"${connector_file}" \
            "${CONNECT_URL}/connectors/${connector_name}/config"
    else
        echo "Creating new connector ${connector_name}..."
        curl -X POST \
            -H "Content-Type: application/json" \
            -d @"${connector_file}" \
            "${CONNECT_URL}/connectors"
    fi

    echo ""
}

# Register all connectors
register_connector "${SCRIPT_DIR}/order-outbox-connector.json"
register_connector "${SCRIPT_DIR}/payment-outbox-connector.json"
register_connector "${SCRIPT_DIR}/fulfillment-outbox-connector.json"

echo "All connectors registered successfully!"

# List all connectors
echo ""
echo "Current connectors:"
curl -s "${CONNECT_URL}/connectors" | jq '.'
