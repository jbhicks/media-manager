#!/bin/bash
# Check actual nft ruleset (not iptables-nft translation)

echo "=== Actual nft Ruleset ==="
docker exec media-manager nft list ruleset

echo ""
echo "=== End ==="
