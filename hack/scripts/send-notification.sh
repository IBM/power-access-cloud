#!/usr/bin/env bash
# Script to send Slack notifications about the automation status

set -o nounset
set -o pipefail

# Configuration
WORK_DIR="${WORK_DIR:-/tmp/centos-images}"
SLACK_WEBHOOK_URL="${SLACK_WEBHOOK_URL:-}"
SLACK_BOT_TOKEN="${SLACK_BOT_TOKEN:-}"
SLACK_CHANNEL_ID="${SLACK_CHANNEL_ID:-}"
JOB_STATUS="${JOB_STATUS:-unknown}"
ENVIRONMENT="${ENVIRONMENT:-production}"
GITHUB_RUN_ID="${GITHUB_RUN_ID:-}"
GITHUB_REPOSITORY="${GITHUB_REPOSITORY:-}"
GITHUB_SERVER_URL="${GITHUB_SERVER_URL:-https://github.com}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Slack is configured (webhook or bot token + channel)
check_slack_config() {
    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        log_info "Using Slack Webhook URL"
        return 0
    elif [ -n "$SLACK_BOT_TOKEN" ] && [ -n "$SLACK_CHANNEL_ID" ]; then
        log_info "Using Slack Bot Token with Channel ID: $SLACK_CHANNEL_ID"
        return 0
    else
        log_warn "Slack not configured. Set either:"
        log_warn "  - SLACK_WEBHOOK_URL, or"
        log_warn "  - SLACK_BOT_TOKEN + SLACK_CHANNEL_ID"
        return 1
    fi
}

# Get image metadata
get_metadata() {
    local metadata_file="${WORK_DIR}/image-metadata.json"
    
    if [ -f "$metadata_file" ]; then
        cat "$metadata_file"
    else
        echo "{}"
    fi
}

# Build success message
build_success_message() {
    local metadata=$(get_metadata)
    local image_name=$(echo "$metadata" | jq -r '.image_name // "N/A"')
    local centos_version=$(echo "$metadata" | jq -r '.centos_version // "N/A"')
    local powervs_image_id=$(echo "$metadata" | jq -r '.powervs_image_id // "N/A"')
    local cos_bucket=$(echo "$metadata" | jq -r '.cos_bucket // "N/A"')
    local cos_object=$(echo "$metadata" | jq -r '.cos_object_name // "N/A"')
    
    local workflow_url=""
    if [ -n "$GITHUB_RUN_ID" ] && [ -n "$GITHUB_REPOSITORY" ]; then
        workflow_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
    fi
    
    # Determine environment emoji
    local env_emoji="🚀"
    if [ "$ENVIRONMENT" = "staging" ]; then
        env_emoji="🧪"
    fi
    
    cat <<EOF
{
  "text": "✅ CentOS Image Automation - Success (${ENVIRONMENT})",
  "blocks": [
    {
      "type": "header",
      "text": {
        "type": "plain_text",
        "text": "✅ CentOS Image Update - Success",
        "emoji": true
      }
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "${env_emoji} *Environment:* \`${ENVIRONMENT}\`"
      }
    },
    {
      "type": "section",
      "fields": [
        {
          "type": "mrkdwn",
          "text": "*Image Name:*\n\`${image_name}\`"
        },
        {
          "type": "mrkdwn",
          "text": "*CentOS Version:*\n${centos_version}"
        },
        {
          "type": "mrkdwn",
          "text": "*PowerVS Image ID:*\n\`${powervs_image_id}\`"
        },
        {
          "type": "mrkdwn",
          "text": "*COS Bucket:*\n\`${cos_bucket}\`"
        }
      ]
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*COS Object:* \`${cos_object}\`"
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Next Steps:*\n• Verify image in PowerVS console\n• Test deploy a VM with the new image\n• Update Catalog CRD with new image ID"
      }
    }
EOF

    if [ -n "$workflow_url" ]; then
        cat <<EOF
    ,
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "<${workflow_url}|View Workflow Run>"
      }
    }
EOF
    fi

    cat <<EOF
  ]
}
EOF
}

# Build failure message
build_failure_message() {
    local workflow_url=""
    if [ -n "$GITHUB_RUN_ID" ] && [ -n "$GITHUB_REPOSITORY" ]; then
        workflow_url="${GITHUB_SERVER_URL}/${GITHUB_REPOSITORY}/actions/runs/${GITHUB_RUN_ID}"
    fi
    
    # Try to get error from logs
    local error_msg="Check workflow logs for details"
    if [ -f "${WORK_DIR}/conversion.log" ]; then
        error_msg=$(tail -20 "${WORK_DIR}/conversion.log" | grep -i "error" | head -5 | sed 's/"/\\"/g' || echo "Check conversion.log for details")
    fi
    
    # Determine environment emoji
    local env_emoji="🚀"
    if [ "$ENVIRONMENT" = "staging" ]; then
        env_emoji="🧪"
    fi
    
    cat <<EOF
{
  "text": "❌ CentOS Image Automation - Failed (${ENVIRONMENT})",
  "blocks": [
    {
      "type": "header",
      "text": {
        "type": "plain_text",
        "text": "❌ CentOS Image Update - Failed",
        "emoji": true
      }
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "${env_emoji} *Environment:* \`${ENVIRONMENT}\`\n*Status:* Failed\n*Error:* ${error_msg}"
      }
    },
    {
      "type": "divider"
    },
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "*Action Required:*\n• Review workflow logs\n• Check PowerVM connectivity\n• Verify IBM Cloud credentials\n• Check disk space on PowerVM"
      }
    }
EOF

    if [ -n "$workflow_url" ]; then
        cat <<EOF
    ,
    {
      "type": "section",
      "text": {
        "type": "mrkdwn",
        "text": "<${workflow_url}|View Workflow Run>"
      }
    }
EOF
    fi

    cat <<EOF
  ]
}
EOF
}

# Send notification to Slack via Webhook
send_via_webhook() {
    local message="$1"
    
    log_info "Sending via Webhook..."
    
    local response=$(curl -s -X POST \
        -H 'Content-Type: application/json' \
        -d "$message" \
        "$SLACK_WEBHOOK_URL")
    
    if [ "$response" == "ok" ]; then
        log_info "Notification sent successfully via webhook"
        return 0
    else
        log_error "Failed to send via webhook: $response"
        return 1
    fi
}

# Send notification to Slack via Bot Token
send_via_bot_token() {
    local message="$1"
    
    log_info "Sending via Bot Token to channel: $SLACK_CHANNEL_ID"
    
    # Build the payload properly using jq to avoid JSON escaping issues
    local payload=$(echo "$message" | jq -c --arg channel "$SLACK_CHANNEL_ID" '. + {channel: $channel}')
    
    # Debug: Show payload (first 200 chars)
    log_info "Payload preview: ${payload:0:200}..."
    
    local response=$(curl -s -X POST \
        -H "Authorization: Bearer $SLACK_BOT_TOKEN" \
        -H 'Content-Type: application/json' \
        -d "$payload" \
        "https://slack.com/api/chat.postMessage")
    
    local ok=$(echo "$response" | jq -r '.ok')
    
    if [ "$ok" == "true" ]; then
        log_info "Notification sent successfully via bot token"
        return 0
    else
        local error=$(echo "$response" | jq -r '.error // "unknown error"')
        log_error "Failed to send via bot token: $error"
        log_error "Response: $response"
        return 1
    fi
}

# Send notification to Slack (auto-detect method)
send_notification() {
    local message="$1"
    
    log_info "Sending notification to Slack..."
    
    if [ -n "$SLACK_WEBHOOK_URL" ]; then
        send_via_webhook "$message"
    elif [ -n "$SLACK_BOT_TOKEN" ] && [ -n "$SLACK_CHANNEL_ID" ]; then
        send_via_bot_token "$message"
    else
        log_error "No Slack configuration found"
        return 1
    fi
}

# Main execution
main() {
    log_info "Preparing notification"
    log_info "Job status: $JOB_STATUS"
    
    # Check if Slack is configured
    if ! check_slack_config; then
        log_info "Skipping notification"
        exit 0
    fi
    
    # Build message based on status
    local message=""
    case "$JOB_STATUS" in
        success)
            log_info "Building success message..."
            message=$(build_success_message)
            ;;
        failure)
            log_info "Building failure message..."
            message=$(build_failure_message)
            ;;
        *)
            log_warn "Unknown job status: $JOB_STATUS"
            message=$(build_failure_message)
            ;;
    esac
    
    # Send notification
    if send_notification "$message"; then
        log_info "Notification process completed"
    else
        log_warn "Notification failed but continuing"
    fi
}

# Run main function
main "$@"


