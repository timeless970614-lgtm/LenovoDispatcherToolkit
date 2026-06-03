# PPM Provisioning Package Analysis Report

## File Information
- **Filename**: PPM-ARL-v1007.20250118.ppkg
- **Version**: 1007.20250118 (January 18, 2025)
- **Target Platform**: Intel Arrow Lake (ARL) - HX/S series
- **Format**: Windows Imaging Format (WIM) - Magic: `MSWI`
- **Compression**: XPress:15
- **Size**: 87,918 bytes (compressed) → 1,088,478 bytes (extracted)
- **Package ID**: {3a73daca-fe3c-4656-a629-3cf7fa3b9105}

## Purpose
This package contains Intel Processor Power Management (PPM) provisioning settings optimized for Arrow Lake (ARL) processors on Windows 11 24H2.

## File Structure
```
├── Multivariant/
│   ├── Multivariant.xml              - Variant selector manifest
│   └── 0/
│       ├── customizations.xml        - Main configuration (490KB)
│       ├── MasterDatastore.xml       - Data store reference
│       └── Prov/
│           ├── RunTime.xml           - Runtime policy selector (by CPU model)
│           └── RunTime/
│               ├── 0__Power_Policy.provxml    - Base power policies
│               ├── 1__Power_Policy.provxml    - Platform-specific policies
│               ├── ...
│               └── 50__Power_Policy.provxml   - Model-specific optimizations
└── CommonSettings/
    ├── CommonSettings.xml
    └── 1/AnswerFile.xml
```

## Power Policy Configuration Categories

### 1. Core Parking Parameters
| Parameter | Description | Impact |
|-----------|-------------|--------|
| CPMinCores | Minimum unparked cores | Higher = lower latency, more power |
| ModuleUnparkPolicy | Core unparking behavior | 1 = aggressive unparking |
| SoftParkLatency | Time before full park (ms) | Higher = longer shallow sleep |
| CpLatencyHintUnpark1 | Latency hint for unparking | Affects scheduling decisions |

### 2. Heterogeneous Scheduling (P-Core / E-Core)
| Parameter | Description | Typical Value |
|-----------|-------------|---------------|
| HeteroIncreaseThreshold | P-Core activation threshold | 254 (very aggressive) |
| HeteroDecreaseThreshold | E-Core fallback threshold | 254 |
| HeteroPolicy | Heterogeneous scheduling mode | 0 = balanced |
| HeteroClass1InitialPerf | Initial P-Core performance | 100% |

### 3. Energy Performance Preference (EPP)
| Parameter | Description | Range |
|-----------|-------------|-------|
| PerfEnergyPreference | Power/performance balance | 0-255 (lower = performance) |
| PerfEnergyPreference1 | Secondary EPP for P-Cores | 0-255 |

**Typical EPP Values by Power Scheme:**
- Performance: 10-45
- Balanced: 45-70
- Power Saver: 70-100

### 4. Frequency Control
| Parameter | Description | Unit |
|-----------|-------------|------|
| MaxFrequency | Maximum P-Core frequency | MHz |
| MaxFrequency1 | Maximum E-Core frequency | MHz |

### 5. Scheduling Policy
| Parameter | Description | Values |
|-----------|-------------|--------|
| SchedulingPolicy | Thread scheduling strategy | 0-3 |
| ShortSchedulingPolicy | Short-term scheduling | 0-3 |

### 6. Performance Autonomy
| Parameter | Description | Values |
|-----------|-------------|--------|
| PerfAutonomousMode | Hardware autonomous control | 0-1 (1 = enabled) |

## CPU Model Targeting

The package targets specific Intel CPU families via `RunTime.xml`:

| Model | CPU Family | Examples |
|-------|------------|----------|
| Model 66 | Ice Lake | Core 10xxx |
| Model 76 | Comet Lake | Core 10xxx |
| Model 126 | Alder Lake | Core 12xxx |
| Model 140 | Raptor Lake | Core 13xxx/14xxx |
| Model 166 | Arrow Lake | Core Ultra 200 |

## Power Scheme GUIDs

| Scheme | GUID |
|--------|------|
| Balanced | 381b4222-f694-41f0-9685-ff5bb260df2e |
| High Performance | 8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c |
| Power Saver | a1841308-3541-4fab-bc81-f71556f20b4a |

## Sub-Policy GUIDs (Per-Application)

| GUID | Purpose |
|------|---------|
| 0c3d5326-944b-4aab-8ad8-fe422a0e50e0 | Power throttling policy |
| 0da965dc-8fcf-4c0b-8efe-8dd5e7bc959a | Performance boost policy |
| 336c7511-f109-4172-bb3a-3ea51f815ada | Efficiency policy |
| 33cc3a0d-45ee-43ca-86c4-695bfc9a313b | Low power policy |

## Key Insights

1. **Heterogeneous Scheduling**: Arrow Lake uses aggressive P-Core activation (threshold=254) for responsiveness
2. **EPP Configuration**: Granular control per power scheme and scenario
3. **Frequency Capping**: Maximum frequency limits for thermal/power management
4. **Core Parking**: Minimum cores set to prevent latency spikes
5. **Model-Specific Tuning**: 51 different policy files for various CPU/platform combinations

## How PPM Provisioning Works

1. **Installation**: Driver INF copies appropriate .ppkg file to `C:\Windows\provisioning\packages\`
2. **Runtime Selection**: Windows matches CPU model + platform role + AoAc status
3. **Policy Application**: Settings injected into registry under:
   ```
   HKLM\SYSTEM\CurrentControlSet\Control\Power\PowerSettings\...
   ```
4. **Dynamic Updates**: DTT (Dynamic Tuning Technology) can adjust policies based on workload

## Relationship with IPF/DTT

```
┌─────────────────────────────────────────────────────────────┐
│                    IPF Framework Manager                     │
│            (Intel Innovation Platform Framework)             │
├─────────────────────────────────────────────────────────────┤
│  ┌──────────────────┐  ┌──────────────────┐                 │
│  │  Processor       │  │  Generic         │                 │
│  │  Participant     │  │  Participant     │                 │
│  └────────┬─────────┘  └────────┬─────────┘                 │
│           │                     │                            │
│           ▼                     ▼                            │
│  ┌──────────────────────────────────────────┐               │
│  │        PPM Provisioning Package          │               │
│  │   (Power policy configuration data)      │               │
│  └──────────────────────────────────────────┘               │
│                      │                                       │
│                      ▼                                       │
│  ┌──────────────────────────────────────────┐               │
│  │      Intel Dynamic Tuning Technology     │               │
│  │      (Runtime policy adjustments)        │               │
│  └──────────────────────────────────────────┘               │
└─────────────────────────────────────────────────────────────┘
```

---
*Analysis generated from PPM-ARL-v1007.20250118.ppkg*
