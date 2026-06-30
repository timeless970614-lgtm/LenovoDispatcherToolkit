//go:build windows

package backend

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	versionDLL               = windows.NewLazySystemDLL("version.dll")
	procGetFileVersionInfoSizeW = versionDLL.NewProc("GetFileVersionInfoSizeW")
	procGetFileVersionInfoW     = versionDLL.NewProc("GetFileVersionInfoW")
	procVerQueryValueW          = versionDLL.NewProc("VerQueryValueW")
)

type vsFixedFileInfo struct {
	Signature        uint32
	StrucVersion     uint32
	FileVersionMS    uint32
	FileVersionLS    uint32
	ProductVersionMS uint32
	ProductVersionLS uint32
	FileFlagsMask    uint32
	FileFlags        uint32
	FileOS           uint32
	FileType         uint32
	FileSubtype      uint32
	FileDateMS       uint32
	FileDateLS       uint32
}

// Cached driver version — computed once on first call, driver version never changes at runtime
var (
	driverVersionOnce sync.Once
	driverVersionVal  string
)

// getDispatcherExeVersion reads the DriverVersion of "Lenovo Dispatcher"
// from the registry (fast path) with WMI/exe fallback.
// Result is cached after the first call.
func getDispatcherExeVersion() string {
	driverVersionOnce.Do(func() {
		// Fast path: read DriverVersion from Control\Class registry
		if v := getDriverVersionFromRegistry(); v != "" {
			driverVersionVal = v
			return
		}
		// Fallback: read FileVersion from service ImagePath
		driverVersionVal = getServiceExeVersion()
	})
	return driverVersionVal
}

// getDriverVersionFromRegistry reads the DriverVersion for "Lenovo Dispatcher"
// from HKLM\SYSTEM\CurrentControlSet\Control\Class\{guid}\nnnn.
// This is the same value shown in Device Manager, but read via registry (instant)
// instead of WMI/PowerShell (3-5 seconds).
func getDriverVersionFromRegistry() string {
	// Step 1: Find the Lenovo Dispatcher device in the Enum tree
	// Try ACPI\IDEA200C (known hardware ID for Lenovo Dispatcher)
	enumKey := `SYSTEM\CurrentControlSet\Enum\ACPI\IDEA200C`
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, enumKey, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return ""
	}
	subKeys, err := k.ReadSubKeyNames(-1)
	k.Close()
	if err != nil || len(subKeys) == 0 {
		return ""
	}

	// Step 2: For each instance, read the Driver value and then DriverVersion from Class
	for _, inst := range subKeys {
		instKey, err := registry.OpenKey(registry.LOCAL_MACHINE, enumKey+"\\"+inst, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		driverVal, _, err := instKey.GetStringValue("Driver")
		instKey.Close()
		if err != nil || driverVal == "" {
			continue
		}
		// driverVal looks like "{4d36e97d-e325-11ce-bfc1-08002be10318}\0074"
		classKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Control\Class\`+driverVal, registry.QUERY_VALUE)
		if err != nil {
			continue
		}
		ver, _, err := classKey.GetStringValue("DriverVersion")
		classKey.Close()
		if err == nil && ver != "" {
			return ver
		}
	}
	return ""
}

// (removed getPnPDriverVersion — replaced by getDriverVersionFromRegistry for speed)

// getServiceExeVersion reads the FileVersion from LNVDispatcherService.exe (legacy fallback).
func getServiceExeVersion() string {
	exePath := readRegString(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement`,
		"ImagePath",
		"",
	)
	if exePath == "" {
		return "N/A"
	}

	exePath = strings.ReplaceAll(exePath, "%SystemRoot%", `C:\Windows`)
	exePath = strings.ReplaceAll(exePath, "%SYSTEMROOT%", `C:\Windows`)

	pathPtr, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return "N/A"
	}

	size, _, _ := procGetFileVersionInfoSizeW.Call(
		uintptr(unsafe.Pointer(pathPtr)), 0,
	)
	if size == 0 {
		return "N/A"
	}

	buf := make([]byte, size)
	ret, _, _ := procGetFileVersionInfoW.Call(
		uintptr(unsafe.Pointer(pathPtr)), 0, size,
		uintptr(unsafe.Pointer(&buf[0])),
	)
	if ret == 0 {
		return "N/A"
	}

	var info *vsFixedFileInfo
	var infoLen uint32
	subBlock, _ := windows.UTF16PtrFromString(`\`)
	ret, _, _ = procVerQueryValueW.Call(
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(subBlock)),
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Pointer(&infoLen)),
	)
	if ret == 0 || infoLen == 0 {
		return "N/A"
	}

	major := (info.FileVersionMS >> 16) & 0xFFFF
	minor := info.FileVersionMS & 0xFFFF
	patch := (info.FileVersionLS >> 16) & 0xFFFF
	build := info.FileVersionLS & 0xFFFF

	return fmt.Sprintf("%d.%d.%d.%d", major, minor, patch, build)
}
