#!/bin/bash

# Quality Results Notification Script
# Sends quality gate results to Slack and other webhooks

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
REPORT_DIR="${PROJECT_ROOT}/reports/quality"

# Configuration
SLACK_WEBHOOK_URL=${SLACK_WEBHOOK_URL:-""}
TEAMS_WEBHOOK_URL=${TEAMS_WEBHOOK_URL:-""}
CUSTOM_WEBHOOK_URL=${CUSTOM_WEBHOOK_URL:-""}

# Read quality report
REPORT_FILE="${REPORT_DIR}/quality-report.json"

if [[ ! -f "$REPORT_FILE" ]]; then
    echo "Quality report not found: ${REPORT_FILE}"
    exit 1
fi

# Parse report
STATUS=$(jq -r '.status' "$REPORT_FILE")
TOTAL_GATES=$(jq -r '.summary.total_gates' "$REPORT_FILE")
PASSED=$(jq -r '.summary.passed' "$REPORT_FILE")
FAILED=$(jq -r '.summary.failed' "$REPORT_FILE")
PASS_RATE=$(jq -r '.summary.pass_rate' "$REPORT_FILE")
TIMESTAMP=$(jq -r '.timestamp' "$REPORT_FILE")

# Determine emoji and color based on status
if [[ "$STATUS" == "PASSED" ]]; then
    EMOJI="✅"
    COLOR="good"
    COLOR_HEX="#36a64f"
else
    EMOJI="❌"
    COLOR="danger"
    COLOR_HEX="#dc3545"
fi

# Build message
build_slack_message() {
    cat << EOF
{
  "attachments": [
    {
      "color": "${COLOR}",
      "blocks": [
        {
          "type": "header",
          "text": {
            "type": "plain_text",
            "text": "${EMOJI} Quality Gates ${STATUS}",
            "emoji": true
          }
        },
        {
          "type": "section",
          "fields": [
            {
              "type": "mrkdwn",
              "text": "*Total Gates:*\n${TOTAL_GATES}"
            },
            {
              "type": "mrkdwn",
              "text": "*Passed:*\n${PASSED}"
            },
            {
              "type": "mrkdwn",
              "text": "*Failed:*\n${FAILED}"
            },
            {
              "type": "mrkdwn",
              "text": "*Pass Rate:*\n${PASS_RATE}%"
            }
          ]
        },
        {
          "type": "context",
          "elements": [
            {
              "type": "mrkdwn",
              "text": "Run at: ${TIMESTAMP}"
            }
          ]
        }
      ]
    }
  ]
}
EOF
}

build_teams_message() {
    cat << EOF
{
  "@type": "MessageCard",
  "@context": "http://schema.org/extensions",
  "themeColor": "${COLOR_HEX}",
  "summary": "Quality Gates ${STATUS}",
  "sections": [{
    "activityTitle": "${EMOJI} Quality Gates ${STATUS}",
    "facts": [
      { "name": "Total Gates", "value": "${TOTAL_GATES}" },
      { "name": "Passed", "value": "${PASSED}" },
      { "name": "Failed", "value": "${FAILED}" },
      { "name": "Pass Rate", "value": "${PASS_RATE}%" }
    ],
    "markdown": true
  }]
}
EOF
}

# Send notifications
send_slack() {
    if [[ -n "$SLACK_WEBHOOK_URL" ]]; then
        echo "Sending Slack notification..."
        curl -s -X POST \
            -H 'Content-type: application/json' \
            --data "$(build_slack_message)" \
            "$SLACK_WEBHOOK_URL"
        echo "Slack notification sent"
    else
        echo "SLACK_WEBHOOK_URL not set, skipping Slack notification"
    fi
}

send_teams() {
    if [[ -n "$TEAMS_WEBHOOK_URL" ]]; then
        echo "Sending Teams notification..."
        curl -s -X POST \
            -H 'Content-type: application/json' \
            --data "$(build_teams_message)" \
            "$TEAMS_WEBHOOK_URL"
        echo "Teams notification sent"
    else
        echo "TEAMS_WEBHOOK_URL not set, skipping Teams notification"
    fi
}

send_custom_webhook() {
    if [[ -n "$CUSTOM_WEBHOOK_URL" ]]; then
        echo "Sending custom webhook..."
        curl -s -X POST \
            -H 'Content-type: application/json' \
            --data "$(cat "$REPORT_FILE")" \
            "$CUSTOM_WEBHOOK_URL"
        echo "Custom webhook sent"
    fi
}

main() {
    echo "========================================="
    echo "Quality Results Notification"
    echo "========================================="
    echo ""
    echo "Status: ${STATUS}"
    echo "Gates: ${PASSED}/${TOTAL_GATES} passed"
    echo ""

    send_slack
    send_teams
    send_custom_webhook

    echo ""
    echo "Notifications complete"
}

main "$@"
