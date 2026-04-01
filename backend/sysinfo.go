//go:build windows

package backend

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// SystemInfo holds system information
type SystemInfo struct {
	CPUName       string  `json:"cpuName"`
	BIOSVersion   string  `json:"biosVersion"`
	OSCaption     string  `json:"osCaption"`
	OSVersion     string  `json:"osVersion"`
	TotalMemoryGB float64 `json:"totalMemoryGB"`
	MemoryType    string  `json:"memoryType"`
}

// memoryStatusEx mirrors MEMORYSTATUSEX from Windows API
type memoryStatusEx struct {
	dwLength                uint32
	dwMemoryLoad            uint32
	ullTotalPhys            uint64
	ullAvailPhys            uint64
	ullTotalPageFile        uint64
	ullAvailPageFile        uint64
	ullTotalVirtual         uint64
	ullAvailVirtual         uint64
	ullAvailExtendedVirtual uint64
}

var (
	kernel32              = windows.NewLazySystemDLL("kernel32.dll")
	procGlobalMemoryStatusEx = kernel32.NewProc("GlobalMemoryStatusEx")
)

// GetSystemInfo reads system information from registry + Windows API (no WMI)
func GetSystemInfo() (SystemInfo, error) {
	info := SystemInfo{}

	// CPU name
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\CentralProcessor\0`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		if v, _, e := k.GetStringValue("ProcessorNameString"); e == nil {
			info.CPUName = strings.TrimSpace(v)
		}
		k.Close()
	}
	if info.CPUName == "" {
		info.CPUName = "N/A"
	}

	// BIOS version
	k, err = registry.OpenKey(
		registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\BIOS`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		if v, _, e := k.GetStringValue("BIOSVersion"); e == nil {
			info.BIOSVersion = strings.TrimSpace(v)
		}
		k.Close()
	}
	if info.BIOSVersion == "" {
		info.BIOSVersion = "N/A"
	}

	// OS caption + version
	k, err = registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		product, _, _ := k.GetStringValue("ProductName")
		display, _, _ := k.GetStringValue("DisplayVersion")
		build, _, _ := k.GetStringValue("CurrentBuildNumber")
		k.Close()
		if product != "" {
			info.OSCaption = strings.TrimSpace(product)
			if display != "" {
				info.OSCaption += " " + strings.TrimSpace(display)
			}
		} else {
			info.OSCaption = "Windows"
		}
		if build != "" {
			info.OSVersion = fmt.Sprintf("Build %s", strings.TrimSpace(build))
		} else {
			info.OSVersion = "N/A"
		}
	} else {
		info.OSCaption = "N/A"
		info.OSVersion = "N/A"
	}

	// Total physical memory via GlobalMemoryStatusEx
	var ms memoryStatusEx
	ms.dwLength = uint32(unsafe.Sizeof(ms))
	ret, _, _ := procGlobalMemoryStatusEx.Call(uintptr(unsafe.Pointer(&ms)))
	if ret != 0 {
		info.TotalMemoryGB = float64(ms.ullTotalPhys) / (1024 * 1024 * 1024)
	}

	info.MemoryType = "LPDDR5 / DDR5"

	return info, nil
}
