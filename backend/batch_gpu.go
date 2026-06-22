//go:build windows

package backend

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/sys/windows"
)

// BatchGPUResult contains all GPU-related data fetched in a single PowerShell call.
type BatchGPUResult struct {
	GPUList       []GPUInfo       `json:"gpuList"`
	IGPUStatus    IGPUStatus      `json:"igpuStatus"`
	NVIDIAStatus  NVIDIAStatus    `json:"nvidiaStatus"`
	GPUPrefStatus GPUPrefStatus   `json:"gpuPrefStatus"`
	IntelGPU      IntelGPUFrequency `json:"intelGPU"`
}

// gpuListCache caches GPU hardware list (doesn't change at runtime)
var (
	gpuListCache     []GPUInfo
	gpuListCacheOnce sync.Once
)

// GetCachedGPUList returns cached GPU list, initializing on first call
func GetCachedGPUList() []GPUInfo {
	gpuListCacheOnce.Do(func() {
		gpuListCache = EnumerateGPUs()
	})
	return gpuListCache
}

// InvalidateGPUCache forces re-enumeration of GPUs on next access
func InvalidateGPUCache() {
	gpuListCacheOnce = sync.Once{}
}

// BatchGPUInit performs all GPU initialization queries in a single PowerShell call.
// This replaces the previous pattern of 8 separate PowerShell processes (~2s) with
// one combined call (~350ms), then augments with Go-native registry reads (<1ms).
func BatchGPUInit() BatchGPUResult {
	result := BatchGPUResult{}

	// ── Phase 1: Single PowerShell for WMI/CIM queries that can't be done in Go ──
	script := `
$ErrorActionPreference = 'SilentlyContinue'

# GPU List (WMI)
$vids = Get-CimInstance Win32_VideoController
foreach ($vid in $vids) {
    $hwId = ''
    $pnpId = ''
    try { $pnpId = $vid.PNPDeviceId } catch {}
    if ($pnpId -ne '') {
        $pnpEsc = $pnpId -replace '\\', '_'
        $regPath = "HKLM:\SYSTEM\CurrentControlSet\Enum\$pnpEsc\Device Parameters"
        $hwIdVal = (Get-ItemProperty -Path $regPath -Name HardwareID -ErrorAction SilentlyContinue).HardwareID
        if ($hwIdVal -is [string[]]) { $hwId = $hwIdVal[0] } elseif ($hwIdVal -is [string]) { $hwId = $hwIdVal }
    }
    $dedRam = [uint64]0
    try { $dedRam = [uint64]$vid.AdapterRAM } catch {}
    $drvDate = ''
    try { $drvDate = $vid.DriverDate.ToString('yyyyMMdd') } catch {}
    Write-Output "GPU|$($vid.Name)|$($vid.DriverVersion)|$drvDate|$dedRam|$hwId|$pnpId"
}

# IGPU Mode (WMI - LENOVO_GAMEZONE_DATA)
try {
    $class = Get-CimInstance -Namespace root/wmi -ClassName LENOVO_GAMEZONE_DATA | Where-Object { $_.InstanceName -like '*GMZN*' } | Select-Object -First 1
    if ($class) {
        $modeResult = $class.GetIGPUModeStatus()
        if ($modeResult.Data -ne $null) {
            Write-Output "IGPU_MODE|Mode:$($modeResult.Data)"
        } else {
            Write-Output "IGPU_MODE|NotAvailable"
        }
    } else {
        Write-Output "IGPU_MODE|NotFound"
    }
} catch {
    Write-Output "IGPU_MODE|Error"
}

# NVIDIA check
$nvDetected = $false
$nvServiceRunning = $false
foreach ($vid in $vids) {
    $name = $vid.Name
    if ($name -match 'NVIDIA|GeForce|RTX|GTX') { $nvDetected = $true }
}
$nvSvc = Get-Service nvcontainer -ErrorAction SilentlyContinue
if ($nvSvc -and $nvSvc.Status -eq 'Running') { $nvServiceRunning = $true }
Write-Output "NVIDIA|$nvDetected|$nvServiceRunning"

# GPU Utilization (Performance Counter)
try {
    $samples = (Get-Counter '\GPU Engine(*engtype_3D)\Utilization Percentage' -ErrorAction Stop).CounterSamples
    $total = ($samples | Measure-Object -Property CookedValue -Sum).Sum
    Write-Output "GPU_UTIL|$([math]::Round($total, 1))"
} catch {
    Write-Output "GPU_UTIL|-1"
}
`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		// PowerShell failed entirely; use Go-native fallbacks
		result.GPUList = []GPUInfo{}
		result.IGPUStatus = IGPUStatus{Available: false, Mode: 0xFFFFFFFF}
		result.NVIDIAStatus = NVIDIAStatus{}
		result.GPUPrefStatus = GetGPUPrefStatusFromCache()
		result.IntelGPU = GetIntelGPUFrequency()
		return result
	}

	// ── Parse PowerShell output ──
	var gpuList []GPUInfo
	var igpuMode uint32 = 0xFFFFFFFF
	var igpuAvail = false
	nvStatus := NVIDIAStatus{}
	var gpuUtil float64 = -1

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Parse GPU entries
		if strings.HasPrefix(line, "GPU|") {
			parts := strings.Split(line, "|")
			if len(parts) < 6 {
				continue
			}
			name := strings.TrimSpace(parts[1])
			driverVersion := strings.TrimSpace(parts[2])
			driverDate := strings.TrimSpace(parts[3])
			dedRam, _ := strconv.ParseUint(strings.TrimSpace(parts[4]), 10, 64)
			hardwareId := strings.TrimSpace(parts[5])
			pnpId := ""
			if len(parts) > 6 {
				pnpId = strings.TrimSpace(parts[6])
			}
			if name == "" {
				continue
			}

			if dedRam > 128*1024*1024*1024 {
				dedRam = 0
			}

			vendorId, deviceId, subVendorId, subSystemId, revisionId := parseHardwareId(hardwareId)
			busNumber := getGPUBusNumber(pnpId)

			isDiscrete := dedRam >= 256*1024*1024
			lower := strings.ToLower(name)
			if strings.Contains(lower, "nvidia") || strings.Contains(lower, "geforce") ||
				strings.Contains(lower, "rtx") || strings.Contains(lower, "gtx") ||
				strings.Contains(lower, "amd") || strings.Contains(lower, "radeon") ||
				strings.Contains(lower, "arc") {
				if dedRam >= 64*1024*1024 {
					isDiscrete = true
				}
			}

			totalMem := getSystemTotalMemory()
			sharedMem := uint64(0)
			if !isDiscrete && totalMem > 0 {
				sharedMem = totalMem / 2
			} else if isDiscrete {
				sharedMem = getGPUSharedMemory(name)
			}

			gpu := GPUInfo{
				Name:            name,
				VendorId:        vendorId,
				DeviceId:        deviceId,
				SubVendorId:     subVendorId,
				SubSystemId:     subSystemId,
				RevisionId:      revisionId,
				DriverVersion:   driverVersion,
				DriverDate:      driverDate,
				DedicatedMemory: dedRam,
				SharedMemory:    sharedMem,
				TotalMemory:     dedRam + sharedMem,
				IsDiscrete:      isDiscrete,
				HardwareId:      hardwareId,
				BusNumber:       busNumber,
			}
			gpuList = append(gpuList, gpu)
		}

		// Parse IGPU mode
		if strings.HasPrefix(line, "IGPU_MODE|Mode:") {
			var m uint32
			_, scanErr := fmt.Sscanf(strings.TrimPrefix(line, "IGPU_MODE|"), "Mode:%d", &m)
			if scanErr == nil {
				igpuMode = m
				igpuAvail = true
			}
		}

		// Parse NVIDIA status
		if strings.HasPrefix(line, "NVIDIA|") {
			parts := strings.Split(strings.TrimPrefix(line, "NVIDIA|"), "|")
			if len(parts) >= 2 {
				nvStatus.Detected = strings.EqualFold(parts[0], "True")
				nvStatus.ServiceRunning = strings.EqualFold(parts[1], "True")
			}
		}

		// Parse GPU utilization
		if strings.HasPrefix(line, "GPU_UTIL|") {
			var v float64
			fmt.Sscanf(strings.TrimPrefix(line, "GPU_UTIL|"), "%f", &v)
			gpuUtil = v
		}
	}

	// NVML check (no PowerShell needed)
	hNVML, _ := windows.LoadLibrary("nvml.dll")
	if hNVML != 0 {
		nvStatus.NVMLLoaded = true
		nvStatus.Detected = true
		windows.FreeLibrary(hNVML)
	}

	// Update GPU list cache
	gpuListCache = gpuList

	// ── Phase 2: Go-native registry reads (instant, <1ms) ──
	gpuPref := GetGPUPrefStatusFromCache()
	intelGPU := GetIntelGPUFrequencyWithUtil(gpuUtil)  // pass pre-fetched util, skip redundant PS call

	result.GPUList = gpuList
	result.IGPUStatus = IGPUStatus{Available: igpuAvail, Mode: igpuMode}
	result.NVIDIAStatus = nvStatus
	result.GPUPrefStatus = gpuPref
	result.IntelGPU = intelGPU

	return result
}
