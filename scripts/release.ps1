<#
.SYNOPSIS
    Bump version and create a release tag.
.DESCRIPTION
    Updates version in config.go and wails.json, commits, and creates a git tag.
    Push with --tags to trigger the release pipeline.
.PARAMETER Version
    The new version number (e.g., 1.2.0)
.EXAMPLE
    .\scripts\release.ps1 -Version 1.2.0
#>

param(
    [Parameter(Mandatory=$true)]
    [ValidatePattern('^\d+\.\d+\.\d+$')]
    [string]$Version
)

$ErrorActionPreference = "Stop"
$rootDir = Split-Path -Parent $PSScriptRoot

Write-Host "Preparing release v$Version" -ForegroundColor Cyan

# 1. Update internal/config/config.go
$configPath = Join-Path $rootDir "internal\config\config.go"
$configContent = Get-Content $configPath -Raw
$configContent = $configContent -replace 'Version:\s*"[^"]*"', ('Version:          "' + $Version + '"')
Set-Content $configPath $configContent -NoNewline
Write-Host "  Updated config.go" -ForegroundColor Green

# 2. Update wails.json
$wailsPath = Join-Path $rootDir "wails.json"
$wailsJson = Get-Content $wailsPath -Raw
$wailsJson = $wailsJson -replace '"productVersion":\s*"[^"]*"', ('"productVersion": "' + $Version + '"')
Set-Content $wailsPath $wailsJson -NoNewline
Write-Host "  Updated wails.json" -ForegroundColor Green

# 3. Git operations
Push-Location $rootDir
try {
    git add internal/config/config.go wails.json
    git commit -m "chore: bump version to $Version"
    git tag "v$Version"
    Write-Host ""
    Write-Host "Version bumped to $Version" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Yellow
    Write-Host "  git push origin main --tags" -ForegroundColor White
    Write-Host ""
    Write-Host "This will trigger the Release workflow on GitHub Actions." -ForegroundColor Gray
} finally {
    Pop-Location
}
