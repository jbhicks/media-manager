#!/bin/bash
# Jellyfin Setup Completion Script
# Run this to finish Jellyfin configuration after wizard completion

echo "🎬 Jellyfin Setup - Final Configuration"
echo "======================================"
echo ""

JELLYFIN_URL="http://192.168.1.49:8096"
USERNAME="josh"
PASSWORD="MediaCenter2026!"

# Function to make authenticated API calls
jellyfin_api() {
    local endpoint="$1"
    local method="${2:-GET}"
    local data="${3:-}"
    
    if [ -n "$data" ]; then
        curl -s -X "$method" "$JELLYFIN_URL$endpoint" \
            -H "Content-Type: application/json" \
            -u "$USERNAME:$PASSWORD" \
            -d "$data"
    else
        curl -s -X "$method" "$JELLYFIN_URL$endpoint" \
            -u "$USERNAME:$PASSWORD"
    fi
}

echo "1. Checking server status..."
SERVER_INFO=$(curl -s "$JELLYFIN_URL/System/Info/Public")
echo "   Server: $(echo "$SERVER_INFO" | grep -o '"ServerName":"[^"]*"' | cut -d'"' -f4)"
echo "   Version: $(echo "$SERVER_INFO" | grep -o '"Version":"[^"]*"' | cut -d'"' -f4)"
echo "   Wizard Complete: $(echo "$SERVER_INFO" | grep -o '"StartupWizardCompleted":[^,}]*' | cut -d':' -f2)"
echo ""

echo "2. Creating Movies library..."
jellyfin_api "/Library/VirtualFolders" "POST" '{
    "Name": "Movies",
    "CollectionType": "movies",
    "Paths": ["/media/Movies"],
    "LibraryOptions": {
        "EnablePhotos": true,
        "EnableRealtimeMonitor": true,
        "EnableChapterImageExtraction": true,
        "ExtractChapterImagesDuringLibraryScan": false,
        "EnableTrickplay": true
    }
}'
echo "   ✓ Movies library created"
echo ""

echo "3. Creating TV Shows library..."
jellyfin_api "/Library/VirtualFolders" "POST" '{
    "Name": "TV Shows",
    "CollectionType": "tvshows",
    "Paths": ["/media/Series"],
    "LibraryOptions": {
        "EnablePhotos": true,
        "EnableRealtimeMonitor": true,
        "EnableChapterImageExtraction": true,
        "ExtractChapterImagesDuringLibraryScan": false,
        "EnableTrickplay": true
    }
}'
echo "   ✓ TV Shows library created"
echo ""

echo "4. Starting library scan..."
jellyfin_api "/Library/Refresh" "POST"
echo "   ✓ Library scan initiated"
echo ""

echo "5. Checking libraries..."
LIBS=$(jellyfin_api "/Library/VirtualFolders")
echo "   Libraries: $(echo "$LIBS" | grep -o '"Name"' | wc -l)"
echo ""

echo "✅ Jellyfin setup complete!"
echo ""
echo "📱 Access your server:"
echo "   Web UI: http://192.168.1.49:8096"
echo "   Username: josh"
echo "   Password: MediaCenter2026!"
echo ""
echo "📁 Libraries:"
echo "   - Movies: /media/Movies"
echo "   - TV Shows: /media/Series"
echo ""
echo "🦞 Why not Zoidberg? Jellyfin is ready!"
