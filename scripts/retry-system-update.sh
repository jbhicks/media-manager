#!/bin/bash
# Retry the failed Omarchy system update
set -euo pipefail

echo "==> Syncing package databases..."
pacman -Sy --noconfirm

echo "==> Running full system upgrade..."
pacman -Syu --noconfirm

echo "==> Verifying media-manager stack after update..."
cd /home/josh/media-manager
docker compose up -d nordvpn media-manager proxy jellyfin

echo "==> Post-update status"
docker ps --format "table {{.Names}}\t{{.Status}}" | head -12
echo ""
pacman -Q docker containerd nordvpn 2>/dev/null || true
