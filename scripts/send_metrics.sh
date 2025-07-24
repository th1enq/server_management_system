#!/bin/bash

ACCESS_TOKEN=$(cat /data/access_token.txt)

send_metrics() {
  RESPONSE=$(curl -s -X POST http://$HOST_IP:8080/api/v1/servers/monitoring \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $ACCESS_TOKEN")

  echo "$(date): Sent metrics, response: $RESPONSE"
}

echo "Start sending metrics every $INTERVAL_TIME seconds..."

while true; do
  send_metrics
  sleep "$INTERVAL_TIME"
done
