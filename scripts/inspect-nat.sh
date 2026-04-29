#!/bin/bash
# Deep NAT table inspection

echo "=== NAT Table Deep Inspection ==="

echo ""
echo "--- Host PREROUTING chain ---"
iptables -t nat -L PREROUTING -n -v

echo ""
echo "--- Host POSTROUTING chain ---"
iptables -t nat -L POSTROUTING -n -v

echo ""
echo "--- Host DOCKER chain (nat) ---"
iptables -t nat -L DOCKER -n -v

echo ""
echo "--- Container NAT table ---"
docker exec media-manager iptables -t nat -L -n -v

echo ""
echo "--- Test with tcpdump ---"
timeout 3 tcpdump -i br-0d7be20b1d70 -n port 8080 2>&1 &
TCPDUMP_PID=$!
sleep 1
curl -s --max-time 2 http://localhost:8080/ > /dev/null 2>&1
sleep 1
kill $TCPDUMP_PID 2>/dev/null

echo ""
echo "--- Check if Docker daemon is managing iptables ---"
docker info 2>/dev/null | grep -i "iptables"

echo ""
echo "=== End Inspection ==="
