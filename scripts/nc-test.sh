#!/bin/bash
# Test with netcat to bypass application layer

echo "=== Netcat TCP Test ==="

echo "Starting netcat listener in container..."
docker exec -d media-manager bash -c "nc -l -p 9999 -e echo 'HELLO FROM CONTAINER'"
sleep 1

echo "Test 1: Connect to netcat from host"
timeout 3 bash -c "echo 'TEST' | nc 10.0.1.2 9999" 2>&1 || echo "FAILED"

echo ""
echo "Test 2: Check conntrack"
docker exec media-manager bash -c "cat /proc/net/nf_conntrack 2>/dev/null | grep ':9999\|:8080' | head -5" || echo "conntrack not available"

echo ""
echo "Test 3: Check if conntrack module is loaded"
docker exec media-manager bash -c "lsmod 2>/dev/null | grep conntrack || cat /proc/modules | grep conntrack" || echo "Cannot check modules"

echo ""
echo "Test 4: Check SYN packets with ss"
docker exec media-manager ss -tan | grep -E ":8080|:9999"

echo ""
echo "=== Done ==="
