//go:build windows
package main

import (
	"fmt"
	"time"
	"lenovo-toolkit/backend"
)

func main() {
	fmt.Println("=== IPF Raw Value Test ===")
	backend.InitIPF()
	time.Sleep(1 * time.Second)

	for i := 0; i < 3; i++ {
		ipf := backend.ReadIPF()
		fmt.Printf("[%d] Raw: SysPwr=%d mW | PL1=%d PL2=%d PL4=%d | Temp_raw=%d cK | Temp_C=%.1f C | Ver=%d Conn=%v\n",
			i, ipf.SystemPower_mW, ipf.PL1_mW, ipf.PL2_mW, ipf.PL4_mW,
			ipf.CpuTemp_cK, ipf.CpuTemp_C, ipf.Version, ipf.Connected)
		time.Sleep(2 * time.Second)
	}
}
