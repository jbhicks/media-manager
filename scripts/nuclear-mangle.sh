#!/bin/bash
# Nuclear option: completely disable NordVPN mangle rules

echo "=== Disabling All NordVPN Mangle Rules ==="

echo "Current nft mangle rules:"
docker exec media-manager nft list table ip mangle 2>&1

echo ""
echo "Flushing mangle chains..."
docker exec media-manager nft flush chain ip mangle PREROUTING 2>/dev/null || true
docker exec media-manager nft flush chain ip mangle POSTROUTING 2>/dev/null || true

echo ""
echo "Setting mangle policies to ACCEPT..."
docker exec media-manager nft 'add chain ip mangle PREROUTING { type filter hook prerouting priority mangle; policy accept; }' 2>/dev/null || true
docker exec media-manager nft 'add chain ip mangle POSTROUTING { type filter hook postrouting priority mangle; policy accept; }' 2>/dev/null || true

echo ""
echo "Verifying mangle is empty:"
docker exec media-manager nft list table ip mangle 2>&1

echo ""
echo "Testing connectivity..."
timeout 3 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works!" || echo "Still failing..."

timeout 3 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works!" || echo "Still failing..."

timeout 3 curl -s http://10.0.1.2:8080/api/stats && echo "SUCCESS: 10.0.1.2:8080 works!" || echo "Still failing..."

echo ""
echo "=== Done ==="
