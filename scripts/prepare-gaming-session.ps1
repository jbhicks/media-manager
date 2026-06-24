# Gaming session cleanup for Windows.
#
# Default mode is a dry run. Use -Apply to actually stop processes and services.
# Use -IncludeLaunchers if you also want to close Steam and the EA app.
# Use -StopVpnServices to stop NordVPN background services.
# Use -DisableVpnAdapter to disable NordVPN adapters as well (requires elevation).

[CmdletBinding()]
param(
    [switch]$Apply,
    [switch]$IncludeLaunchers,
    [switch]$StopVpnServices,
    [switch]$DisableVpnAdapter,
    [switch]$IncludeSearchIndex
)

$ErrorActionPreference = 'Stop'

function Write-Section {
    param(
        [string]$Title
    )

    Write-Host "`n== $Title ==" -ForegroundColor Cyan
}

function Write-Action {
    param(
        [string]$Message,
        [ConsoleColor]$Color = [ConsoleColor]::Yellow
    )

    Write-Host $Message -ForegroundColor $Color
}

function Invoke-StopProcesses {
    param(
        [string[]]$Names,
        [string]$Reason
    )

    foreach ($name in $Names) {
        $processes = @(Get-Process -Name $name -ErrorAction SilentlyContinue)
        if ($processes.Count -eq 0) {
            Write-Action "Skip process $name ($Reason): not running" ([ConsoleColor]::DarkGray)
            continue
        }

        foreach ($process in $processes) {
            $message = "Stop process $($process.ProcessName) (PID $($process.Id)) - $Reason"
            if (-not $Apply) {
                Write-Action "DRY RUN: $message"
                continue
            }

            try {
                Stop-Process -Id $process.Id -Force
                Write-Action "Stopped process $($process.ProcessName) (PID $($process.Id))" ([ConsoleColor]::Green)
            }
            catch {
                Write-Action "Failed to stop process $($process.ProcessName) (PID $($process.Id)): $($_.Exception.Message)" ([ConsoleColor]::Red)
            }
        }
    }
}

function Invoke-StopServices {
    param(
        [string[]]$Names,
        [string]$Reason
    )

    foreach ($name in $Names) {
        $service = Get-Service -Name $name -ErrorAction SilentlyContinue
        if (-not $service) {
            Write-Action "Skip service $name ($Reason): not installed" ([ConsoleColor]::DarkGray)
            continue
        }

        if ($service.Status -eq 'Stopped') {
            Write-Action "Skip service $name ($Reason): already stopped" ([ConsoleColor]::DarkGray)
            continue
        }

        $message = "Stop service $name - $Reason"
        if (-not $Apply) {
            Write-Action "DRY RUN: $message"
            continue
        }

        try {
            Stop-Service -Name $name -Force
            Write-Action "Stopped service $name" ([ConsoleColor]::Green)
        }
        catch {
                Write-Action "Failed to stop service ${name}: $($_.Exception.Message)" ([ConsoleColor]::Red)
        }
    }
}

function Invoke-DisableAdapters {
    param(
        [string[]]$Patterns,
        [string]$Reason
    )

    $adapters = @(Get-NetAdapter -ErrorAction SilentlyContinue | Where-Object {
        $adapter = $_
        $Patterns | Where-Object {
            $adapter.Name -like $_ -or $adapter.InterfaceDescription -like $_
        }
    })

    if ($adapters.Count -eq 0) {
        Write-Action "Skip adapter disable ($Reason): no matching adapters" ([ConsoleColor]::DarkGray)
        return
    }

    foreach ($adapter in $adapters) {
        if ($adapter.Status -eq 'Disabled') {
            Write-Action "Skip adapter $($adapter.Name) ($Reason): already disabled" ([ConsoleColor]::DarkGray)
            continue
        }

        $message = "Disable adapter $($adapter.Name) - $Reason"
        if (-not $Apply) {
            Write-Action "DRY RUN: $message"
            continue
        }

        try {
            Disable-NetAdapter -Name $adapter.Name -Confirm:$false
            Write-Action "Disabled adapter $($adapter.Name)" ([ConsoleColor]::Green)
        }
        catch {
            Write-Action "Failed to disable adapter $($adapter.Name): $($_.Exception.Message)" ([ConsoleColor]::Red)
        }
    }
}

function Show-TargetStatus {
    param(
        [string[]]$ProcessNames,
        [string[]]$ServiceNames
    )

    Write-Section 'Current process state'
    $runningProcesses = @(Get-Process -ErrorAction SilentlyContinue | Where-Object { $ProcessNames -contains $_.ProcessName } | Select-Object ProcessName, Id)
    if ($runningProcesses.Count -eq 0) {
        Write-Action 'No targeted processes are running.' ([ConsoleColor]::DarkGray)
    }
    else {
        $runningProcesses | Sort-Object ProcessName, Id | Format-Table -AutoSize
    }

    Write-Section 'Current service state'
    $services = @(Get-Service -ErrorAction SilentlyContinue | Where-Object { $ServiceNames -contains $_.Name } | Select-Object Name, Status)
    if ($services.Count -eq 0) {
        Write-Action 'No targeted services are installed.' ([ConsoleColor]::DarkGray)
    }
    else {
        $services | Sort-Object Name | Format-Table -AutoSize
    }
}

$processGroups = @(
    @{
        Name = 'VPN and VPN UI';
        Processes = @('NordVPN');
        Reason = 'remove VPN UI and session helpers from the gaming profile';
    },
    @{
        Name = 'Chat and social';
        Processes = @('Discord', 'Signal');
        Reason = 'reduce overlay hooks and background activity';
    },
    @{
        Name = 'NVIDIA overlay';
        Processes = @('NVIDIA Overlay');
        Reason = 'remove overlay and capture overhead';
    },
    @{
        Name = 'VR and mixed reality';
        Processes = @('MixedRealityLinkSvc', 'MixedRealityPairingHelper', 'OVRServer_x64', 'OVRRedir', 'OVRServiceLauncher', 'vorpControl');
        Reason = 'remove VR services when gaming on a normal display';
    }
)

if ($IncludeLaunchers) {
    $processGroups += @{
        Name = 'Game launchers';
        Processes = @('steam', 'steamwebhelper', 'EADesktop', 'EABackgroundService', 'EALocalHostSvc', 'EACefSubProcess');
        Reason = 'reduce launcher overhead after the game is already running or when not needed';
    }
}

if ($IncludeSearchIndex) {
    $processGroups += @{
        Name = 'Search indexing';
        Processes = @('SearchIndexer', 'SearchHost', 'SearchProtocolHost');
        Reason = 'reduce indexing activity during a gaming session';
    }
}

$serviceGroups = @()
if ($StopVpnServices) {
    $serviceGroups += @{
        Name = 'VPN services';
        Services = @('nordvpn-service', 'nordsec-threatprotection-service');
        Reason = 'fully remove NordVPN background services from the session';
    }
}

$allProcessNames = @($processGroups | ForEach-Object { $_.Processes } | Select-Object -Unique)
$allServiceNames = @($serviceGroups | ForEach-Object { $_.Services } | Select-Object -Unique)

Write-Host 'Windows gaming session cleanup' -ForegroundColor Green
if (-not $Apply) {
    Write-Action 'Dry run only. Re-run with -Apply to make changes.'
}

Show-TargetStatus -ProcessNames $allProcessNames -ServiceNames $allServiceNames

foreach ($group in $processGroups) {
    Write-Section $group.Name
    Invoke-StopProcesses -Names $group.Processes -Reason $group.Reason
}

foreach ($group in $serviceGroups) {
    Write-Section $group.Name
    Invoke-StopServices -Names $group.Services -Reason $group.Reason
}

if ($DisableVpnAdapter) {
    Write-Section 'VPN adapters'
    Invoke-DisableAdapters -Patterns @('*Nord*', '*OpenVPN*', '*TAP*') -Reason 'ensure the VPN network path is fully disabled for the session'
}

Write-Section 'Done'
if ($Apply) {
    Write-Action 'Gaming session cleanup finished.' ([ConsoleColor]::Green)
}
else {
    Write-Action 'No changes applied. Use -Apply to execute the cleanup.' ([ConsoleColor]::Yellow)
}