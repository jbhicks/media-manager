#!/bin/bash
# Fix container OUTPUT firewall to allow Docker bridge responses

echo "=== Fixing Container OUTPUT Firewall ==="

echo "Current OUTPUT chain:"
docker exec media-manager iptables -L OUTPUT -n -v | head -15

echo ""
echo "Adding eth0 to OUTPUT chain..."
docker exec media-manager iptables -I OUTPUT 1 -o eth0 -j ACCEPT

echo ""
echo "Updated OUTPUT chain:"
docker exec media-manager iptables -L OUTPUT -n -v | head -15

echo ""
echo "Testing connectivity..."
timeout 5 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works" || echo "TIMEOUT: localhost:8080 still blocked"

timeout 5 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works" || echo "TIMEOUT: 192.168.1.49:8080 still blocked"

echo ""
echo "=== Fix Complete ==="
