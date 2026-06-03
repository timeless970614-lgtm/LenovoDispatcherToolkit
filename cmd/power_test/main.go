//go:build windows

package main

import (
	"fmt"
	"os"

	"lenovo-toolkit/backend/power"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] != "test" {
		return
	}

	// Try to open device (lazy init)
	fmt.Println("[TEST] Opening kernel device...")

	// Test RAPL detection
	fmt.Printf("[TEST] RAPL available: %v\n", power.SupportsRAPL())
	fmt.Printf("[TEST] CPU type: %s\n", power.CPUType())

	// Test ReadRAPL
	reading, err := power.ReadRAPL()
	if err != nil {
		fmt.Printf("[TEST] ReadRAPL error: %v\n", err)
	} else {
		fmt.Printf("[TEST] CPU Power: %.2f W\n", reading.CPUPowerWatts)
		fmt.Printf("[TEST] Core Power: %.2f W\n", reading.CorePowerWatts)
		fmt.Printf("[TEST] GFX/Uncore: %.2f W\n", reading.GfxPowerWatts)
		fmt.Printf("[TEST] DRAM Power: %.2f W\n", reading.DRAMPowerWatts)
	}

	// For GPU, call ReadRAPL a second time to get delta
	reading2, err := power.ReadRAPL()
	if err != nil {
		fmt.Printf("[TEST] ReadRAPL #2 error: %v\n", err)
	} else {
		fmt.Printf("[TEST] Read #2 - CPU Power: %.2f W\n", reading2.CPUPowerWatts)
		fmt.Printf("[TEST] Read #2 - Core Power: %.2f W\n", reading2.CorePowerWatts)
		fmt.Printf("[TEST] Read #2 - GFX/Uncore: %.2f W\n", reading2.GfxPowerWatts)
		fmt.Printf("[TEST] Read #2 - DRAM Power: %.2f W\n", reading2.DRAMPowerWatts)
	}

	fmt.Println("[TEST] Done!")
	power.Shutdown()
}