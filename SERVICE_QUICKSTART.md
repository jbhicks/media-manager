# Media Manager Service - Quick Start

## Deploy Service (One Command)

```bash
make deploy-service
```

This single command will:
- ✅ Build the service binary
- ✅ Install to `~/.local/bin/`
- ✅ Configure systemd user service
- ✅ Enable auto-start on boot
- ✅ Start the service immediately
- ✅ Configure auto-restart on failure

## Service Management

```bash
# View service status
make service-status

# View live logs
make service-logs

# Restart service
make service-restart

# Stop service
make service-stop
```

## Direct systemctl Commands

```bash
# Status
systemctl --user status media-manager-service@$USER.service

# Logs
journalctl --user -u media-manager-service@$USER.service -f

# Restart
systemctl --user restart media-manager-service@$USER.service

# Stop
systemctl --user stop media-manager-service@$USER.service

# Disable auto-start
systemctl --user disable media-manager-service@$USER.service
```

## Web Interface

Once the service is running:
- **URL**: http://localhost:8080 (or next available port)
- The service will log the actual port it's using

## Features

The service includes:
- 🌐 Web UI for library and download management
- 🔍 Automated torrent search via Jackett
- 📥 Download queue management
- 📺 Jellyfin integration
- 🔄 Automatic scheduler for periodic searches
- 📊 Real-time download progress monitoring

## Configuration

Default paths:
- **Database**: `~/.media-manager/media.db`
- **Downloads**: `/mnt/media/Downloads`
- **Library**: `/mnt/media/Movies`
- **Binary**: `~/.local/bin/media-manager-service`
- **Service File**: `~/.config/systemd/user/media-manager-service@.service`

## Troubleshooting

### Service won't start
```bash
# Check logs for errors
journalctl --user -u media-manager-service@$USER.service -n 50

# Rebuild and redeploy
make deploy-service
```

### Find which port the service is using
```bash
journalctl --user -u media-manager-service@$USER.service | grep "Starting HTTP server"
```

### Service not starting on boot
```bash
# Enable lingering (allows user services to run without login)
loginctl enable-linger $USER
```

## Advanced Configuration

See [SERVICE.md](SERVICE.md) for detailed configuration including:
- Jackett setup
- Transmission configuration
- Download rules
- Scheduling options
- Database schema

## Uninstall

```bash
# Stop and disable service
systemctl --user stop media-manager-service@$USER.service
systemctl --user disable media-manager-service@$USER.service

# Remove files
rm ~/.local/bin/media-manager-service
rm ~/.config/systemd/user/media-manager-service@.service
systemctl --user daemon-reload

# Optionally remove data
rm -rf ~/.media-manager/
```
