#!/bin/bash
# SUPER EASY Jellyfin Setup - Just 3 steps!
set -e

echo "🎬 SUPER EASY Jellyfin Setup"
echo "============================"
echo ""

JELLYFIN_URL="http://192.168.1.49:8096"
USER="josh"
PASS="MediaCenter2026!"

# Get authentication token first
echo "Step 1: Authenticating..."
AUTH_JSON=$(curl -s -X POST "$JELLYFIN_URL/Users/AuthenticateByName" \
    -H "Content-Type: application/json" \
    -d "{\"Username\":\"$USER\",\"Pw\":\"$PASS\"}")

# Extract token using grep/sed (no python needed)
TOKEN=$(echo "$AUTH_JSON" | grep -o '"AccessToken":"[^"]*"' | sed 's/.*:"\([^"]*\)".*/\1/')

if [ -z "$TOKEN" ]; then
    echo "❌ Could not authenticate. Is the user created?"
    echo "   Try: curl -X POST $JELLYFIN_URL/Users/New -H 'Content-Type: application/json' -d '{\"Name\":\"josh\",\"Password\":\"MediaCenter2026!\"}'"
    exit 1
fi

echo "✅ Got auth token"
echo ""

# Create Movies library
echo "Step 2: Creating Movies library..."
curl -s -X POST "$JELLYFIN_URL/Library/VirtualFolders" \
    -H "Content-Type: application/json" \
    -H "X-Emby-Token: $TOKEN" \
    -d '{
        "Name": "Movies",
        "CollectionType": "movies",
        "Paths": ["/media/Movies"],
        "LibraryOptions": {
            "EnablePhotos": true,
            "EnableRealtimeMonitor": true,
            "PreferredMetadataLanguage": "en"
        }
    }'
echo "✅ Movies library created"
echo ""

# Create TV Shows library
echo "Step 3: Creating TV Shows library..."
curl -s -X POST "$JELLYFIN_URL/Library/VirtualFolders" \
    -H "Content-Type: application/json" \
    -H "X-Emby-Token: $TOKEN" \
    -d '{
        "Name": "TV Shows",
        "CollectionType": "tvshows",
        "Paths": ["/media/Series"],
        "LibraryOptions": {
            "EnablePhotos": true,
            "EnableRealtimeMonitor": true,
            "PreferredMetadataLanguage": "en"
        }
    }'
echo "✅ TV Shows library created"
echo ""

# Start scan
echo "Step 4: Starting media scan..."
curl -s -X POST "$JELLYFIN_URL/Library/Refresh" \
    -H "X-Emby-Token: $TOKEN"
echo "✅ Scan started"
echo ""

# Verify
echo "Step 5: Verifying..."
LIBS=$(curl -s "$JELLYFIN_URL/Library/VirtualFolders" \
    -H "X-Emby-Token: $TOKEN")

echo "Raw response: $LIBS"
echo ""

echo "🎉 DONE! Jellyfin is ready!"
echo ""
echo "📱 Open: http://192.168.1.49:8096"
echo "👤 Login: josh / MediaCenter2026!"
echo ""
echo "🦞 Why not Zoidberg? Easy peasy!"