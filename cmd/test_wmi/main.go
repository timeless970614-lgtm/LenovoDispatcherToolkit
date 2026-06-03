//go:build windows
package main

import (
	"fmt"
	"lenovo-toolkit/backend"
	"time"
)

func main() {
	fmt.Println("=== WMI Explorer Test ===")
	start := time.Now()
	result := backend.GetWMIExplorer()
	elapsed := time.Since(start)
	
	fmt.Printf("Time: %.2f seconds\n", elapsed.Seconds())
	fmt.Printf("Length: %d chars\n", len(result))
	
	if len(result) > 500 {
		fmt.Println(result[:500])
	} else if len(result) > 0 {
		fmt.Println(result)
	} else {
		fmt.Println("EMPTY RESULT!")
	}
}