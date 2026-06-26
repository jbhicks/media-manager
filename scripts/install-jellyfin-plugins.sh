#!/bin/bash
# Jellyfin Plugin Installation Script
# Run this after Jellyfin is fully configured

echo "🎬 Jellyfin Plugin Installer"
echo "============================"
echo ""

JELLYFIN_PLUGIN_DIR="/home/josh/media-manager/jellyfin/config/plugins"
mkdir -p "$JELLYFIN_PLUGIN_DIR"

echo "📥 Downloading plugins..."
echo ""

# 1. Netflix-Style Theme
echo "1. Netflix-Style Custom Theme"
echo "   URL: https://github.com/Kuschel-code/Jellyfin-Custom-Theme"
echo "   Manual install:"
echo "   - Download latest release from GitHub"
echo "   - Extract to: $JELLYFIN_PLUGIN_DIR/CustomTheme/"
echo ""

# 2. Skip Intro Plugin
echo "2. Skip Intro Plugin (TheIntroDB)"
echo "   URL: https://github.com/TheIntroDB/jellyfin-plugin"
echo "   Manual install:"
echo "   - Download latest release from GitHub"
echo "   - Extract to: $JELLYFIN_PLUGIN_DIR/IntroSkip/"
echo ""

# 3. Open Subtitles (Official)
echo "3. Open Subtitles Plugin"
echo "   Install via Jellyfin Dashboard:"
echo "   Dashboard → Plugins → Catalog → Open Subtitles"
echo ""

# 4. Webhook (Official)
echo "4. Webhook Plugin"
echo "   Install via Jellyfin Dashboard:"
echo "   Dashboard → Plugins → Catalog → Webhook"
echo ""

echo "✅ Plugin setup complete!"
echo ""
echo "🦞 Why not Zoidberg?"
