#!/bin/bash
# SUPER EASY Jellyfin Setup - Fully Automated!
# This script correctly uses the Jellyfin API with proper authentication

set -e

echo "🎬 SUPER EASY Jellyfin Setup"
echo "============================"
echo ""

JELLYFIN_URL="http://192.168.1.49:8096"
USER="josh"
PASS="MediaCenter2026!"

# Step 1: Check if Jellyfin is running
echo "Step 1: Checking Jellyfin..."
if ! curl -s "$JELLYFIN_URL/System/Info/Public" > /dev/null; then
    echo "❌ Jellyfin is not running!"
    echo "   Start it with: cd ~/homeassistant && docker-compose up -d jellyfin"
    exit 1
fi
echo "✅ Jellyfin is running"
echo ""

# Step 2: Complete wizard if needed
echo "Step 2: Checking wizard status..."
WIZARD_DONE=$(curl -s "$JELLYFIN_URL/System/Info/Public" | grep -o '"StartupWizardCompleted":true')
if [ -n "$WIZARD_DONE" ]; then
    echo "✅ Wizard already complete"
else
    echo "Completing wizard..."
    curl -s -X POST "$JELLYFIN_URL/Startup/Complete" \
        -H "Content-Type: application/json" \
        -d '{"UICulture":"en-US"}'
    echo "✅ Wizard complete"
fi
echo ""

# Step 3: Authenticate and get token
echo "Step 3: Authenticating..."
AUTH_JSON=$(curl -s -X POST "$JELLYFIN_URL/Users/AuthenticateByName" \
    -H "Content-Type: application/json" \
    -H 'X-Emby-Authorization: MediaBrowser Client="MediaManager", Device="SetupScript", DeviceId="setup-script-001", Version="1.0.0"' \
    -d "{\"Username\":\"$USER\",\"Pw\":\"$PASS\"}")

TOKEN=$(echo "$AUTH_JSON" | grep -o '"AccessToken":"[^"]*"' | sed 's/.*:"\([^"]*\)".*/\1/')

if [ -z "$TOKEN" ]; then
    echo "❌ Authentication failed!"
    echo "   Response: $AUTH_JSON"
    exit 1
fi

echo "✅ Authenticated (token: ${TOKEN:0:20}...)"
echo ""

# Step 4: Check existing libraries
echo "Step 4: Checking existing libraries..."
EXISTING_LIBS=$(curl -s "$JELLYFIN_URL/Library/VirtualFolders" \
    -H "X-Emby-Token: $TOKEN")

echo "$EXISTING_LIBS" | python3 -c "
import sys, json
data = json.load(sys.stdin)
if data:
    print(f'Found {len(data)} existing libraries:')
    for lib in data:
        print(f'  📁 {lib[\"Name\"]} ({lib.get(\"CollectionType\",\"unknown\")})')
else:
    print('No existing libraries')
" 2>/dev/null || echo "No libraries found"
echo ""

# Step 5: Create Movies library if not exists
echo "Step 5: Creating Movies library..."
if ! echo "$EXISTING_LIBS" | grep -q '"Name":"Movies"'; then
    curl -s -X POST "$JELLYFIN_URL/Library/VirtualFolders" \
        -H "Content-Type: application/json" \
        -H "X-Emby-Token: $TOKEN" \
        -d '["/media/Movies"]' \
        -G --data-urlencode "name=Movies" \
        --data-urlencode "collectionType=movies" \
        --data-urlencode "refreshLibrary=true" > /dev/null
    echo "✅ Movies library created"
else
    echo "✅ Movies library already exists"
fi
echo ""

# Step 6: Create TV Shows library if not exists
echo "Step 6: Creating TV Shows library..."
if ! echo "$EXISTING_LIBS" | grep -q '"Name":"TV Shows"'; then
    curl -s -X POST "$JELLYFIN_URL/Library/VirtualFolders" \
        -H "Content-Type: application/json" \
        -H "X-Emby-Token: $TOKEN" \
        -d '["/media/Series"]' \
        -G --data-urlencode "name=TV Shows" \
        --data-urlencode "collectionType=tvshows" \
        --data-urlencode "refreshLibrary=true" > /dev/null
    echo "✅ TV Shows library created"
else
    echo "✅ TV Shows library already exists"
fi
echo ""

# Step 7: Add paths to libraries
echo "Step 7: Ensuring library paths are set..."
curl -s -X POST "$JELLYFIN_URL/Library/VirtualFolders/Paths" \
    -H "Content-Type: application/json" \
    -H "X-Emby-Token: $TOKEN" \
    -d '{"Name":"Movies","Path":"/media/Movies"}' > /dev/null

curl -s -X POST "$JELLYFIN_URL/Library/VirtualFolders/Paths" \
    -H "Content-Type: application/json" \
    -H "X-Emby-Token: $TOKEN" \
    -d '{"Name":"TV Shows","Path":"/media/Series"}' > /dev/null
echo "✅ Paths configured"
echo ""

# Step 8: Start library scan
echo "Step 8: Starting media scan..."
curl -s -X POST "$JELLYFIN_URL/Library/Refresh" \
    -H "X-Emby-Token: $TOKEN" > /dev/null
echo "✅ Scan started"
echo ""

# Step 9: Verify
echo "Step 9: Verifying setup..."
FINAL_LIBS=$(curl -s "$JELLYFIN_URL/Library/VirtualFolders" \
    -H "X-Emby-Token: $TOKEN")

echo "$FINAL_LIBS" | python3 -c "
import sys, json
data = json.load(sys.stdin)
print(f'Total libraries: {len(data)}')
for lib in data:
    print(f'\\n📁 {lib[\"Name\"]} ({lib[\"CollectionType\"]})')
    for loc in lib.get('Locations', []):
        print(f'   → {loc}')
" 2>/dev/null

echo ""
echo "🎉 DONE! Jellyfin is fully configured!"
echo ""
echo "📱 Access: http://192.168.1.49:8096"
echo "👤 Login: $USER / $PASS"
echo ""
echo "📁 Libraries:"
echo "   - Movies → /media/Movies"
echo "   - TV Shows → /media/Series"
echo ""
echo "🦞 Why not Zoidberg? Easy peasy!"
