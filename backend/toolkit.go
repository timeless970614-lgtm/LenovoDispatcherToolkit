//go:build windows

package backend

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// ToolkitTool represents an installable tool
type ToolkitTool struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Description     string `json:"description"`
	Version         string `json:"version"`
	Category        string `json:"category"` // "system", "monitor", "benchmark", "diagnostic"
	WingetID        string `json:"wingetId"` // Winget package ID
	ExecutableName  string `json:"executableName"` // Main executable to detect/run
	InstallLocation string `json:"installLocation"` // Typical install folder name
	SizeMB          int    `json:"sizeMb"`
	Website         string `json:"website"`
	Vendor          string `json:"vendor"`
}

// ToolkitInstallStatus represents the installation status of a tool
type ToolkitInstallStatus struct {
	ToolID      string `json:"toolId"`
	Installed   bool   `json:"installed"`
	InstallPath string `json:"installPath"`
	Version     string `json:"version"`
	LastChecked string `json:"lastChecked"`
}

// ToolkitInstallProgress represents real-time installation progress
type ToolkitInstallProgress struct {
	ToolID    string `json:"toolId"`
	Status    string `json:"status"` // "idle", "installing", "completed", "error"
	Progress  int    `json:"progress"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

var (
	toolkitTools = []ToolkitTool{
		{
			ID:              "hwinfo64",
			Name:            "HWiNFO64",
			Description:     "Comprehensive hardware monitoring and diagnostics",
			Version:         "7.80",
			Category:        "monitor",
			WingetID:        "REALiX.HWiNFO",
			ExecutableName:  "HWiNFO64.EXE",
			InstallLocation: "HWiNFO64",
			SizeMB:          12,
			Website:         "https://www.hwinfo.com/",
			Vendor:          "REALiX",
		},
		{
			ID:              "cpuz",
			Name:            "CPU-Z",
			Description:     "CPU, motherboard, memory information utility",
			Version:         "2.09",
			Category:        "system",
			WingetID:        "CPUID.CPU-Z",
			ExecutableName:  "cpuz_x64.exe",
			InstallLocation: "CPUID\\CPU-Z",
			SizeMB:          2,
			Website:         "https://www.cpuid.com/softwares/cpu-z.html",
			Vendor:          "CPUID",
		},
		{
			ID:              "gpuz",
			Name:            "GPU-Z",
			Description:     "GPU information and monitoring utility",
			Version:         "2.60.0",
			Category:        "monitor",
			WingetID:        "TechPowerUp.GPU-Z",
			ExecutableName:  "GPU-Z.exe",
			InstallLocation: "GPU-Z",
			SizeMB:          8,
			Website:         "https://www.techpowerup.com/gpuz/",
			Vendor:          "TechPowerUp",
		},
		{
			ID:              "hwmonitor",
			Name:            "HWMonitor",
			Description:     "Hardware monitoring for temperatures, voltages, fan speeds",
			Version:         "1.53",
			Category:        "monitor",
			WingetID:        "CPUID.HWMonitor",
			ExecutableName:  "HWMonitor.exe",
			InstallLocation: "CPUID\\HWMonitor",
			SizeMB:          2,
			Website:         "https://www.cpuid.com/softwares/hwmonitor.html",
			Vendor:          "CPUID",
		},
		{
			ID:              "crystaldiskinfo",
			Name:            "CrystalDiskInfo",
			Description:     "SSD/HDD health monitoring and SMART data",
			Version:         "9.6.0",
			Category:        "diagnostic",
			WingetID:        "CrystalDewWorld.CrystalDiskInfo",
			ExecutableName:  "DiskInfo64.exe",
			InstallLocation: "CrystalDiskInfo",
			SizeMB:          5,
			Website:         "https://crystalmark.info/en/software/crystaldiskinfo/",
			Vendor:          "Crystal Dew World",
		},
		{
			ID:              "crystaldiskmark",
			Name:            "CrystalDiskMark",
			Description:     "Disk benchmark tool for measuring read/write speeds",
			Version:         "8.0.5",
			Category:        "benchmark",
			WingetID:        "CrystalDewWorld.CrystalDiskMark",
			ExecutableName:  "DiskMark64.exe",
			InstallLocation: "CrystalDiskMark",
			SizeMB:          4,
			Website:         "https://crystalmark.info/en/software/crystaldiskmark/",
			Vendor:          "Crystal Dew World",
		},
		{
			ID:              "msiafterburner",
			Name:            "MSI Afterburner",
			Description:     "GPU overclocking, monitoring, and video capture",
			Version:         "4.6.5",
			Category:        "monitor",
			WingetID:        "MSI.Afterburner",
			ExecutableName:  "MSIAfterburner.exe",
			InstallLocation: "MSI Afterburner",
			SizeMB:          50,
			Website:         "https://www.msi.com/Landing/afterburner/",
			Vendor:          "MSI",
		},
		{
			ID:              "prime95",
			Name:            "Prime95",
			Description:     "CPU stress testing and stability benchmark",
			Version:         "30.19",
			Category:        "benchmark",
			WingetID:        "GIMPS.Prime95",
			ExecutableName:  "prime95.exe",
			InstallLocation: "Prime95",
			SizeMB:          12,
			Website:         "https://www.mersenne.org/download/",
			Vendor:          "GIMPS",
		},
		{
			ID:              "furmark",
			Name:            "FurMark",
			Description:     "GPU stress test and benchmark",
			Version:         "1.40.0",
			Category:        "benchmark",
			WingetID:        "Geeks3D.FurMark",
			ExecutableName:  "FurMark.exe",
			InstallLocation: "Geeks3D\\FurMark",
			SizeMB:          60,
			Website:         "https://www.geeks3d.com/furmark/",
			Vendor:          "Geeks3D",
		},
		{
			ID:              "throttlestop",
			Name:            "ThrottleStop",
			Description:     "CPU monitoring and power management tool",
			Version:         "9.7.1",
			Category:        "monitor",
			WingetID:        "TechPowerUp.ThrottleStop",
			ExecutableName:  "ThrottleStop.exe",
			InstallLocation: "ThrottleStop",
			SizeMB:          3,
			Website:         "https://www.techpowerup.com/download/throttlestop/",
			Vendor:          "TechPowerUp",
		},
		{
			ID:              "cpuid",
			Name:            "CPUID",
			Description:     "System information utility with hardware detection",
			Version:         "4.60",
			Category:        "system",
			WingetID:        "CPUID.CPUID",
			ExecutableName:  "cpuid.exe",
			InstallLocation: "CPUID",
			SizeMB:          3,
			Website:         "https://www.cpuid.com/",
			Vendor:          "CPUID",
		},
		{
			ID:              "speccy",
			Name:            "Speccy",
			Description:     "Fast, lightweight system information tool",
			Version:         "1.33",
			Category:        "system",
			WingetID:        "Piriform.Speccy",
			ExecutableName:  "Speccy64.exe",
			InstallLocation: "Speccy",
			SizeMB:          10,
			Website:         "https://www.ccleaner.com/speccy",
			Vendor:          "Piriform",
		},
		{
			ID:              "pcmark10",
			Name:            "PCMark 10",
			Description:     "Industry-standard PC benchmark for Windows 10/11",
			Version:         "2.2",
			Category:        "benchmark",
			WingetID:        "UL.PCMark10",
			ExecutableName:  "PCMark10.exe",
			InstallLocation: "UL\\PCMark 10",
			SizeMB:          2500,
			Website:         "https://benchmarks.ul.com/pcmark10",
			Vendor:          "UL Solutions",
		},
	}

	toolkitProgressMutex sync.Mutex
	toolkitProgressMap   = make(map[string]ToolkitInstallProgress)
)

// GetToolkitTools returns the list of available tools
func GetToolkitTools() []ToolkitTool {
	return toolkitTools
}

// GetToolkitInstallDir returns the tools directory (for compatibility)
func GetToolkitInstallDir() string {
	return `C:\Program Files`
}

// findToolExecutable searches for the tool executable in Program Files
func findToolExecutable(tool *ToolkitTool) string {
	searchPaths := []string{
		filepath.Join(os.Getenv("ProgramFiles"), tool.InstallLocation, tool.ExecutableName),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), tool.InstallLocation, tool.ExecutableName),
		filepath.Join(os.Getenv("ProgramFiles"), tool.InstallLocation, strings.Replace(tool.ExecutableName, "_x64", "", -1)),
		filepath.Join(os.Getenv("ProgramFiles(x86)"), tool.InstallLocation, strings.Replace(tool.ExecutableName, "_x64", "", -1)),
	}

	// Also try glob patterns for versioned folders
	for _, base := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
		pattern := filepath.Join(base, tool.InstallLocation+"*", tool.ExecutableName)
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			return matches[0]
		}
		// Also try without version suffix in executable name
		pattern2 := filepath.Join(base, tool.InstallLocation+"*", strings.Replace(tool.ExecutableName, "_x64", "", -1))
		matches2, _ := filepath.Glob(pattern2)
		if len(matches2) > 0 {
			return matches2[0]
		}
	}

	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// CheckToolkitInstalled checks if a tool is installed
func CheckToolkitInstalled(toolID string) ToolkitInstallStatus {
	status := ToolkitInstallStatus{
		ToolID:      toolID,
		Installed:   false,
		LastChecked: time.Now().Format("2006-01-02 15:04:05"),
	}

	var tool *ToolkitTool
	for i := range toolkitTools {
		if toolkitTools[i].ID == toolID {
			tool = &toolkitTools[i]
			break
		}
	}
	if tool == nil {
		return status
	}

	exePath := findToolExecutable(tool)
	if exePath != "" {
		status.Installed = true
		status.InstallPath = exePath
	}

	return status
}

// CheckAllToolkitInstalled returns installation status for all tools
func CheckAllToolkitInstalled() []ToolkitInstallStatus {
	var statuses []ToolkitInstallStatus
	for _, tool := range toolkitTools {
		statuses = append(statuses, CheckToolkitInstalled(tool.ID))
	}
	return statuses
}

// emitProgress emits installation progress
func emitProgress(toolID, status string, progress int, message string) {
	p := ToolkitInstallProgress{
		ToolID:    toolID,
		Status:    status,
		Progress:  progress,
		Message:   message,
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
	}
	toolkitProgressMutex.Lock()
	toolkitProgressMap[toolID] = p
	toolkitProgressMutex.Unlock()
}

// GetToolkitProgress returns current progress for a tool
func GetToolkitProgress(toolID string) ToolkitInstallProgress {
	toolkitProgressMutex.Lock()
	defer toolkitProgressMutex.Unlock()
	if p, ok := toolkitProgressMap[toolID]; ok {
		return p
	}
	return ToolkitInstallProgress{ToolID: toolID, Status: "idle"}
}

// InstallToolkitTool installs a tool using winget
func InstallToolkitTool(toolID string) ToolkitInstallProgress {
	var tool *ToolkitTool
	for i := range toolkitTools {
		if toolkitTools[i].ID == toolID {
			tool = &toolkitTools[i]
			break
		}
	}
	if tool == nil {
		return ToolkitInstallProgress{
			ToolID:  toolID,
			Status:  "error",
			Message: "Tool not found: " + toolID,
		}
	}

	emitProgress(toolID, "installing", 0, "Installing "+tool.Name+" via winget...")

	// Check if winget is available
	wingetPath, err := exec.LookPath("winget")
	if err != nil {
		// No winget, open download page
		emitProgress(toolID, "error", 0, "winget not found. Opening download page...")
		exec.Command("rundll32", "url.dll,FileProtocolHandler", tool.Website).Start()
		return GetToolkitProgress(toolID)
	}

	// Run winget install asynchronously
	go func() {
		cmd := exec.Command(wingetPath, "install", "--id", tool.WingetID, "--accept-source-agreements", "--accept-package-agreements", "--silent")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		
		_, err := cmd.CombinedOutput()
		if err != nil {
			// Winget failed, open download page
			emitProgress(toolID, "error", 0, "Not available via winget. Opening download page...")
			exec.Command("rundll32", "url.dll,FileProtocolHandler", tool.Website).Start()
			return
		}

		emitProgress(toolID, "completed", 100, tool.Name+" installed successfully!")

		// Auto-launch after successful install
		time.Sleep(2 * time.Second)
		if err := RunToolkitTool(toolID); err != nil {
			// Just log, don't fail the install
			fmt.Printf("Auto-launch failed: %v\n", err)
		}
	}()

	return GetToolkitProgress(toolID)
}

// UninstallToolkitTool uninstalls a tool using winget
func UninstallToolkitTool(toolID string) error {
	var tool *ToolkitTool
	for i := range toolkitTools {
		if toolkitTools[i].ID == toolID {
			tool = &toolkitTools[i]
			break
		}
	}
	if tool == nil {
		return fmt.Errorf("tool not found: %s", toolID)
	}

	cmd := exec.Command("winget", "uninstall", "--id", tool.WingetID, "-h")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	return cmd.Run()
}

// RunToolkitTool launches an installed tool
func RunToolkitTool(toolID string) error {
	status := CheckToolkitInstalled(toolID)
	if !status.Installed {
		return fmt.Errorf("tool not installed: %s", toolID)
	}

	cmd := visibleCmd(status.InstallPath)
	return cmd.Start()
}

// OpenToolkitFolder opens Program Files in Explorer
func OpenToolkitFolder() error {
	cmd := exec.Command("explorer.exe", os.Getenv("ProgramFiles"))
	return cmd.Start()
}

// IsToolkitBusy checks if a tool is currently being installed
func IsToolkitBusy(toolID string) bool {
	toolkitProgressMutex.Lock()
	defer toolkitProgressMutex.Unlock()
	if p, ok := toolkitProgressMap[toolID]; ok {
		return p.Status == "installing"
	}
	return false
}

// GetInstalledToolkitTools returns list of installed tool IDs
func GetInstalledToolkitTools() []string {
	var installed []string
	for _, tool := range toolkitTools {
		if status := CheckToolkitInstalled(tool.ID); status.Installed {
			installed = append(installed, tool.ID)
		}
	}
	return installed
}

// CheckWingetAvailable checks if winget is available on the system
func CheckWingetAvailable() bool {
	_, err := exec.LookPath("winget")
	return err == nil
}

// ToolkitJSON returns tools as JSON string (for debugging)
func ToolkitJSON() string {
	data, _ := json.MarshalIndent(toolkitTools, "", "  ")
	return string(data)
}
