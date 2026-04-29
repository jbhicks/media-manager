#!/bin/bash
# Fix NordVPN mangle rules using iptables-nft

echo "=== Fixing NordVPN Mangle Rules (Alternative Method) ==="

echo "Checking current mangle rules..."
docker exec media-manager iptables-nft -t mangle -L -n -v

echo ""
echo "Adding Docker bridge exceptions using iptables-nft..."

# Add rule to accept incoming Docker traffic before the drop
docker exec media-manager iptables-nft -t mangle -I PREROUTING 1 \
    -i eth0 -s 10.0.1.0/24 -j ACCEPT

# Add rule to accept outgoing Docker traffic  
docker exec media-manager iptables-nft -t mangle -I POSTROUTING 1 \
    -o eth0 -d 10.0.1.0/24 -j ACCEPT

echo ""
echo "Updated mangle rules:"
docker exec media-manager iptables-nft -t mangle -L -n -v

echo ""
echo "Testing connectivity..."
timeout 3 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works!" || echo "Still failing..."

timeout 3 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works!" || echo "Still failing..."

echo ""
echo "=== Fix Complete ==="
