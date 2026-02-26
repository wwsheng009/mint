# PowerShell script to run fiber_counter with ONE SPACE to reproduce the bug
# Get script directory and change to it
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location -Path $scriptDir

$env:TUI_DEBUG_ALL = "true"
$env:MINT_USE_FIBER = "true"

$logfile = "fiber_counter_one_space_debug_$(Get-Date -Format 'yyyyMMdd_HHmmss').log"

Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host "Running fiber_counter with ONE SPACE (repro bug)" -ForegroundColor Red
Write-Host "Working directory: $scriptDir" -ForegroundColor Gray
Write-Host "Output to: $logfile" -ForegroundColor Yellow
Write-Host "Will exit after 10 seconds" -ForegroundColor Yellow
Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host ""

# Start the process
$process = Start-Process -FilePath "go.exe" -ArgumentList "run", "main_one_space.go" -NoNewWindow -RedirectStandardOutput "$logfile" -RedirectStandardError "$logfile.err" -PassThru

# Wait for 10 seconds
Start-Sleep -Seconds 10

# Kill the process
Stop-Process -Id $process.Id -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host "Process stopped after 10 seconds" -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Standard Output (final screen):" -ForegroundColor Yellow
if (Test-Path $logfile) {
    Get-Content $logfile | Select-Object -First 5
}
Write-Host ""
Write-Host "Log file location:" -ForegroundColor Gray
Write-Host "  $logfile" -ForegroundColor White
Write-Host "  $logfile.err" -ForegroundColor White
