//go:build windows

package backend

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

// ModeCheckInfo holds all mode check information
type ModeCheckInfo struct {
	Model              string              `json:"model"`
	BIOSVersion        string              `json:"biosVersion"`
	MemoryType        string              `json:"memoryType"`
	DriverVersion      string              `json:"driverVersion"`
	DispatcherMode     string              `json:"dispatcherMode"`
	DispatcherVersion  string              `json:"dispatcherVersion"`
	AIEngineMode       string              `json:"aiEngineMode"`
	MainVer            string              `json:"mainVer"`
	DYTCValue          uint32              `json:"dytcValue"`
	DYTCBinary         string              `json:"dytcBinary"`
	EnableFuncValue    uint32              `json:"enableFuncValue"`
	EnableFuncHex      string              `json:"enableFuncHex"`
	EnableFuncPolicies []EnableFuncPolicy  `json:"enableFuncPolicies"`
	Features           []ModeCheckFeature   `json:"features"`
}

// EnableFuncPolicy represents a single policy bit
type EnableFuncPolicy struct {
	Bit     int    `json:"bit"`
	Name    string `json:"name"`
	Desc    string `json:"desc"`
	Enabled bool   `json:"enabled"`
}

// ModeCheckFeature represents a supported feature
type ModeCheckFeature struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Support string `json:"support"`
}

// GetModeCheckInfo returns all mode check information (pure registry, no DLL calls)
func GetModeCheckInfo() ModeCheckInfo {
	info := ModeCheckInfo{}

	info.BIOSVersion = readRegString(
		registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\BIOS`,
		"BIOSVersion",
		"N/A",
	)
	info.Model = readRegString(
		registry.LOCAL_MACHINE,
		`HARDWARE\DESCRIPTION\System\BIOS`,
		"SystemProductName",
		"N/A",
	)
	info.MemoryType = "LPDDR5 / DDR5"

	info.DriverVersion = getDispatcherExeVersion()
	info.DispatcherVersion = info.DriverVersion
	info.MainVer = calcMainVersion(info.DriverVersion)

	sliderKey, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`,
		registry.QUERY_VALUE,
	)
	if err == nil {
		defer sliderKey.Close()

		if v, _, e := sliderKey.GetIntegerValue("ITS_AutomaticModeSetting"); e == nil {
			modeNames := map[uint32]string{
				1: "Battery Saving", 2: "Intelligent Battery Saving",
				3: "Intelligent Auto Quiet", 4: "Intelligent Stand Mode",
				5: "Intelligent Auto Performance", 6: "Intelligent Extreme",
				7: "Extreme Performance", 8: "Yoga Tablet",
				9: "Yoga Tent", 10: "Yoga Flat", 11: "Geek Mode",
			}
			mode := uint32(v)
			if name, ok := modeNames[mode]; ok {
				info.DispatcherMode = fmt.Sprintf("%s (%d)", name, mode)
			} else {
				info.DispatcherMode = fmt.Sprintf("Unknown (%d)", mode)
			}
		} else {
			info.DispatcherMode = "N/A"
		}

		if v, _, e := sliderKey.GetIntegerValue("Policy_AIEngine"); e == nil && uint32(v) == 0xc {
			info.AIEngineMode = "AI Engine"
		} else {
			info.AIEngineMode = "CPU Engine"
		}

		if v, _, e := sliderKey.GetIntegerValue("Policy_DispatcherFunc"); e == nil {
			info.DYTCValue = uint32(v)
		} else {
			info.DYTCValue = 0
		}

		if v, _, e := sliderKey.GetIntegerValue("Policy_EnableFunc"); e == nil {
			info.EnableFuncValue = uint32(v)
		}
	} else {
		info.DispatcherMode = "Registry unavailable"
		info.AIEngineMode = "Unknown"
	}

	info.DYTCBinary = fmt.Sprintf("%032b", info.DYTCValue)
	info.EnableFuncHex = fmt.Sprintf("0x%08X", info.EnableFuncValue)
	info.EnableFuncPolicies = parseEnableFunc(info.EnableFuncValue)
	info.Features = buildFeatureList(info.EnableFuncValue, info.DYTCValue)
	return info
}

// readRegString reads a String value from the registry
func readRegString(root registry.Key, path, name, defaultVal string) string {
	k, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return defaultVal
	}
	defer k.Close()
	val, _, err := k.GetStringValue(name)
	if err != nil {
		return defaultVal
	}
	return val
}

// calcMainVersion returns the main version string
func calcMainVersion(driverVer string) string {
	var major, minor uint32
	fmt.Sscanf(driverVer, "%d.%d", &major, &minor)
	combined := (major << 16) | minor
	if combined >= 0x20000 {
		return "3.0"
	} else if combined >= 0x10000 {
		return "2.0 with SE"
	}
	return "2.0"
}

// readModeCheckReg reads a DWORD from the Dispatcher registry path
func readModeCheckReg(name string, defaultVal uint32) uint32 {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return defaultVal
	}
	defer k.Close()
	val, _, err := k.GetIntegerValue(name)
	if err != nil {
		return defaultVal
	}
	return uint32(val)
}

// parseEnableFunc parses Policy_EnableFunc value into policy list
func parseEnableFunc(value uint32) []EnableFuncPolicy {
	policies := []struct {
		bit  int
		name string
		desc string
	}{
		{0, "EnablePLxMonitor", "Enable PLx monitoring"},
		{1, "EnableSetPLxDouble", "Set PLx double interval"},
		{2, "EnableSetPLxFGChange", "Set PLx on FG change"},
		{3, "EnableResetPLxForce", "Reset PLx force"},
		{4, "EnableOSPowerSyncForce", "Force OS power sync"},
		{5, "EnableOSPowerSync", "Enable OS power sync"},
		{6, "DisableFGEPP_FREQUENCY", "No FG EPP/frequency setting"},
		{7, "DisableG1G2APPTURBOEPP", "No App Turbo EPP/frequency"},
		{8, "DisableDCCOPSTControl", "DCC OPST level not set to 1"},
		{9, "Reserved", "Reserved bit"},
		{10, "DisableG1G2APPLongEPP", "G1G2 App Turbo end not set long EPP"},
		{11, "SendFGAPPInfo2AIChip", "Send app info to AI chip"},
		{12, "WUDefenderAQMtoSTD", "Windows Defender FG to AQM"},
		{13, "AIScheduling", "AI/LLM state cooperation"},
		{14, "iDisableTurbo", "AI/AIGC cooperation"},
		{15, "DisableDynamicTurbo", "Excel dynamic Turbo disabled"},
		{16, "ZipTurbo", "7z compress/decompress Turbo"},
		{17, "bDoAudioCheck", "AIGC audio check enable/disable"},
		{18, "bDisableEPMWork", "EPM work disabled"},
		{19, "bDisableMLPipe", "ML pipeline disabled"},
		{20, "bEnableSETurboInfo", "SE pipeline enabled"},
		{21, "EnableLPECore", "Teams/Edge to LPE core"},
		{22, "Reserved1", "Reserved bit 1"},
		{23, "Reserved2", "Reserved bit 2"},
	}
	result := make([]EnableFuncPolicy, len(policies))
	for i, p := range policies {
		result[i] = EnableFuncPolicy{
			Bit:     p.bit,
			Name:    p.name,
			Desc:    p.desc,
			Enabled: (value>>uint(p.bit))&1 == 1,
		}
	}
	return result
}

// buildFeatureList builds the feature list from DYTC and EnableFunc values
func buildFeatureList(enableFunc, dytcFunc uint32) []ModeCheckFeature {
	check := func(val uint32, bit int) string {
		if (val>>uint(bit))&1 == 1 {
			return "Y"
		}
		return "N"
	}
	return []ModeCheckFeature{
		{"Stop Dispatcher PPM", check(dytcFunc, 4), "Stop PPM / Use DTT EPO"},
		{"Power mode 2-way sync", check(dytcFunc, 21), "FN+Q 2-way sync"},
		{"BSM Dolby control", check(dytcFunc, 26), "Enable Dolby control"},
		{"Process create boost", check(dytcFunc, 14), "APM background process turbo"},
		{"Smart SSD mode switch", check(dytcFunc, 16), "SSD whitelist perf"},
		{"Memory Auto clean", check(dytcFunc, 20), "Auto memory clean"},
		{"Send FPS to AI chip", check(dytcFunc, 7), "Send FPS to AI chip"},
		{"Enable DNPU Turbo", check(dytcFunc, 27), "Enable DNPU turbo"},
		{"Enable dGPU Plug ctrl", check(dytcFunc, 30), "Enable dGPU plug/unplug control"},
		{"NVIDIA GPU OC", check(dytcFunc, 31), "NVIDIA GPU overclocking"},
		{"AI Scheduling", check(enableFunc, 13), "Enable AI/LLM state cooperation"},
		{"LPE Core Assign", check(enableFunc, 21), "Teams/Edge to LPE core"},
		{"Zip Turbo", check(enableFunc, 16), "7z compress/decompress Turbo"},
		{"SendFGAPP to AI", check(enableFunc, 11), "Send app info to AI chip"},
	}
}
