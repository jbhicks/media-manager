# Windows Fyne/Go OpenGL Build Environment Setup
# This script installs MinGW-w64 and sets up your environment for building Fyne/OpenGL Go apps.
# Run this script as Administrator in PowerShell.

# Download MinGW-w64 installer
$mingwUrl = "https://github.com/brechtsanders/winlibs_mingw/releases/download/13.2.0-17.0.6-12.2.0-ucrt-r5/winlibs-x86_64-posix-seh-gcc-13.2.0-llvm-17.0.6-mingw-w64ucrt-12.2.0-r5.7z"
$mingwArchive = "$env:TEMP\mingw.7z"
$mingwDir = "C:\mingw-w64"

Write-Host "Downloading MinGW-w64..."
Invoke-WebRequest -Uri $mingwUrl -OutFile $mingwArchive

Write-Host "Extracting MinGW-w64..."
if (!(Test-Path $mingwDir)) { New-Item -ItemType Directory -Path $mingwDir }
# Requires 7-Zip installed and in PATH
7z x $mingwArchive -o$mingwDir | Out-Null

# Add MinGW-w64 to system PATH for current session
$mingwBin = Get-ChildItem -Path $mingwDir -Directory | Select-Object -First 1 | ForEach-Object { $_.FullName + "\bin" }
$env:Path = "$mingwBin;$env:Path"

# Optionally, add to user PATH permanently
# [Environment]::SetEnvironmentVariable("Path", "$mingwBin;" + [Environment]::GetEnvironmentVariable("Path", "User"), "User")

Write-Host "MinGW-w64 installed. Test with: gcc --version"

# Test OpenGL headers
Write-Host "Testing OpenGL headers..."
$testC = @"
#include <windows.h>
#include <GL/gl.h>
int main() { return 0; }
"@
$testFile = "$env:TEMP\testgl.c"
Set-Content -Path $testFile -Value $testC
& gcc $testFile -lopengl32 -o "$env:TEMP\testgl.exe"
if (Test-Path "$env:TEMP\testgl.exe") {
    Write-Host "OpenGL headers and linker found."
} else {
    Write-Host "OpenGL headers or linker missing!"
}

Write-Host "Setup complete. You can now build Fyne/Go OpenGL projects."
