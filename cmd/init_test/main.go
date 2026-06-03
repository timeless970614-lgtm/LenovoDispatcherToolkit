//go:build windows

package main

import (
	"fmt"
	"time"

	"lenovo-toolkit/backend/power"
)

func main() {
	fmt.Println("=== Power Monitor Init Test ===")

	// Open device directly (skip InstallDriver)
	err := power.OpenDevice()
	if err != nil {
		fmt.Printf("OpenDevice FAIL: %v\n", err)
		return
	}
	fmt.Println("OpenDevice OK")

	// First read seeds baseline
	r1, err := power.ReadRAPL()
	if err != nil {
		fmt.Printf("ReadRAPL #1 FAIL: %v\n", err)
		return
	}
	fmt.Printf("ReadRAPL #1 (baseline): CPU=%.2f Core=%.2f GFX=%.2f DRAM=%.2f\n",
		r1.CPUPowerWatts, r1.CorePowerWatts, r1.GfxPowerWatts, r1.DRAMPowerWatts)
	fmt.Printf("CPUType: %s | SupportsRAPL: %v\n", power.CPUType(), power.SupportsRAPL())

	// Wait and read again
	fmt.Println("Waiting 2s...")
	time.Sleep(2 * time.Second)

	r2, err := power.ReadRAPL()
	if err != nil {
		fmt.Printf("ReadRAPL #2 FAIL: %v\n", err)
		return
	}
	fmt.Printf("ReadRAPL #2 (delta):\n")
	fmt.Printf("  CPU Package: %.2f W\n", r2.CPUPowerWatts)
	fmt.Printf("  CPU Cores:   %.2f W\n", r2.CorePowerWatts)
	fmt.Printf("  GFX/Uncore:  %.2f W\n", r2.GfxPowerWatts)
	fmt.Printf("  DRAM:        %.2f W\n", r2.DRAMPowerWatts)

	power.Shutdown()
	fmt.Println("=== Done ===")
}