#!/bin/bash
# Diagnose network connectivity to media-manager container

echo "=== Media Manager Network Diagnostics ==="
echo ""

echo "--- Host IP Addresses ---"
ip addr show | grep "inet " | grep -v "127.0.0.1"
echo ""

echo "--- Docker Port Mappings ---"
docker port media-manager
echo ""

echo "--- Host Firewall (INPUT chain) ---"
iptables -L INPUT -n -v | grep -E "8080|ACCEPT|DROP" | head -10
echo ""

echo "--- Host Firewall (DOCKER chain) ---"
iptables -t nat -L DOCKER -n -v | grep 8080
echo ""

echo "--- Container Firewall ---"
docker exec media-manager iptables -L INPUT -n -v | grep -E "8080|ACCEPT|DROP" | head -10
echo ""

echo "--- Container Processes ---"
docker exec media-manager ps aux | grep -E "media-manager|nordvpn"
echo ""

echo "--- Container Network Interfaces ---"
docker exec media-manager ip addr show
echo ""

echo "--- Test from Host to Container (localhost) ---"
timeout 5 curl -s http://localhost:8080/api/stats || echo "TIMEOUT or FAILED"
echo ""

echo "--- Test from Host to Container (host IP) ---"
HOST_IP=$(ip route get 1 | awk '{print $7; exit}')
echo "Trying host IP: $HOST_IP"
timeout 5 curl -s http://$HOST_IP:8080/api/stats || echo "TIMEOUT or FAILED"
echo ""

echo "--- Docker Network Info ---"
docker network inspect media-manager_default | grep -A 5 "Containers"
echo ""

echo "=== End Diagnostics ==="
