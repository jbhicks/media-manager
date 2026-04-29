#!/bin/bash
# Full firewall analysis for media-manager container

echo "=== Full Container Firewall Analysis ==="

echo ""
echo "--- INPUT Chain ---"
docker exec media-manager iptables -L INPUT -n -v

echo ""
echo "--- OUTPUT Chain ---"
docker exec media-manager iptables -L OUTPUT -n -v

echo ""
echo "--- FORWARD Chain ---"
docker exec media-manager iptables -L FORWARD -n -v

echo ""
echo "--- NAT Table ---"
docker exec media-manager iptables -t nat -L -n -v

echo ""
echo "--- Check if Docker proxy is running ---"
ps aux | grep docker-proxy | grep 8080

echo ""
echo "--- Docker network rules on host ---"
sudo iptables -t nat -L DOCKER -n -v | grep 8080
echo ""
sudo iptables -L DOCKER -n -v | head -20

echo ""
echo "--- Test direct connection to container IP ---"
docker exec media-manager curl -s http://10.0.1.2:8080/api/stats | head -1

echo ""
echo "--- Container routing table ---"
docker exec media-manager ip route

echo ""
echo "=== End Analysis ==="
