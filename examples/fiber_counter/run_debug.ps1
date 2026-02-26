# PowerShell script to run fiber_counter with debug logging
# Get script directory and change to it
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir

$env:TUI_DEBUG_ALL = "true"
$env:MINT_USE_FIBER = "true"

$logfile = "fiber_counter_debug_$(Get-Date -Format 'yyyyMMdd_HHmmss').log"

Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host "Running fiber_counter with debug logging" -ForegroundColor Cyan
Write-Host "Working directory: $scriptDir" -ForegroundColor Gray
Write-Host "Output to: $logfile" -ForegroundColor Yellow
Write-Host "Will exit after 10 seconds" -ForegroundColor Yellow
Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host ""

# Start the process
$process = Start-Process -FilePath "go.exe" -ArgumentList "run", "test_log.go" -NoNewWindow -RedirectStandardOutput "$logfile" -RedirectStandardError "$logfile.err" -PassThru

# Wait for 10 seconds
Start-Sleep -Seconds 10

# Kill the process
Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host "Process stopped after 10 seconds" -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Standard Output:" -ForegroundColor Yellow
if (Test-Path $logfile) {
    Get-Content $logfile | Select-Object -First 200
}
Write-Host ""
Write-Host "Standard Error (if any):" -ForegroundColor Yellow
if (Test-Path "$logfile.err") {
    Get-Content "$logfile.err" | Select-Object -First 100
}
