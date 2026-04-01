//go:build windows



package backend

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const logDir = `C:\ProgramData\Lenovo\LenovoDispatcher\Logs`

// LogFileInfo holds metadata about a log file
type LogFileInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

// GetLogFiles returns the list of log files sorted by modification time (newest first)
func GetLogFiles() []LogFileInfo {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return nil
	}
	var files []LogFileInfo
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ext := strings.ToUpper(filepath.Ext(name))
		if ext != ".LOG" && ext != ".CSV" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		files = append(files, LogFileInfo{
			Name:    name,
			Size:    info.Size(),
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})
	}
	sort.Slice(files, func(i, j int) bool {
		return files[i].ModTime > files[j].ModTime
	})
	return files
}

// ReadLogTail reads the last N lines from the latest .LOG file
func ReadLogTail(maxLines int) string {
	entries, err := os.ReadDir(logDir)
	if err != nil {
		return fmt.Sprintf("Cannot open log directory: %v", err)
	}
	type fileEntry struct {
		path    string
		modTime time.Time
	}
	var logs []fileEntry
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.ToUpper(filepath.Ext(e.Name())) != ".LOG" {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		logs = append(logs, fileEntry{
			path:    filepath.Join(logDir, e.Name()),
			modTime: info.ModTime(),
		})
	}
	if len(logs) == 0 {
		return "No log files found"
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].modTime.After(logs[j].modTime)
	})
	data, err := os.ReadFile(logs[0].path)
	if err != nil {
		return fmt.Sprintf("Cannot read log: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return strings.Join(lines, "\n")
}

// GetLogSummary returns a structured summary of the latest log for AI analysis
func GetLogSummary() map[string]interface{} {
	content := ReadLogTail(500)
	lines := strings.Split(content, "\n")
	summary := map[string]interface{}{
		"totalLines": len(lines),
		"logContent": content,
	}
	var modes, errors, turboEvents, appEvents, workloadLevels []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "ITS_AutomaticModeSetting") || strings.Contains(line, "SetDYTCMode") || strings.Contains(line, "GearEnergy") {
			modes = append(modes, line)
		}
		if strings.Contains(line, "Failed") || strings.Contains(line, "Error") || strings.Contains(line, "error") {
			errors = append(errors, line)
		}
		if strings.Contains(line, "APP_Turbo") || strings.Contains(line, "APPTurbo") || strings.Contains(line, "TurboTime") {
			turboEvents = append(turboEvents, line)
		}
		if strings.Contains(line, "ForeGroundChange") || strings.Contains(line, "APPName") {
			appEvents = append(appEvents, line)
		}
		if strings.Contains(line, "WorkLoadLevel") {
			workloadLevels = append(workloadLevels, line)
		}
	}
	summary["modeChanges"] = lastN(modes, 10)
	summary["errors"] = lastN(errors, 10)
	summary["turboEvents"] = lastN(turboEvents, 10)
	summary["appEvents"] = lastN(appEvents, 10)
	summary["workloadLevels"] = lastN(workloadLevels, 10)
	return summary
}

func lastN(s []string, n int) []string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
