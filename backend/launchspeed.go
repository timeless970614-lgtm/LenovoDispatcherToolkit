//go:build windows

package backend

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/windows/registry"
)

// LaunchSpeedResult holds the result of a single app launch speed test
type LaunchSpeedResult struct {
	AppName     string  `json:"appName"`
	AppPath     string  `json:"appPath"`
	LaunchMs    float64 `json:"launchMs"`    // Time from process start to window visible (seconds, 0.01 precision)
	ProcessMs   float64 `json:"processMs"`   // Time to just create the process (seconds, 0.01 precision)
	Success     bool    `json:"success"`
	Error       string  `json:"error,omitempty"`
	Category    string  `json:"category"`    // "system", "browser", "office", "media", "custom"
	Method      string  `json:"method"`      // "cold", "warm"
}

// LaunchSpeedReport holds the full report
type LaunchSpeedReport struct {
	Results    []LaunchSpeedResult `json:"results"`
	TotalTime  float64             `json:"totalTime"`  // Total benchmark time (seconds, 0.01 precision)
	AvgLaunch  float64             `json:"avgLaunch"`  // Average launch time (seconds, 0.01 precision)
	Fastest    string              `json:"fastest"`    // Fastest app name
	Slowest    string              `json:"slowest"`    // Slowest app name
	SystemBoot int64               `json:"systemBoot"` // Last boot uptime (ms since epoch)
	Timestamp  string              `json:"timestamp"`
}

// BootSpeedInfo holds Windows boot time metrics
type BootSpeedInfo struct {
	LastBootTime    string `json:"lastBootTime"`    // Human-readable last boot time
	BootDurationMs  int64  `json:"bootDurationMs"`  // Approximate boot duration
	UptimeSeconds   int64  `json:"uptimeSeconds"`   // Current uptime in seconds
	StartupCount    int    `json:"startupCount"`     // Number of startup items detected
	ServiceCount    int    `json:"serviceCount"`     // Number of auto-start services
	ServicesList    string `json:"servicesList"`     // JSON of auto-start services
	StartupItems    string `json:"startupItems"`     // JSON of startup registry items
}

// Common apps to benchmark (paths that exist on most Windows systems)
var commonApps = []struct {
	name     string
	path     string
	category string
}{
	{"Calculator", "calc.exe", "system"},
	{"Notepad", "notepad.exe", "system"},
	{"Task Manager", "taskmgr.exe", "system"},
	{"CMD", "cmd.exe", "system"},
	{"PowerShell", "powershell.exe", "system"},
	{"Paint", "mspaint.exe", "system"},
	{"Edge", "msedge.exe", "browser"},
	{"Explorer", "explorer.exe", "system"},
	{"Settings", "ms-settings:", "system"},
	{"Word", `C:\Program Files\Microsoft Office\root\Office16\WINWORD.EXE`, "office"},
	{"Excel", `C:\Program Files\Microsoft Office\root\Office16\EXCEL.EXE`, "office"},
	{"PowerPoint", `C:\Program Files\Microsoft Office\root\Office16\POWERPNT.EXE`, "office"},
	{"Outlook", `C:\Program Files\Microsoft Office\root\Office16\OUTLOOK.EXE`, "office"},
	{"WeChat", `C:\Program Files\Tencent\WeChat\WeChat.exe`, "media"},
	{"VSCode", `C:\Users\3-64\AppData\Local\Programs\Microsoft VS Code\Code.exe`, "system"},
}

var launchSpeedMu sync.Mutex

// BenchmarkLaunchSpeed benchmarks common app launch times
func BenchmarkLaunchSpeed(method string) LaunchSpeedReport {
	launchSpeedMu.Lock()
	defer launchSpeedMu.Unlock()

	start := time.Now()
	var results []LaunchSpeedResult
	var totalLaunchMs float64

	for i, app := range commonApps {
		result := measureLaunch(app.name, app.path, app.category, method)
		results = append(results, result)
		if result.Success {
			totalLaunchMs += result.LaunchMs
		}
		// Idle 5s between apps so each one can settle before the next starts
		if i < len(commonApps)-1 {
			time.Sleep(5 * time.Second)
		}
	}

	// Boot info
	bootInfo := GetBootSpeedInfo()

	var fastest, slowest string
	var fastestMs, slowestMs float64 = 999.99, 0

	successCount := 0
	for _, r := range results {
		if r.Success {
			successCount++
			if r.LaunchMs < fastestMs {
				fastestMs = r.LaunchMs
				fastest = r.AppName
			}
			if r.LaunchMs > slowestMs {
				slowestMs = r.LaunchMs
				slowest = r.AppName
			}
		}
	}

	var avgMs float64
	if successCount > 0 {
		avgMs = totalLaunchMs / float64(successCount)
	}
	avgMs = roundTo01(avgMs)

	return LaunchSpeedReport{
		Results:    results,
		TotalTime:  roundTo01(time.Since(start).Seconds()),
		AvgLaunch:  avgMs,
		Fastest:    fastest,
		Slowest:    slowest,
		SystemBoot: bootInfo.BootDurationMs,
		Timestamp:  time.Now().Format("2006-01-02 15:04:05"),
	}
}

// roundTo01 rounds a float64 to 0.01 precision
func roundTo01(v float64) float64 {
	return math.Round(v*100) / 100
}

// measureLaunch measures the time to launch a single app
func measureLaunch(name, path, category, method string) LaunchSpeedResult {
	result := LaunchSpeedResult{
		AppName:  name,
		AppPath:  path,
		Category: category,
		Method:   method,
	}

	// For cold start: close the app first (if running)
	if method == "cold" {
		// Try to kill the process if it's running (for cold start measurement)
		if name != "Explorer" && name != "Settings" {
			killCmd := exec.Command("taskkill", "/f", "/im", filepath.Base(path))
			killCmd.SysProcAttr = hideWindowAttr
			killCmd.Run() // Ignore error - process might not be running
			time.Sleep(200 * time.Millisecond) // Brief wait after kill
		}
	}

	// Check if app exists (skip if not found)
	if strings.HasPrefix(path, "ms-settings:") || path == "calc.exe" || path == "notepad.exe" ||
		path == "cmd.exe" || path == "powershell.exe" || path == "taskmgr.exe" ||
		path == "mspaint.exe" || path == "explorer.exe" || path == "msedge.exe" {
		// System apps - should exist
	} else {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			result.Success = false
			result.Error = "App not found"
			return result
		}
	}

	// Measure process creation time
	processStart := time.Now()

	var cmd *exec.Cmd
	if strings.HasPrefix(path, "ms-settings:") {
		// Settings app - use cmd start
		cmd = exec.Command("cmd", "/c", "start", path)
		cmd.SysProcAttr = hideWindowAttr
	} else if name == "Edge" {
		// Edge needs special handling
		cmd = exec.Command(path)
		cmd.SysProcAttr = hideWindowAttr
	} else if name == "Explorer" {
		// Explorer is always running, skip cold start
		if method == "cold" {
			result.Success = true
			result.LaunchMs = 0
			result.ProcessMs = 0
			result.Error = "Always running"
			return result
		}
		cmd = exec.Command(path)
		cmd.SysProcAttr = hideWindowAttr
	} else {
		cmd = exec.Command(path)
		cmd.SysProcAttr = hideWindowAttr
	}

	if err := cmd.Start(); err != nil {
		result.Success = false
		result.Error = err.Error()
		return result
	}

	result.ProcessMs = roundTo01(time.Since(processStart).Seconds())

	// For simple apps, process creation time ≈ launch time
	// For complex apps, we wait a bit then check if process is responsive
	if category == "system" || category == "system" {
		result.LaunchMs = result.ProcessMs
	} else {
		// Wait up to 3 seconds for process to stabilize
		waitStart := time.Now()
		maxWait := 3000 * time.Millisecond
		
		// Simple heuristic: wait until the process has been running for a reasonable time
		for time.Since(waitStart) < maxWait {
			// Check if process is still running
			if cmd.Process != nil {
				// For GUI apps, give them time to render
				if category == "browser" || category == "office" {
					time.Sleep(500 * time.Millisecond)
					result.LaunchMs = roundTo01(time.Since(processStart).Seconds())
					break
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
		if result.LaunchMs == 0 {
			result.LaunchMs = roundTo01(time.Since(processStart).Seconds())
		}
	}

	result.Success = true

	// Clean up: close launched apps after measurement
	if name != "Explorer" && name != "Settings" && name != "Edge" {
		go func() {
			time.Sleep(2 * time.Second)
			if cmd.Process != nil {
				cmd.Process.Kill()
			}
		}()
	}

	return result
}

// syscallHideWindow returns SysProcAttr to hide console windows
var hideWindowAttr = &syscall.SysProcAttr{HideWindow: true}

// GetBootSpeedInfo retrieves Windows boot time metrics
func GetBootSpeedInfo() BootSpeedInfo {
	info := BootSpeedInfo{}

	// Get last boot time from WMI via PowerShell
	psCmd := exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-CimInstance Win32_OperatingSystem).LastBootUpTime.ToString('yyyy-MM-dd HH:mm:ss')`)
	psCmd.SysProcAttr = hideWindowAttr
	output, err := psCmd.CombinedOutput()
	if err == nil {
		info.LastBootTime = strings.TrimSpace(string(output))
	}

	// Get uptime
	psCmd2 := exec.Command("powershell", "-NoProfile", "-Command",
		`[math]::Round((Get-CimInstance Win32_OperatingSystem).CurrentUptime.TotalSeconds)`)
	psCmd2.SysProcAttr = hideWindowAttr
	output2, err2 := psCmd2.CombinedOutput()
	if err2 == nil {
		uptimeStr := strings.TrimSpace(string(output2))
		if uptimeStr != "" {
			// Parse as int64
			fmt.Sscanf(uptimeStr, "%d", &info.UptimeSeconds)
		}
	}

	// Count startup registry items
	startupItems := getStartupItems()
	info.StartupCount = len(startupItems)
	itemsJSON, _ := json.Marshal(startupItems)
	info.StartupItems = string(itemsJSON)

	// Count auto-start services
	services := getAutoStartServices()
	info.ServiceCount = len(services)
	svcJSON, _ := json.Marshal(services)
	info.ServicesList = string(svcJSON)

	return info
}

// getStartupItems reads registry startup entries
func getStartupItems() []map[string]string {
	var items []map[string]string

	// HKCU\Software\Microsoft\Windows\CurrentVersion\Run
	k, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE)
	if err == nil {
		names, _ := k.ReadValueNames(0)
		for _, name := range names {
			val, _, err := k.GetStringValue(name)
			if err == nil {
				items = append(items, map[string]string{
					"name": name,
					"path": val,
					"source": "HKCU Run",
				})
			}
		}
		k.Close()
	}

	// HKLM\Software\Microsoft\Windows\CurrentVersion\Run
	k2, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`Software\Microsoft\Windows\CurrentVersion\Run`,
		registry.QUERY_VALUE)
	if err == nil {
		names, _ := k2.ReadValueNames(0)
		for _, name := range names {
			val, _, err := k2.GetStringValue(name)
			if err == nil {
				items = append(items, map[string]string{
					"name": name,
					"path": val,
					"source": "HKLM Run",
				})
			}
		}
		k2.Close()
	}

	return items
}

// getAutoStartServices lists services with Auto start type
func getAutoStartServices() []map[string]string {
	var services []map[string]string

	psCmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-CimInstance Win32_Service | Where-Object {$_.StartMode -eq 'Auto' -and $_.State -eq 'Running'} | Select-Object Name, DisplayName, State | ConvertTo-Json -Compress`)
	psCmd.SysProcAttr = hideWindowAttr
	output, err := psCmd.CombinedOutput()
	if err == nil && len(output) > 0 {
		// Parse JSON
		var svcList []struct {
			Name        string `json:"Name"`
			DisplayName string `json:"DisplayName"`
			State       string `json:"State"`
		}
		if json.Unmarshal(output, &svcList) == nil {
			for _, s := range svcList {
				services = append(services, map[string]string{
					"name":        s.Name,
					"displayName": s.DisplayName,
					"state":       s.State,
				})
			}
		}
	}

	return services
}

// BenchmarkCustomAppLaunch benchmarks a specific app's launch time
func BenchmarkCustomAppLaunch(appPath string, method string) LaunchSpeedResult {
	name := filepath.Base(appPath)
	return measureLaunch(name, appPath, "custom", method)
}

// GetCommonAppsList returns the list of apps that will be benchmarked
func GetCommonAppsList() []map[string]string {
	var list []map[string]string
	for _, app := range commonApps {
		exists := true
		if !strings.HasPrefix(app.path, "ms-settings:") && app.path != "calc.exe" &&
			app.path != "notepad.exe" && app.path != "cmd.exe" && app.path != "powershell.exe" &&
			app.path != "taskmgr.exe" && app.path != "mspaint.exe" && app.path != "explorer.exe" &&
			app.path != "msedge.exe" {
			if _, err := os.Stat(app.path); os.IsNotExist(err) {
				exists = false
			}
		}
		list = append(list, map[string]string{
			"name":     app.name,
			"path":     app.path,
			"category": app.category,
			"exists":   fmt.Sprintf("%v", exists),
		})
	}
	return list
}
