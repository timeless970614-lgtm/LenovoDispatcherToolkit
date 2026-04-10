//go:build windows

package backend

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"
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

// PPMInfo is defined in ipf.go (same package, no redeclaration needed).

// ── powercfg (primary, works without MSR driver) ─────────────────────────

func querySchemeInfo() (name string, guid string) {
	cmd := hiddenCmd("powercfg", "/getactivescheme")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "Unknown", ""
	}
	reGUID := regexp.MustCompile(`Power Scheme GUID:\s+([0-9a-f-]+)\s+\(([^)]+)\)`)
	m := reGUID.FindStringSubmatch(string(output))
	if len(m) > 2 {
		return strings.TrimSpace(m[2]), strings.TrimSpace(m[1])
	}
	return "Unknown", ""
}

func queryPowerSetting(subgroupGUID, settingGUID string) (acVal int, dcVal int, found bool) {
	cmd := hiddenCmd("powercfg", "/query", "SCHEME_CURRENT", subgroupGUID, settingGUID)
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

func runPowercfg(args ...string) error {
	cmd := hiddenCmd("powercfg", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powercfg %v failed: %w\nOutput: %s", args, err, string(output))
	}
	return nil
}

// ── MSR access via ipf_wrapper.dll (WinMSRIO.dll + MSRIO.sys) ─────────────

var (
	msrDLL   *syscall.LazyDLL
	msrFunc  struct {
		readMsr          uintptr
		getEPP           uintptr
		getEPP1         uintptr
		getFrequencyLimit uintptr
		getHeteroInc     uintptr
		getHeteroDec     uintptr
		getSoftPark      uintptr
	}
	msrOnce sync.Once
	msrErr  error
)

func initMSR() error {
	msrOnce.Do(func() {
		exeDir := getExeDir()
		dllPath := filepath.Join(exeDir, "ipf_wrapper.dll")
		msrDLL = syscall.NewLazyDLL(dllPath)

		resolve := func(name string) uintptr {
			p := msrDLL.NewProc(name)
			if err := p.Find(); err != nil {
				return 0
			}
			return p.Addr()
		}

		msrFunc.readMsr = resolve("IPF_ReadMSR")
		msrFunc.getEPP = resolve("IPF_GetEPP")
		msrFunc.getEPP1 = resolve("IPF_GetEPP1")
		msrFunc.getFrequencyLimit = resolve("IPF_GetFrequencyLimit_MHz")
		msrFunc.getHeteroInc = resolve("IPF_GetHeteroInc")
		msrFunc.getHeteroDec = resolve("IPF_GetHeteroDec")
		msrFunc.getSoftPark = resolve("IPF_GetSoftParkLatency")

		// Try to connect via ipf_wrapper (initializes WinMSRIO)
		if connectFn := resolve("IPF_Connect"); connectFn != 0 {
			syscall.Syscall(connectFn, 0, 0, 0, 0)
		}

		// Test MSR access by reading a known-safe register (IA32_PLATFORM_ID, MSR 0x17)
		if msrFunc.readMsr != 0 {
			var eax, edx uint32
			ok, _, _ := syscall.Syscall(
				msrFunc.readMsr,
				3,
				uintptr(0x17),
				uintptr(unsafe.Pointer(&eax)),
				uintptr(unsafe.Pointer(&edx)),
			)
			if ok == 0 {
				msrErr = fmt.Errorf("MSR access test failed; MSRIO driver may not be installed")
			}
		} else {
			msrErr = fmt.Errorf("ipf_wrapper.dll does not export MSR functions")
		}
	})
	return msrErr
}

// GetPPMSettings reads all PPM values via powercfg (primary, no driver needed).
// NOTE: Some GUIDs (EPP, Hetero, Freq, SoftPark) are writable but not queryable
// on certain Intel platforms. We treat them as Found=true so the UI allows editing.
func GetPPMSettings() PPMSettings {
	const sub = processorSubgroup
	s := PPMSettings{}
	s.SchemeName, s.SchemeGUID = querySchemeInfo()

	ac, dc, found := queryPowerSetting(sub, "36687f9e-e3a5-4dbf-b1dc-15eb381c6863")
	s.EPP = PPMSetting{Name: "EPP (P-Core)", GUID: "36687f9e-e3a5-4dbf-b1dc-15eb381c6863", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "36687f9e-e3a5-4dbf-b1dc-15eb381c6864")
	s.EPP1 = PPMSetting{Name: "EPP1 (E-Core)", GUID: "36687f9e-e3a5-4dbf-b1dc-15eb381c6864", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "06cadf0e-64ed-448a-8927-ce7bf90eb35d")
	s.HeteroInc = PPMSetting{Name: "Hetero Increase Threshold", GUID: "06cadf0e-64ed-448a-8927-ce7bf90eb35d", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "12a0ab44-fe28-4fa9-b3bd-4b64f44960a6")
	s.HeteroDec = PPMSetting{Name: "Hetero Decrease Threshold", GUID: "12a0ab44-fe28-4fa9-b3bd-4b64f44960a6", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "75b0ae3f-bce0-45a7-8c89-c9611c25e100")
	s.MaxFreq = PPMSetting{Name: "Max Frequency (P-Core)", GUID: "75b0ae3f-bce0-45a7-8c89-c9611c25e100", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "75b0ae3f-bce0-45a7-8c89-c9611c25e101")
	s.MaxFreq1 = PPMSetting{Name: "Max Frequency (E-Core)", GUID: "75b0ae3f-bce0-45a7-8c89-c9611c25e101", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "97cfac41-2217-47eb-992d-618b1977c907")
	s.SoftPark = PPMSetting{Name: "Soft Park Latency", GUID: "97cfac41-2217-47eb-992d-618b1977c907", ACValue: ac, DCValue: dc, Found: true}

	ac, dc, found = queryPowerSetting(sub, "893dee8e-2bef-41e0-89c6-b55d0929964c")
	s.CPMinCores = PPMSetting{Name: "Min Processor Cores", GUID: "893dee8e-2bef-41e0-89c6-b55d0929964c", ACValue: ac, DCValue: dc, Found: found}

	ac, dc, found = queryPowerSetting(sub, "bc5038f7-23e0-4960-96da-33abaf5935ec")
	s.MaxPerf = PPMSetting{Name: "Max Processor State", GUID: "bc5038f7-23e0-4960-96da-33abaf5935ec", ACValue: ac, DCValue: dc, Found: found}

	return s
}

// GetPPMInfo reads raw MSR-based PPM values (EPP, Freq, Hetero, SoftPark).
// Falls back to powercfg values if MSR access is unavailable.
func GetPPMInfo() PPMInfo {
	info := PPMInfo{}

	// Try MSR first (requires WinMSRIO driver)
	if err := initMSR(); err == nil {
		callU32 := func(fn uintptr) uint32 {
			if fn == 0 {
				return 0
			}
			v, _, _ := syscall.Syscall(fn, 0, 0, 0, 0)
			return uint32(v)
		}
		info.EPP = callU32(msrFunc.getEPP)
		info.EPP1 = callU32(msrFunc.getEPP1)
		info.FrequencyLimit = callU32(msrFunc.getFrequencyLimit)
		info.HeteroInc = callU32(msrFunc.getHeteroInc)
		info.HeteroDec = callU32(msrFunc.getHeteroDec)
		info.SoftParkLatency = callU32(msrFunc.getSoftPark)
		return info
	}

	// Fallback: read from powercfg (no driver needed)
	ac, _, found := queryPowerSetting(processorSubgroup, "36687f9e-e3a5-4dbf-b1dc-15eb381c6863")
	if found {
		info.EPP = uint32(ac)
	}
	ac, _, found = queryPowerSetting(processorSubgroup, "36687f9e-e3a5-4dbf-b1dc-15eb381c6864")
	if found {
		info.EPP1 = uint32(ac)
	}
	ac, _, found = queryPowerSetting(processorSubgroup, "75b0ae3f-bce0-45a7-8c89-c9611c25e100")
	if found {
		info.FrequencyLimit = uint32(ac)
	}
	ac, _, found = queryPowerSetting(processorSubgroup, "06cadf0e-64ed-448a-8927-ce7bf90eb35d")
	if found {
		info.HeteroInc = uint32(ac)
	}
	ac, _, found = queryPowerSetting(processorSubgroup, "12a0ab44-fe28-4fa9-b3bd-4b64f44960a6")
	if found {
		info.HeteroDec = uint32(ac)
	}
	ac, _, found = queryPowerSetting(processorSubgroup, "97cfac41-2217-47eb-992d-618b1977c907")
	if found {
		info.SoftParkLatency = uint32(ac)
	}

	return info
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

// EPOTStatus holds ML_Scenario EPOT parameters (columns 46-53)
type EPOTStatus struct {
	EPOT                  uint32 `json:"epot"`
	EPP                   uint32 `json:"epp"`
	EPP1                  uint32 `json:"epp1"`
	PPMFrequencyLimit     uint32 `json:"ppmFrequencyLimit"`
	PPMFrequencyLimit1    uint32 `json:"ppmFrequencyLimit1"`
	PPMCpMin              uint32 `json:"ppmCpMin"`
	PPMCpMax              uint32 `json:"ppmCpMax"`
	SoftParking           uint32 `json:"softParking"`
}

// GetEPOTStatus reads ML_Scenario EPOT parameters from multiple sources:
// - EPOT: IPF DLL _IPFV2_CurrentGear() (same as ML_Scenario reads)
// - EPP, EPP1, FrequencyLimit, SoftParking: IPF DLL (MSR-based)
// - PPMCpMin, PPMCpMax: powercfg (Windows power settings)
func GetEPOTStatus() EPOTStatus {
	status := EPOTStatus{}
	
	// Read EPOT/Gear from IPF DLL using _IPFV2_CurrentGear()
	// This is the same API used by ML_Scenario GPUCounter.cpp:5282
	// _CurrentGear = _IPFV2_CurrentGear();
	InitIPF() // Ensure IPF is connected
	gear := GetCurrentGear()
	if gear >= 0 && gear <= 255 {
		status.EPOT = uint32(gear)
	} else {
		// Fallback: try reading from registry
		readRegistryDWord := func(name string) uint32 {
			script := fmt.Sprintf(`
$ErrorActionPreference='SilentlyContinue'
$v = (Get-ItemProperty 'HKLM:\SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider' -Name '%s' -ErrorAction SilentlyContinue).'%s'
if ($null -eq $v) { $v = 0 }
Write-Output $v
`, name, name)
			cmd := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script)
			out, err := cmd.Output()
			if err != nil {
				return 0
			}
			var v uint32
			fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &v)
			return v
		}
		status.EPOT = readRegistryDWord("EPOT")
	}
	
	// Read EPP, EPP1, FrequencyLimit, SoftParking from IPF DLL
	ppmInfo := ReadPPM()
	status.EPP = ppmInfo.EPP
	status.EPP1 = ppmInfo.EPP1
	status.PPMFrequencyLimit = ppmInfo.FrequencyLimit
	status.SoftParking = ppmInfo.SoftParkLatency
	
	// Read PPMFrequencyLimit1 from IPF DLL (if available, otherwise use same as P-Core)
	// Note: IPF DLL may not have separate E-Core frequency limit, use P-Core value as fallback
	status.PPMFrequencyLimit1 = ppmInfo.FrequencyLimit
	
	// Read PPMCpMin and PPMCpMax from powercfg
	// These are Windows power settings for core parking
	readPowercfgValue := func(subgroup string, setting string, ac bool) uint32 {
		acdc := "/SETACVALUEINDEX"
		if !ac {
			acdc = "/SETDCVALUEINDEX"
		}
		_ = acdc // suppress unused warning for now
		
		// First get the active scheme
		schemeCmd := hiddenCmd("powercfg", "/getactivescheme")
		schemeOut, err := schemeCmd.Output()
		if err != nil {
			return 0
		}
		schemeLine := strings.TrimSpace(string(schemeOut))
		// Parse scheme GUID from output like "Power Scheme GUID: 381b4222-f694-41f0-9685-ff5bb260df2e  (Balanced)"
		var schemeGUID string
		if idx := strings.Index(schemeLine, "GUID: "); idx != -1 {
			start := idx + 6
			end := strings.Index(schemeLine[start:], " ")
			if end != -1 {
				schemeGUID = schemeLine[start : start+end]
			}
		}
		if schemeGUID == "" {
			return 0
		}
		
		// Query the setting value
		queryCmd := hiddenCmd("powercfg", "/query", schemeGUID, subgroup, setting)
		queryOut, err := queryCmd.Output()
		if err != nil {
			return 0
		}
		
		// Parse the value from output
		var value uint32
		lines := strings.Split(string(queryOut), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if strings.HasPrefix(line, "Current AC Power Setting Index: ") {
				if ac {
					fmt.Sscanf(strings.TrimPrefix(line, "Current AC Power Setting Index: "), "%x", &value)
					break
				}
			}
			if strings.HasPrefix(line, "Current DC Power Setting Index: ") {
				if !ac {
					fmt.Sscanf(strings.TrimPrefix(line, "Current DC Power Setting Index: "), "%x", &value)
					break
				}
			}
		}
		return value
	}
	
	// GUIDs for processor settings
	const (
		processorSubgroup   = "54533251-82be-4824-96c1-47b60b740d00"
		cpMinCoresSetting   = "0cc5b647-c1df-4637-891a-dec35c318583" // Core parking min cores
		cpMaxCoresSetting   = "ea062031-0e34-4ff1-9b6d-eb1059334028" // Core parking max cores
	)
	
	// Read AC values (you may want to check if on battery and read DC instead)
	status.PPMCpMin = readPowercfgValue(processorSubgroup, cpMinCoresSetting, true)
	status.PPMCpMax = readPowercfgValue(processorSubgroup, cpMaxCoresSetting, true)
	
	return status
}
