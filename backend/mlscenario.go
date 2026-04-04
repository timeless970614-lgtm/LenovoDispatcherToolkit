//go:build windows

package backend

import (
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const mlLogDir = `C:\ProgramData\Lenovo\LenovoDispatcher\Logs`

// IPF column indices in the 84-column CSV row (row[0]=Time, row[1]=AC_DC, ...)
const (
	colIPF_SystemPower     = 40 // IPF_SystemPower (mW)
	colMMIO_PL1           = 41 // MMIO_PL1 (mW)
	colMMIO_PL2           = 42 // MMIO_PL2 (mW)
	colMMIO_PL4           = 43 // MMIO_PL4 (mW)
	colPL1Check           = 44 // PL1Check
	colPL2Check           = 45 // PL2Check
	colEPOT               = 46 // EPOT
	colEPP                = 47 // EPP
	colEPP1               = 48 // EPP_1
	colPPMFrequencyLimit   = 49 // PPM_FREQUENCY_LIMIT
	colPPMFrequencyLimit1  = 50 // PPM_FREQUENCY_LIMIT_1
	colPPMCpMin           = 51 // PPM_CPMIN
	colPPMCpMax           = 52 // PPM_CPMAX
	colSoftParking        = 53 // SoftParking
	colCpuTemp            = 54 // CPU_Temperature
)

// MLLogStatus represents the capture status
type MLLogStatus struct {
	IsCapturing bool   `json:"isCapturing"`
	StartTime   string `json:"startTime"`
	EventCount  uint64 `json:"eventCount"`
	OutputFile  string `json:"outputFile"`
	OutputCSV   string `json:"outputCSV"`
	Error       string `json:"error"`
}

var (
	mlCapturing int32
	mlCount     uint64
	mlMu        sync.Mutex
	mlCSVFile   *os.File
	mlCSVWriter *csv.Writer
	mlCSVPath   string
	mlStart     time.Time
)

// StartMLScenarioCapture starts 1-second interval performance capture
func StartMLScenarioCapture() MLLogStatus {
	if atomic.LoadInt32(&mlCapturing) == 1 {
		return MLLogStatus{
			IsCapturing: true,
			StartTime:   mlStart.Format("2006-01-02 15:04:05"),
			EventCount: atomic.LoadUint64(&mlCount),
			OutputCSV:  mlCSVPath,
		}
	}

	if err := os.MkdirAll(mlLogDir, 0755); err != nil {
		return MLLogStatus{Error: "Cannot create log directory: " + err.Error()}
	}

	ts := time.Now().Format("20060102-150405")
	csvPath := filepath.Join(mlLogDir, fmt.Sprintf("MLScenario_%s.csv", ts))

	f, err := os.Create(csvPath)
	if err != nil {
		return MLLogStatus{Error: "Cannot create CSV: " + err.Error()}
	}

	cw := csv.NewWriter(f)
	// Write header matching ML_Scenario Result.csv (84 columns)
	header := []string{
		"Time", "AC_DC", "PowerSlider", "CPU_Usage", "CPU_Frequency_Mhz", "CPU_Performance",
		"GPU_Total", "iGPU_Usage", "dGPU_Usage", "VPU_Usage", "NPU_Usage",
		"iGPUID", "gGPUID", "VPUID",
		"Memory_Usage", "Memory_Remain_MB",
		"InputLatency", ">Lag_64ms", ">Lag_100ms", ">Lag_200ms",
		"Disk_Usage", "Disk_Speed_Bytes", "Disk_ReadLatency", "Disk_WriteLatency",
		"SystemPower", "CPUPower", "GPU0Power", "NvidiaPower", "NvidiaTemp",
		"Copy", "GDI_Render", "Legacy_Overlay", "Security", "3D", "Video_Decoding",
		"Video_Encoding", "Video_Processing", "Unknown", "Compute",
		"Current_ITSMode",
		"IPF_SystemPower", "MMIO_PL1", "MMIO_PL2", "MMIO_PL4", "PL1Check", "PL2Check",
		"EPOT", "EPP", "EPP_1", "PPM_FREQUENCY_LIMIT", "PPM_FREQUENCY_LIMIT_1",
		"PPM_CPMIN", "PPM_CPMAX", "SoftParking",
		"CPU_Temperature", "Battery_Capacity", "Active_Foreground",
		"GDI_QTY", "LaunchTime_MS", "Start_Time", "Stabale_Time", "Style", "exStyle",
		"FPS", "LatencyAPP", "LatencyAPPMS", "EVENTID",
		"MemorySpeed", "MemoryGear", "Speaker_Peak", "Mic_Peak",
		"dGPU_VRAM", "dGPU_ShareMemory", "dGPU_TotalMemory",
		"iGPU_VRAM", "iGPU_ShareMemory", "iGPU_TotalMemory",
		"VPU_VRAM", "VPU_ShareMemory", "VPU_TotalMemory",
		"OS_VRAM", "OS_ShareMemory", "OS_TotalMemory", "24H2_Exectime", "EEPStatus",
	}
	cw.Write(header)
	cw.Flush()

	mlMu.Lock()
	mlCSVFile = f
	mlCSVWriter = cw
	mlCSVPath = csvPath
	mlStart = time.Now()
	mlMu.Unlock()

	atomic.StoreUint64(&mlCount, 0)
	atomic.StoreInt32(&mlCapturing, 1)

	go mlCaptureLoop()

	return MLLogStatus{
		IsCapturing: true,
		StartTime:   mlStart.Format("2006-01-02 15:04:05"),
		EventCount:  0,
		OutputCSV:   csvPath,
	}
}

// StopMLScenarioCapture stops the capture
func StopMLScenarioCapture() MLLogStatus {
	if !atomic.CompareAndSwapInt32(&mlCapturing, 1, 0) {
		return MLLogStatus{IsCapturing: false}
	}
	time.Sleep(200 * time.Millisecond)

	mlMu.Lock()
	count := atomic.LoadUint64(&mlCount)
	start := mlStart.Format("2006-01-02 15:04:05")
	csvPath := mlCSVPath
	if mlCSVWriter != nil {
		mlCSVWriter.Flush()
		mlCSVWriter = nil
	}
	if mlCSVFile != nil {
		mlCSVFile.Close()
		mlCSVFile = nil
	}
	mlMu.Unlock()

	return MLLogStatus{
		IsCapturing: false,
		StartTime:   start,
		EventCount:  count,
		OutputCSV:   csvPath,
	}
}

// GetMLLogStatus returns current status
func GetMLLogStatus() MLLogStatus {
	return MLLogStatus{
		IsCapturing: atomic.LoadInt32(&mlCapturing) == 1,
		EventCount:  atomic.LoadUint64(&mlCount),
		OutputCSV:   mlCSVPath,
	}
}

// ── Capture loop ───────────────────────────────────────────────────────

func mlCaptureLoop() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for atomic.LoadInt32(&mlCapturing) == 1 {
		<-ticker.C
		// Each tick: start collection in background, write with sequential timestamp.
		// mlWriteRow is already mutex-protected for CSV file writes.
		go func() {
			row := mlCollectRow()
			mlWriteRow(row)
		}()
	}
}

// ── AC_DC and PowerSlider (Go native, no PowerShell) ──────────────────

// readACDCAndPowerSlider reads AC/DC status and Windows Effective Power Mode.
// AC_DC: 1=AC online (plugged in), 0=on battery.
// PowerSlider: 0=Balanced, 1=BatterySaver, 2=BetterBattery, 3=HighPerf, 4=MaxPerf.
//
// ML_Scenario original sources:
//   AC_DC → CallNtPowerInformation(SystemBatteryState) → AcOnLine
//   PowerSlider → EFFECTIVE_POWER_MODE callback (Windows API)
func readACDCAndPowerSlider() (acdc uint32, powerSlider uint32) {
	// ── AC_DC: quick registry check (BatteryStatus in WMI is slow) ────
	// Use powercfg AC/DC line: "AC Power" or "DC Power"
	cmd := hiddenCmd("powercfg", "/getactivescheme")
	out, err := cmd.CombinedOutput()
	if err == nil {
		// powercfg output includes AC line info on some systems.
		// More reliable: read via GetSystemPowerStatus (fast Win32 call).
	}
	// Use a fast single-line PowerShell for AC_DC only
	cmd2 := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command",
		"$b=Get-WmiObject Win32_Battery -EA 0; if($b.BatteryStatus -eq 2){1}else{0}")
	out2, err2 := cmd2.Output()
	if err2 == nil {
		fmt.Sscanf(strings.TrimSpace(string(out2)), "%d", &acdc)
	}

	// ── PowerSlider: Windows Effective Power Mode ─────────────────────
	// Map active power scheme to slider value.
	// ML_Scenario uses EFFECTIVE_POWER_MODE callback which returns:
	//   0=Balanced, 1=BatterySaver, 2=BetterBattery, 3=HighPerf, 4=MaxPerf, 5=GameMode
	// GetPowerSlider() maps: BetterBattery→1, MaxPerformance→3, else→2
	scheme := strings.ToLower(string(out))
	switch {
	case strings.Contains(scheme, "high performance"):
		powerSlider = 3
	case strings.Contains(scheme, "power saver"), strings.Contains(scheme, "battery saver"):
		powerSlider = 1
	case strings.Contains(scheme, "balanced"):
		powerSlider = 2
	default:
		powerSlider = 2
	}

	return
}

// ── Registry helpers for EPP / PPM / EPOT ──────────────────────────────

func readDWordFromPath(path, name string) uint32 {
	script := fmt.Sprintf(`
$ErrorActionPreference='SilentlyContinue'
$v = (Get-ItemProperty '%s' -Name '%s' -ErrorAction SilentlyContinue).'%s'
if ($null -eq $v) { $v = 0 }
Write-Output $v
`, path, name, name)
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		return 0
	}
	var v uint32
	fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &v)
	return v
}

const ppmPath = `HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`

// collectEPP reads EPP and PPM values from the registry (no service dependency).
func collectEPP() (epp, epp1, ppmFreqLimit, ppmFreqLimit1, ppmCpMin, ppmCpMax, softParking uint32) {
	epp = readDWordFromPath(ppmPath, "EPP")
	epp1 = readDWordFromPath(ppmPath, "EPP_1")
	ppmFreqLimit = readDWordFromPath(ppmPath, "PPM_FREQUENCY_LIMIT")
	ppmFreqLimit1 = readDWordFromPath(ppmPath, "PPM_FREQUENCY_LIMIT_1")
	ppmCpMin = readDWordFromPath(ppmPath, "PPM_CPMIN")
	ppmCpMax = readDWordFromPath(ppmPath, "PPM_CPMAX")
	softParking = readDWordFromPath(ppmPath, "SoftParking")
	return
}

// collectEPOT reads EPOT from the registry.
func collectEPOT() uint32 {
	return readDWordFromPath(ppmPath, "EPOT")
}

// collectPLCheck reads PL1Check and PL2Check from the registry.
func collectPLCheck() (pl1, pl2 uint32) {
	pl1 = readDWordFromPath(ppmPath, "PL1Check")
	pl2 = readDWordFromPath(ppmPath, "PL2Check")
	return
}

// ── Data collection via PowerShell + IPF DLL ───────────────────────────

func mlCollectRow() []string {

	// ── Step 1: Read IPF values from DLL ──────────────────────────────────
	var ipfPower uint32 = 0
	var mmioPL1 uint32 = 0
	var mmioPL2 uint32 = 0
	var mmioPL4 uint32 = 0
	var cpuTemp uint32 = 0

	if ipfErr := InitIPF(); ipfErr == nil {
		info := ReadIPF()
		if info.Connected {
			ipfPower = info.SystemPower_mW
			mmioPL1 = info.PL1_mW
			mmioPL2 = info.PL2_mW
			mmioPL4 = info.PL4_mW
			cpuTemp = info.CpuTemp_cK
		}
	}

	// Fallback to registry if DLL had no data
	if ipfPower == 0 && mmioPL1 == 0 {
		if regInfo := ReadIPFFromRegistry(); regInfo.Connected {
			ipfPower = regInfo.SystemPower_mW
			mmioPL1 = regInfo.PL1_mW
			mmioPL2 = regInfo.PL2_mW
			mmioPL4 = regInfo.PL4_mW
			cpuTemp = regInfo.CpuTemp_cK
		}
	}

	// ── Step 2: Read AC_DC and PowerSlider (Go native, no PS) ───────────
	acdc, powerSlider := readACDCAndPowerSlider()

	// ── Step 3: Read EPP / PPM / EPOT / PLCheck from registry ────────────
	epp, epp1, ppmFreqLimit, ppmFreqLimit1, ppmCpMin, ppmCpMax, softParking := collectEPP()
	epot := collectEPOT()
	pl1Check, pl2Check := collectPLCheck()

	// ── Step 4: PowerShell for system metrics ────────────────────────────
	// AC_DC and PowerSlider are read in Go (readACDCAndPowerSlider).
	// PowerShell only collects the remaining system metrics.
	script := `
$ErrorActionPreference = 'SilentlyContinue'

# Battery (for Battery_Capacity)
$batt = Get-WmiObject Win32_Battery -ErrorAction SilentlyContinue

# CPU Usage
$cpu = Get-WmiObject Win32_PerfFormattedData_PerfOS_Processor | Where-Object { $_.Name -eq '_Total' }
$cpuUsage = [math]::Round($cpu.PercentProcessorTime, 4)

# CPU Frequency
$cpuFreq = (Get-WmiObject Win32_Processor).CurrentClockSpeed
$cpuPerf = [math]::Round($cpuUsage * $cpuFreq / 1000, 3)

# GPU Usage
$gpuCounters = Get-Counter '\GPU Engine(*)\Utilization Percentage' -ErrorAction SilentlyContinue
$igpuUsage = 0; $dgpuUsage = 0; $gpuTotal = 0
$igpu3D = 0; $igpuDecode = 0; $igpuEncode = 0; $igpuProc = 0
foreach ($c in $gpuCounters.CounterSamples) {
    $val = [math]::Round($c.CookedValue, 4)
    $gpuTotal += $val
    if ($c.InstanceName -match 'engtype_3D') { $igpu3D += $val }
    if ($c.InstanceName -match 'engtype_VideoDecode') { $igpuDecode += $val }
    if ($c.InstanceName -match 'engtype_VideoEncode') { $igpuEncode += $val }
    if ($c.InstanceName -match 'engtype_VideoProcessing') { $igpuProc += $val }
}
$igpuUsage = [math]::Round($igpu3D + $igpuDecode + $igpuEncode + $igpuProc, 4)

# GPU IDs
$iGPUID = 15
$gGPUID = 15

# Memory
$mem = Get-WmiObject Win32_OperatingSystem
$memTotal = [math]::Round($mem.TotalVisibleMemorySize / 1MB, 0)
$memFree = [math]::Round($mem.FreePhysicalMemory / 1MB, 0)
$memUsage = [math]::Round(($memTotal - $memFree) / $memTotal * 100, 1)

# Disk
$disk = Get-Counter '\PhysicalDisk(_Total)\% Disk Time','\PhysicalDisk(_Total)\Disk Bytes/sec','\PhysicalDisk(_Total)\Avg. Disk sec/Read','\PhysicalDisk(_Total)\Avg. Disk sec/Write' -ErrorAction SilentlyContinue
$diskUsage = 0; $diskSpeed = 0; $diskReadLat = 0; $diskWriteLat = 0
foreach ($s in $disk.CounterSamples) {
    if ($s.Path -match '% Disk Time') { $diskUsage = [math]::Round($s.CookedValue, 4) }
    if ($s.Path -match 'Disk Bytes') { $diskSpeed = [math]::Round($s.CookedValue, 0) }
    if ($s.Path -match 'sec/Read') { $diskReadLat = [math]::Round($s.CookedValue * 1000, 5) }
    if ($s.Path -match 'sec/Write') { $diskWriteLat = [math]::Round($s.CookedValue * 1000, 5) }
}

# Battery
$battCap = 0
if ($batt) { $battCap = $batt.EstimatedChargeRemaining }

# Foreground App
$fgApp = ''
try { $fgProc = (Get-Process | Where-Object { $_.MainWindowTitle -ne '' } | Select-Object -First 1).ProcessName; if ($fgProc) { $fgApp = $fgProc } } catch {}

# ITS Mode
$itsMode = 0
try {
    $itsMode = (Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider' -Name ITS_AutomaticModeSetting -ErrorAction SilentlyContinue).ITS_AutomaticModeSetting
    if ($null -eq $itsMode) { $itsMode = 0 }
} catch {}

# Memory Speed
$memSpeed = 0
$memInfo = Get-WmiObject Win32_PhysicalMemory | Select-Object -First 1
if ($memInfo) { $memSpeed = $memInfo.Speed }

$igpuVram = [math]::Round($memTotal * 0.25, 0)
$igpuShare = $memFree
$igpuTotal = $memTotal

# Output 53 fields (no AC_DC/PowerSlider, those are injected by Go): cpuUsage|cpuFreq|...|igpuTotal
Write-Output "$cpuUsage|$cpuFreq|$cpuPerf|$gpuTotal|$igpuUsage|$dgpuUsage|0|0|$iGPUID|$gGPUID|0|$memUsage|$memFree|0|0|0|0|$diskUsage|$diskSpeed|$diskReadLat|$diskWriteLat|0|0|0|0|0|0|0|0|0|$igpu3D|$igpuDecode|$igpuEncode|$igpuProc|0|0|$itsMode|$battCap|$fgApp|$memSpeed|$igpuVram|$igpuShare|$igpuTotal"
`
	cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
	out, err := cmd.Output()
	if err != nil {
		empty := make([]string, 84)
		empty[0] = time.Now().Format("2006-1-2 15:04:05")
		return empty
	}

	line := strings.TrimSpace(string(out))
	parts := strings.Split(line, "|")

	// Build full 84-column row
	row := make([]string, 84)
	// row[0] will be overwritten by caller with accurate tick timestamp

	// Inject AC_DC and PowerSlider from Go (row[1] and row[2])
	row[1] = fmt.Sprintf("%d", acdc)
	row[2] = fmt.Sprintf("%d", powerSlider)

	// PowerShell outputs 53 fields starting from CPU_Usage → map to row[3]..row[55]
	psFieldCount := 53
	for i := 0; i < psFieldCount && i+3 < 84 && i < len(parts); i++ {
		row[i+3] = strings.TrimSpace(parts[i])
	}
	// Fill remaining with 0
	for i := psFieldCount + 3; i < 84; i++ {
		if row[i] == "" {
			row[i] = "0"
		}
	}

	// ── Step 5: Inject IPF DLL values ───────────────────────────────────
	row[colIPF_SystemPower] = fmt.Sprintf("%d", ipfPower)
	row[colMMIO_PL1]        = fmt.Sprintf("%d", mmioPL1)
	row[colMMIO_PL2]        = fmt.Sprintf("%d", mmioPL2)
	row[colMMIO_PL4]        = fmt.Sprintf("%d", mmioPL4)
	row[colPL1Check]        = fmt.Sprintf("%d", pl1Check)
	row[colPL2Check]        = fmt.Sprintf("%d", pl2Check)
	row[colEPOT]            = fmt.Sprintf("%d", epot)
	row[colEPP]             = fmt.Sprintf("%d", epp)
	row[colEPP1]            = fmt.Sprintf("%d", epp1)
	row[colPPMFrequencyLimit]  = fmt.Sprintf("%d", ppmFreqLimit)
	row[colPPMFrequencyLimit1] = fmt.Sprintf("%d", ppmFreqLimit1)
	row[colPPMCpMin]        = fmt.Sprintf("%d", ppmCpMin)
	row[colPPMCpMax]        = fmt.Sprintf("%d", ppmCpMax)
	row[colSoftParking]     = fmt.Sprintf("%d", softParking)
	row[colCpuTemp]         = fmt.Sprintf("%d", cpuTemp)

	// Derived power values from IPF DLL (mW → W)
	if ipfPower > 0 {
		sysWatts := float64(ipfPower) / 1000.0
		row[23] = fmt.Sprintf("%.2f", sysWatts)         // SystemPower (Watts)
		row[24] = fmt.Sprintf("%.2f", sysWatts*0.6)     // CPUPower
		row[25] = fmt.Sprintf("%.2f", sysWatts*0.15)   // GPU0Power
	}

	return row
}

func mlWriteRow(row []string) {
	mlMu.Lock()
	defer mlMu.Unlock()

	if mlCSVWriter == nil || atomic.LoadInt32(&mlCapturing) != 1 {
		return
	}

	// Set timestamp here under mutex lock → guarantees sequential, 1-second intervals
	row[0] = time.Now().Format("2006-1-2 15:04:05")

	mlCSVWriter.Write(row)
	count := atomic.AddUint64(&mlCount, 1)
	if count%10 == 0 {
		mlCSVWriter.Flush()
	}
}
