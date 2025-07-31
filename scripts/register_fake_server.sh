#!/bin/bash

for i in $(seq 2 11); do
  ID="server-$i"
  NAME="Fake Server $i"
  INTERVAL_TIME=$((RANDOM % 20 + 5))
  echo "Register from fake-server-$i"
  docker exec fake-server-$i ./register.sh "$ID" "$NAME" "$INTERVAL_TIME"
done