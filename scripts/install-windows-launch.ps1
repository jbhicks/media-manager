# Point Windows Search / Win+media at the just-installed GUI binary.
# Called from `make install` after the exe is copied to GOBIN.

param(
    [string]$ExePath
)

$ErrorActionPreference = "Stop"

if ([string]::IsNullOrWhiteSpace($ExePath)) {
    $installDir = $env:GOBIN
    if ([string]::IsNullOrWhiteSpace($installDir)) {
        if (-not [string]::IsNullOrWhiteSpace($env:GOPATH)) {
            $installDir = Join-Path $env:GOPATH "bin"
        } else {
            $installDir = Join-Path $env:USERPROFILE "go\bin"
        }
    }
    $ExePath = Join-Path $installDir "media-manager.exe"
}

if (-not (Test-Path -LiteralPath $ExePath)) {
    throw "Installed GUI not found at $ExePath"
}

$exe = (Get-Item -LiteralPath $ExePath).FullName
$workDir = Split-Path -Parent $exe

function Set-ShortcutTarget([string]$Path, [string]$Target, [string]$WorkingDirectory) {
    if (-not (Test-Path -LiteralPath $Path)) {
        Write-Host "Skip missing shortcut: $Path"
        return
    }
    $shell = New-Object -ComObject WScript.Shell
    $lnk = $shell.CreateShortcut($Path)
    $lnk.TargetPath = $Target
    $lnk.WorkingDirectory = $WorkingDirectory
    $lnk.IconLocation = "$Target,0"
    $lnk.Save()
    $check = $shell.CreateShortcut($Path)
    Write-Host "Shortcut $Path -> $($check.TargetPath)"
}

$desktop = Join-Path $env:USERPROFILE "Desktop\Media Manager.lnk"
$startMenu = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Media Manager\Media Manager.lnk"
Set-ShortcutTarget $desktop $exe $workDir
Set-ShortcutTarget $startMenu $exe $workDir

$appPaths = "HKCU:\Software\Microsoft\Windows\CurrentVersion\App Paths\media-manager.exe"
New-Item -Path $appPaths -Force | Out-Null
Set-ItemProperty -Path $appPaths -Name "(default)" -Value $exe
Set-ItemProperty -Path $appPaths -Name "Path" -Value $workDir
Write-Host "App Paths -> $exe"

$pf = "C:\Program Files (x86)\Media Manager\media-manager.exe"
if (Test-Path -LiteralPath (Split-Path -Parent $pf)) {
    try {
        Copy-Item -LiteralPath $exe -Destination $pf -Force
        Write-Host "Updated $pf"
    } catch {
        Write-Host "Could not update Program Files copy (need admin): $($_.Exception.Message)"
        Write-Host "Win+media uses the Start Menu shortcut, which now points at $exe"
    }
}

Write-Host "Win+media launch target: $exe"

$flag = Join-Path $env:TEMP "media-manager-relaunch.flag"
if (Test-Path -LiteralPath $flag) {
    Remove-Item -LiteralPath $flag -Force
    Start-Process -FilePath $exe -WorkingDirectory $workDir
    Write-Host "Relaunched $exe"
}
