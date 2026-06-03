//go:build windows

package main

import (
	"encoding/binary"
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

const deviceName = "\\\\.\\LunaPowerMon"

type MSRIO struct {
	Index  uint32
	Value  uint64
	Status int32
	Extra  uint64
}

func main() {
	hDev, err := windows.CreateFile(
		windows.StringToUTF16Ptr(deviceName),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		fmt.Printf("FAIL: OpenDevice: %v\n", err)
		return
	}
	defer windows.CloseHandle(hDev)
	fmt.Println("OK: Device opened")

	// Test RAPL_POWER_UNIT (0x606) - all Intel CPUs support this
	testMSRs := []struct {
		name  string
		index uint32
	}{
		{"RAPL_POWER_UNIT", 0x606},
		{"MSR_PLATFORM_INFO", 0xCE},
		{"IA32_APERF", 0xE8},
		{"IA32_MPERF", 0xE7},
	}

	for _, msr := range testMSRs {
		io := MSRIO{Index: msr.index}
		inBuf := make([]byte, unsafe.Sizeof(io))
		outBuf := make([]byte, unsafe.Sizeof(io))
		binary.LittleEndian.PutUint32(inBuf[0:4], msr.index)

		var returned uint32
		err := windows.DeviceIoControl(
			hDev,
			computeIOCTL(0x8011, 0x800, 0, 0),
			&inBuf[0],
			uint32(len(inBuf)),
			&outBuf[0],
			uint32(len(outBuf)),
			&returned,
			nil,
		)
		if err != nil {
			fmt.Printf("MSR %-20s (0x%X): ERR %v\n", msr.name, msr.index, err)
			continue
		}
		val := binary.LittleEndian.Uint64(outBuf[4:12])
		status := int32(binary.LittleEndian.Uint32(outBuf[12:16]))
		fmt.Printf("MSR %-20s (0x%X): 0x%016X [status=%d]\n", msr.name, msr.index, val, status)
	}
}

func computeIOCTL(devType, function, method, access uint32) uint32 {
	return (devType << 16) | (access << 14) | (function << 2) | method
}