# TODO

## Current Task: Fixing NordVPN Connection Issue

### Problem
- NordVPN failing to connect, blocking all downloads
- Download manager has VPN protection that refuses to download without active VPN
- 42 downloads were stuck at 0% progress (cleared to failed status)

### Root Cause
- Missing `wireguard-tools` package required for NordLynx (NordVPN's WireGuard implementation)

### Steps Taken
1. ✅ Identified VPN connection failures with both NordLynx and OpenVPN
2. ✅ Cleared 42 stuck downloads in "downloading" status (set to "failed")
3. ✅ Discovered missing wireguard-tools dependency
4. ✅ Installed wireguard-tools package
5. 🔄 Currently: Completing NordVPN OAuth login via browser

### Next Steps
1. Complete browser login for NordVPN
2. Test VPN connection with `nordvpn connect`
3. Verify download system resumes with VPN active
4. Monitor download progress

### Database State
- Location: `/home/josh/.media-manager/media.db`
- Current status: 42 failed, 3 pending downloads
- Service running on port 8083 via air auto-reload

### Files Modified
- None yet - only database cleanup performed
