//go:build windows

package backend

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// WMIExplorerResult holds the complete WMI exploration output
type WMIExplorerResult struct {
	Namespaces []string          `json:"namespaces"`
	Classes    []WMIClassInfo    `json:"classes"`
	Summary    WMISummary        `json:"summary"`
	Error      string            `json:"error,omitempty"`
}

// WMIClassInfo holds information about a WMI class
type WMIClassInfo struct {
	Name        string        `json:"name"`
	Namespace   string        `json:"namespace"`
	Description string        `json:"description"`
	Methods     []WMIMethod   `json:"methods"`
	Properties  []string      `json:"properties"`
	Category    string        `json:"category"` // "BIOS", "Power", "Fan", "Performance", "Display", "Other"
}

// WMIMethod holds information about a WMI method
type WMIMethod struct {
	Name       string   `json:"name"`
	Prototype  string   `json:"prototype"` // e.g., "uint32 SetMode(uint32 Mode)"
	Parameters []string `json:"parameters"` // e.g., ["Mode: uint32"]
	ReturnType string   `json:"returnType"`
	DLLFunc    string   `json:"dllFunc"` // mapped DLL function name
	DLLFile    string   `json:"dllFile"` // e.g., "LenovoDYTC.dll"
	Notes      string   `json:"notes"`
}

// WMISummary holds a quick summary of available features
type WMISummary struct {
	TotalClasses     int      `json:"totalClasses"`
	TotalMethods     int      `json:"totalMethods"`
	AvailableFeatures []FeatureInfo `json:"availableFeatures"`
}

// FeatureInfo is a high-level feature summary
type FeatureInfo struct {
	FeatureName string   `json:"featureName"`
	WMIClass    string   `json:"wmiClass"`
	WMIMethod   string   `json:"wmiMethod"`
	DLLFunction string   `json:"dllFunction"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
}

// GetWMIExplorer returns a comprehensive overview of available WMI BIOS methods
func GetWMIExplorer() string {
	result := &WMIExplorerResult{}

	// Try to enumerate WMI classes via PowerShell
	psScript := `
$ErrorActionPreference = 'SilentlyContinue'
$classes = @()

# ── Lenovo-specific WMI namespaces ──
$namespaces = @(
    'root\wmi',
    'root\WMI',
    'root\lenovo',
    'root\LENOVO',
    'root\ccm',
    'root\default'
)

$classBlacklist = @{
    '__' = $true
    'MSN_' = $true
    'Win32_' = $false  # allow Win32
    'CIM_' = $false
}

$allClasses = @()

foreach ($ns in $namespaces) {
    try {
        $nsClasses = Get-WmiObject -Namespace $ns -List 2>$null | Where-Object { $_.Name -notlike '__*' }
        foreach ($c in $nsClasses) {
            $name = $c.Name.ToString()
            # Filter: prefer Lenovo/LNV classes, allow Win32, filter junk
            if ($name -like 'LNV*' -or $name -like 'Lenovo*' -or $name -like 'WmiMonitor*' -or $name -like 'Acpi*' -or $name -like 'Thermal*' -or $name -like 'Battery*' -or $name -like 'Win32_BIOS' -or $name -like 'Win32_Processor' -or $name -like 'Win32_ComputerSystem' -or $name -like 'Win32_PnP*' -or $name -like 'Win32_BaseBoard' -or $name -like 'Win32_SystemBIOS' -or $name -like 'MSAcpi*') {
                $allClasses += [PSCustomObject]@{
                    Name = $name
                    Namespace = $ns
                }
            }
        }
    } catch {}
}

# Deduplicate
$allClasses = $allClasses | Sort-Object Name -Unique | Select-Object -First 40

# For each class, get methods
$results = @()
foreach ($c in $allClasses) {
    try {
        $methods = @(Get-WmiObject -Namespace $c.Namespace -Class $c.Name 2>$null | Get-Member -MemberType Method 2>$null | Select-Object -ExpandProperty Name)
        $props = @(Get-WmiObject -Namespace $c.Namespace -Class $c.Name 2>$null | Get-Member -MemberType Property 2>$null | Select-Object -ExpandProperty Name | Select-Object -First 5)
        if ($methods.Count -gt 0 -or $props.Count -gt 0) {
            $results += [PSCustomObject]@{
                Name = $c.Name
                Namespace = $c.Namespace
                Methods = @($methods)
                Properties = @($props)
            }
        }
    } catch {}
}

$results | ConvertTo-Json -Depth 5 -Compress
`

	out, err := runPowershellScript(psScript)
	if err != nil || strings.TrimSpace(out) == "" || strings.TrimSpace(out) == "null" {
		result.Error = "Unable to enumerate WMI classes. Some features may require elevated privileges."
	} else {
		// Parse the JSON output from PowerShell
		var rawClasses []map[string]interface{}
		if err := json.Unmarshal([]byte(out), &rawClasses); err == nil {
			for _, rc := range rawClasses {
				classInfo := WMIClassInfo{
					Name:        getString(rc, "Name"),
					Namespace:   getString(rc, "Namespace"),
					Description: getClassDescription(getString(rc, "Name")),
					Category:    categorizeWMI(getString(rc, "Name")),
				}
				if methods, ok := rc["Methods"].([]interface{}); ok {
					for _, m := range methods {
						if mn, ok := m.(string); ok {
							classInfo.Methods = append(classInfo.Methods, WMIMethod{
								Name:       mn,
								Prototype:  buildMethodPrototype(mn),
								DLLFunc:    mapWMIToDLLFunc(mn, classInfo.Name),
								DLLFile:    mapWMIToDLL(mn, classInfo.Name),
								Notes:      getMethodNotes(mn, classInfo.Name),
							})
						}
					}
				}
				if props, ok := rc["Properties"].([]interface{}); ok {
					for _, p := range props {
						if pn, ok := p.(string); ok {
							classInfo.Properties = append(classInfo.Properties, pn)
						}
					}
				}
				result.Classes = append(result.Classes, classInfo)
			}
		}
	}

	// Build summary features
	result.Summary = buildWMISummary(result.Classes)

	// Collect namespaces
	nsMap := make(map[string]bool)
	for _, c := range result.Classes {
		nsMap[c.Namespace] = true
	}
	for ns := range nsMap {
		result.Namespaces = append(result.Namespaces, ns)
	}
	sort.Strings(result.Namespaces)

	// Sort classes by category then name
	sort.Slice(result.Classes, func(i, j int) bool {
		if result.Classes[i].Category != result.Classes[j].Category {
			return categoryOrder(result.Classes[i].Category) < categoryOrder(result.Classes[j].Category)
		}
		return result.Classes[i].Name < result.Classes[j].Name
	})

	// Count
	result.Summary.TotalClasses = len(result.Classes)
	for _, c := range result.Classes {
		result.Summary.TotalMethods += len(c.Methods)
	}

	data, _ := json.Marshal(result)
	return string(data)
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

func getClassDescription(className string) string {
	descs := map[string]string{
		"LNV_BiosAttributes":              "Lenovo BIOS Attribute Settings",
		"LNV_DYTCInterface":                "Lenovo Dynamic Thermal & Power Control Interface",
		"LNV_FanControl":                   "Lenovo Fan Speed Control",
		"LNV_PowerManagement":              "Lenovo Power Management Settings",
		"WmiMonitorBrightnessMethods":       "WMI Monitor Brightness Control Methods",
		"WmiMonitorBrightness":              "WMI Monitor Brightness Status",
		"MSAcpi_ThermalZoneTemperature":   "ACPI Thermal Zone Temperature",
		"Win32_BIOS":                       "Windows BIOS Information",
		"Win32_Processor":                   "Windows Processor Information",
		"Win32_ComputerSystem":             "Windows Computer System Information",
		"Win32_BaseBoard":                   "Windows Baseboard Information",
	}
	if desc, ok := descs[className]; ok {
		return desc
	}
	return ""
}

func categorizeWMI(className string) string {
	lower := strings.ToLower(className)
	switch {
	case strings.Contains(lower, "bios"):
		return "BIOS"
	case strings.Contains(lower, "dytc") || strings.Contains(lower, "thermal") || strings.Contains(lower, "fan"):
		return "Power & Thermal"
	case strings.Contains(lower, "power") || strings.Contains(lower, "battery"):
		return "Battery"
	case strings.Contains(lower, "wmi") && strings.Contains(lower, "brightness"):
		return "Display"
	case strings.Contains(lower, "npu") || strings.Contains(lower, "gpu") || strings.Contains(lower, "video"):
		return "Graphics"
	case strings.Contains(lower, "win32_bios") || strings.Contains(lower, "win32_computer") || strings.Contains(lower, "win32_baseboard"):
		return "System Info"
	default:
		return "Other"
	}
}

func categoryOrder(cat string) int {
	order := map[string]int{
		"BIOS": 1, "Power & Thermal": 2, "Battery": 3,
		"Display": 4, "Graphics": 5, "System Info": 6, "Other": 7,
	}
	if o, ok := order[cat]; ok {
		return o
	}
	return 99
}

func buildMethodPrototype(methodName string) string {
	// Common BIOS/WMI method signatures
	sigs := map[string]string{
		"Set":                           "uint32 Set()",
		"Get":                           "uint32 Get()",
		"Enable":                        "uint32 Enable()",
		"Disable":                       "uint32 Disable()",
		"SetValue":                      "uint32 SetValue(uint32 Value)",
		"GetValue":                      "uint32 GetValue()",
		"WmiSetBrightness":              "uint32 WmiSetBrightness(uint8 Timeout, uint8 Brightness)",
		"WmiGetBrightness":              "uint8 WmiGetBrightness()",
		"Set_FanMode":                   "uint32 Set_FanMode(uint32 Mode)",
		"Set_DYTCMode":                  "uint32 Set_DYTCMode(uint32 Mode)",
		"GET_Cap_DCC":                   "uint32 GET_Cap_DCC()",
		"Set_GEEKMode":                  "uint32 Set_GEEKMode(uint32 OnOff)",
		"GET_Cap_GEEK":                  "uint32 GET_Cap_GEEK()",
		"Set_ODVMode":                   "uint32 Set_ODVMode(uint32 Index, uint32 Value)",
		"Get_DYTC_CMD_MODE_DISPATCHERFUNCTION": "uint32 Get_DYTC_CMD_MODE_DISPATCHERFUNCTION()",
		"Get_DYTC_CMD_FUNC_CAP":         "uint32 Get_DYTC_CMD_FUNC_CAP()",
		"Get_DYTC_CMD_MODE_NIT_DISPATCHERTHRESHOLD": "uint32 Get_DYTC_CMD_MODE_NIT_DISPATCHERTHRESHOLD()",
		"Set_PPM_State":                  "uint32 Set_PPM_State(uint32 State)",
		"Get_PPM_State":                  "uint32 Get_PPM_State()",
	}
	if sig, ok := sigs[methodName]; ok {
		return sig
	}
	return fmt.Sprintf("uint32 %s()", methodName)
}

func mapWMIToDLLFunc(methodName, className string) string {
	lower := strings.ToLower(className)
	methodLower := strings.ToLower(methodName)

	// Map WMI method to DLL function
	dytcMap := map[string]string{
		"set_dytcmode": "Set_DYTCMode (LenovoDYTC.dll)",
		"set_fanmode":  "Set_FanMode (LenovoDYTC.dll)",
		"set_geekmode": "Set_GEEKMode (LenovoDYTC.dll)",
		"get_cap_dcc":  "GET_Cap_DCC (LenovoDYTC.dll)",
		"get_cap_geek": "GET_Cap_GEEK (LenovoDYTC.dll)",
		"set_odvmode":  "Set_ODVMode (LenovoDYTC.dll)",
		"get_dytc_cmd_mode_dispatcherfunction": "Get_DYTC_CMD_MODE_DISPATCHERFUNCTION (LenovoDYTC.dll)",
		"get_dytc_cmd_func_cap":       "Get_DYTC_CMD_FUNC_CAP (LenovoDYTC.dll)",
		"get_dytc_cmd_mode_nit_dispatcherthreshold": "Get_DYTC_CMD_MODE_NIT_DISPATCHERTHRESHOLD (LenovoDYTC.dll)",
	}

	if strings.Contains(lower, "dytc") || strings.Contains(lower, "fan") {
		if dllFunc, ok := dytcMap[methodLower]; ok {
			return dllFunc
		}
	}

	// WMI brightness methods → directly via WMI namespace, no extra DLL
	if strings.Contains(methodLower, "brightness") && strings.Contains(methodLower, "set") {
		return "Invoke-CimMethod (root/WMI)"
	}
	if strings.Contains(methodLower, "brightness") && strings.Contains(methodLower, "get") {
		return "WmiMonitorBrightness.CurrentBrightness (root/WMI)"
	}

	// Default
	return "Via WMI Provider"
}

func mapWMIToDLL(methodName, className string) string {
	lower := strings.ToLower(className)
	if strings.Contains(lower, "dytc") || strings.Contains(lower, "fan") {
		return "LenovoDYTC.dll"
	}
	return "WMI Provider"
}

func getMethodNotes(methodName, className string) string {
	notes := map[string]string{
		"Set_FanMode":            "Mode: 1=Auto, 2=Boost, 3=Silent",
		"Set_DYTCMode":            "Mode: 1=BSM, 2=IBSM, 3=AQM, 4=STD, 5=APM, 6=IEPM, 7=EPM, 13=DCC",
		"Set_GEEKMode":           "OnOff: 0=Off, 1=On",
		"WmiSetBrightness":       "Requires WmiMonitorBrightnessMethods class",
		"Set_ODVMode":            "Index=0-3, Value=overclocking voltage offset",
		"Get_DYTC_CMD_MODE_DISPATCHERFUNCTION": "Returns 32-bit feature bitmask",
		"GET_Cap_DCC":            "Check DCC capability (1=supported, 0=not)",
		"GET_Cap_GEEK":           "Check GEEK mode capability (1=supported, 0=not)",
	}
	if note, ok := notes[methodName]; ok {
		return note
	}
	return ""
}

func buildWMISummary(classes []WMIClassInfo) WMISummary {
	summary := WMISummary{}

	// High-level features we know about
	features := []FeatureInfo{
		{
			FeatureName: "DYTC Mode Control",
			WMIClass:    "LNV_DYTCInterface",
			WMIMethod:   "Set_DYTCMode",
			DLLFunction: "Set_DYTCMode → LenovoDYTC.dll",
			Description: "Set performance mode (BSM/IBSM/AQM/STD/APM/IEPM/EPM/DCC)",
			Category:    "Power & Thermal",
		},
		{
			FeatureName: "Fan Control",
			WMIClass:    "LNV_FanControl",
			WMIMethod:   "Set_FanMode",
			DLLFunction: "Set_FanMode → LenovoDYTC.dll",
			Description: "Manual fan speed override",
			Category:    "Power & Thermal",
		},
		{
			FeatureName: "GEEK Mode",
			WMIClass:    "LNV_DYTCInterface",
			WMIMethod:   "Set_GEEKMode",
			DLLFunction: "Set_GEEKMode → LenovoDYTC.dll",
			Description: "GEEK mode on/off (turbo boost for specific apps)",
			Category:    "Power & Thermal",
		},
		{
			FeatureName: "ODV Overclocking",
			WMIClass:    "LNV_DYTCInterface",
			WMIMethod:   "Set_ODVMode",
			DLLFunction: "Set_ODVMode → LenovoDYTC.dll",
			Description: "OverDrive voltage adjustment",
			Category:    "Performance",
		},
		{
			FeatureName: "Dispatcher Function Query",
			WMIClass:    "LNV_DYTCInterface",
			WMIMethod:   "Get_DYTC_CMD_MODE_DISPATCHERFUNCTION",
			DLLFunction: "Get_DYTC_CMD_MODE_DISPATCHERFUNCTION → LenovoDYTC.dll",
			Description: "Query 32-bit dispatcher feature flags",
			Category:    "System",
		},
		{
			FeatureName: "DCC Capability Check",
			WMIClass:    "LNV_DYTCInterface",
			WMIMethod:   "GET_Cap_DCC",
			DLLFunction: "GET_Cap_DCC → LenovoDYTC.dll",
			Description: "Check if Dynamic Performance Control is supported",
			Category:    "Power & Thermal",
		},
		{
			FeatureName: "Screen Brightness (WMI)",
			WMIClass:    "WmiMonitorBrightnessMethods",
			WMIMethod:   "WmiSetBrightness",
			DLLFunction: "Invoke-CimMethod (WMI, no extra DLL)",
			Description: "Set screen brightness 0-100%",
			Category:    "Display",
		},
		{
			FeatureName: "Thermal Zone Temp",
			WMIClass:    "MSAcpi_ThermalZoneTemperature",
			WMIMethod:   "(Read Property)",
			DLLFunction: "MSAcpi.sys",
			Description: "Read CPU package temperature",
			Category:    "Monitoring",
		},
	}

	// Mark features as available/unavailable based on what's actually found
	classSet := make(map[string]bool)
	for _, c := range classes {
		classSet[c.Name] = true
	}

	for _, f := range features {
		// If we found the class in our enumeration, mark it as available
		if classSet[f.WMIClass] {
			summary.AvailableFeatures = append(summary.AvailableFeatures, f)
		}
	}

	// If no WMI classes found (e.g. non-Lenovo or no admin), still show known features
	if len(summary.AvailableFeatures) == 0 {
		summary.AvailableFeatures = features
	}

	return summary
}

// InvokeWMI executes a WMI method and returns the result
func InvokeWMI(namespace, className, methodName, params string) string {
	// Validate inputs to prevent injection
	if !isValidIdentifier(namespace) || !isValidIdentifier(className) || !isValidIdentifier(methodName) {
		return fmt.Sprintf(`{"error": "Invalid namespace, class, or method name"}`)
	}

	psScript := fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
try {
    $obj = Get-WmiObject -Namespace '%s' -Class '%s' -ErrorAction Stop | Select-Object -First 1
    $result = $obj.InvokeMethod('%s', $null)
    @{ ReturnValue = $result } | ConvertTo-Json
} catch {
    @{ Error = $_.Exception.Message } | ConvertTo-Json
}
`, namespace, className, methodName)

	out, err := runPowershellScript(psScript)
	if err != nil {
		return fmt.Sprintf(`{"error": "%v"}`, err)
	}
	return strings.TrimSpace(out)
}

var validIdentRe = regexp.MustCompile(`^[a-zA-Z0-9_\\]+$`)

func isValidIdentifier(s string) bool {
	return validIdentRe.MatchString(s) && len(s) < 200
}
