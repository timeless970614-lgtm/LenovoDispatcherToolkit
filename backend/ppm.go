//go:build windows

package backend

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

const (
	highPerfScheme    = "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"
	balancedScheme    = "381b4222-f694-41f0-9685-ff5bb260df2e"
	processorSubgroup = "54533251-82be-4824-96c1-47b60b740d00"
)

// PPMSetting holds AC and DC values for a power setting
type PPMSetting struct {
	Name    string `json:"name"`
	GUID    string `json:"guid"`
	ACValue int    `json:"acValue"`
	DCValue int    `json:"dcValue"`
	Found   bool   `json:"found"`
}

// PPMSettings holds all readable PPM parameters
type PPMSettings struct {
	EPP        PPMSetting `json:"epp"`
	EPP1       PPMSetting `json:"epp1"`
	HeteroInc  PPMSetting `json:"heteroInc"`
	HeteroDec  PPMSetting `json:"heteroDec"`
	MaxFreq    PPMSetting `json:"maxFreq"`
	MaxFreq1   PPMSetting `json:"maxFreq1"`
	SoftPark   PPMSetting `json:"softPark"`
	CPMinCores PPMSetting `json:"cpMinCores"`
	MinPerf    PPMSetting `json:"minPerf"`
	MaxPerf    PPMSetting `json:"maxPerf"`
	SchemeName string     `json:"schemeName"`
	SchemeGUID string     `json:"schemeGUID"`
}

// queryPowerSetting reads AC and DC values for a specific power setting GUID
func queryPowerSetting(subgroupGUID, settingGUID string) (acVal int, dcVal int, found bool) {
	cmd := exec.Command("powercfg", "/query", "SCHEME_CURRENT", subgroupGUID, settingGUID)
	output, err := cmd.CombinedOutput()
	if err != nil || len(output) == 0 {
		return 0, 0, false
	}
	text := string(output)
	if !strings.Contains(text, "Current AC Power Setting Index") {
		return 0, 0, false
	}
	reAC := regexp.MustCompile(`Current AC Power Setting Index:\s+0x([0-9a-fA-F]+)`)
	reDC := regexp.MustCompile(`Current DC Power Setting Index:\s+0x([0-9a-fA-F]+)`)
	acMatch := reAC.FindStringSubmatch(text)
	dcMatch := reDC.FindStringSubmatch(text)
	if len(acMatch) > 1 {
		v, _ := strconv.ParseInt(acMatch[1], 16, 64)
		acVal = int(v)
	}
	if len(dcMatch) > 1 {
		v, _ := strconv.ParseInt(dcMatch[1], 16, 64)
		dcVal = int(v)
	}
	return acVal, dcVal, true
}

// querySchemeInfo reads the current power scheme name
func querySchemeInfo() (name string, guid string) {
	cmd := exec.Command("powercfg", "/getactivescheme")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Unknown", ""
	}
	text := string(output)
	reGUID := regexp.MustCompile(`Power Scheme GUID:\s+([0-9a-f-]+)\s+\(([^)]+)\)`)
	m := reGUID.FindStringSubmatch(text)
	if len(m) > 2 {
		return strings.TrimSpace(m[2]), strings.TrimSpace(m[1])
	}
	return "Unknown", ""
}

// GetPPMSettings reads all current PPM power settings
func GetPPMSettings() PPMSettings {
	const sub = processorSubgroup
	s := PPMSettings{}
	s.SchemeName, s.SchemeGUID = querySchemeInfo()

	ac, dc, found := queryPowerSetting(sub, "36687f9e-e3a5-4dbf-b1dc-15eb381c6863")
	s.EPP = PPMSetting{Name: "EPP (P-Core)", GUID: "36687f9e-e3a5-4dbf-b1dc-15eb381c6863", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "36687f9e-e3a5-4dbf-b1dc-15eb381c6864")
	s.EPP1 = PPMSetting{Name: "EPP1 (E-Core)", GUID: "36687f9e-e3a5-4dbf-b1dc-15eb381c6864", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "06cadf0e-64ed-448a-8927-ce7bf90eb35d")
	s.HeteroInc = PPMSetting{Name: "Hetero Increase Threshold", GUID: "06cadf0e-64ed-448a-8927-ce7bf90eb35d", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "12a0ab44-fe28-4fa9-b3bd-4b64f44960a6")
	s.HeteroDec = PPMSetting{Name: "Hetero Decrease Threshold", GUID: "12a0ab44-fe28-4fa9-b3bd-4b64f44960a6", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "75b0ae3f-bce0-45a7-8c89-c9611c25e100")
	s.MaxFreq = PPMSetting{Name: "Max Frequency (P-Core)", GUID: "75b0ae3f-bce0-45a7-8c89-c9611c25e100", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "75b0ae3f-bce0-45a7-8c89-c9611c25e101")
	s.MaxFreq1 = PPMSetting{Name: "Max Frequency (E-Core)", GUID: "75b0ae3f-bce0-45a7-8c89-c9611c25e101", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "97cfac41-2217-47eb-992d-618b1977c907")
	s.SoftPark = PPMSetting{Name: "Soft Park Latency", GUID: "97cfac41-2217-47eb-992d-618b1977c907", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "893dee8e-2bef-41e0-89c6-b55d0929964c")
	s.CPMinCores = PPMSetting{Name: "Min Processor State", GUID: "893dee8e-2bef-41e0-89c6-b55d0929964c", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "bc5038f7-23e0-4960-96da-33abaf5935ec")
	s.MaxPerf = PPMSetting{Name: "Max Processor State", GUID: "bc5038f7-23e0-4960-96da-33abaf5935ec", ACValue: ac, DCValue: dc, Found: found}

	return s
}

// runPowercfg executes a powercfg command
func runPowercfg(args ...string) error {
	cmd := exec.Command("powercfg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powercfg %v failed: %w\nOutput: %s", args, err, string(output))
	}
	return nil
}

// SetPowerSettingRaw sets a power setting by GUID with raw AC/DC values
func SetPowerSettingRaw(settingGUID string, acValue, dcValue int) error {
	if err := runPowercfg("/setacvalueindex", "SCHEME_CURRENT", processorSubgroup, settingGUID, fmt.Sprintf("%d", acValue)); err != nil {
		return err
	}
	if err := runPowercfg("/setdcvalueindex", "SCHEME_CURRENT", processorSubgroup, settingGUID, fmt.Sprintf("%d", dcValue)); err != nil {
		return err
	}
	return runPowercfg("/setactive", "SCHEME_CURRENT")
}

// ApplyHetero applies Hetero scheduling thresholds
func ApplyHetero(increase, decrease int) error {
	multiplier := 17040385
	heteroIncreaseGUID := "06cadf0e-64ed-448a-8927-ce7bf90eb35d"
	heteroDecreaseGUID := "12a0ab44-fe28-4fa9-b3bd-4b64f44960a6"
	if err := runPowercfg("/setacvalueindex", "SCHEME_CURRENT", processorSubgroup, heteroIncreaseGUID, fmt.Sprintf("%d", increase*multiplier)); err != nil {
		return err
	}
	if err := runPowercfg("/setdcvalueindex", "SCHEME_CURRENT", processorSubgroup, heteroIncreaseGUID, fmt.Sprintf("%d", increase*multiplier)); err != nil {
		return err
	}
	if err := runPowercfg("/setacvalueindex", "SCHEME_CURRENT", processorSubgroup, heteroDecreaseGUID, fmt.Sprintf("%d", decrease*multiplier)); err != nil {
		return err
	}
	if err := runPowercfg("/setdcvalueindex", "SCHEME_CURRENT", processorSubgroup, heteroDecreaseGUID, fmt.Sprintf("%d", decrease*multiplier)); err != nil {
		return err
	}
	return runPowercfg("/setactive", "SCHEME_CURRENT")
}

// ApplyEPP applies Energy Performance Preference settings
func ApplyEPP(epp, epp1 int) error {
	perfEPPGUID := "36687f9e-e3a5-4dbf-b1dc-15eb381c6863"
	if err := runPowercfg("/setacvalueindex", "SCHEME_CURRENT", processorSubgroup, perfEPPGUID, fmt.Sprintf("%d", epp)); err != nil {
		return err
	}
	if err := runPowercfg("/setdcvalueindex", "SCHEME_CURRENT", processorSubgroup, perfEPPGUID, fmt.Sprintf("%d", epp1)); err != nil {
		return err
	}
	return runPowercfg("/setactive", "SCHEME_CURRENT")
}

// ApplyMaxFrequency applies maximum processor frequency settings
func ApplyMaxFrequency(freq, freq1 int) error {
	procFreqMaxGUID := "75b0ae3f-bce0-45a7-8c89-c9611c25e100"
	if err := runPowercfg("/setacvalueindex", "SCHEME_CURRENT", processorSubgroup, procFreqMaxGUID, fmt.Sprintf("%d", freq)); err != nil {
		return err
	}
	if err := runPowercfg("/setdcvalueindex", "SCHEME_CURRENT", processorSubgroup, procFreqMaxGUID, fmt.Sprintf("%d", freq1)); err != nil {
		return err
	}
	return runPowercfg("/setactive", "SCHEME_CURRENT")
}

// ApplySoftParkLatency applies SoftParkLatency settings
func ApplySoftParkLatency(ac, dc int) error {
	softParkGUID := "97cfac41-2217-47eb-992d-618b1977c907"
	if err := runPowercfg("/setacvalueindex", "SCHEME_CURRENT", processorSubgroup, softParkGUID, fmt.Sprintf("%d", ac)); err != nil {
		return err
	}
	if err := runPowercfg("/setdcvalueindex", "SCHEME_CURRENT", processorSubgroup, softParkGUID, fmt.Sprintf("%d", dc)); err != nil {
		return err
	}
	return runPowercfg("/setactive", "SCHEME_CURRENT")
}

// RestoreDefaults restores default power schemes
func RestoreDefaults() error {
	runPowercfg("/setactive", highPerfScheme)
	runPowercfg("-restoredefaultschemes")
	return runPowercfg("/setactive", balancedScheme)
}
