//go:build windows

package backend

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"lenovo-toolkit/backend/power"
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

// GetSystemPowerInfo reads real-time power consumption.
// Priority order:
//   1. Kernel driver RAPL (MSR) → CPU package power (fastest, no subprocess)
//   2. NVML (nvml.dll) → GPU power (direct DLL call)
//   3. IPF DLL → system power + PL limits (Lenovo proprietary)
//   4. PowerShell Get-Counter → fallback for all readings
//   5. hm_smi → NPU power (if NPU available)
func GetSystemPowerInfo() SystemPowerInfo {
	info := SystemPowerInfo{}

	// ── Step 1: RAPL via kernel driver (CPU package power) ─────────────────
	raplUsed := false
	// Try RAPL — ReadRAPL() internally auto-detects CPU type and opens the device
	raplRead, err := readRAPLPowerCached()
	if err == nil && raplRead.CPUPowerWatts > 0 {
		info.CPUPowerWatts = raplRead.CPUPowerWatts
		raplUsed = true
	}

	// ── Step 2: NVML (GPU power) ──────────────────────────────────────────
	gpus := power.ReadGPUPower()
	for _, g := range gpus {
		if g.PowerAvailable && g.PowerWatts > 0 {
			info.GPUPowerWatts = g.PowerWatts
			break
		}
	}

	// ── Step 3: IPF DLL (system power + PL limits) ────────────────────────
	InitIPF()
	if ipfFunc.getSystemPower != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getSystemPower, 0, 0, 0, 0)
		info.IPFPowerMW = uint32(val)
		info.SysPowerWatts = float64(info.IPFPowerMW) / 1000.0
	}
	if ipfFunc.getCpuTemp != 0 {
		val, _, _ := syscall.Syscall(ipfFunc.getCpuTemp, 0, 0, 0, 0)
		// V1 GetCpuTempValue returns deciKelvin (K * 10)
		info.CPUTempC = float64(uint32(val))/10.0 - 273.15
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

	// ── Step 4: Fallback PowerShell (single subprocess for all) ───────────
	// Only spawn subprocess if driver/NVML didn't cover everything.
	needFallback := !raplUsed || info.GPUPowerWatts == 0

	if needFallback || info.CPUUtilPct == 0 || info.CPUFreqMHz == 0 {
		result := runPowerFallback()
		if result != "" {
			if !raplUsed {
				if v := parseValueField(result, "CPU:"); v > 0 {
					info.CPUPowerWatts = v
				}
			}
			if info.GPUPowerWatts == 0 {
				if v := parseValueField(result, "GPU:"); v > 0 {
					info.GPUPowerWatts = v
				}
			}
			if info.CPUUtilPct == 0 {
				if v := parseValueField(result, "UTIL:"); v > 0 {
					info.CPUUtilPct = v
				}
			}
			if info.CPUFreqMHz == 0 {
				if v := parseValueField(result, "FREQ:"); v > 100 {
					info.CPUFreqMHz = v
				}
			}
		}
	}

	// Last resort: estimate CPU power from total system power
	if info.CPUPowerWatts == 0 && info.SysPowerWatts > 0 {
		info.CPUPowerWatts = info.SysPowerWatts * 0.65
	}

	// ── Step 5: NPU power (hm_smi) ───────────────────────────────────────
	info.NPUPowerWatts = readNPUPower()

	return info
}

// ── RAPL cached reader ───────────────────────────────────────────────────

var (
	raplCache   power.RAPLReading
	raplCacheMu sync.Mutex
	raplLastTs  int64
)

func readRAPLPowerCached() (power.RAPLReading, error) {
	reading, err := power.ReadRAPL()
	if err != nil {
		return reading, err
	}
	raplCacheMu.Lock()
	raplCache = reading
	raplLastTs = reading.Timestamp
	raplCacheMu.Unlock()
	return reading, nil
}

// ── Fallback: single PowerShell invocation for all remaining counters ──────

func runPowerFallback() string {
	script := `
$cpuPwr=0; $gpuPwr=0; $cpuUtil=0; $freq=0
try {
	$cpus = Get-Counter '\Processor Power(0,*)\Package Core Power' -ErrorAction SilentlyContinue
	if ($cpus) {
		foreach ($c in $cpus.CounterSamples) {
			if ($c.InstanceName -match '^0,') { $cpuPwr = [math]::Round($c.CookedValue / 1000, 3) }
		}
	}
	$gpus = Get-Counter '\GPU Engine(*)\GPU Power' -ErrorAction SilentlyContinue
	if ($gpus) {
		foreach ($c in $gpus.CounterSamples) {
			if ($c.CookedValue -gt $gpuPwr) { $gpuPwr = [math]::Round($c.CookedValue, 3) }
		}
	}
	$util = Get-Counter '\Processor Information(_Total)\% Processor Performance' -ErrorAction SilentlyContinue
	if ($util) { $cpuUtil = [math]::Round(100 - $util.CounterSamples[0].CookedValue, 1); if ($cpuUtil -lt 0) { $cpuUtil = 0 } }
	$frq = Get-Counter '\Processor Information(_Total)\% Performance Frequency' -ErrorAction SilentlyContinue
	if ($frq) { $freq = [math]::Round($frq.CounterSamples[0].CookedValue, 0) }
} catch {}
Write-Host CPU:$cpuPwr GPU:$gpuPwr UTIL:$cpuUtil FREQ:$freq
`
	var out []byte
	var err error
	out, err = hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// parseValueField extracts a numeric value from a key:value formatted string.
// Example: parseValueField("CPU:12.5 GPU:45.0 UTIL:23", "CPU:") → 12.5
func parseValueField(s, key string) float64 {
	idx := strings.Index(s, key)
	if idx < 0 {
		return 0
	}
	s = s[idx+len(key):]
	// Find end of value (space or end-of-string)
	end := strings.IndexAny(s, " \t\r\n")
	valStr := s
	if end >= 0 {
		valStr = s[:end]
	}
	v, err := strconv.ParseFloat(valStr, 64)
	if err != nil {
		return 0
	}
	return v
}

// ── NPU Power ────────────────────────────────────────────────────────────

func readNPUPower() float64 {
	if npuFunc.getBoardPower == 0 {
		return 0
	}
	hmPath := hmSMI
	script := fmt.Sprintf(`
$np = 0
try {
	$out = & '%s' -g power -d 0 2>$null
	foreach ($line in $out) {
		if ($line -match 'Board_Power') {
			$parts = $line -split '\s+'
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
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return 0
	}
	val := strings.TrimSpace(string(out))
	v, err := strconv.ParseFloat(val, 64)
	if err != nil || v <= 0 {
		return 0
	}
	return v
}

// ── Cached power info for polling ─────────────────────────────────────

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