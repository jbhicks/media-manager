# Jellyfin Integration Guide

## What is Jellyfin?

Jellyfin is a **free and open-source media server** that lets you stream your downloaded movies to any device (Smart TV, phone, tablet, web browser, etc.). It's similar to Plex but completely free with no premium features hidden behind a paywall.

## Why Jellyfin Instead of Just File Sharing?

| Feature | Samba/File Share | Jellyfin Media Server |
|---------|------------------|----------------------|
| Beautiful interface with posters | ❌ | ✅ |
| Movie metadata (cast, plot, ratings) | ❌ | ✅ |
| Automatic thumbnail generation | ❌ | ✅ |
| Watch history & resume playback | ❌ | ✅ |
| Transcoding for compatibility | ❌ | ✅ |
| Remote access from anywhere | ❌ | ✅ |
| Multi-user support | ❌ | ✅ |
| Smart TV apps | ❌ | ✅ |

## Installation (Already Done!)

Jellyfin is already installed and running in Docker:

```bash
# Check if running
docker ps | grep jellyfin

# View logs
docker logs jellyfin

# Restart
docker restart jellyfin

# Stop
docker stop jellyfin

# Start
cd /home/josh/media-manager
UID=$(id -u) GID=$(id -g) docker-compose -f docker-compose-jellyfin.yml up -d
```

## Initial Setup

### 1. Access Jellyfin Web Interface

Open your browser to: **http://localhost:8096**

### 2. Run the Setup Wizard

On first launch, Jellyfin will show a setup wizard:

1. **Select Language**: English (or your preference)
2. **Create Admin Account**:
   - Username: `admin` (or your choice)
   - Password: (choose a secure password)
3. **Add Media Library**:
   - Click "Add Media Library"
   - Content type: `Movies`
   - Display name: `Movies`
   - Folder: Click the `+` button and select `/media/movies`
   - Click OK
4. **Skip** the following optional steps:
   - Preferred Metadata Language (defaults are fine)
   - Remote Access (can configure later)
5. **Finish** setup

### 3. Create API Key for Media Manager

After setup is complete:

1. Go to **Dashboard** (click the hamburger menu → Admin → Dashboard)
2. Navigate to **API Keys** (under Advanced section)
3. Click **+ New API Key**
4. Enter app name: `Media Manager`
5. Click **OK**
6. **Copy the API key** that appears (you'll need this next)

### 4. Configure Media Manager

Update the database with your API key:

```bash
# Open the database
sqlite3 ~/.media-manager/media.db

# Update the config (replace YOUR_API_KEY_HERE with the actual key)
UPDATE service_configs SET 
  jellyfin_url='http://localhost:8096', 
  jellyfin_api_key='YOUR_API_KEY_HERE';

# Verify it worked
SELECT jellyfin_url, jellyfin_api_key FROM service_configs;

# Exit
.quit
```

### 5. Restart Media Manager Service

```bash
# Kill existing service
lsof -ti:42069 | xargs -r kill -9
lsof -ti:8083 | xargs -r kill -9

# Start with TMDB API key
cd /home/josh/media-manager
TMDB_API_KEY='6d4a63549daad44f3abcb460750bb7d1' \
./tmp/media-manager-service > tmp/service.log 2>&1 &
```

## How It Works

```
[Media Manager Downloads Movie]
          ↓
[Organizes to /mnt/media/Movies/Movie Name (Year)/]
          ↓
[Triggers Jellyfin Library Refresh via API]
          ↓
[Jellyfin Scans Folder]
          ↓
[Jellyfin Fetches Metadata from TMDB/TVDB]
          ↓
[Movie Appears in Jellyfin with Poster, Plot, Cast]
          ↓
[Stream to Any Device!]
```

## Accessing Your Movies

### Web Browser
- Go to http://localhost:8096
- Log in with your admin account
- Browse and play movies

### Smart TV Apps
Jellyfin has apps for:
- Samsung Smart TVs
- LG WebOS
- Android TV / Google TV
- Apple TV
- Roku
- Fire TV

Search for "Jellyfin" in your TV's app store.

### Mobile Apps
- **Android**: [Jellyfin for Android](https://play.google.com/store/apps/details?id=org.jellyfin.mobile)
- **iOS**: [Jellyfin Mobile](https://apps.apple.com/app/jellyfin-mobile/id1480192618)

### Desktop Apps
- **Windows/Mac/Linux**: [Jellyfin Media Player](https://github.com/jellyfin/jellyfin-media-player/releases)

## Remote Access (Optional)

To access your library from outside your home network:

### Option 1: Reverse Proxy (Recommended)
Use Nginx or Caddy with SSL:
```nginx
# Nginx config example
server {
    listen 443 ssl;
    server_name jellyfin.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:8096;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### Option 2: Port Forwarding
Forward port 8096 on your router to your server's local IP.
⚠️ **Security warning**: Use strong passwords if exposing to internet!

## Troubleshooting

### Jellyfin Not Starting
```bash
# Check logs
docker logs jellyfin

# Check if port is in use
lsof -i :8096

# Restart container
docker restart jellyfin
```

### Library Not Refreshing
```bash
# Check media manager logs
tail -f /home/josh/media-manager/tmp/service.log | grep JELLYFIN

# Manually trigger refresh
curl -X POST http://localhost:8096/Library/Refresh \
  -H "X-MediaBrowser-Token: YOUR_API_KEY"
```

### Movies Not Appearing
1. Check file permissions: `ls -la /mnt/media/Movies/`
2. Ensure files are in Plex-compatible structure:
   ```
   /mnt/media/Movies/
     Movie Name (2025)/
       Movie Name (2025).mkv
   ```
3. Manually trigger library scan in Jellyfin Dashboard

### Can't Access from Other Devices
- Ensure devices are on the same network
- Use server's local IP instead of localhost: `http://192.168.X.X:8096`
- Check firewall rules: `sudo ufw allow 8096`

## Configuration Files

- **Docker Compose**: `/home/josh/media-manager/docker-compose-jellyfin.yml`
- **Jellyfin Config**: `~/jellyfin/config/`
- **Jellyfin Cache**: `~/jellyfin/cache/`
- **Media Files**: `/mnt/media/Movies/` (read-only mount)

## Useful Commands

```bash
# View Jellyfin logs in real-time
docker logs -f jellyfin

# Restart Jellyfin
docker restart jellyfin

# Update Jellyfin to latest version
docker-compose -f docker-compose-jellyfin.yml pull
docker-compose -f docker-compose-jellyfin.yml up -d

# Backup Jellyfin config
tar -czf jellyfin-backup-$(date +%Y%m%d).tar.gz ~/jellyfin/config/

# Check Jellyfin system info
curl -s http://localhost:8096/System/Info/Public | jq
```

## Next Steps

1. ✅ Complete initial setup wizard
2. ✅ Create API key
3. ✅ Configure media manager
4. ✅ Download a movie via suggestions
5. ✅ Wait for it to complete and see it appear in Jellyfin!
6. 🎬 Install Jellyfin app on your Smart TV
7. 🍿 Enjoy your movies!

## Resources

- **Jellyfin Website**: https://jellyfin.org
- **Documentation**: https://jellyfin.org/docs
- **Community**: https://forum.jellyfin.org
- **GitHub**: https://github.com/jellyfin/jellyfin

---

**Questions?** Check the logs in `/home/josh/media-manager/tmp/service.log` for any Jellyfin-related errors.
