#!/bin/bash
# NordVPN wrapper - reports status from the containerized VPN sidecar
# Place in ~/.local/bin/nordvpn (login shells prepend ~/.local/bin to PATH)

CONTAINER_ID=$(docker ps -q -f name=^nordvpn$ 2>/dev/null)

if [ -z "$CONTAINER_ID" ]; then
    echo "Status: Disconnected"
    echo "Container not running"
    exit 1
fi

docker exec "$CONTAINER_ID" nordvpn "$@" 2>/dev/null
