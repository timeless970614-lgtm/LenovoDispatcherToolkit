//go:build windows

package backend

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ==================== Namespace Enumeration ====================

// WMIScopeResult holds discovered namespaces
type WMIScopeResult struct {
	Namespaces []string `json:"namespaces"`
	Error      string   `json:"error,omitempty"`
}

// EnumerateAllNamespaces recursively discovers all WMI namespaces
func EnumerateAllNamespaces() string {
	result := &WMIScopeResult{}
	nsMap := make(map[string]bool)
	var walk func(root string)
	walk = func(root string) {
		ps := fmt.Sprintf(`$ErrorActionPreference='SilentlyContinue'; try { (New-Object System.Management.ManagementClass("%s","__NAMESPACE",$null)).PSBase.GetInstances() | %% { $n="%s\"+$_.Name; $n } } catch {}`, root, root+"\\")
		out, err := runPowershellScript(ps)
		if err == nil {
			for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
				line = strings.TrimSpace(line)
				if line == "" || nsMap[line] {
					continue
				}
				nsMap[line] = true
				walk(line)
			}
		}
	}
	walk("root")
	for ns := range nsMap {
		result.Namespaces = append(result.Namespaces, ns)
	}
	// Always include the well-known ones even if walk fails
	wellKnown := []string{
		"root\\cimv2", "root\\wmi", "root\\ccm", "root\\default",
		"root\\security", "root\\RSOP", "root\\StandardCimv2",
		"root\\WMI", "root\\lenovo", "root\\LENOVO",
	}
	for _, wk := range wellKnown {
		found := false
		for _, ns := range result.Namespaces {
			if strings.EqualFold(ns, wk) {
				found = true
				break
			}
		}
		if !found {
			result.Namespaces = append(result.Namespaces, wk)
		}
	}
	sortSafe(result.Namespaces)
	data, _ := json.Marshal(result)
	return string(data)
}

func sortSafe(s []string) {
	// Simple bubble-ish sort by length then alpha
	n := len(s)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if strings.ToLower(s[i]) > strings.ToLower(s[j]) {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}

// ==================== Class Browser ====================

// WMIClassBrowse holds class info for the browse tab
type WMIClassBrowse struct {
	Name       string         `json:"name"`
	Namespace  string         `json:"namespace"`
	IsDynamic  bool           `json:"isDynamic"`
	IsStatic   bool           `json:"isStatic"`
	Description string        `json:"description"`
	Properties []WMIBrowseProperty `json:"properties"`
	Methods    []WMIBrowseMethod   `json:"methods"`
	Derivation []string       `json:"derivation"`
	Qualifiers []WMIQualifier `json:"qualifiers"`
}

type WMIBrowseProperty struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	IsArray     bool           `json:"isArray"`
	IsKey       bool           `json:"isKey"`
	Description string         `json:"description"`
	Qualifiers  []WMIQualifier `json:"qualifiers"`
}

type WMIBrowseMethod struct {
	Name       string          `json:"name"`
	ReturnType string          `json:"returnType"`
	Parameters []WMIParameter `json:"parameters"`
	Description string         `json:"description"`
	Qualifiers []WMIQualifier `json:"qualifiers"`
	IsStatic   bool           `json:"isStatic"`
}

type WMIParameter struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	IsIn     bool   `json:"isIn"`
	IsOut    bool   `json:"isOut"`
	IsArray  bool   `json:"isArray"`
	Optional bool   `json:"optional"`
}

type WMIQualifier struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

// GetClassesInNamespace returns all classes in a namespace (no property/method details yet)
func GetClassesInNamespace(namespace string) string {
	if !isValidIdentifier(namespace) {
		return `{"error":"Invalid namespace"}`
	}
	psScript := fmt.Sprintf(`
$ErrorActionPreference='Stop'
$ns="%s"
$result=@()
try {
  $searcher=New-Object System.Management.ManagementObjectSearcher("$ns","select * from meta_class")
  foreach ($c in $searcher.Get()) {
    $name=$c.Name
    $dynamic=$false; $static=$false
    foreach ($q in $c.Qualifiers) { if($q.Name -eq "dynamic"){$dynamic=$true} if($q.Name -eq "static"){$static=$true} }
    $desc=""
    try { $desc=$c.GetPropertyQualifierValue("Description","Value") } catch {}
    $propCount=0; $methCount=0
    try { $propCount=($c.Properties.Count) } catch {}
    try { $methCount=($c.Methods.Count) } catch {}
    $obj=[PSCustomObject]@{
      name=$name; isDynamic=$dynamic; isStatic=$static;
      description=$desc; propertyCount=$propCount; methodCount=$methCount;
      derivation=@($c.Derivation)
    }
    $result+=$obj
  }
  $result | ConvertTo-Json -Depth 3
} catch {
  @{ error=$_.Exception.Message } | ConvertTo-Json
}
`, escapePSScript(namespace))
	out, err := runPowershellScript(psScript)
	if err != nil {
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}
	return strings.TrimSpace(out)
}

// GetClassDetails returns full property + method details for one class
func GetClassDetails(namespace, className string) string {
	if !isValidIdentifier(namespace) || !isValidIdentifier(className) {
		return `{"error":"Invalid namespace or class name"}`
	}
	psScript := fmt.Sprintf(`
$ErrorActionPreference='Stop'
$ns="%s"
$cn="%s"
try {
  $mc=New-Object System.Management.ManagementClass($ns, $cn, $null)
  $mc.Options.UseAmendedQualifiers=$true
  $result=[PSCustomObject]@{
    name=$cn; namespace=$ns; description="";
    isDynamic=$false; isStatic=$false;
    properties=@(); methods=@(); derivation=@();
    qualifiers=@()
  }
  # Qualifiers on class
  foreach ($q in $mc.Qualifiers) {
    $result.qualifiers+=@{name=$q.Name; value=$q.Value}
  }
  # Description from qualifier
  try { $result.description=$mc.GetQualifierValue("Description") } catch {}
  # Dynamic/Static
  foreach ($q in $mc.Qualifiers) { if($q.Name -eq "dynamic"){$result.isDynamic=$true} if($q.Name -eq "static"){$result.isStatic=$true} }

  # Properties
  foreach ($p in $mc.Properties) {
    $pq=@()
    foreach ($q in $p.Qualifiers) { $pq+=@{name=$q.Name;value=$q.Value} }
    $desc=""
    try { $desc=$p.Qualifiers["Description"].Value } catch {}
    $isKey=$false
    foreach ($q in $p.Qualifiers) { if($q.Name -eq "key"){$isKey=$true} }
    $result.properties+=@{name=$p.Name;type=$p.Type.ToString();isArray=$p.IsArray;isKey=$isKey;description=$desc;qualifiers=$pq}
  }

  # Methods
  foreach ($m in $mc.Methods) {
    $mq=@()
    foreach ($q in $m.Qualifiers) { $mq+=@{name=$q.Name;value=$q.Value} }
    $mdesc=""
    try { $mdesc=$m.Qualifiers["Description"].Value } catch {}
    $isStatic=$false
    foreach ($q in $m.Qualifiers) { if($q.Name -eq "static"){$isStatic=$true} }
    $params=@()
    foreach ($p in $m.InParameters.Properties) {
      $params+=@{name=$p.Name;type=$p.Type.ToString();isIn=$true;isOut=$false;isArray=$p.IsArray;optional=$false}
    }
    foreach ($p in $m.OutParameters.Properties) {
      $params+=@{name=$p.Name;type=$p.Type.ToString();isIn=$false;isOut=$true;isArray=$p.IsArray;optional=$false}
    }
    $result.methods+=@{name=$m.Name;returnType=$m.ReturnType.ToString();parameters=$params;description=$mdesc;qualifiers=$mq;isStatic=$isStatic}
  }

  # Derivation
  $result.derivation=@($mc.Derivation)

  $result | ConvertTo-Json -Depth 6
} catch {
  @{ error=$_.Exception.Message } | ConvertTo-Json
}
`, escapePSScript(namespace), escapePSScript(className))
	out, err := runPowershellScript(psScript)
	if err != nil {
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}
	return strings.TrimSpace(out)
}

func escapePSScript(s string) string {
	s = strings.ReplaceAll(s, "`", "``")
	s = strings.ReplaceAll(s, "\"", "`\"")
	s = strings.ReplaceAll(s, "$", "`$")
	return s
}

// ==================== Event Classes ====================

// WMIEventClass holds event class info
type WMIEventClass struct {
	Name       string   `json:"name"`
	Namespace  string   `json:"namespace"`
	Description string  `json:"description"`
	Derivation []string `json:"derivation"`
	Properties []WMIBrowseProperty `json:"properties"`
}

// GetEventClasses returns all classes derived from __Event
func GetEventClasses() string {
	psScript := `
$ErrorActionPreference='SilentlyContinue'
$namespaces=@("root\cimv2","root\wmi","root\ccm","root\default")
$result=@()
foreach($ns in $namespaces) {
  try {
    $searcher=New-Object System.Management.ManagementObjectSearcher("$ns","select * from meta_class where __this ISA '__Event'")
    foreach($c in $searcher.Get()) {
      $props=@()
      foreach($p in $c.Properties){ $props+=@{name=$p.Name;type=$p.Type.ToString();isArray=$p.IsArray;isKey=$false;description="";qualifiers=@()} }
      $result+=@{name=$c.Name;namespace=$ns;derivation=@($c.Derivation);properties=$props;description=""}
    }
  } catch {}
}
$result | ConvertTo-Json -Depth 4
`
	out, err := runPowershellScript(psScript)
	if err != nil {
		return fmt.Sprintf(`{"error":"%v"}`, err)
	}
	return strings.TrimSpace(out)
}

// ==================== Code Generation ====================

// CodeGenRequest holds parameters for code generation
type CodeGenRequest struct {
	Namespace  string   `json:"namespace"`
	ClassName  string   `json:"className"`
	Properties []string `json:"properties"`
	Where      string   `json:"where"`
	Language   string   `json:"language"` // "powershell", "vbscript", "csharp", "vbnet"
	Target     string   `json:"target"`   // "local", "remote", "group"
	RemoteHost string   `json:"remoteHost"`
}

// GenerateWMIQueryCode generates WMI query code in the requested language
func GenerateWMIQueryCode(reqJSON string) string {
	var req CodeGenRequest
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return fmt.Sprintf(`{"error":"Invalid request: %v"}`, err)
	}
	switch strings.ToLower(req.Language) {
	case "powershell":
		return generatePowerShellQuery(req)
	case "vbscript":
		return generateVBScriptQuery(req)
	case "csharp":
		return generateCSharpQuery(req)
	case "vbnet":
		return generateVBNetQuery(req)
	default:
		return `{"error":"Unsupported language. Use: powershell, vbscript, csharp, vbnet"}`
	}
}

func generatePowerShellQuery(req CodeGenRequest) string {
	ns := req.Namespace
	cn := req.ClassName
	props := req.Properties
	where := req.Where

	code := "$ErrorActionPreference = 'Stop'\n"
	code += "# ===== WMI Query Generated by Lenovo Dispatcher Toolkit =====\n"
	code += fmt.Sprintf("# Namespace: %s | Class: %s\n\n", ns, cn)

	if req.Target == "remote" && req.RemoteHost != "" {
		code += fmt.Sprintf("$cred = Get-Credential -Message \"WMI access for %s\"\n", req.RemoteHost)
		code += fmt.Sprintf("$session = New-CimSession -ComputerName \"%s\" -Credential $cred\n", req.RemoteHost)
		code += fmt.Sprintf("$classes = Get-CimInstance -CimSession $session -Namespace \"%s\" -ClassName \"%s\"", ns, cn)
	} else {
		code += fmt.Sprintf("$classes = Get-CimInstance -Namespace \"%s\" -ClassName \"%s\"", ns, cn)
	}
	if where != "" {
		code += fmt.Sprintf(" | Where-Object { %s }", where)
	}
	code += "\n\n"

	if len(props) > 0 {
		code += "foreach ($item in $classes) {\n"
		for _, p := range props {
			code += fmt.Sprintf("    Write-Host \"%s: $($item.%s)\"\n", p, p)
		}
		code += "}\n"
	} else {
		code += "foreach ($item in $classes) {\n"
		code += "    $item | Format-List\n"
		code += "}\n"
	}

	if req.Target == "remote" && req.RemoteHost != "" {
		code += "\nRemove-CimSession $session\n"
	}

	return wrapCodeResult(code)
}

func generateVBScriptQuery(req CodeGenRequest) string {
	ns := req.Namespace
	cn := req.ClassName
	props := req.Properties
	where := req.Where

	code := "strComputer = \".\"\n"
	code += "Set objWMIService = GetObject(\"winmgmts:\\\\\" & strComputer & \"\\\\" + ns + "\")\n"
	code += "Set colItems = objWMIService.ExecQuery(\"SELECT * FROM " + cn
	if where != "" {
		code += " WHERE " + where
	}
	code += "\",,48)\n\n"
	code += "For Each objItem in colItems\n"
	if len(props) > 0 {
		for _, p := range props {
			code += "    Wscript.Echo \"" + p + ": \" & objItem." + p + "\n"
		}
	} else {
		code += "    Wscript.Echo \"Instance found\"\n"
	}
	code += "Next\n"

	return wrapCodeResult(code)
}

func generateCSharpQuery(req CodeGenRequest) string {
	ns := req.Namespace
	cn := req.ClassName
	props := req.Properties

	code := "using System;\n"
	code += "using System.Management;\n\n"
	code += "namespace WMIQuery\n{\n"
	code += "    class Program\n"
	code += "    {\n"
	code += "        static void Main(string[] args)\n"
	code += "        {\n"
	code += "            try\n"
	code += "            {\n"
	code += "                ManagementObjectSearcher searcher = new ManagementObjectSearcher(\n"
	code += "                    @\"" + ns + "\",\n"
	code += "                    \"SELECT * FROM " + cn + "\");\n\n"
	code += "                foreach (ManagementObject obj in searcher.Get())\n"
	code += "                {\n"
	if len(props) > 0 {
		for _, p := range props {
			code += "                    Console.WriteLine(\"" + p + ": \" + obj[\"" + p + "\"]);\n"
		}
	} else {
		code += "                    foreach (PropertyData p in obj.Properties)\n"
		code += "                        Console.WriteLine(p.Name + \": \" + (p.Value ?? \"\"));\n"
	}
	code += "                    Console.WriteLine(\"---\");\n"
	code += "                }\n"
	code += "            }\n"
	code += "            catch (Exception ex) { Console.WriteLine(\"Error: \" + ex.Message); }\n"
	code += "        }\n"
	code += "    }\n"
	code += "}\n"

	return wrapCodeResult(code)
}

func generateVBNetQuery(req CodeGenRequest) string {
	ns := req.Namespace
	cn := req.ClassName
	props := req.Properties

	code := "Imports System\n"
	code += "Imports System.Management\n\n"
	code += "Module Program\n"
	code += "    Sub Main()\n"
	code += "        Try\n"
	code += "            Dim searcher As New ManagementObjectSearcher(\n"
	code += "                \"" + ns + "\",\n"
	code += "                \"SELECT * FROM " + cn + "\")\n\n"
	code += "            For Each obj As ManagementObject In searcher.Get()\n"
	if len(props) > 0 {
		for _, p := range props {
			code += "                Console.WriteLine(\"" + p + ": \" & obj(\"" + p + "\"))\n"
		}
	} else {
		code += "                For Each p As PropertyData In obj.Properties\n"
		code += "                    Console.WriteLine(p.Name & \": \" & (p.Value ?? \"\"))\n"
		code += "                Next\n"
	}
	code += "                Console.WriteLine(\"---\")\n"
	code += "            Next\n"
	code += "        Catch ex As Exception\n"
	code += "            Console.WriteLine(\"Error: \" & ex.Message)\n"
	code += "        End Try\n"
	code += "    End Sub\n"
	code += "End Module\n"

	return wrapCodeResult(code)
}

func wrapCodeResult(code string) string {
	type codeResult struct {
		Code  string `json:"code"`
		Error string `json:"error,omitempty"`
	}
	r := codeResult{Code: code}
	data, _ := json.Marshal(r)
	return string(data)
}

// GenerateWMIMethodCode generates WMI method invocation code
func GenerateWMIMethodCode(reqJSON string) string {
	var req struct {
		Namespace string        `json:"namespace"`
		ClassName string        `json:"className"`
		Method    string        `json:"method"`
		InParams  []WMIParameter `json:"inParams"`
		Language  string        `json:"language"`
	}
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return fmt.Sprintf(`{"error":"Invalid request: %v"}`, err)
	}
	ns := req.Namespace
	cn := req.ClassName
	mn := req.Method

	switch strings.ToLower(req.Language) {
	case "powershell":
		code := "# WMI Method Invocation - PowerShell\n\n"
		code += fmt.Sprintf("$class = Get-CimClass -Namespace \"%s\" -ClassName \"%s\"\n", ns, cn)
		code += fmt.Sprintf("$instance = Get-CimInstance -Namespace \"%s\" -ClassName \"%s\" | Select-Object -First 1\n", ns, cn)
		code += fmt.Sprintf("$result = Invoke-CimMethod -InputObject $instance -MethodName \"%s\"", mn)
		if len(req.InParams) > 0 {
			code += " -Arguments @{\n"
			for _, p := range req.InParams {
				code += fmt.Sprintf("    %s = <%s value> # type: %s\n", p.Name, p.Name, p.Type)
			}
			code += "}"
		}
		code += "\n$result | Format-List\n"
		return wrapCodeResult(code)
	case "vbscript":
		code := "strComputer = \".\"\n"
		code += "Set objWMIService = GetObject(\"winmgmts:\\\\\" & strComputer & \"\\\\" + ns + "\")\n"
		code += fmt.Sprintf("Set objInstance = objWMIService.Get(\"%s.%s\")\n", cn, cn)
		code += fmt.Sprintf("ret = objInstance.%s(", mn)
		if len(req.InParams) > 0 {
			names := ""
			for _, p := range req.InParams {
				if names != "" { names += ", " }
				names += p.Name
			}
			code += names
		}
		code += ")\nWscript.Echo \"Return: \" & ret\n"
		return wrapCodeResult(code)
	default:
		return `{"error":"Unsupported language for method code. Use: powershell, vbscript"}`
	}
}

// GenerateWMIEventCode generates WMI event subscription code
func GenerateWMIEventCode(reqJSON string) string {
	var req struct {
		EventClass string `json:"eventClass"`
		Namespace  string `json:"namespace"`
		Condition string `json:"condition"`
		Language  string `json:"language"`
	}
	if err := json.Unmarshal([]byte(reqJSON), &req); err != nil {
		return fmt.Sprintf(`{"error":"Invalid request: %v"}`, err)
	}
	switch strings.ToLower(req.Language) {
	case "powershell":
		code := "# WMI Event Subscription - PowerShell\n\n"
		code += "Register-CimIndicationEvent -Namespace \"root\\cimv2\" `\n"
		code += fmt.Sprintf("    -ClassName \"%s\" `\n", req.EventClass)
		code += "    -SourceIdentifier \"WMIMonitor\" `\n"
		code += "    -Action {\n"
		code += "        Write-Host \"WMI Event received: $Event.SourceEventArgs.NewEvent | Out-String\"\n"
		code += "    }\n\n"
		code += "# To unsubscribe: Unregister-Event -SourceIdentifier \"WMIMonitor\"\n"
		return wrapCodeResult(code)
	case "vbscript":
		code := "Set objWMIService = GetObject(\"winmgmts:\\\\.\\root\\cimv2\")\n"
		code += "Set colEvents = objWMIService.ExecNotificationQuery(\"SELECT * FROM " + req.EventClass + "\")\n"
		code += "Do\n"
		code += "    Set objEvent = colEvents.NextEvent\n"
		code += "    Wscript.Echo \"Event received\"\n"
		code += "Loop\n"
		return wrapCodeResult(code)
	default:
		return `{"error":"Unsupported language for event code. Use: powershell, vbscript"}`
	}
}
