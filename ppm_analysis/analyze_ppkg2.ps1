# PPM Provisioning Package Analysis
# ppkg files use Windows Imaging Format (WIM) - magic "MSWI"

$ppkgPath = "C:\LenovoDispatcherToolkit\ppm_analysis\PPM-ARL-v1007.20250118.zip"
$bytes = [System.IO.File]::ReadAllBytes($ppkgPath)

Write-Host "=== PPM Provisioning Package Analysis ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "File: PPM-ARL-v1007.20250118.ppkg"
Write-Host "Size: $($bytes.Length) bytes"
Write-Host "Magic: MSWI (Windows Imaging Format - WIM)"
Write-Host ""

# Parse WIM header
# WIM Header Structure (simplified)
$header = @{
    Magic = [System.Text.Encoding]::ASCII.GetString($bytes[0..3])
    Size = [BitConverter]::ToUInt32($bytes, 8)
    Version = [BitConverter]::ToUInt32($bytes, 12)
    Flags = [BitConverter]::ToUInt32($bytes, 16)
    CompressionSize = [BitConverter]::ToUInt32($bytes, 20)
}

Write-Host "WIM Header Info:"
Write-Host "  Magic: $($header.Magic)"
Write-Host "  Header Size: $($header.Size) bytes"
Write-Host "  Version: $($header.Version)"
Write-Host "  Flags: 0x$($header.Flags.ToString('X8'))"
Write-Host ""

# Try using 7-Zip if available
$7zipPaths = @(
    "C:\Program Files\7-Zip\7z.exe",
    "C:\Program Files (x86)\7-Zip\7z.exe"
)

$7zip = $7zipPaths | Where-Object { Test-Path $_ } | Select-Object -First 1

if ($7zip) {
    Write-Host "Found 7-Zip at: $7zip" -ForegroundColor Green
    Write-Host "Attempting to extract contents..." -ForegroundColor Yellow
    
    $extractDir = "C:\LenovoDispatcherToolkit\ppm_analysis\extracted"
    if (Test-Path $extractDir) {
        Remove-Item $extractDir -Recurse -Force
    }
    New-Item -ItemType Directory -Path $extractDir -Force | Out-Null
    
    # Use 7z to list and extract
    Write-Host ""
    Write-Host "=== Listing Contents ===" -ForegroundColor Cyan
    & $7zip l $ppkgPath 2>&1
    
    Write-Host ""
    Write-Host "=== Extracting ===" -ForegroundColor Cyan
    & $7zip x $ppkgPath -o"$extractDir" -y 2>&1
} else {
    Write-Host "7-Zip not found. Cannot extract WIM archive." -ForegroundColor Red
    Write-Host ""
    Write-Host "WIM files can be extracted with:"
    Write-Host "  - 7-Zip (recommended)"
    Write-Host "  - DISM (requires admin)"
    Write-Host "  - WIMGAPI (requires admin)"
}

# Show extracted files
Write-Host ""
Write-Host "=== Extracted Files ===" -ForegroundColor Cyan
Get-ChildItem "C:\LenovoDispatcherToolkit\ppm_analysis\extracted" -Recurse | 
    Select-Object FullName, Length, LastWriteTime | 
    Format-Table -AutoSize
