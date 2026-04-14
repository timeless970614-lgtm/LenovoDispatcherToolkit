// +build ignore

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"unsafe"
)

type hmDeviceInfo struct {
	NumDevices uint32
	_          [4]byte
	DeviceIDs  [32]uint32
}

var npuDLL *syscall.LazyDLL
var npuFunc struct {
	getDeviceInfo uintptr
}

func getExeDir() string {
	exe, _ := os.Executable()
	return filepath.Dir(exe)
}

func npuDynFindDLL() (string, error) {
	exeDir := getExeDir()
	candidates := []string{
		filepath.Join(exeDir, "libhal_xh2a.dll"),
		filepath.Join(exeDir, "build", "bin", "libhal_xh2a.dll"),
		filepath.Join(exeDir, "..", "..", "backend", "dynamic_npu", "libhal_xh2a.dll"),
		`C:\Users\3-64\source\repos\Project1\Project1\libhal_xh2a.dll`,
		`C:\Program Files (x86)\houmo-drv-xh2_v1.1.0\hal\lib\libhal_xh2a.dll`,
		`C:\LenovoDispatcherToolkit\build\bin\libhal_xh2a.dll`,
		`C:\LenovoDispatcherToolkit\backend\dynamic_npu\libhal_xh2a.dll`,
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("not found")
}

func main() {
	fmt.Println("=== NPU DLL Test (Standalone Go) ===")

	dllPath, err := npuDynFindDLL()
	if err != nil {
		fmt.Printf("DLL not found: %v\n", err)
		return
	}
	fmt.Printf("DLL found at: %s\n", dllPath)

	// Load DLL
	dll, err := syscall.LoadDLL(dllPath)
	if err != nil {
		fmt.Printf("LoadLibrary failed: %v\n", err)
		return
	}
	defer dll.Release()

	// Get function
	proc, err := dll.FindProc("hm_sys_get_device_info")
	if err != nil {
		fmt.Printf("GetProcAddress failed: %v\n", err)
		return
	}
	fmt.Printf("Function address: %x\n", proc.Addr())

	// Call the function
	var info hmDeviceInfo
	ret, _, err := syscall.Syscall(proc.Addr(), 1, uintptr(unsafe.Pointer(&info)), 0, 0)
	fmt.Printf("\nResult:\n")
	fmt.Printf("  Return value (RAX): %d\n", ret)
	fmt.Printf("  info.NumDevices:     %d\n", info.NumDevices)
	fmt.Printf("  info.DeviceIDs:     %v\n", info.DeviceIDs[:info.NumDevices])
	if info.NumDevices == 0 && ret != 0 {
		fmt.Printf("\n  NOTE: NumDevices=0 but ret=%d — using ret as device count\n", ret)
	}
}
