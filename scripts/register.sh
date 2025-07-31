#!/bin/bash

SERVER_ID="$1"
SERVER_NAME="$2"
DESCRIPTION="This is $SERVER_NAME"
LOCATION="Data Center A"
OS="Linux"
INTERVAL_TIME="$3"

RESPONSE=$(curl -s -X POST http://$HOST_IP:8080/api/v1/servers/register \
  -H "Content-Type: application/json" \
  -d "{
    \"server_id\": \"$SERVER_ID\",
    \"server_name\": \"$SERVER_NAME\",
    \"description\": \"$DESCRIPTION\",
    \"location\": \"$LOCATION\",
    \"os\": \"$OS\",
    \"interval_time\": $INTERVAL_TIME
  }")

echo "Register response: $RESPONSE"

ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.data.access_token')
if [ "$ACCESS_TOKEN" != "null" ]; then
  echo "$ACCESS_TOKEN" > /data/access_token.txt
  echo "Token saved to /data/access_token.txt"
else
  echo "Failed to get token"
fi

REFRESH_TOKEN=$(echo "$RESPONSE" | jq -r '.data.refresh_token')
if [ "$REFRESH_TOKEN" != "null" ]; then
  echo "$REFRESH_TOKEN" > /data/refresh_token.txt
  echo "Token saved to /data/refresh_token.txt"
else
  echo "Failed to get token"
fi

cat > .env <<EOF
SERVER_ID=$SERVER_ID
SERVER_NAME=$SERVER_NAME
DESCRIPTION=$DESCRIPTION
LOCATION=$LOCATION
OS=$OS
INTERVAL_TIME=$INTERVAL_TIME
EOF