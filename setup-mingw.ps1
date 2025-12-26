# Setup MinGW-w64 for Fyne development on Windows
Write-Host "Installing MinGW-w64 toolchain..." -ForegroundColor Green

# Install MinGW-w64 gcc
& C:\msys64\usr\bin\bash.exe -lc "pacman -S --noconfirm mingw-w64-x86_64-gcc mingw-w64-x86_64-pkg-config"

# Add to PATH permanently
$mingwPath = "C:\msys64\mingw64\bin"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

if ($currentPath -notlike "*$mingwPath*") {
    Write-Host "Adding MinGW to PATH..." -ForegroundColor Green
    [Environment]::SetEnvironmentVariable("Path", "$currentPath;$mingwPath", "User")
    $env:Path = "$env:Path;$mingwPath"
    Write-Host "MinGW added to PATH" -ForegroundColor Green
} else {
    Write-Host "MinGW already in PATH" -ForegroundColor Yellow
}

# Verify installation
Write-Host "`nVerifying installation..." -ForegroundColor Green
& gcc --version

Write-Host "`nSetup complete! You may need to restart your terminal." -ForegroundColor Green
Write-Host "Run 'make build' to build your project." -ForegroundColor Green
