//go:build windows

package backend

import (
	"golang.org/x/sys/windows/registry"
)

// EPOTStatus represents EPOT parameters for ML_Scenario
type EPOTStatus struct {
	EPOT1 uint32 `json:"epot1"`
	EPOT2 uint32 `json:"epot2"`
	EPOT3 uint32 `json:"epot3"`
	EPOT4 uint32 `json:"epot4"`
	Valid bool   `json:"valid"`
}

// GetEPOTStatus reads EPOT parameters from registry
func GetEPOTStatus() EPOTStatus {
	status := EPOTStatus{}

	// GPU class GUID path
	classPath := `SYSTEM\CurrentControlSet\Control\Class\{4d36e968-e325-11ce-bfc1-08002be10318}`

	// Open GPU class key
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, classPath, registry.READ)
	if err != nil {
		return status
	}
	defer k.Close()

	// Enumerate subkeys (0000, 0001, etc.)
	subkeys, err := k.ReadSubKeyNames(-1)
	if err != nil {
		return status
	}

	// Try each subkey to find EPOT values
	for _, subkey := range subkeys {
		gpuKeyPath := classPath + `\` + subkey
		gpuKey, err := registry.OpenKey(registry.LOCAL_MACHINE, gpuKeyPath, registry.READ)
		if err != nil {
			continue
		}

		epot1, _, err1 := gpuKey.GetIntegerValue("EPOT1")
		epot2, _, err2 := gpuKey.GetIntegerValue("EPOT2")
		epot3, _, err3 := gpuKey.GetIntegerValue("EPOT3")
		epot4, _, err4 := gpuKey.GetIntegerValue("EPOT4")
		gpuKey.Close()

		if err1 == nil && err2 == nil && err3 == nil && err4 == nil {
			status.EPOT1 = uint32(epot1)
			status.EPOT2 = uint32(epot2)
			status.EPOT3 = uint32(epot3)
			status.EPOT4 = uint32(epot4)
			status.Valid = true
			break
		}
	}

	return status
}
