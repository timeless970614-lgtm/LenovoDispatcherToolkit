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

func (a *App) SetIntelGPUFrequencyRange(minFreq, maxFreq uint32) backend.IntelGPUFreqTestResult {
	return backend.SetIntelGPUFrequencyRange(minFreq, maxFreq)
}

func (a *App) TestIntelGPUFrequency(testType string) backend.IntelGPUFreqTestResult {
	return backend.TestIntelGPUFrequency(testType)
}

func (a *App) CtlFrequencySetRange(adapterIndex, minFreq, maxFreq uint32) backend.IntelGPUFreqTestResult {
	return backend.CtlFrequencySetRange(adapterIndex, minFreq, maxFreq)
}

func (a *App) GetIntelDriverDownloadURL() string {
	return backend.GetIntelDriverDownloadURL()
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
