//go:build windows

package power

import (
	"fmt"
	"math"
	"sync"
	"syscall"
	"unsafe"
)

// ── NVML (NVIDIA Management Library) ──────────────────────────────────
// Use syscall.LazyDLL to call nvml.dll functions directly.
// Supported GPUs: NVIDIA Kepler and later.

var (
	nvmlMu       sync.Mutex
	nvmlDLL      *syscall.LazyDLL
	nvmlInitProc   *syscall.LazyProc
	nvmlCountProc  *syscall.LazyProc
	nvmlHandleProc *syscall.LazyProc
	nvmlPowerProc  *syscall.LazyProc
	nvmlShutdownProc *syscall.LazyProc
	nvmlTempProc    *syscall.LazyProc
	nvmlNameProc     *syscall.LazyProc
	nvmlLoaded       bool
)

const nvmlSuccess = 0

func loadNVMLProc(name string) *syscall.LazyProc {
	return nvmlDLL.NewProc(name)
}

func initNVML() error {
	nvmlMu.Lock()
	defer nvmlMu.Unlock()

	if nvmlLoaded {
		return nil
	}

	nvmlDLL = syscall.NewLazyDLL("nvml.dll")
	if nvmlDLL == nil {
		return fmt.Errorf("nvml.dll: LazyDLL init failed")
	}

	// Try to load — this will fail gracefully if DLL not present
	err := nvmlDLL.Load()
	if err != nil {
		return fmt.Errorf("nvml.dll not available: %w", err)
	}

	nvmlInitProc = loadNVMLProc("nvmlInit_v2")
	nvmlCountProc = loadNVMLProc("nvmlDeviceGetCount")
	nvmlHandleProc = loadNVMLProc("nvmlDeviceGetHandleByIndex_v2")
	nvmlPowerProc = loadNVMLProc("nvmlDeviceGetPowerUsage")
	nvmlTempProc = loadNVMLProc("nvmlDeviceGetTemperature")
	nvmlNameProc = loadNVMLProc("nvmlDeviceGetName")
	nvmlShutdownProc = loadNVMLProc("nvmlShutdown")

	// Call nvmlInit
	rc, _, _ := nvmlInitProc.Call()
	if rc != nvmlSuccess {
		return fmt.Errorf("nvmlInit_v2 failed: rc=%d", rc)
	}

	nvmlLoaded = true
	return nil
}

func shutdownNVML() {
	nvmlMu.Lock()
	defer nvmlMu.Unlock()
	if !nvmlLoaded {
		return
	}
	if nvmlShutdownProc != nil {
		nvmlShutdownProc.Call()
	}
	nvmlLoaded = false
}

// NVMLGPUPower holds per-GPU power reading.
type NVMLGPUPower struct {
	Index       uint32
	Name        string
	PowerWatts  float64
	TempC       float64
	Available   bool
}

// ReadNVMLPower reads all NVIDIA GPU power via nvml.dll.
func ReadNVMLPower() []NVMLGPUPower {
	if err := initNVML(); err != nil {
		return nil
	}

	var count uint32
	rc, _, _ := nvmlCountProc.Call(uintptr(unsafe.Pointer(&count)))
	if rc != nvmlSuccess || count == 0 {
		return nil
	}

	const maxDevs = 8
	if count > maxDevs {
		count = maxDevs
	}

	result := make([]NVMLGPUPower, 0, count)
	for i := uint32(0); i < count; i++ {
		var handle uintptr

		rc, _, _ := nvmlHandleProc.Call(uintptr(i), uintptr(unsafe.Pointer(&handle)))
		if rc != nvmlSuccess || handle == 0 {
			continue
		}

		gpu := NVMLGPUPower{
			Index:     i,
			Available: true,
		}

		// Get name
		nameBuf := make([]byte, 64)
		rc, _, _ = nvmlNameProc.Call(handle, uintptr(unsafe.Pointer(&nameBuf[0])), 64)
		if rc == nvmlSuccess {
			gpu.Name = string(nameBuf[:clenString(nameBuf)])
		}

		// Get power (mW → W)
		var powerMw uint32
		rc, _, _ = nvmlPowerProc.Call(handle, uintptr(unsafe.Pointer(&powerMw)))
		if rc == nvmlSuccess {
			gpu.PowerWatts = float64(powerMw) / 1000.0
			if gpu.PowerWatts < 0 || gpu.PowerWatts > 1000 {
				gpu.PowerWatts = 0
				gpu.Available = false
			}
		} else {
			gpu.Available = false
		}

		// Get temperature
		var temp uint32
		rc, _, _ = nvmlTempProc.Call(handle, 0 /*NvmlTemperatureGpu*/, uintptr(unsafe.Pointer(&temp)))
		if rc == nvmlSuccess {
			gpu.TempC = float64(temp)
		}

		result = append(result, gpu)
	}
	return result
}

func clenString(b []byte) int {
	for i, c := range b {
		if c == 0 {
			return i
		}
	}
	return len(b)
}

// ShutdownNVML closes the NVML library cleanly.
func ShutdownNVML() {
	shutdownNVML()
}

// ── Intel GPU Power (RAPL PP1) ─────────────────────────────────────────
// Intel integrated GPU power is available via RAPL PP1 counter.
// Already read as part of RAPL reading; see rapl.go.

// ── GPU Power Reading (public API) ─────────────────────────────────────

// GPUPowerReading holds one GPU's power reading.
type GPUPowerReading struct {
	Name          string  `json:"name"`
	Vendor        string  `json:"vendor"`     // "nvidia", "intel", "amd"
	PowerWatts    float64 `json:"powerWatts"`
	TempC         float64 `json:"tempC"`
	PowerAvailable bool    `json:"powerAvailable"`
	Source        string  `json:"source"`     // "nvml", "rapl", "estimated"
}

// ReadGPUPower collects GPU power from all available sources.
func ReadGPUPower() []GPUPowerReading {
	var gpus []GPUPowerReading

	// NVIDIA via NVML
	nvmlGPUs := ReadNVMLPower()
	for _, nv := range nvmlGPUs {
		gpus = append(gpus, GPUPowerReading{
			Name:           nv.Name,
			Vendor:         "nvidia",
			PowerWatts:     nv.PowerWatts,
			TempC:          nv.TempC,
			PowerAvailable: nv.Available,
			Source:         "nvml",
		})
	}

	// Intel iGPU via RAPL PP1
	if SupportsRAPL() {
		// PP1 (GFX/Uncore) energy is read inside ReadRAPL()
		// We expose it via the RAPL reading, not here separately.
		// This entry is just a marker that iGPU power is available.
		if len(gpus) == 0 {
			gpus = append(gpus, GPUPowerReading{
				Name:           "Intel Graphics (iGPU)",
				Vendor:         "intel",
				PowerAvailable: true,
				Source:         "rapl",
			})
		}
	}

	return gpus
}

// ── System Power Estimation ─────────────────────────────────────────────

// EstimateSystemPower estimates total system power consumption.
// Approximation: CPU package + GPU + fixed overhead (~12W).
// For accurate readings on battery systems, battery discharge rate can be used.
func EstimateSystemPower(cpuWatts float64, gpuReadings []GPUPowerReading) float64 {
	total := cpuWatts

	for _, g := range gpuReadings {
		if g.PowerAvailable && g.PowerWatts > 0 {
			total += g.PowerWatts
		}
	}

	// Fixed overhead estimate (chipset, RAM, drives, fans)
	const fixedOverheadW = 12.0
	overheadScale := 1.0
	if total > 100 {
		overheadScale = 1.5
	} else if total < 20 {
		overheadScale = 0.5
	}
	total += fixedOverheadW * overheadScale

	// Round to 1 decimal
	return math.Round(total*10.0) / 10.0
}
