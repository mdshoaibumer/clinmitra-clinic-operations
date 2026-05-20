<#
.SYNOPSIS
    Development helper commands for ClinMitra Dental.
.EXAMPLE
    .\scripts\dev.ps1 test       # Run all Go tests
    .\scripts\dev.ps1 cover      # Run tests with coverage
    .\scripts\dev.ps1 lint       # Run linting checks
    .\scripts\dev.ps1 build      # Build production binary
    .\scripts\dev.ps1 dev        # Start dev mode
#>

param(
    [Parameter(Position=0)]
    [ValidateSet("test", "cover", "lint", "build", "dev", "clean")]
    [string]$Command = "test"
)

$ErrorActionPreference = "Stop"
$rootDir = Split-Path -Parent $PSScriptRoot

Push-Location $rootDir
try {
    switch ($Command) {
        "test" {
            Write-Host "Running tests..." -ForegroundColor Cyan
            go test -race ./...
        }
        "cover" {
            Write-Host "Running tests with coverage..." -ForegroundColor Cyan
            go test -race -coverprofile=coverage.out -covermode=atomic ./...
            go tool cover -func=coverage.out | Select-String "total:"
            Write-Host "`nFull report: go tool cover -html=coverage.out" -ForegroundColor Gray
        }
        "lint" {
            Write-Host "Running linting..." -ForegroundColor Cyan
            go vet ./...
            $unformatted = gofmt -l .
            if ($unformatted) {
                Write-Host "Unformatted files:" -ForegroundColor Red
                $unformatted | ForEach-Object { Write-Host "  $_" }
                exit 1
            }
            Write-Host "✓ All checks passed" -ForegroundColor Green
        }
        "build" {
            Write-Host "Building production binary..." -ForegroundColor Cyan
            wails build -clean
            Write-Host "✓ Build complete: build/bin/" -ForegroundColor Green
        }
        "dev" {
            Write-Host "Starting development mode..." -ForegroundColor Cyan
            wails dev
        }
        "clean" {
            Write-Host "Cleaning build artifacts..." -ForegroundColor Cyan
            Remove-Item -Recurse -Force "build/bin/*" -ErrorAction SilentlyContinue
            Remove-Item -Recurse -Force "frontend/dist" -ErrorAction SilentlyContinue
            Remove-Item "coverage.out" -ErrorAction SilentlyContinue
            Write-Host "✓ Cleaned" -ForegroundColor Green
        }
    }
} finally {
    Pop-Location
}
