//go:build windows

package backend

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"golang.org/x/sys/windows/registry"
)

const (
	HM_SYS_DEVICE_NAME_LEN = 16
	HM_SYS_DEVICE_SN_LEN   = 32
	HM_MAX_DEVICES         = 32
)

type hmDeviceInfo struct {
	NumDevices uint32
	DeviceIDs  [HM_MAX_DEVICES]uint32
}

type hmMemInfo struct {
	MemTotal uint32
	MemUsed  uint32
	MemAvail uint32
}

type hmDvfsMode int32

const (
	HM_DVFS_PERFORMANCE hmDvfsMode = 0
	HM_DVFS_ONDEMAND    hmDvfsMode = 1
	HM_DVFS_MODE_MAX    hmDvfsMode = 2
)

const (
	HM_FW_UPGRADE_SUCCESS               = 0
	HM_FW_UPGRADE_INVALID_PARAM         = 1
	HM_FW_UPGRADE_MEM_ALLOC_FAILED      = 2
	HM_FW_UPGRADE_FILE_OPER_FAILED     = 3
	HM_FW_UPGRADE_INVALID_HEADER       = 4
	HM_FW_UPGRADE_INVALID_MAGIC        = 5
	HM_FW_UPGRADE_INVALID_META         = 6
	HM_FW_UPGRADE_INVALID_IMAGE_SIZE   = 7
	HM_FW_UPGRADE_INVALID_IMAGE_CRC    = 8
	HM_FW_UPGRADE_DEVICE_OPER_FAILED   = 9
	HM_FW_UPGRADE_FLASH_PROGRAM_FAILED = 10
	HM_FW_UPGRADE_THREAD_ERROR         = 11
	HM_FW_UPGRADE_BUSY                 = 12
)

type npuFuncs struct {
	getDeviceInfo       uintptr
	checkDeviceIndex    uintptr
	getVendorID         uintptr
	getDeviceSN         uintptr
	getDeviceName       uintptr
	getComputingPower   uintptr
	getCoreCount        uintptr
	getIPUUtiliRate     uintptr
	getIPUCoreUtiliRate uintptr
	getIPUVoltage       uintptr
	getIPUFrequency     uintptr
	getMemInfo          uintptr
	getTemperature      uintptr
	getBuildtime        uintptr
	getVersion          uintptr
	getDriverVersion    uintptr
	getDeviceVersion    uintptr
	getDDRSize          uintptr
	getBDF              uintptr
	getBandwidth        uintptr
	getBoardPower       uintptr
	getDVFSMode         uintptr
	setDVFSMode         uintptr
	getCTCPHYID         uintptr
	flashRead           uintptr
	flashProgram        uintptr
}

// NPU scheduler state
type NPOSchedulerState struct {
	Running       bool    `json:"running"`
	DevIndex      int     `json:"devIndex"`
	CurMode       string  `json:"curMode"`
	CurUtilPct    float64 `json:"curUtilPct"`
	CurTempC      float64 `json:"curTempC"`
	CurPowerW     float64 `json:"curPowerW"`
	CurFreqMHz    float64 `json:"curFreqMHz"`
	CurLockMaxMHz int     `json:"curLockMaxMHz"`
	CurLockMinMHz int     `json:"curLockMinMHz"`
	Decision      string  `json:"decision"`       // why mode was chosen
	LastSwitch    string  `json:"lastSwitch"`     // timestamp of last switch
}

type NPOSchedulerSettings struct {
	TempWarnC   float64 `json:"tempWarnC"`    // > this → force ONDEMAND
	TempCritC   float64 `json:"tempCritC"`    // > this → force lowest power
	UtilHighPct  float64 `json:"utilHighPct"`  // > this → PERFORMANCE (0-100)
	UtilLowPct   float64 `json:"utilLowPct"`   // < this → ONDEMAND (0-100)
	CheckSec     int     `json:"checkSec"`     // polling interval seconds
}

var (
	npuDLL          *syscall.LazyDLL
	npuFunc         npuFuncs
	npuOnce         sync.Once
	npuOnceErr      error
	npuMutex   sync.RWMutex
)

func InitNPU() error {
	npuOnce.Do(func() {
		npuOnceErr = npuLoad()
	})
	return npuOnceErr
}

// npuLoad loads libhal_xh2a.dll from known installation locations.
// Search order mirrors the C++ test.cpp: exe dir, source repo, driver install path.
// npuLoadDebug stores the last load attempt result for diagnostics.
// Exposed via a Wails diagnostic binding so the UI can display it.
var npuLoadDebug = ""

func npuLoad() error {
	// ── Step 1: Collect all candidate DLL paths ──────────────────────────────
	var candidates []string

	exeDir := getExeDir()

	// A) Exe directory (same folder as the app) — highest priority
	candidates = append(candidates, filepath.Join(exeDir, "libhal_xh2a.dll"))

	// B) Wails build output (exe is at build/bin)
	candidates = append(candidates, filepath.Join(exeDir, "build", "bin", "libhal_xh2a.dll"))

	// C) Known relative-layout paths
	relPaths := []string{
		// From exe at C:\LenovoDispatcherToolkit\build\bin\ → project root
		filepath.Join(exeDir, "..", "..", "backend", "dynamic_npu", "libhal_xh2a.dll"),
		filepath.Join(exeDir, "..", "..", "..", "backend", "dynamic_npu", "libhal_xh2a.dll"),
		// From project root C:\LenovoDispatcherToolkit\
		filepath.Join(exeDir, "backend", "dynamic_npu", "libhal_xh2a.dll"),
	}
	for _, p := range relPaths {
		abs, _ := filepath.Abs(p)
		candidates = append(candidates, abs)
	}

	// C) Source repo (where the original C++ test lived)
	candidates = append(candidates, `C:\Users\3-64\source\repos\Project1\Project1\libhal_xh2a.dll`)

	// D) Scan Registry for installed houmo-drv-xh2* drivers
	//    HKLM\SOFTWARE\Houmo, HKLM\SOFTWARE\WOW6432Node\Houmo, etc.
	registryRoots := []string{
		`SOFTWARE\Houmo`,
		`SOFTWARE\WOW6432Node\Houmo`,
		`SOFTWARE\houmo`,
		`SOFTWARE\WOW6432Node\houmo`,
		`SOFTWARE\houmo-drv-xh2`,
		`SOFTWARE\WOW6432Node\houmo-drv-xh2`,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`,
	}
	for _, root := range registryRoots {
		key, err := registry.OpenKey(registry.LOCAL_MACHINE, root, registry.QUERY_VALUE)
		if err != nil {
			continue
		}

		// Read all sub-key names at once
		subKeyNames, err := key.ReadSubKeyNames(0)
		key.Close()
		if err != nil {
			continue
		}

		for _, subKeyName := range subKeyNames {
			// Open the sub-key and read InstallLocation / ImagePath
			subKey, err := registry.OpenKey(registry.LOCAL_MACHINE,
				filepath.Join(root, subKeyName), registry.QUERY_VALUE)
			if err != nil {
				continue
			}
			// Try reading InstallLocation (REG_SZ)
			for _, valName := range []string{"InstallLocation", "Install_Path", "Path", "ImagePath"} {
				instDir, _, err := subKey.GetStringValue(valName)
				if err == nil && instDir != "" {
					// Try both root and hal\lib sub-folder
					for _, sub := range []string{``, `hal\lib`, `tools\hm_smi`} {
						dir := instDir
						if sub != `` {
							dir = filepath.Join(instDir, sub)
						}
						candidates = append(candidates, filepath.Join(dir, `libhal_xh2a.dll`))
					}
				}
			}
			subKey.Close()
		}
	}

	// E) Known installation paths for houmo-drv-xh2
	knownRoots := []string{
		`C:\Program Files (x86)\`,
		`C:\Program Files\`,
	}
	// Glob for houmo-drv-xh2* folders
	for _, root := range knownRoots {
		entries, err := os.ReadDir(root)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			name := strings.ToLower(e.Name())
			if strings.Contains(name, "houmo-drv-xh2") || strings.Contains(name, "houmo") && strings.Contains(name, "xh2") {
				for _, sub := range []string{``, `hal\lib`, `lib`, `bin`} {
					dir := e.Name()
					if sub != `` {
						dir = filepath.Join(e.Name(), sub)
					}
					candidates = append(candidates, filepath.Join(root, dir, `libhal_xh2a.dll`))
				}
			}
		}
	}

	// Deduplicate
	seen := make(map[string]bool)
	var unique []string
	for _, c := range candidates {
		abs, _ := filepath.Abs(c)
		if !seen[abs] {
			seen[abs] = true
			unique = append(unique, abs)
		}
	}
	candidates = unique

	// ── Step 2: Find the first path that exists ───────────────────────────────
	var dllPath string
	var tried []string
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			dllPath = p
			break
		}
		tried = append(tried, p)
	}

	if dllPath == "" {
		msg := "[NPU Load] DLL not found in any of these paths:\n"
		for _, p := range tried {
			msg += fmt.Sprintf("  MISSING: %s\n", p)
		}
		npuLoadDebug = msg
		return fmt.Errorf("libhal_xh2a.dll not found in %d paths; see diagnostic log", len(tried))
	}

	// ── Step 3: Load DLL and resolve functions ─────────────────────────────────
	npuDLL = syscall.NewLazyDLL(dllPath)

	// Build a readable load report
	loadReport := fmt.Sprintf(
		"[NPU Load] DLL found at: %s\n\n[Function resolution]\n", dllPath)

	// Write diagnostic to DLL location directory
	diagPath := filepath.Join(filepath.Dir(dllPath), "npu_diag.txt")
	if f, err := os.Create(diagPath); err == nil {
		fmt.Fprintf(f, "=== NPU DIAGNOSTIC ===\n")
		fmt.Fprintf(f, "dllPath: %s\n", dllPath)
		fmt.Fprintf(f, "diagPath: %s\n", diagPath)
		fmt.Fprintf(f, "\n%s\n", loadReport)
		f.Close()
	} else {
		npuLoadDebug += "[Diag] Failed to write " + diagPath + ": " + err.Error() + "\n"
	}

	resolve := func(name string) uintptr {
		p := npuDLL.NewProc(name)
		if err := p.Find(); err != nil {
			loadReport += fmt.Sprintf("  MISSING: %s\n", name)
			return 0
		}
		loadReport += fmt.Sprintf("  OK:      %s\n", name)
		return p.Addr()
	}

	npuFunc.getDeviceInfo       = resolve("hm_sys_get_device_info")
	npuFunc.checkDeviceIndex    = resolve("hm_sys_check_device_index")
	npuFunc.getVendorID         = resolve("hm_sys_get_vendor_id")
	npuFunc.getDeviceSN         = resolve("hm_sys_get_device_sn")
	npuFunc.getDeviceName       = resolve("hm_sys_get_device_name")
	npuFunc.getComputingPower   = resolve("hm_sys_get_computing_power")
	npuFunc.getCoreCount        = resolve("hm_sys_get_core_count")
	npuFunc.getIPUUtiliRate     = resolve("hm_sys_get_ipu_utili_rate")
	npuFunc.getIPUCoreUtiliRate = resolve("hm_sys_get_ipu_core_utili_rate")
	npuFunc.getIPUVoltage       = resolve("hm_sys_get_ipu_voltage")
	npuFunc.getIPUFrequency     = resolve("hm_sys_get_ipu_frequency")
	npuFunc.getMemInfo          = resolve("hm_sys_get_mem_info")
	npuFunc.getTemperature      = resolve("hm_sys_get_temperature")
	npuFunc.getBuildtime        = resolve("hm_sys_get_buildtime")
	npuFunc.getVersion          = resolve("hm_sys_get_version")
	npuFunc.getDriverVersion    = resolve("hm_sys_get_driver_version")
	npuFunc.getDeviceVersion    = resolve("hm_sys_get_device_version")
	npuFunc.getDDRSize          = resolve("hm_sys_get_ddr_size")
	npuFunc.getBDF              = resolve("hm_sys_get_bdf")
	npuFunc.getBandwidth        = resolve("hm_sys_get_bandwidth")
	npuFunc.getBoardPower       = resolve("hm_sys_get_board_power")
	npuFunc.getDVFSMode         = resolve("hm_sys_get_dvfs_mode")
	npuFunc.setDVFSMode         = resolve("hm_sys_set_dvfs_mode")
	npuFunc.getCTCPHYID         = resolve("hm_sys_get_ctc_phy_id")
	npuFunc.flashRead           = resolve("hm_flash_read")
	npuFunc.flashProgram        = resolve("hm_flash_program")

		npuLoadDebug = loadReport // store for UI diagnostics

	// Also append to diagnostic file in DLL directory
	if dllPath != "" {
		diagPath := filepath.Join(filepath.Dir(dllPath), "npu_diag.txt")
		if f, err := os.OpenFile(diagPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
			fmt.Fprintf(f, "\n=== NPU LOAD COMPLETE ===\n")
			f.Close()
		}
	}

	if npuFunc.getDeviceInfo == 0 {
		return fmt.Errorf("hm_sys_get_device_info not found in DLL; the loaded DLL does not export the expected HAL API. See diagnostic log.")
	}

	// ── Step 4: Quick enumerate devices (mirrors C++ test.cpp) ─────────────────
	// Call hm_sys_get_device_info immediately to confirm the driver is functional.
	var info hmDeviceInfo
	ret, _, _ := syscall.Syscall(npuFunc.getDeviceInfo, 1, uintptr(unsafe.Pointer(&info)), 0, 0)
	n := info.NumDevices
	if n == 0 {
		n = uint32(ret)
	}
	npuLoadDebug += fmt.Sprintf("\n[NPU Probe] hm_sys_get_device_info returned: numDevices=%d ret=%d\n", n, ret)

	if n == 0 {
		// Treat 0 devices as a valid (but empty) enumeration, not an error.
		// This matches C++ test.cpp behavior: "found 0 npu device" is printed, no error is thrown.
		npuLoadDebug += "[NPU Probe] WARNING: 0 devices reported — NPU may be disabled in BIOS or not present.\n"
		return nil // initialization succeeded; device list will be empty
	}

	npuLoadDebug += fmt.Sprintf("[NPU Probe] Found %d device(s): %v\n", n, info.DeviceIDs[:n])
	return nil
}

// GetNPULoadDebug exposes the diagnostic log to the Wails frontend.
func GetNPULoadDebug() string {
	return npuLoadDebug
}

// GetNPUREport returns a detailed diagnostic report combining the DLL load
// report with the raw probe output, for display in the UI diagnostic panel.
func GetNPUREport() string {
	// Force InitNPU() to populate npuLoadDebug
	_ = InitNPU()
	report, _ := NPURawProbe()
	// Write to diagnostic file when GetNPUREport is called
	diagPath := filepath.Join(getExeDir(), "..", "..", "wails_npu_diag.txt")
	if f, err := os.Create(diagPath); err == nil {
		fmt.Fprintf(f, "=== Wails GetNPUREport Called ===\n")
		fmt.Fprintf(f, "exeDir: %s\n", getExeDir())
		fmt.Fprintf(f, "diagPath: %s\n", diagPath)
		fmt.Fprintf(f, "\nnpuLoadDebug:\n%s\n", npuLoadDebug)
		f.Close()
	}

	// TEST: Try direct LoadLibrary (same as standalone Go program)
	testDlls := []string{
		`C:\Users\3-64\source\repos\Project1\Project1\libhal_xh2a.dll`,
		`C:\LenovoDispatcherToolkit\build\bin\libhal_xh2a.dll`,
		`C:\Program Files (x86)\houmo-drv-xh2_v1.1.0\hal\lib\libhal_xh2a.dll`,
	}
	for _, testPath := range testDlls {
		if _, err := os.Stat(testPath); err != nil {
			npuLoadDebug += "\n[Direct Load] " + testPath + " - FILE NOT FOUND\n"
			continue
		}
		dll, err := syscall.LoadDLL(testPath)
		if err != nil {
			npuLoadDebug += "\n[Direct Load] " + testPath + " - LoadLibrary FAILED: " + err.Error() + "\n"
			continue
		}
		proc, err := dll.FindProc("hm_sys_get_device_info")
		dll.Release()
		if err != nil {
			npuLoadDebug += "\n[Direct Load] " + testPath + " - GetProcAddress FAILED: " + err.Error() + "\n"
			continue
		}
		type hmDI struct {
			NumDevices uint32
			_          [4]byte
			DeviceIDs  [32]uint32
		}
		var info hmDI
		ret, _, _ := syscall.Syscall(proc.Addr(), 1, uintptr(unsafe.Pointer(&info)), 0, 0)
		npuLoadDebug += "\n[Direct Load] " + testPath + " - ret=" + fmt.Sprintf("%d", ret) + " NumDevices=" + fmt.Sprintf("%d", info.NumDevices) + "\n"
	}
	return npuLoadDebug + "\n\n[Raw Probe Report]\n" + report
}

// NPURawProbe calls hm_sys_get_device_info and returns the raw bytes so the UI
// can see exactly what the driver returned. This is diagnostic-only.
func NPURawProbe() (string, error) {
	if err := InitNPU(); err != nil {
		return "", err
	}

	npuMutex.RLock()
	defer npuMutex.RUnlock()

	// Test both possible struct layouts
	type hmDeviceInfoV1 struct { // no padding: NumDevices(4) + DeviceIDs[32](128) = 132 bytes
		NumDevices uint32
		DeviceIDs  [HM_MAX_DEVICES]uint32
	}
	type hmDeviceInfoV2 struct { // with 4-byte padding after NumDevices: offset 0 + pad(4) + offset 8
		NumDevices uint32
		_         uint32 // padding
		DeviceIDs  [HM_MAX_DEVICES]uint32
	}
	type hmDeviceInfoV3 struct { // 8-byte padding (num + padding + maybe reserve)
		NumDevices    uint32
		_             uint32 // padding
		_             uint32 // reserved
		DeviceIDs     [HM_MAX_DEVICES]uint32
	}

	var raw [256]byte
	var infoV1 hmDeviceInfoV1
	var infoV2 hmDeviceInfoV2
	var infoV3 hmDeviceInfoV3

	ret1, _, _ := syscall.Syscall(npuFunc.getDeviceInfo, 1, uintptr(unsafe.Pointer(&infoV1)), 0, 0)
	ret2, _, _ := syscall.Syscall(npuFunc.getDeviceInfo, 1, uintptr(unsafe.Pointer(&infoV2)), 0, 0)
	ret3, _, _ := syscall.Syscall(npuFunc.getDeviceInfo, 1, uintptr(unsafe.Pointer(&infoV3)), 0, 0)

	// Also capture raw bytes
	_, _, _ = syscall.Syscall(npuFunc.getDeviceInfo, 1, uintptr(unsafe.Pointer(&raw)), 0, 0)

	v1Devs := make([]uint32, 0)
	for i := uint32(0); i < infoV1.NumDevices && i < HM_MAX_DEVICES; i++ {
		v1Devs = append(v1Devs, infoV1.DeviceIDs[i])
	}
	v2Devs := make([]uint32, 0)
	for i := uint32(0); i < infoV2.NumDevices && i < HM_MAX_DEVICES; i++ {
		v2Devs = append(v2Devs, infoV2.DeviceIDs[i])
	}
	v3Devs := make([]uint32, 0)
	for i := uint32(0); i < infoV3.NumDevices && i < HM_MAX_DEVICES; i++ {
		v3Devs = append(v3Devs, infoV3.DeviceIDs[i])
	}

	// Parse raw as if first uint32 is numDevices at offset 0,4,8
	rawN0 := uint32(0)
	rawN4 := uint32(0)
	rawN8 := uint32(0)
	if len(raw) >= 4 {
		rawN0 = uint32(raw[0]) | uint32(raw[1])<<8 | uint32(raw[2])<<16 | uint32(raw[3])<<24
	}
	if len(raw) >= 8 {
		rawN4 = uint32(raw[4]) | uint32(raw[5])<<8 | uint32(raw[6])<<16 | uint32(raw[7])<<24
	}
	if len(raw) >= 12 {
		rawN8 = uint32(raw[8]) | uint32(raw[9])<<8 | uint32(raw[10])<<16 | uint32(raw[11])<<24
	}

	report := fmt.Sprintf(
		"[Raw Probe] sizeof(hm_device_info_t) expected ~136 bytes\n\n"+
			"V1 (no pad,  offset[0]=num):  ret=%d  numDevices=%d  deviceIds=%v\n"+
			"V2 (4B pad,  offset[8]=num):  ret=%d  numDevices=%d  deviceIds=%v\n"+
			"V3 (8B pad,  offset[12]=num): ret=%d  numDevices=%d  deviceIds=%v\n\n"+
			"Raw bytes[0..63]: % x\n"+
			"  offset[0] num=%d  offset[4] num=%d  offset[8] num=%d\n\n"+
			"npuLoadDebug:\n%s",
		ret1, infoV1.NumDevices, v1Devs,
		ret2, infoV2.NumDevices, v2Devs,
		ret3, infoV3.NumDevices, v3Devs,
		raw[:64],
		rawN0, rawN4, rawN8,
		npuLoadDebug,
	)
	return report, nil
}

// NPUDeviceInfo holds the enumerated device list.
type NPUDeviceInfo struct {
	NumDevices uint32   `json:"numDevices"`
	DeviceIDs  []uint32 `json:"deviceIds"`
}

func GetNPUDeviceInfo() (NPUDeviceInfo, error) {
	if err := InitNPU(); err != nil {
		return NPUDeviceInfo{}, err
	}
	var info hmDeviceInfo
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	ret, _, _ := syscall.Syscall(npuFunc.getDeviceInfo, 1, uintptr(unsafe.Pointer(&info)), 0, 0)
	devCount := info.NumDevices
	if devCount == 0 {
		devCount = uint32(ret)
	}
	// DIAGNOSTIC: always write to file so we know what happened
	diagPath := "C:\\LenovoDispatcherToolkit\\build\\bin\\npu_devcount.txt"
	if f, err := os.Create(diagPath); err == nil {
		fmt.Fprintf(f, "LazyDLL: NumDevices=%d ret=%d devCount=%d\n", info.NumDevices, ret, devCount)
		if devCount == 0 {
			fmt.Fprintf(f, "Calling direct LoadLibrary fallback...\n")
			if altCount, altIDs, altErr := tryDirectDLLLoad(); altErr == nil {
				fmt.Fprintf(f, "Fallback result: altCount=%d err=%v\n", altCount, altErr)
				if altCount > 0 {
					devIDs := make([]uint32, altCount)
					copy(devIDs, altIDs)
					fmt.Fprintf(f, "Fallback SUCCEEDED: returning %d devices\n", altCount)
					f.Close()
					return NPUDeviceInfo{NumDevices: altCount, DeviceIDs: devIDs}, nil
				} else {
					fmt.Fprintf(f, "Fallback returned 0\n")
				}
			} else {
				fmt.Fprintf(f, "Fallback error: %v\n", altErr)
			}
		}
		f.Close()
	}
	devIDs := make([]uint32, 0, devCount)
	for i := uint32(0); i < devCount && i < HM_MAX_DEVICES; i++ {
		devIDs = append(devIDs, info.DeviceIDs[i])
	}
	return NPUDeviceInfo{NumDevices: devCount, DeviceIDs: devIDs}, nil
}

func CheckNPUDeviceIndex(devIndex int) bool {
	if err := InitNPU(); err != nil {
		return false
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	ret, _, _ := syscall.Syscall(npuFunc.checkDeviceIndex, 1, uintptr(devIndex), 0, 0)
	return int(ret) == 0
}

// NPUDeviceProperties holds basic device properties.
type NPUDeviceProperties struct {
	VendorID       int    `json:"vendorId"`
	SerialNumber   string `json:"serialNumber"`
	ModelName      string `json:"modelName"`
	ComputingPower int    `json:"computingPowerTOPS"`
	CoreCount      int    `json:"coreCount"`
	DDRSizeBytes   uint64 `json:"ddrSizeBytes"`
	DDRSizeMB      uint32 `json:"ddrSizeMB"`
	FirmwareVer    string `json:"firmwareVersion"`
}

func GetNPUDeviceProperties(devIndex int) (NPUDeviceProperties, error) {
	if err := InitNPU(); err != nil {
		return NPUDeviceProperties{}, err
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	prop := NPUDeviceProperties{}
	if npuFunc.getVendorID != 0 {
		v, _, _ := syscall.Syscall(npuFunc.getVendorID, 1, uintptr(devIndex), 0, 0)
		prop.VendorID = int(v)
	}
	if npuFunc.getDeviceSN != 0 {
		var snBuf [HM_SYS_DEVICE_SN_LEN]byte
		syscall.Syscall(npuFunc.getDeviceSN, 3, uintptr(devIndex), uintptr(unsafe.Pointer(&snBuf[0])), uintptr(HM_SYS_DEVICE_SN_LEN))
		prop.SerialNumber = cstring(snBuf[:])
	}
	if npuFunc.getDeviceName != 0 {
		var nameBuf [HM_SYS_DEVICE_NAME_LEN]byte
		syscall.Syscall(npuFunc.getDeviceName, 3, uintptr(devIndex), uintptr(unsafe.Pointer(&nameBuf[0])), uintptr(HM_SYS_DEVICE_NAME_LEN))
		prop.ModelName = cstring(nameBuf[:])
	}
	if npuFunc.getComputingPower != 0 {
		v, _, _ := syscall.Syscall(npuFunc.getComputingPower, 1, uintptr(devIndex), 0, 0)
		prop.ComputingPower = int(v)
	}
	if npuFunc.getCoreCount != 0 {
		v, _, _ := syscall.Syscall(npuFunc.getCoreCount, 1, uintptr(devIndex), 0, 0)
		prop.CoreCount = int(v)
	}
	if npuFunc.getDDRSize != 0 {
		var ddr uint64
		syscall.Syscall(npuFunc.getDDRSize, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&ddr)), 0)
		prop.DDRSizeBytes = ddr
		prop.DDRSizeMB = uint32(ddr / 1024 / 1024)
	}
	if npuFunc.getDeviceVersion != 0 {
		var verBuf [64]byte
		syscall.Syscall(npuFunc.getDeviceVersion, 3, uintptr(devIndex), uintptr(unsafe.Pointer(&verBuf[0])), uintptr(64))
		prop.FirmwareVer = cstring(verBuf[:])
	}
	return prop, nil
}

// tryDirectDLLLoad attempts direct LoadLibrary as fallback when LazyDLL returns 0 devices.
// Standalone Go test confirms this approach works (returns 1 device).
func tryDirectDLLLoad() (uint32, []uint32, error) {
	testPaths := []string{
		`C:\Users\3-64\source\repos\Project1\Project1\libhal_xh2a.dll`,
		`C:\LenovoDispatcherToolkit\build\bin\libhal_xh2a.dll`,
		`C:\Program Files (x86)\houmo-drv-xh2_v1.1.0\hal\lib\libhal_xh2a.dll`,
	}
	type hmDI struct {
		NumDevices uint32
		_          [4]byte
		DeviceIDs  [32]uint32
	}
	for _, p := range testPaths {
		if _, err := os.Stat(p); err != nil {
			continue
		}
		dll, err := syscall.LoadDLL(p)
		if err != nil {
			continue
		}
		defer dll.Release()
		proc, err := dll.FindProc("hm_sys_get_device_info")
		if err != nil {
			continue
		}
		var info hmDI
		ret, _, _ := syscall.Syscall(proc.Addr(), 1, uintptr(unsafe.Pointer(&info)), 0, 0)
		n := info.NumDevices
		if n == 0 {
			n = uint32(ret)
		}
		if n > 0 {
			devIDs := make([]uint32, n)
			for i := uint32(0); i < n; i++ {
				devIDs[i] = info.DeviceIDs[i]
			}
			return n, devIDs, nil
		}
	}
	return 0, nil, fmt.Errorf("all DLL paths returned 0 devices")
}

// NPUDeviceMetrics holds real-time runtime metrics.
type NPUDeviceMetrics struct {
	IPUUtiliRate    float64   `json:"ipuUtiliRate"`
	IPUVoltageMV    float64   `json:"ipuVoltageMV"`
	IPUFrequencyHz uint64    `json:"ipuFrequencyHz"`
	BoardPowerW    float64   `json:"boardPowerW"`
	TemperatureC   float64   `json:"temperatureC"`
	MemTotalMB     uint32    `json:"memTotalMB"`
	MemUsedMB      uint32    `json:"memUsedMB"`
	MemAvailMB     uint32    `json:"memAvailMB"`
	CoreUtiliPct   []float64 `json:"coreUtiliPct"` // per-core utilization, 0-100
}

func GetNPUDeviceMetrics(devIndex int) (NPUDeviceMetrics, error) {
	if err := InitNPU(); err != nil {
		return NPUDeviceMetrics{}, err
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	m := NPUDeviceMetrics{}
	if npuFunc.getIPUUtiliRate != 0 {
		v, _, _ := syscall.Syscall(npuFunc.getIPUUtiliRate, 1, uintptr(devIndex), 0, 0)
		m.IPUUtiliRate = mathFloat(v)
	}
	if npuFunc.getIPUVoltage != 0 {
		var volt float32
		syscall.Syscall(npuFunc.getIPUVoltage, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&volt)), 0)
		m.IPUVoltageMV = float64(volt)
	}
	if npuFunc.getIPUFrequency != 0 {
		var freq uint64
		syscall.Syscall(npuFunc.getIPUFrequency, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&freq)), 0)
		m.IPUFrequencyHz = freq
	}
	if npuFunc.getBoardPower != 0 {
		var power float32
		syscall.Syscall(npuFunc.getBoardPower, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&power)), 0)
		m.BoardPowerW = float64(power)
	}
	if npuFunc.getTemperature != 0 {
		var temp float32
		syscall.Syscall(npuFunc.getTemperature, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&temp)), 0)
		m.TemperatureC = float64(temp)
	}
	if npuFunc.getMemInfo != 0 {
		var mem hmMemInfo
		syscall.Syscall(npuFunc.getMemInfo, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&mem)), 0)
		m.MemTotalMB = mem.MemTotal
		m.MemUsedMB = mem.MemUsed
		m.MemAvailMB = mem.MemAvail
	}
	// Per-core utilization: get core count from IPUUtiliRate call
	var coreCount uint32
	if npuFunc.getCoreCount != 0 {
		v, _, _ := syscall.Syscall(npuFunc.getCoreCount, 1, uintptr(devIndex), 0, 0)
		coreCount = uint32(v)
	}
	if coreCount > 0 && npuFunc.getIPUCoreUtiliRate != 0 {
		for i := uint32(0); i < coreCount; i++ {
			v, _, _ := syscall.Syscall(npuFunc.getIPUCoreUtiliRate, 3, uintptr(devIndex), uintptr(i), 0)
			m.CoreUtiliPct = append(m.CoreUtiliPct, mathFloat(v)*100)
		}
	}
	return m, nil
}

func GetNPUPerCoreUtili(devIndex int, coreCount int) ([]float64, error) {
	if err := InitNPU(); err != nil {
		return nil, err
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	rates := make([]float64, 0, coreCount)
	for i := 0; i < coreCount; i++ {
		v, _, _ := syscall.Syscall(npuFunc.getIPUCoreUtiliRate, 3, uintptr(devIndex), uintptr(uint32(i)), 0)
		rates = append(rates, mathFloat(v))
	}
	return rates, nil
}

// NPUDeviceOverview mirrors test.exe print_device_info() output for a single device.
type NPUDeviceOverview struct {
	// Basic Info
	DevID           int     `json:"devId"`
	DevName         string  `json:"devName"`
	VendorID        int     `json:"vendorId"`
	Serial          string  `json:"serial"`
	ComputingPower  int     `json:"computingPower"`
	CoreCount       int     `json:"coreCount"`
	DDRSizeMB       uint32  `json:"ddrSizeMB"`
	// DVFS Mode
	DVFSMode        string  `json:"dvfsMode"`
	DVFSModeDesc   string  `json:"dvfsModeDesc"`
	// Runtime Metrics
	IPUUtiliPct    float64 `json:"ipuUtiliPct"`
	IPUFreqGHz     float64 `json:"ipuFreqGHz"`
	IPUVoltageMV   float64 `json:"ipuVoltageMV"`
	TemperatureC   float64 `json:"temperatureC"`
	BoardPowerW    float64 `json:"boardPowerW"`
	// Memory Info
	MemTotalMB     uint32  `json:"memTotalMB"`
	MemUsedMB      uint32  `json:"memUsedMB"`
	MemAvailMB     uint32  `json:"memAvailMB"`
	// Per-Core
	CoreUtiliPct   []float64 `json:"coreUtiliPct"`
	// Version Info
	SDKVersion     string  `json:"sdkVersion"`
	DriverVersion  string  `json:"driverVersion"`
	FirmwareVer    string  `json:"firmwareVer"`
}

// GetNPUDeviceOverview returns full device info for devIndex, mirroring test.exe print_device_info().
func GetNPUDeviceOverview(devIndex int) (NPUDeviceOverview, error) {
	props, err := GetNPUDeviceProperties(devIndex)
	if err != nil {
		return NPUDeviceOverview{}, err
	}
	metrics, err := GetNPUDeviceMetrics(devIndex)
	if err != nil {
		return NPUDeviceOverview{}, err
	}
	sdkInfo, _ := GetNPUSDKInfo()
	dvfsMode, _ := NPUGetDVFSMode(devIndex)
	dvfsDesc := dvfsModeDesc(dvfsMode)

	// Get firmware version
	var firmware [64]byte
	npuMutex.RLock()
	if npuFunc.getDeviceVersion != 0 {
		syscall.Syscall(npuFunc.getDeviceVersion, 3, uintptr(devIndex), uintptr(unsafe.Pointer(&firmware[0])), uintptr(64))
	}
	npuMutex.RUnlock()

	return NPUDeviceOverview{
		DevID:           devIndex,
		DevName:         props.ModelName,
		VendorID:        props.VendorID,
		Serial:          props.SerialNumber,
		ComputingPower:  props.ComputingPower,
		CoreCount:       props.CoreCount,
		DDRSizeMB:       props.DDRSizeMB,
		DVFSMode:        dvfsMode,
		DVFSModeDesc:    dvfsDesc,
		IPUUtiliPct:     metrics.IPUUtiliRate * 100,
		IPUFreqGHz:      float64(metrics.IPUFrequencyHz) / 1e9,
		IPUVoltageMV:    metrics.IPUVoltageMV,
		TemperatureC:    metrics.TemperatureC,
		BoardPowerW:     metrics.BoardPowerW,
		MemTotalMB:      metrics.MemTotalMB,
		MemUsedMB:       metrics.MemUsedMB,
		MemAvailMB:      metrics.MemAvailMB,
		CoreUtiliPct:    metrics.CoreUtiliPct,
		SDKVersion:      sdkInfo.SDKVersion,
		DriverVersion:   sdkInfo.DriverVersion,
		FirmwareVer:     cstring(firmware[:]),
	}, nil
}

// dvfsModeDesc returns a human-readable description for each DVFS mode.
func dvfsModeDesc(mode string) string {
	switch mode {
	case "PERFORMANCE":
		return "Fixed 1400 MHz, max throughput, high power"
	case "ONDEMAND":
		return "Dynamic 200-1400 MHz, auto-adjust, low power"
	case "POWERSAVE":
		return "Fixed minimum frequency, lowest power"
	case "USERSPACE":
		return "User-defined frequency via clock lock"
	case "CONSERVATIVE":
		return "Slow ramp, conservative power saving"
	default:
		return "Auto-adjust frequency and power"
	}
}

// NPUSDKInfo holds SDK and driver version strings.
type NPUSDKInfo struct {
	Buildtime     string `json:"buildtime"`
	SDKVersion    string `json:"sdkVersion"`
	DriverVersion string `json:"driverVersion"`
}

func GetNPUSDKInfo() (NPUSDKInfo, error) {
	if err := InitNPU(); err != nil {
		return NPUSDKInfo{}, err
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	info := NPUSDKInfo{}
	if npuFunc.getBuildtime != 0 {
		var buf [64]byte
		syscall.Syscall(npuFunc.getBuildtime, 2, uintptr(unsafe.Pointer(&buf[0])), uintptr(64), 0)
		info.Buildtime = cstring(buf[:])
	}
	if npuFunc.getVersion != 0 {
		var buf [64]byte
		syscall.Syscall(npuFunc.getVersion, 2, uintptr(unsafe.Pointer(&buf[0])), uintptr(64), 0)
		info.SDKVersion = cstring(buf[:])
	}
	if npuFunc.getDriverVersion != 0 {
		var buf [64]byte
		syscall.Syscall(npuFunc.getDriverVersion, 2, uintptr(unsafe.Pointer(&buf[0])), uintptr(64), 0)
		info.DriverVersion = cstring(buf[:])
	}
	return info, nil
}

func NPUGetDVFSMode(devIndex int) (string, error) {
	if err := InitNPU(); err != nil {
		return "", err
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	var mode int32
	syscall.Syscall(npuFunc.getDVFSMode, 2, uintptr(devIndex), uintptr(unsafe.Pointer(&mode)), 0)
	switch hmDvfsMode(mode) {
	case HM_DVFS_PERFORMANCE:
		return "PERFORMANCE", nil
	case HM_DVFS_ONDEMAND:
		return "ONDEMAND", nil
	}
	return "UNKNOWN", nil
}

func NPUSetDVFSMode(devIndex int, mode string) error {
	if err := InitNPU(); err != nil {
		return err
	}
	var m int32
	switch strings.ToUpper(mode) {
	case "PERFORMANCE":
		m = int32(HM_DVFS_PERFORMANCE)
	case "ONDEMAND":
		m = int32(HM_DVFS_ONDEMAND)
	default:
		return fmt.Errorf("invalid DVFS mode: %s", mode)
	}
	npuMutex.Lock()
	defer npuMutex.Unlock()
	ret, _, _ := syscall.Syscall(npuFunc.setDVFSMode, 2, uintptr(devIndex), uintptr(m), 0)
	if int(ret) < 0 {
		return fmt.Errorf("hm_sys_set_dvfs_mode failed for dev %d", devIndex)
	}
	return nil
}

// NPUPCIeInfo holds PCIe BDF and bandwidth (Windows: hm_sys_get_bdf is Linux/Android only).
type NPUPCIeInfo struct {
	BDF       string `json:"bdf"`
	Bandwidth string `json:"bandwidth"`
}

func GetNPUPCIeInfo(_ int) (NPUPCIeInfo, error) {
	return NPUPCIeInfo{}, nil
}

// NPUCTCPHYInfo holds CTC PHY group/chip IDs.
type NPUCTCPHYInfo struct {
	GroupID int `json:"groupId"`
	ChipID  int `json:"chipId"`
}

func GetNPUCTCPHYInfo(devIndex int) (NPUCTCPHYInfo, error) {
	if err := InitNPU(); err != nil {
		return NPUCTCPHYInfo{}, err
	}
	npuMutex.RLock()
	defer npuMutex.RUnlock()
	var gid, cid int32
	ret, _, _ := syscall.Syscall(npuFunc.getCTCPHYID, 3, uintptr(devIndex), uintptr(unsafe.Pointer(&gid)), uintptr(unsafe.Pointer(&cid)))
	if int(ret) < 0 {
		return NPUCTCPHYInfo{}, fmt.Errorf("hm_sys_get_ctc_phy_id failed for dev %d", devIndex)
	}
	return NPUCTCPHYInfo{GroupID: int(gid), ChipID: int(cid)}, nil
}

// NPUFullReport is a consolidated report of all NPU data.
type NPUFullReport struct {
	DeviceCount int               `json:"deviceCount"`
	SDKInfo     NPUSDKInfo       `json:"sdkInfo"`
	Devices     []NPUDeviceReport `json:"devices"`
}

// NPUDeviceReport is per-device consolidated data.
type NPUDeviceReport struct {
	Index        int                   `json:"index"`
	Properties   NPUDeviceProperties   `json:"properties"`
	Metrics      NPUDeviceMetrics      `json:"metrics"`
	PCIeInfo     NPUPCIeInfo           `json:"pcieInfo"`
	DVFSMode     string                `json:"dvfsMode"`
	DVFSModeDesc string                `json:"dvfsModeDesc"`
	CTCPHYInfo   NPUCTCPHYInfo         `json:"ctcPhyInfo"`
}

func GetNPUFullReport() (NPUFullReport, error) {
	devInfo, err := GetNPUDeviceInfo()
	if err != nil {
		return NPUFullReport{}, err
	}
	sdkInfo, _ := GetNPUSDKInfo()
	report := NPUFullReport{
		DeviceCount: int(devInfo.NumDevices),
		SDKInfo:     sdkInfo,
		Devices:     make([]NPUDeviceReport, 0, devInfo.NumDevices),
	}
	for _, devID := range devInfo.DeviceIDs {
		props, _    := GetNPUDeviceProperties(int(devID))
		metrics, _  := GetNPUDeviceMetrics(int(devID))
		dvfsMode, _ := NPUGetDVFSMode(int(devID))
		ctcInfo, _  := GetNPUCTCPHYInfo(int(devID))
		report.Devices = append(report.Devices, NPUDeviceReport{
			Index: int(devID), Properties: props, Metrics: metrics,
			PCIeInfo: NPUPCIeInfo{}, DVFSMode: dvfsMode,
			DVFSModeDesc: dvfsModeDesc(dvfsMode), CTCPHYInfo: ctcInfo,
		})
	}
	return report, nil
}

// cstring converts a C-style null-terminated byte slice to a Go string.
func cstring(b []byte) string {
	n := len(b)
	for i, c := range b {
		if c == 0 {
			n = i
			break
		}
	}
	return string(b[:n])
}

// mathFloat reinterprets float32 bits stored as uintptr back to float64.
func mathFloat(bits uintptr) float64 {
	return float64(math.Float32frombits(uint32(bits)))
}

// ------------------------------------------------------------------
// hm_smi wrapper (mirrors C++ test.cpp run_hm_smi / parse_smi_field)
// ------------------------------------------------------------------

var hmSMI = `C:\Program Files (x86)\houmo-drv-xh2_v1.1.0\tools\hm_smi\hm_smi.exe`

// runHMCLI executes hm_smi with given args and returns stdout+stderr.
func runHMCLI(args ...string) (string, error) {
	cmd := hiddenCmd(hmSMI, args...)
	cmd.Env = os.Environ()
	out, err := cmd.CombinedOutput()
	if err != nil {
		return string(out), err
	}
	return string(out), nil
}

// parseSMIField extracts a field value from hm_smi output.
// Handles formats like "DVFS_Mode            : performance" or "DVFS_Mode : performance".
func parseSMIField(output, field string) string {
	idx := strings.Index(output, field)
	if idx == -1 {
		return ""
	}
	rest := output[idx+len(field):]
	// skip whitespace and colon
	for len(rest) > 0 && (rest[0] == ' ' || rest[0] == ':' || rest[0] == '\t') {
		rest = rest[1:]
	}
	// take until newline or carriage return
	end := 0
	for end < len(rest) && rest[end] != '\n' && rest[end] != '\r' {
		end++
	}
	return strings.TrimRight(rest[:end], " ")
}

// NPUPowerStatus holds the full power/DVFS snapshot from hm_smi -a -d <n>.
type NPUPowerStatus struct {
	DVFSMode       string  `json:"dvfsMode"`        // e.g. "performance", "ondemand", "powerlimit"
	CurIPUFreqMHz  float64 `json:"curIpuFreqMHz"`  // MHz
	LockIPUMaxMHz  float64 `json:"lockIpuMaxMHz"`  // MHz
	LockIPUMinMHz  float64 `json:"lockIpuMinMHz"`  // MHz
	IPULoadPct     float64 `json:"ipuLoadPct"`      // %
	BoardPowerW    float64 `json:"boardPowerW"`    // W
	DDRTotalMB     float64 `json:"ddrTotalMB"`      // MB
	DDRFreeMB      float64 `json:"ddrFreeMB"`       // MB
	CoreNum        int     `json:"coreNum"`
	CoreFreqMHz    float64 `json:"coreFreqMHz"`    // MHz
	CoreVoltageMV  float64 `json:"coreVoltageMV"`   // mV
	Core0UtilPct   float64 `json:"core0UtilPct"`   // %
	Core1UtilPct   float64 `json:"core1UtilPct"`   // %
	AvgUtilPct     float64 `json:"avgUtilPct"`     // %
	DDR0TempC      float64 `json:"ddr0TempC"`       // C
	DDR2TempC      float64 `json:"ddr2TempC"`       // C
	DDR4TempC      float64 `json:"ddr4TempC"`       // C
	DDR5TempC      float64 `json:"ddr5TempC"`       // C
	Core0TempC     float64 `json:"core0TempC"`      // C
	Core1TempC     float64 `json:"core1TempC"`      // C
}

// GetNPUPowerStatus returns a snapshot of power/DVFS/memory/temp for device <devIndex>.
func GetNPUPowerStatus(devIndex int) (NPUPowerStatus, error) {
	// First ensure NPU is enumerated via DLL (to confirm device exists)
	if err := InitNPU(); err != nil {
		return NPUPowerStatus{}, err
	}
	info, err := GetNPUDeviceInfo()
	if err != nil {
		return NPUPowerStatus{}, err
	}
	if devIndex < 0 || devIndex >= int(info.NumDevices) {
		return NPUPowerStatus{}, fmt.Errorf("device index %d out of range (have %d device(s))", devIndex, info.NumDevices)
	}

	out, err := runHMCLI("-a", "-d", fmt.Sprintf("%d", devIndex))
	if err != nil {
		return NPUPowerStatus{}, fmt.Errorf("hm_smi failed: %v\n%s", err, out)
	}

	s := NPUPowerStatus{}

	// DVFS / power block
	s.DVFSMode = strings.ToLower(parseSMIField(out, "DVFS_Mode"))

	s.CurIPUFreqMHz = parseFloat(parseSMIField(out, "Cur_Ipu_Freq"))
	s.LockIPUMaxMHz = parseFloat(parseSMIField(out, "Lock_Ipu_Max_Clock"))
	s.LockIPUMinMHz = parseFloat(parseSMIField(out, "Lock_Ipu_Min_Clock"))
	s.IPULoadPct = parseFloat(parseSMIField(out, "IPU_Load"))
	s.BoardPowerW = parseFloat(parseSMIField(out, "Board_Power"))

	// IPU Infos
	s.CoreNum = int(parseFloat(parseSMIField(out, "Core_Num")))
	s.CoreFreqMHz = parseFloat(parseSMIField(out, "Core_Freq"))
	s.CoreVoltageMV = parseFloat(parseSMIField(out, "Voltage"))
	s.Core0UtilPct = parseFloat(parseSMIField(out, "Core0_Util"))
	s.Core1UtilPct = parseFloat(parseSMIField(out, "Core1_Util"))
	s.AvgUtilPct = parseFloat(parseSMIField(out, "Average_Util"))

	// DDR memory
	s.DDRTotalMB = parseFloat(parseSMIField(out, "DDR_Memory_Total"))
	s.DDRFreeMB = parseFloat(parseSMIField(out, "DDR_Memory_Free"))

	// Temperature
	s.DDR0TempC = parseFloat(parseSMIField(out, "DDR0"))
	s.DDR2TempC = parseFloat(parseSMIField(out, "DDR2"))
	s.DDR4TempC = parseFloat(parseSMIField(out, "DDR4"))
	s.DDR5TempC = parseFloat(parseSMIField(out, "DDR5"))
	s.Core0TempC = parseFloat(parseSMIField(out, "Core0"))
	s.Core1TempC = parseFloat(parseSMIField(out, "Core1"))

	return s, nil
}

// NPUPowerAction represents the result of a power/DVFS action.
type NPUPowerAction struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	NewMode   string `json:"newMode,omitempty"`
	NewMaxMHz int    `json:"newMaxMHz,omitempty"`
	NewMinMHz int    `json:"newMinMHz,omitempty"`
}

// SetNPUMode sets DVFS mode: "performance" | "ondemand" | "powerlimit".
func SetNPUMode(devIndex int, mode string) (NPUPowerAction, error) {
	// Validate device
	info, err := GetNPUDeviceInfo()
	if err != nil {
		return NPUPowerAction{}, err
	}
	if devIndex < 0 || devIndex >= int(info.NumDevices) {
		return NPUPowerAction{}, fmt.Errorf("device index out of range")
	}

	var out string
	switch strings.ToLower(mode) {
	case "performance":
		out, err = runHMCLI("-g", "performance", "-d", fmt.Sprintf("%d", devIndex))
	case "ondemand":
		out, err = runHMCLI("-g", "ondemand", "-d", fmt.Sprintf("%d", devIndex))
	case "powerlimit":
		// Set to powerlimit mode (power limit values set separately via SetNPUPowerLimit)
		out, err = runHMCLI("-g", "powerlimit", "-d", fmt.Sprintf("%d", devIndex))
	default:
		return NPUPowerAction{}, fmt.Errorf("unsupported mode: %s (use: performance, ondemand, powerlimit)", mode)
	}

	if err != nil {
		return NPUPowerAction{}, fmt.Errorf("hm_smi failed: %v\n%s", err, out)
	}

	// Verify
	status, _ := GetNPUPowerStatus(devIndex)
	return NPUPowerAction{
		Success: true,
		Message: fmt.Sprintf("DVFS mode set to %s", mode),
		NewMode: status.DVFSMode,
	}, nil
}

// SetNPUPowerLimit sets the power limit for the given device (requires POWERLIMIT DVFS mode).
// maxW and minW are in Watts. Call SetNPUMode(dev, "powerlimit") first if needed.
func SetNPUPowerLimit(devIndex int, maxW, minW float64) (NPUPowerAction, error) {
	info, err := GetNPUDeviceInfo()
	if err != nil {
		return NPUPowerAction{}, err
	}
	if devIndex < 0 || devIndex >= int(info.NumDevices) {
		return NPUPowerAction{}, fmt.Errorf("device index out of range")
	}
	if maxW < 3 || maxW > 50 || minW < 1 || minW > maxW {
		return NPUPowerAction{}, fmt.Errorf("power limit out of range (3-50W, min ≤ max)")
	}

	// Switch to powerlimit mode first, then apply limit
	_, err = SetNPUMode(devIndex, "powerlimit")
	if err != nil {
		return NPUPowerAction{}, fmt.Errorf("failed to set POWERLIMIT mode: %v", err)
	}

	// hm_smi: -g powerlimit -pl maxWw minWw
	out, err := runHMCLI("-g", "powerlimit", "-pl",
		fmt.Sprintf("%.0fw", maxW), fmt.Sprintf("%.0fw", minW),
		"-d", fmt.Sprintf("%d", devIndex))
	if err != nil {
		return NPUPowerAction{Success: false, Message: fmt.Sprintf("hm_smi failed: %v\n%s", err, out)}, nil
	}
	return NPUPowerAction{Success: true, Message: fmt.Sprintf("Power limit set: %.0fW – %.0fW", maxW, minW)}, nil
}

// SetNPUClockLock locks IPU clock range (max/min in MHz, range 700–1400).
func SetNPUClockLock(devIndex, maxMHz, minMHz int) (NPUPowerAction, error) {
	if maxMHz < 700 || maxMHz > 1400 || minMHz < 700 || minMHz > 1400 {
		return NPUPowerAction{}, fmt.Errorf("clock must be between 700 and 1400 MHz")
	}
	if minMHz > maxMHz {
		return NPUPowerAction{}, fmt.Errorf("min clock cannot exceed max clock")
	}

	info, err := GetNPUDeviceInfo()
	if err != nil {
		return NPUPowerAction{}, err
	}
	if devIndex < 0 || devIndex >= int(info.NumDevices) {
		return NPUPowerAction{}, fmt.Errorf("device index out of range")
	}

	out, err := runHMCLI("-lc", fmt.Sprintf("%d", maxMHz), fmt.Sprintf("%d", minMHz), "-d", fmt.Sprintf("%d", devIndex))
	if err != nil {
		return NPUPowerAction{}, fmt.Errorf("hm_smi -lc failed: %v\n%s", err, out)
	}

	return NPUPowerAction{
		Success:   true,
		Message:   fmt.Sprintf("Clock locked to %d–%d MHz", minMHz, maxMHz),
		NewMaxMHz: maxMHz,
		NewMinMHz: minMHz,
	}, nil
}

// ResetNPUDefaults resets DVFS to PERFORMANCE and unlocks clocks (700–1400).
func ResetNPUDefaults(devIndex int) (NPUPowerAction, error) {
	info, err := GetNPUDeviceInfo()
	if err != nil {
		return NPUPowerAction{}, err
	}
	if devIndex < 0 || devIndex >= int(info.NumDevices) {
		return NPUPowerAction{}, fmt.Errorf("device index out of range")
	}

	_, err1 := runHMCLI("-g", "performance", "-d", fmt.Sprintf("%d", devIndex))
	_, err2 := runHMCLI("-lc", "1400", "700", "-d", fmt.Sprintf("%d", devIndex))

	if err1 != nil || err2 != nil {
		return NPUPowerAction{}, fmt.Errorf("reset failed (perf=%v, lock=%v)", err1, err2)
	}

	return NPUPowerAction{
		Success:   true,
		Message:   "Reset to PERFORMANCE (700–1400 MHz)",
		NewMode:   "performance",
		NewMaxMHz: 1400,
		NewMinMHz: 700,
	}, nil
}

// parseFloat parses "700.0 Mhz" / "3.53 W" etc. into float64.
func parseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	// Strip non-numeric suffixes like "Mhz", "mV", "W", "MB", "C", "%"
	re := regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)
	m := re.FindString(s)
	if m == "" {
		return 0
	}
	var f float64
	fmt.Sscanf(m, "%f", &f)
	return f
}

// ─── Smart Auto-Scheduler ───────────────────────────────────────────────────

const hmSchedulerVer = "v1.1.0"

var (
	schedMu       sync.RWMutex
	schedRunning  bool
	schedStopCh  chan struct{}
	schedState    NPOSchedulerState
	schedSettings NPOSchedulerSettings
)

// StartNPUScheduler starts the background scheduler for the given device.
// Settings are applied and the goroutine polls every <CheckSec> seconds.
func StartNPUScheduler(devIndex int, settings NPOSchedulerSettings) error {
	// Validate device
	if err := InitNPU(); err != nil {
		return fmt.Errorf("NPU not initialised: %v", err)
	}
	info, err := GetNPUDeviceInfo()
	if err != nil {
		return fmt.Errorf("cannot enumerate devices: %v", err)
	}
	if devIndex < 0 || devIndex >= int(info.NumDevices) {
		return fmt.Errorf("device index %d out of range (have %d device(s))", devIndex, info.NumDevices)
	}

	schedMu.Lock()
	if schedRunning {
		schedMu.Unlock()
		return fmt.Errorf("scheduler already running for device %d; stop it first", schedState.DevIndex)
	}
	schedRunning = true
	schedStopCh = make(chan struct{})
	schedSettings = settings
	// Fill defaults for any zero fields
	if schedSettings.TempWarnC == 0 {
		schedSettings.TempWarnC = 80.0
	}
	if schedSettings.TempCritC == 0 {
		schedSettings.TempCritC = 90.0
	}
	if schedSettings.UtilHighPct == 0 {
		schedSettings.UtilHighPct = 85.0
	}
	if schedSettings.UtilLowPct == 0 {
		schedSettings.UtilLowPct = 20.0
	}
	if schedSettings.CheckSec == 0 {
		schedSettings.CheckSec = 5
	}
	schedState = NPOSchedulerState{
		Running:  true,
		DevIndex: devIndex,
		Decision: "scheduler started",
	}
	schedMu.Unlock()

	go schedRunner(devIndex)
	return nil
}

// StopNPUScheduler stops the running scheduler.
func StopNPUScheduler() error {
	schedMu.Lock()
	defer schedMu.Unlock()
	if !schedRunning {
		return fmt.Errorf("no scheduler is running")
	}
	schedRunning = false
	close(schedStopCh)
	schedState.Running = false
	schedState.Decision = "stopped"
	return nil
}

// GetNPOSchedulerState returns the current scheduler state.
func GetNPOSchedulerState() (NPOSchedulerState, error) {
	schedMu.RLock()
	defer schedMu.RUnlock()
	return schedState, nil
}

// GetNPOSchedulerSettings returns the current scheduler settings.
func GetNPOSchedulerSettings() (NPOSchedulerSettings, error) {
	schedMu.RLock()
	defer schedMu.RUnlock()
	return schedSettings, nil
}

// schedRunner is the background polling loop.
func schedRunner(devIndex int) {
	for {
		select {
		case <-schedStopCh:
			return
		case <-time.After(time.Duration(schedSettings.CheckSec) * time.Second):
		}

		schedMu.RLock()
		if !schedRunning {
			schedMu.RUnlock()
			return
		}
		devIdx := schedState.DevIndex
		schedMu.RUnlock()

		// Collect current metrics directly from DLL (faster than spawning hm_smi)
		var curUtil, curTemp, curPower float32
		var curFreq uint64
		var curDVFS int32

		npuMutex.RLock()
		if npuFunc.getIPUUtiliRate != 0 {
			v, _, _ := syscall.Syscall(npuFunc.getIPUUtiliRate, 1, uintptr(devIdx), 0, 0)
			curUtil = float32(mathFloat(v))
		}
		if npuFunc.getTemperature != 0 {
			syscall.Syscall(npuFunc.getTemperature, 2, uintptr(devIdx), uintptr(unsafe.Pointer(&curTemp)), 0)
		}
		if npuFunc.getBoardPower != 0 {
			syscall.Syscall(npuFunc.getBoardPower, 2, uintptr(devIdx), uintptr(unsafe.Pointer(&curPower)), 0)
		}
		if npuFunc.getIPUFrequency != 0 {
			syscall.Syscall(npuFunc.getIPUFrequency, 2, uintptr(devIdx), uintptr(unsafe.Pointer(&curFreq)), 0)
		}
		if npuFunc.getDVFSMode != 0 {
			syscall.Syscall(npuFunc.getDVFSMode, 2, uintptr(devIdx), uintptr(unsafe.Pointer(&curDVFS)), 0)
		}
		npuMutex.RUnlock()

		utilPct := float64(curUtil) * 100.0
		tempC := float64(curTemp)
		powerW := float64(curPower)
		freqMHz := float64(curFreq) / 1e6

		// Smart logic: decide target DVFS mode
		targetMode := int(curDVFS) // default: keep current
		reason := ""

		if tempC >= schedSettings.TempCritC && tempC > 0 {
			targetMode = 1 // ONDEMAND
			reason = fmt.Sprintf("CRIT TEMP %.1fC → ONDEMAND (auto-throttle)", tempC)
		} else if tempC >= schedSettings.TempWarnC && tempC > 0 {
			targetMode = 1 // ONDEMAND
			reason = fmt.Sprintf("HIGH TEMP %.1fC → ONDEMAND", tempC)
		} else if utilPct >= schedSettings.UtilHighPct && utilPct >= 0 {
			targetMode = 0 // PERFORMANCE
			reason = fmt.Sprintf("HIGH LOAD %.1f%% → PERFORMANCE", utilPct)
		} else if utilPct < schedSettings.UtilLowPct && utilPct >= 0 {
			targetMode = 1 // ONDEMAND
			reason = fmt.Sprintf("LOW LOAD %.1f%% → ONDEMAND", utilPct)
		} else {
			reason = fmt.Sprintf("STABLE (util=%.0f%% temp=%.1fC) — no change", utilPct, tempC)
		}

		// Apply if mode changed
		if targetMode != int(curDVFS) {
			npuMutex.RLock()
			if npuFunc.setDVFSMode != 0 {
				syscall.Syscall(npuFunc.setDVFSMode, 2, uintptr(devIdx), uintptr(targetMode), 0)
			}
			npuMutex.RUnlock()
		}

		// Get clock lock info from hm_smi
		out, _ := runHMCLI("-a", "-d", fmt.Sprintf("%d", devIdx))
		lockMax := int(parseFloat(parseSMIField(out, "Lock_Ipu_Max_Clock")))
		lockMin := int(parseFloat(parseSMIField(out, "Lock_Ipu_Min_Clock")))

		modeName := []string{"PERFORMANCE", "ONDEMAND"}[targetMode]
		if targetMode < 0 || targetMode > 1 {
			modeName = fmt.Sprintf("MODE%d", targetMode)
		}

		// Update shared state
		schedMu.Lock()
		schedState.CurMode = modeName
		schedState.CurUtilPct = utilPct
		schedState.CurTempC = tempC
		schedState.CurPowerW = powerW
		schedState.CurFreqMHz = freqMHz
		schedState.CurLockMaxMHz = lockMax
		schedState.CurLockMinMHz = lockMin
		schedState.Decision = reason
		schedState.LastSwitch = time.Now().Format("15:04:05")
		schedMu.Unlock()
	}
}
