# Python3 Alias Setup for PowerShell
# Create python3 alias pointing to py (Python 3.13)

# PowerShell profile path
$profilePath = "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
$profileDir = Split-Path $profilePath -Parent

# Create profile directory if it doesn't exist
if (-not (Test-Path $profileDir)) {
    New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
}

# Create profile file if it doesn't exist
if (-not (Test-Path $profilePath)) {
    New-Item -ItemType File -Path $profilePath -Force | Out-Null
}

# Check if python3 alias already exists
$content = Get-Content $profilePath -Raw
if ($content -match 'Set-Alias -Name python3') {
    Write-Host "[OK] python3 alias already exists in profile" -ForegroundColor Green
    Write-Host ""
    Write-Host "Current aliases:" -ForegroundColor Yellow
    Get-Alias python3 -ErrorAction SilentlyContinue
    Write-Host ""
    Write-Host "Restart PowerShell to apply" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 0
}

# Add python3 alias to profile
$aliasContent = @"

# Python3 alias - python3 points to Python 3.13 (via py launcher)
Set-Alias -Name python3 -Value py
"@

Add-Content -Path $profilePath -Value $aliasContent

Write-Host "[OK] python3 alias added to profile" -ForegroundColor Green
Write-Host ""
Write-Host "Profile location: $profilePath" -ForegroundColor Gray
Write-Host ""
Write-Host "Alias created: python3 -> py" -ForegroundColor Yellow
Write-Host ""
Write-Host "Usage after restart:" -ForegroundColor Cyan
Write-Host "  python3 --version" -ForegroundColor Gray
Write-Host "  python3 script.py" -ForegroundColor Gray
Write-Host "  python3 -m pip install <package>" -ForegroundColor Gray
Write-Host ""
Write-Host "[!] IMPORTANT: Restart PowerShell to apply changes" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
