//go:build windows

package backend

import (
	"fmt"
	"sync"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

// GPUStatusWatcher uses WaitForMultipleObjects to monitor multiple registry keys efficiently
type GPUStatusWatcher struct {
	mu           sync.RWMutex
	hKeys        []windows.Handle    // Registry key handles
	hEvents      []windows.Handle    // Event handles for each key
	hStopEvent   windows.Handle      // Stop signal event
	keyPaths     []string            // Registry paths being watched
	running      bool
	callbacks    []func(GPUPrefStatus)
}

var (
	gpuWatcher    *GPUStatusWatcher
	gpuStatusCache GPUPrefStatus
	gpuStatusMu   sync.RWMutex
)

// GPU status registry paths to monitor
var gpuRegistryPaths = []string{
	"SYSTEM\\CurrentControlSet\\Services\\LenovoProcessManagement\\Performance\\PowerSlider",
	"SOFTWARE\\Lenovo\\SmartEngine\\ModuleSettings\\GPU",
	"SOFTWARE\\Lenovo\\GameZone",
	"SOFTWARE\\Lenovo\\PowerManagement",
}

// NewGPUStatusWatcher creates a watcher that monitors all GPU status registry keys
// using a single WaitForMultipleObjects call for maximum efficiency
func NewGPUStatusWatcher() (*GPUStatusWatcher, error) {
	w := &GPUStatusWatcher{
		callbacks: make([]func(GPUPrefStatus), 0),
	}

	// Create stop event (manual-reset, initially not signaled)
	hStop, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create stop event: %w", err)
	}
	w.hStopEvent = hStop

	// Open all registry keys and create events for each
	for _, path := range gpuRegistryPaths {
		pathPtr, err := syscall.UTF16PtrFromString(path)
		if err != nil {
			continue
		}

		var hKey windows.Handle
		err = windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, pathPtr, 0, 
			windows.KEY_READ|windows.KEY_NOTIFY, &hKey)
		if err != nil {
			continue // Skip keys that don't exist
		}

		// Create event for this key
		hEvent, err := windows.CreateEvent(nil, 1, 0, nil)
		if err != nil {
			windows.RegCloseKey(hKey)
			continue
		}

		w.hKeys = append(w.hKeys, hKey)
		w.hEvents = append(w.hEvents, hEvent)
		w.keyPaths = append(w.keyPaths, path)
	}

	if len(w.hKeys) == 0 {
		windows.CloseHandle(hStop)
		return nil, fmt.Errorf("no valid registry keys to watch")
	}

	return w, nil
}

// OnChange registers a callback for GPU status changes
func (w *GPUStatusWatcher) OnChange(callback func(GPUPrefStatus)) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.callbacks = append(w.callbacks, callback)
}

// Watch starts monitoring all registry keys using WaitForMultipleObjects
// This is a single-threaded, event-driven approach - NO POLLING
func (w *GPUStatusWatcher) Watch() {
	w.mu.Lock()
	w.running = true
	w.mu.Unlock()

	defer w.cleanup()

	numKeys := len(w.hKeys)
	
	// Build handle array: [stopEvent, event1, event2, ...]
	allHandles := make([]windows.Handle, numKeys+1)
	allHandles[0] = w.hStopEvent
	for i := 0; i < numKeys; i++ {
		allHandles[i+1] = w.hEvents[i]
	}

	for {
		// Register for change notification on ALL keys
		for i := 0; i < numKeys; i++ {
			windows.ResetEvent(w.hEvents[i]) // Reset before registering
			err := windows.RegNotifyChangeKeyValue(
				w.hKeys[i],
				true, // watch subtree
				windows.REG_NOTIFY_CHANGE_LAST_SET|windows.REG_NOTIFY_CHANGE_NAME,
				w.hEvents[i],
				true, // async
			)
			if err != nil {
				// Key might have been deleted, try to re-open
				w.reopenKey(i)
			}
		}

		// Wait for ANY event to signal (including stop)
		// WaitForMultipleObjects waits on all handles efficiently
		result, err := windows.WaitForMultipleObjects(
			allHandles,
			false, // wait for ANY event, not all
			windows.INFINITE,
		)

		if err != nil {
			// Error occurred, wait a bit and retry
			result, _ = windows.WaitForMultipleObjects(
				allHandles,
				false,
				1000, // 1 second timeout
			)
		}

		if result == windows.WAIT_FAILED {
			continue
		}

		// Check if stop event was signaled (index 0)
		if result == windows.WAIT_OBJECT_0 {
			// Stop requested
			return
		}

		// One of the registry keys changed
		// result - WAIT_OBJECT_0 gives the index of signaled handle
		signaledIndex := int(result - windows.WAIT_OBJECT_0)
		if signaledIndex > 0 && signaledIndex <= numKeys {
			// Registry key at index (signaledIndex-1) changed
			// Read new status and notify callbacks
			w.notifyChange()
		}
	}
}

// reopenKey attempts to reopen a registry key if it was closed/invalidated
func (w *GPUStatusWatcher) reopenKey(index int) {
	if index >= len(w.hKeys) || index >= len(w.keyPaths) {
		return
	}

	// Close old handles
	if w.hKeys[index] != 0 {
		windows.RegCloseKey(w.hKeys[index])
	}

	// Reopen
	pathPtr, err := syscall.UTF16PtrFromString(w.keyPaths[index])
	if err != nil {
		return
	}

	var hKey windows.Handle
	err = windows.RegOpenKeyEx(windows.HKEY_LOCAL_MACHINE, pathPtr, 0,
		windows.KEY_READ|windows.KEY_NOTIFY, &hKey)
	if err == nil {
		w.hKeys[index] = hKey
	}
}

// notifyChange reads current GPU status and notifies all callbacks
func (w *GPUStatusWatcher) notifyChange() {
	status := readGPUStatusDirect()

	// Update cache
	gpuStatusMu.Lock()
	gpuStatusCache = status
	gpuStatusMu.Unlock()

	// Notify callbacks (non-blocking)
	w.mu.RLock()
	callbacks := make([]func(GPUPrefStatus), len(w.callbacks))
	copy(callbacks, w.callbacks)
	w.mu.RUnlock()

	for _, cb := range callbacks {
		if cb != nil {
			go cb(status)
		}
	}
}

// Stop signals the watcher to stop and cleans up resources
func (w *GPUStatusWatcher) Stop() {
	w.mu.Lock()
	if !w.running {
		w.mu.Unlock()
		return
	}
	w.running = false
	w.mu.Unlock()

	// Signal stop event to wake up WaitForMultipleObjects
	if w.hStopEvent != 0 {
		windows.SetEvent(w.hStopEvent)
	}
}

// cleanup releases all resources
func (w *GPUStatusWatcher) cleanup() {
	// Close all registry keys
	for _, hKey := range w.hKeys {
		if hKey != 0 {
			windows.RegCloseKey(hKey)
		}
	}
	w.hKeys = nil

	// Close all events
	for _, hEvent := range w.hEvents {
		if hEvent != 0 {
			windows.CloseHandle(hEvent)
		}
	}
	w.hEvents = nil

	// Close stop event
	if w.hStopEvent != 0 {
		windows.CloseHandle(w.hStopEvent)
		w.hStopEvent = 0
	}
}

// StartGPUStatusWatcher starts the GPU status registry watcher
func StartGPUStatusWatcher() error {
	if gpuWatcher != nil {
		return nil
	}

	watcher, err := NewGPUStatusWatcher()
	if err != nil {
		return fmt.Errorf("failed to create GPU status watcher: %w", err)
	}

	gpuWatcher = watcher

	// Initial read
	gpuStatusCache = readGPUStatusDirect()

	// Start watching in background
	go watcher.Watch()

	return nil
}

// StopGPUStatusWatcher stops the GPU status watcher
func StopGPUStatusWatcher() {
	if gpuWatcher != nil {
		gpuWatcher.Stop()
		gpuWatcher = nil
	}
}

// GetGPUStatusCached returns cached GPU status (instant, no registry read)
func GetGPUStatusCached() GPUPrefStatus {
	gpuStatusMu.RLock()
	defer gpuStatusMu.RUnlock()
	return gpuStatusCache
}

// OnGPUStatusChange registers a callback for GPU status changes
func OnGPUStatusChange(callback func(GPUPrefStatus)) {
	if gpuWatcher == nil {
		return
	}
	gpuWatcher.mu.Lock()
	defer gpuWatcher.mu.Unlock()
	gpuWatcher.callbacks = append(gpuWatcher.callbacks, callback)
}

// RemoveGPUStatusCallbacks removes all GPU status callbacks
func RemoveGPUStatusCallbacks() {
	if gpuWatcher == nil {
		return
	}
	gpuWatcher.mu.Lock()
	defer gpuWatcher.mu.Unlock()
	gpuWatcher.callbacks = nil
}

// readGPUStatusDirect reads GPU status directly from registry
func readGPUStatusDirect() GPUPrefStatus {
	result := GPUPrefStatus{
		Available: false,
	}

	// First check if Dispatcher service is running
	if !isDispatcherServiceRunning() {
		result.Available = false
		result.Label = "Dispatcher Service Stopped"
		result.Value = 0
		result.PCMStatus = 0
		result.PCMStatusAvail = false
		result.PCMLabel = "N/A"
		return result
	}

	// Read iGPUStatus from SmartEngine (PCM_GPUStatus)
	pcmStatus, pcmAvail := readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE,
		"SOFTWARE\\Lenovo\\SmartEngine\\ModuleSettings\\GPU", "iGPUStatus")

	// Read PE_GPUPrefStatus from PowerSlider
	peStatus, peAvail := readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE,
		"SYSTEM\\CurrentControlSet\\Services\\LenovoProcessManagement\\Performance\\PowerSlider", "PE_GPUPrefStatus")

	// Fallback to GameZone and PowerManagement
	if !pcmAvail {
		pcmStatus, pcmAvail = readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE,
			"SOFTWARE\\Lenovo\\GameZone", "ITS_GPUHybridModeSetting")
	}
	if !pcmAvail {
		pcmStatus, pcmAvail = readDWordFromRegistry(windows.HKEY_LOCAL_MACHINE,
			"SOFTWARE\\Lenovo\\PowerManagement", "ITS_GPUHybridModeSetting")
	}

	// Determine final label - check PE_GPUPrefStatus first
	// PE_GPUPrefStatus not existing (=0) means no Dispatcher service controlling GPU
	// Even if PE_GPUPrefStatus is missing from registry, treat as Dispatcher not active
	if !pcmAvail && !peAvail {
		result.Available = false
		result.Label = "Not Available"
		result.Value = 0
	} else if !peAvail || peStatus == 0 {
		// PE_GPUPrefStatus missing or = 0: Dispatcher not controlling GPU
		result.Available = false
		result.Label = "Dispatcher Service Stopped"
		result.Value = 0
		result.PCMStatus = pcmStatus
		result.PCMStatusAvail = pcmAvail
		result.PCMLabel = pcmStatusLabel(pcmStatus)
		return result
	} else if !pcmAvail {
		// PCM_GPUStatus N/A (SmartEngine not installed), but Dispatcher is running.
		// Fall back to PE_GPUPrefStatus: 1=DIS, 2=UMA
		if peAvail && peStatus == 1 {
			result.Label = "DIS (Hybrid)"
			result.Value = 1
			result.Available = true
		} else if peAvail && peStatus == 2 {
			result.Label = "UMA (IGPU)"
			result.Value = 2
			result.Available = true
		} else {
			// peStatus missing or 0: Dispatcher not controlling GPU
			result.Label = "Dispatcher Service Stopped"
			result.Value = 0
			result.Available = false
		}
	} else {
		if pcmAvail && pcmStatus == 1 {
			result.Label = "UMA (IGPU)"
			result.Value = 2
			result.Available = true
		} else if pcmAvail && pcmStatus == 3 {
			result.Label = "DIS (Hybrid)"
			result.Value = 1
			result.Available = true
		} else if pcmAvail && pcmStatus == 2 {
			if peAvail && peStatus == 1 {
				result.Label = "DIS (Hybrid)"
				result.Value = 1
				result.Available = true
			} else if peAvail && peStatus == 2 {
				result.Label = "UMA (IGPU)"
				result.Value = 2
				result.Available = true
			} else {
				// peStatus missing or 0: Dispatcher not controlling GPU
				result.Label = "Dispatcher Service Stopped"
				result.Value = 0
				result.Available = false
			}
		} else if peAvail {
			result.Value = peStatus
			result.Label = gpuPrefStatusLabel(peStatus)
			result.Available = true
		} else {
			result.Label = "Not Available"
			result.Value = 0
		}
	}

	result.PCMStatus = pcmStatus
	result.PCMStatusAvail = pcmAvail
	result.PCMLabel = pcmStatusLabel(pcmStatus)

	return result
}

// isDispatcherServiceRunning checks if LenovoProcessManagement service is running
func isDispatcherServiceRunning() bool {
	// Use syscall to check service status directly (faster than PowerShell)
	scm, err := windows.OpenSCManager(nil, nil, windows.SC_MANAGER_CONNECT)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(scm)

	serviceName, _ := windows.UTF16PtrFromString("LenovoProcessManagement")
	hService, err := windows.OpenService(scm, serviceName, windows.SERVICE_QUERY_STATUS)
	if err != nil {
		return false
	}
	defer windows.CloseServiceHandle(hService)

	var status windows.SERVICE_STATUS
	err = windows.QueryServiceStatus(hService, &status)
	if err != nil {
		return false
	}

	return status.CurrentState == windows.SERVICE_RUNNING
}

// readDWordFromRegistry reads a DWORD value from registry
func readDWordFromRegistry(hKeyRoot windows.Handle, subKeyPath, valueName string) (uint32, bool) {
	subKeyPtr, err := syscall.UTF16PtrFromString(subKeyPath)
	if err != nil {
		return 0, false
	}

	var hKey windows.Handle
	err = windows.RegOpenKeyEx(hKeyRoot, subKeyPtr, 0, windows.KEY_READ, &hKey)
	if err != nil {
		return 0, false
	}
	defer windows.RegCloseKey(hKey)

	valueNamePtr, err := syscall.UTF16PtrFromString(valueName)
	if err != nil {
		return 0, false
	}

	var valueType uint32
	var valueData uint32
	var valueSize uint32 = 4

	err = windows.RegQueryValueEx(hKey, valueNamePtr, nil, &valueType, 
		(*byte)(unsafe.Pointer(&valueData)), &valueSize)
	if err != nil || valueType != windows.REG_DWORD {
		return 0, false
	}

	return valueData, true
}

// GetGPUPrefStatusFromCache returns cached GPU status (for frontend)
func GetGPUPrefStatusFromCache() GPUPrefStatus {
	return GetGPUStatusCached()
}

// GetCachedGPUStatus returns raw cached values
func GetCachedGPUStatus() (iGPUStatus uint32, iAvail bool, peStatus uint32, peAvail bool) {
	gpuStatusMu.RLock()
	cache := gpuStatusCache
	gpuStatusMu.RUnlock()
	return cache.PCMStatus, cache.PCMStatusAvail, cache.Value, cache.Available
}