#!/usr/bin/env bash
set -euo pipefail

# ----------------------------
# Required env vars:
# BASE_URL, CLIENT_SECRET
#
# Optional env vars:
# CLIENT_ID (defaults to 'user-manage-client')
#
# Usage:
#   ./kc_user.sh --disable <username>
#   ./kc_user.sh --enable <username>
# ----------------------------

if [[ $# -ne 2 ]]; then
  echo "Usage: $0 [--enable|--disable] <username>"
  exit 2
fi

ACTION="$1"
USERNAME="$2"

case "$ACTION" in
  --enable)  ENABLE_VALUE=true ;;
  --disable) ENABLE_VALUE=false ;;
  *)
    echo "Invalid action: $ACTION"
    echo "Usage: $0 [--enable|--disable] <username>"
    exit 2
    ;;
esac

: "${BASE_URL:?Please set BASE_URL (eg https://auth.example.com)}"
CLIENT_ID="${CLIENT_ID:-user-manage-client}" && echo "Using CLIENT_ID: $CLIENT_ID"
: "${CLIENT_SECRET:?Please set CLIENT_SECRET for the mentioned CLIENT_ID}"

REALM=pac
TOKEN_ENDPOINT="${BASE_URL%/}/realms/${REALM}/protocol/openid-connect/token"
ADMIN_USERS_ENDPOINT_BASE="${BASE_URL%/}/admin/realms/${REALM}/users"

# 1) Get access token
TOKEN_JSON=$(curl -sS -X POST "$TOKEN_ENDPOINT" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  --data-urlencode "grant_type=client_credentials" \
  --data-urlencode "client_id=${CLIENT_ID}" \
  --data-urlencode "client_secret=${CLIENT_SECRET}" )

ACCESS_TOKEN=$(echo "$TOKEN_JSON" | sed -n 's/.*"access_token":"\([^"]*\)".*/\1/p')

if [[ -z "$ACCESS_TOKEN" ]]; then
  echo "Failed to obtain access token. Response:"
  echo "$TOKEN_JSON"
  exit 3
fi

# 2) Find user by username
SEARCH_RESP=$(curl -sS -X GET "${ADMIN_USERS_ENDPOINT_BASE}?username=${USERNAME}" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Accept: application/json" )

USER_ID=$(echo "$SEARCH_RESP" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p' | head -n1)
if [[ -z "$USER_ID" ]]; then
  echo "User '${USERNAME}' not found."
  exit 4
fi

# 3) Fetch user JSON
USER_JSON=$(curl -sS -X GET "${ADMIN_USERS_ENDPOINT_BASE}/${USER_ID}" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Accept: application/json" )

# 4) Replace enabled flag
UPDATED_JSON=$(echo "$USER_JSON" | sed -E "s/\"enabled\":(true|false)/\"enabled\":${ENABLE_VALUE}/")

# 5) Update user
PUT_RESP=$(curl -sS -w "\n%{http_code}" -X PUT "${ADMIN_USERS_ENDPOINT_BASE}/${USER_ID}" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  --data-binary "$UPDATED_JSON" )

HTTP_BODY=$(printf '%s' "$PUT_RESP" | sed '$d')
HTTP_CODE=$(printf '%s' "$PUT_RESP" | tail -n1)

if [[ "$HTTP_CODE" == "204" || "$HTTP_CODE" == "200" ]]; then
  echo "User '${USERNAME}' (id: ${USER_ID}) updated successfully → enabled=${ENABLE_VALUE}"
else
  echo "Failed to update user. HTTP $HTTP_CODE"
  echo "Response body:"
  echo "$HTTP_BODY"
  exit 5
fi
