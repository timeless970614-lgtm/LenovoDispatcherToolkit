# Build script for MSR Test Application
# Based on: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx

param(
    [switch]$Release,
    [switch]$Clean,
    [switch]$Run
)

$ErrorActionPreference = "Stop"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  MSR Test Application Build Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

$ProjectRoot = "C:\LenovoDispatcher\MSR"
$ProjectFile = Join-Path $ProjectRoot "MSR_Test_App.csproj"

# Change to project directory
Set-Location $ProjectRoot

# Check if .NET SDK is installed
Write-Host "Checking .NET SDK..." -ForegroundColor Yellow
$dotnetVersion = & dotnet --version 2>$null
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: .NET SDK not found. Please install .NET 8.0 SDK." -ForegroundColor Red
    exit 1
}
Write-Host "  .NET SDK Version: $dotnetVersion" -ForegroundColor Green

# Clean if requested
if ($Clean) {
    Write-Host ""
    Write-Host "Cleaning build artifacts..." -ForegroundColor Yellow
    & dotnet clean -c $(if($Release){"Release"}else{"Debug"}) 2>$null
    $binPath = Join-Path $ProjectRoot "bin"
    $objPath = Join-Path $ProjectRoot "obj"
    if (Test-Path $binPath) { Remove-Item $binPath -Recurse -Force }
    if (Test-Path $objPath) { Remove-Item $objPath -Recurse -Force }
    Write-Host "  Clean complete." -ForegroundColor Green
}

# Restore packages
Write-Host ""
Write-Host "Restoring NuGet packages..." -ForegroundColor Yellow
& dotnet restore $ProjectFile
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Package restore failed." -ForegroundColor Red
    exit 1
}
Write-Host "  Restore complete." -ForegroundColor Green

# Build
Write-Host ""
Write-Host "Building..." -ForegroundColor Yellow
$config = if($Release){"Release"}else{"Debug"}
& dotnet build $ProjectFile -c $config --no-restore
if ($LASTEXITCODE -ne 0) {
    Write-Host "ERROR: Build failed." -ForegroundColor Red
    exit 1
}
Write-Host "  Build complete." -ForegroundColor Green

# Output location
$outputDir = Join-Path $ProjectRoot "bin\$config\net8.0"
Write-Host ""
Write-Host "Output:" -ForegroundColor Cyan
Write-Host "  $outputDir" -ForegroundColor Gray

# List output files
if (Test-Path $outputDir) {
    Write-Host ""
    Write-Host "Built files:" -ForegroundColor Green
    Get-ChildItem $outputDir -File | ForEach-Object {
        Write-Host "  - $($_.Name) ($([Math]::Round($_.Length/1KB, 1)) KB)" -ForegroundColor White
    }
}

# Run if requested
if ($Run) {
    Write-Host ""
    Write-Host "Running..." -ForegroundColor Yellow
    Write-Host ""
    & dotnet run -c $config --no-build
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Build script completed" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
