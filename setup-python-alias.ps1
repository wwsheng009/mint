# Python Alias Setup for PowerShell
# Add this to your PowerShell profile to create python -> py alias

# Create PowerShell profile if it doesn't exist
$profilePath = "$env:USERPROFILE\Documents\WindowsPowerShell\Microsoft.PowerShell_profile.ps1"
$profileDir = Split-Path $profilePath -Parent

if (-not (Test-Path $profileDir)) {
    New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
}

# Check if alias already exists
if (Test-Path $profilePath) {
    $content = Get-Content $profilePath -Raw
    if ($content -match 'Set-Alias -Name python -Value py') {
        Write-Host "[OK] Python alias already exists in profile" -ForegroundColor Green
        Write-Host ""
        Write-Host "Current aliases:" -ForegroundColor Yellow
        Get-Alias python -ErrorAction SilentlyContinue
        Write-Host ""
        Write-Host "Restart PowerShell to apply" -ForegroundColor Yellow
        Write-Host ""
        Write-Host "Press any key to exit..."
        $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
        exit 0
    }
}

# Add alias to profile
$aliasLine = @"

# Python alias - points to Python 3.13 launcher
Set-Alias -Name python -Value py
"@

Add-Content -Path $profilePath -Value $aliasLine

Write-Host "[OK] Python alias added to profile" -ForegroundColor Green
Write-Host ""
Write-Host "Profile location: $profilePath" -ForegroundColor Gray
Write-Host ""
Write-Host "Alias created: python -> py" -ForegroundColor Yellow
Write-Host ""
Write-Host "Usage after restart:" -ForegroundColor Cyan
Write-Host "  python --version" -ForegroundColor Gray
Write-Host "  python script.py" -ForegroundColor Gray
Write-Host "  python -m pip install <package>" -ForegroundColor Gray
Write-Host ""
Write-Host "[!] IMPORTANT: Restart PowerShell to apply changes" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press any key to exit..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
