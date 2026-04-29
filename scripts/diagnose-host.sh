#!/bin/bash
# Check host-level Docker bridge connectivity

echo "=== Host Bridge Diagnostics ==="

echo ""
echo "--- Bridge Interface ---"
ip addr show br-0d7be20b1d70

echo ""
echo "--- Bridge Routing ---"
ip route | grep 10.0.1

echo ""
echo "--- Host FORWARD Chain ---"
iptables -L FORWARD -n -v | head -20

echo ""
echo "--- Host Filter Table (DOCKER-USER) ---"
iptables -L DOCKER-USER -n -v 2>/dev/null || echo "DOCKER-USER chain not found"

echo ""
echo "--- IP Forwarding ---"
cat /proc/sys/net/ipv4/ip_forward

echo ""
echo "--- Bridge Interface Status ---"
cat /sys/class/net/br-0d7be20b1d70/operstate

echo ""
echo "=== End Diagnostics ==="
