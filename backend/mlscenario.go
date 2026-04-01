//go:build windows

package backend

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const dispatcherLogDir = `C:\ProgramData\Lenovo\LenovoDispatcher\Logs`

type MLLogStatus struct {
	IsCapturing bool   `json:"isCapturing"`
	StartTime   string `json:"startTime"`
	EventCount  uint64 `json:"eventCount"`
	OutputFile  string `json:"outputFile"`
	Error       string `json:"error"`
}

var (
	isCapturing int32
	eventCount uint64
	captureMu  sync.Mutex
	logFile    *os.File
	logPath    string
)

func StartMLScenarioCapture() MLLogStatus {
	if atomic.LoadInt32(&isCapturing) == 1 {
		return MLLogStatus{
			IsCapturing: true,
			StartTime:   time.Now().Format("2006-01-02 15:04:05"),
			EventCount: atomic.LoadUint64(&eventCount),
			OutputFile:  logPath,
		}
	}

	if err := os.MkdirAll(dispatcherLogDir, 0755); err != nil {
		return MLLogStatus{Error: "Cannot access log directory: " + err.Error()}
	}

	latest, err := findLatestDispatcherLog()
	if err != nil {
		return MLLogStatus{Error: "No dispatcher log found: " + err.Error()}
	}

	timestamp := time.Now().Format("20060102-150405")
	outPath := filepath.Join(dispatcherLogDir, fmt.Sprintf("MLScenario_%s.LOG", timestamp))

	outFile, err := os.Create(outPath)
	if err != nil {
		return MLLogStatus{Error: "Cannot create output file: " + err.Error()}
	}

	outFile.WriteString(fmt.Sprintf("[%s] ML Scenario Capture Started\n", time.Now().Format("03-01-06-15-04-05")))
	outFile.WriteString(fmt.Sprintf("Tailing: %s\n", filepath.Base(latest)))
	outFile.WriteString("=========================================================\n")
	outFile.Sync()

	captureMu.Lock()
	logFile = outFile
	logPath = outPath
	captureMu.Unlock()

	atomic.StoreUint64(&eventCount, 0)
	atomic.StoreInt32(&isCapturing, 1)
	go captureTailLoop(latest)

	return MLLogStatus{
		IsCapturing: true,
		StartTime:   time.Now().Format("2006-01-02 15:04:05"),
		EventCount:  0,
		OutputFile:  outPath,
	}
}

func StopMLScenarioCapture() MLLogStatus {
	if !atomic.CompareAndSwapInt32(&isCapturing, 1, 0) {
		return MLLogStatus{IsCapturing: false}
	}
	time.Sleep(100 * time.Millisecond)

	captureMu.Lock()
	count := atomic.LoadUint64(&eventCount)
	if logFile != nil {
		logFile.WriteString(fmt.Sprintf("\n[%s] Capture Stopped | Events: %d\n",
			time.Now().Format("03-01-06-15-04-05"), count))
		logFile.Sync()
		logFile.Close()
		logFile = nil
	}
	captureMu.Unlock()

	return MLLogStatus{
		IsCapturing: false,
		EventCount:  count,
		OutputFile:  logPath,
	}
}

func GetMLLogStatus() MLLogStatus {
	return MLLogStatus{
		IsCapturing: atomic.LoadInt32(&isCapturing) == 1,
		EventCount: atomic.LoadUint64(&eventCount),
		OutputFile: logPath,
	}
}

func findLatestDispatcherLog() (string, error) {
	entries, err := os.ReadDir(dispatcherLogDir)
	if err != nil {
		return "", err
	}
	var latest os.FileInfo
	var latestPath string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToUpper(e.Name()), ".LOG") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if latest == nil || info.ModTime().After(latest.ModTime()) {
			latest = info
			latestPath = filepath.Join(dispatcherLogDir, e.Name())
		}
	}
	if latestPath == "" {
		return "", fmt.Errorf("no .LOG files found")
	}
	return latestPath, nil
}

func captureTailLoop(initialPath string) {
	currentPath := initialPath
	pos := int64(0)

	if f, err := os.Open(currentPath); err == nil {
		info, _ := f.Stat()
		pos = info.Size()
		f.Close()
	}

	for atomic.LoadInt32(&isCapturing) == 1 {
		newPath, err := findLatestDispatcherLog()
		if err == nil && newPath != currentPath {
			writeLine("--- Log rotated: " + filepath.Base(newPath) + " ---")
			currentPath = newPath
			pos = 0
		}

		f, err := os.Open(currentPath)
		if err != nil {
			time.Sleep(2 * time.Second)
			continue
		}
		info, _ := f.Stat()
		size := info.Size()

		if pos > size {
			pos = 0
		}
		if size > pos {
			f.Seek(pos, 0)
			scanner := bufio.NewScanner(f)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line != "" {
					parseAndWriteLine(line)
				}
			}
			pos = size
		}
		f.Close()
		time.Sleep(500 * time.Millisecond)
	}
}

func parseAndWriteLine(line string) {
	count := atomic.AddUint64(&eventCount, 1)
	var desc string

	switch {
	case strings.Contains(line, "OEMV_UPDATE") || strings.Contains(line, "SendOEMVUpdate"):
		desc = parseOEMVLine(line)
	case strings.Contains(line, "MLScenario") || strings.Contains(line, "IPFProcess") || strings.Contains(line, "WorkLoadLevel"):
		desc = parseMLScenarioLine(line)
	case strings.Contains(line, "ForeGroundChange") || strings.Contains(line, "ProName:"):
		desc = parseFGAppLine(line)
	case strings.Contains(line, "APP Turbo") || strings.Contains(line, "APPTurbo") || strings.Contains(line, "TurboTime"):
		desc = "TURBO: " + extractAfterFirst(line, "]")
	case strings.Contains(line, "ITS_AutomaticModeSetting") || strings.Contains(line, "SetDYTCMode") || strings.Contains(line, "GearEnergy"):
		desc = "DYTC_MODE: " + extractAfterFirst(line, "]")
	case strings.Contains(line, "DTT") || strings.Contains(line, "DYTCCmd"):
		desc = "DTT: " + extractAfterFirst(line, "::")
	default:
		desc = "LOG: " + extractAfterFirst(line, "]")
	}

	writeLine(fmt.Sprintf("#%d %s", count, desc))
}

func parseOEMVLine(line string) string {
	idx := extractBetween(line, "index", ",")
	val := extractBetween(line, "value", ",")
	rc := extractAfterFirst(line, "rc")
	if idx != "" && val != "" {
		return fmt.Sprintf("OEMV_UPDATE idx=%s val=%s rc=%s", idx, val, rc)
	}
	return "OEMV_UPDATE: " + extractAfterFirst(line, "]")
}

func parseMLScenarioLine(line string) string {
	msg := extractBetween(line, "MLScenarioWorkerThread::", "")
	if msg == "" {
		msg = extractBetween(line, "IPFProcess_Func,", "")
	}
	if msg == "" {
		msg = extractAfterFirst(line, "::")
	}
	return "ML_SCENARIO: " + msg
}

func parseFGAppLine(line string) string {
	pid := extractBetween(line, "PID", ",")
	name := extractBetween(line, "ProName:", "")
	if name == "" {
		name = extractBetween(line, "ProName:", " ")
	}
	if name != "" {
		return fmt.Sprintf("FG_APP PID=%s app=%s", pid, name)
	}
	return "FG_APP: " + extractAfterFirst(line, ":")
}

func extractBetween(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	i += len(start)
	if end == "" {
		return strings.TrimSpace(s[i:])
	}
	j := strings.Index(s[i:], end)
	if j < 0 {
		return strings.TrimSpace(s[i:])
	}
	return strings.TrimSpace(s[i : i+j])
}

func extractAfterFirst(s, sep string) string {
	if i := strings.Index(s, sep); i >= 0 {
		return strings.TrimSpace(s[i+len(sep):])
	}
	return s
}

func writeLine(line string) {
	captureMu.Lock()
	defer captureMu.Unlock()
	if logFile != nil && atomic.LoadInt32(&isCapturing) == 1 {
		logFile.WriteString(line + "\n")
		if atomic.LoadUint64(&eventCount)%20 == 0 {
			logFile.Sync()
		}
	}
}
