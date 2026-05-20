//go:build windows

package backend

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"syscall"
)

// SystemPowerInfo holds real-time power consumption data (HWInfo-style).
type SystemPowerInfo struct {
	CPUPowerWatts float64 `json:"cpuPowerWatts"` // CPU package power in Watts
	GPUPowerWatts float64 `json:"gpuPowerWatts"` // GPU power in Watts
	NPUPowerWatts float64 `json:"npuPowerWatts"` // NPU board power in Watts
	SysPowerWatts float64 `json:"sysPowerWatts"` // Total system power in Watts
	IPFPowerMW    uint32  `json:"ipfPowerMW"`    // IPF system power in milliwatts
	CPUUtilPct    float64 `json:"cpuUtilPct"`    // CPU utilization %
	PL1Watts      float64 `json:"pl1Watts"`      // PL1 in Watts
	PL2Watts      float64 `json:"pl2Watts"`      // PL2 in Watts
	PL4Watts      float64 `json:"pl4Watts"`      // PL4 in Watts
	CPUFreqMHz    float64 `json:"cpuFreqMHz"`    // CPU frequency in MHz
	CPUTempC      float64 `json:"cpuTempC"`      // CPU temperature in C
}

// GetSystemPowerInfo reads real-time power consumption using HWInfo-style approach:
//  1. IPF DLL for system power and PL limits
//  2. PowerShell Get-Counter for CPU/GPU power sensors
//  3. hm_smi for NPU power (if available)
func GetSystemPowerInfo() SystemPowerInfo {
	info := SystemPowerInfo{}

	// Try IPF DLL first (fastest, most accurate)
	InitIPF()
	if ipfFunc.getSystemPower != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getSystemPower, 0, 0, 0, 0)
		info.IPFPowerMW = uint32(val)
		info.SysPowerWatts = float64(info.IPFPowerMW) / 1000.0
	}
	if ipfFunc.getCpuTemp != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getCpuTemp, 0, 0, 0, 0)
		info.CPUTempC = float64(uint32(val)) / 1000.0
	}
	if ipfFunc.getPL1 != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getPL1, 0, 0, 0, 0)
		info.PL1Watts = float64(uint32(val)) / 1000.0
	}
	if ipfFunc.getPL2 != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getPL2, 0, 0, 0, 0)
		info.PL2Watts = float64(uint32(val)) / 1000.0
	}
	if ipfFunc.getPL4 != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getPL4, 0, 0, 0, 0)
		info.PL4Watts = float64(uint32(val)) / 1000.0
	}

	// PowerShell Get-Counter for CPU package power (HWInfo sensor style)
	cpuPowerScript := `
$cpuPower = 0
try {
	$cpuCounters = Get-Counter '\Processor Power(0,*)\Package Core Power' -ErrorAction SilentlyContinue
	if ($cpuCounters) {
		foreach ($c in $cpuCounters.CounterSamples) {
			if ($c.InstanceName -match '^0,') { $cpuPower = [math]::Round($c.CookedValue / 1000, 3) }
		}
	}
} catch {}
$cpuPower
`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", cpuPowerScript).Output()
	if err == nil {
		val := strings.TrimSpace(string(out))
		if v, err := strconv.ParseFloat(val, 64); err == nil && v > 0 {
			info.CPUPowerWatts = v
		}
	}

	// Fallback: use IPF system power * 0.65 as CPU estimate
	if info.CPUPowerWatts == 0 && info.SysPowerWatts > 0 {
		info.CPUPowerWatts = info.SysPowerWatts * 0.65
	}

	// GPU power via GPU Engine counters (HWInfo-style)
	gpuPowerScript := `
$gpuPower = 0
try {
	$gpuCounters = Get-Counter '\GPU Engine(*)\GPU Power' -ErrorAction SilentlyContinue
	if ($gpuCounters) {
		foreach ($c in $gpuCounters.CounterSamples) {
			$val = [math]::Round($c.CookedValue, 3)
			if ($val -gt $gpuPower) { $gpuPower = $val }
		}
	}
} catch {}
$gpuPower
`
	out, err = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", gpuPowerScript).Output()
	if err == nil {
		val := strings.TrimSpace(string(out))
		if v, err := strconv.ParseFloat(val, 64); err == nil && v > 0 {
			info.GPUPowerWatts = v
		}
	}

	// CPU utilization via Get-Counter
	cpuUtilScript := `
$util = 0
try {
	$cpuUtil = Get-Counter '\Processor Information(_Total)\% Processor Performance' -ErrorAction SilentlyContinue
	if ($cpuUtil) {
		$util = [math]::Round(100 - $cpuUtil.CounterSamples[0].CookedValue, 1)
		if ($util -lt 0) { $util = 0 }
	} else {
		$cpuUtil2 = Get-Counter '\Processor(_Total)\% Processor Time' -ErrorAction SilentlyContinue
		if ($cpuUtil2) { $util = [math]::Round($cpuUtil2.CounterSamples[0].CookedValue, 1) }
	}
} catch {}
$util
`
	out, err = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", cpuUtilScript).Output()
	if err == nil {
		val := strings.TrimSpace(string(out))
		if v, err := strconv.ParseFloat(val, 64); err == nil {
			info.CPUUtilPct = v
		}
	}

	// CPU frequency
	freqScript := `
$freq = 0
try {
	$freqCounters = Get-Counter '\Processor Information(_Total)\% Performance Frequency' -ErrorAction SilentlyContinue
	if ($freqCounters) {
		$freq = [math]::Round($freqCounters.CounterSamples[0].CookedValue, 0)
	}
} catch {}
$freq
`
	out, err = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", freqScript).Output()
	if err == nil {
		val := strings.TrimSpace(string(out))
		if v, err := strconv.ParseFloat(val, 64); err == nil && v > 100 {
			info.CPUFreqMHz = v
		}
	}

	// NPU power via hm_smi (if NPU available)
	if npuFunc.getBoardPower != 0 {
		hmPath := hmSMI
		npuPowerScript := fmt.Sprintf(`
$np = 0
try {
	$out = & '%s' -g power -d 0 2>$null
	foreach ($line in $out) {
		if ($line -match 'Board_Power') {
			$parts = $line -split 's+'
			foreach ($p in $parts) {
				if ([double]::TryParse($p, [ref]$null)) {
					$np = [math]::Round([double]$p, 3)
					break
				}
			}
		}
	}
} catch {}
$np
`, hmPath)
		out, _ = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", npuPowerScript).Output()
		val := strings.TrimSpace(string(out))
		if v, err := strconv.ParseFloat(val, 64); err == nil && v > 0 {
			info.NPUPowerWatts = v
		}
	}

	return info
}

// Cached power info for polling
var (
	sysPowerCache   SystemPowerInfo
	sysPowerCacheMu sync.Mutex
)

// GetCachedSystemPower returns the most recent power reading (thread-safe).
func GetCachedSystemPower() SystemPowerInfo {
	sysPowerCacheMu.Lock()
	defer sysPowerCacheMu.Unlock()
	return sysPowerCache
}

// UpdateCachedSystemPower refreshes the cache and returns the new value.
func UpdateCachedSystemPower() SystemPowerInfo {
	sysPowerCacheMu.Lock()
	defer sysPowerCacheMu.Unlock()
	sysPowerCache = GetSystemPowerInfo()
	return sysPowerCache
}
