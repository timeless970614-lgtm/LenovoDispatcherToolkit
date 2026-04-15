//go:build windows

package backend

import (
	"encoding/json"
	"os/exec"
	"strings"
	"syscall"
)

// PPMPlatformInfo represents platform information
type PPMPlatformInfo struct {
	CPUName      string `json:"cpuName"`
	Cores        int    `json:"cores"`
	Threads      int    `json:"threads"`
	Platform     string `json:"platform"`
	Architecture string `json:"architecture"`
}

// PPMDriverInfo represents a PPM driver
type PPMDriverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Date    string `json:"date"`
}

// pnpmDriverRaw is used for JSON unmarshaling from PowerShell output
type pnpmDriverRaw struct {
	DeviceName    string `json:"DeviceName"`
	DriverVersion string `json:"DriverVersion"`
	DriverDate    string `json:"DriverDate"`
}

// runPowerShellHidden executes PowerShell command without showing window
func runPowerShellHidden(command string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", command)
	
	// Hide window on Windows
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
	
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

// GetPPMPlatformInfo retrieves platform information
func GetPPMPlatformInfo() *PPMPlatformInfo {
	info := &PPMPlatformInfo{
		Platform:     "Intel",
		Architecture: "x64",
	}

	// Get CPU info via PowerShell (hidden window)
	output, err := runPowerShellHidden(
		"Get-WmiObject -Class Win32_Processor | Select-Object -First 1 | ConvertTo-Json")
	if err != nil {
		return info
	}

	// Parse JSON
	var cpu struct {
		Name                      string `json:"Name"`
		NumberOfCores             int    `json:"NumberOfCores"`
		NumberOfLogicalProcessors int    `json:"NumberOfLogicalProcessors"`
	}
	
	if err := json.Unmarshal([]byte(output), &cpu); err == nil {
		info.CPUName = cpu.Name
		info.Cores = cpu.NumberOfCores
		info.Threads = cpu.NumberOfLogicalProcessors
	}

	return info
}

// GetPPMDrivers retrieves PPM related drivers
func GetPPMDrivers() []PPMDriverInfo {
	var drivers []PPMDriverInfo

	// PowerShell command to get PPM drivers (hidden window)
	output, err := runPowerShellHidden(
		`Get-WmiObject -Class Win32_PnPSignedDriver | Where-Object { $_.DeviceName -like "*PPM*" -or $_.DeviceName -like "*Dynamic Tuning*" -or $_.DeviceName -like "*Innovation Platform*" -or $_.DeviceName -like "*Processor Participant*" } | Select-Object DeviceName, DriverVersion, DriverDate | ConvertTo-Json`)
	if err != nil {
		return drivers
	}

	// Parse JSON output
	var rawDrivers []pnpmDriverRaw
	if err := json.Unmarshal([]byte(output), &rawDrivers); err != nil {
		// Try single object
		var singleDriver pnpmDriverRaw
		if err := json.Unmarshal([]byte(output), &singleDriver); err == nil {
			rawDrivers = []pnpmDriverRaw{singleDriver}
		}
	}

	// Convert to PPMDriverInfo
	for _, raw := range rawDrivers {
		driver := PPMDriverInfo{
			Name:    raw.DeviceName,
			Version: raw.DriverVersion,
			Date:    raw.DriverDate,
		}
		drivers = append(drivers, driver)
	}

	return drivers
}

// FormatDate converts PowerShell date format to YYYY-MM-DD
func FormatDate(dateStr string) string {
	// PowerShell date format: "20241216000000.******+***"
	if len(dateStr) >= 8 {
		return dateStr[0:4] + "-" + dateStr[4:6] + "-" + dateStr[6:8]
	}
	return dateStr
}

// GetDriverDisplayName returns a short display name for a driver
func GetDriverDisplayName(name string) string {
	name = strings.TrimSpace(name)
	switch {
	case strings.Contains(name, "Framework Manager"):
		return "IPF Framework Manager"
	case strings.Contains(name, "Processor Participant"):
		return "IPF Processor Participant"
	case strings.Contains(name, "Generic Participant"):
		return "IPF Generic Participant"
	case strings.Contains(name, "Dynamic Tuning") && !strings.Contains(name, "Updater"):
		return "Intel DTT"
	case strings.Contains(name, "PPM Provisioning"):
		return "PPM Provisioning"
	default:
		return name
	}
}
