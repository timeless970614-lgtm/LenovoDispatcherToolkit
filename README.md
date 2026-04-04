# Lenovo Dispatcher Toolkit

A Windows desktop toolkit for Lenovo ThinkPad power management, thermal mode control, and ML scenario performance logging.

Built with **Go + Wails v2** (WebView2), featuring a dark/light theme UI with English/Chinese support.

## Features

### 🖥️ Dashboard
- Real-time system monitoring: CPU, GPU, Memory, Disk, Battery
- Network status, OS info, hardware overview

### ⚡ PPM Driver — Live Power Settings
- Read and modify Intel PPM parameters via `powercfg`:
  - **EPP** (P-Core / E-Core) — Energy Performance Preference
  - **Hetero** Increase / Decrease Thresholds
  - **Max Frequency** (P-Core / E-Core)
  - **Soft Park Latency**
  - **CP Min Cores**, **Min / Max Processor State**
- Auto Refresh with configurable interval (1s/2s/3s/5s)
- Save changes take effect immediately via `powercfg /setactive`

### 🔧 Function Check
- Intel GPU feature validation and status reporting

### 🌡️ Fixed Thermal Mode
- **Standard Modes**: Battery Saving, Extreme Performance
- **Intelligent Mode**: Intelligent Battery Saving (IBSM), Intelligent Auto Quiet (AQM), Intelligent Stand Mode (STD), Intelligent Auto Performance (APM), Intelligent Extreme (IEPM), DCC Mode
- Policy Enable Function display with per-policy enable/disable status
- DYTC mode detection and live monitoring

### 📊 ML Scenario Log
- 1-second interval CSV capture matching ML_Scenario `Result.csv` format (84 columns)
- **IPF data** read directly from `LenovoIPFV2.dll` / `LenovoIPF.dll` via C++ wrapper (no service dependency):
  - `IPF_SystemPower` (mW), `MMIO_PL1` / `PL2` / `PL4` (mW)
  - `CPU_Temperature` (centiKelvin)
- **PPM data** from Windows registry / powercfg:
  - `EPP`, `EPP_1`, `PPM_FREQUENCY_LIMIT`, `PPM_CPMIN` / `PPM_CPMAX`, `SoftParking`, `EPOT`
- System metrics via PowerShell: CPU usage/freq, GPU usage, memory, disk, battery, ITS mode
- Output: `C:\ProgramData\Lenovo\LenovoDispatcher\Logs\MLScenario_*.csv`

### 🤖 AI Analysis
- Performance analysis tools

### ⚙️ Settings
- Theme: Dark / Light
- Language: English / 中文
- Auto Refresh Interval (default 1s)
- Auto Start with Windows
- Minimize to Tray

## Architecture

```
LenovoDispatcherToolkit/
├── app.go                          # Wails app entry, Bindings
├── main.go                         # Window config
├── wails.json                      # Wails project config
├── backend/                        # Go backend
│   ├── ipf_wrapper/                # C++ DLL wrapper for LenovoIPF*.dll
│   │   ├── ipf_wrapper.h           #   C export interface
│   │   ├── ipf_wrapper.cpp         #   Implementation (V1/V2/MSR)
│   │   ├── ipf_wrapper.def         #   DLL export table
│   │   └── BUILD.md                #   Build instructions
│   ├── ipf.go                      # Go dynamic DLL loading (syscall.LazyDLL)
│   ├── ppm.go                      # PPM power settings (powercfg)
│   ├── mlscenario.go              # ML log capture (1s interval CSV)
│   ├── modecheck.go               # DYTC / thermal mode
│   ├── funccheck.go               # Intel GPU function check
│   ├── dytc.go                    # DYTC interface
│   ├── intelgpu.go                # Intel GPU queries
│   ├── sysinfo.go                 # System info
│   ├── dispatcher.go              # Lenovo dispatcher service
│   ├── dynamiclog.go              # Named pipe log reader
│   └── ...
└── frontend/                       # Vue 3 + JS (no build step)
    └── src/
        ├── App.vue                 # Main layout + sidebar navigation
        └── pages/
            ├── Dashboard.vue
            ├── PPMDriver.vue
            ├── FunctionCheck.vue
            ├── ModeCheck.vue
            ├── DYTCMode.vue
            ├── AIAnalysis.vue
            ├── Settings.vue
            └── About.vue
```

## Prerequisites

- **Go** 1.22+
- **Node.js** 18+ (for frontend tooling)
- **Wails CLI v2**: `go install github.com/wailsapp/wails/v2/cmd/wails@latest`
- **Visual Studio 2022** (or Build Tools) — for compiling `ipf_wrapper.dll`
- **Windows 10/11** (x64)

## Build

### 1. Compile ipf_wrapper.dll (C++ → DLL)

```bat
cd backend\ipf_wrapper
cl /LD /EHsc /D_CRT_SECURE_NO_WARNINGS ipf_wrapper.cpp ipf_wrapper.def /Fe:ipf_wrapper.dll
```

Or open `build.bat` in Visual Studio Developer Command Prompt.

### 2. Build the application

```bash
wails build
```

Output: `build\bin\LenovoDispatcherToolkit.exe`

### 3. Copy runtime DLLs

Copy these DLLs alongside the EXE:

| DLL | Source |
|-----|--------|
| `LenovoIPFV2.dll` | From ML_Scenario build output |
| `LenovoIPF.dll` | From ML_Scenario build output |
| `WinMSRIO.dll` | From ML_Scenario build output |
| `Lenovo.Event.dll` | From ML_Scenario build output |

## IPF Data Flow

```
LenovoDispatcherToolkit.exe
  → ipf_wrapper.dll (C++ wrapper)
    → LenovoIPFV2.dll (V2, JSON API)  ← preferred
    → LenovoIPF.dll   (V1, ESIF SDK)  ← fallback
    → WinMSRIO.dll → MSRIO.sys        ← MSR reads (requires driver)
      → CPU RAPL (power), PL1/2/4, EPP, Frequency, Hetero, SoftPark
```

## CSV Column Reference

ML Scenario log outputs 84 columns matching ML_Scenario `Result.csv`:

| Col | Field | Unit |
|-----|-------|------|
| 0 | Time | `YYYY-M-D HH:MM:SS` |
| 1 | AC_DC | 0=Battery, 1=AC |
| 2 | PowerSlider | 1=Save, 2=Balanced, 3=Perf |
| 3-4 | CPU Usage, Frequency | %, MHz |
| 5-10 | GPU Usage (total/iGPU/dGPU/VPU/NPU) | % |
| 23-25 | System/CPU/GPU Power | W |
| 40 | IPF_SystemPower | mW |
| 41-43 | MMIO_PL1 / PL2 / PL4 | mW |
| 44-45 | PL1Check / PL2Check | — |
| 46 | EPOT | — |
| 47-48 | EPP / EPP_1 | 0-15 |
| 49-50 | PPM_FREQUENCY_LIMIT / _1 | — |
| 51-53 | PPM_CPMIN / CPMAX / SoftParking | — |
| 54 | CPU_Temperature | cK |

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.14 | 2026-04-04 | IPF/PPM DLL integration, ML Log rewrite, UI improvements |
| 1.0.13 | — | Initial release |

## License

Internal use only.
