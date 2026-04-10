// ipf_wrapper.h
// C interface for LenovoIPFV2.dll (and LenovoIPF.dll fallback)
// Used by LenovoToolkit Go backend via cgo.

#pragma once

#include <windows.h>

#ifdef __cplusplus
extern "C" {
#endif

// ── Connection ──────────────────────────────────────────────────────────────

// Connects to IPF V2 (LenovoIPFV2.dll). Returns version (1, 2) or 0 on failure.
int IPF_Connect(void);

// Returns current IPF version: 1=V1(ESIF), 2=V2(JSON), 0=not connected.
int IPF_GetVersion(void);

// Disconnects from IPF.
void IPF_Disconnect(void);

// ── Read values (all return milliwatts for power, 0 on failure) ───────────

// IPF_SystemPower: System/CPU package power from RAPL via IPF SDK.
unsigned int IPF_GetSystemPower_mW(void);

// CPU temperature in centiKelvin (e.g. 3230 = 32.30°C above absolute zero → ~50°C).
unsigned int IPF_GetCpuTemp_cK(void);

// PL1: Sustained power limit (mW).
unsigned int IPF_GetPL1_mW(void);

// PL2: Burst/turbo power limit (mW).
unsigned int IPF_GetPL2_mW(void);

// PL4: Peak power limit (mW). May return 0 if hardware does not support it.
unsigned int IPF_GetPL4_mW(void);

// ── Convenience: all PL values in one call ───────────────────────────────
// Sets pl[0]=PL1, pl[1]=PL2, pl[2]=PL4 (all in mW). Pass array of at least 3.
void IPF_GetAllPL_mW(unsigned int pl[3]);

// ── MSR reads (Intel CPU) ──────────────────────────────────────────────────
// Read a Model Specific Register by index, returns full 64-bit value in *eax,*edx.
// Returns 1 on success, 0 on failure (no MSR support or access denied).
// Caller provides eax/edx pointers (pass NULL to ignore).
int IPF_ReadMSR(unsigned int msrIndex, unsigned int* eaxOut, unsigned int* edxOut);

// ── Convenience MSR reads (Intel) ─────────────────────────────────────────
// EPP (Energy Performance Preference): MSR 0x1B0 IA32_ENERGY_PERF_BIAS, bits 0-3
// Returns EPP value 0-15 (0=performance, 15=battery), or 0xFFFF on failure.
unsigned int IPF_GetEPP(void);

// EPP for E-Core / secondary EPP: MSR 0x1B0 upper bits or MSR 0x644
// Returns EPP1 value, or 0xFFFF on failure.
unsigned int IPF_GetEPP1(void);

// Max frequency limit from HWP MSR 0x770, returns MHz, or 0 on failure.
unsigned int IPF_GetFrequencyLimit_MHz(void);

// Hysteresis settings (Hetero Increase/Decrease): values in platform units.
// Returns raw value from MSR, or 0 on failure.
unsigned int IPF_GetHeteroInc(void);
unsigned int IPF_GetHeteroDec(void);

// Soft Park Latency from MSR, returns raw value in 100ns units.
unsigned int IPF_GetSoftParkLatency(void);

// ── IPFV2 Current Gear (EPOT) ─────────────────────────────────────────────
// Returns the current EPOT/Gear level (0-9) from LenovoIPFV2.dll.
// This reads directly from the firmware/IPF layer, same as ML_Scenario.
// Returns -1 if not available or on error.
int IPF_GetCurrentGear(void);

// ── DLL path configuration ───────────────────────────────────────────────
// Sets the directory where LenovoIPFV2.dll / LenovoIPF.dll / WinMSRIO.dll are located.
// Call before IPF_Connect(). Pass empty string to use default search.
void IPF_SetDllPath(const char* path);

#ifdef __cplusplus
}
#endif
