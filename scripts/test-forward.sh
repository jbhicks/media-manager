#!/bin/bash
# Fix host FORWARD chain for Docker bridge traffic

echo "=== Fixing Host FORWARD Chain ==="

echo "Current FORWARD chain:"
iptables -L FORWARD -n -v | head -10

echo ""
echo "DOCKER-FORWARD chain:"
iptables -L DOCKER-FORWARD -n -v | head -20

echo ""
echo "Testing if FORWARD policy is blocking..."

# Temporarily set FORWARD to ACCEPT to test
iptables -P FORWARD ACCEPT
echo "FORWARD policy set to ACCEPT (temporary)"

echo ""
echo "Testing connectivity..."
timeout 3 curl -s http://localhost:8080/api/stats && echo "SUCCESS!" || echo "Still failing..."

echo ""
echo "=== Fix Information ==="
echo "If this works, the issue is the FORWARD chain policy."
echo "To make permanent, you need to add a rule like:"
echo "  iptables -A DOCKER-USER -i br-0d7be20b1d70 -j ACCEPT"
echo "  iptables -A DOCKER-USER -o br-0d7be20b1d70 -j ACCEPT"
echo ""
echo "Or permanently:"
echo "  iptables -P FORWARD ACCEPT"
