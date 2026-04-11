//go:build windows

package backend

import (
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"unsafe"
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

var (
	npuDLL     *syscall.LazyDLL
	npuFunc    npuFuncs
	npuOnce    sync.Once
	npuOnceErr error
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
func npuLoad() error {
	// Search paths (in order of priority, like C++ test.cpp)
	searchPaths := []string{
		filepath.Join(getExeDir(), "libhal_xh2a.dll"),
		filepath.Join(getExeDir(), "..", "..", "backend", "dynamic_npu", "libhal_xh2a.dll"),
		`C:\Users\3-64\source\repos\Project1\Project1\libhal_xh2a.dll`,
		`C:\Program Files (x86)\houmo-drv-xh2_v1.1.0\libhal_xh2a.dll`,
		`C:\Program Files (x86)\houmo-drv-xh2_v1.0.0\libhal_xh2a.dll`,
		`C:\Program Files\houmo-drv-xh2_v1.1.0\libhal_xh2a.dll`,
		`C:\Program Files\houmo-drv-xh2_v1.0.0\libhal_xh2a.dll`,
	}

	var dllPath string
	for _, p := range searchPaths {
		if _, err := os.Stat(p); err == nil {
			dllPath = p
			break
		}
	}

	if dllPath == "" {
		return fmt.Errorf("libhal_xh2a.dll not found in any search path; " +
			"ensure Houmo NPU driver is installed or copy DLL to app directory")
	}

	npuDLL = syscall.NewLazyDLL(dllPath)
	resolve := func(name string) uintptr {
		p := npuDLL.NewProc(name)
		if err := p.Find(); err != nil {
			return 0
		}
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
	if npuFunc.getDeviceInfo == 0 {
		return fmt.Errorf("libhal_xh2a.dll not found or hm_sys_get_device_info not exported; ensure Houmo HAL driver is installed")
	}
	return nil
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

// NPUDeviceMetrics holds real-time runtime metrics.
type NPUDeviceMetrics struct {
	IPUUtiliRate    float64 `json:"ipuUtiliRate"`
	IPUVoltageMV    float64 `json:"ipuVoltageMV"`
	IPUFrequencyHz  uint64  `json:"ipuFrequencyHz"`
	BoardPowerW     float64 `json:"boardPowerW"`
	TemperatureC    float64 `json:"temperatureC"`
	MemTotalMB      uint32  `json:"memTotalMB"`
	MemUsedMB       uint32  `json:"memUsedMB"`
	MemAvailMB      uint32  `json:"memAvailMB"`
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
	Index      int                 `json:"index"`
	Properties NPUDeviceProperties `json:"properties"`
	Metrics    NPUDeviceMetrics   `json:"metrics"`
	PCIeInfo   NPUPCIeInfo       `json:"pcieInfo"`
	DVFSMode   string             `json:"dvfsMode"`
	CTCPHYInfo NPUCTCPHYInfo     `json:"ctcPhyInfo"`
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
		props, _   := GetNPUDeviceProperties(int(devID))
		metrics, _ := GetNPUDeviceMetrics(int(devID))
		dvfsMode, _ := NPUGetDVFSMode(int(devID))
		ctcInfo, _  := GetNPUCTCPHYInfo(int(devID))
		report.Devices = append(report.Devices, NPUDeviceReport{
			Index: int(devID), Properties: props, Metrics: metrics,
			PCIeInfo: NPUPCIeInfo{}, DVFSMode: dvfsMode, CTCPHYInfo: ctcInfo,
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
	cmd := exec.Command(hmSMI, args...)
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
		// powerlimit requires -pl maxw minw; use 30W/5W as reasonable defaults
		out, err = runHMCLI("-g", "powerlimit", "-pl", "30w", "5w", "-d", fmt.Sprintf("%d", devIndex))
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
