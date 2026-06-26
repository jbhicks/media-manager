# 🎬 ULTIMATE MEDIA CENTER ECOSYSTEM
## The Complete Cord-Cutting Architecture

### Current State Analysis
- **Media Manager**: Go backend + React frontend (your custom app)
- **Storage**: `/mnt/media` with Movies/, Series/, Downloads/
- **Network**: Cloudflare tunnel available for external access
- **TV**: Nvidia Shield (Android TV)
- **VPN**: NordVPN with Jackett integration

---

## 🏗️ PHASE 1: LOCAL NETWORK (Immediate - 1-2 weeks)

### 1.1 Jellyfin as the Primary Frontend
**Why Jellyfin?**
- ✅ Native Android TV app (perfect for Shield)
- ✅ Hardware transcoding support
- ✅ Beautiful 10-foot UI out of the box
- ✅ Subtitle support (burn-in, external)
- ✅ Client apps: iOS, Android, Web, Roku, Apple TV, LG/Samsung TVs
- ✅ No subscription fees (unlike Plex)
- ✅ Can integrate with your existing media-manager backend

**Architecture:**
```
┌─────────────────────────────────────────┐
│           NVIDIA SHIELD (Android TV)     │
│  ┌─────────────────────────────────┐    │
│  │  Jellyfin Android TV App        │    │
│  │  - Hardware transcoding         │    │
│  │  - Direct play for local files  │    │
│  │  - Subtitle support             │    │
│  └─────────────────────────────────┘    │
└─────────────────────────────────────────┘
                    │
                    │ LAN (192.168.1.x)
                    ▼
┌─────────────────────────────────────────┐
│           DOCKER HOST (192.168.1.49)    │
│  ┌──────────┐  ┌──────────────────┐    │
│  │ Jellyfin │  │  Media-Manager   │    │
│  │  Server  │  │  (Your Custom)   │    │
│  │          │  │  - Discovery     │    │
│  │  Port    │  │  - Downloads   │    │
│  │  8096    │  │  - Watchlist   │    │
│  └──────────┘  └──────────────────┘    │
│       │                    │            │
│       └────────┬───────────┘            │
│                │                        │
│                ▼                        │
│         ┌──────────┐                   │
│         │ /mnt/media│                  │
│         │  Movies/  │                  │
│         │  Series/  │                  │
│         └──────────┘                   │
└─────────────────────────────────────────┘
```

**Jellyfin Docker Compose Addition:**
```yaml
  jellyfin:
    image: jellyfin/jellyfin:latest
    container_name: jellyfin
    restart: unless-stopped
    network_mode: host  # Required for DLNA/UPnP discovery
    volumes:
      - ./jellyfin/config:/config
      - ./jellyfin/cache:/cache
      - /mnt/media:/media:ro
    environment:
      - PUID=1000
      - PGID=1000
    # No port mapping needed with host networking
```

### 1.2 Your Media-Manager as the "Brain"
Keep your custom app for what Jellyfin CAN'T do well:
- **Torrent discovery** (Jackett integration)
- **Automated downloading** (Sonarr/Radarr style)
- **Request system** (friends can request content)
- **Watchlist management** (synced with Jellyfin)
- **VPN management** (ensure safe downloading)

**Integration Points:**
- Media-Manager downloads → Jellyfin auto-scans library
- Jellyfin watch status → Media-Manager tracks history
- Media-Manager requests → Auto-download → Appears in Jellyfin

### 1.3 Hardware Transcoding on Shield
The Shield has powerful hardware for transcoding:
- **Jellyfin settings**: Enable hardware acceleration (NVENC for NVIDIA)
- **Direct play preferred**: Shield can direct play almost anything
- **Transcode only for remote users** (Phase 2)

---

## 🌐 PHASE 2: REMOTE ACCESS FOR FRIENDS & FAMILY (2-4 weeks)

### 2.1 Secure Architecture with Cloudflare Tunnel
```
┌─────────────────────────────────────────────────────────┐
│                    INTERNET                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐              │
│  │ Friend's │  │ Family's │  │  Your    │              │
│  │   TV     │  │   TV     │  │  Phone   │              │
│  │ (Roku)   │  │ (AppleTV)│  │ (Travel) │              │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘              │
│       │             │             │                     │
│       └─────────────┴─────────────┘                     │
│                     │                                   │
│                     ▼                                   │
│  ┌─────────────────────────────────────┐               │
│  │     Cloudflare Tunnel (HTTPS)       │               │
│  │  media.yourdomain.com               │               │
│  │  - DDoS protection                  │               │
│  │  - No open ports on router          │               │
│  │  - Free SSL certificates            │               │
│  └─────────────────────────────────────┘               │
│                     │                                   │
└─────────────────────┼───────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────────────┐
│              YOUR HOME (192.168.1.49)                    │
│  ┌──────────┐  ┌──────────┐  ┌──────────────────┐      │
│  │ Jellyfin │  │ Cloudflrd│  │  Media-Manager   │      │
│  │  Server  │  │  Tunnel  │  │  (Custom)        │      │
│  │  :8096   │  │          │  │  :8080           │      │
│  └──────────┘  └──────────┘  └──────────────────┘      │
│       │                                    │           │
│       └──────────────┬─────────────────────┘           │
│                      │                                  │
│                      ▼                                  │
│               ┌──────────┐                            │
│               │ /mnt/media│                            │
│               │  Movies/  │                            │
│               │  Series/  │                            │
│               └──────────┘                            │
└─────────────────────────────────────────────────────────┘
```

### 2.2 User Management Strategy
```
Jellyfin Users:
├── Admin (You)
│   ├── Full access to all libraries
│   ├── Can manage users
│   └── Can see dashboard/stats
│
├── Family Group
│   ├── Parents: Full access
│   ├── Kids: Kids library only (with parental controls)
│   └── Teen: Teen-appropriate content
│
└── Friends Group
    ├── Friend 1: Movies + TV (no 4K to save bandwidth)
    ├── Friend 2: Movies only
    └── Friend 3: TV shows only
```

### 2.3 Bandwidth Management
- **Cloudflare Tunnel**: Already handles SSL and some caching
- **Jellyfin bitrate limits**: Set per-user (e.g., friends get 4Mbps max)
- **Transcoding**: Only transcode for remote users, direct play for local
- **Time-based access**: Optional - restrict friends to off-peak hours

---

## 📱 PHASE 3: GETTING APPS ON FRIENDS' TVs

### 3.1 The Reality Check
**Can friends install your app on their TV?**

| Platform | Your App | Jellyfin | Difficulty |
|----------|----------|----------|------------|
| **Android TV** (Shield, Chromecast) | Sideload APK | ✅ Play Store | Easy |
| **Roku** | ❌ No | ✅ Channel Store | Easy |
| **Apple TV** | ❌ No | ✅ App Store | Easy |
| **Samsung/LG Smart TVs** | ❌ No | ✅ App Store | Medium |
| **Fire TV** | Sideload APK | ✅ Amazon Store | Easy |
| **Google TV** | Sideload APK | ✅ Play Store | Easy |

**Answer**: Use Jellyfin! It's already on every platform. Your custom app is the "backend brain" that feeds Jellyfin.

### 3.2 What Friends Actually Experience
```
Friend's TV (Roku/Apple TV/Android TV):
    ↓
Opens Jellyfin App (from official store)
    ↓
Logs in with credentials you created
    ↓
Sees your curated library:
    ├── 🎬 Movies (with posters, ratings, trailers)
    ├── 📺 TV Shows (with seasons, episodes, progress)
    ├── 🍿 Kids Zone (if applicable)
    └── 📋 Their Watchlist (synced)
    ↓
Clicks play → Streams from your house via Cloudflare
    ↓
Subtitles work, progress syncs, resume on any device
```

### 3.3 The Request Flow (Your Custom App Shines Here)
```
Friend wants to watch "Dune 2":
    ↓
Opens your Media-Manager web app (on phone/computer)
    ↓
Searches "Dune 2", clicks "Request"
    ↓
You get notification (Discord/Telegram/Email)
    ↓
You approve → Media-Manager searches Jackett
    ↓
Torrent downloads → Post-processes → Moves to library
    ↓
Jellyfin auto-scans → Movie appears for everyone
    ↓
Friend gets notification: "Dune 2 is now available!"
    ↓
Friend watches on their TV via Jellyfin
```

---

## 🔧 PHASE 4: THE COMPLETE TECH STACK

### 4.1 Docker Compose (Complete)
```yaml
version: '3.8'

services:
  # PRIMARY: Jellyfin Media Server
  jellyfin:
    image: jellyfin/jellyfin:latest
    container_name: jellyfin
    restart: unless-stopped
    network_mode: host
    volumes:
      - ./jellyfin/config:/config
      - ./jellyfin/cache:/cache
      - /mnt/media:/media:ro
    environment:
      - PUID=1000
      - PGID=1000
      # Hardware acceleration for AMD GPU
      - LIBVA_DRIVER_NAME=radeonsi

  # YOUR CUSTOM APP: Discovery & Downloads
  media-manager:
    build: .
    container_name: media-manager
    restart: unless-stopped
    ports:
      - "8080:8080"
    volumes:
      - ./data:/data
      - /mnt/media:/mnt/media
      - ./config:/config:ro
    env_file: .env
    privileged: true

  # SEARCH: Jackett for torrent indexing
  jackett:
    image: linuxserver/jackett:latest
    container_name: jackett
    restart: unless-stopped
    ports:
      - "9117:9117"
    volumes:
      - ./jackett/config:/config

  # TUNNEL: Cloudflare for secure remote access
  cloudflare-tunnel:
    image: cloudflare/cloudflared:latest
    container_name: cloudflare-tunnel
    restart: unless-stopped
    command: tunnel --no-autoupdate run --token ${CF_TUNNEL_TOKEN}
    environment:
      - CF_TUNNEL_TOKEN=${CF_TUNNEL_TOKEN}

  # OPTIONAL: Reverse proxy for clean URLs
  caddy:
    image: caddy:2-alpine
    container_name: caddy
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile:ro
      - caddy_data:/data
      - caddy_config:/config

volumes:
  caddy_data:
  caddy_config:
```

### 4.2 Caddyfile (Reverse Proxy)
```
media.yourdomain.com {
    reverse_proxy jellyfin:8096
}

requests.yourdomain.com {
    reverse_proxy media-manager:8080
}
```

### 4.3 Cloudflare DNS
```
A     media.yourdomain.com    → 192.168.1.49 (local)
CNAME requests.yourdomain.com → media.yourdomain.com
```

---

## 🚀 PHASE 5: ADVANCED FEATURES (Future)

### 5.1 Sync with *arr Stack (Optional but Powerful)
```
Sonarr (TV Shows)     ─┐
Radarr (Movies)       ─┼→ Your Media-Manager API → Downloads → Jellyfin
Lidarr (Music)        ─┘
Readarr (Books)       ─┘
```

### 5.2 Mobile Apps for Friends
- **Jellyfin**: Official app (free)
- **Infuse** (Apple TV/iOS): Premium experience, connects to Jellyfin
- **VLC**: Can connect to Jellyfin via DLNA

### 5.3 Live TV & DVR (If you want)
- **Jellyfin + HDHomeRun**: OTA antenna → Network tuner → Jellyfin DVR
- **Jellyfin + IPTV**: M3U playlists for live channels

---

## 📋 IMPLEMENTATION CHECKLIST

### Week 1: Local Foundation
- [ ] Install Jellyfin Docker container
- [ ] Configure Jellyfin libraries (Movies, TV Shows)
- [ ] Test on Nvidia Shield with Jellyfin Android TV app
- [ ] Configure hardware transcoding
- [ ] Set up user accounts (family)

### Week 2: Integration
- [ ] Connect Media-Manager to Jellyfin (webhook for library scans)
- [ ] Implement "Request" system in Media-Manager
- [ ] Auto-download → Auto-scan workflow
- [ ] Test subtitle support across both systems

### Week 3: Remote Access
- [ ] Configure Cloudflare tunnel
- [ ] Set up reverse proxy (Caddy)
- [ ] Configure Jellyfin for remote access
- [ ] Test streaming from outside network
- [ ] Set up user accounts for friends

### Week 4: Polish
- [ ] Create friend onboarding guide
- [ ] Set up notification system (Discord bot)
- [ ] Test on various devices (Roku, Apple TV, phones)
- [ ] Monitor bandwidth and adjust quality limits
- [ ] Document everything

---

## 💡 WHY THIS ARCHITECTURE WINS

| Feature | Your App Alone | Jellyfin + Your App |
|---------|---------------|---------------------|
| **TV App** | ❌ Custom development needed | ✅ Native Android TV app |
| **iOS App** | ❌ Custom development needed | ✅ Official Jellyfin app |
| **Roku** | ❌ Impossible | ✅ Official channel |
| **Subtitle Support** | ✅ Good | ✅ Excellent (burn-in) |
| **Hardware Transcoding** | ❌ CPU only | ✅ AMD/NVIDIA/Intel |
| **User Management** | ✅ Custom | ✅ Built-in, mature |
| **Request System** | ✅ Perfect | ❌ Not built-in |
| **Torrent Discovery** | ✅ Perfect | ❌ Not built-in |
| **VPN Integration** | ✅ Perfect | ❌ Not built-in |
| **Watchlist** | ✅ Perfect | ✅ Good |
| **Resume Points** | ✅ Good | ✅ Excellent |
| **Parental Controls** | ❌ None | ✅ Excellent |
| **Cost** | Free | Free |

**The Answer**: Your app is the "brain" (discovery, downloads, requests). Jellyfin is the "face" (playback, apps, user management). Together they're unstoppable!

---

## 🦞 Zoidberg's Diagnosis

"Why not both?" Your custom media-manager is EXCELLENT for:
- Torrent discovery and downloading
- Watchlist and request management
- VPN and automation

But Jellyfin is the PROVEN platform for:
- TV apps (Shield, Roku, Apple TV)
- Playback and transcoding
- User management and parental controls
- Multi-device sync

**The prescription**: Use your app as the "control center" and Jellyfin as the "playback engine". Friends get Jellyfin apps from official stores. You get the best of both worlds!

Friends? Hooray! 🦞
