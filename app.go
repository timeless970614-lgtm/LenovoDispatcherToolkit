package main

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/wailsapp/wails/v2/pkg/runtime"
	"lenovo-toolkit/backend"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ============ System Info ============

// GetSystemInfo returns system information
func (a *App) GetSystemInfo() backend.SystemInfo {
	info, err := backend.GetSystemInfo()
	if err != nil {
		return backend.SystemInfo{
			CPUName:       "Error: " + err.Error(),
			BIOSVersion:   "N/A",
			OSCaption:     "N/A",
			OSVersion:     "N/A",
			TotalMemoryGB: 0,
		}
	}
	return info
}

// ============ Dispatcher Info ============

// GetDispatcherInfo returns Dispatcher driver and registry info
func (a *App) GetDispatcherInfo() backend.DispatcherInfo {
	info, err := backend.GetDispatcherInfo()
	if err != nil {
		return backend.DispatcherInfo{
			DriverVersion: "N/A",
			Description:   "Driver not found",
			CurrentMode:   "Unknown",
			AIEngineMode:  "Unknown",
			AutoMode:      0,
		}
	}
	return info
}

// ============ Service Control ============

// GetServiceStatus returns the current status of LenovoProcessManagement service
func (a *App) GetServiceStatus() string {
	status, err := backend.GetServiceStatus()
	if err != nil {
		return "Error: " + err.Error()
	}
	return status
}

// StartService starts the LenovoProcessManagement service
func (a *App) StartService() string {
	err := backend.StartService()
	if err != nil {
		return "Error: " + err.Error()
	}
	return "Service started successfully"
}

// StopService stops the LenovoProcessManagement service
func (a *App) StopService() string {
	err := backend.StopService()
	if err != nil {
		return "Error: " + err.Error()
	}
	return "Service stopped successfully"
}

// RestartService restarts the LenovoProcessManagement service
func (a *App) RestartService() string {
	err := backend.RestartService()
	if err != nil {
		return "Error: " + err.Error()
	}
	return "Service restarted successfully"
}

// ============ PPM Settings ============

// GetPPMSettings reads all current PPM power settings
func (a *App) GetPPMSettings() backend.PPMSettings {
	return backend.GetPPMSettings()
}

// SetPowerSettingRaw sets a power setting by GUID with raw AC/DC values
func (a *App) SetPowerSettingRaw(settingGUID string, acValue, dcValue int) string {
	err := backend.SetPowerSettingRaw(settingGUID, acValue, dcValue)
	if err != nil {
		return "Error: " + err.Error()
	}
	return "OK"
}

// ApplyHetero applies Hetero scheduling settings
func (a *App) ApplyHetero(increase, decrease int) string {
	if increase < 0 || increase > 100 || decrease < 0 || decrease > 100 {
		return fmt.Sprintf("Error: values must be between 0 and 100")
	}
	err := backend.ApplyHetero(increase, decrease)
	if err != nil {
		return "Error: " + err.Error()
	}
	return fmt.Sprintf("Hetero settings applied: Increase=%d, Decrease=%d", increase, decrease)
}

// ApplyEPP applies EPP settings
func (a *App) ApplyEPP(epp, epp1 int) string {
	if epp < 0 || epp > 100 || epp1 < 0 || epp1 > 100 {
		return "Error: values must be between 0 and 100"
	}
	err := backend.ApplyEPP(epp, epp1)
	if err != nil {
		return "Error: " + err.Error()
	}
	return fmt.Sprintf("EPP settings applied: EPP=%d, EPP1=%d", epp, epp1)
}

// ApplyMaxFrequency applies max frequency settings
func (a *App) ApplyMaxFrequency(freq, freq1 int) string {
	if freq < 0 || freq > 100 || freq1 < 0 || freq1 > 100 {
		return "Error: values must be between 0 and 100"
	}
	err := backend.ApplyMaxFrequency(freq, freq1)
	if err != nil {
		return "Error: " + err.Error()
	}
	return fmt.Sprintf("Max frequency settings applied: Freq=%d, Freq1=%d", freq, freq1)
}

// ApplySoftParkLatency applies SoftParkLatency settings
func (a *App) ApplySoftParkLatency(ac, dc int) string {
	if ac < 0 || ac > 100 || dc < 0 || dc > 100 {
		return "Error: values must be between 0 and 100"
	}
	err := backend.ApplySoftParkLatency(ac, dc)
	if err != nil {
		return "Error: " + err.Error()
	}
	return fmt.Sprintf("SoftParkLatency settings applied: AC=%d, DC=%d", ac, dc)
}

// RestoreDefaults restores default power settings
func (a *App) RestoreDefaults() string {
	err := backend.RestoreDefaults()
	if err != nil {
		return "Error: " + err.Error()
	}
	return "Default settings restored successfully"
}

// ============ DYTC Functions ============

// GetDYTCInfo returns all DYTC related information
func (a *App) GetDYTCInfo() backend.DYTCInfo {
	info, err := backend.GetDYTCInfo()
	if err != nil {
		return backend.DYTCInfo{
			CurrentMode:           "Error: " + err.Error(),
			CurrentDispatcherMode: "N/A",
			AIEngineMode:          "N/A",
		}
	}
	return *info
}

// SetDYTCMode sets the DYTC thermal mode (BSM, STD, APM, AQM, EPM, IEPM, DCC)
func (a *App) SetDYTCMode(modeName string) string {
	result, err := backend.SetDYTCModeByName(modeName)
	if err != nil {
		return "Error: " + err.Error()
	}
	return result
}

// SetGEEKMode enables or disables GEEK mode
func (a *App) SetGEEKMode(enable bool) string {
	err := backend.SetGEEKMode(enable)
	if err != nil {
		return "Error: " + err.Error()
	}
	if enable {
		return "GEEK Mode enabled"
	}
	return "GEEK Mode disabled"
}

// SetODV sets ODV (OverDrive Voltage) mode
func (a *App) SetODV(index, value uint32) string {
	err := backend.SetODV(index, value)
	if err != nil {
		return "Error: " + err.Error()
	}
	return fmt.Sprintf("ODV set: Index=%d, Value=%d", index, value)
}

// CheckDYTCCapabilities returns device DYTC capabilities
func (a *App) CheckDYTCCapabilities() map[string]bool {
	return map[string]bool{
		"BSM":  backend.CheckBSM(),
		"GEEK": backend.CheckGEEK(),
		"DCC":  backend.CheckDCC(),
	}
}

// ============ Function Check (GPU & System Diagnostics) ============

// EnumerateGPUs returns a list of all GPUs using WMI
func (a *App) EnumerateGPUs() []backend.GPUInfo {
	return backend.EnumerateGPUs()
}

// EnumerateGPUProcesses returns a list of processes that might be using GPU
func (a *App) EnumerateGPUProcesses() []backend.GPUProcess {
	return backend.EnumerateGPUProcesses()
}

// GetIGPUMode returns the current IGPU mode status from WMI
func (a *App) GetIGPUMode() backend.IGPUStatus {
	available, mode := backend.GetIGPUModeStatusWMI()
	return backend.IGPUStatus{
		Available: available,
		Mode:      mode,
	}
}

// GetGPUPrefStatus reads PE_GPUPrefStatus registry value in real-time
func (a *App) GetGPUPrefStatus() backend.GPUPrefStatus {
	return backend.GetGPUPrefStatus()
}

// Intel GPU Frequency Control
func (a *App) GetIntelGPUFrequency() backend.IntelGPUFrequency {
	return backend.GetIntelGPUFrequency()
}

// SetIntelGPUFrequencyRange sets iGPU frequency range via IGC API (float64 MHz).
func (a *App) SetIntelGPUFrequencyRange(minFreq, maxFreq float64) backend.IntelGPUFreqTestResult {
	return backend.SetIntelGPUFrequencyRange(minFreq, maxFreq)
}

func (a *App) TestIntelGPUFrequency(testType string) backend.IntelGPUFreqTestResult {
	return backend.TestIntelGPUFrequency(testType)
}

func (a *App) GetIntelDriverDownloadURL() string {
	return backend.GetIntelDriverDownloadURL()
}

// GetIntelGPUUtilization returns current GPU 3D engine utilization (0-100%).
// Lightweight call for periodic polling from the frontend.
func (a *App) GetIntelGPUUtilization() float64 {
	return backend.GetIntelGPUUtilization()
}

func (a *App) StartGPUStatusWatcher() error {
	err := backend.StartGPUStatusWatcher()
	if err != nil {
		return err
	}

	// Register callback to push events to frontend when GPU status changes
	backend.OnGPUStatusChange(func(status backend.GPUPrefStatus) {
		if a.ctx != nil {
			runtime.EventsEmit(a.ctx, "gpu:status-change", status)
		}
	})

	return nil
}

func (a *App) StopGPUStatusWatcher() {
	backend.RemoveGPUStatusCallbacks()
	backend.StopGPUStatusWatcher()
}

func (a *App) GetGPUPrefStatusFromCache() backend.GPUPrefStatus {
	return backend.GetGPUPrefStatusFromCache()
}

func (a *App) GetCachedGPUStatus() (uint32, bool, uint32, bool) {
	return backend.GetCachedGPUStatus()
}

// SetIGPUMode sets the IGPU mode (0=DGPU Plug In, 1=DGPU Plug Out)
func (a *App) SetIGPUMode(mode uint32) backend.SetResult {
	success, returnedMode := backend.SetIGPUModeStatusWMI(mode)
	if success {
		return backend.SetResult{
			Success: true,
			Message: fmt.Sprintf("IGPU mode set successfully. Mode=%d", returnedMode),
		}
	}
	return backend.SetResult{
		Success: false,
		Message: fmt.Sprintf("Failed to set IGPU mode. Error code: %d", returnedMode),
	}
}

// CheckNVIDIAStatus checks if NVIDIA GPU is present and its status
func (a *App) CheckNVIDIAStatus() backend.NVIDIAStatus {
	return backend.CheckNVIDIAStatus()
}

// ============ Mode Check ============

// GetModeCheckInfo returns all mode check information
func (a *App) GetModeCheckInfo() backend.ModeCheckInfo {
	return backend.GetModeCheckInfo()
}

// ============ Pin Mode ============

// GetPinnedDYTCMode returns the currently pinned mode name, or "" if not pinned
func (a *App) GetPinnedDYTCMode() string {
	return backend.GetPinnedDYTCMode()
}

// PinDYTCMode pins the given mode (writes Policy_Override=3 + ITS_AutomaticModeSetting)
func (a *App) PinDYTCMode(modeId string) error {
	return backend.PinDYTCMode(modeId)
}

// UnpinDYTCMode removes the pin (restores Policy_Override=0)
func (a *App) UnpinDYTCMode() error {
	return backend.UnpinDYTCMode()
}

// ============ SSD Control ============

// GetSSDInfo returns all physical SSD drives and their mode status
func (a *App) GetSSDInfo() []backend.SSDInfo {
	return backend.GetSSDInfo()
}

// SetSSDMode sets the SSD mode (0=Standard, 1=Performance, 2=PowerSaving, 3=Default)
func (a *App) SetSSDMode(physicalDriveIndex int, mode int) backend.SSDModeResult {
	return backend.SetSSDMode(physicalDriveIndex, backend.SSDMode(mode))
}

// ============ ML Scenario Log Capture ============

// StartMLScenarioCapture starts capturing ML Scenario events from the named pipe
func (a *App) StartMLScenarioCapture() backend.MLLogStatus {
	return backend.StartMLScenarioCapture()
}

// StopMLScenarioCapture stops the capture and saves the log file
func (a *App) StopMLScenarioCapture() backend.MLLogStatus {
	return backend.StopMLScenarioCapture()
}

// GetMLLogStatus returns the current capture session status
func (a *App) GetMLLogStatus() backend.MLLogStatus {
	return backend.GetMLLogStatus()
}

// OpenFolder opens a folder in Windows Explorer
func (a *App) OpenFolder(path string) error {
	cmd := exec.Command("explorer.exe", path)
	return cmd.Start()
}

// ============ AI Analysis ============

// GetLogFiles returns the list of dispatcher log files
func (a *App) GetLogFiles() []backend.LogFileInfo {
	return backend.GetLogFiles()
}

// ReadLogTail reads the last N lines from the latest log file
func (a *App) ReadLogTail(maxLines int) string {
	return backend.ReadLogTail(maxLines)
}

// GetLogSummary returns structured log summary for AI analysis
func (a *App) GetLogSummary() map[string]interface{} {
	return backend.GetLogSummary()
}

// ============ Uninstall ============

// UninstallDispatcher uninstalls the Lenovo Process Management driver
func (a *App) UninstallDispatcher() backend.UninstallResult {
	return backend.UninstallDispatcherSimple()
}

// ============ Dynamic Log ============

// GetDynamicLogStatus checks if dynamic log is enabled
func (a *App) GetDynamicLogStatus() bool {
	return backend.GetDynamicLogStatus()
}

// EnableDynamicLog enables the dynamic log and restarts the service
func (a *App) EnableDynamicLog() backend.DynamicLogResult {
	return backend.EnableDynamicLog()
}

// ============ Auto Gear & EPOT ============

// GetEPOTStatus returns ML_Scenario EPOT parameters
func (a *App) GetEPOTStatus() backend.EPOTStatus {
	return backend.GetEPOTStatus()
}

// GetGPUAutoGear returns the current Auto Gear setting
func (a *App) GetGPUAutoGear() backend.GPUAutoGear {
	return backend.GetGPUAutoGear()
}

// SetGPUAutoGear sets the Auto Gear value
func (a *App) SetGPUAutoGear(value uint32) backend.SetResult {
	return backend.SetGPUAutoGear(value)
}

// UninstallDTT runs the DTT uninstall script
func (a *App) UninstallDTT() string {
	return backend.UninstallDTT()
}

// UninstallDTTUI runs the DTT UI uninstall script
func (a *App) UninstallDTTUI() string {
	return backend.UninstallDTTUI()
}

// ============ Dynamic NPU (Houmo AI M50 NPU HAL) ============

// GetNPUFullReport returns a full NPU status report for all detected Houmo NPU devices.
// This consolidates device enumeration, properties, metrics, DVFS mode, and CTC PHY info.
func (a *App) GetNPUFullReport() backend.NPUFullReport {
	report, err := backend.GetNPUFullReport()
	if err != nil {
		return backend.NPUFullReport{}
	}
	return report
}

// GetNPUDeviceInfo enumerates all Houmo NPU devices and returns their IDs.
func (a *App) GetNPUDeviceInfo() backend.NPUDeviceInfo {
	info, err := backend.GetNPUDeviceInfo()
	if err != nil {
		return backend.NPUDeviceInfo{}
	}
	return info
}

// GetNPUDeviceProperties reads static properties for a given device index.
func (a *App) GetNPUDeviceProperties(devIndex int) backend.NPUDeviceProperties {
	prop, err := backend.GetNPUDeviceProperties(devIndex)
	if err != nil {
		return backend.NPUDeviceProperties{}
	}
	return prop
}

// GetNPUDeviceMetrics reads real-time runtime metrics for a given device index.
func (a *App) GetNPUDeviceMetrics(devIndex int) backend.NPUDeviceMetrics {
	m, err := backend.GetNPUDeviceMetrics(devIndex)
	if err != nil {
		return backend.NPUDeviceMetrics{}
	}
	return m
}

// NPUGetDVFSMode reads the current DVFS mode for a device (PERFORMANCE or ONDEMAND).
func (a *App) NPUGetDVFSMode(devIndex int) string {
	mode, err := backend.NPUGetDVFSMode(devIndex)
	if err != nil {
		return "Error: " + err.Error()
	}
	return mode
}

// NPUSetDVFSMode sets the DVFS mode for a device.
// Valid modes: "PERFORMANCE", "ONDEMAND".
func (a *App) NPUSetDVFSMode(devIndex int, mode string) string {
	err := backend.NPUSetDVFSMode(devIndex, mode)
	if err != nil {
		return "Error: " + err.Error()
	}
	return "DVFS mode set to " + mode + " successfully"
}

// GetNPUSDKInfo reads Houmo HAL SDK and driver version info.
func (a *App) GetNPUSDKInfo() backend.NPUSDKInfo {
	info, err := backend.GetNPUSDKInfo()
	if err != nil {
		return backend.NPUSDKInfo{}
	}
	return info
}

func (a *App) GetNPUPowerStatus(devIndex int) backend.NPUPowerStatus {
	status, err := backend.GetNPUPowerStatus(devIndex)
	if err != nil {
		return backend.NPUPowerStatus{}
	}
	return status
}

func (a *App) SetNPUMode(devIndex int, mode string) backend.NPUPowerAction {
	result, err := backend.SetNPUMode(devIndex, mode)
	if err != nil {
		return backend.NPUPowerAction{Success: false, Message: err.Error()}
	}
	return result
}

func (a *App) SetNPUClockLock(devIndex, maxMHz, minMHz int) backend.NPUPowerAction {
	result, err := backend.SetNPUClockLock(devIndex, maxMHz, minMHz)
	if err != nil {
		return backend.NPUPowerAction{Success: false, Message: err.Error()}
	}
	return result
}

func (a *App) ResetNPUDefaults(devIndex int) backend.NPUPowerAction {
	result, err := backend.ResetNPUDefaults(devIndex)
	if err != nil {
		return backend.NPUPowerAction{Success: false, Message: err.Error()}
	}
	return result
}

// GetNPUREport returns a detailed diagnostic report of NPU DLL loading,
// function resolution, and device enumeration. Use this when the UI shows
// N/A or 0 devices to understand exactly what the backend encountered.
func (a *App) GetNPUREport() string {
	return backend.GetNPUREport()
}

// GetNPUDeviceInfoWrapper returns device enumeration result and exposes
// the detailed init error if any.
func (a *App) GetNPUDeviceInfoWrapper() (backend.NPUDeviceInfo, string) {
	info, err := backend.GetNPUDeviceInfo()
	if err != nil {
		return backend.NPUDeviceInfo{}, err.Error()
	}
	return info, ""
}

func (a *App) GetNPUDeviceOverview(devIndex int) backend.NPUDeviceOverview {
	overview, err := backend.GetNPUDeviceOverview(devIndex)
	if err != nil {
		return backend.NPUDeviceOverview{}
	}
	return overview
}

// ============ ETL Trace Analyzer ============

// IsElevated returns whether the current process has administrator privileges
func (a *App) IsElevated() bool {
	return backend.IsElevated()
}

// GetETLProfiles returns available WPR profile options for the UI
func (a *App) GetETLProfiles() []backend.ETLProfile {
	return backend.GetETLProfiles()
}

// StartETLCapture starts a WPR trace with the given profile ID
// durationSecs: capture duration in seconds (0 = indefinite, caller must call StopETLCapture)
func (a *App) StartETLCapture(profile string, durationSecs int) backend.ETLCaptureState {
	return backend.StartETLCapture(profile, durationSecs)
}

// StopETLCapture stops the running WPR trace and returns trace info
func (a *App) StopETLCapture() backend.ETLTraceInfo {
	return backend.StopETLCapture()
}

// GetETLCaptureStatus returns the current capture state
func (a *App) GetETLCaptureStatus() backend.ETLCaptureState {
	return backend.GetETLCaptureStatus()
}

// GetETLTraceList returns list of previously captured traces
func (a *App) GetETLTraceList() []backend.ETLTraceInfo {
	return backend.GetETLTraceList()
}

// AnalyzeETLFile parses an ETL file and returns structured analysis results
func (a *App) AnalyzeETLFile(etlPath string) backend.ETLAnalysisResult {
	return backend.AnalyzeETLFile(etlPath)
}