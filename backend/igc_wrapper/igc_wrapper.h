// igc_wrapper.h
// Wrapper for Intel GPU Control Library (igc.dll / ControlLib.dll)
// Exposes a flat C API for Go cgo / syscall binding.
//
// Intel GPU Control Library real API:
//   ctlInit()                          -> ctl_api_handle_t
//   ctlEnumerateDevices()              -> ctl_device_adapter_handle_t[]
//   ctlEnumFrequencyDomains()          -> ctl_freq_handle_t[]
//   ctlFrequencyGetRange()             -> ctl_freq_range_t {min, max} MHz
//   ctlFrequencySetRange()             -> set {min, max} MHz
//   ctlClose()
//
// This wrapper loads igc.dll at runtime (delay-load style) so the
// application starts even when the DLL is absent.

#pragma once
#ifdef __cplusplus
extern "C" {
#endif

#include <windows.h>

// ── Result codes ──────────────────────────────────────────────────────────────
#define IGC_OK                  0
#define IGC_ERR_NOT_LOADED     -1   // igc.dll not found / failed to load
#define IGC_ERR_NO_DEVICE      -2   // no Intel GPU adapter found
#define IGC_ERR_NO_FREQ_DOMAIN -3   // no frequency domain on adapter
#define IGC_ERR_API_FAIL       -4   // ctlXxx() returned non-zero
#define IGC_ERR_NOT_INIT       -5   // IGC_Init() not called yet

// ── Opaque handle (index into internal table) ─────────────────────────────────
typedef int IGC_HANDLE;   // 0-based adapter index; -1 = invalid

// ── Frequency info ────────────────────────────────────────────────────────────
typedef struct {
    double minFreqMHz;      // hardware minimum (read-only capability)
    double maxFreqMHz;      // hardware maximum (read-only capability)
    double currentMinMHz;   // current software min limit
    double currentMaxMHz;   // current software max limit
    double requestedMHz;    // last requested frequency
    double tdpMHz;          // TDP frequency
    double efficientMHz;    // efficient frequency
    double actualMHz;       // actual current frequency
} IGC_FreqInfo;

// ── Adapter info ──────────────────────────────────────────────────────────────
typedef struct {
    char   name[256];       // adapter name (UTF-8)
    char   driverVersion[64];
    int    adapterIndex;
    int    freqDomainCount;
} IGC_AdapterInfo;

// ── API ───────────────────────────────────────────────────────────────────────

// Initialize the IGC library. Must be called once before other functions.
// Returns IGC_OK or IGC_ERR_NOT_LOADED.
__declspec(dllexport) int IGC_Init(void);

// Release all resources. Safe to call even if Init failed.
__declspec(dllexport) void IGC_Close(void);

// Returns the number of Intel GPU adapters found (0 if none / not init).
__declspec(dllexport) int IGC_GetAdapterCount(void);

// Fill *info for adapter at adapterIndex. Returns IGC_OK or error code.
__declspec(dllexport) int IGC_GetAdapterInfo(int adapterIndex, IGC_AdapterInfo* info);

// Get frequency range for the first GPU frequency domain of adapterIndex.
// Returns IGC_OK or error code.
__declspec(dllexport) int IGC_GetFreqInfo(int adapterIndex, IGC_FreqInfo* info);

// Set frequency range [minMHz, maxMHz] for adapterIndex.
// Pass 0 for min/max to use hardware defaults.
// Returns IGC_OK or error code.
__declspec(dllexport) int IGC_SetFreqRange(int adapterIndex, double minMHz, double maxMHz);

// Human-readable error string for a result code.
__declspec(dllexport) const char* IGC_ErrorString(int code);

#ifdef __cplusplus
}
#endif
