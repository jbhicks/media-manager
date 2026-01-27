# Wabbajack Mod Downloader Script
# This script extracts mod information from a Wabbajack modlist and downloads them

param(
    [string]$WabbajackFile = "C:\wabbajack\4.0.5.0\downloaded_mod_lists\wj-featured_@@_fo4vre.wabbajack",
    [string]$OutputDir = "C:\wabbajack\downloads",
    [string]$NexusAPIKey = "",  # Optional: Get from https://www.nexusmods.com/users/myaccount?tab=api
    [switch]$ExtractOnly,  # Only extract list, don't download
    [switch]$UseAria2c  # Use aria2c for faster downloads
)

$ErrorActionPreference = "Stop"

Write-Host "=== Wabbajack Mod Extractor ===" -ForegroundColor Cyan
Write-Host "Modlist: $WabbajackFile" -ForegroundColor Yellow

# Create output directory
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# Function to extract .wabbajack file (it's a zip archive)
function Extract-ModlistMetadata {
    param([string]$WabbajackPath)
    
    Write-Host "`nExtracting modlist metadata..." -ForegroundColor Green
    
    $tempDir = Join-Path $env:TEMP "wabbajack_extract_$(Get-Random)"
    New-Item -ItemType Directory -Force -Path $tempDir | Out-Null
    
    try {
        # .wabbajack files are zip archives
        Add-Type -AssemblyName System.IO.Compression.FileSystem
        [System.IO.Compression.ZipFile]::ExtractToDirectory($WabbajackPath, $tempDir)
        
        # Look for modlist JSON
        $modlistFile = Get-ChildItem -Path $tempDir -Filter "modlist" -Recurse -File | Select-Object -First 1
        
        if ($modlistFile) {
            Write-Host "Found modlist file: $($modlistFile.FullName)" -ForegroundColor Green
            $modlistJson = Get-Content $modlistFile.FullName -Raw | ConvertFrom-Json
            return $modlistJson
        } else {
            Write-Host "ERROR: Could not find modlist file in archive" -ForegroundColor Red
            return $null
        }
    } finally {
        # Cleanup
        if (Test-Path $tempDir) {
            Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
        }
    }
}

# Function to parse log file for mod information
function Get-ModsFromLog {
    param([string]$LogPath)
    
    Write-Host "`nParsing Wabbajack log file..." -ForegroundColor Green
    
    $mods = @()
    $logContent = Get-Content $LogPath
    
    foreach ($line in $logContent) {
        if ($line -match 'Downloading (.+?)(\||$)') {
            $filename = $matches[1].Trim()
            
            # Try to extract Nexus mod ID from filename pattern: modname-MODID-version.ext
            if ($filename -match '-(\d+)-[\d\-\.]+\.(7z|zip|rar)$') {
                $modId = $matches[1]
                $mods += [PSCustomObject]@{
                    FileName = $filename
                    ModID = $modId
                    Game = "fallout4"  # Fallout 4 VR uses fallout4 on Nexus
                }
            } else {
                $mods += [PSCustomObject]@{
                    FileName = $filename
                    ModID = $null
                    Game = "fallout4"
                }
            }
        }
    }
    
    return $mods | Sort-Object -Property FileName -Unique
}

# Function to generate download URLs
function Get-NexusDownloadUrl {
    param(
        [string]$Game,
        [int]$ModID,
        [string]$FileName
    )
    
    # Nexus manual download URL format
    return "https://www.nexusmods.com/$Game/mods/$ModID?tab=files"
}

# Main execution
$logFile = "C:\wabbajack\4.0.5.0\logs\Wabbajack.current.log"

if (Test-Path $logFile) {
    Write-Host "Using log file: $logFile" -ForegroundColor Yellow
    $mods = Get-ModsFromLog -LogPath $logFile
    
    Write-Host "`nFound $($mods.Count) unique mods" -ForegroundColor Green
    
    # Export to CSV for easy viewing
    $csvPath = Join-Path $OutputDir "modlist.csv"
    $mods | Export-Csv -Path $csvPath -NoTypeInformation -Encoding UTF8
    Write-Host "Exported mod list to: $csvPath" -ForegroundColor Green
    
    # Create download URLs file
    $urlsPath = Join-Path $OutputDir "download_urls.txt"
    $mods | Where-Object { $_.ModID } | ForEach-Object {
        Get-NexusDownloadUrl -Game $_.Game -ModID $_.ModID -FileName $_.FileName
    } | Out-File $urlsPath -Encoding UTF8
    
    Write-Host "Exported Nexus URLs to: $urlsPath" -ForegroundColor Green
    
    # Create aria2c input file for batch downloading
    $aria2Path = Join-Path $OutputDir "aria2c_input.txt"
    $mods | Where-Object { $_.ModID } | ForEach-Object {
        $url = Get-NexusDownloadUrl -Game $_.Game -ModID $_.ModID -FileName $_.FileName
        "$url`n  out=$($_.FileName)"
    } | Out-File $aria2Path -Encoding UTF8
    
    Write-Host "Exported aria2c input to: $aria2Path" -ForegroundColor Green
    
    # Show some sample mods
    Write-Host "`nSample mods (first 10):" -ForegroundColor Cyan
    $mods | Select-Object -First 10 | Format-Table -AutoSize
    
    if (-not $ExtractOnly) {
        Write-Host "`n=== Download Instructions ===" -ForegroundColor Yellow
        Write-Host "Option 1: Manual Download from Nexus" -ForegroundColor White
        Write-Host "  - Open: $urlsPath" -ForegroundColor Gray
        Write-Host "  - Visit each URL and manually download" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Option 2: Use NexusMods API (requires Premium)" -ForegroundColor White
        Write-Host "  - Get API key from: https://www.nexusmods.com/users/myaccount?tab=api" -ForegroundColor Gray
        Write-Host "  - Run script with: -NexusAPIKey 'YOUR_KEY'" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Option 3: Use aria2c (fast multi-threaded downloader)" -ForegroundColor White
        Write-Host "  - Install aria2c: choco install aria2 (or download from GitHub)" -ForegroundColor Gray
        Write-Host "  - Note: Nexus requires authentication, aria2c won't work without API" -ForegroundColor Gray
        Write-Host ""
        Write-Host "Option 4: Continue using Wabbajack (recommended)" -ForegroundColor White
        Write-Host "  - Wabbajack handles authentication and CDN selection automatically" -ForegroundColor Gray
    }
    
} else {
    Write-Host "ERROR: Log file not found at $logFile" -ForegroundColor Red
    Write-Host "Please run Wabbajack first to generate the log" -ForegroundColor Yellow
}

Write-Host "`n=== Script Complete ===" -ForegroundColor Cyan
