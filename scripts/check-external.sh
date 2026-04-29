#!/bin/bash
# Check if external access to port 8080 is blocked

echo "=== External Access Check ==="
echo ""
echo "Host firewall INPUT chain:"
sudo iptables -L INPUT -n -v | grep -E "8080|ACCEPT|DROP" | head -10
echo ""
echo "Host firewall FORWARD chain:"
sudo iptables -L FORWARD -n -v | head -10
echo ""
echo "Testing external access from host IP:"
curl -s --max-time 3 http://192.168.1.49:8080/api/stats && echo "EXTERNAL ACCESS WORKS" || echo "EXTERNAL ACCESS BLOCKED"
echo ""
echo "=== End Check ==="
