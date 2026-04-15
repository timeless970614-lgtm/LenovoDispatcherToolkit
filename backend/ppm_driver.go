//go:build windows

package backend

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
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
	Name         string `json:"name"`
	Version      string `json:"version"`
	Date         string `json:"date"`
	Manufacturer string `json:"manufacturer"`
	Location     string `json:"location"`
}

// GetPPMPlatformInfo retrieves platform information
func GetPPMPlatformInfo() *PPMPlatformInfo {
	info := &PPMPlatformInfo{
		Platform:     "Intel",
		Architecture: "x64",
	}

	// Get CPU info via PowerShell
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		"Get-WmiObject -Class Win32_Processor | Select-Object -First 1 | ConvertTo-Json")
	output, err := cmd.Output()
	if err == nil {
		// Parse JSON-like output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, `"Name"`) {
				re := regexp.MustCompile(`"Name"\s*:\s*"([^"]+)"`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					info.CPUName = matches[1]
				}
			}
			if strings.HasPrefix(line, `"NumberOfCores"`) {
				re := regexp.MustCompile(`"NumberOfCores"\s*:\s*(\d+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					fmt.Sscanf(matches[1], "%d", &info.Cores)
				}
			}
			if strings.HasPrefix(line, `"NumberOfLogicalProcessors"`) {
				re := regexp.MustCompile(`"NumberOfLogicalProcessors"\s*:\s*(\d+)`)
				matches := re.FindStringSubmatch(line)
				if len(matches) > 1 {
					fmt.Sscanf(matches[1], "%d", &info.Threads)
				}
			}
		}
	}

	return info
}

// GetPPMDrivers retrieves PPM related drivers
func GetPPMDrivers() []PPMDriverInfo {
	var drivers []PPMDriverInfo

	// PowerShell command to get PPM drivers
	cmd := exec.Command("powershell", "-NoProfile", "-Command",
		`Get-WmiObject -Class Win32_PnPSignedDriver | Where-Object { 
			$_.DeviceName -like "*PPM*" -or 
			$_.DeviceName -like "*Dynamic Tuning*" -or 
			$_.DeviceName -like "*Innovation Platform*" -or
			$_.DeviceName -like "*Processor Participant*"
		} | Select-Object DeviceName, DriverVersion, DriverDate, Manufacturer, Location | ConvertTo-Json`)

	output, err := cmd.Output()
	if err != nil {
		return drivers
	}

	// Parse output
	text := string(output)
	
	// Check if it's an array or single object
	if strings.Contains(text, `"DeviceName"`) {
		// Single object or array
		entries := parseDriverJSON(text)
		drivers = append(drivers, entries...)
	}

	return drivers
}

func parseDriverJSON(text string) []PPMDriverInfo {
	var drivers []PPMDriverInfo

	// Simple JSON parsing for the format returned by ConvertTo-Json
	// Split by array elements if array
	var entries []string
	if strings.HasPrefix(strings.TrimSpace(text), "[") {
		// Array - split by }, {
		re := regexp.MustCompile(`\{\s*"DeviceName"`)
		entries = re.Split(text, -1)[1:] // Skip first empty
	} else {
		entries = []string{text}
	}

	for _, entry := range entries {
		driver := PPMDriverInfo{}
		
		// Extract DeviceName
		re := regexp.MustCompile(`"DeviceName"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(entry); len(matches) > 1 {
			driver.Name = matches[1]
		}
		
		// Extract DriverVersion
		re = regexp.MustCompile(`"DriverVersion"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(entry); len(matches) > 1 {
			driver.Version = matches[1]
		}
		
		// Extract DriverDate
		re = regexp.MustCompile(`"DriverDate"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(entry); len(matches) > 1 {
			driver.Date = matches[1]
		}
		
		// Extract Manufacturer
		re = regexp.MustCompile(`"Manufacturer"\s*:\s*"([^"]+)"`)
		if matches := re.FindStringSubmatch(entry); len(matches) > 1 {
			driver.Manufacturer = matches[1]
		}
		
		// Extract Location
		re = regexp.MustCompile(`"Location"\s*:\s*"([^"]*)"`)
		if matches := re.FindStringSubmatch(entry); len(matches) > 1 {
			driver.Location = matches[1]
		}

		if driver.Name != "" {
			drivers = append(drivers, driver)
		}
	}

	return drivers
}
