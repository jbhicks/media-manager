#!/bin/bash
# Fix NordVPN mangle rules to allow Docker bridge traffic

echo "=== Fixing NordVPN Mangle Rules ==="

echo "Current mangle PREROUTING:"
docker exec media-manager nft list chain ip mangle PREROUTING

echo ""
echo "Current mangle POSTROUTING:"
docker exec media-manager nft list chain ip mangle POSTROUTING

echo ""
echo "Adding Docker bridge exceptions to mangle table..."

# Add rule to accept incoming Docker traffic before the drop
docker exec media-manager nft insert rule ip mangle PREROUTING \
    position 1 iifname "eth0" ip saddr 10.0.1.0/24 counter accept

# Add rule to accept outgoing Docker traffic before the drop  
docker exec media-manager nft insert rule ip mangle POSTROUTING \
    position 2 oifname "eth0" ip daddr 10.0.1.0/24 counter accept

echo ""
echo "Updated mangle rules:"
docker exec media-manager nft list chain ip mangle PREROUTING
echo "---"
docker exec media-manager nft list chain ip mangle POSTROUTING

echo ""
echo "Testing connectivity..."
timeout 3 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works!" || echo "Still failing..."

timeout 3 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works!" || echo "Still failing..."

echo ""
echo "=== Fix Complete ==="
