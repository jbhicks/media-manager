# 🌐 Media Manager - Network Access Guide
## Your media center is accessible from any device on your local network!

---

## 📱 NETWORK URLs (Accessible from any device on 192.168.1.x)

### Media Manager (Custom App)
- **Frontend**: http://192.168.1.49:5178
- **Backend API**: http://192.168.1.49:8084
- **Health Check**: http://192.168.1.49:8084/api/health

### Jellyfin Media Server
- **Web UI**: http://192.168.1.49:8096
- **Public Info**: http://192.168.1.49:8096/System/Info/Public

---

## ✅ LIVE-TESTED FEATURES (Verified Working)

### From Browser on Local Network:

**1. Discover Page** ✅
- URL: http://192.168.1.49:5178/discover
- Shows real movies from TMDB API
- Categories: Trending, Popular, Now Playing, Upcoming, Top Rated

**2. Movie Detail Page** ✅
- URL: http://192.168.1.49:5178/movie/1084244
- Shows: Title, Download button, Add to Watchlist button
- Sections: Overview, Cast, Crew, Movie Info

**3. TV Interface (10-foot UI)** ✅
- URL: http://192.168.1.49:5178/tv
- TV navigation: Home, Discover, Watchlist, History, Settings
- Large text for TV viewing
- Works with D-pad/arrow keys

**4. Backend API** ✅
- Health: http://192.168.1.49:8084/api/health
- Returns: `{"status":"healthy"}`
- All endpoints accessible from network

**5. Jellyfin Server** ✅
- URL: http://192.168.1.49:8096
- Version: 10.11.5
- Ready for setup wizard

---

## 🏠 HOW TO ACCESS FROM OTHER DEVICES

### On Your Phone/Tablet:
1. Connect to same WiFi (192.168.1.x network)
2. Open browser
3. Go to: http://192.168.1.49:5178

### On Nvidia Shield (Android TV):
1. Open browser or Jellyfin app
2. For Media Manager: http://192.168.1.49:5178/tv
3. For Jellyfin: http://192.168.1.49:8096

### On Laptop/Desktop:
1. Open browser
2. Go to: http://192.168.1.49:5178

---

## 🔧 NETWORK CONFIGURATION

**Server IP**: 192.168.1.49
**Backend Port**: 8084 (binds to 0.0.0.0 - all interfaces)
**Frontend Port**: 5178 (Vite dev server, host: 0.0.0.0)
**Jellyfin Port**: 8096 (Docker, network_mode: host)

**Why it works**:
- Backend explicitly binds to `0.0.0.0:8084` (all network interfaces)
- Vite configured with `host: '0.0.0.0'`
- Jellyfin uses Docker host networking
- No firewall blocking on local network

---

## 🚀 QUICK START

```bash
# Start all services
cd /home/josh/media-manager
./scripts/dev-server.sh start

# Check status
./scripts/dev-server.sh status

# Access from any device:
# http://192.168.1.49:5178
```

---

## 🦞 ZOIDBERG'S VERDICT

"Why not Zoidberg? Your media center is accessible from EVERYWHERE on your network! Phone, tablet, laptop, Nvidia Shield - they can ALL access your movies! It's like a real Netflix server running in your house! HOORAY FOR NETWORKED FRIENDS! 🦞🎉"
