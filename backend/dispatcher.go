//go:build windows

package backend

import (
	"fmt"
	"sync"

	"golang.org/x/sys/windows/registry"
)

// DispatcherInfo holds Dispatcher driver and mode information
type DispatcherInfo struct {
	DriverVersion string `json:"driverVersion"`
	Description   string `json:"description"`
	CurrentMode   string `json:"currentMode"`
	AIEngineMode  string `json:"aiEngineMode"`
	AutoMode      int    `json:"autoMode"`
	ITSCurrentMode string `json:"itsCurrentMode"`
	ITSTargetMode  string `json:"itsTargetMode"`
}

// Cached description — service display name never changes at runtime
var (
	descriptionOnce sync.Once
	descriptionVal  string
)

// getCachedDescription returns the cached service display name
func getCachedDescription() string {
	descriptionOnce.Do(func() {
		descriptionVal = readRegString(
			registry.LOCAL_MACHINE,
			`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement`,
			"DisplayName",
			"Lenovo Process Management",
		)
	})
	return descriptionVal
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
	info.DriverVersion = getDispatcherExeVersion() // cached via sync.Once
	info.Description = getCachedDescription()       // cached via sync.Once
	values, err := ReadAllDispatcherValues()
	if err != nil {
		info.CurrentMode = "Registry unavailable"
		info.AIEngineMode = "Unknown"
		return info, nil
	}
	// ITS_CurrentSetting: actual hardware mode from DTT driver
	// ITS_AutomaticModeSetting: target mode set by Dispatcher (used as primary source for currentMode)
	itsCurrent := values["ITS_CurrentSetting"]
	itsTarget := values["ITS_AutomaticModeSetting"]

	// Populate individual ITS mode fields
	if modeName, ok := dispatcherModeMap[itsCurrent]; ok {
		info.ITSCurrentMode = fmt.Sprintf("%s (%d)", modeName, itsCurrent)
	} else if itsCurrent == 0 {
		info.ITSCurrentMode = "N/A (DTT not active)"
	} else {
		info.ITSCurrentMode = fmt.Sprintf("Unknown (%d)", itsCurrent)
	}
	if modeName, ok := dispatcherModeMap[itsTarget]; ok {
		info.ITSTargetMode = fmt.Sprintf("%s (%d)", modeName, itsTarget)
	} else if itsTarget == 0 {
		info.ITSTargetMode = "N/A"
	} else {
		info.ITSTargetMode = fmt.Sprintf("Unknown (%d)", itsTarget)
	}

	// Use ITS_AutomaticModeSetting (Dispatcher target mode) when available, fallback to ITS_CurrentSetting (DTT hardware mode)
	currentSetting := itsTarget
	if currentSetting == 0 {
		currentSetting = itsCurrent
	}
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
