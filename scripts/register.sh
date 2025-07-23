#!/bin/bash

SERVER_ID="$1"
SERVER_NAME="$2"
DESCRIPTION="This is $SERVER_NAME"
LOCATION="Data Center A"
OS="Linux"

RESPONSE=$(curl -s -X POST http://$HOST_IP:8080/api/v1/servers/register \
  -H "Content-Type: application/json" \
  -d "{
    \"server_id\": \"$SERVER_ID\",
    \"server_name\": \"$SERVER_NAME\",
    \"description\": \"$DESCRIPTION\",
    \"location\": \"$LOCATION\",
    \"os\": \"$OS\"
  }")

echo "Register response: $RESPONSE"

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.data.access_token')

if [ "$ACCESS_TOKEN" != "null" ]; then
  echo "$ACCESS_TOKEN" > /data/token.txt
  echo "Token saved to /data/token.txt"
else
  echo "Failed to get token"
fi
