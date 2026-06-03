//go:build windows

package power

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// ── RAPL MSR Registers (Intel) ─────────────────────────────────────
const (
	msrRAPLUnit       = 0x606 // RAPL_POWER_UNIT
	msrPkgEnergy      = 0x611 // PKG_ENERGY_STATUS
	msrDramEnergy     = 0x619 // DRAM_ENERGY_STATUS
	msrPp0Energy      = 0x639 // PP0_ENERGY_STATUS (cores)
	msrPp1Energy      = 0x641 // PP1_ENERGY_STATUS (gfx/uncore)
)

// ── RAPL MSR Registers (AMD) ─────────────────────────────────────
const (
	msrAmdPkgEnergy   = 0xC001029B // AMD Family 17h+ Package Energy
)

// ── RAPL Power Reading ──────────────────────────────────────────

// RAPLReading represents a single RAPL power reading.
type RAPLReading struct {
	CPUPowerWatts  float64 `json:"cpuPowerWatts"`
	DRAMPowerWatts float64 `json:"dramPowerWatts"`
	CorePowerWatts float64 `json:"corePowerWatts"`
	GfxPowerWatts  float64 `json:"gfxPowerWatts"`
	Timestamp       int64   `json:"timestamp"`
}

// ── RAPL State ──────────────────────────────────────────────────

type raplState struct {
	mu         sync.RWMutex
	lastTime   int64
	lastPkg    uint64
	lastDram   uint64
	lastPp0    uint64
	lastPp1    uint64

	energyUnits float64 // joules per counter tick
	dramUnits   float64
	pp0Units    float64
	pp1Units    float64

	cpuType string // "intel", "amd", "unknown"
}

var rapl = &raplState{cpuType: "unknown"}

// detectCPUType reads a known MSR to distinguish Intel vs AMD.
func (r *raplState) detect() {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cpuType != "unknown" {
		return
	}

	// Ensure kernel driver is open (idempotent)
	ensureReady()

	// Try Intel RAPL unit register
	_, err := ReadMSR(msrRAPLUnit)
	if err == nil {
		r.cpuType = "intel"
		return
	}

	// Try AMD package energy
	_, err = ReadMSR(msrAmdPkgEnergy)
	if err == nil {
		r.cpuType = "amd"
		return
	}

	r.cpuType = "unknown"
}

// readEnergyUnits reads the energy unit multiplier from RAPL_POWER_UNIT.
// Caller must hold r.mu.
func (r *raplState) readEnergyUnits() error {
	if r.cpuType != "intel" {
		return fmt.Errorf("CPU type %s: RAPL unit register not available", r.cpuType)
	}

	val, err := ReadMSR(msrRAPLUnit)
	if err != nil {
		return err
	}

	// Bits 8-12: energy status units = 1/(2^N) joules
	eu := (val >> 8) & 0x1F
	r.energyUnits = 1.0 / math.Pow(2.0, float64(eu))

	// Bits 16-19: DRAM energy units
	du := (val >> 16) & 0x1F
	r.dramUnits = 1.0 / math.Pow(2.0, float64(du))

	// Bits 24-27: PP0 (core) units — same as package
	r.pp0Units = r.energyUnits

	// Bits 32-35: PP1 (uncore/gfx) units
	pu := (val >> 32) & 0x1F
	if pu == 0 {
		r.pp1Units = r.energyUnits
	} else {
		r.pp1Units = 1.0 / math.Pow(2.0, float64(pu))
	}

	return nil
}

// readCurrentEnergy reads raw energy counters from MSRs.
// Caller must hold r.mu.
func (r *raplState) readCurrentEnergy() (pkg, dram, pp0, pp1 uint64, err error) {

	if r.cpuType == "intel" {
		pkg, err = ReadMSR(msrPkgEnergy)
		if err != nil {
			return
		}
		pkg &= 0xFFFFFFFF

		dram, _ = ReadMSR(msrDramEnergy)
		dram &= 0xFFFFFFFF

		pp0, _ = ReadMSR(msrPp0Energy)
		pp0 &= 0xFFFFFFFF

		pp1, _ = ReadMSR(msrPp1Energy)
		pp1 &= 0xFFFFFFFF

		return
	}

	if r.cpuType == "amd" {
		pkg, err = ReadMSR(msrAmdPkgEnergy)
		if err != nil {
			return
		}
		return
	}

	err = fmt.Errorf("CPU type %s: RAPL not available", r.cpuType)
	return
}

// deltaCounter handles 32-bit counter overflow.
func deltaCounter(prev, curr uint64) uint64 {
	if curr >= prev {
		return curr - prev
	}
	// Overflow: counter wrapped
	return (0xFFFFFFFF - prev) + curr + 1
}

// ReadRAPL reads current CPU power via RAPL.
// Returns a reading; CPUPowerWatts==0 if not enough data yet.
func ReadRAPL() (RAPLReading, error) {
	nowMs := time.Now().UnixMilli()

	// Detect CPU type outside the lock (avoids deadlock with detect())
	if rapl.cpuType == "unknown" {
		rapl.detect()
	}

	rapl.mu.Lock()
	defer rapl.mu.Unlock()

	// Read energy units if not yet done
	if rapl.energyUnits == 0 {
		if err := rapl.readEnergyUnits(); err != nil {
			return RAPLReading{}, fmt.Errorf("readEnergyUnits: %w", err)
		}
	}

	// Read raw counters
	pkg, dram, pp0, pp1, err := rapl.readCurrentEnergy()
	if err != nil {
		return RAPLReading{}, fmt.Errorf("readCurrentEnergy: %w", err)
	}

	reading := RAPLReading{
		Timestamp: nowMs,
	}

	// Calculate delta (need previous values for power calculation)
	if rapl.lastTime > 0 && rapl.lastPkg != 0 {
		dt := float64(nowMs-rapl.lastTime) / 1000.0
		if dt <= 0 {
			dt = 0.001
		}

		deltaPkg := deltaCounter(rapl.lastPkg, pkg)
		reading.CPUPowerWatts = (float64(deltaPkg) * rapl.energyUnits) / dt

		deltaDram := deltaCounter(rapl.lastDram, dram)
		reading.DRAMPowerWatts = (float64(deltaDram) * rapl.dramUnits) / dt

		deltaPp0 := deltaCounter(rapl.lastPp0, pp0)
		reading.CorePowerWatts = (float64(deltaPp0) * rapl.pp0Units) / dt

		deltaPp1 := deltaCounter(rapl.lastPp1, pp1)
		reading.GfxPowerWatts = (float64(deltaPp1) * rapl.pp1Units) / dt

		// Sanity clamp
		const maxCPU = 500.0
		const maxGFX = 200.0
		if reading.CPUPowerWatts < 0 || reading.CPUPowerWatts > maxCPU {
			reading.CPUPowerWatts = 0
		}
		if reading.DRAMPowerWatts < 0 || reading.DRAMPowerWatts > 100 {
			reading.DRAMPowerWatts = 0
		}
		if reading.CorePowerWatts < 0 || reading.CorePowerWatts > maxCPU {
			reading.CorePowerWatts = 0
		}
		if reading.GfxPowerWatts < 0 || reading.GfxPowerWatts > maxGFX {
			reading.GfxPowerWatts = 0
		}
	}

	// Update state for next reading
	rapl.lastTime = nowMs
	rapl.lastPkg = pkg
	rapl.lastDram = dram
	rapl.lastPp0 = pp0
	rapl.lastPp1 = pp1

	return reading, nil
}

// CPUType returns the detected CPU type.
func CPUType() string {
	rapl.mu.RLock()
	defer rapl.mu.RUnlock()
	return rapl.cpuType
}

// SupportsRAPL returns true if RAPL is available.
func SupportsRAPL() bool {
	rapl.mu.RLock()
	defer rapl.mu.RUnlock()
	return rapl.cpuType == "intel" || rapl.cpuType == "amd"
}
