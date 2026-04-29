#!/bin/bash
# Fix mangle rules to accept all Docker bridge traffic

echo "=== Fixing Mangle Rules (All Docker Bridge Traffic) ==="

echo "Removing old rule..."
docker exec media-manager iptables-nft -t mangle -D PREROUTING -i eth0 -s 10.0.1.0/24 -j ACCEPT 2>/dev/null || true
docker exec media-manager iptables-nft -t mangle -D POSTROUTING -o eth0 -d 10.0.1.0/24 -j ACCEPT 2>/dev/null || true

echo "Adding blanket accept for eth0..."
docker exec media-manager iptables-nft -t mangle -I PREROUTING 1 -i eth0 -j ACCEPT
docker exec media-manager iptables-nft -t mangle -I POSTROUTING 1 -o eth0 -j ACCEPT

echo ""
echo "Updated mangle rules:"
docker exec media-manager iptables-nft -t mangle -L -n -v

echo ""
echo "Testing connectivity..."
timeout 3 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works!" || echo "Still failing..."

timeout 3 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works!" || echo "Still failing..."

echo ""
echo "=== Fix Complete ==="
