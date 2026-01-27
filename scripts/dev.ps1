# Development script for media-manager on Windows
# This replaces 'make dev' with a Windows-native solution

# Add required tools to PATH
$env:Path = "C:\msys64\mingw64\bin;$env:USERPROFILE\go\bin;C:\Users\joshu\AppData\Local\Microsoft\WinGet\Packages\Gyan.FFmpeg_Microsoft.Winget.Source_8wekyb3d8bbwe\ffmpeg-8.0.1-full_build\bin;$env:Path"

Write-Host "Building clear-previews..." -ForegroundColor Green
$env:CGO_ENABLED = "1"
go build -o bin/clear-previews.exe ./cmd/clear-previews/main.go

if (-not (Test-Path "tmp")) {
    New-Item -ItemType Directory -Path "tmp" | Out-Null
}

Write-Host "Clearing previews..." -ForegroundColor Green
& bin/clear-previews.exe

Write-Host "Starting development mode with hot-reload..." -ForegroundColor Green
$env:CLEAR_DB_ON_START = "true"
# Run air through cmd.exe to avoid PowerShell 'start' alias issues
cmd /c "air"

