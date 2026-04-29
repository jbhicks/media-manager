#!/bin/bash
# Check container ARP settings

echo "=== Container ARP Diagnostics ==="

echo ""
echo "--- Container ARP settings ---"
docker exec media-manager bash -c "
for f in /proc/sys/net/ipv4/conf/eth0/*arp*; do
  echo \"\$f: \$(cat \$f)\" 2>/dev/null
done
"

echo ""
echo "--- Container interface stats ---"
docker exec media-manager cat /proc/net/dev | grep eth0

echo ""
echo "--- Host bridge ARP table ---"
ip neigh show dev br-0d7be20b1d70

echo ""
echo "--- Host bridge fdb ---"
bridge fdb show br br-0d7be20b1d70 | grep veth27efca7

echo ""
echo "--- Test direct layer 2 with arping ---"
which arping >/dev/null 2>&1 && sudo arping -c 3 -I br-0d7be20b1d70 10.0.1.2 || echo "arping not available"

echo ""
echo "=== End Diagnostics ==="
