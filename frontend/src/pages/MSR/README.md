# MSR Test Tools - README
# Based on: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx

## Project Structure

```
MSR/
├── MSR_Test_App.csproj    # .NET Test Application
├── Program.cs              # Main test application source
├── README.md               # This file
├── Test-MSR.ps1            # PowerShell MSR testing script
├── Build.ps1               # Build script
└── driver/
    ├── MSRDriver.c         # Kernel driver source (WDM)
    ├── MSRDriver.h         # Driver header with MSR definitions
    └── MSRDriver.inf       # Driver installation file
```

## MSR Registers from PPT

Based on the MSR.pptx slides, the following MSRs are supported:

### Power Management MSRs
| Address | Name | Description | Access |
|---------|------|-------------|--------|
| 0x1B0 | IA32_ENERGY_PERF_BIAS | Energy Performance Bias Hint | RW |
| 0x198 | IA32_PERF_CTL | P-State Target Control | RW |
| 0x199 | IA32_PERF_STATUS | Current P-State (read-only) | RO |
| 0x1A0 | IA32_MISC_ENABLE | Turbo Boost Control | RW |
| 0x19A | IA32_CLOCK_MODULATION | Clock Modulation (T-state) | RW |

### Frequency Monitoring MSRs
| Address | Name | Description | Access |
|---------|------|-------------|--------|
| 0xE7 | IA32_APERF | Actual Performance Counter | RO |
| 0xE8 | IA32_MPERF | Maximum Performance Counter | RO |

### Uncore Frequency MSRs
| Address | Name | Description | Access |
|---------|------|-------------|--------|
| 0x620 | MSR_UNCORE_RATIO_LIMIT | Uncore Frequency Limits | RW |

### RAPL MSRs
| Address | Name | Description | Access |
|---------|------|-------------|--------|
| 0x606 | MSR_RAPL_POWER_UNIT | Power/Energy/Time Units | RO |
| 0x610 | MSR_PKG_POWER_LIMIT | Package Power Limit | RW |
| 0x611 | MSR_PKG_ENERGY_STATUS | Package Energy Consumed | RO |
| 0x614 | MSR_PKG_POWER_INFO | Package Power Range Info | RO |
| 0x638 | MSR_PP0_POWER_LIMIT | CPU Cores Power Limit | RW |
| 0x639 | MSR_PP0_ENERGY_STATUS | CPU Cores Energy | RO |
| 0x618 | MSR_DRAM_POWER_LIMIT | DRAM Power Limit | RW |
| 0x619 | MSR_DRAM_ENERGY_STATUS | DRAM Energy | RO |

## Building

### .NET Application
```powershell
cd C:\LenovoDispatcher\MSR
dotnet restore
dotnet build -c Release
```

### Kernel Driver (requires WDK)
```powershell
# Open driver/MSRDriver.sln in Visual Studio with WDK
# Build for x64 Release
```

## Running

### .NET Application
```powershell
# Requires WinRing0 driver installed
.\bin\Release\net8.0\MSR_Test_App.exe
```

### PowerShell Script (No Driver Required)
```powershell
.\Test-MSR.ps1
```

## Installation

### WinRing0 Driver
For the .NET application, you need WinRing0:
1. Download WinRing0 from: https://github.com/george-cheng/WinRing0
2. Copy WinRing0x64.sys to C:\Windows\System32\drivers\
3. Run as Administrator:
   ```cmd
   sc create WinRing0 type= kernel binPath= C:\Windows\System32\drivers\WinRing0x64.sys
   sc start WinRing0
   ```

## Safety Warnings

⚠️ **WARNING: MSR modifications can cause system instability!**

- Always backup settings before modification
- Some changes require administrator privileges
- Incorrect values may damage hardware or void warranty
- Turbo Boost and clock modulation changes may cause crashes
- Power limit changes should stay within CPU specifications

## PPT Reference

This project was created based on:
- File: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx
- Topics:
  - Slide 5: Energy Performance Bias
  - Slide 6: Intel P-State Machine
  - Slide 7: Hardware-coordination feedback (APERF/MPERF)
  - Slide 10: Uncore Frequency Scaling
  - Slide 18: Dynamic Duty Cycle Modulation
  - Slides 54-58: RAPL (Running Average Power Limit)
