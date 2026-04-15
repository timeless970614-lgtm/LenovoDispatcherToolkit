Get-WmiObject -Class Win32_PnPSignedDriver | Where-Object { 
    $_.DeviceName -like "*PPM*" -or 
    $_.DeviceName -like "*Dynamic Tuning*" -or 
    $_.DeviceName -like "*Innovation Platform*" -or
    $_.DeviceName -like "*Processor Participant*"
} | Select-Object DeviceName, DriverVersion, DriverDate | ConvertTo-Json
