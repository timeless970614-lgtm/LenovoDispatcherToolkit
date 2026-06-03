//go:build windows

package backend

import (
	"fmt"
	"os/exec"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// DynamicLogResult represents the result of enable log operation
type DynamicLogResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

const (
	dynamicLogRegistryPath  = "SYSTEM\\CurrentControlSet\\Services\\LenovoProcessManagement\\Performance\\PowerSlider"
	dynamicLogValueName     = "Policy_DynamicPlxlog"
)

// GetDynamicLogStatus checks if dynamic log is enabled
func GetDynamicLogStatus() bool {
	value, ok := readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE, dynamicLogRegistryPath, dynamicLogValueName)
	if !ok {
		return false
	}
	return value == 1
}

// EnableDynamicLog enables the dynamic log and restarts the service
func EnableDynamicLog() DynamicLogResult {
	result := DynamicLogResult{
		Success: false,
	}

	// Check if the value already exists and is 1
	currentValue, exists := readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE, dynamicLogRegistryPath, dynamicLogValueName)
	
	if exists && currentValue == 1 {
		// Already enabled, just restart service
		restartErr := restartLenovoService()
		if restartErr != nil {
			result.Message = fmt.Sprintf("Log already enabled, but failed to restart service: %v", restartErr)
			return result
		}
		result.Success = true
		result.Message = "Dynamic log already enabled. Service restarted successfully."
		return result
	}

	// Need to set the value to 1
	err := setDWordInRegistry(windows.HKEY_LOCAL_MACHINE, dynamicLogRegistryPath, dynamicLogValueName, 1)
	if err != nil {
		result.Message = fmt.Sprintf("Failed to set registry value: %v", err)
		return result
	}

	// Restart the service
	restartErr := restartLenovoService()
	if restartErr != nil {
		result.Message = fmt.Sprintf("Registry updated, but failed to restart service: %v", restartErr)
		return result
	}

	result.Success = true
	result.Message = "Dynamic log enabled and service restarted successfully."
	return result
}

// setDWordInRegistry sets a DWORD value in the registry
func setDWordInRegistry(hKeyRoot windows.Handle, subKeyPath, valueName string, value uint32) error {
	subKeyPtr, err := syscall.UTF16PtrFromString(subKeyPath)
	if err != nil {
		return err
	}

	var hKey windows.Handle
	err = windows.RegOpenKeyEx(hKeyRoot, subKeyPtr, 0, windows.KEY_WRITE, &hKey)
	if err != nil {
		return err
	}
	defer windows.RegCloseKey(hKey)

	valueNamePtr, err := syscall.UTF16PtrFromString(valueName)
	if err != nil {
		return err
	}

	// Use syscall to call RegSetValueEx
	ret, _, err := syscall.NewLazyDLL("advapi32.dll").NewProc("RegSetValueExW").Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(valueNamePtr)),
		0,
		uintptr(windows.REG_DWORD),
		uintptr(unsafe.Pointer(&value)),
		4,
	)
	if ret != 0 {
		return err
	}
	return nil
}

// restartLenovoService restarts the Lenovo Process Management service
func restartLenovoService() error {
	// Stop the service
	stopCmd := exec.Command("net", "stop", "LenovoProcessManagement")
	stopCmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	stopCmd.Run() // Ignore error if service wasn't running

	// Start the service
	startCmd := exec.Command("net", "start", "LenovoProcessManagement")
	startCmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	return startCmd.Run()
}
