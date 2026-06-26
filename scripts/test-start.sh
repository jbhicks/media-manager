#!/bin/bash
# Quick start script for testing media-manager backend

cd /home/josh/media-manager

# Kill any existing service
pkill -f "media-manager-service" 2>/dev/null || true
sleep 1

# Start the service with correct API key
export PORT=8084
export DB_PATH="./data/media.db"
export DOWNLOAD_PATH="/mnt/media/Downloads"
export LIBRARY_PATH="/mnt/media/Movies"
export JACKETT_URL="http://192.168.1.49:9117"
export JACKETT_API_KEY="xzvzhznfjsnit25fpa8mloxh66h40svz"
export TMDB_API_KEY="6d4a63549daad44f3abcb460750bb7d1"
export TMDB_READ_ACCESS_TOKEN="eyJhbGciOiJIUzI1NiJ9.eyJhdWQiOiI2ZDRhNjM1NDlkYWFkNDRmM2FiY2I0NjA3NTBiYjdkMSIsIm5iZiI6MTc2NjUxMzY5Ni42NjUsInN1YiI6IjY5NGFkYzIwOTc1ODBhODBhZThlYzM2OSIsInNjb3BlcyI6WyJhcGlfcmVhZCJdLCJ2ZXJzaW9uIjoxfQ.HxGs710OGotl47wHPAv9CC629BsDwyfARS3ibPStFD4"

./tmp/media-manager-service > tmp/service-test.log 2>&1 &
PID=$!
echo "Started media-manager-service with PID: $PID"
echo $PID > tmp/test-service.pid
sleep 5

# Test health endpoint
echo "Testing health endpoint..."
curl -s "http://localhost:8084/api/health" -w "
HTTP_STATUS: %{http_code}
"

# Test trending movies
echo ""
echo "Testing trending movies..."
curl -s "http://localhost:8084/api/discover/movies/trending" | head -c 200
