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

// SystemInfo holds system information
type SystemInfo struct {
	CPUName       string  `json:"CPUName"`
	CodeName      string  `json:"CodeName"`
	BIOSVersion   string  `json:"BIOSVersion"`
	OSCaption     string  `json:"OSCaption"`
	OSVersion     string  `json:"OSVersion"`
	TotalMemoryGB float64 `json:"TotalMemoryGB"`
	MemoryType    string  `json:"MemoryType"`
}

// Intel CPU Model (Family 6) to Code Name mapping
// Model is extracted from ProcessorId: Family=6, Model=bits 4-11, Stepping=bits 0-3
var intelCodeNames = map[uint32]string{
	// Alder Lake
	151: "Alder Lake",
	154: "Alder Lake",
	// Raptor Lake
	183: "Raptor Lake",
	186: "Raptor Lake",
	191: "Raptor Lake",
	// Meteor Lake
	170: "Meteor Lake",
	172: "Meteor Lake",
	// Arrow Lake
	197: "Arrow Lake",
	198: "Arrow Lake",
	// Lunar Lake
	189: "Lunar Lake",
	// Panther Lake
	204: "Panther Lake",
	// Tiger Lake
	140: "Tiger Lake",
	141: "Tiger Lake",
	// Ice Lake
	126: "Ice Lake",
	// Comet Lake
	166: "Comet Lake",
	// Rocket Lake
	167: "Rocket Lake",
	// Kaby Lake
	142: "Kaby Lake",
	158: "Kaby Lake",
	// Skylake
	78:  "Skylake",
	94:  "Skylake",
	// Broadwell
	61:  "Broadwell",
	71:  "Broadwell",
	// Haswell
	60:  "Haswell",
	63:  "Haswell",
	69:  "Haswell",
	70:  "Haswell",
	// Ivy Bridge
	58:  "Ivy Bridge",
	// Sandy Bridge
	42:  "Sandy Bridge",
	45:  "Sandy Bridge",
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

// Cached values — computed once on first call, never change at runtime
var (
	cpuCodeNameOnce sync.Once
	cpuCodeNameVal  string
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

	// Get Code Name from CPU model (cached after first call — never changes at runtime)
	cpuCodeNameOnce.Do(func() {
		cpuCodeNameVal = getCPUCodeName()
	})
	info.CodeName = cpuCodeNameVal

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

		// Fix Windows 11 detection: registry ProductName may still show "Windows 10"
		// Windows 11 has build number >= 22000
		buildNum := 0
		if build != "" {
			fmt.Sscanf(strings.TrimSpace(build), "%d", &buildNum)
		}
		isWin11 := buildNum >= 22000

		if product != "" {
			product = strings.TrimSpace(product)
			// Replace "Windows 10" with "Windows 11" if build >= 22000
			if isWin11 && strings.Contains(product, "Windows 10") {
				product = strings.Replace(product, "Windows 10", "Windows 11", 1)
			}
			info.OSCaption = product
			if display != "" {
				info.OSCaption += " " + strings.TrimSpace(display)
			}
		} else {
			if isWin11 {
				info.OSCaption = "Windows 11"
			} else {
				info.OSCaption = "Windows 10"
			}
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

// getCPUCodeName reads CPU Identifier from registry and returns the Intel Code Name.
// Registry path: HKLM\HARDWARE\DESCRIPTION\System\CentralProcessor\0 → "Identifier"
// Value format: "Intel64 Family 6 Model 198 Stepping 2"
// This replaces the old PowerShell+WMI approach which took 1-3 seconds;
// registry read is ~0.1ms.
func getCPUCodeName() string {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\CentralProcessor\0`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return "Unknown"
	}
	defer k.Close()

	identifier, _, err := k.GetStringValue("Identifier")
	if err != nil || identifier == "" {
		return "Unknown"
	}

	// Parse "Intel64 Family 6 Model 198 Stepping 2"
	var family, model uint32
	_, err = fmt.Sscanf(identifier, "Intel64 Family %d Model %d", &family, &model)
	if err != nil {
		// Try alternate format without "Intel64" prefix
		_, err = fmt.Sscanf(identifier, "Family %d Model %d", &family, &model)
		if err != nil {
			return "Unknown"
		}
	}

	// For Intel CPUs (Family 6), the registry already stores the full model number
	// (Extended Model << 4 | Model), so no need to manually combine.
	if name, ok := intelCodeNames[model]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (Model %d)", model)
}
