#!/bin/bash
# Check host-level firewall affecting Docker proxy responses

echo "=== Host Firewall Check ==="

echo ""
echo "--- Host Filter Table (Full) ---"
iptables -L -n -v --line-numbers

echo ""
echo "--- Host NAT Table (Full) ---"
iptables -t nat -L -n -v --line-numbers

echo ""
echo "--- Check Docker proxy process details ---"
ps aux | grep docker-proxy | grep 8080

echo ""
echo "--- Test connection through Docker proxy ---"
timeout 2 curl -v http://127.0.0.1:8080/api/stats 2>&1 | grep -E "Trying|Connected|HTTP"

echo ""
echo "--- Check if host has any DROP rules for docker0 or bridge interfaces ---"
iptables -L -n -v | grep -E "br-|docker0|DROP"

echo ""
echo "=== End Check ==="
