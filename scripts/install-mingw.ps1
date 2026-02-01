# Windows MinGW-w64 and 7-Zip Automated Installer
# This script will:
# 1. Download and install 7-Zip if missing
# 2. Download and extract MinGW-w64
# 3. Add MinGW-w64 to the user PATH
# 4. Test gcc installation

$ErrorActionPreference = 'Stop'

function Install-7Zip {
    $sevenZipPath = "$env:ProgramFiles\7-Zip\7z.exe"
    if (!(Test-Path $sevenZipPath)) {
        Write-Host "7-Zip not found. Downloading and installing..."
        $url = "https://www.7-zip.org/a/7z2301-x64.exe"
        $installer = "$env:TEMP\7zsetup.exe"
        Invoke-WebRequest -Uri $url -OutFile $installer
        Start-Process -FilePath $installer -ArgumentList "/S" -Wait
        Remove-Item $installer
    } else {
        Write-Host "7-Zip already installed."
    }
}

function Install-MinGW {
    $mingwUrl = "https://github.com/brechtsanders/winlibs_mingw/releases/download/13.2.0-17.0.6-12.2.0-ucrt-r5/winlibs-x86_64-posix-seh-gcc-13.2.0-llvm-17.0.6-mingw-w64ucrt-12.2.0-r5.7z"
    $mingwArchive = "$env:TEMP\mingw.7z"
    $mingwDir = "C:\mingw-w64"
    if (!(Test-Path $mingwDir)) {
        Write-Host "Downloading MinGW-w64..."
        Invoke-WebRequest -Uri $mingwUrl -OutFile $mingwArchive
        Write-Host "Extracting MinGW-w64..."
        & "$env:ProgramFiles\7-Zip\7z.exe" x $mingwArchive -o$mingwDir | Out-Null
        Remove-Item $mingwArchive
    } else {
        Write-Host "MinGW-w64 already exists at $mingwDir."
    }
    # Find the bin directory
    $binDir = Get-ChildItem -Path $mingwDir -Directory | Select-Object -First 1 | ForEach-Object { $_.FullName + "\bin" }
    return $binDir
}

function Add-ToUserPath($binDir) {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($currentPath -notlike "*$binDir*") {
        Write-Host "Adding $binDir to user PATH..."
        [Environment]::SetEnvironmentVariable("Path", "$binDir;" + $currentPath, "User")
    } else {
        Write-Host "$binDir already in user PATH."
    }
}

function Test-GCC {
    Write-Host "Testing gcc..."
    $gcc = & gcc --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "gcc is installed and working."
    } else {
        Write-Host "gcc is NOT working. You may need to restart your terminal or log out/in for PATH changes to take effect."
    }
}

Install-7Zip
$binDir = Install-MinGW
Add-ToUserPath $binDir
$env:Path = "$binDir;" + $env:Path
Test-GCC
Write-Host "MinGW-w64 and 7-Zip setup complete. If gcc is not found, restart your terminal or log out/in."
