# Stop the Media Manager GUI so `make install` can overwrite the exe.
# Does not stop media-manager-service. Does not kill explorer.

$ErrorActionPreference = "Stop"

$flag = Join-Path $env:TEMP "media-manager-relaunch.flag"
if (Test-Path -LiteralPath $flag) {
    Remove-Item -LiteralPath $flag -Force
}

$procs = @(Get-Process -Name "media-manager" -ErrorAction SilentlyContinue)
if ($procs.Count -eq 0) {
    Write-Host "Media Manager was not running"
    return
}

Set-Content -LiteralPath $flag -Value "relaunch" -Encoding ASCII

$pids = @($procs | Select-Object -ExpandProperty Id)
Write-Host ("Stopping media-manager PIDs: " + ($pids -join ", "))
Stop-Process -Id $pids -Force

$deadline = (Get-Date).AddSeconds(15)
while ((Get-Date) -lt $deadline) {
    $still = @(Get-Process -Name "media-manager" -ErrorAction SilentlyContinue)
    if ($still.Count -eq 0) {
        Write-Host "Media Manager stopped; it will relaunch after install"
        return
    }
    Start-Sleep -Milliseconds 200
}

throw "media-manager is still running after 15s; cannot overwrite locked exe"
