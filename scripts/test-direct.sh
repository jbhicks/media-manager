#!/bin/bash
# Test direct TCP to container and check all possible blocking points

echo "=== Direct TCP Test to Container ==="

echo "Test 1: Direct to container IP (10.0.1.2)"
curl --max-time 2 -v http://10.0.1.2:8080/api/stats 2>&1 | head -20

echo ""
echo "Test 2: Check if Docker proxy is listening on host"
netstat -tlnp 2>/dev/null | grep 8080 || ss -tlnp | grep 8080

echo ""
echo "Test 3: Test from container to host bridge"
docker exec media-manager curl -s --max-time 2 http://10.0.1.1:8080/api/stats 2>&1 | head -1

echo ""
echo "Test 4: Check all nftables tables"
docker exec media-manager nft list tables

echo ""
echo "Test 5: Check for any other filter chains"
docker exec media-manager iptables-nft -L -n --line-numbers | grep -E "Chain|DROP|REJECT"

echo ""
echo "Test 6: Host output chain for Docker traffic"
iptables -L OUTPUT -n -v | grep -E "10\.0\.1|docker"

echo ""
echo "=== End Tests ==="
