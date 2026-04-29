#!/bin/bash
# Complete iptables dump from container

echo "=== Complete Container Firewall Dump ==="

echo ""
echo "--- Filter Table (Full) ---"
docker exec media-manager iptables-nft -L -n -v --line-numbers

echo ""
echo "--- Mangle Table (Full) ---"
docker exec media-manager iptables-nft -t mangle -L -n -v --line-numbers

echo ""
echo "--- Raw Table ---"
docker exec media-manager iptables-nft -t raw -L -n -v 2>/dev/null || echo "No raw table"

echo ""
echo "--- Security Table ---"
docker exec media-manager iptables-nft -t security -L -n -v 2>/dev/null || echo "No security table"

echo ""
echo "--- Container IP Info ---"
docker exec media-manager ip addr show

echo ""
echo "--- Container Routes ---"
docker exec media-manager ip route

echo ""
echo "=== End Dump ==="
