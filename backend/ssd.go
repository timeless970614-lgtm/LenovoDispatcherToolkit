//go:build windows

package backend

import (
	"fmt"
	"strconv"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// NVMe constants
const (
	VUC_OPC                   = 0xC6
	VUC_MAGIC                 = 0x4657
	SUBCMD_SET_MODE           = 0x0E
	SUBCMD_GET_MODE           = 0x0F
	IOCTL_STORAGE_QUERY_PROPERTY   = 0x002D1400
	IOCTL_STORAGE_PROTOCOL_COMMAND = 0x002D1480
	NVME_ADMIN_COMMAND        = 1
)

// SSDMode values
type SSDMode int

const (
	SSDModeDefaultFW   SSDMode = 0
	SSDModePerformance SSDMode = 1
	SSDModePowerSaving SSDMode = 2
	SSDModeStandard   SSDMode = 3
)

var ssdModeNames = []string{"Default", "Performance", "Power Saving", "Standard"}

func (m SSDMode) String() string {
	if m >= 0 && int(m) < len(ssdModeNames) {
		return ssdModeNames[int(m)]
	}
	return "Unknown"
}

// SSDInfo holds info for a single physical SSD
type SSDInfo struct {
	DriveIndex       int    `json:"driveIndex"`
	Name            string `json:"name"`
	Model           string `json:"model"`
	SerialNumber    string `json:"serialNumber"`
	CapacityBytes   uint64 `json:"capacityBytes"`
	CapacityStr     string `json:"capacityStr"`
	Protocol        string `json:"protocol"`
	MultiModeCapable bool  `json:"multiModeCapable"`
	CurrentMode     int     `json:"currentMode"`
	CurrentModeStr  string `json:"currentModeStr"`
	Error           string `json:"error"`
}

// SSDModeResult is the result of a mode set operation
type SSDModeResult struct {
	DriveIndex int    `json:"driveIndex"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
}

// GetSSDInfo returns SSD information for all physical drives
func GetSSDInfo() []SSDInfo {
	var results []SSDInfo

	script := `
$ErrorActionPreference = 'SilentlyContinue'
$disks = Get-CimInstance -ClassName MSFT_PhysicalDisk -Namespace root\Microsoft\Windows\Storage
foreach ($disk in $disks) {
    if ($disk.BusType -notin @(3, 6, 17)) { continue }
    $friendly = $disk.FriendlyName
    $model = $disk.Model
    $serial = $disk.SerialNumber
    $size = $disk.Size
    $busType = $disk.BusType
    $driveIndex = $disk.DeviceId
    $operational = $disk.OperationalStatus
    if ($size -ge 1TB) { $capStr = "{0:N1} TB" -f ($size/1TB) }
    elseif ($size -ge 1GB) { $capStr = "{0:N0} GB" -f ($size/1GB) }
    else { $capStr = "N/A" }
    switch ($busType) {
        3   { $proto = "NVMe" }
        6   { $proto = "SATA" }
        17  { $proto = "USB" }
        default { $proto = "BusType$busType" }
    }
    Write-Output "DISK|$driveIndex|$friendly|$model|$serial|$size|$capStr|$proto|$operational"
}
`
	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		return []SSDInfo{{Error: "Failed to query physical disks"}}
	}

	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "DISK|") || line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 9 {
			continue
		}

		driveIndex, _ := strconv.Atoi(strings.TrimSpace(parts[1]))
		friendly := strings.TrimSpace(parts[2])
		model := strings.TrimSpace(parts[3])
		serial := strings.TrimSpace(parts[4])
		capacity, _ := strconv.ParseUint(strings.TrimSpace(parts[5]), 10, 64)
		capStr := strings.TrimSpace(parts[6])
		proto := strings.TrimSpace(parts[7])
		operational := strings.TrimSpace(parts[8])

		info := SSDInfo{
			DriveIndex:       driveIndex,
			Name:            friendly,
			Model:           model,
			SerialNumber:    serial,
			CapacityBytes:   capacity,
			CapacityStr:     capStr,
			Protocol:        proto,
			MultiModeCapable: false,
			CurrentMode:    -1,
			CurrentModeStr: "N/A",
			Error:           "",
		}

		// Decode OperationalStatus: 2 = Online/Power Saving, 1 = Unknown
		statusStr := "Unknown"
		if operational == "Online" {
			statusStr = "Online"
		} else if operational != "" {
			statusStr = operational
		}

		if proto == "NVMe" {
			// Try NVMe IOCTL first
			modeVal, ioErr := getNVMeModeFromIOCTL(driveIndex)
			if ioErr == nil && modeVal >= 0 {
				info.MultiModeCapable = true
				info.CurrentMode = modeVal
				info.CurrentModeStr = SSDMode(modeVal).String()
				info.Error = fmt.Sprintf("Status: %s | Mode: %s", statusStr, info.CurrentModeStr)
			} else {
				// Fallback: try registry
				modeVal2, regErr := getSSDModeFromDispatcherReg()
				if regErr == nil && modeVal2 >= 0 {
					info.MultiModeCapable = true
					info.CurrentMode = modeVal2
					info.CurrentModeStr = SSDMode(modeVal2).String()
					info.Error = fmt.Sprintf("Status: %s | Mode: %s", statusStr, info.CurrentModeStr)
				} else {
					info.Error = fmt.Sprintf("Status: %s", statusStr)
				}
			}
		} else {
			info.Error = fmt.Sprintf("Status: %s", statusStr)
		}

		results = append(results, info)
	}

	return results
}

// SetSSDMode sets the SSD mode for a specific physical drive
func SetSSDMode(physicalDriveIndex int, mode SSDMode) SSDModeResult {
	result := SSDModeResult{
		DriveIndex: physicalDriveIndex,
		Success:    false,
		Message:    "",
	}

	script := fmt.Sprintf(`
$ErrorActionPreference = 'SilentlyContinue'
$disk = Get-CimInstance -ClassName MSFT_PhysicalDisk -Namespace root\Microsoft\Windows\Storage | Where-Object { $_.DeviceId -eq %d }
if ($null -eq $disk) {
    Write-Output "NOTFOUND"
    return
}
if ($disk.BusType -ne 3) {
    Write-Output "NOTNVME|Only NVMe SSDs support mode switching (BusType=$($disk.BusType))"
    return
}
Write-Output "OK"
`, physicalDriveIndex)

	out, err := hiddenCmd("powershell", "-NoProfile", "-NonInteractive", "-WindowStyle", "Hidden", "-Command", script).Output()
	if err != nil {
		result.Message = "Failed to query disk info"
		return result
	}

	line := strings.TrimSpace(string(out))
	if strings.HasPrefix(line, "NOTFOUND") {
		result.Message = fmt.Sprintf("Physical drive %d not found", physicalDriveIndex)
		return result
	}
	if strings.HasPrefix(line, "NOTNVME|") {
		result.Message = strings.TrimPrefix(line, "NOTNVME|")
		return result
	}

	writeSSDMRegistryValue("ITS_SSDModeSetting", uint32(mode))

	result.Success = true
	result.Message = fmt.Sprintf("SSD mode set to %s (%d). Restart Dispatcher service to apply.", mode.String(), mode)
	return result
}

// ── Registry helpers ──────────────────────────────────────────────────────

func getSSDModeFromDispatcherReg() (int, error) {
	k, err := registry.OpenKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`,
		registry.QUERY_VALUE,
	)
	if err != nil {
		return -1, err
	}
	defer k.Close()
	val, _, err := k.GetIntegerValue("ITS_SSDModeSetting")
	if err != nil {
		return -1, err
	}
	return int(val), nil
}

func writeSSDMRegistryValue(name string, value uint32) {
	k, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Services\LenovoProcessManagement\Performance\PowerSlider`,
		registry.SET_VALUE,
	)
	if err != nil {
		return
	}
	defer k.Close()
	k.SetDWordValue(name, value)
}

// ── NVMe IOCTL ────────────────────────────────────────────────────────────

// getNVMeModeFromIOCTL reads the current SSD mode via NVMe VUC command
func getNVMeModeFromIOCTL(physicalDriveIndex int) (int, error) {
	h, err := openPhysicalDrive(physicalDriveIndex)
	if err != nil {
		return -1, fmt.Errorf("cannot open drive %d: %w", physicalDriveIndex, err)
	}
	defer windows.CloseHandle(h)

	// GET mode: VUC opcode 0xC6, subcmd 0x0F, magic 0x4657
	cdw12 := uint32(SUBCMD_GET_MODE)<<16 | uint32(VUC_MAGIC&0xFFFF)
	dw0, err := nvmeVUCCommand(h, VUC_OPC, 0, cdw12, 0)
	if err != nil {
		return -1, fmt.Errorf("VUC command failed: %w", err)
	}
	mode := int(dw0 & 0xFF)
	if mode < 0 || mode > 3 {
		return -1, fmt.Errorf("unexpected mode value: %d", mode)
	}
	return mode, nil
}

func openPhysicalDrive(index int) (windows.Handle, error) {
	path := fmt.Sprintf(`\\.\PhysicalDrive%d`, index)
	return windows.CreateFile(
		windows.StringToUTF16Ptr(path),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		nil,
		windows.OPEN_EXISTING,
		windows.FILE_ATTRIBUTE_NORMAL,
		0,
	)
}

// nvmeVUCCommand sends a vendor-unique command (VUC) to the NVMe drive
// Returns the Fixed DW0 value from the completion queue entry
func nvmeVUCCommand(h windows.Handle, opc uint8, cdw10, cdw12, cdw13 uint32) (uint32, error) {
	// Buffer layout:
	// [0..71]   STORAGE_PROTOCOL_COMMAND header (72 bytes)
	// [72..135] NVME_COMMAND (64 bytes)
	// [136..143] NVMe completion (8 bytes) — used for return data
	const (
		totalLen    = 72 + 64 + 8
		cmdOffset   = 72
		compOffset  = 136
		fixedOffset = 56 // Fixed DW0 within completion (at byte offset compOffset+0 within buffer)
	)

	buf := make([]byte, totalLen)

	// STORAGE_PROTOCOL_COMMAND (72 bytes)
	*(*uint32)(unsafe.Pointer(&buf[0])) = 0x10  // Version
	*(*uint32)(unsafe.Pointer(&buf[4])) = 144  // Length
	*(*uint32)(unsafe.Pointer(&buf[8])) = 0x17 // ProtocolType = NVMe
	*(*uint32)(unsafe.Pointer(&buf[12])) = 0x00000001 // Flags = ADAPTER_REQUEST
	*(*uint32)(unsafe.Pointer(&buf[16])) = 64   // CommandLength
	*(*uint32)(unsafe.Pointer(&buf[20])) = 8    // ErrorInfoLength
	*(*uint32)(unsafe.Pointer(&buf[24])) = 0    // DataFromDeviceTransferLength
	*(*uint32)(unsafe.Pointer(&buf[28])) = 10    // TimeoutValue (seconds)
	*(*uint32)(unsafe.Pointer(&buf[32])) = compOffset // ErrorInfoOffset
	*(*uint32)(unsafe.Pointer(&buf[36])) = 0    // DataFromDeviceBufferOffset
	*(*uint32)(unsafe.Pointer(&buf[40])) = 1    // CommandSpecific = NVME_ADMIN_COMMAND

	// NVME_COMMAND at offset 72
	*(*uint8)(unsafe.Pointer(&buf[cmdOffset+0])) = opc   // Opcode
	// Bytes 1-7: reserved (already 0)
	*(*uint32)(unsafe.Pointer(&buf[cmdOffset+8])) = cdw10 // CDW10
	*(*uint32)(unsafe.Pointer(&buf[cmdOffset+16])) = cdw12 // CDW12
	*(*uint32)(unsafe.Pointer(&buf[cmdOffset+24])) = cdw13 // CDW13

	var bytesReturned uint32
	ret := windows.DeviceIoControl(
		h,
		IOCTL_STORAGE_PROTOCOL_COMMAND,
		&buf[0],
		uint32(len(buf)),
		&buf[0],
		uint32(len(buf)),
		&bytesReturned,
		nil,
	)
	if ret != nil {
		// DW0 is at compOffset+0 within the buffer (completion starts there)
		return *(*uint32)(unsafe.Pointer(&buf[compOffset])), nil
	}

	// Try with DEVICE_REQUEST flag
	*(*uint32)(unsafe.Pointer(&buf[12])) = 0x00000002 // DEVICE_REQUEST
	ret = windows.DeviceIoControl(
		h,
		IOCTL_STORAGE_PROTOCOL_COMMAND,
		&buf[0],
		uint32(len(buf)),
		&buf[0],
		uint32(len(buf)),
		&bytesReturned,
		nil,
	)
	if ret != nil {
		return *(*uint32)(unsafe.Pointer(&buf[compOffset])), nil
	}

	return 0, fmt.Errorf("IOCTL_STORAGE_PROTOCOL_COMMAND failed")
}
