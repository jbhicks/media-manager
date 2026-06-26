#!/bin/bash
# Final Setup Script for Josh's Ultimate Media Center
# Run this with: sudo bash finish-media-center.sh

echo "🎬 ULTIMATE MEDIA CENTER - FINAL SETUP"
echo "======================================="
echo ""

# Fix permissions
echo "1. Fixing Jellyfin permissions..."
chown -R josh:josh /home/josh/media-manager/jellyfin/config/
chmod -R 755 /home/josh/media-manager/jellyfin/config/

# Restart Jellyfin to apply changes
echo "2. Restarting Jellyfin..."
cd /home/josh/media-manager
docker compose restart jellyfin

echo ""
echo "✅ Setup complete!"
echo ""
echo "📱 NEXT STEPS:"
echo "==============="
echo ""
echo "1. Open Jellyfin Setup Wizard:"
echo "   http://192.168.1.49:8096"
echo ""
echo "2. Create admin account"
echo ""
echo "3. Add Media Libraries:"
echo "   - Movies: /media/Movies"
echo "   - TV Shows: /media/Series"
echo ""
echo "4. Install Plugins (Dashboard → Plugins → Catalog):"
echo "   - Open Subtitles"
echo "   - Webhook"
echo "   - Playback Reporting"
echo ""
echo "5. For Netflix Theme and Skip Intro:"
echo "   - Download from GitHub releases"
echo "   - Extract to jellyfin/config/plugins/"
echo ""
echo "6. Configure Cloudflare Tunnel (optional):"
echo "   - Install cloudflared"
echo "   - Create tunnel: cloudflared tunnel create media-center"
echo "   - Route to http://localhost:8096"
echo ""
echo "7. Create friend accounts:"
echo "   Dashboard → Users → Add User"
echo ""
echo "🦞 Why not Zoidberg? Your media center is ready!"
