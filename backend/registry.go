//go:build windows

package backend

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

const registryPath = `SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`

// ReadDWORD reads a DWORD value from the Dispatcher registry path
func ReadDWORD(valueName string) (uint32, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return 0, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()
	val, _, err := k.GetIntegerValue(valueName)
	if err != nil {
		return 0, fmt.Errorf("failed to read value %s: %w", valueName, err)
	}
	return uint32(val), nil
}

// WriteDWORD writes a DWORD value to the Dispatcher registry path
func WriteDWORD(valueName string, value uint32) error {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open registry key for writing: %w", err)
	}
	defer k.Close()
	if err := k.SetDWordValue(valueName, value); err != nil {
		return fmt.Errorf("failed to write value %s: %w", valueName, err)
	}
	return nil
}

// ReadAllDispatcherValues reads all relevant Dispatcher registry values
func ReadAllDispatcherValues() (map[string]uint32, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, registryPath, registry.QUERY_VALUE)
	if err != nil {
		return nil, fmt.Errorf("failed to open registry key: %w", err)
	}
	defer k.Close()

	keys := []string{
		"ITS_CurrentSetting",
		"ITS_AutomaticModeSetting",
		"Policy_DynamicPLx",
		"Policy_DynamicPLxLog",
		"Policy_AIEngine",
	}
	result := make(map[string]uint32)
	for _, key := range keys {
		val, _, err := k.GetIntegerValue(key)
		if err != nil {
			result[key] = 0
		} else {
			result[key] = uint32(val)
		}
	}
	return result, nil
}
