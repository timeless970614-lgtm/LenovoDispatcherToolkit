/*
 * Lenovo Power Monitor Kernel Driver
 * Provides IOCTL-based MSR/PCI/MMIO access for RAPL power monitoring.
 *
 * MSR registers:
 *   - 0x606: RAPL_POWER_UNIT (energy unit scaling)
 *   - 0x611: PKG_ENERGY_STATUS (package energy counter)
 *   - 0x619: DRAM_ENERGY_STATUS (DRAM energy)
 *   - 0x639: PP0_ENERGY_STATUS (core energy)
 *   - 0x641: PP1_ENERGY_STATUS (graphics energy)
 *   - 0xC001029B: AMD PKG_ENERGY_STAT
 *
 * Build: See build.bat
 * Load:  sc create LunaPowerMon type=kernel binPath="C:\path\lenovo_power.sys"
 * Start: sc start LunaPowerMon
 */

#include <ntddk.h>

/* ── Device Names ─────────────────────────────────────────────────────────── */
#define DEVICE_NAME_U  L"\\Device\\LunaPowerMon"
#define SYMLINK_NAME_U L"\\DosDevices\\LunaPowerMon"

#define FILE_DEVICE_LPM 0x8011

/* ── IOCTL Codes ───────────────────────────────────────────────────────────── */
#define IOCTL_LPM_READ_MSR     CTL_CODE(FILE_DEVICE_LPM, 0x800, METHOD_BUFFERED, FILE_ANY_ACCESS)
#define IOCTL_LPM_WRITE_MSR    CTL_CODE(FILE_DEVICE_LPM, 0x801, METHOD_BUFFERED, FILE_ANY_ACCESS)
#define IOCTL_LPM_READ_PCI_CFG CTL_CODE(FILE_DEVICE_LPM, 0x802, METHOD_BUFFERED, FILE_ANY_ACCESS)
#define IOCTL_LPM_READ_PHYS_MEM CTL_CODE(FILE_DEVICE_LPM, 0x803, METHOD_BUFFERED, FILE_ANY_ACCESS)

/* ── IOCTL Structures ──────────────────────────────────────────────────────── */
#pragma pack(push, 1)
typedef struct _MSR_IO {
    ULONG   Index;       /* MSR index [input+output] */
    ULONG64 Value;       /* MSR value [output] */
    LONG    Status;      /* 0=success, !0=fault [output] */
} MSR_IO;

typedef struct _PCI_CFG_IO {
    ULONG Bus;           /* [input] */
    ULONG Dev;           /* [input] */
    ULONG Func;          /* [input] */
    ULONG Offset;        /* [input] */
    ULONG Size;          /* [input] 1/2/4 bytes */
    ULONG Value;         /* [output] */
    LONG  Status;        /* [output] */
} PCI_CFG_IO;

typedef struct _PHYS_MEM_IO {
    ULONG64 PhysAddr;    /* [input] */
    ULONG   Size;        /* [input] max 256 */
    UCHAR   Data[256];   /* [output] */
    LONG    Status;      /* [output] */
} PHYS_MEM_IO;
#pragma pack(pop)

/* ── Prototypes ────────────────────────────────────────────────────────────── */
DRIVER_INITIALIZE DriverEntry;
DRIVER_UNLOAD     DriverUnload;
DRIVER_DISPATCH   DispatchCreateClose;
DRIVER_DISPATCH   DispatchDeviceControl;

/* ── Driver Entry ──────────────────────────────────────────────────────────── */
NTSTATUS DriverEntry(
    PDRIVER_OBJECT  DriverObject,
    PUNICODE_STRING RegistryPath)
{
    NTSTATUS        status;
    PDEVICE_OBJECT  deviceObject = NULL;
    UNICODE_STRING  deviceName, symlinkName;

    UNREFERENCED_PARAMETER(RegistryPath);

    DbgPrint("[LPM] DriverEntry\n");

    RtlInitUnicodeString(&deviceName, DEVICE_NAME_U);
    RtlInitUnicodeString(&symlinkName, SYMLINK_NAME_U);

    status = IoCreateDevice(
        DriverObject,
        0,                              /* DeviceExtensionSize */
        &deviceName,
        FILE_DEVICE_UNKNOWN,
        FILE_DEVICE_SECURE_OPEN,
        FALSE,                          /* Exclusive */
        &deviceObject);

    if (!NT_SUCCESS(status)) {
        DbgPrint("[LPM] IoCreateDevice failed: 0x%08X\n", status);
        return status;
    }

    status = IoCreateSymbolicLink(&symlinkName, &deviceName);
    if (!NT_SUCCESS(status)) {
        DbgPrint("[LPM] IoCreateSymbolicLink failed: 0x%08X\n", status);
        IoDeleteDevice(deviceObject);
        return status;
    }

    deviceObject->Flags |= DO_BUFFERED_IO;
    deviceObject->Flags &= ~DO_DEVICE_INITIALIZING;

    DriverObject->MajorFunction[IRP_MJ_CREATE]         = DispatchCreateClose;
    DriverObject->MajorFunction[IRP_MJ_CLOSE]          = DispatchCreateClose;
    DriverObject->MajorFunction[IRP_MJ_DEVICE_CONTROL] = DispatchDeviceControl;
    DriverObject->DriverUnload                         = DriverUnload;

    DbgPrint("[LPM] Driver loaded successfully\n");
    return STATUS_SUCCESS;
}

/* ── Driver Unload ─────────────────────────────────────────────────────────── */
VOID DriverUnload(PDRIVER_OBJECT DriverObject)
{
    UNICODE_STRING symlinkName;

    RtlInitUnicodeString(&symlinkName, SYMLINK_NAME_U);

    DbgPrint("[LPM] DriverUnload\n");
    IoDeleteSymbolicLink(&symlinkName);
    if (DriverObject->DeviceObject) {
        IoDeleteDevice(DriverObject->DeviceObject);
    }
}

/* ── Create/Close Dispatch ─────────────────────────────────────────────────── */
NTSTATUS DispatchCreateClose(PDEVICE_OBJECT DeviceObject, PIRP Irp)
{
    UNREFERENCED_PARAMETER(DeviceObject);
    Irp->IoStatus.Status      = STATUS_SUCCESS;
    Irp->IoStatus.Information = 0;
    IoCompleteRequest(Irp, IO_NO_INCREMENT);
    return STATUS_SUCCESS;
}

/* ── __readmsr / __writemsr (intrinsics) ──────────────────────────────────── */
#ifdef _MSC_VER
#pragma intrinsic(__readmsr, __writemsr)
#endif

/* ── READ_PCI_CONFIG helper ────────────────────────────────────────────────── */
static ULONG ReadPCIConfigSpace(ULONG Bus, ULONG Dev, ULONG Func, ULONG Offset, ULONG Size)
{
    ULONG64 addr = 0x80000000ULL |
                   (((ULONG64)Bus & 0xFF) << 16) |
                   (((ULONG64)Dev & 0x1F) << 11) |
                   (((ULONG64)Func & 0x07) << 8) |
                   (Offset & 0xFC);

    /* Use HAL bus data access */
    BUS_DATA_TYPE busDataType = (Bus > 0) ? PCIConfiguration : PCIConfiguration;

    if (Bus > 0) {
        /* PCI Express - use PCIEXPRESS bus data type */
        ULONG slotNumber = (Dev << 3) | Func;
        return HalGetBusData(PCIConfiguration, Bus, slotNumber, (PVOID)(ULONG_PTR)Offset, Size);
    } else {
        return HalGetBusData(PCIConfiguration, Bus, Dev, (PVOID)(ULONG_PTR)Offset, Size);
    }
}

/* ── Device Control Dispatch ───────────────────────────────────────────────── */
NTSTATUS DispatchDeviceControl(PDEVICE_OBJECT DeviceObject, PIRP Irp)
{
    PIO_STACK_LOCATION  stack;
    ULONG               ioctl;
    PVOID               buffer;
    ULONG               inLen, outLen;
    NTSTATUS            status = STATUS_SUCCESS;
    ULONG               info   = 0;

    UNREFERENCED_PARAMETER(DeviceObject);

    stack  = IoGetCurrentIrpStackLocation(Irp);
    ioctl  = stack->Parameters.DeviceIoControl.IoControlCode;
    buffer = Irp->AssociatedIrp.SystemBuffer;
    inLen  = stack->Parameters.DeviceIoControl.InputBufferLength;
    outLen = stack->Parameters.DeviceIoControl.OutputBufferLength;

    switch (ioctl) {

    /* ── IOCTL_LPM_READ_MSR ────────────────────────────────────────────── */
    case IOCTL_LPM_READ_MSR:
    {
        if (inLen < sizeof(MSR_IO) || outLen < sizeof(MSR_IO)) {
            status = STATUS_BUFFER_TOO_SMALL;
            break;
        }

        MSR_IO* msr = (MSR_IO*)buffer;
        msr->Value  = 0;
        msr->Status = 0;

        __try {
            msr->Value = __readmsr(msr->Index);
        }
        __except (EXCEPTION_EXECUTE_HANDLER) {
            msr->Status = GetExceptionCode();
            status = STATUS_ACCESS_VIOLATION;
        }

        info = sizeof(MSR_IO);
        break;
    }

    /* ── IOCTL_LPM_WRITE_MSR ───────────────────────────────────────────── */
    case IOCTL_LPM_WRITE_MSR:
    {
        if (inLen < sizeof(MSR_IO)) {
            status = STATUS_BUFFER_TOO_SMALL;
            break;
        }

        MSR_IO* msr = (MSR_IO*)buffer;
        msr->Status = 0;

        __try {
            __writemsr(msr->Index, msr->Value);
        }
        __except (EXCEPTION_EXECUTE_HANDLER) {
            msr->Status = GetExceptionCode();
            status = STATUS_ACCESS_VIOLATION;
        }

        info = sizeof(MSR_IO);
        break;
    }

    /* ── IOCTL_LPM_READ_PCI_CFG ────────────────────────────────────────── */
    case IOCTL_LPM_READ_PCI_CFG:
    {
        if (inLen < sizeof(PCI_CFG_IO) || outLen < sizeof(PCI_CFG_IO)) {
            status = STATUS_BUFFER_TOO_SMALL;
            break;
        }

        PCI_CFG_IO* pci = (PCI_CFG_IO*)buffer;
        pci->Value  = 0;
        pci->Status = 0;

        __try {
            pci->Value = ReadPCIConfigSpace(
                pci->Bus, pci->Dev, pci->Func, pci->Offset, pci->Size);
        }
        __except (EXCEPTION_EXECUTE_HANDLER) {
            pci->Status = GetExceptionCode();
            status = STATUS_ACCESS_VIOLATION;
        }

        info = sizeof(PCI_CFG_IO);
        break;
    }

    /* ── IOCTL_LPM_READ_PHYS_MEM ────────────────────────────────────────── */
    case IOCTL_LPM_READ_PHYS_MEM:
    {
        if (inLen < sizeof(PHYS_MEM_IO) || outLen < sizeof(PHYS_MEM_IO)) {
            status = STATUS_BUFFER_TOO_SMALL;
            break;
        }

        PHYS_MEM_IO* mem = (PHYS_MEM_IO*)buffer;
        ULONG size = mem->Size;

        if (size > sizeof(mem->Data)) {
            size = sizeof(mem->Data);
            mem->Size = size;
        }

        mem->Status = 0;
        RtlZeroMemory(mem->Data, sizeof(mem->Data));

        PHYSICAL_ADDRESS physAddr;
        physAddr.QuadPart = (LONGLONG)mem->PhysAddr;

        __try {
            PVOID mapped = MmMapIoSpace(physAddr, size, MmNonCached);
            if (mapped) {
                RtlCopyMemory(mem->Data, mapped, size);
                MmUnmapIoSpace(mapped, size);
            } else {
                mem->Status = 1;
                status = STATUS_UNSUCCESSFUL;
            }
        }
        __except (EXCEPTION_EXECUTE_HANDLER) {
            mem->Status = GetExceptionCode();
            status = STATUS_ACCESS_VIOLATION;
        }

        info = sizeof(PHYS_MEM_IO);
        break;
    }

    default:
        status = STATUS_INVALID_DEVICE_REQUEST;
        break;
    }

    Irp->IoStatus.Status      = status;
    Irp->IoStatus.Information = info;
    IoCompleteRequest(Irp, IO_NO_INCREMENT);
    return status;
}