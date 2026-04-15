Get-WmiObject -Class Win32_PnPSignedDriver | Where-Object { 
    $_.DeviceName -like "*PPM Provisioning*"
} | Select-Object DeviceName, InfName, DriverVersion, DriverDate | ConvertTo-Json
