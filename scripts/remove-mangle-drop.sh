#!/bin/bash
# Remove NordVPN DROP rules from mangle to test connectivity

echo "=== Removing NordVPN Mangle DROP Rules ==="

echo "Current PREROUTING:"
docker exec media-manager iptables-nft -t mangle -L PREROUTING --line-numbers -n

echo ""
echo "Deleting DROP rule from PREROUTING (line 3)..."
docker exec media-manager iptables-nft -t mangle -D PREROUTING 3

echo ""
echo "Current POSTROUTING:"
docker exec media-manager iptables-nft -t mangle -L POSTROUTING --line-numbers -n

echo ""
echo "Deleting DROP rule from POSTROUTING..."
# Find the line number of the DROP rule
DROP_LINE=$(docker exec media-manager iptables-nft -t mangle -L POSTROUTING --line-numbers -n | grep DROP | awk '{print $1}')
if [ -n "$DROP_LINE" ]; then
    docker exec media-manager iptables-nft -t mangle -D POSTROUTING $DROP_LINE
    echo "Deleted line $DROP_LINE"
else
    echo "No DROP rule found"
fi

echo ""
echo "Testing connectivity..."
timeout 3 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works!" || echo "Still failing..."

timeout 3 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works!" || echo "Still failing..."

echo ""
echo "=== Done ==="
