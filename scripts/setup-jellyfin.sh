#!/bin/bash

# Jellyfin Setup Script for Media Manager

set -e

echo "==================================="
echo " Jellyfin Setup for Media Manager"
echo "==================================="
echo ""

# Check if Jellyfin is running
if ! docker ps | grep -q jellyfin; then
    echo "❌ Jellyfin container is not running!"
    echo ""
    echo "Starting Jellyfin..."
    cd /home/josh/media-manager
    UID=$(id -u) GID=$(id -g) docker-compose -f docker-compose-jellyfin.yml up -d
    sleep 5
fi

echo "✅ Jellyfin is running!"
echo ""
echo "📍 Access Jellyfin at: http://localhost:8096"
echo ""
echo "==================================="
echo " Initial Setup Instructions"
echo "==================================="
echo ""
echo "1. Open http://localhost:8096 in your browser"
echo "2. Follow the setup wizard:"
echo "   - Create an admin account"
echo "   - Add a Movies library pointing to: /media/movies"
echo "   - Skip other optional steps"
echo ""
echo "3. After setup, create an API key:"
echo "   - Go to Dashboard → API Keys"
echo "   - Click 'New API Key'"
echo "   - Name it 'Media Manager'"
echo "   - Copy the API key"
echo ""
echo "4. Update the database:"
echo "   sqlite3 ~/.media-manager/media.db"
echo "   UPDATE service_configs SET jellyfin_url='http://localhost:8096', jellyfin_api_key='YOUR_API_KEY';"
echo "   .quit"
echo ""
echo "5. Restart the media manager service"
echo ""
echo "==================================="
echo ""

# Check Jellyfin health
echo "Checking Jellyfin status..."
sleep 2

if curl -s http://localhost:8096/health > /dev/null 2>&1; then
    echo "✅ Jellyfin is healthy and ready!"
else
    echo "⏳ Jellyfin is starting up... (this may take 30-60 seconds)"
    echo "   Visit http://localhost:8096 when ready"
fi

echo ""
echo "Done! Follow the instructions above to complete setup."
