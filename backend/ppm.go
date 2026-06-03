//go:build windows

package backend

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// PPMSetting represents a single power setting value
type PPMSetting struct {
	Name    string `json:"name"`
	GUID    string `json:"guid"`
	ACValue uint32 `json:"acValue"`
	DCValue uint32 `json:"dcValue"`
	Found   bool   `json:"found"`
}

// PPMSettings represents all PPM power settings
type PPMSettings struct {
	SchemeName string      `json:"schemeName"`
	SchemeGUID string      `json:"schemeGUID"`
	MinPerf    PPMSetting  `json:"minPerf"`
	MaxPerf    PPMSetting  `json:"maxPerf"`
	CPMinCores PPMSetting  `json:"cpMinCores"`
	EPP        PPMSetting  `json:"epp"`
	EPP1       PPMSetting  `json:"epp1"`
	HeteroInc  PPMSetting  `json:"heteroInc"`
	HeteroDec  PPMSetting  `json:"heteroDec"`
	MaxFreq    PPMSetting  `json:"maxFreq"`
	MaxFreq1   PPMSetting  `json:"maxFreq1"`
	SoftPark   PPMSetting  `json:"softPark"`
}

// PPM setting aliases and GUIDs
var ppmSettings = map[string]struct {
	alias string
	guid  string
	name  string
}{
	"minPerf":    {"PROCTHROTTLEMIN", "893dee8e-2bef-41e0-89c6-b55d092996a4", "Min Processor State"},
	"maxPerf":    {"PROCTHROTTLEMAX", "bc5038f7-23e0-4960-96da-33abaf5935ec", "Max Processor State"},
	"cpMinCores": {"CPMINCORES", "0cc5b647-c1df-4637-891a-dec35c318583", "Min Processor Cores"},
	"epp":        {"PERFEPP", "ea062031-0e34-4ff1-9b6d-eb1059334028", "EPP (P-Core)"},
	"epp1":       {"PERFEPP1", "91a8d7a5-9396-4b52-b188-5dd9b7d6f160", "EPP1 (E-Core)"},
	"heteroInc":  {"HETEROINCREASETHRESHOLD", "b000397d-9b0b-483d-98c9-692a6064cf55", "Hetero Increase Threshold"},
	"heteroDec":  {"HETERODECREASETHRESHOLD", "7f2f5cfa-f10c-4823-b5e1-e943e45c3637", "Hetero Decrease Threshold"},
	"maxFreq":    {"PROCFREQMAX", "75b0ae3f-bce0-45a7-8c89-c9611c25e100", "Max Frequency (P-Core)"},
	"maxFreq1":   {"PROCFREQMAX1", "bc6eb96b-bd1d-4a53-8c8d-51c35a0e7070", "Max Frequency (E-Core)"},
	"softPark":   {"SOFTPARKLATENCY", "943fd819-38c1-4f5d-8c8e-06d2c9a80e3f", "Soft Park Latency"},
}

// GetPPMSettings retrieves current PPM power settings
func GetPPMSettings() *PPMSettings {
	settings := &PPMSettings{}

	// Get active power scheme
	schemeCmd := exec.Command("powercfg", "/getactivescheme")
	schemeOutput, err := schemeCmd.CombinedOutput()
	if err == nil {
		// Parse: "Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)"
		re := regexp.MustCompile(`Power Scheme GUID: ([a-f0-9-]+)\s+\(([^)]+)\)`)
		matches := re.FindStringSubmatch(string(schemeOutput))
		if len(matches) >= 3 {
			settings.SchemeGUID = matches[1]
			settings.SchemeName = matches[2]
		}
	}

	// Get each setting using alias
	settings.MinPerf = getPPMSettingByAlias("minPerf")
	settings.MaxPerf = getPPMSettingByAlias("maxPerf")
	settings.CPMinCores = getPPMSettingByAlias("cpMinCores")
	settings.EPP = getPPMSettingByAlias("epp")
	settings.EPP1 = getPPMSettingByAlias("epp1")
	settings.HeteroInc = getPPMSettingByAlias("heteroInc")
	settings.HeteroDec = getPPMSettingByAlias("heteroDec")
	settings.MaxFreq = getPPMSettingByAlias("maxFreq")
	settings.MaxFreq1 = getPPMSettingByAlias("maxFreq1")
	settings.SoftPark = getPPMSettingByAlias("softPark")

	return settings
}

func getPPMSettingByAlias(key string) PPMSetting {
	info, ok := ppmSettings[key]
	if !ok {
		return PPMSetting{Name: key, Found: false}
	}

	setting := PPMSetting{
		Name: info.name,
		GUID: info.guid,
	}

	// Query using alias
	cmd := exec.Command("powercfg", "/query", "SCHEME_CURRENT", "SUB_PROCESSOR", info.alias)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Try with GUID if alias fails
		cmd = exec.Command("powercfg", "/query", "SCHEME_CURRENT", "SUB_PROCESSOR", info.guid)
		output, err = cmd.CombinedOutput()
	}

	if err == nil {
		acVal, dcVal := parsePowerCfgOutputFull(string(output))
		if acVal != nil {
			setting.ACValue = *acVal
			setting.Found = true
		}
		if dcVal != nil {
			setting.DCValue = *dcVal
		}
	}

	return setting
}

func parsePowerCfgOutputFull(output string) (*uint32, *uint32) {
	var acVal, dcVal *uint32

	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.Contains(line, "Current AC Power Setting Index:") {
			// Extract hex value
			re := regexp.MustCompile(`0x([0-9a-fA-F]+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				var val uint32
				fmt.Sscanf(matches[1], "%x", &val)
				acVal = &val
			}
		} else if strings.Contains(line, "Current DC Power Setting Index:") {
			re := regexp.MustCompile(`0x([0-9a-fA-F]+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 2 {
				var val uint32
				fmt.Sscanf(matches[1], "%x", &val)
				dcVal = &val
			}
		}
	}

	return acVal, dcVal
}

// SetPowerSettingRaw sets a power setting value directly by GUID
func SetPowerSettingRaw(guid string, acValue, dcValue uint32) string {
	// Set AC value
	acCmd := exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", guid, fmt.Sprintf("%d", acValue))
	if err := acCmd.Run(); err != nil {
		return fmt.Sprintf("Failed to set AC value: %v", err)
	}

	// Set DC value
	dcCmd := exec.Command("powercfg", "/setdcvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", guid, fmt.Sprintf("%d", dcValue))
	if err := dcCmd.Run(); err != nil {
		return fmt.Sprintf("Failed to set DC value: %v", err)
	}

	// Apply settings
	applyCmd := exec.Command("powercfg", "/setactive", "SCHEME_CURRENT")
	if err := applyCmd.Run(); err != nil {
		return fmt.Sprintf("Failed to apply settings: %v", err)
	}

	return "OK"
}

// ApplyHetero applies Hetero scheduling settings
func ApplyHetero(increase, decrease int) error {
	incInfo := ppmSettings["heteroInc"]
	decInfo := ppmSettings["heteroDec"]

	// Set AC values
	cmd := exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", incInfo.guid, fmt.Sprintf("%d", increase))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set hetero increase: %v", err)
	}

	cmd = exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", decInfo.guid, fmt.Sprintf("%d", decrease))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set hetero decrease: %v", err)
	}

	// Apply
	cmd = exec.Command("powercfg", "/setactive", "SCHEME_CURRENT")
	return cmd.Run()
}

// ApplyEPP applies EPP settings
func ApplyEPP(epp, epp1 int) error {
	eppInfo := ppmSettings["epp"]
	epp1Info := ppmSettings["epp1"]

	// Set AC values
	cmd := exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", eppInfo.guid, fmt.Sprintf("%d", epp))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set EPP: %v", err)
	}

	cmd = exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", epp1Info.guid, fmt.Sprintf("%d", epp1))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set EPP1: %v", err)
	}

	// Apply
	cmd = exec.Command("powercfg", "/setactive", "SCHEME_CURRENT")
	return cmd.Run()
}

// ApplyMaxFrequency applies max frequency settings
func ApplyMaxFrequency(freq, freq1 int) error {
	freqInfo := ppmSettings["maxFreq"]
	freq1Info := ppmSettings["maxFreq1"]

	// Set AC values
	cmd := exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", freqInfo.guid, fmt.Sprintf("%d", freq))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set max frequency: %v", err)
	}

	cmd = exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", freq1Info.guid, fmt.Sprintf("%d", freq1))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set max frequency1: %v", err)
	}

	// Apply
	cmd = exec.Command("powercfg", "/setactive", "SCHEME_CURRENT")
	return cmd.Run()
}

// ApplySoftParkLatency applies soft park latency setting
func ApplySoftParkLatency(ac, dc int) error {
	softParkInfo := ppmSettings["softPark"]

	// Set AC value
	cmd := exec.Command("powercfg", "/setacvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", softParkInfo.guid, fmt.Sprintf("%d", ac))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set soft park latency AC: %v", err)
	}

	// Set DC value
	cmd = exec.Command("powercfg", "/setdcvalueindex", "SCHEME_CURRENT", "SUB_PROCESSOR", softParkInfo.guid, fmt.Sprintf("%d", dc))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to set soft park latency DC: %v", err)
	}

	// Apply
	cmd = exec.Command("powercfg", "/setactive", "SCHEME_CURRENT")
	return cmd.Run()
}

// RestoreDefaults restores default processor power settings
func RestoreDefaults() error {
	// Use powercfg to restore default settings for the current scheme
	cmd := exec.Command("powercfg", "/restoredefaultschemes")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to restore defaults: %v", err)
	}
	return nil
}
