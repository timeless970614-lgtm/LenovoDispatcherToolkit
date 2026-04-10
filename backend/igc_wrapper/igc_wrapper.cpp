// igc_wrapper.cpp
// Runtime-loads igc.dll (Intel GPU Control Library) and exposes a flat C API.
//
// Intel GPU Control Library types used here are defined inline to avoid
// requiring the SDK headers at build time.  The struct layouts match the
// public SDK (github.com/intel/drivers.gpu.control-library).

#include "igc_wrapper.h"
#include <windows.h>
#include <string>
#include <vector>
#include <cstring>
#include <cstdio>

// ─────────────────────────────────────────────────────────────────────────────
// Minimal Intel GPU Control Library type definitions
// (mirrors ctl_api.h from the public SDK)
// ─────────────────────────────────────────────────────────────────────────────

typedef void*   ctl_api_handle_t;
typedef void*   ctl_device_adapter_handle_t;
typedef void*   ctl_freq_handle_t;
typedef int     ctl_result_t;

#define CTL_RESULT_SUCCESS 0

// ctl_init_args_t  (simplified – only fields we need)
typedef struct _ctl_init_args_t {
    uint32_t Size;          // sizeof(ctl_init_args_t)
    uint8_t  Version;       // 0
    uint64_t AppVersion;    // caller version
    uint32_t flags;         // 0
    uint32_t SupportedVersion;
    uint8_t  reserved[116];
} ctl_init_args_t;

// ctl_device_adapter_properties_t  (simplified)
typedef struct _ctl_device_adapter_properties_t {
    uint32_t Size;
    uint8_t  Version;
    void*    pDeviceID;
    uint32_t device_id;
    uint32_t rev_id;
    uint32_t Frequency;
    char     name[100];
    uint8_t  DeviceType;
    uint8_t  reserved[128];
} ctl_device_adapter_properties_t;

// ctl_freq_properties_t
typedef struct _ctl_freq_properties_t {
    uint32_t Size;
    uint8_t  Version;
    uint8_t  type;          // 0=GPU, 1=Memory
    uint8_t  canControl;
    double   min;           // MHz
    double   max;           // MHz
    uint8_t  reserved[64];
} ctl_freq_properties_t;

// ctl_freq_range_t
typedef struct _ctl_freq_range_t {
    double min;   // MHz  (-1 = use default)
    double max;   // MHz  (-1 = use default)
} ctl_freq_range_t;

// ctl_freq_state_t
typedef struct _ctl_freq_state_t {
    uint32_t Size;
    uint8_t  Version;
    double   request;       // MHz
    double   tdp;           // MHz
    double   efficient;     // MHz
    double   actual;        // MHz
    uint32_t throttleReasons;
    uint8_t  reserved[64];
} ctl_freq_state_t;

// ─────────────────────────────────────────────────────────────────────────────
// Function pointer types
// ─────────────────────────────────────────────────────────────────────────────

typedef ctl_result_t (WINAPI *PFN_ctlInit)(ctl_init_args_t*, ctl_api_handle_t*);
typedef ctl_result_t (WINAPI *PFN_ctlClose)(ctl_api_handle_t);
typedef ctl_result_t (WINAPI *PFN_ctlEnumerateDevices)(ctl_api_handle_t, uint32_t*, ctl_device_adapter_handle_t*);
typedef ctl_result_t (WINAPI *PFN_ctlGetDeviceProperties)(ctl_device_adapter_handle_t, ctl_device_adapter_properties_t*);
typedef ctl_result_t (WINAPI *PFN_ctlEnumFrequencyDomains)(ctl_device_adapter_handle_t, uint32_t*, ctl_freq_handle_t*);
typedef ctl_result_t (WINAPI *PFN_ctlFrequencyGetProperties)(ctl_freq_handle_t, ctl_freq_properties_t*);
typedef ctl_result_t (WINAPI *PFN_ctlFrequencyGetRange)(ctl_freq_handle_t, ctl_freq_range_t*);
typedef ctl_result_t (WINAPI *PFN_ctlFrequencySetRange)(ctl_freq_handle_t, const ctl_freq_range_t*);
typedef ctl_result_t (WINAPI *PFN_ctlFrequencyGetState)(ctl_freq_handle_t, ctl_freq_state_t*);

// ─────────────────────────────────────────────────────────────────────────────
// Module state
// ─────────────────────────────────────────────────────────────────────────────

static HMODULE g_hDll = NULL;
static ctl_api_handle_t g_hApi = NULL;
static bool g_initialized = false;

static PFN_ctlInit                  pfn_ctlInit                  = NULL;
static PFN_ctlClose                 pfn_ctlClose                 = NULL;
static PFN_ctlEnumerateDevices      pfn_ctlEnumerateDevices      = NULL;
static PFN_ctlGetDeviceProperties   pfn_ctlGetDeviceProperties   = NULL;
static PFN_ctlEnumFrequencyDomains  pfn_ctlEnumFrequencyDomains  = NULL;
static PFN_ctlFrequencyGetProperties pfn_ctlFrequencyGetProperties = NULL;
static PFN_ctlFrequencyGetRange     pfn_ctlFrequencyGetRange     = NULL;
static PFN_ctlFrequencySetRange     pfn_ctlFrequencySetRange     = NULL;
static PFN_ctlFrequencyGetState     pfn_ctlFrequencyGetState     = NULL;

// Cached adapter handles
static std::vector<ctl_device_adapter_handle_t> g_adapters;
// Cached freq domain handles per adapter (first GPU domain only)
static std::vector<ctl_freq_handle_t> g_freqHandles;

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

#define LOAD_PROC(dll, name) \
    pfn_##name = (PFN_##name)GetProcAddress(dll, #name); \
    if (!pfn_##name) { FreeLibrary(dll); g_hDll = NULL; return IGC_ERR_NOT_LOADED; }

static int loadDll() {
    // Try igc.dll first (newer), then ControlLib.dll (older name)
    g_hDll = LoadLibraryA("igc.dll");
    if (!g_hDll) g_hDll = LoadLibraryA("ControlLib.dll");
    if (!g_hDll) return IGC_ERR_NOT_LOADED;

    LOAD_PROC(g_hDll, ctlInit)
    LOAD_PROC(g_hDll, ctlClose)
    LOAD_PROC(g_hDll, ctlEnumerateDevices)
    LOAD_PROC(g_hDll, ctlGetDeviceProperties)
    LOAD_PROC(g_hDll, ctlEnumFrequencyDomains)
    LOAD_PROC(g_hDll, ctlFrequencyGetProperties)
    LOAD_PROC(g_hDll, ctlFrequencyGetRange)
    LOAD_PROC(g_hDll, ctlFrequencySetRange)
    LOAD_PROC(g_hDll, ctlFrequencyGetState)
    return IGC_OK;
}

// Find the first GPU-type frequency domain handle for an adapter.
// Returns NULL if none found.
static ctl_freq_handle_t findGPUFreqDomain(ctl_device_adapter_handle_t hAdapter) {
    uint32_t count = 0;
    if (pfn_ctlEnumFrequencyDomains(hAdapter, &count, NULL) != CTL_RESULT_SUCCESS || count == 0)
        return NULL;

    std::vector<ctl_freq_handle_t> handles(count);
    if (pfn_ctlEnumFrequencyDomains(hAdapter, &count, handles.data()) != CTL_RESULT_SUCCESS)
        return NULL;

    for (uint32_t i = 0; i < count; i++) {
        ctl_freq_properties_t props = {};
        props.Size = sizeof(props);
        if (pfn_ctlFrequencyGetProperties(handles[i], &props) == CTL_RESULT_SUCCESS) {
            if (props.type == 0) // GPU domain
                return handles[i];
        }
    }
    // Fallback: return first domain
    return handles[0];
}

// ─────────────────────────────────────────────────────────────────────────────
// Public API implementation
// ─────────────────────────────────────────────────────────────────────────────

extern "C" {

int IGC_Init(void) {
    if (g_initialized) return IGC_OK;

    int rc = loadDll();
    if (rc != IGC_OK) return rc;

    ctl_init_args_t args = {};
    args.Size       = sizeof(args);
    args.AppVersion = 0x00010000; // 1.0

    ctl_result_t res = pfn_ctlInit(&args, &g_hApi);
    if (res != CTL_RESULT_SUCCESS) {
        FreeLibrary(g_hDll);
        g_hDll = NULL;
        return IGC_ERR_API_FAIL;
    }

    // Enumerate adapters
    uint32_t adapterCount = 0;
    pfn_ctlEnumerateDevices(g_hApi, &adapterCount, NULL);
    if (adapterCount == 0) {
        pfn_ctlClose(g_hApi);
        FreeLibrary(g_hDll);
        g_hDll = NULL;
        return IGC_ERR_NO_DEVICE;
    }

    g_adapters.resize(adapterCount);
    pfn_ctlEnumerateDevices(g_hApi, &adapterCount, g_adapters.data());

    // Pre-cache GPU freq domain handles
    g_freqHandles.resize(adapterCount, NULL);
    for (uint32_t i = 0; i < adapterCount; i++) {
        g_freqHandles[i] = findGPUFreqDomain(g_adapters[i]);
    }

    g_initialized = true;
    return IGC_OK;
}

void IGC_Close(void) {
    if (!g_initialized) return;
    g_adapters.clear();
    g_freqHandles.clear();
    if (g_hApi) { pfn_ctlClose(g_hApi); g_hApi = NULL; }
    if (g_hDll) { FreeLibrary(g_hDll); g_hDll = NULL; }
    g_initialized = false;
}

int IGC_GetAdapterCount(void) {
    if (!g_initialized) return 0;
    return (int)g_adapters.size();
}

int IGC_GetAdapterInfo(int adapterIndex, IGC_AdapterInfo* info) {
    if (!g_initialized) return IGC_ERR_NOT_INIT;
    if (!info) return IGC_ERR_API_FAIL;
    if (adapterIndex < 0 || adapterIndex >= (int)g_adapters.size()) return IGC_ERR_NO_DEVICE;

    ctl_device_adapter_properties_t props = {};
    props.Size = sizeof(props);
    ctl_result_t res = pfn_ctlGetDeviceProperties(g_adapters[adapterIndex], &props);
    if (res != CTL_RESULT_SUCCESS) return IGC_ERR_API_FAIL;

    memset(info, 0, sizeof(*info));
    strncpy_s(info->name, sizeof(info->name), props.name, _TRUNCATE);
    info->adapterIndex    = adapterIndex;
    info->freqDomainCount = (g_freqHandles[adapterIndex] != NULL) ? 1 : 0;
    return IGC_OK;
}

int IGC_GetFreqInfo(int adapterIndex, IGC_FreqInfo* info) {
    if (!g_initialized) return IGC_ERR_NOT_INIT;
    if (!info) return IGC_ERR_API_FAIL;
    if (adapterIndex < 0 || adapterIndex >= (int)g_adapters.size()) return IGC_ERR_NO_DEVICE;

    ctl_freq_handle_t hFreq = g_freqHandles[adapterIndex];
    if (!hFreq) return IGC_ERR_NO_FREQ_DOMAIN;

    memset(info, 0, sizeof(*info));

    // Hardware capability range
    ctl_freq_properties_t props = {};
    props.Size = sizeof(props);
    if (pfn_ctlFrequencyGetProperties(hFreq, &props) == CTL_RESULT_SUCCESS) {
        info->minFreqMHz = props.min;
        info->maxFreqMHz = props.max;
    }

    // Current software limits
    ctl_freq_range_t range = {};
    if (pfn_ctlFrequencyGetRange(hFreq, &range) == CTL_RESULT_SUCCESS) {
        info->currentMinMHz = range.min;
        info->currentMaxMHz = range.max;
    }

    // Actual state
    ctl_freq_state_t state = {};
    state.Size = sizeof(state);
    if (pfn_ctlFrequencyGetState(hFreq, &state) == CTL_RESULT_SUCCESS) {
        info->requestedMHz  = state.request;
        info->tdpMHz        = state.tdp;
        info->efficientMHz  = state.efficient;
        info->actualMHz     = state.actual;
    }

    return IGC_OK;
}

int IGC_SetFreqRange(int adapterIndex, double minMHz, double maxMHz) {
    if (!g_initialized) return IGC_ERR_NOT_INIT;
    if (adapterIndex < 0 || adapterIndex >= (int)g_adapters.size()) return IGC_ERR_NO_DEVICE;

    ctl_freq_handle_t hFreq = g_freqHandles[adapterIndex];
    if (!hFreq) return IGC_ERR_NO_FREQ_DOMAIN;

    ctl_freq_range_t range = { minMHz, maxMHz };
    ctl_result_t res = pfn_ctlFrequencySetRange(hFreq, &range);
    return (res == CTL_RESULT_SUCCESS) ? IGC_OK : IGC_ERR_API_FAIL;
}

const char* IGC_ErrorString(int code) {
    switch (code) {
        case IGC_OK:                  return "OK";
        case IGC_ERR_NOT_LOADED:      return "igc.dll not found or failed to load";
        case IGC_ERR_NO_DEVICE:       return "No Intel GPU adapter found";
        case IGC_ERR_NO_FREQ_DOMAIN:  return "No frequency domain on this adapter";
        case IGC_ERR_API_FAIL:        return "Intel GPU Control Library API call failed";
        case IGC_ERR_NOT_INIT:        return "IGC_Init() not called";
        default:                      return "Unknown error";
    }
}

} // extern "C"
