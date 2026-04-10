//go:build windows

package backend

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

const (
	MinIntelDriverVersion  = "32.0.101.8000"
	IntelDriverDownloadURL = "https://www.intel.cn/content/www/cn/zh/download-center/home.html"
)

// ── Result codes from igc_wrapper.dll ────────────────────────────────────────
const (
	igcOK               = 0
	igcErrNotLoaded     = -1
	igcErrNoDevice      = -2
	igcErrNoFreqDomain  = -3
	igcErrApiFail       = -4
	igcErrNotInit       = -5
)

// ── C struct mirrors (must match igc_wrapper.h exactly) ──────────────────────

// IGC_FreqInfo mirrors igc_wrapper.h IGC_FreqInfo
type igcFreqInfo struct {
	MinFreqMHz     float64
	MaxFreqMHz     float64
	CurrentMinMHz  float64
	CurrentMaxMHz  float64
	RequestedMHz   float64
	TdpMHz         float64
	EfficientMHz   float64
	ActualMHz      float64
}

// IGC_AdapterInfo mirrors igc_wrapper.h IGC_AdapterInfo
type igcAdapterInfo struct {
	Name            [256]byte
	DriverVersion   [64]byte
	AdapterIndex    int32
	FreqDomainCount int32
}

// ── DLL loader (lazy, once) ───────────────────────────────────────────────────

var (
	igcWrapperDLL *syscall.LazyDLL
	igcFn         struct {
		init           uintptr
		close          uintptr
		getAdapterCount uintptr
		getAdapterInfo  uintptr
		getFreqInfo     uintptr
		setFreqRange    uintptr
		errorString     uintptr
	}
	igcOnce sync.Once
	igcErr  error
)

func igcLoad() {
	igcOnce.Do(func() {
		exeDir := getExeDir()
		dllPath := exeDir + `\igc_wrapper.dll`
		igcWrapperDLL = syscall.NewLazyDLL(dllPath)

		resolve := func(name string) uintptr {
			p := igcWrapperDLL.NewProc(name)
			if err := p.Find(); err != nil {
				igcErr = fmt.Errorf("igc_wrapper.dll: %s not found: %v", name, err)
				return 0
			}
			return p.Addr()
		}

		igcFn.init            = resolve("IGC_Init")
		igcFn.close           = resolve("IGC_Close")
		igcFn.getAdapterCount = resolve("IGC_GetAdapterCount")
		igcFn.getAdapterInfo  = resolve("IGC_GetAdapterInfo")
		igcFn.getFreqInfo     = resolve("IGC_GetFreqInfo")
		igcFn.setFreqRange    = resolve("IGC_SetFreqRange")
		igcFn.errorString     = resolve("IGC_ErrorString")
	})
}

// igcInit calls IGC_Init() and returns the result code.
func igcInit() int {
	igcLoad()
	if igcErr != nil || igcFn.init == 0 {
		return igcErrNotLoaded
	}
	ret, _, _ := syscall.Syscall(igcFn.init, 0, 0, 0, 0)
	return int(int32(ret))
}

// igcErrorString returns a human-readable string for a result code.
func igcErrorString(code int) string {
	if igcFn.errorString == 0 {
		return fmt.Sprintf("igc_wrapper not loaded (code %d)", code)
	}
	ret, _, _ := syscall.Syscall(igcFn.errorString, 1, uintptr(code), 0, 0)
	if ret == 0 {
		return fmt.Sprintf("unknown error %d", code)
	}
	return syscall.UTF16ToString((*[256]uint16)(unsafe.Pointer(ret))[:])
}

// igcErrorStringA reads a C string (char*) returned by IGC_ErrorString.
func igcErrorStringA(code int) string {
	if igcFn.errorString == 0 {
		return fmt.Sprintf("igc_wrapper not loaded (code %d)", code)
	}
	ret, _, _ := syscall.Syscall(igcFn.errorString, 1, uintptr(int32(code)), 0, 0)
	if ret == 0 {
		return fmt.Sprintf("unknown error %d", code)
	}
	// Read null-terminated C string
	ptr := (*[256]byte)(unsafe.Pointer(ret))
	n := 0
	for n < 256 && ptr[n] != 0 {
		n++
	}
	return string(ptr[:n])
}

// ── Public Go types ───────────────────────────────────────────────────────────

// IntelGPUFrequency is the data returned to the frontend.
type IntelGPUFrequency struct {
	Available        bool    `json:"available"`
	MinFreq          float64 `json:"minFreq"`
	MaxFreq          float64 `json:"maxFreq"`
	CurrentMin       float64 `json:"currentMin"`
	CurrentMax       float64 `json:"currentMax"`
	RequestedMHz     float64 `json:"requestedMHz"`
	ActualMHz        float64 `json:"actualMHz"`
	TdpMHz           float64 `json:"tdpMHz"`
	EfficientMHz     float64 `json:"efficientMHz"`
	GPUName          string  `json:"gpuName"`
	DriverVersion    string  `json:"driverVersion"`
	DriverDate       string  `json:"driverDate"`
	MinDriverVersion string  `json:"minDriverVersion"`
	DriverOK         bool    `json:"driverOK"`
	AdapterIndex     int     `json:"adapterIndex"`
	Error            string  `json:"error"`
}

// IntelGPUFreqTestResult is returned by set/test operations.
type IntelGPUFreqTestResult struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	MinFreq float64 `json:"minFreq"`
	MaxFreq float64 `json:"maxFreq"`
}

// ── Implementation ────────────────────────────────────────────────────────────

// GetIntelGPUFrequency queries the iGPU frequency info via igc_wrapper.dll.
// Falls back to WMI for GPU name / driver version if IGC is unavailable.
func GetIntelGPUFrequency() IntelGPUFrequency {
	result := IntelGPUFrequency{
		MinDriverVersion: MinIntelDriverVersion,
	}

	// Always populate GPU name + driver version from WMI (works without IGC)
	fillGPUInfoFromWMI(&result)

	// Try IGC wrapper
	rc := igcInit()
	if rc != igcOK {
		result.Error = igcErrorStringA(rc)
		return result
	}

	// Find the first Intel adapter
	adapterCount, _, _ := syscall.Syscall(igcFn.getAdapterCount, 0, 0, 0, 0)
	if int(adapterCount) == 0 {
		result.Error = "No Intel GPU adapter found by IGC"
		return result
	}

	// Use adapter 0 (first Intel GPU)
	adapterIdx := 0

	// Get adapter info (name)
	var ai igcAdapterInfo
	rc2, _, _ := syscall.Syscall(igcFn.getAdapterInfo, 2,
		uintptr(adapterIdx),
		uintptr(unsafe.Pointer(&ai)),
		0)
	if int(int32(rc2)) == igcOK {
		n := 0
		for n < len(ai.Name) && ai.Name[n] != 0 {
			n++
		}
		if n > 0 {
			result.GPUName = string(ai.Name[:n])
		}
	}

	// Get frequency info
	var fi igcFreqInfo
	rc3, _, _ := syscall.Syscall(igcFn.getFreqInfo, 2,
		uintptr(adapterIdx),
		uintptr(unsafe.Pointer(&fi)),
		0)
	if int(int32(rc3)) != igcOK {
		result.Error = igcErrorStringA(int(int32(rc3)))
		return result
	}

	result.Available    = true
	result.AdapterIndex = adapterIdx
	result.MinFreq      = fi.MinFreqMHz
	result.MaxFreq      = fi.MaxFreqMHz
	result.CurrentMin   = fi.CurrentMinMHz
	result.CurrentMax   = fi.CurrentMaxMHz
	result.RequestedMHz = fi.RequestedMHz
	result.ActualMHz    = fi.ActualMHz
	result.TdpMHz       = fi.TdpMHz
	result.EfficientMHz = fi.EfficientMHz

	return result
}

// SetIntelGPUFrequencyRange calls IGC_SetFreqRange via igc_wrapper.dll.
func SetIntelGPUFrequencyRange(minFreq, maxFreq float64) IntelGPUFreqTestResult {
	result := IntelGPUFreqTestResult{MinFreq: minFreq, MaxFreq: maxFreq}

	if minFreq > maxFreq {
		result.Message = "Min frequency cannot be greater than max frequency"
		return result
	}

	rc := igcInit()
	if rc != igcOK {
		result.Message = "IGC not available: " + igcErrorStringA(rc)
		return result
	}

	// syscall with two float64 args: pass as uintptr via unsafe
	minBits := *(*uintptr)(unsafe.Pointer(&minFreq))
	maxBits := *(*uintptr)(unsafe.Pointer(&maxFreq))

	rc2, _, _ := syscall.Syscall(igcFn.setFreqRange, 3,
		uintptr(0), // adapterIndex 0
		minBits,
		maxBits)

	if int(int32(rc2)) == igcOK {
		result.Success = true
		result.Message = fmt.Sprintf("iGPU frequency range set: %.0f – %.0f MHz", minFreq, maxFreq)
	} else {
		result.Message = fmt.Sprintf("IGC_SetFreqRange failed: %s", igcErrorStringA(int(int32(rc2))))
	}
	return result
}

// TestIntelGPUFrequency runs a named test scenario.
func TestIntelGPUFrequency(testType string) IntelGPUFreqTestResult {
	freq := GetIntelGPUFrequency()
	if !freq.Available {
		return IntelGPUFreqTestResult{
			Success: false,
			Message: "Intel GPU not available: " + freq.Error,
		}
	}
	switch testType {
	case "min":
		return SetIntelGPUFrequencyRange(freq.MinFreq, freq.MinFreq)
	case "max":
		return SetIntelGPUFrequencyRange(freq.MaxFreq, freq.MaxFreq)
	case "dynamic":
		return SetIntelGPUFrequencyRange(freq.MinFreq, freq.MaxFreq)
	default:
		return IntelGPUFreqTestResult{Success: false, Message: "Unknown test type: " + testType}
	}
}

// GetIntelDriverDownloadURL returns the Intel driver download URL.
func GetIntelDriverDownloadURL() string { return IntelDriverDownloadURL }

// ── WMI fallback ──────────────────────────────────────────────────────────────

func fillGPUInfoFromWMI(result *IntelGPUFrequency) {
	script := `
$ErrorActionPreference = 'SilentlyContinue'
$gpu = Get-WmiObject Win32_VideoController |
       Where-Object { $_.Name -match 'Intel.*Graphics|Intel.*UHD|Intel.*Iris|Intel.*Arc' } |
       Select-Object -First 1
if ($gpu) {
    Write-Host "NAME:$($gpu.Name)"
    Write-Host "DRIVER:$($gpu.DriverVersion)"
    $d = $gpu.DriverDate
    if ($d -and $d.Length -ge 8) {
        Write-Host "DATE:$($d.Substring(0,4))-$($d.Substring(4,2))-$($d.Substring(6,2))"
    }
} else {
    Write-Host "NA:No Intel GPU found"
}
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return
	}
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimRight(line, "\r")
		switch {
		case strings.HasPrefix(line, "NAME:"):
			result.GPUName = strings.TrimPrefix(line, "NAME:")
		case strings.HasPrefix(line, "DRIVER:"):
			result.DriverVersion = strings.TrimPrefix(line, "DRIVER:")
			result.DriverOK = compareDriverVersion(result.DriverVersion, MinIntelDriverVersion) >= 0
		case strings.HasPrefix(line, "DATE:"):
			result.DriverDate = strings.TrimPrefix(line, "DATE:")
		}
	}
}

// compareDriverVersion compares two "A.B.C.D" version strings.
// Returns 1 if v1 > v2, 0 if equal, -1 if v1 < v2.
func compareDriverVersion(v1, v2 string) int {
	parts1 := strings.Split(v1, ".")
	parts2 := strings.Split(v2, ".")
	maxLen := len(parts1)
	if len(parts2) > maxLen {
		maxLen = len(parts2)
	}
	for i := 0; i < maxLen; i++ {
		var n1, n2 int
		if i < len(parts1) {
			fmt.Sscanf(parts1[i], "%d", &n1)
		}
		if i < len(parts2) {
			fmt.Sscanf(parts2[i], "%d", &n2)
		}
		if n1 > n2 {
			return 1
		}
		if n1 < n2 {
			return -1
		}
	}
	return 0
}
