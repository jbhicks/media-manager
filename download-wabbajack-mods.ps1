# Advanced Wabbajack Mod Downloader
# Downloads mods from extracted Wabbajack modlist using Nexus API

param(
    [string]$ModlistCSV = "C:\wabbajack\downloads\modlist.csv",
    [string]$OutputDir = "C:\wabbajack\manual_downloads",
    [string]$NexusAPIKey = "",  # Get from https://www.nexusmods.com/users/myaccount?tab=api
    [int]$Parallel = 4,  # Number of concurrent downloads
    [switch]$OpenInBrowser  # Open each mod page in browser for manual download
)

$ErrorActionPreference = "Continue"

Write-Host "=== Wabbajack Mod Downloader ===" -ForegroundColor Cyan

# Create output directory
New-Item -ItemType Directory -Force -Path $OutputDir | Out-Null

# Load mod list
if (-not (Test-Path $ModlistCSV)) {
    Write-Host "ERROR: Modlist CSV not found at $ModlistCSV" -ForegroundColor Red
    Write-Host "Run extract-wabbajack-mods.ps1 first!" -ForegroundColor Yellow
    exit 1
}

$mods = Import-Csv $ModlistCSV | Where-Object { $_.ModID -and $_.ModID -ne "" }
Write-Host "Loaded $($mods.Count) mods with Nexus IDs" -ForegroundColor Green

# Function to download from Nexus with API
function Download-NexusMod {
    param(
        [string]$ApiKey,
        [string]$Game,
        [int]$ModID,
        [string]$FileName,
        [string]$DestPath
    )
    
    $headers = @{
        "apikey" = $ApiKey
        "User-Agent" = "Wabbajack-Helper/1.0"
    }
    
    try {
        # Get mod info
        $modInfoUrl = "https://api.nexusmods.com/v1/games/$Game/mods/$ModID.json"
        $modInfo = Invoke-RestMethod -Uri $modInfoUrl -Headers $headers
        
        Write-Host "  Mod: $($modInfo.name)" -ForegroundColor Cyan
        
        # Get file list
        $filesUrl = "https://api.nexusmods.com/v1/games/$Game/mods/$ModID/files.json"
        $filesInfo = Invoke-RestMethod -Uri $filesUrl -Headers $headers
        
        # Find matching file
        $file = $filesInfo.files | Where-Object { $_.file_name -eq $FileName } | Select-Object -First 1
        
        if (-not $file) {
            Write-Host "  WARNING: Exact file not found, trying first main file..." -ForegroundColor Yellow
            $file = $filesInfo.files | Where-Object { $_.category_name -eq "MAIN" } | Select-Object -First 1
        }
        
        if ($file) {
            # Get download link (requires Premium)
            $downloadUrl = "https://api.nexusmods.com/v1/games/$Game/mods/$ModID/files/$($file.file_id)/download_link.json"
            $downloadLinks = Invoke-RestMethod -Uri $downloadUrl -Headers $headers
            
            if ($downloadLinks -and $downloadLinks.Count -gt 0) {
                $dlUrl = $downloadLinks[0].URI
                $outFile = Join-Path $DestPath $file.file_name
                
                Write-Host "  Downloading: $($file.file_name)" -ForegroundColor Green
                Invoke-WebRequest -Uri $dlUrl -OutFile $outFile -Headers $headers
                Write-Host "  SUCCESS: Downloaded to $outFile" -ForegroundColor Green
                return $true
            }
        }
        
        Write-Host "  ERROR: Could not get download link (Premium required?)" -ForegroundColor Red
        return $false
        
    } catch {
        Write-Host "  ERROR: $($_.Exception.Message)" -ForegroundColor Red
        return $false
    }
}

# Function to open in browser
function Open-ModInBrowser {
    param(
        [string]$Game,
        [int]$ModID
    )
    
    $url = "https://www.nexusmods.com/$Game/mods/${ModID}?tab=files"
    Start-Process $url
}

# Main execution
if ($OpenInBrowser) {
    Write-Host "`nOpening mod pages in browser for manual download..." -ForegroundColor Yellow
    Write-Host "Press CTRL+C to stop" -ForegroundColor Gray
    
    $count = 0
    foreach ($mod in $mods) {
        $count++
        Write-Host "[$count/$($mods.Count)] Opening: $($mod.FileName)" -ForegroundColor Cyan
        Open-ModInBrowser -Game $mod.Game -ModID $mod.ModID
        Start-Sleep -Milliseconds 1500  # Delay to avoid overwhelming browser
    }
    
} elseif ($NexusAPIKey) {
    Write-Host "`nDownloading with Nexus API..." -ForegroundColor Yellow
    Write-Host "Note: This requires Nexus Premium for direct downloads" -ForegroundColor Gray
    
    $successful = 0
    $failed = 0
    
    foreach ($mod in $mods) {
        Write-Host "`n[$($successful + $failed + 1)/$($mods.Count)] Processing: $($mod.FileName)" -ForegroundColor Cyan
        
        if (Download-NexusMod -ApiKey $NexusAPIKey -Game $mod.Game -ModID $mod.ModID -FileName $mod.FileName -DestPath $OutputDir) {
            $successful++
        } else {
            $failed++
        }
        
        Start-Sleep -Milliseconds 500  # Rate limiting
    }
    
    Write-Host "`n=== Download Summary ===" -ForegroundColor Cyan
    Write-Host "Successful: $successful" -ForegroundColor Green
    Write-Host "Failed: $failed" -ForegroundColor Red
    
} else {
    Write-Host "`n=== Instructions ===" -ForegroundColor Yellow
    Write-Host "Choose a download method:" -ForegroundColor White
    Write-Host ""
    Write-Host "Method 1: Open all mod pages in browser (recommended)" -ForegroundColor Cyan
    Write-Host "  .\download-wabbajack-mods.ps1 -OpenInBrowser" -ForegroundColor Gray
    Write-Host "  Then manually download each file (slow but reliable)" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Method 2: Use Nexus API (requires Premium)" -ForegroundColor Cyan
    Write-Host "  Get API key: https://www.nexusmods.com/users/myaccount?tab=api" -ForegroundColor Gray
    Write-Host "  .\download-wabbajack-mods.ps1 -NexusAPIKey 'YOUR_KEY'" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Method 3: Use Wabbajack's download folder" -ForegroundColor Cyan
    Write-Host "  Point your browser to download to: C:\wabbajack\downloads" -ForegroundColor Gray
    Write-Host "  Then open pages with: -OpenInBrowser" -ForegroundColor Gray
    Write-Host "  Wabbajack will detect already-downloaded files" -ForegroundColor Gray
}

Write-Host "`n=== Script Complete ===" -ForegroundColor Cyan
