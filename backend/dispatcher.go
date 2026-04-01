//go:build windows

package backend

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// DispatcherInfo holds Dispatcher driver and mode information
type DispatcherInfo struct {
	DriverVersion string `json:"driverVersion"`
	Description   string `json:"description"`
	CurrentMode   string `json:"currentMode"`
	AIEngineMode  string `json:"aiEngineMode"`
	AutoMode      int    `json:"autoMode"`
}

// dispatcherModeMap maps ITS_AutomaticModeSetting values to mode names
var dispatcherModeMap = map[uint32]string{
	11: "Geek Mode",
	10: "Yoga Flat",
	9:  "Yoga Tent",
	8:  "Yoga Tablet",
	7:  "Extreme Performance",
	6:  "Intelligent Extreme",
	5:  "Intelligent Auto Performance",
	4:  "Intelligent Stand Mode",
	3:  "Intelligent Auto Quiet",
	2:  "Intelligent Battery Saving",
	1:  "Battery Saving",
}

// GetDispatcherInfo retrieves Dispatcher driver info and current mode (pure registry, no WMI)
func GetDispatcherInfo() (DispatcherInfo, error) {
	info := DispatcherInfo{}
	info.DriverVersion = getDispatcherExeVersion()
	info.Description = readRegString(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement`,
		"DisplayName",
		"Lenovo Process Management",
	)
	values, err := ReadAllDispatcherValues()
	if err != nil {
		info.CurrentMode = "Registry unavailable"
		info.AIEngineMode = "Unknown"
		return info, nil
	}
	currentSetting := values["ITS_AutomaticModeSetting"]
	if modeName, ok := dispatcherModeMap[currentSetting]; ok {
		info.CurrentMode = fmt.Sprintf("%s (%d)", modeName, currentSetting)
	} else {
		info.CurrentMode = fmt.Sprintf("Unknown (%d)", currentSetting)
	}
	aiEngine := values["Policy_AIEngine"]
	if aiEngine == 0xc {
		info.AIEngineMode = "AI Engine"
	} else {
		info.AIEngineMode = "CPU Engine"
	}
	info.AutoMode = int(currentSetting)
	return info, nil
}
