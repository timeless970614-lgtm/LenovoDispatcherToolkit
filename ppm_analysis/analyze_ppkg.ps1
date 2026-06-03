# Read and analyze ppkg file header
$ppkgPath = "C:\LenovoDispatcherToolkit\ppm_analysis\PPM-ARL-v1007.20250118.zip"
$bytes = [System.IO.File]::ReadAllBytes($ppkgPath)

# Show first 64 bytes as hex
Write-Host "First 64 bytes (hex):"
$hex = [BitConverter]::ToString($bytes[0..63])
Write-Host $hex

# Try to detect format from magic bytes
Write-Host ""
Write-Host "Magic bytes analysis:"
$magic = [System.Text.Encoding]::ASCII.GetString($bytes[0..3])
Write-Host "Magic: $magic"

# Check if it's a CAB file (MSCF)
if ($magic -eq "MSCF") {
    Write-Host "Detected: CAB file format"
}
# Check if it's a ZIP file (PK)
elseif ($bytes[0] -eq 0x50 -and $bytes[1] -eq 0x4B) {
    Write-Host "Detected: ZIP file format"
}
else {
    Write-Host "Unknown format"
}

# Show file size
Write-Host ""
Write-Host "File size: $($bytes.Length) bytes"
