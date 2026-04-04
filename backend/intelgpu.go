//go:build windows

package backend

import (
	"fmt"
	"strings"
	"syscall"
)

const (
	// Minimum required Intel GPU driver version for IGC API
	// Format: XX.X.XXX.XXXX (e.g., 32.0.101.8626)
	MinIntelDriverVersion = "32.0.101.8000"
	IntelDriverDownloadURL = "https://www.intel.cn/content/www/cn/zh/download-center/home.html"
)

// IntelGPUFrequency represents iGPU frequency control capabilities
type IntelGPUFrequency struct {
	Available       bool   `json:"available"`
	MinFreq         uint32 `json:"minFreq"`
	MaxFreq         uint32 `json:"maxFreq"`
	CurrentMin      uint32 `json:"currentMin"`
	CurrentMax      uint32 `json:"currentMax"`
	Step            uint32 `json:"step"`
	GPUName         string `json:"gpuName"`
	DriverVersion   string `json:"driverVersion"`
	DriverDate      string `json:"driverDate"`
	MinDriverVersion string `json:"minDriverVersion"`
	DriverOK        bool   `json:"driverOK"`
	Error           string `json:"error"`
}

// IntelGPUFreqTestResult represents a test result
type IntelGPUFreqTestResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	MinFreq uint32 `json:"minFreq"`
	MaxFreq uint32 `json:"maxFreq"`
}

var igcDll uintptr

// InitIGCLibrary initializes Intel GPU Control Library
func InitIGCLibrary() error {
	dll, err := syscall.LoadLibrary("igc.dll")
	if err != nil {
		dll, err = syscall.LoadLibrary("ControlLib.dll")
		if err != nil {
			return fmt.Errorf("Intel GPU Control Library not found: %v", err)
		}
	}
	igcDll = uintptr(dll)
	return nil
}

// GetIntelGPUFrequency gets current iGPU frequency range
func GetIntelGPUFrequency() IntelGPUFrequency {
	result := IntelGPUFrequency{
		Available:        false,
		MinDriverVersion: MinIntelDriverVersion,
	}

	// Try to load IGC library
	if igcDll == 0 {
		if err := InitIGCLibrary(); err != nil {
			result.Error = err.Error()
			return result
		}
	}

	// Use PowerShell to query GPU info via WMI as fallback
	script := `
$ErrorActionPreference = 'SilentlyContinue'
try {
    $gpu = Get-WmiObject Win32_VideoController | Where-Object { $_.Name -match 'Intel.*Graphics|Intel.*UHD|Intel.*Iris|Intel.*Arc' } | Select-Object -First 1
    if ($gpu) {
        $name = $gpu.Name
        $driver = $gpu.DriverVersion
        $date = $gpu.DriverDate
        Write-Host "NAME:$name"
        Write-Host "DRIVER:$driver"
        Write-Host "DATE:$date"
    } else {
        Write-Host "NA:No Intel GPU found"
    }
} catch {
    Write-Host "NA:$($_.Exception.Message)"
}
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		result.Error = "Failed to query GPU: " + err.Error()
		return result
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "NAME:") {
			result.GPUName = strings.TrimPrefix(line, "NAME:")
			result.Available = true
		}
		if strings.HasPrefix(line, "DRIVER:") {
			result.DriverVersion = strings.TrimPrefix(line, "DRIVER:")
		}
		if strings.HasPrefix(line, "DATE:") {
			dateStr := strings.TrimPrefix(line, "DATE:")
			// Parse date like "20260311000000.000000-000"
			if len(dateStr) >= 8 {
				result.DriverDate = dateStr[0:4] + "-" + dateStr[4:6] + "-" + dateStr[6:8]
			}
		}
		if strings.HasPrefix(line, "NA:") {
			result.Error = strings.TrimPrefix(line, "NA:")
			return result
		}
	}

	// Check driver version
	result.DriverOK = compareDriverVersion(result.DriverVersion, MinIntelDriverVersion) >= 0

	// Read frequency from registry (Intel GPU stores freq settings here)
	freqScript := `
$ErrorActionPreference = 'SilentlyContinue'
# Try to read from Intel GPU registry
$keys = @(
    'HKLM:\SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}',
    'HKLM:\SOFTWARE\Intel\Display'
)
foreach ($k in $keys) {
    $subkeys = Get-ChildItem $k -ErrorAction SilentlyContinue
    foreach ($sk in $subkeys) {
        $props = Get-ItemProperty $sk.PSPath -ErrorAction SilentlyContinue
        if ($props.MinFreq -or $props.MaxFreq) {
            Write-Host "MIN:$($props.MinFreq)"
            Write-Host "MAX:$($props.MaxFreq)"
            return
        }
    }
}
# Default frequency range for Intel iGPU (typical values)
Write-Host "MIN:300"
Write-Host "MAX:1500"
Write-Host "STEP:50"
`
	cmd = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", freqScript)
	out, _ = cmd.Output()
	lines = strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MIN:") {
			fmt.Sscanf(line, "MIN:%d", &result.MinFreq)
			result.CurrentMin = result.MinFreq
		}
		if strings.HasPrefix(line, "MAX:") {
			fmt.Sscanf(line, "MAX:%d", &result.MaxFreq)
			result.CurrentMax = result.MaxFreq
		}
		if strings.HasPrefix(line, "STEP:") {
			fmt.Sscanf(line, "STEP:%d", &result.Step)
		}
	}

	if result.Step == 0 {
		result.Step = 50 // Default step
	}

	return result
}

// SetIntelGPUFrequencyRange sets the GPU frequency range
func SetIntelGPUFrequencyRange(minFreq, maxFreq uint32) IntelGPUFreqTestResult {
	result := IntelGPUFreqTestResult{
		Success:  false,
		MinFreq:  minFreq,
		MaxFreq:  maxFreq,
	}

	// Validate
	if minFreq > maxFreq {
		result.Message = "Min frequency cannot be greater than max frequency"
		return result
	}

	// Try to set via registry (Intel GPU control)
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
try {
    # Write to Intel GPU registry
    $keys = Get-ChildItem 'HKLM:\SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}' -ErrorAction SilentlyContinue
    foreach ($k in $keys) {
        $props = Get-ItemProperty $k.PSPath -ErrorAction SilentlyContinue
        if ($props.DriverDesc -match 'Intel') {
            Set-ItemProperty -Path $k.PSPath -Name 'MinFreq' -Value %d -Type DWord -Force -ErrorAction SilentlyContinue
            Set-ItemProperty -Path $k.PSPath -Name 'MaxFreq' -Value %d -Type DWord -Force -ErrorAction SilentlyContinue
            Write-Host "OK"
            return
        }
    }
    Write-Host "NA:No Intel GPU registry key found"
} catch {
    Write-Host "NA:$($_.Exception.Message)"
}
`, minFreq, maxFreq)

	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		result.Message = "Failed to set frequency: " + err.Error()
		return result
	}

	s := strings.TrimSpace(string(out))
	if s == "OK" {
		result.Success = true
		result.Message = fmt.Sprintf("Frequency range set to %d-%d MHz", minFreq, maxFreq)
	} else {
		result.Message = s
	}

	return result
}

// TestIntelGPUFrequency tests frequency control functionality
func TestIntelGPUFrequency(testType string) IntelGPUFreqTestResult {
	freq := GetIntelGPUFrequency()
	if !freq.Available {
		return IntelGPUFreqTestResult{
			Success:  false,
			Message:  "Intel GPU not available: " + freq.Error,
		}
	}

	switch testType {
	case "min":
		// Set to minimum frequency
		return SetIntelGPUFrequencyRange(freq.MinFreq, freq.MinFreq)
	case "max":
		// Set to maximum frequency
		return SetIntelGPUFrequencyRange(freq.MaxFreq, freq.MaxFreq)
	case "dynamic":
		// Restore dynamic range
		return SetIntelGPUFrequencyRange(freq.MinFreq, freq.MaxFreq)
	case "stress":
		// Stress test: rapidly switch between min and max
		for i := 0; i < 5; i++ {
			SetIntelGPUFrequencyRange(freq.MaxFreq, freq.MaxFreq)
			SetIntelGPUFrequencyRange(freq.MinFreq, freq.MinFreq)
		}
		// Restore
		final := SetIntelGPUFrequencyRange(freq.MinFreq, freq.MaxFreq)
		final.Message = "Stress test completed (5 cycles). " + final.Message
		return final
	default:
		return IntelGPUFreqTestResult{
			Success:  false,
			Message:  "Unknown test type: " + testType,
		}
	}
}

// CallIGCAPI calls Intel GPU Control Library API (placeholder for actual DLL call)
func CallIGCAPI(apiName string, params ...uintptr) (uintptr, error) {
	if igcDll == 0 {
		return 0, fmt.Errorf("IGC library not initialized")
	}

	proc, err := syscall.GetProcAddress(syscall.Handle(igcDll), apiName)
	if err != nil {
		return 0, fmt.Errorf("API %s not found: %v", apiName, err)
	}

	// Call the procedure (simplified - actual implementation would need proper parameter handling)
	switch len(params) {
	case 0:
		ret, _, _ := syscall.Syscall(proc, 0, 0, 0, 0)
		return ret, nil
	case 1:
		ret, _, _ := syscall.Syscall(proc, 1, params[0], 0, 0)
		return ret, nil
	case 2:
		ret, _, _ := syscall.Syscall(proc, 2, params[0], params[1], 0)
		return ret, nil
	case 3:
		ret, _, _ := syscall.Syscall(proc, 3, params[0], params[1], params[2])
		return ret, nil
	default:
		ret, _, _ := syscall.Syscall6(proc, uintptr(len(params)), params[0], params[1], params[2], params[3], params[4], params[5])
		return ret, nil
	}
}

// CtlFrequencySetRange wraps the Intel GPU Control Library CtlFrequencySetRange API
func CtlFrequencySetRange(adapterIndex uint32, minFreq, maxFreq uint32) IntelGPUFreqTestResult {
	// Try actual IGC DLL call first
	ret, err := CallIGCAPI("CtlFrequencySetRange", uintptr(adapterIndex), uintptr(minFreq), uintptr(maxFreq))
	if err == nil && ret == 0 {
		return IntelGPUFreqTestResult{
			Success:  true,
			Message:  fmt.Sprintf("CtlFrequencySetRange(%d, %d, %d) = 0", adapterIndex, minFreq, maxFreq),
			MinFreq:  minFreq,
			MaxFreq:  maxFreq,
		}
	}

	// Fallback to registry method
	return SetIntelGPUFrequencyRange(minFreq, maxFreq)
}

// CloseIGCLibrary releases the IGC library
func CloseIGCLibrary() {
	if igcDll != 0 {
		syscall.FreeLibrary(syscall.Handle(igcDll))
		igcDll = 0
	}
}

// compareDriverVersion compares two driver version strings
// Returns: 1 if v1 > v2, 0 if v1 == v2, -1 if v1 < v2
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

// GetIntelDriverDownloadURL returns the Intel driver download URL
func GetIntelDriverDownloadURL() string {
	return IntelDriverDownloadURL
}