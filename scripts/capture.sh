#!/bin/bash
# Capture packets on Docker bridge to diagnose traffic flow

echo "=== Packet Capture Test ==="

echo "Starting tcpdump on bridge..."
timeout 5 tcpdump -i br-0d7be20b1d70 -n host 10.0.1.2 and port 8080 -c 10 2>&1 &
TCPDUMP_PID=$!

sleep 1

echo "Sending test request..."
timeout 2 curl -s http://10.0.1.2:8080/api/stats > /dev/null 2>&1 || true

sleep 3
kill $TCPDUMP_PID 2>/dev/null

echo ""
echo "=== Capture Complete ==="
