/**
 * Houmo DNPU IPU Stress Test
 * ============================
 * Simple stress test that continuously triggers IPU kernels
 * to create artificial load for power/frequency monitoring.
 *
 * This doesn't need any model - it just exercises the IPU compute units.
 *
 * Compile (VS Developer Command Prompt):
 *   cl /EHsc /W4 /O2 ipu_stress.cpp /Fe:ipu_stress.exe user32.lib
 *
 * Run: ipu_stress.exe [num_kernels] [iterations]
 *   num_kernels: Number of concurrent kernels (default: 16)
 *   iterations:  Loop count (default: 1000)
 */

#include <windows.h>
#include <stdio.h>
#include <stdint.h>
#include <iostream>
#include <chrono>
#include <vector>
#include <thread>
#include <atomic>

#define WIN32_LEAN_AND_MEAN
#include <windows.h>

// Load DLL
static HMODULE g_dll = NULL;

#define HM_DECL(ret, name, args) \
    typedef ret (WINAPI *name##_t) args; \
    static name##_t name = NULL;

HM_DECL(int, hm_get_device_num, (void))
HM_DECL(int, hm_get_device_info, (void* info))
HM_DECL(int, hm_get_ipu_utili_rate, (int dev_id))
HM_DECL(int, hm_get_temperature, (int dev_id, float* temp))
HM_DECL(int, hm_get_board_power, (int dev_id, float* power))
HM_DECL(int, hm_get_dvfs_mode, (int dev_id, int* mode))
HM_DECL(int, hm_set_dvfs_mode, (int dev_id, int mode))
HM_DECL(int, hm_ipu_kernel_group_create, (int dev_id, int group_id))
HM_DECL(int, hm_ipu_kernel_group_destroy, (int dev_id, int group_id))
HM_DECL(int, hm_ipu_kernel_group_run, (int dev_id, int group_id, int kernel_id, void* inputs, void* outputs))
HM_DECL(int, hm_ipu_kernel_group_sync, (int dev_id, int group_id))
HM_DECL(int, hm_dvfs_get_ipu_freq, (int dev_id, uint64_t* freq))
HM_DECL(int, hm_reset_device, (int dev_id))
HM_DECL(int, hm_get_error_code, (int dev_id))

static bool load_dll() {
    const wchar_t* paths[] = {
        L"C:\\Program Files (x86)\\houmo-drv-xh2_v1.1.0\\hal\\lib\\libhal_xh2a.dll",
        L"libhal_xh2a.dll",
    };
    for (auto p : paths) {
        g_dll = LoadLibraryW(p);
        if (g_dll) {
            std::wcout << L"  DLL: " << p << L"\n";
            #define LOAD(sym) sym = (sym##_t)GetProcAddress(g_dll, #sym); \
                if (!sym) { std::cerr << "  Missing: " #sym << "\n"; return false; }
            LOAD(hm_get_device_num);
            LOAD(hm_get_device_info);
            LOAD(hm_get_ipu_utili_rate);
            LOAD(hm_get_temperature);
            LOAD(hm_get_board_power);
            LOAD(hm_get_dvfs_mode);
            LOAD(hm_set_dvfs_mode);
            LOAD(hm_ipu_kernel_group_create);
            LOAD(hm_ipu_kernel_group_destroy);
            LOAD(hm_ipu_kernel_group_run);
            LOAD(hm_ipu_kernel_group_sync);
            LOAD(hm_dvfs_get_ipu_freq);
            LOAD(hm_reset_device);
            LOAD(hm_get_error_code);
            return true;
        }
    }
    return false;
}

static std::atomic<bool> g_running{true};

static void monitor_thread(int dev_id, int interval_ms) {
    while (g_running) {
        float power = -1;
        float temp = -1;
        int ipu_util = -1;
        uint64_t freq = 0;
        int dvfs = 0;

        hm_get_board_power(dev_id, &power);
        hm_get_temperature(dev_id, &temp);
        ipu_util = hm_get_ipu_utili_rate(dev_id);
        hm_dvfs_get_ipu_freq(dev_id, &freq);
        hm_get_dvfs_mode(dev_id, &dvfs);

        const char* dvfs_name = "UNKNOWN";
        if (dvfs == 0) dvfs_name = "PERFORMANCE";
        else if (dvfs == 1) dvfs_name = "ONDEMAND";
        else if (dvfs == 2) dvfs_name = "POWERLIMIT";

        printf("[MON] Power=%6.2fW  IPU_Util=%3d%%  Freq=%6.0fMHz  Temp=%5.1fC  Mode=%s\r",
               power, ipu_util, freq / 1e6, temp, dvfs_name);
        fflush(stdout);
        Sleep(interval_ms);
    }
}

static void print_banner() {
    printf("\n");
    printf("  ===========================================================\n");
    printf("  Houmo DNPU - IPU Stress Test\n");
    printf("  ===========================================================\n");
    printf("\n");
}

int main(int argc, char** argv) {
    print_banner();

    // Parse args
    int num_kernels = 16;
    int iterations = 1000;
    if (argc >= 2) num_kernels = atoi(argv[1]);
    if (argc >= 3) iterations = atoi(argv[2]);

    printf("  Loading driver...\n\n");
    if (!load_dll()) {
        std::cerr << "  [ERROR] Cannot load libhal_xh2a.dll\n";
        return -1;
    }

    int num_devs = hm_get_device_num();
    if (num_devs <= 0) {
        std::cerr << "  [ERROR] No DNPU device found!\n";
        return -1;
    }
    printf("  Found %d DNPU device(s)\n\n", num_devs);

    int dev_id = 0;

    // Set to PERFORMANCE mode for max load
    printf("  Setting DVFS mode to PERFORMANCE...\n");
    hm_set_dvfs_mode(dev_id, 0);

    // Start monitor thread
    std::thread mon(monitor_thread, dev_id, 1000);

    printf("  Starting IPU stress test...\n");
    printf("  Kernels: %d  Iterations: %d\n\n", num_kernels, iterations);

    auto start = std::chrono::steady_clock::now();

    // Try to create kernel groups and run them
    // This will create load on the IPU
    for (int i = 0; i < iterations && g_running; i++) {
        // Create a group
        int group_id = i % num_kernels;
        int ret = hm_ipu_kernel_group_create(dev_id, group_id);
        if (ret != 0) {
            // Group might already exist or API not available
            // Try to reset the device periodically
            if (i % 100 == 0) {
                hm_reset_device(dev_id);
            }
            continue;
        }

        // Run with empty input (exercises IPU control path)
        // This may or may not work depending on firmware
        uint8_t dummy_input[64] = {0};
        uint8_t dummy_output[64] = {0};
        hm_ipu_kernel_group_run(dev_id, group_id, 0, dummy_input, dummy_output);
        hm_ipu_kernel_group_sync(dev_id, group_id);
        hm_ipu_kernel_group_destroy(dev_id, group_id);

        if (i % 100 == 0) {
            printf("  Progress: %d / %d iterations\r", i, iterations);
            fflush(stdout);
        }
    }

    g_running = false;
    mon.join();

    auto end = std::chrono::steady_clock::now();
    double elapsed = std::chrono::duration<double>(end - start).count();

    printf("\n\n");
    printf("  ===========================================================\n");
    printf("  Stress Test Complete!\n");
    printf("  Elapsed time: %.2f seconds\n", elapsed);
    printf("  Iterations:  %d\n", iterations);
    printf("  ===========================================================\n");
    printf("\n");

    if (g_dll) FreeLibrary(g_dll);
    return 0;
}
