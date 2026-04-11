/*
 * MSRDriver.h - Intel MSR Access Kernel Driver
 * =============================================
 * Based on: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx
 * =============================================
 * 
 * This header defines the MSR registers and IOCTL codes for
 * communicating with the kernel driver.
 */

#pragma once

#include <windows.h>

#ifdef __cplusplus
extern "C" {
#endif

// =============================================
// MSR Register Addresses (from PPT slides)
// =============================================

// Performance and Power Management MSRs
#define MSR_IA32_ENERGY_PERF_BIAS     0x1B0   // Slide 5: Energy Performance Bias
#define MSR_IA32_PERF_CTL              0x198   // Slide 6: Performance Control
#define MSR_IA32_PERF_STATUS           0x199   // Slide 6: Current P-state (read-only)
#define MSR_IA32_MISC_ENABLE           0x1A0   // Slide 6: Turbo Boost control
#define MSR_IA32_CLOCK_MODULATION      0x19A   // Slide 18: Clock Modulation (T-state)
#define MSR_IA32_APERF                 0xE7    // Slide 7: Actual Performance Counter
#define MSR_IA32_MPERF                 0xE8    // Slide 7: Maximum Performance Counter

// Uncore Frequency Scaling (Slide 10)
#define MSR_UNCORE_RATIO_LIMIT         0x620

// RAPL Registers (Slides 54-58)
#define MSR_RAPL_POWER_UNIT             0x606   // Units: power, energy, time
#define MSR_PKG_POWER_LIMIT             0x610   // Package power limit
#define MSR_PKG_ENERGY_STATUS           0x611   // Package energy consumed
#define MSR_PKG_POWER_INFO              0x614   // Package power range info
#define MSR_PKG_PERF_STATUS             0x613   // Package throttled time
#define MSR_PP0_POWER_LIMIT             0x638   // CPU cores power limit
#define MSR_PP0_ENERGY_STATUS           0x639   // CPU cores energy
#define MSR_PP0_POLICY                  0x63A   // PP0 power policy
#define MSR_PP1_POWER_LIMIT             0x640   // Uncore/Graphics power limit
#define MSR_PP1_ENERGY_STATUS           0x641   // PP1 energy
#define MSR_DRAM_POWER_LIMIT            0x618   // DRAM power limit
#define MSR_DRAM_ENERGY_STATUS          0x619   // DRAM energy
#define MSR_DRAM_POWER_INFO             0x61C   // DRAM power info

// =============================================
// IOCTL Codes
// =============================================

#define FILE_DEVICE_MSR_DRIVER         0x8000

#define IOCTL_MSR_READ             CTL_CODE(FILE_DEVICE_MSR_DRIVER, 0x800, METHOD_BUFFERED, FILE_ANY_ACCESS)
#define IOCTL_MSR_WRITE            CTL_CODE(FILE_DEVICE_MSR_DRIVER, 0x801, METHOD_BUFFERED, FILE_ANY_ACCESS)
#define IOCTL_MSR_GET_INFO         CTL_CODE(FILE_DEVICE_MSR_DRIVER, 0x802, METHOD_BUFFERED, FILE_ANY_ACCESS)
#define IOCTL_MSR_LIST_SUPPORTED   CTL_CODE(FILE_DEVICE_MSR_DRIVER, 0x803, METHOD_BUFFERED, FILE_ANY_ACCESS)

// =============================================
// Data Structures
// =============================================

#pragma pack(push, 1)

// MSR Read/Write Request
typedef struct _MSR_REQUEST {
    ULONG ProcessorNumber;    // CPU core to access
    ULONG MsrAddress;        // MSR register address
    ULONGLONG Value;         // Value read/written
} MSR_REQUEST, *PMSR_REQUEST;

// MSR Information
typedef struct _MSR_INFO {
    ULONG MsrAddress;                    // Register address
    WCHAR Name[64];                      // Register name
    WCHAR Description[256];              // Description from Intel manual
    BOOLEAN IsReadOnly;                  // Read-only flag
    BOOLEAN IsSupported;                 // Supported on current CPU
    ULONGLONG CurrentValue;              // Current value
    ULONGLONG MinValue;                  // Minimum value
    ULONGLONG MaxValue;                  // Maximum value
} MSR_INFO, *PMSR_INFO;

// Supported MSR List Entry
typedef struct _SUPPORTED_MSR {
    ULONG MsrAddress;
    WCHAR Name[32];
    WCHAR Description[128];
} SUPPORTED_MSR, *PSUPPORTED_MSR;

#pragma pack(pop)

// =============================================
// Energy Performance Bias Values (Slide 5)
// =============================================

typedef enum _ENERGY_PERF_BIAS {
    EPB_HIGHEST_PERFORMANCE = 0,     // 0 = Highest Performance
    EPB_BALANCE = 6,                // 6 = Balance (default)
    EPB_MAX_ENERGY_SAVINGS = 15      // 15 = Maximum Energy Savings
} ENERGY_PERF_BIAS;

// =============================================
// Helper Macros
// =============================================

// Extract bits from MSR value
#define MSR_GET_BITS(Value, Start, Length) \
    ((Value >> Start) & ((1ULL << Length) - 1))

// Set bits in MSR value
#define MSR_SET_BITS(Value, Start, Length, NewValue) \
    (Value & ~(((1ULL << Length) - 1) << Start)) | \
    (((NewValue) & ((1ULL << Length) - 1)) << Start)

// RAPL Power Unit conversion (Slide 55)
#define RAPL_POWER_UNIT(power_unit_msr) \
    (1.0 / (1ULL << (power_unit_msr & 0xF)))

#define RAPL_ENERGY_UNIT(power_unit_msr) \
    (1.0 / (1ULL << ((power_unit_msr >> 8) & 0x1F)))

#define RAPL_TIME_UNIT(power_unit_msr) \
    (1.0 / (1ULL << ((power_unit_msr >> 16) & 0xF)))

// =============================================
// Function Declarations
// =============================================

// User-mode API
HANDLE WINAPI MsrOpenDriver(VOID);
BOOL WINAPI MsrCloseDriver(HANDLE hDevice);
BOOL WINAPI MsrReadMsr(HANDLE hDevice, ULONG ProcessorNumber, ULONG MsrAddress, PULONGLONG Value);
BOOL WINAPI MsrWriteMsr(HANDLE hDevice, ULONG ProcessorNumber, ULONG MsrAddress, ULONGLONG Value);
BOOL WINAPI MsrGetInfo(HANDLE hDevice, ULONG MsrAddress, PMSR_INFO Info);
BOOL WINAPI MsrListSupported(HANDLE hDevice, PSUPPORTED_MSR List, PULONG Count);

// Convenience functions
BOOL WINAPI MsrSetEnergyPerfBias(HANDLE hDevice, ULONG ProcessorNumber, UCHAR Bias);
UCHAR WINAPI MsrGetEnergyPerfBias(HANDLE hDevice, ULONG ProcessorNumber);
BOOL WINAPI MsrSetTurboBoost(HANDLE hDevice, ULONG ProcessorNumber, BOOLEAN Enable);
BOOLEAN WINAPI MsrGetTurboBoost(HANDLE hDevice, ULONG ProcessorNumber);
BOOL WINAPI MsrSetPackagePowerLimit(HANDLE hDevice, ULONG ProcessorNumber, ULONG PowerLimitMw, ULONG TimeWindowMs);
BOOL WINAPI MsrGetPackagePowerLimit(HANDLE hDevice, ULONG ProcessorNumber, PULONG PowerLimitMw, PULONG TimeWindowMs);

#ifdef __cplusplus
}
#endif
