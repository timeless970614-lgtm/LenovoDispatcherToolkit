//go:build windows

package backend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
)

// ==================== Types ====================

// ETLCaptureState describes the current state of an ETL capture session
type ETLCaptureState struct {
	IsCapturing bool   `json:"isCapturing"`
	Profile     string `json:"profile"`
	StartTime   string `json:"startTime"`
	Duration    int    `json:"durationSecs"`
	OutputPath  string `json:"outputPath"`
	Status      string `json:"status"`
	Error       string `json:"error"`
}

// ETLProfile describes a single WPR profile
type ETLProfile struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// ETLTraceInfo describes a captured ETL file
type ETLTraceInfo struct {
	Path        string `json:"path"`
	SizeMB      string `json:"sizeMB"`
	CapturedAt  string `json:"capturedAt"`
	Duration    int    `json:"durationSecs"`
	Profile     string `json:"profile"`
	ProfileName string `json:"profileName"`
}

// ETLAnalysisResult is the structured result returned to the UI
type ETLAnalysisResult struct {
	TraceInfo     ETLTraceInfo         `json:"traceInfo"`
	CPU           CPUAnalysis          `json:"cpu"`
	Disk          DiskAnalysis         `json:"disk"`
	Network       NetworkAnalysis      `json:"network"`
	Power         PowerAnalysis        `json:"power"`
	GPU           GPUAnalysis          `json:"gpu"`
	DPCISR        DPCISRAnalysis        `json:"dpcrisr"`
	Profile       ProfileAnalysis      `json:"profile"`
	Summary       string               `json:"summary"`
	RawCSVPath    string               `json:"rawCSVPath"`
	IsElevated    bool                 `json:"isElevated"`
	RawCSVLines   []string             `json:"rawCSVLines"`
}

// CPUAnalysis holds CPU-related trace analysis
type CPUAnalysis struct {
	BusyProcesses   []CPUProcessItem `json:"busyProcesses"`
	CPUUsagePct      string            `json:"cpuUsagePct"`
	ContextSwitches  int64             `json:"contextSwitches"`
	Interrupts       int64             `json:"interrupts"`
	HardIRQs        int64             `json:"hardIrqs"`
	DPCsQueued      int64             `json:"dpcsQueued"`
	 DPCsDropped     int64             `json:"dpcsDropped"`
}

// CPUProcessItem represents a busy process in CPU analysis
type CPUProcessItem struct {
	ProcessName string  `json:"processName"`
	PID         uint32  `json:"pid"`
	CPUPct      float64 `json:"cpuPct"`
	Duration    string  `json:"duration"`
}

// DiskAnalysis holds Disk I/O analysis
type DiskAnalysis struct {
	TopReaders       []DiskIOItem `json:"topReaders"`
	TopWriters       []DiskIOItem `json:"topWriters"`
	TotalReadMB      string       `json:"totalReadMB"`
	TotalWrittenMB   string       `json:"totalWrittenMB"`
	ReadOpsPerSec    string       `json:"readOpsPerSec"`
	WriteOpsPerSec   string       `json:"writeOpsPerSec"`
	AvgLatencyMs     string       `json:"avgLatencyMs"`
}

// DiskIOItem represents a disk I/O item
type DiskIOItem struct {
	ProcessName string `json:"processName"`
	Path        string `json:"path"`
	IOType      string `json:"ioType"`
	Count       int64  `json:"count"`
	SizeMB      string `json:"sizeMB"`
}

// NetworkAnalysis holds Network I/O analysis
type NetworkAnalysis struct {
	TotalSentMB      string        `json:"totalSentMB"`
	TotalRecvMB      string        `json:"totalRecvMB"`
	TCPConnections   int           `json:"tcpConnections"`
	TopConnections   []ConnItem    `json:"topConnections"`
}

// ConnItem represents a TCP/UDP connection item
type ConnItem struct {
	LocalAddr      string `json:"localAddr"`
	RemoteAddr     string `json:"remoteAddr"`
	State          string `json:"state"`
	BytesSent      int64  `json:"bytesSent"`
	BytesReceived  int64  `json:"bytesReceived"`
	ProcessName    string `json:"processName"`
}

// PowerAnalysis holds Power/Energy analysis
type PowerAnalysis struct {
	PlatformIdle      string         `json:"platformIdle"`
	CPUPower          string         `json:"cpuPower"`
	PackagePower      string         `json:"packagePower"`
	GPUPower          string         `json:"gpuPower"`
	S0ixDuration      string         `json:"s0ixDuration"`
	S0ixTransitions   int            `json:"s0ixTransitions"`
	ProcessorFreqMHz  string         `json:"processorFreqMHz"`
}

// GPUAnalysis holds GPU activity analysis
type GPUAnalysis struct {
	GPUEngineUtilPct  string          `json:"gpuEngineUtilPct"`
	GPUMemoryUsedMB   string          `json:"gpuMemoryUsedMB"`
	GPUContextCreated int            `json:"gpuContextCreated"`
	GPUEngines        []GPUEngineItem `json:"gpuEngines"`
}

// GPUEngineItem represents a GPU engine item
type GPUEngineItem struct {
	EngineName string `json:"engineName"`
	UtilPct    string `json:"utilPct"`
}

// DPCISRAnalysis holds DPC/ISR latency analysis
type DPCISRAnalysis struct {
	HighDPCLatencyProcs []LatencyItem `json:"highDpcLatencyProcs"`
	HighISRLatencyProcs []LatencyItem `json:"highIsrLatencyProcs"`
	AvgDPCMs            string        `json:"avgDpcMs"`
	AvgISRMs            string        `json:"avgIsrMs"`
	MaxDPCMs            string        `json:"maxDpcMs"`
	MaxISRMs            string        `json:"maxIsrMs"`
}

// LatencyItem represents a high-latency item
type LatencyItem struct {
	ProcessName string `json:"processName"`
	Module      string `json:"module"`
	MaxLatencyMs string `json:"maxLatencyMs"`
	AvgLatencyMs string `json:"avgLatencyMs"`
	Count       int64  `json:"count"`
}

// ProfileAnalysis holds which providers were active
type ProfileAnalysis struct {
	ProfileName     string   `json:"profileName"`
	ProvidersActive []string `json:"providersActive"`
}

// ==================== Constants ====================

var wprPath = `C:\Windows\System32\wpr.exe`
var xperfPath = `C:\Program Files (x86)\Windows Kits\10\Windows Performance Toolkit\xperf.exe`
var tracerptPath = `C:\Windows\System32\tracerpt.exe`

var etlOutputDir = `C:\Users\Public\ETL_Traces`
var captureState = ETLCaptureState{IsCapturing: false}
var captureMu sync.Mutex // protects wpr -stop calls from goroutine and manual stop

// ==================== Helpers ====================

func init() {
	os.MkdirAll(etlOutputDir, 0755)
}

func runHidden(name string, args ...string) (string, error) {
	cmd := hiddenCmd(name, args...)
	cmd.Dir = `C:\Windows\System32`
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func isElevated() bool {
	return windows.GetCurrentProcessToken().IsElevated()
}

// IsElevated is the exported version for the app layer
func IsElevated() bool {
	return isElevated()
}

// ==================== ETL Profiles ====================

// GetETLProfiles returns available WPR profiles for the UI
func GetETLProfiles() []ETLProfile {
	return []ETLProfile{
		{ID: "GeneralProfile", Name: "General", Description: "CPU, Disk, Network, Power, GPU - First level triage", Category: "All"},
		{ID: "CPU", Name: "CPU Usage", Description: "Per-process CPU usage, context switches, interrupts", Category: "CPU"},
		{ID: "DiskIO", Name: "Disk I/O", Description: "Disk read/write activity, file I/O, minifilter", Category: "Storage"},
		{ID: "Network", Name: "Network", Description: "Network send/receive activity, TCP connections", Category: "Network"},
		{ID: "Power", Name: "Power", Description: "C-states, S0ix, P-states, processor frequency", Category: "Power"},
		{ID: "GPU", Name: "GPU", Description: "GPU engine utilization, context activity, memory", Category: "GPU"},
		{ID: "Heap", Name: "Heap", Description: "Heap allocation and usage analysis", Category: "Memory"},
		{ID: "Pool", Name: "Pool", Description: "Kernel pool nonpaged/paged usage", Category: "Memory"},
	}
}

// ==================== ETL Capture ====================

// StartETLCapture starts a WPR trace with the given profile using a visible cmd window.
// The cmd window runs wpr so the user can see the output.
// If durationSecs > 0, the trace auto-stops after that many seconds.
// duration=0 means the trace runs until the user clicks Stop Trace in the UI.
func ForceStopWPR() {
	// Cancel any running WPR sessions to ensure a clean start.
	logFile := etlOutputDir + "\\wpr_cancel_log.txt"
	os.MkdirAll(etlOutputDir, 0755)

	// Call wpr -cancel directly (no cmd.exe layer)
	cmd := wprCmd("wpr.exe", "-cancel", "-instancename", "dispatcher_trace")
	out, err := cmd.CombinedOutput()

	f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString(fmt.Sprintf("[%s] ForceStopWPR cancel: err=%v out=%s\n",
		time.Now().Format("15:04:05.000"), err, string(out)))
	f.Close()

	// Also try without instancename to cancel any session
	cmd2 := wprCmd("wpr.exe", "-cancel")
	out2, _ := cmd2.CombinedOutput()
	f2, _ := os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
	f2.WriteString(fmt.Sprintf("[%s] ForceStopWPR cancel (no instancename): out=%s\n",
		time.Now().Format("15:04:05.000"), string(out2)))
	f2.Close()
}

func StartETLCapture(profile string, durationSecs int) ETLCaptureState {
	// Clean up any stale sessions from previous runs first
	ForceStopWPR()

	if !isElevated() {
		return ETLCaptureState{
			IsCapturing: false,
			Status:      "error",
			Error:       "Administrator privileges required. Please run the application as Administrator.",
		}
	}
	if captureState.IsCapturing {
		return ETLCaptureState{
			IsCapturing: true,
			Status:      "error",
			Error:       "A trace is already in progress.",
		}
	}

	outputFile := filepath.Join(etlOutputDir, fmt.Sprintf("trace_%s_%s.etl",
		profile, time.Now().Format("20060102_150405")))

	// Build the wpr pipeline: start trace and wait in a visible cmd window
	// The cmd window shows wpr -start output so user knows it started
	sleepSec := durationSecs
	if sleepSec <= 0 {
		sleepSec = 30
	}
	// WPR start: use memory mode (default, no -filemode) to avoid temp dir issues.
	// WPR start: memory mode (default, no -filemode) with named session.
	// -instancename dispatcher_trace: name the session so -stop can find it.
	// Use wprCmd (CREATE_NEW_CONSOLE) for reliable ETW session management.
	cmd := wprCmd("wpr.exe", "-start", profile, "-instancename", "dispatcher_trace")
	startOut, startErr := cmd.CombinedOutput()

	// Log the start result
	logFile := etlOutputDir + "\\wpr_goroutine_log.txt"
	f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	f.WriteString(fmt.Sprintf("[%s] wpr -start: profile=%s -instancename dispatcher_trace\n", time.Now().Format("15:04:05.000"), profile))
	f.WriteString(fmt.Sprintf("[%s] wpr -start exit: %v\n", time.Now().Format("15:04:05.000"), startErr))
	f.WriteString(fmt.Sprintf("[%s] wpr -start output:\n%s\n", time.Now().Format("15:04:05.000"), string(startOut)))
	f.Close()

	if startErr != nil {
		return ETLCaptureState{
			IsCapturing: false,
			Status:      "error",
			Error:       fmt.Sprintf("WPR start failed: %s", string(startOut)),
		}
	}

	// Start succeeded - launch visible cmd window for countdown display
	countdownScript := fmt.Sprintf(`echo WPR trace started with profile %s. Auto-stopping in %d seconds... && timeout /t %d`, profile, sleepSec, sleepSec)
	countdownCmd := visibleCmd("cmd.exe", "/C", countdownScript+" & pause")
	countdownCmd.Dir = etlOutputDir
	countdownCmd.Start() // Don't wait - let user see the countdown

	captureState = ETLCaptureState{
		IsCapturing: true,
		Profile:     profile,
		StartTime:   time.Now().Format("15:04:05"),
		Duration:    durationSecs,
		OutputPath:  outputFile,
		Status:      "recording",
	}

	// Background: auto-stop after duration ends and update UI state
	go func() {
		logFile := etlOutputDir + "\\wpr_goroutine_log.txt"
		f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		f.WriteString(fmt.Sprintf("[%s] goroutine started, will stop in %ds\n", time.Now().Format("15:04:05.000"), sleepSec))
		f.Close()

		time.Sleep(time.Duration(sleepSec) * time.Second)

		f, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
		f.WriteString(fmt.Sprintf("[%s] sleep done, taking lock...\n", time.Now().Format("15:04:05.000")))
		f.Close()

		captureMu.Lock()
		if !captureState.IsCapturing {
			captureMu.Unlock()
			f, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
			f.WriteString(fmt.Sprintf("[%s] not capturing, returning\n", time.Now().Format("15:04:05.000")))
			f.Close()
			return
		}
		outputPath := captureState.OutputPath
		f.WriteString(fmt.Sprintf("[%s] capturing=true, path=%s, running wpr -stop\n", time.Now().Format("15:04:05.000"), outputPath))
		f.Close()

		captureState = ETLCaptureState{IsCapturing: false, Status: "auto-stopped", OutputPath: outputPath}
		captureMu.Unlock()

		// Run wpr -stop outside the lock so UI doesn't block
		// WPR stop syntax: wpr -stop <output.etl> -force -instancename <session>
		// -force: skip warning about non-.etl extension
		// -instancename must match the -start value
		// Use wprCmd for reliable ETW session management
		cmd := wprCmd("wpr.exe", "-stop", outputPath, "-force", "-instancename", "dispatcher_trace")
		f, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
		f.WriteString(fmt.Sprintf("[%s] wpr -stop: path=%s -force -instancename dispatcher_trace\n",
			time.Now().Format("15:04:05.000"), outputPath))
		f.Close()
		out, err := cmd.CombinedOutput()
		f, _ = os.OpenFile(logFile, os.O_APPEND|os.O_WRONLY, 0644)
		f.WriteString(fmt.Sprintf("[%s] wpr -stop done. err=%v output=%s\n", time.Now().Format("15:04:05.000"), err, string(out)))
		f.Close()
	}()

	return captureState
}

// StopETLCapture stops the running WPR trace and returns trace info
func StopETLCapture() ETLTraceInfo {
	captureMu.Lock()
	defer captureMu.Unlock()

	if !captureState.IsCapturing {
		return ETLTraceInfo{}
	}

	outputPath := captureState.OutputPath

	// Run wpr -stop and -merge silently (goroutine may have already done this)
	// Use wprCmd for reliable ETW session management
	cmd := wprCmd("wpr.exe", "-stop", outputPath, "-force", "-instancename", "dispatcher_trace")
	cmd.Run() // ignore error — may already be stopped

	captureState = ETLCaptureState{IsCapturing: false, Status: "stopped"}

	// Get file info
	var infoSizeMB, capturedAt string
	if fi, err := os.Stat(outputPath); err == nil {
		infoSizeMB = fmt.Sprintf("%.1f", float64(fi.Size())/1024/1024)
		capturedAt = fi.ModTime().Format("2006-01-02 15:04:05")
	}

	duration := 0
	if captureState.StartTime != "" {
		if t, err := time.Parse("15:04:05", captureState.StartTime); err == nil {
			duration = int(time.Since(t).Seconds())
		}
	}

	profileName := captureState.Profile
	for _, p := range GetETLProfiles() {
		if p.ID == captureState.Profile {
			profileName = p.Name
			break
		}
	}

	ti := ETLTraceInfo{
		Path:        outputPath,
		SizeMB:      infoSizeMB,
		CapturedAt:  capturedAt,
		Duration:    duration,
		Profile:     captureState.Profile,
		ProfileName: profileName,
	}

	return ti
}

// GetETLCaptureStatus returns current capture state
func GetETLCaptureStatus() ETLCaptureState {
	return captureState
}

// ==================== ETL Analysis ====================

// AnalyzeETLFile parses an ETL file and returns structured analysis
func AnalyzeETLFile(etlPath string) ETLAnalysisResult {
	result := ETLAnalysisResult{
		IsElevated: isElevated(),
		TraceInfo:   ETLTraceInfo{Path: etlPath},
	}

	// Get file size
	if fi, err := os.Stat(etlPath); err == nil {
		result.TraceInfo.SizeMB = fmt.Sprintf("%.1f", float64(fi.Size())/1024/1024)
		result.TraceInfo.CapturedAt = fi.ModTime().Format("2006-01-02 15:04:05")
	}

	if !isElevated() {
		result.Summary = "Administrator privileges required for full ETL analysis. Run as Admin for complete results."
		return result
	}

	csvDir := filepath.Dir(etlPath)
	csvBase := filepath.Base(etlPath)
	csvName := strings.TrimSuffix(csvBase, filepath.Ext(csvBase))
	summaryCSV := filepath.Join(csvDir, csvName+"_summary.csv")
	_ = filepath.Join(csvDir, csvName+"_detail.csv") // reserved for future detail export

	// Run tracerpt to export CSV
	tracerptArgs := []string{
		etlPath,
		"-o", summaryCSV,
		"-of", "CSV",
		"-y",
	}
	out, err := runHidden(tracerptPath, tracerptArgs...)
	if err != nil {
		result.Summary = fmt.Sprintf("tracerpt failed: %s\nOutput: %s", err.Error(), out)
		return result
	}

	// Try to read summary CSV
	if data, err := os.ReadFile(summaryCSV); err == nil {
		lines := strings.Split(string(data), "\n")
		result.RawCSVLines = takeLines(lines, 100)
		result = parseSummaryCSV(result, string(data))
	}

	// Try xperf for CPU profile analysis
	if runtime.GOARCH == "amd64" {
		result = runXperfCPUDump(result, etlPath)
	}

	return result
}

func takeLines(lines []string, n int) []string {
	if len(lines) > n {
		return lines[:n]
	}
	return lines
}

// runXperfCPUDump runs xperf to get CPU stack data
func runXperfCPUDump(result ETLAnalysisResult, etlPath string) ETLAnalysisResult {
	csvDir := filepath.Dir(etlPath)
	csvName := strings.TrimSuffix(filepath.Base(etlPath), filepath.Ext(etlPath))
	cpuCSV := filepath.Join(csvDir, csvName+"_cpu.csv")

	// xperf -i <etl> -o <csv> -t -pretty  (text mode, no symbol resolve needed)
	args := []string{
		"-i", etlPath,
		"-o", cpuCSV,
		"-d", // dump as CSV
	}
	out, err := runHidden(xperfPath, args...)
	if err != nil {
		// xperf not critical - tracerpt already gave us summary
		_ = out
		return result
	}

	if data, err := os.ReadFile(cpuCSV); err == nil {
		result = parseCPUCSV(result, string(data))
	}

	return result
}

// parseSummaryCSV parses the tracerpt summary CSV and fills the result
func parseSummaryCSV(result ETLAnalysisResult, csvData string) ETLAnalysisResult {
	scanner := bufio.NewScanner(strings.NewReader(csvData))
	var headers []string
	var rows [][]string

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		// Simple CSV parse (handles quoted fields)
		fields, _ := parseCSVRow(line)
		if headers == nil {
			headers = fields
		} else {
			rows = append(rows, fields)
		}
	}

	// Find column indices
	col := make(map[string]int)
	for i, h := range headers {
		col[strings.ToLower(strings.TrimSpace(h))] = i
	}

	// Parse based on the "Name" column (common in WPA summary CSVs)
	var nameIdx, valueIdx int
	if v, ok := col["name"]; ok {
		nameIdx = v
	} else {
		return result
	}
	if v, ok := col["value"]; ok {
		valueIdx = v
	} else if v, ok := col["count"]; ok {
		valueIdx = v
	} else if v, ok := col["duration"]; ok {
		valueIdx = v
	} else {
		valueIdx = 1
	}

	var totalCPUPct float64
	var netSent, netRecv, cpuPower float64

	for _, row := range rows {
		if len(row) <= nameIdx || len(row) <= valueIdx {
			continue
		}
		name := strings.TrimSpace(strings.ToLower(row[nameIdx]))
		valStr := strings.TrimSpace(row[valueIdx])

		switch {
		case strings.Contains(name, "processor") && strings.Contains(name, "%"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				totalCPUPct = v
				result.CPU.CPUUsagePct = fmt.Sprintf("%.1f%%", v)
			}
		case strings.Contains(name, "cpu") && strings.Contains(name, "%"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				result.CPU.CPUUsagePct = fmt.Sprintf("%.1f%%", v)
			}
		case strings.Contains(name, "disk") && strings.Contains(name, "read") && strings.Contains(name, "mb"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				result.Disk.TotalReadMB = fmt.Sprintf("%.1f MB", v)
			}
		case strings.Contains(name, "disk") && strings.Contains(name, "write") && strings.Contains(name, "mb"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				result.Disk.TotalWrittenMB = fmt.Sprintf("%.1f MB", v)
			}
		case strings.Contains(name, "network") && strings.Contains(name, "sent"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				netSent = v
				result.Network.TotalSentMB = fmt.Sprintf("%.1f MB", v)
			}
		case strings.Contains(name, "network") && strings.Contains(name, "receive"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				netRecv = v
				result.Network.TotalRecvMB = fmt.Sprintf("%.1f MB", v)
			}
		case strings.Contains(name, "s0ix") || strings.Contains(name, "idle"):
			result.Power.S0ixDuration = valStr
		case strings.Contains(name, "cpu") && strings.Contains(name, "power"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				cpuPower = v
				result.Power.CPUPower = fmt.Sprintf("%.1f W", v)
			}
		case strings.Contains(name, "gpu") && strings.Contains(name, "power"):
			if v, err := strconv.ParseFloat(valStr, 64); err == nil {
				result.Power.GPUPower = fmt.Sprintf("%.1f W", v)
			}
		case strings.Contains(name, "context") && strings.Contains(name, "switch"):
			if v, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				result.CPU.ContextSwitches = v
			}
		case strings.Contains(name, "interrupt"):
			if v, err := strconv.ParseInt(valStr, 10, 64); err == nil {
				result.CPU.Interrupts = v
			}
		}
	}

	// Build summary
	var parts []string
	if totalCPUPct > 0 {
		parts = append(parts, fmt.Sprintf("CPU: %.1f%%", totalCPUPct))
	}
	if cpuPower > 0 {
		parts = append(parts, fmt.Sprintf("CPU Power: %.1fW", cpuPower))
	}
	if netSent > 0 || netRecv > 0 {
		parts = append(parts, fmt.Sprintf("Network: ↓%.1fMB ↑%.1fMB", netRecv, netSent))
	}
	if len(parts) > 0 {
		result.Summary = strings.Join(parts, " | ")
	} else {
		result.Summary = "Trace captured. Open in WPA for detailed analysis: " + filepath.Base(result.TraceInfo.Path)
	}

	return result
}

// parseCPUCSV parses xperf CPU CSV and fills CPU analysis
func parseCPUCSV(result ETLAnalysisResult, csvData string) ETLAnalysisResult {
	scanner := bufio.NewScanner(strings.NewReader(csvData))
	var headers []string
	rowCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		fields, _ := parseCSVRow(line)
		if headers == nil {
			headers = fields
			continue
		}
		rowCount++
		if rowCount > 2000 {
			break // limit processing
		}

		// Try to find process name and CPU% columns
		var procName string
		var cpuPct float64
		for i, h := range headers {
			h = strings.ToLower(strings.TrimSpace(h))
			if len(fields) <= i {
				continue
			}
			val := strings.TrimSpace(fields[i])
			if h == "process" || h == "name" || h == "image" {
				procName = val
			}
			if (h == "%cpu" || h == "cpu" || h == "%") && val != "" && val != "0" {
				if v, err := strconv.ParseFloat(val, 64); err == nil {
					cpuPct = v
				}
			}
		}
		if procName != "" && cpuPct > 0 && len(result.CPU.BusyProcesses) < 10 {
			result.CPU.BusyProcesses = append(result.CPU.BusyProcesses, CPUProcessItem{
				ProcessName: procName,
				CPUPct:      cpuPct,
			})
		}
	}

	// Sort by CPU%
	sort.Slice(result.CPU.BusyProcesses, func(i, j int) bool {
		return result.CPU.BusyProcesses[i].CPUPct > result.CPU.BusyProcesses[j].CPUPct
	})
	if len(result.CPU.BusyProcesses) > 10 {
		result.CPU.BusyProcesses = result.CPU.BusyProcesses[:10]
	}

	return result
}

// parseCSVRow parses a single CSV row (handles basic quoted fields)
func parseCSVRow(line string) ([]string, error) {
	var fields []string
	cur := ""
	inQuote := false
	for i := 0; i < len(line); i++ {
		c := line[i]
		switch c {
		case '"':
			if inQuote && i+1 < len(line) && line[i+1] == '"' {
				cur += "\""
				i++
			} else {
				inQuote = !inQuote
			}
		case ',':
			if !inQuote {
				fields = append(fields, cur)
				cur = ""
			} else {
				cur += ","
			}
		default:
			cur += string(c)
		}
	}
	fields = append(fields, cur)
	return fields, nil
}

// GetETLTraceList returns list of previously captured ETL traces
func GetETLTraceList() []ETLTraceInfo {
	var traces []ETLTraceInfo
	entries, err := os.ReadDir(etlOutputDir)
	if err != nil {
		return nil
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".etl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		profileName := "General"
		name := strings.ToLower(e.Name())
		if strings.Contains(name, "cpu") {
			profileName = "CPU"
		} else if strings.Contains(name, "disk") {
			profileName = "DiskIO"
		} else if strings.Contains(name, "network") {
			profileName = "Network"
		} else if strings.Contains(name, "power") {
			profileName = "Power"
		} else if strings.Contains(name, "gpu") {
			profileName = "GPU"
		}
		traces = append(traces, ETLTraceInfo{
			Path:        filepath.Join(etlOutputDir, e.Name()),
			SizeMB:      fmt.Sprintf("%.1f", float64(info.Size())/1024/1024),
			CapturedAt:  info.ModTime().Format("2006-01-02 15:04:05"),
			ProfileName: profileName,
		})
	}
	sort.Slice(traces, func(i, j int) bool {
		return traces[i].CapturedAt > traces[j].CapturedAt
	})
	return traces
}