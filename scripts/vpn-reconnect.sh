#!/bin/bash
# Reconnect containerized NordVPN for media-manager
set -euo pipefail

cd /home/josh/media-manager

if ! docker ps -q -f name=^nordvpn$ | grep -q .; then
    echo "Starting nordvpn container..."
    docker compose up -d nordvpn
    sleep 5
fi

echo "Reconnecting VPN inside container..."
docker exec nordvpn nordvpn disconnect 2>/dev/null || true
sleep 2

if ! docker exec nordvpn nordvpn connect "${NORDVPN_SERVER:-United_States}"; then
    echo "Connect failed, restarting nordvpn container..."
    docker compose restart nordvpn
    sleep 10
    docker exec nordvpn nordvpn connect "${NORDVPN_SERVER:-United_States}" || true
fi

docker exec nordvpn nordvpn status
