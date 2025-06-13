#!/bin/bash

SUBNET="192.168.100"
IMAGE="nginx"
NETWORK="custom-net"

docker network create \
  --subnet=192.168.100.0/24 \
  --gateway=192.168.100.1 \
  custom-net

for i in $(seq 2 10); do
  IP="${SUBNET}.${i}"
  NAME="fake-server-${i}"

  docker run -d \
    --name "$NAME" \
    --network "$NETWORK" \
    --ip "$IP" \
    "$IMAGE"
done
