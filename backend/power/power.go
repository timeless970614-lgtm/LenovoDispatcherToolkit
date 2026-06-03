//go:build windows

package power

import (
	"log"
	"path/filepath"
	"time"
)

// ── Public API Types ────────────────────────────────────────────

// PowerReading combines all power sources into a single reading.
type PowerReading struct {
	CPUPackageW  float64 `json:"cpuPackageW"`
	CPUCoresW    float64 `json:"cpuCoresW"`
	CPUUncoreW   float64 `json:"cpuUncoreW"`
	DRAMPowerW   float64 `json:"dramPowerW"`
	GPUPowerW    float64 `json:"gpuPowerW"`
	GPUName      string  `json:"gpuName"`
	GPUTempC     float64 `json:"gpuTempC"`
	SystemTotalW float64 `json:"systemTotalW"`
	CPUSupported bool    `json:"cpuSupported"`
	GPUSupported bool    `json:"gpuSupported"`
	Timestamp    int64   `json:"timestamp"`
}

// InitResult from driver initialization.
type InitResult struct {
	DriverLoaded  bool   `json:"driverLoaded"`
	RAPLAvailable bool   `json:"raplAvailable"`
	NVMLAvailable bool   `json:"nvmlAvailable"`
	CPUType       string `json:"cpuType"`
	Error         string `json:"error,omitempty"`
}

var (
	initialized   bool
	driverSvcPath string
)

// Initialize power monitoring. driverPath = full path to lenovo_power.sys.
func Initialize(driverPath string) InitResult {
	result := InitResult{}
	driverSvcPath = driverPath

	// Step 1: Install and start kernel driver
	if !IsDriverLoaded() {
		err := InstallDriver(driverPath)
		if err != nil {
			result.Error = "driver install: " + err.Error()
			return result
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Step 2: Open device
	err := OpenDevice()
	if err != nil {
		result.Error = "device open: " + err.Error()
		return result
	}
	result.DriverLoaded = true

	// Step 3: Detect CPU type
	rapl.detect()
	result.CPUType = rapl.cpuType
	result.RAPLAvailable = (rapl.cpuType == "intel" || rapl.cpuType == "amd")

	// Step 4: Check NVML
	if err := initNVML(); err == nil {
		result.NVMLAvailable = true
	}

	initialized = true
	log.Printf("[power] Initialized: driver=%t RAPL=%t NVML=%t CPU=%s",
		result.DriverLoaded, result.RAPLAvailable, result.NVMLAvailable, result.CPUType)

	return result
}

// Shutdown cleanly releases all resources.
func Shutdown() {
	initialized = false
	CloseDevice()
	shutdownNVML()
	if driverSvcPath != "" {
		UninstallDriver()
	}
}

// ReadPower reads all available power readings in one call.
func ReadPower() PowerReading {
	reading := PowerReading{
		Timestamp: time.Now().UnixMilli(),
	}

	if !initialized {
		return reading
	}

	// RAPL (CPU power)
	if rapl.cpuType != "unknown" {
		r, err := ReadRAPL()
		if err == nil {
			reading.CPUPackageW = r.CPUPowerWatts
			reading.CPUCoresW = r.CorePowerWatts
			reading.CPUUncoreW = r.GfxPowerWatts
			reading.DRAMPowerW = r.DRAMPowerWatts
			reading.CPUSupported = true
		}
	}

	// GPU power
	gpus := ReadGPUPower()
	if len(gpus) > 0 {
		for _, g := range gpus {
			if g.PowerAvailable && g.PowerWatts > 0 {
				reading.GPUPowerW = g.PowerWatts
				reading.GPUName = g.Name
				reading.GPUTempC = g.TempC
				reading.GPUSupported = true
				break
			}
		}
	}

	// System total (estimated)
	if reading.CPUPackageW > 0 {
		reading.SystemTotalW = EstimateSystemPower(reading.CPUPackageW, gpus)
	}

	return reading
}

// DefaultDriverPath returns the expected driver location next to the EXE.
func DefaultDriverPath() string {
	if driverSvcPath != "" {
		return filepath.Dir(driverSvcPath) + "\\lenovo_power.sys"
	}
	return "lenovo_power.sys"
}

// ensureReady tries to open the kernel device if not yet done.
// Safe to call multiple times — idempotent.
// Does NOT require admin if the driver was already installed/service-started.
func ensureReady() {
	deviceMu.Lock()
	d := hDev
	deviceMu.Unlock()
	if d != 0 {
		return
	}
	_ = OpenDevice()
}
