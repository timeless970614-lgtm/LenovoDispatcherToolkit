//go:build windows

package backend

import (
	"fmt"
	"strings"
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

// getDispatcherExeVersion reads the DriverVersion of "Lenovo Dispatcher"
// from Win32_PnPSignedDriver (matching Device Manager).
func getDispatcherExeVersion() string {
	// Try WMI first (what Device Manager shows)
	if v := getPnPDriverVersion("Lenovo Dispatcher"); v != "" {
		return v
	}
	// Fallback: read FileVersion from service ImagePath
	return getServiceExeVersion()
}

// getPnPDriverVersion queries Win32_PnPSignedDriver for a device's DriverVersion.
func getPnPDriverVersion(deviceName string) string {
	script := fmt.Sprintf(
		`Get-CimInstance Win32_PnPSignedDriver | Where-Object { $_.DeviceName -eq '%s' } | Select-Object -ExpandProperty DriverVersion`,
		deviceName,
	)
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return ""
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return ""
	}
	return v
}

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
