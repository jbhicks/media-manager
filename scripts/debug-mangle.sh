#!/bin/bash
# Debug mangle rules with conntrack and packet tracing

echo "=== Deep Packet Inspection ==="

echo "--- Check which iptables backend is active ---"
docker exec media-manager update-alternatives --display iptables 2>/dev/null

echo ""
echo "--- Check legacy mangle table ---"
docker exec media-manager iptables-legacy -t mangle -L -n -v

echo ""
echo "--- Connection tracking state ---"
docker exec media-manager bash -c "
# Check if conntrack is available
if command -v conntrack > /dev/null 2>&1; then
    conntrack -L 2>/dev/null | grep ':8080' | head -5
else
    echo 'conntrack not available'
fi
"

echo ""
echo "--- Try with explicit mark ---"
# Try adding rule with mark match to see if that's what NordVPN expects
docker exec media-manager iptables-nft -t mangle -I PREROUTING 1 -i eth0 -m mark --mark 0xe1f1 -j ACCEPT 2>/dev/null || echo "mark rule failed"

echo ""
echo "--- Test with conntrack marks ---"
docker exec media-manager bash -c "
# List nft ruleset to see actual backend
nft list ruleset 2>/dev/null | grep -A5 'chain PREROUTING' | head -20
"

echo ""
echo "=== End Inspection ==="
