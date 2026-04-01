//go:build windows

package backend

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// modeToValue maps DYTC mode name to ITS_AutomaticModeSetting value
var modeToValue = map[string]uint32{
	"BSM": 1, "IBSM": 2, "AQM": 3, "STD": 4,
	"APM": 5, "IEPM": 6, "EPM": 7, "DCC": 13,
}

var valueToMode = map[uint32]string{
	1: "BSM", 2: "IBSM", 3: "AQM", 4: "STD",
	5: "APM", 6: "IEPM", 7: "EPM", 13: "DCC",
}

// GetPinnedDYTCMode returns the currently pinned mode name, or "" if not pinned.
// Pinned = Policy_Override == 3
func GetPinnedDYTCMode() string {
	override := readModeCheckReg("Policy_Override", 0)
	if override != 3 {
		return ""
	}
	mode := readModeCheckReg("ITS_AutomaticModeSetting", 0)
	if name, ok := valueToMode[mode]; ok {
		return name
	}
	return fmt.Sprintf("Mode %d", mode)
}

// PinDYTCMode pins the given mode by writing ITS_AutomaticModeSetting + Policy_Override=3.
// The Dispatcher service reads these on startup and restores the mode.
func PinDYTCMode(modeId string) error {
	val, ok := modeToValue[modeId]
	if !ok {
		return fmt.Errorf("unknown mode: %s", modeId)
	}
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("cannot open registry for write: %w", err)
	}
	defer k.Close()

	if err := k.SetDWordValue("ITS_AutomaticModeSetting", val); err != nil {
		return fmt.Errorf("failed to write ITS_AutomaticModeSetting: %w", err)
	}
	// Policy_Override = 3 -> fixed mode, Dispatcher won't auto-switch on restart
	if err := k.SetDWordValue("Policy_Override", 3); err != nil {
		return fmt.Errorf("failed to write Policy_Override: %w", err)
	}
	return nil
}

// UnpinDYTCMode removes the pin by restoring Policy_Override to 0 (auto).
func UnpinDYTCMode() error {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("cannot open registry for write: %w", err)
	}
	defer k.Close()

	if err := k.SetDWordValue("Policy_Override", 0); err != nil {
		return fmt.Errorf("failed to write Policy_Override: %w", err)
	}
	return nil
}
