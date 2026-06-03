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

	// Get Code Name from CPU model
	info.CodeName = getCPUCodeName()

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

// getCPUCodeName reads CPU ProcessorId via WMI and returns the Intel Code Name
func getCPUCodeName() string {
	// Use PowerShell to get ProcessorId (same method as CPU-Z)
	script := `
$ErrorActionPreference = 'SilentlyContinue'
$cpu = Get-CimInstance Win32_Processor | Select-Object -First 1
if ($cpu.ProcessorId) {
    $id = $cpu.ProcessorId
    # Parse Family, Model, Stepping from ProcessorId
    # ProcessorId format: BFEBFBFF000C06C2 (hex, little-endian)
    # For Intel CPUs: EAX contains CPUID info
    # EAX[3:0] = Stepping, EAX[7:4] = Model, EAX[11:8] = Family
    # EAX[19:16] = Extended Model (when Family=6 or 15)
    # Full Model = (Extended Model << 4) + Model
    $eax = [Convert]::ToInt32($id.Substring(8,8), 16)
    $family = ($eax -shr 8) -band 0xF
    $model = ($eax -shr 4) -band 0xF
    $ext_model = ($eax -shr 16) -band 0xF
    if ($family -eq 6 -or $family -eq 15) {
        $model = ($ext_model -shl 4) -bor $model
    }
    Write-Output "Model:$model"
} else {
    Write-Output "N/A"
}
`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return "Unknown"
	}
	
	output := strings.TrimSpace(string(out))
	if output == "N/A" || output == "" {
		return "Unknown"
	}
	
	// Parse Model from output
	var model uint32
	if _, err := fmt.Sscanf(output, "Model:%d", &model); err != nil {
		return "Unknown"
	}
	
	// Look up Code Name
	if name, ok := intelCodeNames[model]; ok {
		return name
	}
	return fmt.Sprintf("Unknown (Model %d)", model)
}
