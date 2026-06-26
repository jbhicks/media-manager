#!/bin/bash
# NordVPN wrapper - checks status from inside the media-manager container
# Place this in your PATH and the backend will use it to check VPN status

CONTAINER_ID=$(docker ps -q -f name=media-manager)

if [ -z "$CONTAINER_ID" ]; then
    echo "Status: Disconnected"
    echo "Container not running"
    exit 1
fi

# Run nordvpn command inside the container
docker exec "$CONTAINER_ID" nordvpn "$@" 2>/dev/null