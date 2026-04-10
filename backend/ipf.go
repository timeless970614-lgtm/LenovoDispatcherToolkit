//go:build windows

package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

// ── DLL function pointers ──────────────────────────────────────────────────

type ipfFuncs struct {
	connect              uintptr
	disconnect           uintptr
	getVersion           uintptr
	getSystemPower       uintptr
	getCpuTemp           uintptr
	getPL1               uintptr
	getPL2               uintptr
	getPL4               uintptr
	getAllPL             uintptr
	setDllPath           uintptr
	readMSR              uintptr
	getEPP               uintptr
	getEPP1              uintptr
	getFrequencyLimit    uintptr
	getHeteroInc         uintptr
	getHeteroDec         uintptr
	getSoftParkLatency   uintptr
	getCurrentGear       uintptr
}

var (
	ipfDLL     *syscall.LazyDLL
	ipfFunc    ipfFuncs
	ipfOnce    sync.Once
	ipfOnceErr error
	ipfVersion int
	ipfMutex   sync.Mutex
)

// ── Init: load DLL and resolve symbols ────────────────────────────────────

// InitIPF loads ipf_wrapper.dll and resolves all function pointers.
// Safe to call multiple times; only runs once.
func InitIPF() error {
	ipfOnce.Do(func() {
		ipfOnceErr = ipfLoad()
	})
	return ipfOnceErr
}

// ipfLoad loads the DLL and resolves all function pointers.
func ipfLoad() error {
	exeDir := getExeDir()
	dllPath := filepath.Join(exeDir, "ipf_wrapper.dll")

	ipfDLL = syscall.NewLazyDLL(dllPath)

	resolve := func(name string) uintptr {
		p := ipfDLL.NewProc(name)
		if err := p.Find(); err != nil {
			return 0
		}
		return p.Addr()
	}

	ipfFunc.connect = resolve("IPF_Connect")
	ipfFunc.disconnect = resolve("IPF_Disconnect")
	ipfFunc.getVersion = resolve("IPF_GetVersion")
	ipfFunc.getSystemPower = resolve("IPF_GetSystemPower_mW")
	ipfFunc.getCpuTemp = resolve("IPF_GetCpuTemp_cK")
	ipfFunc.getPL1 = resolve("IPF_GetPL1_mW")
	ipfFunc.getPL2 = resolve("IPF_GetPL2_mW")
	ipfFunc.getPL4 = resolve("IPF_GetPL4_mW")
	ipfFunc.getAllPL = resolve("IPF_GetAllPL_mW")
	ipfFunc.setDllPath = resolve("IPF_SetDllPath")
	ipfFunc.readMSR = resolve("IPF_ReadMSR")
	ipfFunc.getEPP = resolve("IPF_GetEPP")
	ipfFunc.getEPP1 = resolve("IPF_GetEPP1")
	ipfFunc.getFrequencyLimit = resolve("IPF_GetFrequencyLimit_MHz")
	ipfFunc.getHeteroInc = resolve("IPF_GetHeteroInc")
	ipfFunc.getHeteroDec = resolve("IPF_GetHeteroDec")
	ipfFunc.getSoftParkLatency = resolve("IPF_GetSoftParkLatency")
	ipfFunc.getCurrentGear = resolve("IPF_GetCurrentGear")

	// Check critical functions
	if ipfFunc.connect == 0 || ipfFunc.getSystemPower == 0 || ipfFunc.getAllPL == 0 {
		return fmt.Errorf("ipf_wrapper.dll is missing one or more required functions; " +
			"make sure it is built from ipf_wrapper.cpp and ipf_wrapper.def")
	}

	// Connect and detect version
	rc, _, _ := syscall.Syscall(ipfFunc.connect, 0, 0, 0, 0)
	ipfVersion = int(rc)
	if ipfVersion == 0 {
		return fmt.Errorf("IPF_Connect failed; no IPF service available")
	}

	return nil
}

// getExeDir returns the directory containing the running executable.
func getExeDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	return filepath.Dir(exe)
}

// ── Public read API ─────────────────────────────────────────────────────────

// IPFInfo holds the raw IPF data read from LenovoIPF*.dll.
type IPFInfo struct {
	SystemPower_mW  uint32  `json:"systemPower_mW"`  // IPF_SystemPower
	PL1_mW          uint32  `json:"pl1_mW"`          // MMIO_PL1
	PL2_mW          uint32  `json:"pl2_mW"`          // MMIO_PL2
	PL4_mW          uint32  `json:"pl4_mW"`          // MMIO_PL4
	CpuTemp_cK      uint32  `json:"cpuTemp_cK"`      // raw centiKelvin
	CpuTemp_C       float64 `json:"cpuTemp_C"`       // Celsius
	Version         int     `json:"version"`          // 1=V1, 2=V2
	Connected       bool    `json:"connected"`
}

// PPMInfo holds MSR-based PPM values (EPP, Freq, Hetero, SoftPark).
type PPMInfo struct {
	EPP              uint32 `json:"epp"`               // EPP P-Core (MSR 0x1B0, 0-15)
	EPP1             uint32 `json:"epp1"`              // EPP E-Core
	FrequencyLimit   uint32 `json:"freqLimit"`         // Max freq from HWP (0-255 or MHz)
	HeteroInc        uint32 `json:"heteroInc"`         // Hetero Increase Threshold
	HeteroDec        uint32 `json:"heteroDec"`         // Hetero Decrease Threshold
	SoftParkLatency  uint32 `json:"softParkLatency"`   // Soft Park Latency
}

// ReadIPF reads all IPF values in one call.
func ReadIPF() IPFInfo {
	info := IPFInfo{Version: ipfVersion, Connected: ipfVersion > 0}

	ipfMutex.Lock()
	defer ipfMutex.Unlock()

	if ipfFunc.getSystemPower != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getSystemPower, 0, 0, 0, 0)
		info.SystemPower_mW = uint32(val)
	}

	if ipfFunc.getAllPL != 0 {
		var pl [3]uint32
		syscall.Syscall(
			ipfFunc.getAllPL, 1,
			uintptr(unsafe.Pointer(&pl[0])), 0, 0,
		)
		info.PL1_mW = pl[0]
		info.PL2_mW = pl[1]
		info.PL4_mW = pl[2]
	}

	if ipfFunc.getCpuTemp != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getCpuTemp, 0, 0, 0, 0)
		info.CpuTemp_cK = uint32(val)
		info.CpuTemp_C = float64(uint32(val))/100.0 - 273.15
	}

	return info
}

// ReadPPM reads all MSR-based PPM values (EPP, frequency, hetero, softpark).
func ReadPPM() PPMInfo {
	info := PPMInfo{}

	ipfMutex.Lock()
	defer ipfMutex.Unlock()

	readU32 := func(fn uintptr) uint32 {
		if fn == 0 {
			return 0
		}
		val, _, _ := syscall.Syscall(fn, 0, 0, 0, 0)
		return uint32(val)
	}

	info.EPP = readU32(ipfFunc.getEPP)
	info.EPP1 = readU32(ipfFunc.getEPP1)
	info.FrequencyLimit = readU32(ipfFunc.getFrequencyLimit)
	info.HeteroInc = readU32(ipfFunc.getHeteroInc)
	info.HeteroDec = readU32(ipfFunc.getHeteroDec)
	info.SoftParkLatency = readU32(ipfFunc.getSoftParkLatency)

	return info
}

// ReadIPFFromRegistry reads IPF values from the registry fallback.
// This is the original LenovoToolkit approach: read from
// HKLM\...\LenovoProcessManagement\Performance\PowerSlider.
// Use this when the IPF DLL is not available.
func ReadIPFFromRegistry() IPFInfo {
	info := IPFInfo{Connected: false}

	script := `
$ErrorActionPreference = 'SilentlyContinue'
$path = 'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider'
$ipfPower = (Get-ItemProperty $path -Name IPF_SystemPower -ErrorAction SilentlyContinue).IPF_SystemPower
$mmioPL1  = (Get-ItemProperty $path -Name MMIO_PL1  -ErrorAction SilentlyContinue).MMIO_PL1
$mmioPL2  = (Get-ItemProperty $path -Name MMIO_PL2  -ErrorAction SilentlyContinue).MMIO_PL2
$mmioPL4  = (Get-ItemProperty $path -Name MMIO_PL4  -ErrorAction SilentlyContinue).MMIO_PL4
$cpuTemp  = (Get-ItemProperty $path -Name CPU_Temperature -ErrorAction SilentlyContinue).CPU_Temperature
if ($null -eq $ipfPower) { $ipfPower = 0 }
if ($null -eq $mmioPL1)  { $mmioPL1  = 0 }
if ($null -eq $mmioPL2)  { $mmioPL2  = 0 }
if ($null -eq $mmioPL4)  { $mmioPL4  = 0 }
if ($null -eq $cpuTemp)  { $cpuTemp  = 0 }
Write-Output "$ipfPower|$mmioPL1|$mmioPL2|$mmioPL4|$cpuTemp"
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return info
	}
	line := strings.TrimSpace(string(out))
	parts := strings.Split(line, "|")
	if len(parts) >= 5 {
		fmt.Sscanf(parts[0], "%d", &info.SystemPower_mW)
		fmt.Sscanf(parts[1], "%d", &info.PL1_mW)
		fmt.Sscanf(parts[2], "%d", &info.PL2_mW)
		fmt.Sscanf(parts[3], "%d", &info.PL4_mW)
		fmt.Sscanf(parts[4], "%d", &info.CpuTemp_cK)
		info.CpuTemp_C = float64(info.CpuTemp_cK)/100.0 - 273.15
		info.Connected = true
	}
	return info
}

// CloseIPF disconnects from the IPF DLL.
func CloseIPF() {
	ipfMutex.Lock()
	defer ipfMutex.Unlock()
	if ipfFunc.disconnect != 0 {
		syscall.Syscall(ipfFunc.disconnect, 0, 0, 0, 0)
	}
	ipfFunc = ipfFuncs{}
	ipfVersion = 0
}

// GetIPFVersion returns the detected IPF version (1=V1, 2=V2, 0=not connected).
func GetIPFVersion() int {
	return ipfVersion
}

// GetCurrentGear returns the current EPOT/Gear level (0-9) from LenovoIPFV2.dll.
// This is the same API used by ML_Scenario: _IPFV2_CurrentGear().
// Returns -1 if not available (V1 DLL or not connected).
func GetCurrentGear() int32 {
	ipfMutex.Lock()
	defer ipfMutex.Unlock()
	if ipfFunc.getCurrentGear == 0 {
		return -1
	}
	gear, _, _ := syscall.Syscall(ipfFunc.getCurrentGear, 0, 0, 0, 0)
	return int32(gear)
}
