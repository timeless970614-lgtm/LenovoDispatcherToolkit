// ipf_wrapper.cpp
// Implementation of ipf_wrapper.h

#include "ipf_wrapper.h"
#include <windows.h>
#include <iostream>
#include <string>

// ─────────────────────────────────────────────────────────────────────────────
// WinMSRIO.dll function types (for MSR reading)
// ─────────────────────────────────────────────────────────────────────────────
typedef BOOL(WINAPI* PFN_InitMSRIO)(void);
typedef VOID(WINAPI* PFN_DeinitMSRIO)(void);
typedef BOOL(WINAPI* PFN_Rdmsr)(DWORD index, PDWORD eax, PDWORD edx);

// ─────────────────────────────────────────────────────────────────────────────
// Internal function pointer types matching the DLL signatures
// ─────────────────────────────────────────────────────────────────────────────

// V2 types (LenovoIPFV2.dll)
typedef int(WINAPI* PFN_IPFV2_CONNECT)(void);
typedef void(WINAPI* PFN_IPFV2_DISCONNECT)(void);
typedef int(WINAPI* PFN_IPFV2_CHECK)(void);
typedef unsigned int(WINAPI* PFN_IPFV2_GET_SYSTEM_POWER)(void);
typedef unsigned int(WINAPI* PFN_IPFV2_GET_CPU_TEMP)(void);
typedef void(WINAPI* PFN_IPFV2_GET_DEFAULT_PLX)(unsigned int* pl, int len);
typedef unsigned int(WINAPI* PFN_IPFV2_GET_OEMV)(int id);
typedef int(WINAPI* PFN_IPFV2_CurrentGear)(void);

// V1 types (LenovoIPF.dll)
typedef int(WINAPI* PFN_V1_INITIAL)(const char* name);
typedef void(WINAPI* PFN_V1_DISCONNECT)(void);
typedef int(WINAPI* PFN_V1_GET_DEFAULT_PLX)(unsigned int arr[]);
typedef unsigned int(WINAPI* PFN_V1_GET_SYSTEM_POWER)(void);
typedef unsigned int(WINAPI* PFN_V1_GET_CPU_TEMP)(void);

// ─────────────────────────────────────────────────────────────────────────────
// Module handles
// ─────────────────────────────────────────────────────────────────────────────
static HMODULE g_hV2 = NULL;
static HMODULE g_hV1 = NULL;
static HMODULE g_hMSR = NULL;   // WinMSRIO.dll
static int     g_Version = 0;   // 1=V1, 2=V2
static bool    g_Connected = false;

// Path where the DLLs are located (set via IPF_SetDllPath)
static std::string g_DllPath = "";

// V2 function pointers
static PFN_IPFV2_CONNECT        fpV2_Connect        = nullptr;
static PFN_IPFV2_DISCONNECT     fpV2_Disconnect     = nullptr;
static PFN_IPFV2_CHECK          fpV2_Check          = nullptr;
static PFN_IPFV2_GET_SYSTEM_POWER fpV2_GetSystemPower = nullptr;
static PFN_IPFV2_GET_CPU_TEMP  fpV2_GetCpuTemp     = nullptr;
static PFN_IPFV2_GET_DEFAULT_PLX fpV2_GetDefaultPLX = nullptr;
static PFN_IPFV2_GET_OEMV      fpV2_GetOEMV        = nullptr;
static PFN_IPFV2_CurrentGear   fpV2_CurrentGear    = nullptr;

// V1 function pointers
static PFN_V1_INITIAL        fpV1_Initial        = nullptr;
static PFN_V1_DISCONNECT      fpV1_Disconnect     = nullptr;
static PFN_V1_GET_DEFAULT_PLX fpV1_GetDefaultPLX  = nullptr;
static PFN_V1_GET_SYSTEM_POWER fpV1_GetSystemPower = nullptr;
static PFN_V1_GET_CPU_TEMP   fpV1_GetCpuTemp     = nullptr;

// WinMSRIO function pointers
static PFN_InitMSRIO  fp_InitMSRIO  = nullptr;
static PFN_DeinitMSRIO fp_DeinitMSRIO = nullptr;
static PFN_Rdmsr      fp_Rdmsr      = nullptr;
static bool g_MSRInit = false;

// ─────────────────────────────────────────────────────────────────────────────
// DLL loading helpers
// ─────────────────────────────────────────────────────────────────────────────

static std::wstring makePath(const std::string& dllName) {
    if (g_DllPath.empty()) {
        return std::wstring(dllName.begin(), dllName.end());
    }
    std::wstring wp = std::wstring(g_DllPath.begin(), g_DllPath.end());
    if (!wp.empty() && wp.back() != L'\\' && wp.back() != L'/') wp += L"\\";
    wp += std::wstring(dllName.begin(), dllName.end());
    return wp;
}

#define LOAD_PROC_EX(hModule, type, name, outVar, fatal) \
    outVar = (type)GetProcAddress(hModule, name); \
    if (!outVar) { \
        std::cerr << "[IPF-Wrapper] " << (fatal ? "FATAL: " : "") << "Failed to load: " << name << std::endl; \
    }

#define LOAD_PROC(hModule, type, name, outVar) \
    LOAD_PROC_EX(hModule, type, name, outVar, false)

// ─────────────────────────────────────────────────────────────────────────────
// WinMSRIO DLL loading
// ─────────────────────────────────────────────────────────────────────────────
static bool loadMSR() {
    const wchar_t* dllName = L"WinMSRIO.dll";
    g_hMSR = LoadLibraryW(makePath(std::string(dllName, dllName + wcslen(dllName))).c_str());
    if (!g_hMSR) {
        std::cerr << "[IPF-Wrapper] Could not load WinMSRIO.dll" << std::endl;
        return false;
    }

    LOAD_PROC(g_hMSR, PFN_InitMSRIO,   "InitializeMSRIO",  fp_InitMSRIO);
    LOAD_PROC(g_hMSR, PFN_DeinitMSRIO, "DeinitializeMSRIO", fp_DeinitMSRIO);
    LOAD_PROC(g_hMSR, PFN_Rdmsr,       "Rdmsr",            fp_Rdmsr);

    if (!fp_InitMSRIO || !fp_Rdmsr) {
        std::cerr << "[IPF-Wrapper] WinMSRIO.dll missing required functions" << std::endl;
        FreeLibrary(g_hMSR);
        g_hMSR = NULL;
        return false;
    }

    // Initialize MSR access (opens kernel driver)
    BOOL ok = fp_InitMSRIO();
    if (!ok) {
        std::cerr << "[IPF-Wrapper] InitializeMSRIO() failed" << std::endl;
        FreeLibrary(g_hMSR);
        g_hMSR = NULL;
        return false;
    }

    g_MSRInit = true;
    std::cout << "[IPF-Wrapper] WinMSRIO.dll initialized OK" << std::endl;
    return true;
}

// ─────────────────────────────────────────────────────────────────────────────
// V2 DLL loading
// ─────────────────────────────────────────────────────────────────────────────
static bool loadV2() {
    const wchar_t* dllNames[] = {
        L"LenovoIPFV2.dll",
        L"LenovoIPF.dll",
    };

    HMODULE h = NULL;
    for (int i = 0; i < 2; i++) {
        std::wstring fullPath = makePath(std::string(dllNames[i], dllNames[i] + wcslen(dllNames[i])));
        h = LoadLibraryW(fullPath.c_str());
        if (h) break;
    }

    if (!h) {
        std::cerr << "[IPF-Wrapper] Could not load LenovoIPFV2.dll" << std::endl;
        return false;
    }

    fpV2_Connect = (PFN_IPFV2_CONNECT)GetProcAddress(h, "_IPFV2_Connect");
    if (fpV2_Connect) {
        g_hV2 = h;
        LOAD_PROC(g_hV2, PFN_IPFV2_DISCONNECT,       "_IPFV2_DisConnection",    fpV2_Disconnect);
        LOAD_PROC(g_hV2, PFN_IPFV2_CHECK,           "_IPFV2_CheckConnect",    fpV2_Check);
        LOAD_PROC(g_hV2, PFN_IPFV2_GET_SYSTEM_POWER, "_IPFV2_GetSystemPowerValue", fpV2_GetSystemPower);
        LOAD_PROC(g_hV2, PFN_IPFV2_GET_CPU_TEMP,    "_IPFV2_GetCpuTempValue", fpV2_GetCpuTemp);
        LOAD_PROC(g_hV2, PFN_IPFV2_GET_DEFAULT_PLX, "_IPFV2_GetDefaultPLX",   fpV2_GetDefaultPLX);
        LOAD_PROC(g_hV2, PFN_IPFV2_GET_OEMV,        "_IPFV2_GETOEMV",          fpV2_GetOEMV);
        LOAD_PROC(g_hV2, PFN_IPFV2_CurrentGear,     "_IPFV2_CurrentGear",     fpV2_CurrentGear);

        if (fpV2_Connect && fpV2_GetSystemPower && fpV2_GetDefaultPLX) {
            // Optional: CurrentGear function (may not exist on all versions)
            // Just load it if available, don't fail on missing
            g_Version = 2;
            return true;
        }
        FreeLibrary(h);
        g_hV2 = NULL;
    }

    FreeLibrary(h);
    g_hV2 = NULL;
    return false;
}

// ─────────────────────────────────────────────────────────────────────────────
// V1 DLL loading
// ─────────────────────────────────────────────────────────────────────────────
static bool loadV1() {
    const wchar_t* dllName = L"LenovoIPF.dll";
    g_hV1 = LoadLibraryW(makePath(std::string(dllName, dllName + wcslen(dllName))).c_str());
    if (!g_hV1) {
        std::cerr << "[IPF-Wrapper] Could not load LenovoIPF.dll" << std::endl;
        return false;
    }

    LOAD_PROC(g_hV1, PFN_V1_INITIAL,         "InitialConnection_Win32",      fpV1_Initial);
    LOAD_PROC(g_hV1, PFN_V1_DISCONNECT,      "DisConnection",                 fpV1_Disconnect);
    LOAD_PROC(g_hV1, PFN_V1_GET_DEFAULT_PLX, "GetDefaultPLX",                  fpV1_GetDefaultPLX);
    LOAD_PROC(g_hV1, PFN_V1_GET_SYSTEM_POWER,"GetSystemPowerValue",            fpV1_GetSystemPower);
    LOAD_PROC(g_hV1, PFN_V1_GET_CPU_TEMP,    "GetCpuTempValue",                fpV1_GetCpuTemp);

    if (fpV1_Initial && fpV1_GetSystemPower && fpV1_GetDefaultPLX) {
        g_Version = 1;
        return true;
    }

    std::cerr << "[IPF-Wrapper] V1 DLL loaded but required functions missing" << std::endl;
    FreeLibrary(g_hV1);
    g_hV1 = NULL;
    return false;
}

// ─────────────────────────────────────────────────────────────────────────────
// Public API Implementation
// ─────────────────────────────────────────────────────────────────────────────

void IPF_SetDllPath(const char* path) {
    g_DllPath = path ? path : "";
}

int IPF_Connect(void) {
    if (g_Connected) return g_Version;

    // Load WinMSRIO first (needed for MSR reads regardless of IPF version)
    loadMSR();

    // Try V2 first, then V1
    if (loadV2()) {
        std::cout << "[IPF-Wrapper] Using V2 (LenovoIPFV2)" << std::endl;
    } else if (loadV1()) {
        std::cout << "[IPF-Wrapper] Using V1 (LenovoIPF)" << std::endl;
    } else {
        std::cerr << "[IPF-Wrapper] No IPF DLL available" << std::endl;
        g_Version = 0;
    }

    if (g_Version == 1) {
        int rc = fpV1_Initial("LenovoToolkit");
        if (rc != 0) {
            std::cerr << "[IPF-Wrapper] V1 InitialConnection failed: " << rc << std::endl;
            FreeLibrary(g_hV1);
            g_hV1 = NULL;
            g_Version = 0;
            return 0;
        }
    } else if (g_Version == 2) {
        int rc = fpV2_Connect();
        if (rc < 2) {
            std::cerr << "[IPF-Wrapper] V2 Connect returned: " << rc << std::endl;
        }
    }

    g_Connected = true;
    return g_Version;
}

int IPF_GetVersion(void) {
    return g_Version;
}

void IPF_Disconnect(void) {
    if (!g_Connected) return;
    if (g_Version == 1 && fpV1_Disconnect) {
        fpV1_Disconnect();
    } else if (g_Version == 2 && fpV2_Disconnect) {
        fpV2_Disconnect();
    }
    if (g_hV1) { FreeLibrary(g_hV1); g_hV1 = NULL; }
    if (g_hV2) { FreeLibrary(g_hV2); g_hV2 = NULL; }
    if (g_MSRInit && fp_DeinitMSRIO) {
        fp_DeinitMSRIO();
        g_MSRInit = false;
    }
    if (g_hMSR) { FreeLibrary(g_hMSR); g_hMSR = NULL; }
    g_Connected = false;
    g_Version = 0;
}

unsigned int IPF_GetSystemPower_mW(void) {
    if (!g_Connected) return 0;
    if (g_Version == 2 && fpV2_GetSystemPower) {
        return fpV2_GetSystemPower();
    }
    if (g_Version == 1 && fpV1_GetSystemPower) {
        return fpV1_GetSystemPower();
    }
    return 0;
}

unsigned int IPF_GetCpuTemp_cK(void) {
    if (!g_Connected) return 0;
    if (g_Version == 2 && fpV2_GetCpuTemp) {
        return fpV2_GetCpuTemp();
    }
    if (g_Version == 1 && fpV1_GetCpuTemp) {
        return fpV1_GetCpuTemp();
    }
    return 0;
}

unsigned int IPF_GetPL1_mW(void) {
    unsigned int pl[3] = {0, 0, 0};
    IPF_GetAllPL_mW(pl);
    return pl[0];
}

unsigned int IPF_GetPL2_mW(void) {
    unsigned int pl[3] = {0, 0, 0};
    IPF_GetAllPL_mW(pl);
    return pl[1];
}

unsigned int IPF_GetPL4_mW(void) {
    unsigned int pl[3] = {0, 0, 0};
    IPF_GetAllPL_mW(pl);
    return pl[2];
}

void IPF_GetAllPL_mW(unsigned int pl[3]) {
    pl[0] = 0; pl[1] = 0; pl[2] = 0;
    if (!g_Connected) return;

    if (g_Version == 2 && fpV2_GetDefaultPLX) {
        fpV2_GetDefaultPLX(pl, 3);
        // Try to get PL4 via OEMV
        if (fpV2_GetOEMV) {
            unsigned int pl4 = fpV2_GetOEMV(4);
            if (pl4 > 0 && pl4 < 1000000) pl[2] = pl4;
        }
        return;
    }

    if (g_Version == 1 && fpV1_GetDefaultPLX) {
        unsigned int arr[2] = {0, 0};
        fpV1_GetDefaultPLX(arr);
        pl[0] = arr[0];
        pl[1] = arr[1];
        pl[2] = 0;
        return;
    }
}

// ─────────────────────────────────────────────────────────────────────────────
// MSR reading implementation
// ─────────────────────────────────────────────────────────────────────────────

int IPF_ReadMSR(unsigned int msrIndex, unsigned int* eax, unsigned int* edx) {
    if (!g_MSRInit || !fp_Rdmsr) return 0;

    DWORD eaxVal = 0, edxVal = 0;
    BOOL ok = fp_Rdmsr((DWORD)msrIndex, &eaxVal, &edxVal);
    if (!ok) return 0;

    if (eax) *eax = (unsigned int)eaxVal;
    if (edx) *edx = (unsigned int)edxVal;
    return 1;
}

static unsigned int callReadMsr(unsigned int msrIndex) {
    DWORD eax = 0, edx = 0;
    if (!IPF_ReadMSR(msrIndex, (unsigned int*)&eax, (unsigned int*)&edx)) return 0xFFFFFFFF;
    return (unsigned int)eax;
}

unsigned int IPF_GetEPP(void) {
    // MSR 0x1B0 = IA32_ENERGY_PERF_BIAS; bits 0-3 = EPP (0=perf, 15=battery)
    return callReadMsr(0x1B0) & 0xF;
}

unsigned int IPF_GetEPP1(void) {
    // MSR 0x1B0 bits 24-27 = EPP for E-Core on hybrid CPUs
    unsigned int val = callReadMsr(0x1B0);
    unsigned int epp1 = (val >> 24) & 0xF;
    if (epp1 == 0) epp1 = callReadMsr(0x644) & 0xFF;  // fallback: MSR 0x644
    return epp1;
}

unsigned int IPF_GetFrequencyLimit_MHz(void) {
    // MSR 0x770 = IA32_HWP_REQUEST; bits 15:8 = Maximum Performance Level
    unsigned int val = callReadMsr(0x770);
    unsigned int maxPerf = (val >> 8) & 0xFF;
    return (maxPerf == 0xFF) ? 0 : maxPerf;
}

unsigned int IPF_GetHeteroInc(void) {
    // MSR 0x1A0 bits 20-27 = Hetero Increase Threshold
    return (callReadMsr(0x1A0) >> 20) & 0xFF;
}

unsigned int IPF_GetHeteroDec(void) {
    // MSR 0x1A0 bits 28-31 = Hetero Decrease Threshold
    return (callReadMsr(0x1A0) >> 28) & 0xF;
}

unsigned int IPF_GetSoftParkLatency(void) {
    // MSR 0x1A0 bits 0-11 = Soft Park Latency (platform-specific)
    return callReadMsr(0x1A0) & 0xFFF;
}

int IPF_GetCurrentGear(void) {
    // Read current EPOT/Gear level (0-9) from LenovoIPFV2.dll
    // This is the same API used by ML_Scenario: _IPFV2_CurrentGear()
    if (!g_Connected) return -1;
    
    if (g_Version == 2 && fpV2_CurrentGear) {
        int gear = fpV2_CurrentGear();
        return gear;
    }
    
    // V1 DLL does not have this function, return -1
    return -1;
}
