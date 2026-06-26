#!/bin/bash
# Quick start script for testing media-manager backend

cd /home/josh/media-manager

# Kill any existing service
pkill -f "media-manager-service" 2>/dev/null || true
sleep 1

# Start the service
export PORT=8084
export DB_PATH="./data/media.db"
export DOWNLOAD_PATH="/mnt/media/Downloads"
export LIBRARY_PATH="/mnt/media/Movies"
export JACKETT_URL="http://192.168.1.49:9117"
export JACKETT_API_KEY="xzvzhz...0svz"
export TMDB_API_KEY="6d4a63...b7d1"

./tmp/media-manager-service > tmp/service-test.log 2>&1 &
PID=$!
echo "Started media-manager-service with PID: $PID"
echo $PID > tmp/test-service.pid
sleep 5

# Test health endpoint
echo "Testing health endpoint..."
curl -s "http://localhost:8084/api/health" -w "\nHTTP_STATUS: %{http_code}\n"
