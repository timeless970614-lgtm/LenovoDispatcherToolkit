//go:build windows

package backend

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// UninstallResult represents the result of uninstall operation
type UninstallResult struct {
	Success       bool   `json:"success"`
	Message       string `json:"message"`
	DriversRemoved int    `json:"driversRemoved"`
}

// UninstallDispatcher uninstalls the Lenovo Process Management dispatcher driver
func UninstallDispatcher() UninstallResult {
	result := UninstallResult{
		Success:       false,
		DriversRemoved: 0,
	}

	// Create a batch script to uninstall the driver
	batchScript := `@echo off
set substring_1=lnvprocessmanagement.inf
set counter1=0
set prev_line=""

pnputil /enum-drivers > drivers.txt

set "File2Read=drivers.txt"
If Not Exist "%File2Read%" (Goto :Error)

setlocal EnableExtensions EnableDelayedExpansion
for /f "delims=" %%a in ('Type "%File2Read%"') do (
 set /a count+=1
 set "Line[!count!]=%%a"
)

For /L %%i in (1,1,%Count%) do (
 echo "!Line[%%i]!" | findstr /C:"!substring_1!" 1>nul
 if errorlevel 1 (
 rem pattern not found
 ) ELSE (
 call :getOemNum "!Line[%%i]!" "!prev_line!"
 set /a counter1+=1
 )
 set "prev_line=!Line[%%i]!"
)

echo "%substring_1% found and removed: %counter1% times"
del drivers.txt 2>NUL
exit /b 0

:getOemNum
for /F "tokens=2 delims=:" %%a in ("%~2%") do ( 
 call :deleteOEM %%a
)
exit /b

:deleteOEM
set strDelete="remove %1"
echo %strDelete%
pnputil /delete-driver %1 /force /uninstall
exit /b

:ERROR 
ECHO Uninstall failed, refer to logFile.
echo The file "%File2Read%" does not exist !
del drivers.txt 2>NUL
exit /b 1
`

	// Create temp directory for the batch script
	tempDir := os.TempDir()
	batchPath := filepath.Join(tempDir, "uninstall_dispatcher.bat")

	// Write the batch script
	err := os.WriteFile(batchPath, []byte(batchScript), 0644)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to create uninstall script: %v", err)
		return result
	}
	defer os.Remove(batchPath)

	// Execute the batch script with admin privileges
	cmd := exec.Command("cmd", "/C", batchPath)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		result.Message = fmt.Sprintf("Uninstall script failed: %v", err)
		return result
	}

	// Parse output to count removed drivers
	outputStr := string(output)
	if strings.Contains(outputStr, "found and removed:") {
		// Extract count from output
		parts := strings.Split(outputStr, "found and removed:")
		if len(parts) > 1 {
			fmt.Sscanf(strings.TrimSpace(parts[1]), "%d", &result.DriversRemoved)
		}
	}

	result.Success = true
	if result.DriversRemoved > 0 {
		result.Message = fmt.Sprintf("Successfully uninstalled dispatcher. %d driver(s) removed.", result.DriversRemoved)
	} else {
		result.Message = "Uninstall completed. No dispatcher drivers found to remove."
	}

	return result
}

// UninstallDispatcherSimple uses pnputil directly to find and remove the driver
func UninstallDispatcherSimple() UninstallResult {
	result := UninstallResult{
		Success:       false,
		DriversRemoved: 0,
	}

	// Enumerate drivers and find lnvprocessmanagement.inf
	cmd := exec.Command("pnputil", "/enum-drivers")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	output, err := cmd.Output()
	if err != nil {
		result.Message = fmt.Sprintf("Failed to enumerate drivers: %v", err)
		return result
	}

	// Parse output to find OEM drivers for lnvprocessmanagement.inf
	lines := strings.Split(string(output), "\n")
	var oemDrivers []string
	var prevLine string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "lnvprocessmanagement.inf") {
			// Previous line contains the OEM driver name
			if prevLine != "" {
				// Extract OEM name from "Published Name: oemXX.inf"
				if strings.Contains(prevLine, "Published Name:") {
					parts := strings.Split(prevLine, ":")
					if len(parts) > 1 {
						oemName := strings.TrimSpace(parts[1])
						oemDrivers = append(oemDrivers, oemName)
					}
				}
			}
		}
		prevLine = line
	}

	if len(oemDrivers) == 0 {
		result.Success = true
		result.Message = "No dispatcher drivers found to uninstall."
		return result
	}

	// Delete each OEM driver
	for _, oem := range oemDrivers {
		cmd := exec.Command("pnputil", "/delete-driver", oem, "/force", "/uninstall")
		cmd.SysProcAttr = &syscall.SysProcAttr{
			HideWindow: true,
		}
		_, err := cmd.Output()
		if err == nil {
			result.DriversRemoved++
		}
	}

	result.Success = true
	result.Message = fmt.Sprintf("Successfully removed %d dispatcher driver(s).", result.DriversRemoved)
	return result
}