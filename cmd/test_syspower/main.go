//go:build windows
package main

import (
	"fmt"
	"time"
	"lenovo-toolkit/backend"
)

func main() {
	fmt.Println("=== Frontend Flow Simulation ===")
	
	// Simulate: first call = GetSystemPowerInfo() on mount
	info1 := backend.GetSystemPowerInfo()
	fmt.Printf("Call#1 GetSystemPowerInfo: CPU=%.2fW GPU=%.2fW Sys=%.2fW Temp=%.1fC Freq=%.0fMHz\n",
		info1.CPUPowerWatts, info1.GPUPowerWatts, info1.SysPowerWatts, info1.CPUTempC, info1.CPUFreqMHz)
	
	time.Sleep(3 * time.Second)
	
	// Simulate: second call = UpdateCachedSystemPower() via polling
	info2 := backend.UpdateCachedSystemPower()
	fmt.Printf("Call#2 UpdateCached:       CPU=%.2fW GPU=%.2fW Sys=%.2fW Temp=%.1fC Freq=%.0fMHz\n",
		info2.CPUPowerWatts, info2.GPUPowerWatts, info2.SysPowerWatts, info2.CPUTempC, info2.CPUFreqMHz)
	
	fmt.Printf("PL1=%.2fW PL2=%.2fW PL4=%.2fW CPUUtil=%.1f%% NPU=%.2fW\n",
		info2.PL1Watts, info2.PL2Watts, info2.PL4Watts, info2.CPUUtilPct, info2.NPUPowerWatts)
}
