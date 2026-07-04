#!/bin/bash
# Finish containerized NordVPN migration for media-manager
# Run: sudo bash ~/finish-vpn-migration.sh
set -euo pipefail

echo "==> Restarting containerd + docker (fix version mismatch)"
systemctl restart containerd
sleep 2
systemctl restart docker
sleep 3

if ! docker info >/dev/null 2>&1; then
    echo "ERROR: Docker failed to start"
    systemctl status docker --no-pager
    exit 1
fi
echo "    Docker OK"

echo "==> Disabling native NordVPN daemon (VPN lives in container now)"
systemctl disable --now nordvpnd 2>/dev/null || true

echo "==> Configuring host DNS via Pi-hole"
systemctl unmask systemd-resolved 2>/dev/null || true
grep -q "^DNS=127.0.0.1" /etc/systemd/resolved.conf || \
    sed -i "s/^#DNS=.*/DNS=127.0.0.1/" /etc/systemd/resolved.conf
systemctl enable --now systemd-resolved
systemctl restart systemd-resolved

cd /home/josh/pihole
docker compose up -d

echo "==> Waiting for Pi-hole DNS..."
for i in $(seq 1 15); do
    if dig @127.0.0.1 google.com +short +time=1 +tries=1 2>/dev/null | grep -q .; then
        echo "    Pi-hole DNS ready"
        break
    fi
    sleep 2
done

echo "==> Starting media-manager VPN stack"
cd /home/josh/media-manager
docker rm -f 124fbf3419f0_media-manager 2>/dev/null || true
docker compose up -d nordvpn media-manager proxy jellyfin

echo "==> Waiting for container VPN..."
for i in $(seq 1 30); do
    if docker exec nordvpn nordvpn status 2>/dev/null | grep -q "Status: Connected"; then
        echo "    Container VPN connected"
        break
    fi
    sleep 3
done

echo ""
echo "=== Verification ==="
echo "--- Native VPN (should be disabled) ---"
systemctl is-enabled nordvpnd 2>/dev/null || echo "nordvpnd: disabled"

echo "--- Container VPN ---"
docker exec nordvpn nordvpn status 2>/dev/null | head -8 || echo "Container VPN not ready"

echo "--- DNS ---"
resolvectl status 2>/dev/null | head -15 || cat /etc/resolv.conf
dig @127.0.0.1 google.com +short 2>/dev/null | head -2

echo "--- Public IPs (host vs container should differ) ---"
HOST_IP=$(curl -s --max-time 10 https://ifconfig.me 2>/dev/null || echo unknown)
CONTAINER_IP=$(docker exec nordvpn curl -s --max-time 10 https://ifconfig.me 2>/dev/null || echo unknown)
echo "Host IP:      $HOST_IP  (home ISP)"
echo "Container IP: $CONTAINER_IP  (VPN exit)"

echo "--- Running containers ---"
docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"

echo ""
echo "Done. Architecture:"
echo "  - Host: no VPN, Pi-hole DNS, Jellyfin streams locally"
echo "  - Container: nordvpn sidecar routes media-manager downloads through VPN"
