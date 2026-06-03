//go:build windows

package power

import (
	"encoding/binary"
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// ── Device Constants ──────────────────────────────────────────────────────

const (
	deviceName    = "\\\\.\\LunaPowerMon"
	serviceName   = "LunaPowerMon"
	driverPathKey = "LunaPowerMonDriver" // registry key for driver path storage

	// IOCTL codes (must match driver.c)
	ioctlReadMSR     = 0x80112000 + 0x800*4 // CTL_CODE(0x8011, 0x800, METHOD_BUFFERED, FILE_ANY_ACCESS)
	ioctlWriteMSR    = 0x80112004 + 0x800*4
	ioctlReadPCICfg  = 0x80112008 + 0x800*4
	ioctlReadPhysMem = 0x8011200C + 0x800*4
)

var (
	deviceMu sync.Mutex
	hDev     windows.Handle
)

// ── IOCTL helpers ──────────────────────────────────────────────────────────

func computeIOCTL(devType, function, method, access uint32) uint32 {
	return (devType << 16) | (access << 14) | (function << 2) | method
}

func ioctlCode(function uint32) uint32 {
	return computeIOCTL(0x8011, function, 0, 0) // METHOD_BUFFERED=0, FILE_ANY_ACCESS=0
}

func callIOCTL(code uint32, inBuf []byte, outBuf []byte) error {
	deviceMu.Lock()
	defer deviceMu.Unlock()

	if hDev == 0 {
		return fmt.Errorf("driver device not open")
	}

	var inPtr *byte
	if len(inBuf) > 0 {
		inPtr = &inBuf[0]
	}

	var outPtr *byte
	if len(outBuf) > 0 {
		outPtr = &outBuf[0]
	}

	var returned uint32
	err := windows.DeviceIoControl(
		hDev,
		code,
		inPtr,
		uint32(len(inBuf)),
		outPtr,
		uint32(len(outBuf)),
		&returned,
		nil,
	)
	return err
}

// ── MSR Read ──────────────────────────────────────────────────────────────

// ReadMSR reads a Model-Specific Register value via kernel driver.
// Returns 0 on error (check Error field).
func ReadMSR(msrIndex uint32) (uint64, error) {
	io := MSRIO{Index: msrIndex}

	inBuf := make([]byte, unsafe.Sizeof(io))
	outBuf := make([]byte, unsafe.Sizeof(io))

	binary.LittleEndian.PutUint32(inBuf[0:4], msrIndex)

	if err := callIOCTL(ioctlCode(0x800), inBuf, outBuf); err != nil {
		return 0, fmt.Errorf("ReadMSR(0x%X): %w", msrIndex, err)
	}

	io.Value = binary.LittleEndian.Uint64(outBuf[4:12])
	io.Status = int32(binary.LittleEndian.Uint32(outBuf[12:16]))

	if io.Status != 0 {
		return 0, fmt.Errorf("ReadMSR(0x%X): GPF/UD (status=%d)", msrIndex, io.Status)
	}

	return io.Value, nil
}

// ── Driver Lifecycle ──────────────────────────────────────────────────────

// InstallDriver creates the kernel service and loads it.
// driverPath must be a full path to lenovo_power.sys.
func InstallDriver(driverPath string) error {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("OpenSCManager: %w", err)
	}
	defer windows.CloseServiceHandle(scm)

	pathPtr, err := windows.UTF16PtrFromString(driverPath)
	if err != nil {
		return err
	}
	namePtr, _ := windows.UTF16PtrFromString(serviceName)

	// Try to create the service
	svc, err := windows.CreateService(
		scm,
		namePtr,
		namePtr,
		windows.SERVICE_ALL_ACCESS,
		windows.SERVICE_KERNEL_DRIVER,
		windows.SERVICE_DEMAND_START,
		windows.SERVICE_ERROR_NORMAL,
		pathPtr,
		nil, nil, nil, nil, nil)
	if err != nil {
		// Service may already exist - try to open
		svc, err = windows.OpenService(scm, namePtr, windows.SERVICE_ALL_ACCESS)
		if err != nil {
			return fmt.Errorf("CreateService/OpenService: %w", err)
		}
	}
	defer windows.CloseServiceHandle(svc)

	// Start the driver
	err = windows.StartService(svc, 0, nil)
	if err != nil && err != windows.ERROR_SERVICE_ALREADY_RUNNING {
		return fmt.Errorf("StartService: %w", err)
	}

	return nil
}

// UninstallDriver stops and removes the kernel service.
func UninstallDriver() error {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("OpenSCManager: %w", err)
	}
	defer windows.CloseServiceHandle(scm)

	namePtr, _ := windows.UTF16PtrFromString(serviceName)
	svc, err := windows.OpenService(scm, namePtr, windows.SERVICE_ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("OpenService: %w", err)
	}
	defer windows.CloseServiceHandle(svc)

	// Stop the service
	var status windows.SERVICE_STATUS
	windows.ControlService(svc, windows.SERVICE_CONTROL_STOP, &status)

	// Delete the service
	return windows.DeleteService(svc)
}

// OpenDevice opens the kernel driver device for IOCTL access.
func OpenDevice() error {
	devPtr, err := syscall.UTF16PtrFromString(deviceName)
	if err != nil {
		return err
	}

	hDev, err = windows.CreateFile(
		devPtr,
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		0, // exclusive
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if err != nil {
		return fmt.Errorf("CreateFile(%s): %w", deviceName, err)
	}
	return nil
}

// CloseDevice closes the kernel driver device handle.
func CloseDevice() error {
	deviceMu.Lock()
	defer deviceMu.Unlock()
	if hDev != 0 {
		windows.CloseHandle(hDev)
		hDev = 0
	}
	return nil
}

// IsDriverLoaded checks if the driver service is running.
func IsDriverLoaded() bool {
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(scm)

	namePtr, _ := windows.UTF16PtrFromString(serviceName)
	svc, err := windows.OpenService(scm, namePtr, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(svc)

	var status windows.SERVICE_STATUS
	err = windows.QueryServiceStatus(svc, &status)
	if err != nil {
		return false
	}
	return status.CurrentState == windows.SERVICE_RUNNING
}