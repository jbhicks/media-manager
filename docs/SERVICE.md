# Media Manager Service Mode

This guide explains how to set up and use the Media Manager in headless service mode for automatic torrent downloads.

## Overview

The service mode allows Media Manager to run as a background daemon that automatically:
- Searches torrent trackers for movies and TV shows
- Downloads matching content based on configurable rules
- Integrates with torrent clients (currently supports Transmission)
- Monitors and manages download tasks

## Architecture

The service uses:
- **SQLite Database** for configuration, download rules, tasks, and search results
- **Jackett** for unified torrent tracker searching
- **Transmission** (or compatible clients) for torrent downloads
- **Systemd** for service management on Linux

## Prerequisites

1. **Jackett** - Torrent tracker aggregator
   - Install from: https://github.com/Jackett/Jackett
   - Configure your preferred trackers
   - Note your API key

2. **Transmission** - Torrent client
   ```bash
   sudo apt install transmission-daemon  # Debian/Ubuntu
   sudo pacman -S transmission-cli        # Arch Linux
   ```

3. **Go 1.24+** - For building the service

## Building

```bash
cd media-manager

# Build service binary
make build-service

# Or build both GUI and service
make build-all
```

## Installation

### Option 1: System Service (Recommended)

```bash
# Build and create install script
make install-service

# Run the generated sudo script
./sudo_install_service.sh

# Enable service for your user
sudo systemctl enable media-manager-service@$USER
sudo systemctl start media-manager-service@$USER

# Check status
sudo systemctl status media-manager-service@$USER

# View logs
sudo journalctl -u media-manager-service@$USER -f
```

### Option 2: Manual Execution

```bash
# Run directly
./bin/media-manager-service --service

# Or with output to file
./bin/media-manager-service --service > service.log 2>&1 &
```

## Configuration

### Database Setup

The service uses the same SQLite database as the GUI at `~/.media-manager/media.db`.

On first run, the service creates a default configuration. You can modify it by directly updating the database or by creating a configuration utility.

### Configure Download Sources

Add Jackett as a download source:

```sql
sqlite3 ~/.media-manager/media.db

INSERT INTO download_sources (name, type, url, api_key, enabled, priority, created_at, updated_at)
VALUES (
  'Jackett',
  'jackett',
  'http://localhost:9117',
  'YOUR_JACKETT_API_KEY',
  1,
  10,
  datetime('now'),
  datetime('now')
);
```

### Configure Download Rules

Create automatic download rules:

```sql
-- Example: Auto-download 1080p movies
INSERT INTO download_rules (
  name, enabled, media_type, search_query, quality,
  min_seeders, auto_download, destination_path,
  created_at, updated_at
) VALUES (
  'Auto Download Movies 1080p',
  1,
  'movies',
  'New Release 2024',
  '1080p',
  5,
  1,
  '/media/downloads/movies',
  datetime('now'),
  datetime('now')
);

-- Example: Auto-download TV shows
INSERT INTO download_rules (
  name, enabled, media_type, search_query, quality,
  min_seeders, auto_download, destination_path,
  created_at, updated_at
) VALUES (
  'Auto Download Breaking Bad',
  1,
  'tv',
  'Breaking Bad',
  '1080p',
  3,
  1,
  '/media/downloads/tv',
  datetime('now'),
  datetime('now')
);
```

### Service Configuration

Update service settings:

```sql
-- Enable downloads with 5 minute intervals
UPDATE service_configs SET
  download_enabled = 1,
  schedule_interval = 300,
  max_concurrent_downloads = 3,
  torrent_client_type = 'transmission',
  torrent_client_host = 'localhost:9091',
  torrent_client_user = '',
  torrent_client_pass = '';
```

## Database Schema

Key tables:

- **download_sources** - Configured search providers (Jackett, etc.)
- **download_rules** - Automatic download rules
- **download_tasks** - Active and completed downloads
- **search_results** - Cached search results (24hr expiry)
- **service_configs** - Service configuration

## Monitoring

### Check Service Status

```bash
# Systemd status
sudo systemctl status media-manager-service@$USER

# View logs
sudo journalctl -u media-manager-service@$USER -n 100
```

### Monitor Downloads

```sql
-- View active downloads
SELECT id, title, status, progress, seeders
FROM download_tasks
WHERE status = 'downloading'
ORDER BY created_at DESC;

-- View completed downloads
SELECT id, title, download_path, completed_at
FROM download_tasks
WHERE status = 'completed'
ORDER BY completed_at DESC;

-- View download rules and last run time
SELECT id, name, enabled, last_run, search_query
FROM download_rules
ORDER BY last_run DESC;
```

## Transmission Setup

Configure Transmission for remote access:

1. Stop transmission daemon:
   ```bash
   sudo systemctl stop transmission-daemon
   ```

2. Edit `/etc/transmission-daemon/settings.json`:
   ```json
   {
     "rpc-enabled": true,
     "rpc-bind-address": "127.0.0.1",
     "rpc-port": 9091,
     "rpc-whitelist-enabled": false,
     "rpc-authentication-required": false
   }
   ```

3. Start transmission daemon:
   ```bash
   sudo systemctl start transmission-daemon
   ```

## Troubleshooting

### Service won't start

```bash
# Check logs for errors
sudo journalctl -u media-manager-service@$USER -n 50

# Verify binary exists
ls -la /usr/local/bin/media-manager-service

# Test manual execution
/usr/local/bin/media-manager-service --service
```

### No downloads happening

1. Verify downloads are enabled:
   ```sql
   SELECT download_enabled FROM service_configs;
   ```

2. Check if rules exist and are enabled:
   ```sql
   SELECT * FROM download_rules WHERE enabled = 1;
   ```

3. Verify Jackett is accessible:
   ```bash
   curl http://localhost:9117/api/v2.0/indexers/all/results?apikey=YOUR_KEY&Query=test
   ```

4. Check Transmission is running:
   ```bash
   sudo systemctl status transmission-daemon
   ```

### Downloads fail

1. Check download task errors:
   ```sql
   SELECT title, error FROM download_tasks WHERE status = 'failed';
   ```

2. Verify destination paths exist and are writable:
   ```bash
   ls -la /media/downloads/movies
   ```

3. Check Transmission download directory permissions

## Advanced Configuration

### Multiple Download Sources

You can add multiple Jackett instances or other tracker sources:

```sql
INSERT INTO download_sources (name, type, url, api_key, enabled, priority, created_at, updated_at)
VALUES 
  ('Jackett Primary', 'jackett', 'http://localhost:9117', 'key1', 1, 10, datetime('now'), datetime('now')),
  ('Jackett Secondary', 'jackett', 'http://tracker2.local:9117', 'key2', 1, 5, datetime('now'), datetime('now'));
```

### Quality Filters

Common quality strings for rules:
- `1080p`, `2160p`, `4K`, `720p`
- `BluRay`, `WEB-DL`, `WEBRip`
- `HEVC`, `x265`, `x264`

### Size Limits

Set `max_size` in bytes (e.g., 5GB = 5368709120):

```sql
UPDATE download_rules 
SET max_size = 5368709120 
WHERE name = 'Auto Download Movies 1080p';
```

## Future Features

Planned enhancements:
- Web UI for managing rules and downloads
- RSS feed monitoring
- Sonarr/Radarr integration
- qBittorrent support
- Notification system (webhook, email, Discord)
- Bandwidth scheduling
- Duplicate detection improvements

## Contributing

See the main README for contribution guidelines.
