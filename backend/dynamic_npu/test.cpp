/**
 * Houmo DNPU (M50) Control Panel v1.1.0
 * =======================================
 *
 * Features:
 *   1. Display device info
 *   2. DVFS mode switching (Performance / Power-Saving)
 *   3. Real-time monitoring
 *   4. Smart auto-scheduler
 *   5. Parameter reference
 *
 * Compile (VS Developer Command Prompt):
 *   cl /EHsc /W4 /O2 test.cpp /Fe:test.exe user32.lib
 *
 * Run:
 *   Place libhal_xh2a.dll in the same directory as test.exe
 */

#define NOMINMAX
#define _CRT_SECURE_NO_WARNINGS
#include <windows.h>
#include <iostream>
#include <iomanip>
#include <string>
#include <vector>
#include <chrono>
#include <cstring>
#include <cstdlib>
#include <cstdio>
#include <algorithm>
#include <limits>
#include <conio.h>

// ============================================================
// Houmo DNPU API Types
// ============================================================

#define HM_SYS_DEVICE_NAME_LEN 16
#define HM_SYS_DEVICE_SN_LEN   32
#define HM_MAX_DEVICES          32

enum hm_dvfs_mode {
    HM_DVFS_PERFORMANCE = 0,
    HM_DVFS_ONDEMAND,
    HM_DVFS_USERSPACE,
    HM_DVFS_INDEPENDENT,
    HM_DVFS_EFFICIENCY_PREFILL,
    HM_DVFS_EFFICIENCY_DECODE,
    HM_DVFS_POWERSAVE_PREFILL,
    HM_DVFS_POWERSAVE_DECODE,
    HM_DVFS_MODE_MAX,
};

struct hm_device_info {
    uint32_t num_devices;
    uint32_t device_ids[HM_MAX_DEVICES];
};

struct hm_mem_info {
    uint32_t mem_total;
    uint32_t mem_used;
    uint32_t mem_avail;
};

// Function pointer types
typedef uint32_t (*PFN_hm_sys_get_device_info)(struct hm_device_info* info);
typedef int     (*PFN_hm_sys_get_vendor_id)(int dev_index);
typedef int     (*PFN_hm_sys_get_device_sn)(int dev_index, char sn[], int len);
typedef int     (*PFN_hm_sys_get_device_name)(int dev_index, char name[], int len);
typedef int     (*PFN_hm_sys_get_computing_power)(int dev_index);
typedef int     (*PFN_hm_sys_get_core_count)(int dev_index);
typedef float   (*PFN_hm_sys_get_ipu_utili_rate)(int dev_index);
typedef float   (*PFN_hm_sys_get_ipu_core_utili_rate)(int dev_index, uint32_t core_id);
typedef int     (*PFN_hm_sys_get_ipu_voltage)(int dev_index, float* voltage);
typedef int     (*PFN_hm_sys_get_ipu_frequency)(int dev_index, uint64_t* frequency);
typedef int     (*PFN_hm_sys_get_mem_info)(int dev_index, struct hm_mem_info* mem_info);
typedef int     (*PFN_hm_sys_get_temperature)(int dev_index, float* temp);
typedef int     (*PFN_hm_sys_get_buildtime)(char* buildtime, size_t len);
typedef int     (*PFN_hm_sys_get_version)(char* version, size_t len);
typedef int     (*PFN_hm_sys_get_driver_version)(char* version, size_t len);
typedef int     (*PFN_hm_sys_get_device_version)(int dev_index, char* version, size_t len);
typedef uint32_t (*PFN_hm_sys_get_ddr_size)(int dev_index, uint64_t* ddr_size);
typedef int     (*PFN_hm_sys_get_board_power)(int dev_index, float* power);
typedef int     (*PFN_hm_sys_get_dvfs_mode)(int dev_index, enum hm_dvfs_mode* mode);
typedef int     (*PFN_hm_sys_set_dvfs_mode)(int dev_index, enum hm_dvfs_mode mode);

// DVFS frequency control APIs (from dvfs_api.h)
typedef int     (*PFN_xh2a_dvfs_get_ipu_frequency)(int dev_id, uint64_t* frequency);
typedef int     (*PFN_xh2a_dvfs_set_ipu_frequency)(int dev_id, uint64_t frequency);
typedef int     (*PFN_xh2a_dvfs_get_cpu_frequency)(int dev_id, uint64_t* frequency);
typedef int     (*PFN_xh2a_dvfs_set_cpu_frequency)(int dev_id, uint64_t frequency);
typedef int     (*PFN_xh2a_dvfs_get_ddr_frequency)(int dev_id, uint64_t* frequency);
typedef int     (*PFN_xh2a_dvfs_set_ddr_frequency)(int dev_id, uint64_t frequency);
typedef int     (*PFN_xh2a_dvfs_get_noc_frequency)(int dev_id, uint64_t* frequency);
typedef int     (*PFN_xh2a_dvfs_set_noc_frequency)(int dev_id, uint64_t frequency);

// Power limit APIs (if available)
typedef int     (*PFN_xh2a_dvfs_get_power_limit)(int dev_id, float* min_power, float* max_power);
typedef int     (*PFN_xh2a_dvfs_set_power_limit)(int dev_id, float power_limit);
typedef int     (*PFN_hm_sys_set_dvfs_mode)(int dev_index, enum hm_dvfs_mode mode);

// ============================================================
// API Wrapper
// ============================================================

struct HoumoAPI {
    HMODULE hDll = nullptr;
    bool ok = false;

    PFN_hm_sys_get_device_info        pfn_get_device_info        = nullptr;
    PFN_hm_sys_get_vendor_id          pfn_get_vendor_id          = nullptr;
    PFN_hm_sys_get_device_sn          pfn_get_device_sn          = nullptr;
    PFN_hm_sys_get_device_name        pfn_get_device_name        = nullptr;
    PFN_hm_sys_get_computing_power    pfn_get_computing_power    = nullptr;
    PFN_hm_sys_get_core_count         pfn_get_core_count         = nullptr;
    PFN_hm_sys_get_ipu_utili_rate     pfn_get_ipu_utili_rate     = nullptr;
    PFN_hm_sys_get_ipu_core_utili_rate pfn_get_ipu_core_utili_rate = nullptr;
    PFN_hm_sys_get_ipu_voltage        pfn_get_ipu_voltage        = nullptr;
    PFN_hm_sys_get_ipu_frequency      pfn_get_ipu_frequency      = nullptr;
    PFN_hm_sys_get_mem_info           pfn_get_mem_info           = nullptr;
    PFN_hm_sys_get_temperature         pfn_get_temperature         = nullptr;
    PFN_hm_sys_get_buildtime           pfn_get_buildtime           = nullptr;
    PFN_hm_sys_get_version             pfn_get_version             = nullptr;
    PFN_hm_sys_get_driver_version      pfn_get_driver_version      = nullptr;
    PFN_hm_sys_get_device_version      pfn_get_device_version      = nullptr;
    PFN_hm_sys_get_ddr_size            pfn_get_ddr_size            = nullptr;
    PFN_hm_sys_get_board_power         pfn_get_board_power         = nullptr;
    PFN_hm_sys_get_dvfs_mode           pfn_get_dvfs_mode           = nullptr;
    PFN_hm_sys_set_dvfs_mode           pfn_set_dvfs_mode           = nullptr;

    // DVFS frequency control
    PFN_xh2a_dvfs_get_ipu_frequency    pfn_dvfs_get_ipu_freq       = nullptr;
    PFN_xh2a_dvfs_set_ipu_frequency    pfn_dvfs_set_ipu_freq       = nullptr;
    PFN_xh2a_dvfs_get_cpu_frequency    pfn_dvfs_get_cpu_freq       = nullptr;
    PFN_xh2a_dvfs_set_cpu_frequency    pfn_dvfs_set_cpu_freq       = nullptr;
    PFN_xh2a_dvfs_get_ddr_frequency    pfn_dvfs_get_ddr_freq       = nullptr;
    PFN_xh2a_dvfs_set_ddr_frequency    pfn_dvfs_set_ddr_freq       = nullptr;
    PFN_xh2a_dvfs_get_noc_frequency    pfn_dvfs_get_noc_freq       = nullptr;
    PFN_xh2a_dvfs_set_noc_frequency    pfn_dvfs_set_noc_freq       = nullptr;

    bool load(const wchar_t* dllPath) {
        hDll = LoadLibraryW(dllPath);
        if (!hDll) {
            wchar_t errBuf[256];
            FormatMessageW(FORMAT_MESSAGE_FROM_SYSTEM | FORMAT_MESSAGE_IGNORE_INSERTS,
                nullptr, GetLastError(), MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
                errBuf, 256, nullptr);
            std::wcerr << L"[ERROR] Failed to load DLL: " << errBuf << L"\n";
            return false;
        }

        pfn_get_device_info         = (PFN_hm_sys_get_device_info)GetProcAddress(hDll, "hm_sys_get_device_info");
        if (!pfn_get_device_info)         { std::cerr << "[ERROR] hm_sys_get_device_info not found\n"; return false; }
        pfn_get_vendor_id           = (PFN_hm_sys_get_vendor_id)GetProcAddress(hDll, "hm_sys_get_vendor_id");
        if (!pfn_get_vendor_id)           { std::cerr << "[ERROR] hm_sys_get_vendor_id not found\n"; return false; }
        pfn_get_device_sn           = (PFN_hm_sys_get_device_sn)GetProcAddress(hDll, "hm_sys_get_device_sn");
        if (!pfn_get_device_sn)           { std::cerr << "[ERROR] hm_sys_get_device_sn not found\n"; return false; }
        pfn_get_device_name         = (PFN_hm_sys_get_device_name)GetProcAddress(hDll, "hm_sys_get_device_name");
        if (!pfn_get_device_name)         { std::cerr << "[ERROR] hm_sys_get_device_name not found\n"; return false; }
        pfn_get_computing_power     = (PFN_hm_sys_get_computing_power)GetProcAddress(hDll, "hm_sys_get_computing_power");
        if (!pfn_get_computing_power)     { std::cerr << "[ERROR] hm_sys_get_computing_power not found\n"; return false; }
        pfn_get_core_count          = (PFN_hm_sys_get_core_count)GetProcAddress(hDll, "hm_sys_get_core_count");
        if (!pfn_get_core_count)          { std::cerr << "[ERROR] hm_sys_get_core_count not found\n"; return false; }
        pfn_get_ipu_utili_rate      = (PFN_hm_sys_get_ipu_utili_rate)GetProcAddress(hDll, "hm_sys_get_ipu_utili_rate");
        if (!pfn_get_ipu_utili_rate)      { std::cerr << "[ERROR] hm_sys_get_ipu_utili_rate not found\n"; return false; }
        pfn_get_ipu_core_utili_rate = (PFN_hm_sys_get_ipu_core_utili_rate)GetProcAddress(hDll, "hm_sys_get_ipu_core_utili_rate");
        if (!pfn_get_ipu_core_utili_rate) { std::cerr << "[ERROR] hm_sys_get_ipu_core_utili_rate not found\n"; return false; }
        pfn_get_ipu_voltage         = (PFN_hm_sys_get_ipu_voltage)GetProcAddress(hDll, "hm_sys_get_ipu_voltage");
        if (!pfn_get_ipu_voltage)         { std::cerr << "[ERROR] hm_sys_get_ipu_voltage not found\n"; return false; }
        pfn_get_ipu_frequency       = (PFN_hm_sys_get_ipu_frequency)GetProcAddress(hDll, "hm_sys_get_ipu_frequency");
        if (!pfn_get_ipu_frequency)       { std::cerr << "[ERROR] hm_sys_get_ipu_frequency not found\n"; return false; }
        pfn_get_mem_info            = (PFN_hm_sys_get_mem_info)GetProcAddress(hDll, "hm_sys_get_mem_info");
        if (!pfn_get_mem_info)           { std::cerr << "[ERROR] hm_sys_get_mem_info not found\n"; return false; }
        pfn_get_temperature          = (PFN_hm_sys_get_temperature)GetProcAddress(hDll, "hm_sys_get_temperature");
        if (!pfn_get_temperature)          { std::cerr << "[ERROR] hm_sys_get_temperature not found\n"; return false; }
        pfn_get_buildtime            = (PFN_hm_sys_get_buildtime)GetProcAddress(hDll, "hm_sys_get_buildtime");
        if (!pfn_get_buildtime)            { std::cerr << "[ERROR] hm_sys_get_buildtime not found\n"; return false; }
        pfn_get_version              = (PFN_hm_sys_get_version)GetProcAddress(hDll, "hm_sys_get_version");
        if (!pfn_get_version)              { std::cerr << "[ERROR] hm_sys_get_version not found\n"; return false; }
        pfn_get_driver_version       = (PFN_hm_sys_get_driver_version)GetProcAddress(hDll, "hm_sys_get_driver_version");
        if (!pfn_get_driver_version)       { std::cerr << "[ERROR] hm_sys_get_driver_version not found\n"; return false; }
        pfn_get_device_version       = (PFN_hm_sys_get_device_version)GetProcAddress(hDll, "hm_sys_get_device_version");
        if (!pfn_get_device_version)       { std::cerr << "[ERROR] hm_sys_get_device_version not found\n"; return false; }
        pfn_get_board_power          = (PFN_hm_sys_get_board_power)GetProcAddress(hDll, "hm_sys_get_board_power");
        if (!pfn_get_board_power)          { std::cerr << "[ERROR] hm_sys_get_board_power not found\n"; return false; }
        pfn_get_dvfs_mode            = (PFN_hm_sys_get_dvfs_mode)GetProcAddress(hDll, "hm_sys_get_dvfs_mode");
        if (!pfn_get_dvfs_mode)            { std::cerr << "[ERROR] hm_sys_get_dvfs_mode not found\n"; return false; }
        pfn_set_dvfs_mode            = (PFN_hm_sys_set_dvfs_mode)GetProcAddress(hDll, "hm_sys_set_dvfs_mode");
        if (!pfn_set_dvfs_mode)            { std::cerr << "[ERROR] hm_sys_set_dvfs_mode not found\n"; return false; }

        // Optional DVFS frequency APIs (may not exist in all versions)
        pfn_dvfs_get_ipu_freq = (PFN_xh2a_dvfs_get_ipu_frequency)GetProcAddress(hDll, "xh2a_dvfs_get_ipu_frequency");
        pfn_dvfs_set_ipu_freq = (PFN_xh2a_dvfs_set_ipu_frequency)GetProcAddress(hDll, "xh2a_dvfs_set_ipu_frequency");
        pfn_dvfs_get_cpu_freq = (PFN_xh2a_dvfs_get_cpu_frequency)GetProcAddress(hDll, "xh2a_dvfs_get_cpu_frequency");
        pfn_dvfs_set_cpu_freq = (PFN_xh2a_dvfs_set_cpu_frequency)GetProcAddress(hDll, "xh2a_dvfs_set_cpu_frequency");
        pfn_dvfs_get_ddr_freq = (PFN_xh2a_dvfs_get_ddr_frequency)GetProcAddress(hDll, "xh2a_dvfs_get_ddr_frequency");
        pfn_dvfs_set_ddr_freq = (PFN_xh2a_dvfs_set_ddr_frequency)GetProcAddress(hDll, "xh2a_dvfs_set_ddr_frequency");
        pfn_dvfs_get_noc_freq = (PFN_xh2a_dvfs_get_noc_frequency)GetProcAddress(hDll, "xh2a_dvfs_get_noc_frequency");
        pfn_dvfs_set_noc_freq = (PFN_xh2a_dvfs_set_noc_frequency)GetProcAddress(hDll, "xh2a_dvfs_set_noc_frequency");

        ok = true;
        return true;
    }

    ~HoumoAPI() {
        if (hDll) FreeLibrary(hDll);
    }
};

static HoumoAPI g_api;

// ============================================================
// Constants
// ============================================================

static const float TEMP_WARNING    = 80.0f;
static const float TEMP_CRITICAL   = 90.0f;
static const float UTIL_HIGH       = 0.85f;
static const float UTIL_LOW        = 0.20f;
static const int   MONITOR_INTERVAL  = 2000;
static const int   AUTO_ADJUST_DELAY = 5000;

// ============================================================
// Utility
// ============================================================

static void cls() { system("cls"); }

static void print_line(char c = '-', int w = 76) {
    std::cout << std::string(w, c) << "\n";
}

static void print_title(const char* t) {
    std::cout << "\n"; print_line('=');
    std::cout << "  " << t << "\n"; print_line('=');
}

static void print_sub(const char* t) {
    std::cout << "\n[ " << t << " ]\n"; print_line('-');
}

static void pause_console() {
    std::cout << "\nPress ENTER to continue...";
    std::cin.get();
}

static int read_int(const char* prompt, int lo, int hi) {
    int v;
    while (true) {
        std::cout << prompt;
        if (std::cin >> v) {
            if (v >= lo && v <= hi) {
                std::cin.ignore(std::numeric_limits<std::streamsize>::max(), '\n');
                return v;
            }
        }
        std::cin.clear();
        std::cin.ignore(std::numeric_limits<std::streamsize>::max(), '\n');
        std::cout << "  Invalid input. Please enter " << lo << " ~ " << hi << ".\n";
    }
}

static void read_float(const char* prompt, float* out, float default_val) {
    std::cout << prompt << " (default " << default_val << "): ";
    std::string s;
    std::getline(std::cin, s);
    if (!s.empty()) {
        *out = (float)atof(s.c_str());
    } else {
        *out = default_val;
    }
}

static std::string fmt_mem(uint32_t mb) {
    if (mb >= 1024) {
        char buf[32];
        sprintf(buf, "%.1f GB", mb / 1024.0);
        return std::string(buf);
    }
    return std::to_string(mb) + " MB";
}

static const char* dvfs_name(enum hm_dvfs_mode m) {
    switch (m) {
        case HM_DVFS_PERFORMANCE:       return "PERFORMANCE";
        case HM_DVFS_ONDEMAND:          return "ONDEMAND";
        case HM_DVFS_USERSPACE:         return "USERSPACE";
        case HM_DVFS_INDEPENDENT:       return "INDEPENDENT";
        case HM_DVFS_EFFICIENCY_PREFILL: return "EFFICIENCY_PREFILL";
        case HM_DVFS_EFFICIENCY_DECODE: return "EFFICIENCY_DECODE";
        case HM_DVFS_POWERSAVE_PREFILL: return "POWERSAVE_PREFILL";
        case HM_DVFS_POWERSAVE_DECODE:  return "POWERSAVE_DECODE";
        default:                        return "UNKNOWN";
    }
}

// ============================================================
// Device Info
// ============================================================

struct DeviceInfo {
    int dev_id = 0;
    int vendor_id = 0;
    char serial[HM_SYS_DEVICE_SN_LEN] = {0};
    char name[HM_SYS_DEVICE_NAME_LEN] = {0};
    int computing_power = 0;
    int core_count = 0;
    float ipu_util = -1.0f;
    float ipu_freq = 0.0f;
    float ipu_voltage = -1.0f;
    float temperature = -1.0f;
    float board_power = -1.0f;
    struct hm_mem_info mem_info = {0,0,0};
    char driver_version[128] = {0};
    char device_version[128] = {0};
    char sdk_version[128] = {0};
    enum hm_dvfs_mode dvfs_mode = HM_DVFS_PERFORMANCE;
};

static bool get_device_info(int dev_id, DeviceInfo* info) {
    if (!info) return false;
    memset(info, 0, sizeof(DeviceInfo));
    info->dev_id = dev_id;

    info->vendor_id = g_api.pfn_get_vendor_id(dev_id);
    g_api.pfn_get_device_sn(dev_id, info->serial, sizeof(info->serial));
    g_api.pfn_get_device_name(dev_id, info->name, sizeof(info->name));
    info->computing_power = g_api.pfn_get_computing_power(dev_id);
    info->core_count = g_api.pfn_get_core_count(dev_id);

    info->ipu_util = g_api.pfn_get_ipu_utili_rate(dev_id);

    uint64_t freq = 0;
    g_api.pfn_get_ipu_frequency(dev_id, &freq);
    info->ipu_freq = (float)freq;

    g_api.pfn_get_ipu_voltage(dev_id, &info->ipu_voltage);
    g_api.pfn_get_temperature(dev_id, &info->temperature);
    g_api.pfn_get_board_power(dev_id, &info->board_power);
    g_api.pfn_get_mem_info(dev_id, &info->mem_info);

    g_api.pfn_get_driver_version(info->driver_version, sizeof(info->driver_version));
    g_api.pfn_get_device_version(dev_id, info->device_version, sizeof(info->device_version));
    g_api.pfn_get_version(info->sdk_version, sizeof(info->sdk_version));
    g_api.pfn_get_dvfs_mode(dev_id, &info->dvfs_mode);

    return true;
}

static void print_device_info(const DeviceInfo& info) {
    std::cout.setf(std::ios::fixed);

    print_sub("Basic Info");
    std::cout << "  Device ID:       " << info.dev_id << "\n";
    std::cout << "  Device Name:    " << info.name << "\n";
    std::cout << "  Vendor ID:      0x" << std::hex << info.vendor_id << std::dec << "\n";
    std::cout << "  Serial:         " << info.serial << "\n";
    std::cout << "  Computing Power: " << info.computing_power << " TOPS\n";
    std::cout << "  Core Count:     " << info.core_count << "\n";

    print_sub("DVFS Mode");
    const char* modeDesc = (info.dvfs_mode == HM_DVFS_PERFORMANCE)
        ? "Fixed 1400 MHz, max throughput, high power"
        : "Dynamic 200-1400 MHz, auto-adjust, low power";
    std::cout << "  Mode:           " << dvfs_name(info.dvfs_mode) << "\n";
    std::cout << "  Description:    " << modeDesc << "\n";

    print_sub("Runtime Metrics");

    if (info.ipu_util >= 0) {
        std::cout << "  IPU Utilization: " << std::setprecision(1) << info.ipu_util * 100.0f << " %\n";
    } else {
        std::cout << "  IPU Utilization:  N/A\n";
    }

    if (info.ipu_freq > 0) {
        std::cout << "  IPU Frequency:    " << std::setprecision(3) << info.ipu_freq / 1e9f << " GHz\n";
    } else {
        std::cout << "  IPU Frequency:    N/A\n";
    }

    if (info.ipu_voltage >= 0) {
        std::cout << "  IPU Voltage:      " << std::setprecision(2) << info.ipu_voltage << " mV\n";
    }

    std::cout << "  Temperature:     ";
    if (info.temperature >= 0) {
        if (info.temperature >= TEMP_CRITICAL) {
            std::cout << "\033[1;31m" << std::setprecision(1) << info.temperature << " C\033[0m [!! DANGER]\n";
        } else if (info.temperature >= TEMP_WARNING) {
            std::cout << "\033[1;33m" << std::setprecision(1) << info.temperature << " C\033[0m [!! WARNING]\n";
        } else {
            std::cout << std::setprecision(1) << info.temperature << " C [OK]\n";
        }
    } else {
        std::cout << "N/A\n";
    }

    if (info.board_power >= 0) {
        std::cout << "  Board Power:     " << std::setprecision(2) << info.board_power << " W\n";
    }

    print_sub("Memory Info");
    std::cout << "  Total:           " << fmt_mem(info.mem_info.mem_total) << "\n";
    std::cout << "  Used:            " << fmt_mem(info.mem_info.mem_used) << "\n";
    std::cout << "  Available:       " << fmt_mem(info.mem_info.mem_avail) << "\n";

    if (info.core_count > 0) {
        print_sub("Per-Core Utilization");
        for (int i = 0; i < info.core_count; i++) {
            float cu = g_api.pfn_get_ipu_core_utili_rate(info.dev_id, (uint32_t)i);
            if (cu >= 0) {
                std::cout << "  Core[" << i << "]:           " << std::setprecision(1)
                          << cu * 100.0f << " %\n";
            }
        }
    }

    print_sub("Version Info");
    std::cout << "  SDK Version:     " << info.sdk_version << "\n";
    std::cout << "  Driver Version:  " << info.driver_version << "\n";
    std::cout << "  Firmware Version: " << info.device_version << "\n";
}

// ============================================================
// DVFS Operations
// ============================================================

static bool set_dvfs_mode(int dev_id, enum hm_dvfs_mode mode) {
    int ret = g_api.pfn_set_dvfs_mode(dev_id, mode);
    if (ret == 0) {
        std::cout << "[Device " << dev_id << "] DVFS mode -> " << dvfs_name(mode) << "\n";
        return true;
    }
    std::cerr << "[Device " << dev_id << "] Failed to set DVFS mode (error: " << ret << ")\n";
    return false;
}

static void show_current_dvfs(int dev_id) {
    enum hm_dvfs_mode m;
    if (g_api.pfn_get_dvfs_mode(dev_id, &m) == 0) {
        std::cout << "  Current Mode: " << dvfs_name(m) << "\n";
    }
}

// ============================================================
// Menu
// ============================================================

static void show_menu() {
    cls();
    print_title("Houmo DNPU Control Panel v1.2.0");
    std::cout << "\n";
    std::cout << "  [1] Show Device Info\n";
    std::cout << "  [2] Switch DVFS Mode\n";
    std::cout << "  [3] Real-time Monitor\n";
    std::cout << "  [4] Smart Auto-Scheduler\n";
    std::cout << "  [5] Parameter Reference\n";
    std::cout << "  [6] Power Management\n";
    std::cout << "  [0] Exit\n";
    std::cout << "\n"; print_line();
}

// ============================================================
// Feature 1: Show Devices
// ============================================================

static void func_show_devices() {
    cls();
    print_title("Device Info");

    struct hm_device_info dinfo = {0};
    uint32_t ret = g_api.pfn_get_device_info(&dinfo);
    uint32_t n = dinfo.num_devices ? dinfo.num_devices : ret;

    if (n == 0) {
        std::cerr << "\n  [ERROR] No DNPU device found!\n";
        pause_console();
        return;
    }

    std::cout << "\n  Found " << n << " DNPU device(s)\n";

    for (uint32_t i = 0; i < n; ++i) {
        int id = (int)dinfo.device_ids[i];
        DeviceInfo di;
        if (get_device_info(id, &di)) {
            std::cout << "\n"; print_device_info(di);
        }
        if (i < n - 1) { std::cout << "\n"; print_line(); }
    }

    pause_console();
}

// ============================================================
// Feature 2: Switch DVFS
// ============================================================

static void func_set_dvfs() {
    struct hm_device_info dinfo = {0};
    uint32_t ret = g_api.pfn_get_device_info(&dinfo);
    uint32_t n = dinfo.num_devices ? dinfo.num_devices : ret;
    if (n == 0) { std::cerr << "\n  [ERROR] No device found!\n"; pause_console(); return; }

    int dev = 0;
    if (n > 1) dev = read_int("\n  Select Device ID: ", 0, (int)(n - 1));

    while (true) {
        cls();
        print_title("DVFS Mode Switch");
        std::cout << "\n  Device ID: " << dev << "\n";
        show_current_dvfs(dev);

        std::cout << "\n  Select DVFS Mode:\n\n";
        std::cout << "  [0] PERFORMANCE\n";
        std::cout << "      - Fixed 1400 MHz\n";
        std::cout << "      - Max throughput\n";
        std::cout << "      - High power consumption\n";
        std::cout << "      - Use for: batch inference, compute-intensive tasks\n\n";
        std::cout << "  [1] ONDEMAND\n";
        std::cout << "      - Dynamic 200-1400 MHz\n";
        std::cout << "      - Auto-adjusts based on load\n";
        std::cout << "      - Low power consumption\n";
        std::cout << "      - Use for: standby, idle, low-load scenarios\n\n";
        std::cout << "  [2] Back to Main Menu\n";
        std::cout << "\n"; print_line();

        int ch = read_int("  Select [0-2]: ", 0, 2);
        if (ch == 2) break;

        enum hm_dvfs_mode mode = (ch == 0) ? HM_DVFS_PERFORMANCE : HM_DVFS_ONDEMAND;
        std::cout << "\n  Switching DVFS mode...\n\n";
        if (set_dvfs_mode(dev, mode)) {
            Sleep(1000);
            DeviceInfo di;
            if (get_device_info(dev, &di)) {
                std::cout << "  After switching:\n";
                show_current_dvfs(dev);
                if (di.ipu_freq > 0) {
                    std::cout << "  Current Freq: " << std::fixed << std::setprecision(2)
                              << di.ipu_freq / 1e6 << " MHz\n";
                }
                if (di.board_power >= 0) {
                    std::cout << "  Current Power: " << di.board_power << " W\n";
                }
            }
        }

        std::cout << "\n  Continue switching? [y/N]: ";
        char c; std::cin >> c;
        std::cin.ignore(std::numeric_limits<std::streamsize>::max(), '\n');
        if (c != 'y' && c != 'Y') break;
    }
}

// ============================================================
// Feature 3: Real-time Monitor
// ============================================================

static void func_realtime_monitor() {
    struct hm_device_info dinfo = {0};
    uint32_t ret = g_api.pfn_get_device_info(&dinfo);
    uint32_t n = dinfo.num_devices ? dinfo.num_devices : ret;
    if (n == 0) { std::cerr << "\n  [ERROR] No device found!\n"; pause_console(); return; }

    int dev = 0;
    if (n > 1) dev = read_int("\n  Select Device ID to monitor: ", 0, (int)(n - 1));

    int duration = read_int("\n  Duration (seconds, 0=unlimited): ", 0, 3600);

    cls();
    print_title("Real-time Monitor");
    std::cout << "\n  Device ID: " << dev;
    std::cout << "\n  Press Ctrl+C to exit\n";
    print_line();

    auto start = std::chrono::steady_clock::now();
    int iter = 0;

    while (true) {
        if (duration > 0) {
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                std::chrono::steady_clock::now() - start).count();
            if (elapsed >= duration) break;
            std::cout << "\n  [" << std::setw(4) << elapsed << "s] ";
        } else {
            std::cout << "\n  [" << std::setw(4) << ++iter << "] ";
        }

        DeviceInfo di;
        if (get_device_info(dev, &di)) {
            std::cout << "DVFS: " << dvfs_name(di.dvfs_mode) << " | ";

            std::cout.setf(std::ios::fixed);
            if (di.ipu_util >= 0) {
                std::cout << "Util: " << std::setw(5) << std::setprecision(1)
                          << di.ipu_util * 100.0f << "% | ";
            }
            if (di.ipu_freq > 0) {
                std::cout << "Freq: " << std::setprecision(2)
                          << di.ipu_freq / 1e9f << " GHz | ";
            }
            if (di.temperature >= 0) {
                std::cout << "Temp: " << std::setprecision(1);
                if (di.temperature >= TEMP_WARNING) {
                    std::cout << "\033[1;33m" << di.temperature << "C\033[0m";
                } else {
                    std::cout << di.temperature << "C";
                }
                std::cout << " | ";
            }
            if (di.board_power >= 0) {
                std::cout << "Power: " << di.board_power << " W | ";
            }
            std::cout << "Mem: " << di.mem_info.mem_used << "/"
                      << di.mem_info.mem_total << " MB";
        }
        std::cout << std::flush;
        Sleep(MONITOR_INTERVAL);
    }

    std::cout << "\n\n  Monitor ended.\n";
    pause_console();
}

// ============================================================
// Feature 4: Smart Scheduler
// ============================================================

static void func_smart_scheduler() {
    struct hm_device_info dinfo = {0};
    uint32_t ret = g_api.pfn_get_device_info(&dinfo);
    uint32_t n = dinfo.num_devices ? dinfo.num_devices : ret;
    if (n == 0) { std::cerr << "\n  [ERROR] No device found!\n"; pause_console(); return; }

    int dev = 0;
    if (n > 1) dev = read_int("\n  Select Device ID: ", 0, (int)(n - 1));

    cls();
    print_title("Smart Scheduler - Setup");
    std::cout << "\n  [Tip] Press ENTER to use default value\n\n";

    float tempWarn = TEMP_WARNING;
    float tempCrit = TEMP_CRITICAL;
    float utilHigh = UTIL_HIGH;
    float utilLow  = UTIL_LOW;

    read_float("  Temperature Warning Threshold", &tempWarn, TEMP_WARNING);
    read_float("  Temperature Critical Threshold", &tempCrit, TEMP_CRITICAL);
    read_float("  High Load Threshold (%)", &utilHigh, UTIL_HIGH * 100.0f);
    utilHigh /= 100.0f;
    read_float("  Low Load Threshold (%)", &utilLow, UTIL_LOW * 100.0f);
    utilLow /= 100.0f;

    int duration = read_int("\n  Duration (seconds, 0=unlimited): ", 0, 3600);

    cls();
    print_title("Smart Scheduler - Running");
    std::cout << "\n  Device ID: " << dev;
    std::cout << "\n  Strategy:\n";
    std::cout << "    - Load > " << (utilHigh*100) << "% -> PERFORMANCE\n";
    std::cout << "    - Load < " << (utilLow*100) << "% -> ONDEMAND\n";
    std::cout << "    - Temp > " << tempWarn << "C -> Force ONDEMAND\n";
    std::cout << "    - Temp > " << tempCrit << "C -> Force lowest power\n";
    std::cout << "\n  Press Ctrl+C to exit\n";
    print_line();

    enum hm_dvfs_mode last_mode = HM_DVFS_ONDEMAND;
    auto start = std::chrono::steady_clock::now();
    int iter = 0;

    while (true) {
        if (duration > 0) {
            auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                std::chrono::steady_clock::now() - start).count();
            if (elapsed >= duration) break;
            std::cout << "\n  [" << std::setw(4) << elapsed << "s] ";
        } else {
            std::cout << "\n  [" << std::setw(4) << ++iter << "] ";
        }

        DeviceInfo di;
        if (!get_device_info(dev, &di)) {
            std::cout << "Failed to get device info";
            Sleep(AUTO_ADJUST_DELAY);
            continue;
        }

        bool highUtil = (di.ipu_util >= 0) && (di.ipu_util > utilHigh);
        bool lowUtil  = (di.ipu_util >= 0) && (di.ipu_util < utilLow);
        bool highTemp = (di.temperature >= 0) && (di.temperature > tempWarn);
        bool critTemp = (di.temperature >= 0) && (di.temperature > tempCrit);

        enum hm_dvfs_mode target = di.dvfs_mode;
        const char* reason = "";

        if (critTemp) {
            target = HM_DVFS_ONDEMAND;
            reason = "[!! CRITICAL TEMP -> POWER SAVE]";
        } else if (highTemp) {
            target = HM_DVFS_ONDEMAND;
            reason = "[!  HIGH TEMP -> POWER SAVE]";
        } else if (highUtil) {
            target = HM_DVFS_PERFORMANCE;
            reason = "[>> HIGH LOAD -> PERFORMANCE]";
        } else if (lowUtil) {
            target = HM_DVFS_ONDEMAND;
            reason = "[<< LOW LOAD -> POWER SAVE]";
        } else {
            reason = "[== LOAD OK]";
        }

        if (target != last_mode) {
            set_dvfs_mode(dev, target);
            last_mode = target;
        }

        std::cout << "Mode: " << dvfs_name(di.dvfs_mode) << " " << reason << "\n    ";

        std::cout.setf(std::ios::fixed);
        std::cout << "Util: " << std::setprecision(1);
        if (di.ipu_util >= 0) std::cout << di.ipu_util * 100.0f << "%";
        else std::cout << "N/A";

        std::cout << " | Temp: ";
        if (di.temperature >= 0) {
            if (highTemp) std::cout << "\033[1;33m" << di.temperature << "C\033[0m";
            else std::cout << di.temperature << "C";
        } else std::cout << "N/A";

        std::cout << " | Freq: ";
        if (di.ipu_freq > 0) std::cout << std::setprecision(2) << di.ipu_freq / 1e9f << " GHz";
        else std::cout << "N/A";

        std::cout << " | Power: ";
        if (di.board_power >= 0) std::cout << di.board_power << " W";
        else std::cout << "N/A";

        std::cout << std::flush;
        Sleep(AUTO_ADJUST_DELAY);
    }

    std::cout << "\n\n  Scheduler ended.\n";
    pause_console();
}

// ============================================================
// Feature 5: Parameter Reference
// ============================================================

static void func_param_ref() {
    cls();
    print_title("DNPU Adjustable Parameters Reference");

    std::cout << R"(
  +------------------------------------------------------------------+
  |                    Adjustable Parameters                         |
  +------------------------------------------------------------------+
  |
  |  [1] DVFS Mode
  |  -----------------------------------------------------------
  |
  |  +-------------------+------------------------------------------+
  |  | PERFORMANCE       | Fixed 1400 MHz                          |
  |  | (High Performance)| Max throughput, lowest latency          |
  |  |                   | Highest power & heat                     |
  |  |                   | Use for: batch inference, compute tasks   |
  |  +-------------------+------------------------------------------+
  |  | ONDEMAND          | Dynamic 200-1400 MHz                    |
  |  | (Power Saving)    | Auto up/down based on load              |
  |  |                   | Lowest power & heat                     |
  |  |                   | Use for: standby, idle, low-load        |
  |  +-------------------+------------------------------------------+
  |
  |  [2] Smart Scheduler Strategy
  |  -----------------------------------------------------------
  |
  |  +-------------------+------------------------------------------+
  |  | Scenario          | Recommended Setting                     |
  |  +-------------------+------------------------------------------+
  |  | Deep Learning Inf | PERFORMANCE + temp monitor             |
  |  | Real-time Inf     | PERFORMANCE (latency-sensitive)         |
  |  | Batch Processing  | PERFORMANCE (throughput-first)         |
  |  | Dev/Debug         | ONDEMAND (reduce heat)                 |
  |  | Long Deployment   | ONDEMAND + temp protection            |
  |  | Edge Device       | ONDEMAND + smart scheduler (efficiency)|
  |  +-------------------+------------------------------------------+
  |
  |  [3] Read-only Metrics
  |  -----------------------------------------------------------
  |
  |  +------------------+-------+----------------------------------+
  |  | Metric           | Unit  | Description                     |
  |  +------------------+-------+----------------------------------+
  |  | IPU Utilization  | %     | Processor load, >85% = heavy    |
  |  | IPU Frequency    | MHz   | PERFORMANCE = 1400MHz fixed    |
  |  | IPU Voltage      | mV    | Scales with frequency           |
  |  | Temperature      | C     | >80C warning, >90C danger      |
  |  | Board Power      | W     | Real-time power consumption    |
  |  | DDR Memory       | MB    | Video memory usage             |
  |  +------------------+-------+----------------------------------+
  |
  |  [4] Firmware Update
  |  -----------------------------------------------------------
  |
  |  CLI command:
  |  "C:\Program Files (x86)\houmo-drv-xh2_v1.1.0\
  |   tools\hm_upgrade_cli\hm_upgrade_cli.exe"
  |       list-devices
  |       burn-image -d 0 -i <firmware.img path>
  |
  |  ! WARNING: Do NOT power off during flash!
  |
  |  [5] API Quick Reference
  |  -----------------------------------------------------------
  |
  |  DVFS Control:
  |    hm_sys_set_dvfs_mode(dev, PERFORMANCE)    Set performance mode (1400MHz)
  |    hm_sys_set_dvfs_mode(dev, ONDEMAND)        Set power-saving mode (auto 200-1400MHz)
  |    hm_sys_get_dvfs_mode(dev, &mode)          Get current DVFS mode (ret=0 ok)
  |
  |  Utilization:
  |    hm_sys_get_ipu_utili_rate(dev)             Get overall IPU load, ret=float (0.0~1.0)
  |    hm_sys_get_ipu_core_utili_rate(dev, cid)   Get per-core load, ret=float (0.0~1.0)
  |
  |  Frequency & Voltage:
  |    hm_sys_get_ipu_frequency(dev, &freq)       Get IPU freq (Hz), ret=0 ok
  |    hm_sys_get_ipu_voltage(dev, &volt)          Get IPU voltage (mV), ret=0 ok
  |
  |  Power & Thermal:
  |    hm_sys_get_temperature(dev, &temp)          Get die temperature (C), ret=0 ok
  |    hm_sys_get_board_power(dev, &power)         Get board power (W), ret=0 ok
  |
  |  Memory:
  |    hm_sys_get_mem_info(dev, &mem)             Get DDR info (total/used/avail MB)
  |    hm_sys_get_ddr_size(dev, &ddr_size)        Get total DDR size (bytes), ret=u32 ok
  |
  |  Device Info:
  |    hm_sys_get_device_info(&info)              Get all device IDs & count
  |    hm_sys_get_device_sn(dev, sn[], len)       Get device serial number
  |    hm_sys_get_device_name(dev, name[], len)   Get device name string
  |    hm_sys_get_computing_power(dev)            Get TOPS rating, ret=int
  |    hm_sys_get_core_count(dev)                 Get IPU core count, ret=int
  |
  |  Version:
  |    hm_sys_get_version(buf, len)               Get SDK runtime version
  |    hm_sys_get_driver_version(buf, len)        Get driver version
  |    hm_sys_get_device_version(dev, buf, len)   Get firmware version on device
  |
  +------------------------------------------------------------------+
)";

    pause_console();
}

// ============================================================
// Feature 6: Power Management
// ============================================================

// Helper: Run hm_smi.exe command and capture output
static std::string run_hm_smi(const std::string& args) {
    std::string result;
    char cmd[512];
    snprintf(cmd, sizeof(cmd),
        "\"C:\\Program Files (x86)\\houmo-drv-xh2_v1.1.0\\tools\\hm_smi\\hm_smi.exe\" %s 2>&1",
        args.c_str());

    FILE* pipe = _popen(cmd, "r");
    if (!pipe) return "";

    char buf[256];
    while (fgets(buf, sizeof(buf), pipe)) {
        result += buf;
    }
    _pclose(pipe);
    return result;
}

// Helper: Parse hm_smi output for specific field
static std::string parse_smi_field(const std::string& output, const std::string& field) {
    size_t pos = output.find(field);
    if (pos == std::string::npos) return "N/A";
    pos += field.length();
    while (pos < output.length() && (output[pos] == ' ' || output[pos] == ':')) pos++;
    size_t end = pos;
    while (end < output.length() && output[end] != '\n' && output[end] != '\r') end++;
    std::string val = output.substr(pos, end - pos);
    // Trim trailing spaces
    while (!val.empty() && (val.back() == ' ' || val.back() == '\t')) val.pop_back();
    return val;
}

static void func_power_management() {
    struct hm_device_info dinfo = {0};
    uint32_t ret = g_api.pfn_get_device_info(&dinfo);
    uint32_t n = dinfo.num_devices ? dinfo.num_devices : ret;
    if (n == 0) { std::cerr << "\n  [ERROR] No device found!\n"; pause_console(); return; }

    int dev = 0;
    if (n > 1) dev = read_int("\n  Select Device ID: ", 0, (int)(n - 1));

    while (true) {
        cls();
        print_title("Power Management");

        // Get info from hm_smi for lock clock status
        std::string smi_out = run_hm_smi("-a -d " + std::to_string(dev));

        std::string dvfs_mode    = parse_smi_field(smi_out, "DVFS_Mode");
        std::string cur_ipu_freq = parse_smi_field(smi_out, "Cur_Ipu_Freq");
        std::string lock_max     = parse_smi_field(smi_out, "Lock_Ipu_Max_Clock");
        std::string lock_min     = parse_smi_field(smi_out, "Lock_Ipu_Min_Clock");
        std::string ipu_load     = parse_smi_field(smi_out, "IPU_Load");
        std::string board_power  = parse_smi_field(smi_out, "Board_Power");

        // Get current metrics via API
        float temp = -1; g_api.pfn_get_temperature(dev, &temp);

        std::cout << "\n";
        print_sub("Power Management Status");
        std::cout << "  DVFS Mode:          " << dvfs_mode << "\n";
        std::cout << "  Current IPU Freq:   " << cur_ipu_freq << "\n";
        std::cout << "  Lock IPU Max Clock: " << lock_max << "\n";
        std::cout << "  Lock IPU Min Clock: " << lock_min << "\n";
        std::cout << "  IPU Load:           " << ipu_load << "\n";
        std::cout << "  Board Power:        " << board_power << "\n";
        std::cout << "  Temperature:        ";
        if (temp >= 0) {
            if (temp >= TEMP_CRITICAL) std::cout << "\033[1;31m" << temp << " C\033[0m [DANGER]\n";
            else if (temp >= TEMP_WARNING) std::cout << "\033[1;33m" << temp << " C\033[0m [WARNING]\n";
            else std::cout << temp << " C\n";
        } else std::cout << "N/A\n";

        print_sub("Actions");
        std::cout << "  [1] Set DVFS Mode (PERFORMANCE / ONDEMAND / POWERLIMIT)\n";
        std::cout << "  [2] Lock IPU Clock Range (Max/Min MHz)\n";
        std::cout << "  [3] Set Power Limit (Watts)\n";
        std::cout << "  [4] Reset to Default Settings\n";
        std::cout << "  [0] Back to Main Menu\n";

        int ch = read_int("\n  Select [0-4]: ", 0, 4);

        if (ch == 0) break;

        switch (ch) {
            case 1: {
                print_sub("Select DVFS Mode");
                std::cout << "  [1] PERFORMANCE  - Fixed 1400 MHz, max throughput\n";
                std::cout << "  [2] ONDEMAND     - Dynamic based on load (200-1400 MHz)\n";
                std::cout << "  [3] POWERLIMIT   - Dynamic based on power limit\n";
                int mode = read_int("\n  Select [1-3]: ", 1, 3);

                std::string cmd;
                if (mode == 1) {
                    cmd = "-g performance -d " + std::to_string(dev);
                } else if (mode == 2) {
                    cmd = "-g ondemand -d " + std::to_string(dev);
                } else {
                    std::cout << "\n  Power Limit mode requires max/min power values.\n";
                    int max_w = read_int("  Enter max power (W): ", 5, 50);
                    int min_w = read_int("  Enter min power (W): ", 3, max_w);
                    cmd = "-g powerlimit -pl " + std::to_string(max_w) + "w " + std::to_string(min_w) + "w -d " + std::to_string(dev);
                }

                std::cout << "\n  Executing: hm_smi " << cmd << "\n";
                std::string result = run_hm_smi(cmd);
                if (result.find("error") != std::string::npos || result.find("Error") != std::string::npos) {
                    std::cout << "  Result: " << result << "\n";
                } else {
                    std::cout << "  DVFS mode changed successfully!\n";
                }
                pause_console();
                break;
            }
            case 2: {
                print_sub("Lock IPU Clock Range");
                std::cout << "  Current: Max=" << lock_max << ", Min=" << lock_min << "\n";
                std::cout << "  Range:   700 - 1400 MHz\n\n";
                int max_mhz = read_int("  Enter max IPU clock (MHz): ", 700, 1400);
                int min_mhz = read_int("  Enter min IPU clock (MHz): ", 700, max_mhz);

                std::string cmd = "-lc " + std::to_string(max_mhz) + " " + std::to_string(min_mhz) + " -d " + std::to_string(dev);
                std::cout << "\n  Executing: hm_smi " << cmd << "\n";
                run_hm_smi(cmd);
                std::cout << "\n  IPU clock range locked to " << min_mhz << " - " << max_mhz << " MHz\n\n";

                // Real-time monitoring after setting
                std::cout << "  --- Real-time Monitoring ---\n";
                std::cout << "  Press ENTER to stop\n\n";
                print_line('-', 60);
                std::cout << "  Time    IPU_Freq      Power    Load     Temp\n";
                print_line('-', 60);

                auto mon_start = std::chrono::steady_clock::now();
                bool mon_done = false;
                while (!mon_done) {
                    // Check if ENTER is waiting in buffer (non-blocking)
                    if (_kbhit()) {
                        char c = _getch();
                        if (c == '\r' || c == '\n') { mon_done = true; break; }
                    }

                    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                        std::chrono::steady_clock::now() - mon_start).count();

                    std::string mon_out = run_hm_smi("-a -d " + std::to_string(dev));
                    std::string mf = parse_smi_field(mon_out, "Cur_Ipu_Freq");
                    std::string mp = parse_smi_field(mon_out, "Board_Power");
                    std::string ml = parse_smi_field(mon_out, "IPU_Load");

                    float temp_m = -1; g_api.pfn_get_temperature(dev, &temp_m);

                    std::cout << "  [" << std::setw(5) << elapsed << "s] "
                              << std::setw(14) << mf << " "
                              << std::setw(9) << mp << " "
                              << std::setw(8) << ml << " ";
                    if (temp_m >= 0) {
                        if (temp_m >= TEMP_CRITICAL) std::cout << "\033[1;31m" << temp_m << " C\033[0m\033[1G\033[2K";
                        else if (temp_m >= TEMP_WARNING) std::cout << "\033[1;33m" << temp_m << " C\033[0m\033[1G\033[2K";
                        else std::cout << temp_m << " C";
                    }
                    std::cout << "\n" << std::flush;
                    Sleep(2000);
                }
                std::cout << "\n  Monitoring stopped.\n";
                pause_console();
                break;
            }
            case 3: {
                print_sub("Set Power Limit");
                std::cout << "  Current Power: " << board_power << "\n";
                std::cout << "  Note: This sets the DVFS mode to POWERLIMIT\n\n";
                int max_w = read_int("  Enter max power limit (W): ", 5, 50);
                int min_w = read_int("  Enter min power limit (W): ", 3, max_w);

                std::string cmd = "-g powerlimit -pl " + std::to_string(max_w) + "w " + std::to_string(min_w) + "w -d " + std::to_string(dev);
                std::cout << "\n  Executing: hm_smi " << cmd << "\n";
                run_hm_smi(cmd);
                std::cout << "\n  Power limit set to " << min_w << " - " << max_w << " W\n\n";

                // Real-time monitoring after setting
                std::cout << "  --- Real-time Monitoring ---\n";
                std::cout << "  Press ENTER to stop\n\n";
                print_line('-', 60);
                std::cout << "  Time    IPU_Freq      Power      Load     Temp\n";
                print_line('-', 60);

                auto mon_start = std::chrono::steady_clock::now();
                bool mon_done = false;
                while (!mon_done) {
                    if (_kbhit()) {
                        char c = _getch();
                        if (c == '\r' || c == '\n') { mon_done = true; break; }
                    }

                    auto elapsed = std::chrono::duration_cast<std::chrono::seconds>(
                        std::chrono::steady_clock::now() - mon_start).count();

                    std::string mon_out = run_hm_smi("-a -d " + std::to_string(dev));
                    std::string mf = parse_smi_field(mon_out, "Cur_Ipu_Freq");
                    std::string mp = parse_smi_field(mon_out, "Board_Power");
                    std::string ml = parse_smi_field(mon_out, "IPU_Load");
                    std::string dvfs_m = parse_smi_field(mon_out, "DVFS_Mode");

                    float temp_m = -1; g_api.pfn_get_temperature(dev, &temp_m);

                    std::cout << "  [" << std::setw(5) << elapsed << "s] "
                              << std::setw(14) << mf << " "
                              << std::setw(11) << mp << " "
                              << std::setw(8) << ml << " ";
                    if (temp_m >= 0) {
                        if (temp_m >= TEMP_CRITICAL) std::cout << "\033[1;31m" << temp_m << " C\033[0m\033[1G\033[2K";
                        else if (temp_m >= TEMP_WARNING) std::cout << "\033[1;33m" << temp_m << " C\033[0m\033[1G\033[2K";
                        else std::cout << temp_m << " C";
                    }
                    std::cout << "\n" << std::flush;
                    Sleep(2000);
                }
                std::cout << "\n  Monitoring stopped.\n";
                pause_console();
                break;
            }
            case 4: {
                print_sub("Reset to Defaults");
                std::cout << "  This will reset DVFS to PERFORMANCE and unlock clocks.\n";
                std::cout << "  Continue? [y/N]: ";
                char c; std::cin >> c;
                std::cin.ignore(std::numeric_limits<std::streamsize>::max(), '\n');
                if (c == 'y' || c == 'Y') {
                    run_hm_smi("-g performance -d " + std::to_string(dev));
                    run_hm_smi("-lc 1400 700 -d " + std::to_string(dev));
                    std::cout << "\n  Settings reset to default (PERFORMANCE, 700-1400 MHz)\n";
                }
                pause_console();
                break;
            }
        }
    }
}

// ============================================================
// Main
// ============================================================

int main(int argc, char** argv) {
    (void)argc; (void)argv;

    std::ios::sync_with_stdio(false);
    std::cin.tie(nullptr);

    cls();
    std::cout << "\n  Loading Houmo DNPU driver...\n\n";

    const wchar_t* dllPaths[] = {
        L"libhal_xh2a.dll",
        L"C:\\Users\\3-64\\source\\repos\\Project1\\Project1\\libhal_xh2a.dll",
        L"C:\\Program Files (x86)\\houmo-drv-xh2_v1.1.0\\libhal_xh2a.dll",
    };

    bool loaded = false;
    for (auto path : dllPaths) {
        std::wcout << L"  Trying: " << path << L" ... ";
        if (g_api.load(path)) {
            std::wcout << L"OK!\n";
            loaded = true;
            break;
        }
        std::cout << "FAIL\n";
    }

    if (!loaded) {
        std::cerr << "\n  [ERROR] Cannot load libhal_xh2a.dll!\n";
        std::cerr << "  Make sure the DLL is in the same directory as this exe.\n";
        std::cerr << "\n  Press ENTER to exit...";
        std::cin.get();
        return -1;
    }

    Sleep(1500);

    struct hm_device_info dinfo = {0};
    uint32_t ret = g_api.pfn_get_device_info(&dinfo);
    uint32_t n = dinfo.num_devices ? dinfo.num_devices : ret;

    if (n == 0) {
        cls();
        std::cerr << "\n  =========================================\n";
        std::cerr << "  ! No DNPU device found\n";
        std::cerr << "  =========================================\n";
        std::cerr << "\n  Checklist:\n";
        std::cerr << "  1. Is the DNPU device connected and powered?\n";
        std::cerr << "  2. Is the driver (houmo-drv-xh2_v1.1.0) installed?\n";
        std::cerr << "  3. Is the device being used by another program?\n";
        std::cerr << "\n  Press ENTER to exit...";
        std::cin.get();
        return -1;
    }

    std::cout << "  Found " << n << " DNPU device(s)\n\n";
    Sleep(1000);

    while (true) {
        show_menu();
        int ch = read_int("  Select [0-6]: ", 0, 6);

        switch (ch) {
            case 1: func_show_devices();      break;
            case 2: func_set_dvfs();          break;
            case 3: func_realtime_monitor();  break;
            case 4: func_smart_scheduler();    break;
            case 5: func_param_ref();          break;
            case 6: func_power_management();   break;
            case 0:
                std::cout << "\n  Goodbye!\n\n";
                return 0;
        }
    }

    return 0;
}
