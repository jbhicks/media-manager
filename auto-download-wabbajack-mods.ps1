# Automated Wabbajack Downloader with Browser Automation
# Uses Selenium WebDriver to automate Nexus downloads

param(
    [string]$ModlistCSV = "C:\wabbajack\downloads\modlist.csv",
    [string]$DownloadDir = "C:\wabbajack\downloads",
    [string]$NexusUsername = "",
    [string]$NexusPassword = "",
    [switch]$SetupOnly,  # Just install dependencies
    [int]$StartFrom = 1,  # Resume from specific mod number
    [int]$BatchSize = 0  # Download only N mods (0 = all)
)

$ErrorActionPreference = "Stop"

Write-Host "=== Automated Wabbajack Downloader ===" -ForegroundColor Cyan

# Check if Selenium is installed
function Test-Selenium {
    try {
        $null = Get-Package Selenium.WebDriver -ErrorAction Stop
        return $true
    } catch {
        return $false
    }
}

# Install Selenium
function Install-Selenium {
    Write-Host "Installing Selenium WebDriver..." -ForegroundColor Yellow
    
    # Install NuGet provider
    if (-not (Get-PackageProvider -Name NuGet -ErrorAction SilentlyContinue)) {
        Install-PackageProvider -Name NuGet -Force -Scope CurrentUser
    }
    
    # Install Selenium
    Install-Package Selenium.WebDriver -Force -Scope CurrentUser -Source https://www.nuget.org/api/v2
    Install-Package Selenium.WebDriver.ChromeDriver -Force -Scope CurrentUser -Source https://www.nuget.org/api/v2
    
    Write-Host "Selenium installed!" -ForegroundColor Green
}

# Setup check
if ($SetupOnly) {
    if (Test-Selenium) {
        Write-Host "Selenium is already installed" -ForegroundColor Green
    } else {
        Install-Selenium
    }
    Write-Host "`nSetup complete!" -ForegroundColor Cyan
    Write-Host "Now run without -SetupOnly to start downloading" -ForegroundColor Yellow
    exit 0
}

# Verify Selenium
if (-not (Test-Selenium)) {
    Write-Host "ERROR: Selenium not installed" -ForegroundColor Red
    Write-Host "Run with -SetupOnly first to install dependencies" -ForegroundColor Yellow
    exit 1
}

# Load Selenium assemblies
$seleniumPath = (Get-Package Selenium.WebDriver).Source
$chromeDriverPath = (Get-Package Selenium.WebDriver.ChromeDriver).Source
Add-Type -Path (Join-Path (Split-Path $seleniumPath) "lib\netstandard2.0\WebDriver.dll")

# Load mod list
if (-not (Test-Path $ModlistCSV)) {
    Write-Host "ERROR: Modlist CSV not found" -ForegroundColor Red
    Write-Host "Run extract-wabbajack-mods.ps1 first!" -ForegroundColor Yellow
    exit 1
}

$allMods = Import-Csv $ModlistCSV | Where-Object { $_.ModID -and $_.ModID -ne "" }
Write-Host "Loaded $($allMods.Count) mods" -ForegroundColor Green

# Apply filters
$mods = $allMods | Select-Object -Skip ($StartFrom - 1)
if ($BatchSize -gt 0) {
    $mods = $mods | Select-Object -First $BatchSize
}

Write-Host "Will process $($mods.Count) mods (starting from #$StartFrom)" -ForegroundColor Yellow

# Setup Chrome driver
$chromeOptions = New-Object OpenQA.Selenium.Chrome.ChromeOptions
$chromeOptions.AddUserProfilePreference("download.default_directory", $DownloadDir)
$chromeOptions.AddUserProfilePreference("download.prompt_for_download", $false)
$chromeOptions.AddUserProfilePreference("disable-popup-blocking", "true")

$driverService = [OpenQA.Selenium.Chrome.ChromeDriverService]::CreateDefaultService((Split-Path $chromeDriverPath))
$driver = New-Object OpenQA.Selenium.Chrome.ChromeDriver($driverService, $chromeOptions)

try {
    # Login to Nexus if credentials provided
    if ($NexusUsername -and $NexusPassword) {
        Write-Host "`nLogging in to Nexus Mods..." -ForegroundColor Yellow
        $driver.Navigate().GoToUrl("https://users.nexusmods.com/auth/sign_in")
        Start-Sleep -Seconds 2
        
        $driver.FindElementById("user_login").SendKeys($NexusUsername)
        $driver.FindElementById("user_password").SendKeys($NexusPassword)
        $driver.FindElementByName("commit").Click()
        
        Start-Sleep -Seconds 3
        Write-Host "Logged in!" -ForegroundColor Green
    } else {
        Write-Host "`nWARNING: No credentials provided, you may need to log in manually" -ForegroundColor Yellow
        Write-Host "The browser will open - please log in when prompted" -ForegroundColor Gray
        Start-Sleep -Seconds 5
    }
    
    # Download each mod
    $count = 0
    foreach ($mod in $mods) {
        $count++
        $modNum = $StartFrom + $count - 1
        
        Write-Host "`n[$modNum/$($allMods.Count)] Processing: $($mod.FileName)" -ForegroundColor Cyan
        
        $url = "https://www.nexusmods.com/$($mod.Game)/mods/$($mod.ModID)?tab=files"
        $driver.Navigate().GoToUrl($url)
        Start-Sleep -Seconds 2
        
        try {
            # Look for download buttons
            $downloadButtons = $driver.FindElementsByXPath("//button[contains(text(), 'Manual Download') or contains(text(), 'Mod Manager Download')]")
            
            if ($downloadButtons.Count -gt 0) {
                $downloadButtons[0].Click()
                Write-Host "  Clicked download button" -ForegroundColor Green
                Start-Sleep -Seconds 2
                
                # Handle any popups or confirmations
                try {
                    $slowDownload = $driver.FindElementByXPath("//a[contains(text(), 'Slow Download')]")
                    if ($slowDownload) {
                        $slowDownload.Click()
                        Write-Host "  Started download" -ForegroundColor Green
                    }
                } catch {
                    Write-Host "  Download started (no confirmation needed)" -ForegroundColor Green
                }
                
                Start-Sleep -Seconds 3
            } else {
                Write-Host "  WARNING: No download button found" -ForegroundColor Yellow
            }
            
        } catch {
            Write-Host "  ERROR: $($_.Exception.Message)" -ForegroundColor Red
        }
        
        # Small delay between mods
        Start-Sleep -Seconds 2
    }
    
    Write-Host "`n=== Download Complete ===" -ForegroundColor Cyan
    Write-Host "Check $DownloadDir for downloaded files" -ForegroundColor Green
    Write-Host "`nPress Enter to close browser..." -ForegroundColor Yellow
    Read-Host
    
} finally {
    $driver.Quit()
}
