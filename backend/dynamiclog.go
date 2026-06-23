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
	dynamicLogRegistryPath = "SYSTEM\\CurrentControlSet\\Services\\LenovoProcessManagement\\Performance\\PowerSlider"
	dynamicLogValueName    = "Policy_DynamicPlxlog"

	// Windows Error Reporting LocalDumps paths
	werLocalDumpsBase      = "SOFTWARE\\Microsoft\\Windows\\Windows Error Reporting\\LocalDumps"
	dumpExe1               = "LNV_DES.exe"
	dumpExe2               = "LNVDispatcherService.exe"
	dumpFolderValueName    = "DumpFolder"
	dumpTypeValueName      = "DumpType"

	// DumpFolder = %PROGRAMDATA%\Lenovo\LenovoDispatcher (REG_EXPAND_SZ)
	dumpFolderData = "%PROGRAMDATA%\\Lenovo\\LenovoDispatcher"
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

// GetDynamicDumpStatus checks if WER LocalDumps is configured for both Dispatcher executables
func GetDynamicDumpStatus() bool {
	// Check if both exe dump entries exist with DumpType=2
	return isDumpConfigured(dumpExe1) && isDumpConfigured(dumpExe2)
}

// isDumpConfigured checks if WER LocalDumps is set for a specific exe
func isDumpConfigured(exeName string) bool {
	subKey := werLocalDumpsBase + "\\" + exeName
	value, ok := readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE, subKey, dumpTypeValueName)
	if !ok {
		return false
	}
	return value == 2
}

// EnableDynamicDump configures WER LocalDumps for both Dispatcher executables and restarts the service
func EnableDynamicDump() DynamicLogResult {
	result := DynamicLogResult{
		Success: false,
	}

	// Configure dump for both executables
	for _, exeName := range []string{dumpExe1, dumpExe2} {
		if err := configureDumpForExe(exeName); err != nil {
			result.Message = fmt.Sprintf("Failed to configure dump for %s: %v", exeName, err)
			return result
		}
	}

	// Restart the service
	restartErr := restartLenovoService()
	if restartErr != nil {
		result.Message = fmt.Sprintf("Dump configured, but failed to restart service: %v", restartErr)
		return result
	}

	result.Success = true
	result.Message = "Dispatcher dump enabled (WER LocalDumps → %PROGRAMDATA%\\Lenovo\\LenovoDispatcher, Full dump). Service restarted."
	return result
}

// configureDumpForExe creates the WER LocalDumps registry entries for one exe
func configureDumpForExe(exeName string) error {
	subKey := werLocalDumpsBase + "\\" + exeName

	// Create/open the key with write access
	subKeyPtr, err := syscall.UTF16PtrFromString(subKey)
	if err != nil {
		return fmt.Errorf("UTF16 path: %w", err)
	}

	var hKey windows.Handle
	disposition := uint32(0)
	ret, _, err := syscall.NewLazyDLL("advapi32.dll").NewProc("RegCreateKeyExW").Call(
		uintptr(windows.HKEY_LOCAL_MACHINE),
		uintptr(unsafe.Pointer(subKeyPtr)),
		0, 0, 0,
		uintptr(windows.KEY_WRITE),
		0,
		uintptr(unsafe.Pointer(&hKey)),
		uintptr(unsafe.Pointer(&disposition)),
	)
	if ret != 0 {
		return fmt.Errorf("RegCreateKeyEx: %w", err)
	}
	defer windows.RegCloseKey(hKey)

	// Set DumpFolder (REG_EXPAND_SZ)
	folderUTF16 := windows.StringToUTF16(dumpFolderData)
	folderDataLen := uint32(len(folderUTF16) * 2) // StringToUTF16 already includes null terminator

	ret, _, err = syscall.NewLazyDLL("advapi32.dll").NewProc("RegSetValueExW").Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(mustUTF16Ptr(dumpFolderValueName))),
		0,
		uintptr(windows.REG_EXPAND_SZ),
		uintptr(unsafe.Pointer(&folderUTF16[0])),
		uintptr(folderDataLen),
	)
	if ret != 0 {
		return fmt.Errorf("Set DumpFolder: %w", err)
	}

	// Set DumpType (REG_DWORD = 2, full dump)
	dumpType := uint32(2)
	ret, _, err = syscall.NewLazyDLL("advapi32.dll").NewProc("RegSetValueExW").Call(
		uintptr(hKey),
		uintptr(unsafe.Pointer(mustUTF16Ptr(dumpTypeValueName))),
		0,
		uintptr(windows.REG_DWORD),
		uintptr(unsafe.Pointer(&dumpType)),
		4,
	)
	if ret != 0 {
		return fmt.Errorf("Set DumpType: %w", err)
	}

	return nil
}

// mustUTF16Ptr converts a string to UTF16 pointer, panicking on error
func mustUTF16Ptr(s string) *uint16 {
	ptr, err := syscall.UTF16PtrFromString(s)
	if err != nil {
		panic(err)
	}
	return ptr
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
