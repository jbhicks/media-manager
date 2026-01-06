# Wabbajack Mod Downloader Scripts

These scripts help you download all mods from a Wabbajack modlist without dealing with the slow, error-prone manual download prompts.

## Quick Start

### Step 1: Extract Mod List
```powershell
.\extract-wabbajack-mods.ps1 -ExtractOnly
```

This creates:
- `C:\wabbajack\downloads\modlist.csv` - Full list of mods
- `C:\wabbajack\downloads\download_urls.txt` - Nexus URLs
- `C:\wabbajack\downloads\aria2c_input.txt` - For aria2c downloader

### Step 2: Choose Download Method

#### Method A: Semi-Automated (Recommended)
Opens each mod page in your browser automatically. You click "Download" manually but don't have to search for each mod.

```powershell
.\download-wabbajack-mods.ps1 -OpenInBrowser
```

**Tips:**
- Set your browser to download to `C:\wabbajack\downloads`
- Wabbajack will detect already-downloaded files
- You can stop/resume anytime

#### Method B: Fully Automated (Experimental)
Uses browser automation to click download buttons for you. Requires Selenium.

```powershell
# First-time setup
.\auto-download-wabbajack-mods.ps1 -SetupOnly

# Then download (without login - you'll need to log in manually when browser opens)
.\auto-download-wabbajack-mods.ps1

# Or with auto-login
.\auto-download-wabbajack-mods.ps1 -NexusUsername "youruser" -NexusPassword "yourpass"

# Download in batches (resume-friendly)
.\auto-download-wabbajack-mods.ps1 -BatchSize 50  # Download 50 at a time
.\auto-download-wabbajack-mods.ps1 -StartFrom 51 -BatchSize 50  # Next batch
```

#### Method C: Nexus API (Premium Only)
If you have Nexus Premium, you can use the API for direct downloads.

```powershell
# Get your API key from https://www.nexusmods.com/users/myaccount?tab=api
.\download-wabbajack-mods.ps1 -NexusAPIKey "YOUR_API_KEY_HERE"
```

## Resuming Downloads

The scripts are designed to be resume-friendly:

1. **Wabbajack integration**: If you download to `C:\wabbajack\downloads`, Wabbajack will detect and use already-downloaded files
2. **Batch downloading**: Use `-BatchSize` and `-StartFrom` to download in chunks
3. **Check logs**: The main Wabbajack log shows what's already downloaded

## Troubleshooting

### "Execution policy" error
```powershell
Set-ExecutionPolicy -Scope Process -ExecutionPolicy Bypass
```

### Downloads going to wrong folder
Check your browser's download settings and set it to: `C:\wabbajack\downloads`

### Selenium not working
Make sure Chrome is installed. If issues persist:
```powershell
.\auto-download-wabbajack-mods.ps1 -SetupOnly
```

### Some mods fail
This is normal - some mods may be:
- Removed from Nexus
- Require manual authorization
- Behind adult content walls

For these, you'll need to download manually or let Wabbajack prompt you.

## Files Created

- `modlist.csv` - Complete list with Nexus IDs
- `download_urls.txt` - Direct links to mod pages  
- `aria2c_input.txt` - Input file for aria2c (advanced users)

## Integration with Wabbajack

After downloading:
1. Files should be in `C:\wabbajack\downloads`
2. Restart your Wabbajack installation
3. Wabbajack will verify and use the downloaded files
4. You'll only get prompts for mods that failed/weren't downloaded

## Notes

- **689 mods** total in Fallout VR Essentials
- **Free Nexus** = Manual downloads (slow but works)
- **Nexus Premium** = Faster downloads via API
- **Browser automation** = Middle ground (still free)

The semi-automated method (Method A) is recommended as it's reliable and doesn't require Premium, just some patience.

## Current Status

As of extraction, found **689 unique mods** to download for Fallout VR Essentials.

## Advanced: Using aria2c

If you want to use aria2c (fast downloader) with the Nexus API:

```powershell
# Install aria2
choco install aria2

# Download using API (Premium required for direct links)
aria2c -i C:\wabbajack\downloads\aria2c_input.txt -j 4 --header="apikey: YOUR_API_KEY"
```

Note: Standard Nexus downloads require authentication, so aria2c alone won't work without the API.
