//go:build windows

package backend

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
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

// AnalysisStep represents one analysis stage with progress
// ETLAnalysisStep represents one analysis stage with progress
type ETLAnalysisStep struct {
	Step       int    `json:"step"`
	Name       string `json:"name"`
	Status     string `json:"status"` // "done", "running", "pending", "error"
	Detail     string `json:"detail"`
}

// ETLEventSummary holds event type enumeration results
type ETLEventSummary struct {
	EventName string `json:"eventName"`
	Count     int    `json:"count"`
	Category  string `json:"category"` // Process, Disk, Network, Power, GPU, Other
}

// ETLIssueItem represents a searched issue
type ETLIssueItem struct {
	Keyword    string `json:"keyword"`
	FoundCount int    `json:"foundCount"`
	Samples    []string `json:"samples"` // first 3 matched lines (truncated)
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
	Steps         []ETLAnalysisStep      `json:"steps"`
	EventTypes    []ETLEventSummary      `json:"eventTypes"`
	Issues        []ETLIssueItem         `json:"issues"`
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
		{ID: "GeneralProfile", Name: "General", Description: "CPU, Disk, Power, GPU - First level triage", Category: "All"},
		{ID: "CPU", Name: "CPU Usage", Description: "Per-process CPU usage, context switches, interrupts", Category: "CPU"},
		{ID: "DiskIO", Name: "Disk I/O", Description: "Disk read/write activity, file I/O, minifilter", Category: "Storage"},
		{ID: "Power", Name: "Power", Description: "C-states, S0ix, P-states, processor frequency", Category: "Power"},
		{ID: "GPU", Name: "GPU", Description: "GPU engine utilization, context activity, memory", Category: "GPU"},
	}
}

// ==================== ETL Capture ====================

// StartETLCapture starts a WPR trace with the given profile using a visible cmd window.
// The cmd window runs wpr so the user can see the output.
// If durationSecs > 0, the trace auto-stops after that many seconds.
// duration=0 means the trace runs until the user clicks Stop Trace in the UI.
func ForceStopWPR() {
	// ForceStopWPR cancels any running WPR sessions to ensure a clean start.
	// Since the WPR ETW subsystem runs in a detached process tree, a single -cancel
	// may not fully clean up before the next -start races in. We retry up to 3 times
	// with a short pause between each attempt, then verify with -status before
	// returning so the caller (StartETLCapture) can safely call -start.
	logFile := etlOutputDir + "\\wpr_cancel_log.txt"
	os.MkdirAll(etlOutputDir, 0755)

	cancelSession := func(name string, args ...string) (string, error) {
		cmd := wprCmd(name, args...)
		out, err := cmd.CombinedOutput()
		return string(out), err
	}

	log := func(msg string) {
		f, _ := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
		f.WriteString(fmt.Sprintf("[%s] %s\n", time.Now().Format("15:04:05.000"), msg))
		f.Close()
	}

	// Retry cancel up to 3 times — the ETW subsystem may need a moment to fully stop
	for attempt := 1; attempt <= 3; attempt++ {
		out1, err1 := cancelSession("wpr.exe", "-cancel", "-instancename", "dispatcher_trace")
		log(fmt.Sprintf("ForceStopWPR attempt %d: cancel dispatcher_trace err=%v out=%s", attempt, err1, out1))

		out2, err2 := cancelSession("wpr.exe", "-cancel")
		log(fmt.Sprintf("ForceStopWPR attempt %d: cancel (all) err=%v out=%s", attempt, err2, out2))

		// Give the ETW subsystem a moment to fully tear down the session
		time.Sleep(500 * time.Millisecond)

		// Verify — if -status says not recording, we're done
		statusOut, _ := cancelSession("wpr.exe", "-status")
		log(fmt.Sprintf("ForceStopWPR attempt %d: status=%s", attempt, statusOut))
		if strings.Contains(strings.ToLower(statusOut), "wpr is not recording") {
			log("ForceStopWPR: verified clean (not recording)")
			return
		}
	}

	log(fmt.Sprintf("ForceStopWPR: WARNING — sessions may still be running after 3 attempts. User may need to run 'wpr -cancel' manually or reboot."))
}

func StartETLCapture(profile string, durationSecs int) ETLCaptureState {
	// Clean up any stale sessions from previous runs first
	ForceStopWPR()
	// Extra safety margin: give the ETW subsystem a moment to fully release resources
	time.Sleep(300 * time.Millisecond)

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
		TraceInfo:  ETLTraceInfo{Path: etlPath},
		Steps:      make([]ETLAnalysisStep, 0, 5),
	}

	// ── Step 1: 基本信息确认 ──
	step1 := ETLAnalysisStep{Step: 1, Name: "基本信息确认", Status: "done"}
	if fi, err := os.Stat(etlPath); err == nil {
		result.TraceInfo.SizeMB = fmt.Sprintf("%.1f MB", float64(fi.Size())/1024/1024)
		result.TraceInfo.CapturedAt = fi.ModTime().Format("2006-01-02 15:04:05")
		etlName := filepath.Base(etlPath)
		step1.Detail = fmt.Sprintf("文件: %s | 大小: %.1f MB | 时间: %s",
			etlName, float64(fi.Size())/1024/1024, fi.ModTime().Format("2006-01-02 15:04:05"))
	} else {
		step1.Status = "error"
		step1.Detail = fmt.Sprintf("无法读取文件: %v", err)
	}
	result.Steps = append(result.Steps, step1)

	if !isElevated() {
		step2 := ETLAnalysisStep{Step: 2, Name: "转换格式 (tracerpt)", Status: "error",
			Detail: "需要管理员权限才能进行完整分析。请以管理员身份运行。"}
		result.Steps = append(result.Steps, step2)
		result.Summary = "Administrator privileges required for full ETL analysis. Run as Admin for complete results."
		return result
	}

	// ── Step 2: 转换格式 (tracerpt) ──
	csvDir := filepath.Dir(etlPath)
	csvBase := filepath.Base(etlPath)
	csvName := strings.TrimSuffix(csvBase, filepath.Ext(csvBase))
	summaryCSV := filepath.Join(csvDir, csvName+"_summary.csv")
	result.RawCSVPath = summaryCSV

	tracerptArgs := []string{
		etlPath,
		"-o", summaryCSV,
		"-of", "CSV",
		"-y",
	}
	out, err := runHidden(tracerptPath, tracerptArgs...)
	step2 := ETLAnalysisStep{Step: 2, Name: "转换格式 (tracerpt)", Status: "done"}
	if err != nil {
		step2.Status = "error"
		step2.Detail = fmt.Sprintf("tracerpt 执行失败: %s", err.Error())
		result.Steps = append(result.Steps, step2)
		result.Summary = fmt.Sprintf("tracerpt failed: %s\nOutput: %s", err.Error(), out)
		return result
	}
	// Trim tracerpt verbose output (hundreds of "Event 65535..." lines)
	outLines := strings.Split(strings.TrimSpace(out), "\n")
	if len(outLines) > 5 {
		step2.Detail = fmt.Sprintf("CSV: %s | 输出 %d 行 (显示前5行)\n%s",
			filepath.Base(summaryCSV), len(outLines),
			strings.Join(outLines[:5], "\n"))
	} else {
		step2.Detail = fmt.Sprintf("CSV: %s | 输出: %s", filepath.Base(summaryCSV), strings.TrimSpace(out))
	}
	result.Steps = append(result.Steps, step2)

	// Read CSV
	csvData, csvErr := os.ReadFile(summaryCSV)
	if csvErr != nil {
		step3 := ETLAnalysisStep{Step: 3, Name: "查看事件类型", Status: "error",
			Detail: fmt.Sprintf("无法读取 CSV: %v", csvErr)}
		result.Steps = append(result.Steps, step3)
		return result
	}
	csvStr := string(csvData)
	lines := strings.Split(csvStr, "\n")
	result.RawCSVLines = takeLines(lines, 100)

	// ── Step 3: 查看事件类型 ──
	eventCounts := make(map[string]int)
	csvLineCount := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		csvLineCount++
		if csvLineCount == 1 {
			continue // skip header
		}
		// Try to extract event name from CSV columns
		fields, _ := parseCSVRow(line)
		for _, f := range fields {
			f = strings.TrimSpace(f)
			// Match typical ETW event names like "Process/Start", "DiskIO/Read"
			if strings.Contains(f, "/") && !strings.HasPrefix(f, "\"") && len(f) < 80 {
				eventCounts[f]++
				break
			}
		}
	}

	var eventSummaries []ETLEventSummary
	for name, count := range eventCounts {
		cat := categorizeEvent(name)
		eventSummaries = append(eventSummaries, ETLEventSummary{
			EventName: name, Count: count, Category: cat,
		})
	}
	// Sort by count desc
	sort.Slice(eventSummaries, func(i, j int) bool {
		return eventSummaries[i].Count > eventSummaries[j].Count
	})
	result.EventTypes = eventSummaries

	step3 := ETLAnalysisStep{Step: 3, Name: "查看事件类型", Status: "done"}
	if len(eventSummaries) == 0 {
		step3.Detail = fmt.Sprintf("CSV 共 %d 行数据，未能提取到标准事件类型。可查看 Raw CSV。", csvLineCount-1)
	} else {
		top3 := eventSummaries
		if len(top3) > 3 {
			top3 = top3[:3]
		}
		parts := make([]string, 0, 3)
		for _, ev := range top3 {
			parts = append(parts, fmt.Sprintf("%s (%d)", ev.EventName, ev.Count))
		}
		step3.Detail = fmt.Sprintf("发现 %d 类事件 | TOP: %s | CSV: %d 行",
			len(eventSummaries), strings.Join(parts, ", "), csvLineCount-1)
	}
	result.Steps = append(result.Steps, step3)

	// ── Step 4: 定位问题 (搜索 fail/error/warning/exception/timeout) ──
	keywords := []string{"fail", "error", "warning", "exception", "timeout", "critical"}
	var issues []ETLIssueItem

	for _, kw := range keywords {
		var samples []string
		count := 0
		for _, line := range lines {
			lowerLine := strings.ToLower(line)
			if strings.Contains(lowerLine, kw) {
				count++
				if len(samples) < 3 {
					sample := strings.TrimSpace(line)
					if len(sample) > 120 {
						sample = sample[:117] + "..."
					}
					samples = append(samples, sample)
				}
			}
		}
		if count > 0 {
			issues = append(issues, ETLIssueItem{
				Keyword: kw, FoundCount: count, Samples: samples,
			})
		}
	}
	result.Issues = issues

	step4 := ETLAnalysisStep{Step: 4, Name: "定位问题", Status: "done"}
	if len(issues) == 0 {
		step4.Detail = "未发现 fail/error/warning/exception/timeout 关键字 ✅"
	} else {
		parts := make([]string, 0, len(issues))
		for _, iss := range issues {
			parts = append(parts, fmt.Sprintf("%s: %d 次", iss.Keyword, iss.FoundCount))
		}
		step4.Detail = strings.Join(parts, " | ")
	}
	result.Steps = append(result.Steps, step4)

	// ── Step 5: 图形化分析 (WPA) ──
	wpaAvailable := false
	wpaPaths := []string{
		`C:\Program Files (x86)\Windows Kits\10\Windows Performance Toolkit\wpa.exe`,
		`C:\Program Files\Windows Kits\10\Windows Performance Toolkit\wpa.exe`,
	}
	for _, wp := range wpaPaths {
		if _, err := os.Stat(wp); err == nil {
			wpaAvailable = true
			break
		}
	}
	step5 := ETLAnalysisStep{Step: 5, Name: "图形化分析 (WPA)", Status: "done"}
	if wpaAvailable {
		step5.Detail = "WPA 可用，点击下方按钮打开 WPA 进行图形化分析"
	} else {
		step5.Detail = "WPA 未安装。请安装 Windows Performance Toolkit。"
	}
	result.Steps = append(result.Steps, step5)

	// Also do the original parsing
	result = parseSummaryCSV(result, csvStr)

	// Try xperf for CPU profile analysis
	if runtime.GOARCH == "amd64" {
		result = runXperfCPUDump(result, etlPath)
	}

	return result
}

// categorizeEvent returns a category label for an ETW event name
func categorizeEvent(name string) string {
	name = strings.ToLower(name)
	switch {
	case strings.Contains(name, "process"):
		return "Process"
	case strings.Contains(name, "thread"):
		return "Thread"
	case strings.Contains(name, "disk") || strings.Contains(name, "file"):
		return "Disk"
	case strings.Contains(name, "network") || strings.Contains(name, "tcp") || strings.Contains(name, "udp"):
		return "Network"
	case strings.Contains(name, "power") || strings.Contains(name, "cstate") || strings.Contains(name, "pstate"):
		return "Power"
	case strings.Contains(name, "gpu"):
		return "GPU"
	case strings.Contains(name, "heap") || strings.Contains(name, "memory") || strings.Contains(name, "pool"):
		return "Memory"
	default:
		return "Other"
	}
}

// OpenETLInWPA opens the ETL file in Windows Performance Analyzer
func OpenETLInWPA(etlPath string) string {
	wpaPaths := []string{
		`C:\Program Files (x86)\Windows Kits\10\Windows Performance Toolkit\wpa.exe`,
		`C:\Program Files\Windows Kits\10\Windows Performance Toolkit\wpa.exe`,
	}
	for _, wp := range wpaPaths {
		if _, err := os.Stat(wp); err == nil {
			cmd := exec.Command(wp, etlPath)
			cmd.Start()
			return "WPA launched"
		}
	}
	return "WPA not found. Install Windows Performance Toolkit."
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