//go:build windows

package backend

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// GPUInfo represents a single GPU's information
type GPUInfo struct {
	Name            string `json:"name"`
	VendorId        uint32 `json:"vendorId"`
	DeviceId        uint32 `json:"deviceId"`
	SubVendorId     uint32 `json:"subVendorId"`
	SubSystemId     uint32 `json:"subSystemId"`
	RevisionId      uint32 `json:"revisionId"`
	DriverVersion   string `json:"driverVersion"`
	DriverDate      string `json:"driverDate"`
	DedicatedMemory uint64 `json:"dedicatedMemory"`
	SharedMemory    uint64 `json:"sharedMemory"`
	TotalMemory     uint64 `json:"totalMemory"`
	IsDiscrete      bool   `json:"isDiscrete"`
	HardwareId      string `json:"hardwareId"`
	BusNumber       uint32 `json:"busNumber"`
}

// GPUProcess represents a process using GPU
type GPUProcess struct {
	PID    uint32 `json:"pid"`
	Name   string `json:"name"`
	Memory string `json:"memory"`
}

// NVIDIAStatus represents NVIDIA GPU status
type NVIDIAStatus struct {
	Detected       bool `json:"detected"`
	NVMLLoaded    bool `json:"nvmlLoaded"`
	ServiceRunning bool `json:"serviceRunning"`
}

// IGPUStatus represents IGPU mode status
type IGPUStatus struct {
	Available bool   `json:"available"`
	Mode      uint32 `json:"mode"`
}

// GPUPrefStatus represents the real-time GPU mode status from multiple sources
type GPUPrefStatus struct {
	Available              bool   `json:"available"`
	Value                 uint32 `json:"value"`
	Label                 string `json:"label"`
	PCMStatus             uint32 `json:"pcmStatus"`
	PCMStatusAvail        bool   `json:"pcmStatusAvail"`
	PCMLabel              string `json:"pcmLabel"`
	VantageGPUStatus      uint32 `json:"vantageGPUStatus"`
	VantageGPUStatusAvail bool   `json:"vantageGPUStatusAvail"`
	VantageDefaultMode    uint32 `json:"vantageDefaultMode"`
	VantageDefaultModeAvail bool `json:"vantageDefaultModeAvail"`
	PCMServiceRunning     bool   `json:"pcmServiceRunning"`
	VantageServiceRunning bool   `json:"vantageServiceRunning"`
}

// GPUAutoGear represents the auto gear setting for GPU hybrid mode
type GPUAutoGear struct {
	Available bool   `json:"available"`
	Value     uint32 `json:"value"`
}

// SetResult represents a set operation result
type SetResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// EnumerateGPUs returns GPU list using registry + CIM (no WMI)
func EnumerateGPUs() []GPUInfo {
	var gpus []GPUInfo

	script := `
$ErrorActionPreference = 'SilentlyContinue'
$vids = Get-CimInstance Win32_VideoController
foreach ($vid in $vids) {
    $hwId = ''
    $devId = ''
    $name = $vid.Name
    $drvVer = $vid.DriverVersion
    $drvDate = ''
    try { $drvDate = $vid.DriverDate.ToString('yyyyMMdd') } catch {}
    $dedRam = [uint64]0
    try { $dedRam = [uint64]$vid.AdapterRAM } catch {}

    $pnpId = ''
    try { $pnpId = $vid.PNPDeviceId } catch {}
    if ($pnpId -ne '') {
        $pnpId = $pnpId -replace '\\', '_'
        $regPath = "HKLM:\SYSTEM\CurrentControlSet\Enum\$pnpId\Device Parameters"
        $hwIdVal = (Get-ItemProperty -Path $regPath -Name HardwareID -ErrorAction SilentlyContinue).HardwareID
        if ($hwIdVal -is [string[]]) { $hwId = $hwIdVal[0] } elseif ($hwIdVal -is [string]) { $hwId = $hwIdVal }
    }

    Write-Output "GPU|$name|$drvVer|$drvDate|$dedRam|$hwId|$pnpId"
}
`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return gpus
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "GPU|") {
			continue
		}
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

		vendorId, deviceId, subVendorId, subSystemId, revisionId := parseHardwareId(hardwareId)
		busNumber := getGPUBusNumber(pnpId)

		if dedRam > 128*1024*1024*1024 {
			dedRam = 0
		}

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

		gpus = append(gpus, gpu)
	}

	return gpus
}

// EnumerateGPUProcesses returns process list using PowerShell (no WMI)
func EnumerateGPUProcesses() []GPUProcess {
	var processes []GPUProcess

	script := `
$ErrorActionPreference = 'SilentlyContinue'
Get-Process | Where-Object { $_.WorkingSet64 -gt 104857600 } | ForEach-Object {
    Write-Output "$($_.Id)|$($_.ProcessName)|$([math]::Round($_.WorkingSet64/1MB, 1))"
}
`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return processes
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 3 {
			continue
		}
		pid, _ := strconv.ParseUint(strings.TrimSpace(parts[0]), 10, 32)
		name := strings.TrimSpace(parts[1])
		mem := strings.TrimSpace(parts[2])
		processes = append(processes, GPUProcess{
			PID:    uint32(pid),
			Name:   name,
			Memory: fmt.Sprintf("%s MB", mem),
		})
	}

	return processes
}

// CheckNVIDIAStatus checks if NVIDIA GPU is present and its status
func CheckNVIDIAStatus() NVIDIAStatus {
	status := NVIDIAStatus{
		Detected:       false,
		NVMLLoaded:    false,
		ServiceRunning: false,
	}

	hNVML, _ := windows.LoadLibrary("nvml.dll")
	if hNVML != 0 {
		status.NVMLLoaded = true
		status.Detected = true
		windows.FreeLibrary(hNVML)
	}

	script := `Get-CimInstance Win32_VideoController | ForEach-Object { Write-Output $_.Name }`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err == nil {
		lower := strings.ToLower(string(out))
		if strings.Contains(lower, "nvidia") || strings.Contains(lower, "geforce") ||
			strings.Contains(lower, "rtx") || strings.Contains(lower, "gtx") {
			status.Detected = true
		}
	}

	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", "sc query nvcontainer 2>$null")
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "RUNNING") {
		status.ServiceRunning = true
	}

	return status
}

// GetIGPUModeStatusWMI reads IGPU mode via WMI (LENOVO_GAMEZONE_DATA.GetIGPUModeStatus)
// This is the same method used by the official Lenovo DGPUtool.
// Falls back to registry if WMI is not available.
func GetIGPUModeStatusWMI() (bool, uint32) {
	// Try WMI first (same as DGPUtool: CLSID_WbemLocator, root\wmi, LENOVO_GAMEZONE_DATA)
	mode, ok := getIGPUModeViaWMI()
	if ok {
		return true, mode
	}
	// Fallback: try registry
	script := `
$ErrorActionPreference = 'SilentlyContinue'
try {
    $key = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider' -Name ITS_GPUHybridModeSetting -ErrorAction Stop
    Write-Host "Mode:$($key.ITS_GPUHybridModeSetting)"
} catch {
    Write-Host "NotAvailable"
}
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return false, 0xFFFFFFFF
	}
	outputStr := string(output)
	if strings.Contains(outputStr, "NotAvailable") || outputStr == "" {
		return false, 0xFFFFFFFF
	}
	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Mode:") {
			var m uint32
			_, scanErr := fmt.Sscanf(line, "Mode:%d", &m)
			if scanErr == nil {
				return true, m
			}
		}
	}
	return false, 0xFFFFFFFF
}

// getIGPUModeViaWMI calls LENOVO_GAMEZONE_DATA.GetIGPUModeStatus via COM
func getIGPUModeViaWMI() (uint32, bool) {
	// This uses syscall to create a COM object and call the WMI method directly.
	// Same approach as the official DGPUtool.
	// We use PowerShell as a subprocess to avoid needing to set up COM threading in Go.
	script := `
$ErrorActionPreference = 'SilentlyContinue'
try {
    $class = Get-CimInstance -Namespace root/wmi -ClassName LENOVO_GAMEZONE_DATA | Where-Object { $_.InstanceName -like '*GMZN*' } | Select-Object -First 1
    if ($null -eq $class) {
        Write-Output "NotFound"
        return
    }
    $result = $class.GetIGPUModeStatus()
    if ($result.Data -ne $null) {
        Write-Output "Mode:$($result.Data)"
    } else {
        Write-Output "NotAvailable"
    }
} catch {
    Write-Output "Error:$_"
}
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return 0, false
	}
	outputStr := string(output)
	if strings.HasPrefix(outputStr, "Mode:") {
		var m uint32
		_, scanErr := fmt.Sscanf(strings.TrimSpace(outputStr), "Mode:%d", &m)
		if scanErr == nil {
			return m, true
		}
	}
	return 0, false
}

// SetIGPUModeStatusWMI sets IGPU mode via Lenovo GameZone WMI
func SetIGPUModeStatusWMI(mode uint32) (bool, uint32) {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
try {
    $class = Get-CimInstance -Namespace root/wmi -ClassName LENOVO_GAMEZONE_DATA | Where-Object { $_.InstanceName -like '*GMZN*' -or $_.InstanceName -like '*ACPI*' } | Select-Object -First 1
    if ($class) {
        $result = $class.SetIGPUModeStatus(%d)
        Write-Host "Success"
    } else {
        Write-Host "ClassNotFound"
    }
} catch {
    Write-Host "Error:$_"
}`, mode)

	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "Success") {
		return true, mode
	}

	// Try registry write
	return setIGPUModeRegistry(mode)
}

// GetGPUPrefStatus reads GPU mode from multiple registry sources
// Uses cached value from registry watcher if available (instant, no PowerShell)
func GetGPUPrefStatus() GPUPrefStatus {
	// Try cached value first (from registry watcher)
	cached := GetGPUStatusCached()
	if cached.Available || cached.PCMStatusAvail {
		return cached
	}
	
	// Fallback to direct read (slower)
	return readGPUStatusDirect()
}

func pcmStatusLabel(val uint32) string {
	switch val {
	case 0:
		return "无驱动"
	case 1:
		return "集显模式"
	case 2:
		return "智能模式"
	case 3:
		return "双显模式"
	default:
		return fmt.Sprintf("未知 (%d)", val)
	}
}

func gpuPrefStatusLabel(val uint32) string {
	switch val {
	case 2:
		return "智能模式"
	case 1:
		return "集显模式"
	case 3:
		return "双显模式"
	default:
		return fmt.Sprintf("未知 (%d)", val)
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────

func parseHardwareId(hwId string) (vendorId, deviceId, subVendorId, subSystemId, revisionId uint32) {
	if hwId == "" {
		return
	}
	re := regexp.MustCompile(`(?i)VEN_([0-9A-Fa-f]{4})`)
	m := re.FindStringSubmatch(hwId)
	if len(m) >= 2 {
		v, _ := strconv.ParseUint(m[1], 16, 64)
		vendorId = uint32(v)
	}

	re = regexp.MustCompile(`(?i)&DEV_([0-9A-Fa-f]{4})`)
	m = re.FindStringSubmatch(hwId)
	if len(m) >= 2 {
		v, _ := strconv.ParseUint(m[1], 16, 64)
		deviceId = uint32(v)
	}

	re = regexp.MustCompile(`(?i)SUBSYS_([0-9A-Fa-f]{8})`)
	m = re.FindStringSubmatch(hwId)
	if len(m) >= 2 {
		v, _ := strconv.ParseUint(m[1], 16, 64)
		subSystemId = uint32(v)
		subVendorId = uint32(v >> 16)
	}

	re = regexp.MustCompile(`(?i)&REV_([0-9A-Fa-f]{2})`)
	m = re.FindStringSubmatch(hwId)
	if len(m) >= 2 {
		v, _ := strconv.ParseUint(m[1], 16, 64)
		revisionId = uint32(v)
	}

	return
}

func getGPUBusNumber(pnpId string) uint32 {
	if pnpId == "" {
		return 0
	}
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$pnpId = '%s' -replace '\\', '_'
$regPath = "HKLM:\SYSTEM\CurrentControlSet\Enum\$pnpId\Device Parameters"
$busNum = (Get-ItemProperty -Path $regPath -Name BusNumber -ErrorAction SilentlyContinue).BusNumber
if ($busNum -eq $null) { $busNum = 0 }
Write-Output $busNum
`, strings.ReplaceAll(pnpId, "'", "''"))
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return 0
	}
	bus, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 32)
	return uint32(bus)
}

func getSystemTotalMemory() uint64 {
	type MEMORYSTATUSEX struct {
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
	var mem MEMORYSTATUSEX
	mem.dwLength = uint32(unsafe.Sizeof(mem))
	ret, _, _ := windows.NewLazySystemDLL("kernel32.dll").NewProc("GlobalMemoryStatusEx").Call(uintptr(unsafe.Pointer(&mem)))
	if ret == 0 {
		return 0
	}
	return mem.ullTotalPhys
}

func getGPUSharedMemory(gpuName string) uint64 {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$vid = Get-CimInstance Win32_VideoController | Where-Object { $_.Name -like '*%s*' } | Select-Object -First 1
if ($vid) {
    $shared = [uint64]0
    try { $shared = [uint64]$vid.SharedMemory } catch {}
    Write-Output $shared
} else {
    Write-Output 0
}
`, strings.ReplaceAll(gpuName, "'", "''"))
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return 0
	}
	shared, _ := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	return shared
}

func setIGPUModeRegistry(mode uint32) (bool, uint32) {
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$paths = @(
    'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider',
    'HKLM:\SOFTWARE\Lenovo\GameZone',
    'HKLM:\SOFTWARE\Lenovo\PowerManagement'
)
foreach ($path in $paths) {
    if (Test-Path $path) {
        try {
            Set-ItemProperty -Path $path -Name ITS_GPUHybridModeSetting -Value %d -Type DWord -Force
            Write-Host "Success"
            break
        } catch {}
    }
}
`, mode)
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	output, err := cmd.Output()
	if err == nil && strings.Contains(string(output), "Success") {
		return true, mode
	}
	return false, 0xFFFFFFFF
}

// GetGPUAutoGear reads the ITS_GPUHybridModeSetting value from registry
func GetGPUAutoGear() GPUAutoGear {
	result := GPUAutoGear{}
	script := `
$ErrorActionPreference = 'SilentlyContinue'
try {
    $key = Get-ItemProperty -Path 'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider' -Name ITS_GPUHybridModeSetting -ErrorAction Stop
    Write-Host "Value:$($key.ITS_GPUHybridModeSetting)"
} catch {
    Write-Host "NotFound"
}
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		return result
	}
	outputStr := string(output)
	if strings.Contains(outputStr, "NotFound") || outputStr == "" {
		return result
	}
	for _, line := range strings.Split(outputStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "Value:") {
			var v uint32
			_, scanErr := fmt.Sscanf(line, "Value:%d", &v)
			if scanErr == nil {
				result.Available = true
				result.Value = v
				return result
			}
		}
	}
	return result
}

// SetGPUAutoGear sets the ITS_GPUHybridModeSetting value
func SetGPUAutoGear(value uint32) SetResult {
	result := SetResult{Success: false}
	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$paths = @(
    'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider',
    'HKLM:\SOFTWARE\Lenovo\GameZone',
    'HKLM:\SOFTWARE\Lenovo\PowerManagement'
)
$success = $false
foreach ($path in $paths) {
    if (Test-Path $path) {
        try {
            Set-ItemProperty -Path $path -Name ITS_GPUHybridModeSetting -Value %d -Type DWord -Force
            $success = $true
            break
        } catch {}
    }
}
if ($success) {
    Write-Host "Success"
} else {
    Write-Host "Failed"
}
`, value)
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	output, err := cmd.Output()
	if err != nil {
		result.Message = "Failed to execute PowerShell command"
		return result
	}
	if strings.Contains(string(output), "Success") {
		result.Success = true
		result.Message = "Auto Gear setting applied successfully"
	} else {
		result.Message = "Failed to set Auto Gear"
	}
	return result
}
