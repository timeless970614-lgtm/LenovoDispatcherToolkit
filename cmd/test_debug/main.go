//go:build windows
package main

import (
	"fmt"
	"time"
	"lenovo-toolkit/backend/power"
)

func main() {
	fmt.Println("=== ReadRAPL With Delay Test ===")
	
	r, err := power.ReadRAPL()
	fmt.Printf("ReadRAPL#1 (baseline): err=%v CPU=%.2f CPUType=%s\n", err, r.CPUPowerWatts, power.CPUType())
	
	time.Sleep(2 * time.Second)
	
	r2, err2 := power.ReadRAPL()
	fmt.Printf("ReadRAPL#2 (delta):    err=%v CPU=%.2f Core=%.2f GFX=%.2f DRAM=%.2f\n",
		err2, r2.CPUPowerWatts, r2.CorePowerWatts, r2.GfxPowerWatts, r2.DRAMPowerWatts)
}
