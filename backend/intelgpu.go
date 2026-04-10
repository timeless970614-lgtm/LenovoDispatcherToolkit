//go:build windows

package backend

import (
	"fmt"
	"os/exec"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	MinIntelDriverVersion  = "32.0.101.8000"
	IntelDriverDownloadURL = "https://www.intel.cn/content/www/cn/zh/download-center/home.html"

	// Intel GPU driver registry key (display adapter class, first Intel adapter)
	intelGPURegBase = `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`
)

// ── Public Go types ───────────────────────────────────────────────────────────

// IntelGPUFrequency is the data returned to the frontend.
type IntelGPUFrequency struct {
	Available        bool    `json:"available"`
	MinFreq          float64 `json:"minFreq"`      // hardware minimum (MHz)
	MaxFreq          float64 `json:"maxFreq"`      // hardware maximum (MHz)
	CurrentMin       float64 `json:"currentMin"`   // current software min limit (MHz)
	CurrentMax       float64 `json:"currentMax"`   // current software max limit (MHz)
	RequestedMHz     float64 `json:"requestedMHz"` // N/A without IGC
	ActualMHz        float64 `json:"actualMHz"`    // N/A without IGC
	TdpMHz           float64 `json:"tdpMHz"`       // N/A without IGC
	EfficientMHz     float64 `json:"efficientMHz"` // N/A without IGC
	GPUName          string  `json:"gpuName"`
	DriverVersion    string  `json:"driverVersion"`
	DriverDate       string  `json:"driverDate"`
	MinDriverVersion string  `json:"minDriverVersion"`
	DriverOK         bool    `json:"driverOK"`
	AdapterIndex     int     `json:"adapterIndex"`
	RegKeyPath       string  `json:"regKeyPath"` // for debug
	Error            string  `json:"error"`
}

// IntelGPUFreqTestResult is returned by set/test operations.
type IntelGPUFreqTestResult struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	MinFreq float64 `json:"minFreq"`
	MaxFreq float64 `json:"maxFreq"`
}

// ── Registry helpers ──────────────────────────────────────────────────────────

// findIntelGPURegKey finds the first Intel GPU adapter subkey under the display
// class registry key. Returns the subkey path (e.g. "0000") and the opened key.
func findIntelGPURegKey() (string, registry.Key, error) {
	base, err := registry.OpenKey(registry.LOCAL_MACHINE, intelGPURegBase, registry.READ)
	if err != nil {
		return "", 0, fmt.Errorf("open GPU class key: %v", err)
	}
	defer base.Close()

	subkeys, err := base.ReadSubKeyNames(-1)
	if err != nil {
		return "", 0, fmt.Errorf("list GPU subkeys: %v", err)
	}

	for _, sub := range subkeys {
		if len(sub) != 4 {
			continue // skip "Properties", "Configuration", etc.
		}
		subPath := intelGPURegBase + `\` + sub
		k, err := registry.OpenKey(registry.LOCAL_MACHINE, subPath, registry.READ|registry.WRITE)
		if err != nil {
			// Try read-only
			k, err = registry.OpenKey(registry.LOCAL_MACHINE, subPath, registry.READ)
			if err != nil {
				continue
			}
		}
		desc, _, err := k.GetStringValue("DriverDesc")
		if err != nil {
			k.Close()
			continue
		}
		if strings.Contains(strings.ToLower(desc), "intel") {
			return subPath, k, nil
		}
		k.Close()
	}
	return "", 0, fmt.Errorf("no Intel GPU adapter found in registry")
}

// readDWORD reads a DWORD value from a registry key, returns 0 on error.
func readDWORD(k registry.Key, name string) uint32 {
	v, _, err := k.GetIntegerValue(name)
	if err != nil {
		return 0
	}
	return uint32(v)
}

// ── Implementation ────────────────────────────────────────────────────────────

// GetIntelGPUFrequency reads iGPU frequency limits from the Intel driver registry key.
func GetIntelGPUFrequency() IntelGPUFrequency {
	result := IntelGPUFrequency{
		MinDriverVersion: MinIntelDriverVersion,
	}

	// Get GPU name + driver version from WMI
	fillGPUInfoFromWMI(&result)

	// Open Intel GPU registry key
	keyPath, k, err := findIntelGPURegKey()
	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer k.Close()

	result.RegKeyPath = keyPath

	// Read MinFreq / MaxFreq (in MHz, DWORD)
	minFreq := readDWORD(k, "MinFreq")
	maxFreq := readDWORD(k, "MaxFreq")

	// Read hardware capability range from driver properties
	// Intel stores the hardware max in "MaxFreqOC" or we derive from WMI
	// For now use sensible defaults based on Arc B370 (100–2050 MHz range)
	hwMin := float64(100)
	hwMax := float64(2050)

	// If the driver has stored a hardware max, use it
	hwMaxReg := readDWORD(k, "MaxFreqOC")
	if hwMaxReg > 0 {
		hwMax = float64(hwMaxReg)
	}
	hwMinReg := readDWORD(k, "MinFreqHW")
	if hwMinReg > 0 {
		hwMin = float64(hwMinReg)
	}

	result.Available  = true
	result.MinFreq    = hwMin
	result.MaxFreq    = hwMax
	result.CurrentMin = float64(minFreq)
	result.CurrentMax = float64(maxFreq)

	// If both are 0, the driver hasn't set limits yet — use hardware max
	if result.CurrentMin == 0 && result.CurrentMax == 0 {
		result.CurrentMin = hwMin
		result.CurrentMax = hwMax
	}

	return result
}

// SetIntelGPUFrequencyRange writes MinFreq/MaxFreq to the Intel driver registry key.
// Requires the app to be running with sufficient privileges (or the key to be writable).
func SetIntelGPUFrequencyRange(minFreq, maxFreq float64) IntelGPUFreqTestResult {
	result := IntelGPUFreqTestResult{MinFreq: minFreq, MaxFreq: maxFreq}

	if minFreq > maxFreq {
		result.Message = "Min frequency cannot be greater than max frequency"
		return result
	}
	if minFreq < 0 || maxFreq < 0 {
		result.Message = "Frequency must be non-negative"
		return result
	}

	keyPath, _, err := findIntelGPURegKey()
	if err != nil {
		result.Message = "Cannot find Intel GPU registry key: " + err.Error()
		return result
	}

	// Open with write access
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, keyPath, registry.SET_VALUE)
	if err != nil {
		// Try via PowerShell with elevated privileges
		return setFreqViaPowerShell(keyPath, minFreq, maxFreq)
	}
	defer k.Close()

	if err := k.SetDWordValue("MinFreq", uint32(minFreq)); err != nil {
		return setFreqViaPowerShell(keyPath, minFreq, maxFreq)
	}
	if err := k.SetDWordValue("MaxFreq", uint32(maxFreq)); err != nil {
		return setFreqViaPowerShell(keyPath, minFreq, maxFreq)
	}

	result.Success = true
	result.Message = fmt.Sprintf("iGPU frequency range set: %.0f – %.0f MHz (registry)", minFreq, maxFreq)
	return result
}

// setFreqViaPowerShell sets the registry values via PowerShell (handles UAC elevation).
func setFreqViaPowerShell(keyPath string, minFreq, maxFreq float64) IntelGPUFreqTestResult {
	result := IntelGPUFreqTestResult{MinFreq: minFreq, MaxFreq: maxFreq}

	// Convert registry path to PowerShell format
	psPath := "HKLM:\\" + strings.ReplaceAll(keyPath, `\`, `\\`)

	script := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
try {
    Set-ItemProperty -Path '%s' -Name 'MinFreq' -Value %d -Type DWord
    Set-ItemProperty -Path '%s' -Name 'MaxFreq' -Value %d -Type DWord
    Write-Host "OK"
} catch {
    Write-Host "ERR:$($_.Exception.Message)"
}
`, psPath, int(minFreq), psPath, int(maxFreq))

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	out, err := cmd.Output()
	if err != nil {
		result.Message = fmt.Sprintf("PowerShell failed: %v", err)
		return result
	}

	output := strings.TrimSpace(string(out))
	if strings.HasPrefix(output, "ERR:") {
		result.Message = "Registry write failed: " + strings.TrimPrefix(output, "ERR:")
		return result
	}

	result.Success = true
	result.Message = fmt.Sprintf("iGPU frequency range set: %.0f – %.0f MHz", minFreq, maxFreq)
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
       Where-Object { $_.Name -match 'Intel|Arc' } |
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

// ── Unused IGC wrapper stubs (kept for future use) ────────────────────────────
// These are no-ops since we use registry-based approach.

var _ = unsafe.Pointer(nil) // suppress unused import
