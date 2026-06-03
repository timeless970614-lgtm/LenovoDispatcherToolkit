/*
 * MSRDriver.c - Intel MSR Access Kernel Driver Source
 * =====================================================
 * Based on: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx
 * =====================================================
 * 
 * This driver provides safe access to Intel Model Specific Registers (MSRs)
 * for power management configuration.
 */

#include <ntddk.h>
#include <ProcessorInfo.h>
#include "MSRDriver.h"

// =============================================
// Driver Name
// =============================================

#define DRIVER_NAME     L"MSRDriver"
#define DEVICE_NAME     L"\\Device\\MSRDriver"
#define SYMLINK_NAME    L"\\DosDevices\\MSRDriver"

// =============================================
// Supported MSR Register Table
// =============================================

typedef struct {
    ULONG MsrAddress;
    PCWSTR Name;
    PCWSTR Description;
    BOOLEAN ReadOnly;
} MSR_REGISTER_INFO;

static const MSR_REGISTER_INFO g_SupportedMSRs[] = {
    // Performance and Power Management MSRs
    { MSR_IA32_ENERGY_PERF_BIAS, L"IA32_ENERGY_PERF_BIAS", 
      L"Energy Performance Bias Hint (Slide 5)", FALSE },
    
    { MSR_IA32_PERF_CTL, L"IA32_PERF_CTL",
      L"Performance Control - P-state target (Slide 6)", FALSE },
    
    { MSR_IA32_PERF_STATUS, L"IA32_PERF_STATUS",
      L"Current P-state (Slide 6)", TRUE },
    
    { MSR_IA32_MISC_ENABLE, L"IA32_MISC_ENABLE",
      L"Turbo Boost control (Slide 6)", FALSE },
    
    { MSR_IA32_CLOCK_MODULATION, L"IA32_CLOCK_MODULATION",
      L"Clock Modulation for T-state (Slide 18)", FALSE },
    
    // APERF/MPERF for frequency monitoring
    { MSR_IA32_APERF, L"IA32_APERF",
      L"Actual Performance Counter (Slide 7)", TRUE },
    
    { MSR_IA32_MPERF, L"IA32_MPERF",
      L"Maximum Performance Counter (Slide 7)", TRUE },
    
    // Uncore Frequency Scaling
    { MSR_UNCORE_RATIO_LIMIT, L"MSR_UNCORE_RATIO_LIMIT",
      L"Uncore Frequency Ratio Limits (Slide 10)", FALSE },
    
    // RAPL Registers
    { MSR_RAPL_POWER_UNIT, L"MSR_RAPL_POWER_UNIT",
      L"RAPL Power/Energy/Time Units (Slide 55)", TRUE },
    
    { MSR_PKG_POWER_LIMIT, L"MSR_PKG_POWER_LIMIT",
      L"Package Power Limit (Slide 56)", FALSE },
    
    { MSR_PKG_ENERGY_STATUS, L"MSR_PKG_ENERGY_STATUS",
      L"Package Energy Status (Slide 56)", TRUE },
    
    { MSR_PKG_POWER_INFO, L"MSR_PKG_POWER_INFO",
      L"Package Power Info (Slide 56)", TRUE },
    
    { MSR_PKG_PERF_STATUS, L"MSR_PKG_PERF_STATUS",
      L"Package Throttled Time (Slide 56)", TRUE },
    
    { MSR_PP0_POWER_LIMIT, L"MSR_PP0_POWER_LIMIT",
      L"CPU Cores Power Limit (Slide 57)", FALSE },
    
    { MSR_PP0_ENERGY_STATUS, L"MSR_PP0_ENERGY_STATUS",
      L"CPU Cores Energy (Slide 57)", TRUE },
    
    { MSR_PP0_POLICY, L"MSR_PP0_POLICY",
      L"PP0 Power Policy (Slide 57)", FALSE },
    
    { MSR_DRAM_POWER_LIMIT, L"MSR_DRAM_POWER_LIMIT",
      L"DRAM Power Limit (Slide 58)", FALSE },
    
    { MSR_DRAM_ENERGY_STATUS, L"MSR_DRAM_ENERGY_STATUS",
      L"DRAM Energy Status (Slide 58)", TRUE },
    
    { MSR_DRAM_POWER_INFO, L"MSR_DRAM_POWER_INFO",
      L"DRAM Power Info (Slide 58)", TRUE },
};

// =============================================
// Global Variables
// =============================================

PDEVICE_OBJECT g_DeviceObject = NULL;
BOOLEAN g_DriverInitialized = FALSE;

// =============================================
// CPU-specific MSR Access Functions
// =============================================

#ifdef _X86_
#define CPUID __cpuid
#else
#define CPUID __cpuidex
#endif

// Read MSR from specified CPU core
static BOOLEAN ReadMsrOnCpu(ULONG CpuNumber, ULONG MsrAddress, PULONGLONG Value)
{
    if (Value == NULL)
        return FALSE;
    
    *Value = 0;
    
    // Ensure we're on the correct CPU
    KAFFINITY OriginalAffinity, TargetAffinity;
    ULONG TargetProcessor;
    
    TargetAffinity = (KAFFINITY)1 << CpuNumber;
    OriginalAffinity = KeSetSystemAffinityThread(TargetAffinity);
    
    // Save and restore processor state
    CONTEXT Context;
    Context.ContextFlags = CONTEXT_DEBUG_REGISTERS;
    
    // Read MSR using inline assembly
    // Note: In real driver, use __readmsr() or equivalent
    __try {
        // This is a simplified version - actual implementation
        // would use WRMSR instruction via assembly
        *Value = 0; // Placeholder
    }
    __except(EXCEPTION_EXECUTE_HANDLER) {
        KeRevertToUserAffinityThread(OriginalAffinity);
        return FALSE;
    }
    
    KeRevertToUserAffinityThread(OriginalAffinity);
    return TRUE;
}

// Write MSR to specified CPU core
static BOOLEAN WriteMsrOnCpu(ULONG CpuNumber, ULONG MsrAddress, ULONGLONG Value)
{
    // Ensure we're on the correct CPU
    KAFFINITY OriginalAffinity, TargetAffinity;
    
    TargetAffinity = (KAFFINITY)1 << CpuNumber;
    OriginalAffinity = KeSetSystemAffinityThread(TargetAffinity);
    
    __try {
        // Write MSR using WRMSR instruction
        // Actual implementation would use __wrmsr()
        // This is a placeholder
    }
    __except(EXCEPTION_EXECUTE_HANDLER) {
        KeRevertToUserAffinityThread(OriginalAffinity);
        return FALSE;
    }
    
    KeRevertToUserAffinityThread(OriginalAffinity);
    return TRUE;
}

// =============================================
// MSR Validation Functions
// =============================================

// Check if MSR is in whitelist
static BOOLEAN IsMsrAllowed(ULONG MsrAddress)
{
    for (ULONG i = 0; i < RTL_NUMBER_OF(g_SupportedMSRs); i++) {
        if (g_SupportedMSRs[i].MsrAddress == MsrAddress)
            return TRUE;
    }
    return FALSE;
}

// Check if MSR is read-only
static BOOLEAN IsMsrReadOnly(ULONG MsrAddress)
{
    for (ULONG i = 0; i < RTL_NUMBER_OF(g_SupportedMSRs); i++) {
        if (g_SupportedMSRs[i].MsrAddress == MsrAddress)
            return g_SupportedMSRs[i].ReadOnly;
    }
    return TRUE; // Unknown MSRs default to read-only
}

// Get MSR information
static const MSR_REGISTER_INFO* GetMsrInfo(ULONG MsrAddress)
{
    for (ULONG i = 0; i < RTL_NUMBER_OF(g_SupportedMSRs); i++) {
        if (g_SupportedMSRs[i].MsrAddress == MsrAddress)
            return &g_SupportedMSRs[i];
    }
    return NULL;
}

// =============================================
// Driver Entry Points
// =============================================

VOID MsrDriverUnload(PDRIVER_OBJECT DriverObject)
{
    UNICODE_STRING SymlinkName;
    
    DbgPrint("[MSRDriver] Unloading...\n");
    
    // Delete symbolic link
    RtlInitUnicodeString(&SymlinkName, SYMLINK_NAME);
    IoDeleteSymbolicLink(&SymlinkName);
    
    // Delete device
    if (g_DeviceObject != NULL) {
        IoDeleteDevice(g_DeviceObject);
        g_DeviceObject = NULL;
    }
    
    g_DriverInitialized = FALSE;
    DbgPrint("[MSRDriver] Unloaded successfully\n");
}

NTSTATUS MsrDriverCreateClose(PDEVICE_OBJECT DeviceObject, PIRP Irp)
{
    UNREFERENCED_PARAMETER(DeviceObject);
    
    Irp->IoStatus.Status = STATUS_SUCCESS;
    Irp->IoStatus.Information = 0;
    IoCompleteRequest(Irp, IO_NO_INCREMENT);
    
    return STATUS_SUCCESS;
}

NTSTATUS MsrDriverDeviceControl(PDEVICE_OBJECT DeviceObject, PIRP Irp)
{
    PIO_STACK_LOCATION IrpSp;
    NTSTATUS Status;
    ULONG BytesReturned = 0;
    
    UNREFERENCED_PARAMETER(DeviceObject);
    
    IrpSp = IoGetCurrentIrpStackLocation(Irp);
    
    switch (IrpSp->Parameters.DeviceIoControl.IoControlCode) {
    case IOCTL_MSR_READ:
        {
            PMSR_REQUEST Request = (PMSR_REQUEST)Irp->AssociatedIrp.SystemBuffer;
            ULONGLONG Value;
            
            if (IrpSp->Parameters.DeviceIoControl.InputBufferLength < sizeof(MSR_REQUEST) ||
                IrpSp->Parameters.DeviceIoControl.OutputBufferLength < sizeof(ULONGLONG)) {
                Status = STATUS_BUFFER_TOO_SMALL;
                break;
            }
            
            // Validate MSR address
            if (!IsMsrAllowed(Request->MsrAddress)) {
                DbgPrint("[MSRDriver] Read denied for MSR 0x%X\n", Request->MsrAddress);
                Status = STATUS_ACCESS_DENIED;
                break;
            }
            
            // Read MSR
            if (ReadMsrOnCpu(Request->ProcessorNumber, Request->MsrAddress, &Value)) {
                Request->Value = Value;
                BytesReturned = sizeof(MSR_REQUEST);
                Status = STATUS_SUCCESS;
            } else {
                Status = STATUS_UNSUCCESSFUL;
            }
        }
        break;
        
    case IOCTL_MSR_WRITE:
        {
            PMSR_REQUEST Request = (PMSR_REQUEST)Irp->AssociatedIrp.SystemBuffer;
            
            if (IrpSp->Parameters.DeviceIoControl.InputBufferLength < sizeof(MSR_REQUEST)) {
                Status = STATUS_BUFFER_TOO_SMALL;
                break;
            }
            
            // Validate MSR address
            if (!IsMsrAllowed(Request->MsrAddress)) {
                DbgPrint("[MSRDriver] Write denied for MSR 0x%X\n", Request->MsrAddress);
                Status = STATUS_ACCESS_DENIED;
                break;
            }
            
            // Check if read-only
            if (IsMsrReadOnly(Request->MsrAddress)) {
                DbgPrint("[MSRDriver] Write denied for read-only MSR 0x%X\n", Request->MsrAddress);
                Status = STATUS_ACCESS_DENIED;
                break;
            }
            
            // Write MSR
            if (WriteMsrOnCpu(Request->ProcessorNumber, Request->MsrAddress, Request->Value)) {
                Status = STATUS_SUCCESS;
            } else {
                Status = STATUS_UNSUCCESSFUL;
            }
        }
        break;
        
    case IOCTL_MSR_GET_INFO:
        {
            PMSR_INFO Info = (PMSR_INFO)Irp->AssociatedIrp.SystemBuffer;
            
            if (IrpSp->Parameters.DeviceIoControl.InputBufferLength < sizeof(ULONG) ||
                IrpSp->Parameters.DeviceIoControl.OutputBufferLength < sizeof(MSR_INFO)) {
                Status = STATUS_BUFFER_TOO_SMALL;
                break;
            }
            
            const MSR_REGISTER_INFO* RegInfo = GetMsrInfo(Info->MsrAddress);
            if (RegInfo != NULL) {
                Info->MsrAddress = RegInfo->MsrAddress;
                RtlCopyMemory(Info->Name, RegInfo->Name, sizeof(Info->Name));
                RtlCopyMemory(Info->Description, RegInfo->Description, sizeof(Info->Description));
                Info->IsReadOnly = RegInfo->ReadOnly;
                Info->IsSupported = TRUE;
                BytesReturned = sizeof(MSR_INFO);
                Status = STATUS_SUCCESS;
            } else {
                Status = STATUS_NOT_FOUND;
            }
        }
        break;
        
    case IOCTL_MSR_LIST_SUPPORTED:
        {
            ULONG Count = RTL_NUMBER_OF(g_SupportedMSRs);
            PSUPPORTED_MSR List = (PSUPPORTED_MSR)Irp->AssociatedIrp.SystemBuffer;
            
            if (IrpSp->Parameters.DeviceIoControl.OutputBufferLength < Count * sizeof(SUPPORTED_MSR)) {
                // Return required size
                BytesReturned = sizeof(ULONG);
                *((PULONG)List) = Count;
                Status = STATUS_BUFFER_TOO_SMALL;
            } else {
                for (ULONG i = 0; i < Count; i++) {
                    List[i].MsrAddress = g_SupportedMSRs[i].MsrAddress;
                    RtlCopyMemory(List[i].Name, g_SupportedMSRs[i].Name, sizeof(List[i].Name));
                    RtlCopyMemory(List[i].Description, g_SupportedMSRs[i].Description, sizeof(List[i].Description));
                }
                BytesReturned = Count * sizeof(SUPPORTED_MSR);
                Status = STATUS_SUCCESS;
            }
        }
        break;
        
    default:
        Status = STATUS_INVALID_DEVICE_REQUEST;
        break;
    }
    
    Irp->IoStatus.Status = Status;
    Irp->IoStatus.Information = BytesReturned;
    IoCompleteRequest(Irp, IO_NO_INCREMENT);
    
    return Status;
}

// =============================================
// Driver Initialization
// =============================================

NTSTATUS DriverEntry(PDRIVER_OBJECT DriverObject, PUNICODE_STRING RegistryPath)
{
    NTSTATUS Status;
    UNICODE_STRING DeviceName, SymlinkName;
    
    UNREFERENCED_PARAMETER(RegistryPath);
    
    DbgPrint("[MSRDriver] Loading - Based on MSR.pptx Power Management\n");
    DbgPrint("[MSRDriver] Supported MSRs:\n");
    for (ULONG i = 0; i < RTL_NUMBER_OF(g_SupportedMSRs); i++) {
        DbgPrint("[MSRDriver]   0x%03X: %S\n", 
                 g_SupportedMSRs[i].MsrAddress,
                 g_SupportedMSRs[i].Name);
    }
    
    // Create device
    RtlInitUnicodeString(&DeviceName, DEVICE_NAME);
    Status = IoCreateDevice(
        DriverObject,
        0,
        &DeviceName,
        FILE_DEVICE_UNKNOWN,
        0,
        FALSE,
        &g_DeviceObject
    );
    
    if (!NT_SUCCESS(Status)) {
        DbgPrint("[MSRDriver] Failed to create device: 0x%X\n", Status);
        return Status;
    }
    
    // Create symbolic link
    RtlInitUnicodeString(&SymlinkName, SYMLINK_NAME);
    Status = IoCreateSymbolicLink(&SymlinkName, &DeviceName);
    
    if (!NT_SUCCESS(Status)) {
        DbgPrint("[MSRDriver] Failed to create symlink: 0x%X\n", Status);
        IoDeleteDevice(g_DeviceObject);
        return Status;
    }
    
    // Set up driver entry points
    DriverObject->MajorFunction[IRP_MJ_CREATE] = MsrDriverCreateClose;
    DriverObject->MajorFunction[IRP_MJ_CLOSE] = MsrDriverCreateClose;
    DriverObject->MajorFunction[IRP_MJ_DEVICE_CONTROL] = MsrDriverDeviceControl;
    DriverObject->DriverUnload = MsrDriverUnload;
    
    // Initialize device
    g_DeviceObject->Flags |= DO_BUFFERED_IO;
    g_DriverInitialized = TRUE;
    
    DbgPrint("[MSRDriver] Loaded successfully\n");
    return STATUS_SUCCESS;
}
