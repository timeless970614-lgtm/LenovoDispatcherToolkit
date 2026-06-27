//go:build windows

package backend

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed scripts/*.ps1
var scriptFS embed.FS

// scriptsDir returns the on-disk directory for embedded scripts.
// Scripts are extracted once to LOCALAPPDATA\Temp\LenovoToolkit\scripts
// and reused for the lifetime of the process.
var scriptsDir = func() string {
	dir := filepath.Join(os.Getenv("LOCALAPPDATA"), "Temp", "LenovoToolkit", "scripts")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "" // extraction failed — fall back to inline or error
	}
	// Extract each script once
	extractScript := func(name string) string {
		data, err := scriptFS.ReadFile("scripts/" + name)
		if err != nil {
			return ""
		}
		p := filepath.Join(dir, name)
		// Only write if missing or content differs (avoids unnecessary writes on restart)
		if existing, err := os.ReadFile(p); err != nil || string(existing) != string(data) {
			_ = os.WriteFile(p, data, 0644)
		}
		return p
	}
	_ = extractScript("get_acdc.ps1")
	_ = extractScript("read_registry_dword.ps1")
	_ = extractScript("collect_system_metrics.ps1")
	return dir
}()

// runPSFile executes a .ps1 script file with optional arguments and returns stdout.
// Returns "", error on failure.
func runPSFile(scriptName string, args ...string) (string, error) {
	dir := scriptsDir
	if dir == "" {
		return "", fmt.Errorf("scripts dir unavailable")
	}
	scriptPath := filepath.Join(dir, scriptName)
	psArgs := []string{
		"-ExecutionPolicy", "Bypass",
		"-NoProfile",
		"-NonInteractive",
		"-File", scriptPath,
	}
	psArgs = append(psArgs, args...)
	cmd := hiddenCmd("powershell.exe", psArgs...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// ── get_acdc.ps1 ─────────────────────────────────────────────────────────

// GetACDC returns 1 if on AC power, 0 if on battery.
func GetACDC() uint32 {
	out, err := runPSFile("get_acdc.ps1")
	if err != nil || out == "" {
		return 1 // default to AC on failure
	}
	var v uint32
	fmt.Sscanf(out, "%d", &v)
	return v
}

// ── read_registry_dword.ps1 ──────────────────────────────────────────────

// ReadRegistryDWord reads a DWORD from the registry.
// Returns (value, found). found=false means the key or value didn't exist.
func ReadRegistryDWord(path, name string) (uint32, bool) {
	out, err := runPSFile("read_registry_dword.ps1", path, name)
	if err != nil {
		return 0, false
	}
	var v uint32
	if _, parseErr := fmt.Sscanf(strings.TrimSpace(out), "%d", &v); parseErr != nil {
		return 0, false
	}
	return v, true
}

// ── collect_system_metrics.ps1 ─────────────────────────────────────────

// SystemMetrics holds the parsed output of collect_system_metrics.ps1 (53 fields).
// Fields match columns 3–55 in the 84-column ML Scenario CSV.
type SystemMetrics struct {
	CPUUsage    float64
	CPUFreq     float64
	CPUPerf     float64
	GPUTotal    float64
	IGPUUsage   float64
	DGPUUsage   float64
	VPUUsage    float64
	NPUUsage    float64
	IGPUID      int
	DGPUID      int
	VPUID       int
	MemUsage    float64
	MemFree     float64
	InputLat    float64
	Lag64       float64
	Lag100      float64
	Lag200      float64
	DiskUsage   float64
	DiskSpeed   float64
	DiskReadLat float64
	DiskWriteLat float64
	ITSMode     int
	BattCap     int
	FgApp       string
	MemSpeed    int
	IGPUVRAM    int
	IGPUShare   int
	IGPUTotal   int
}

// CollectSystemMetrics calls the external script and parses the output.
func CollectSystemMetrics() (SystemMetrics, error) {
	out, err := runPSFile("collect_system_metrics.ps1")
	if err != nil {
		return SystemMetrics{}, err
	}
	return parseSystemMetrics(out)
}

// parseSystemMetrics converts 27 pipe-delimited values (the ones we actually
// populate, not the hardcoded zeros) into a SystemMetrics struct.
// The script outputs 27 active values; missing fields default to 0.
func parseSystemMetrics(line string) (SystemMetrics, error) {
	parts := strings.Split(strings.TrimSpace(line), "|")
	var m SystemMetrics
	var idx int

	readFloat := func() float64 {
		if idx >= len(parts) {
			return 0
		}
		var f float64
		fmt.Sscanf(strings.TrimSpace(parts[idx]), "%f", &f)
		idx++
		return f
	}
	readInt := func() int {
		if idx >= len(parts) {
			return 0
		}
		var v int
		fmt.Sscanf(strings.TrimSpace(parts[idx]), "%d", &v)
		idx++
		return v
	}

	m.CPUUsage = readFloat()
	m.CPUFreq  = readFloat()
	m.CPUPerf  = readFloat()
	m.GPUTotal = readFloat()
	m.IGPUUsage = readFloat()
	m.DGPUUsage = readFloat()
	m.VPUUsage  = readFloat()
	m.NPUUsage  = readFloat()
	m.IGPUID   = readInt()
	m.DGPUID   = readInt()
	m.VPUID    = readInt()
	m.MemUsage = readFloat()
	m.MemFree  = readFloat()
	// 4 empty slots (InputLat, Lag64, Lag100, Lag200) — skipped
	for i := 0; i < 4; i++ {
		readFloat()
	}
	m.DiskUsage    = readFloat()
	m.DiskSpeed    = readFloat()
	m.DiskReadLat  = readFloat()
	m.DiskWriteLat = readFloat()
	// 5 slots (sysPower, cpuPower, gpuPower, nvidiaPower, nvidiaTemp) — skipped (Go injects from IPF DLL)
	for i := 0; i < 5; i++ {
		readFloat()
	}
	// 10 slots (Copy, GDI_Render, Legacy_Overlay, Security, 3D, Video_Dec, Video_Enc, Video_Proc, Unknown, Compute)
	for i := 0; i < 10; i++ {
		readFloat()
	}
	m.ITSMode  = readInt()
	m.BattCap  = readInt()
	m.FgApp    = func() string {
		if idx >= len(parts) {
			return ""
		}
		s := strings.TrimSpace(parts[idx])
		idx++
		return s
	}()
	m.MemSpeed  = readInt()
	m.IGPUVRAM  = readInt()
	m.IGPUShare = readInt()
	m.IGPUTotal = readInt()

	return m, nil
}
