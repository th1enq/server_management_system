#!/bin/bash

SUBNET="192.168.100"
IMAGE="fake-server-image"  
NETWORK="custom-net"
START=2
COUNT=${1:-10}
HOST_IP=$(ip route get 1 | awk '{print $(NF-2); exit}')

docker network inspect "$NETWORK" >/dev/null 2>&1 || \
docker network create \
  --subnet=${SUBNET}.0/24 \
  --gateway=${SUBNET}.1 \
  "$NETWORK"

for ((i=0; i<COUNT; i++)); do
  IP_SUFFIX=$((START + i))
  IP="${SUBNET}.${IP_SUFFIX}"
  NAME="fake-server-${IP_SUFFIX}"

  docker rm -f "$NAME" >/dev/null 2>&1

  echo "Create container $NAME with IP $IP"

  docker run -d \
    --name "$NAME" \
    --network "$NETWORK" \
    --ip "$IP" \
    -e HOST_IP="$HOST_IP" \
    -v $(pwd)/register.sh:/app/register.sh \
    -v $(pwd)/send_metrics.sh:/app/send_metrics.sh \
    -v $(pwd)/data/${NAME}:/data \
    "$IMAGE"
done
