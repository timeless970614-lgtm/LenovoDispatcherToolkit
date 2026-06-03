//go:build windows
package main

import (
	"fmt"
	"time"
	"lenovo-toolkit/backend"
)

func main() {
	fmt.Println("=== IPF Raw Debug ===")
	backend.InitIPF()
	
	time.Sleep(1 * time.Second)
	
	for i := 0; i < 3; i++ {
		info := backend.GetSystemPowerInfo()
		fmt.Printf("[%d] IPF=%dmW Sys=%.2fW Temp=%.1fC PL1=%dW PL2=%dW\n",
			i, info.IPFPowerMW, info.SysPowerWatts, info.CPUTempC, 
			int(info.PL1Watts), int(info.PL2Watts))
		time.Sleep(2 * time.Second)
	}
}
