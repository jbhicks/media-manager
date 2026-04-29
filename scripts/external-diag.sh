#!/bin/bash
# Complete network diagnostics for external access

echo "=== External Access Diagnostics ==="
echo ""

echo "1. Host IP addresses:"
ip addr show | grep "inet " | grep -v "127.0.0.1" | awk '{print $2}'

echo ""
echo "2. Checking if port 8080 is listening on all interfaces:"
ss -tlnp | grep 8080 || netstat -tlnp | grep 8080

echo ""
echo "3. Host firewall - INPUT chain:"
iptables -L INPUT -n -v | grep -E "8080| policy"

echo ""
echo "4. Host firewall - NAT PREROUTING:"
iptables -t nat -L PREROUTING -n -v | grep 8080

echo ""
echo "5. Docker port mappings:"
docker port media-manager

echo ""
echo "6. Container IP:"
docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' media-manager

echo ""
echo "7. Testing from host to container (external IP):"
timeout 3 curl -s http://192.168.1.49:8080/api/stats && echo "WORKS" || echo "TIMEOUT"

echo ""
echo "8. Testing from host to container (localhost):"
timeout 3 curl -s http://127.0.0.1:8080/api/stats && echo "WORKS" || echo "TIMEOUT"

echo ""
echo "9. Testing from host to container (direct IP):"
timeout 3 curl -s http://10.0.1.2:8080/api/stats && echo "WORKS" || echo "TIMEOUT"

echo ""
echo "10. Container iptables (filter):"
docker exec media-manager iptables-nft -L INPUT -n -v | head -5

echo ""
echo "11. Container iptables (mangle):"
docker exec media-manager iptables-nft -t mangle -L PREROUTING -n -v | head -5

echo ""
echo "=== End Diagnostics ==="
