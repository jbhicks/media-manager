#!/bin/bash
# Fix media-manager container firewall to allow Docker bridge traffic

echo "=== Fixing Container Firewall ==="

# Check current firewall state
echo "Current INPUT policy:"
docker exec media-manager iptables -L INPUT -n | head -3

# Add rule to accept all traffic from Docker bridge (eth0 interface)
echo "Adding Docker bridge acceptance rule..."
docker exec media-manager iptables -I INPUT 1 -i eth0 -j ACCEPT

# Verify the rule was added
echo "Updated INPUT chain:"
docker exec media-manager iptables -L INPUT -n -v | head -10

# Test connectivity from host
echo ""
echo "Testing connectivity..."
timeout 5 curl -s http://localhost:8080/api/stats && echo "SUCCESS: localhost:8080 works" || echo "TIMEOUT: localhost:8080 still blocked"

timeout 5 curl -s http://192.168.1.49:8080/api/stats && echo "SUCCESS: 192.168.1.49:8080 works" || echo "TIMEOUT: 192.168.1.49:8080 still blocked"

echo ""
echo "=== Fix Complete ==="
echo ""
echo "NOTE: This fix is temporary. To make it permanent, you need to update"
echo "the entrypoint.sh script to add the Docker bridge rule during startup."
