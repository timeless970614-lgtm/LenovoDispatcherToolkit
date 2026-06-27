# Changelog

## v1.0.25 (2026-06-27)

### New Features
- **Performance Analysis Knowledge Base**: Expanded ETL Trace knowledge base from 8 to 14 sections based on WinHEC 2026 Lab Guide
  - WAT (Windows Assessment Toolkit) — WAT vs WPR selection, WAC workflow
  - CPA Drill-Down Views — 4 bottleneck types × corresponding WPA views (Browse Wait Chain, Interference CPU, Disk IO DrillDown, Disk Interference)
  - CPU Starvation vs Preemption — root cause distinction and remediation direction
  - ETW MCP AI-Assisted Analysis — GitHub Copilot CLI complete workflow
  - Energy Saver A/B Comparison — CPU frequency/utilization/critical-path quantified impact table
  - Troubleshooting — 8 common WPA/ETW MCP issues with solutions

## v1.0.24 (2026-06-27)

### Improvements
- **Test Function**: Added 5-second idle between app launches in BenchmarkLaunchSpeed() to let each app settle before the next — 15 apps × 5s = ~75s total
- **Test Function**: Fixed user path in GetCommonAppsList() (`3-64` → `Zhou`) for VSCode detection

## v1.0.23 (2026-06-27)

### Performance
- **AIAnalysis.vue Split**: 2236 lines → 120 lines (↓95%), split into 4 sub-components (ETLTrace, SALog, LogAnalysis, ToolkitTab) with `v-if` on-demand rendering
- **ai_agent.go Batch PS**: Merged 12 independent PowerShell processes into 1 batch call + 1 powercfg — ~3s → ~350ms

### Bug Fixes
- **Error Handling**: Fixed 30+ high-risk ignored errors across 7 backend files (etl_analyzer.go, nvidia_api.go, scripts.go, registrywatcher.go, syspower.go, event_log.go)
- **dynamicnpu.go**: 6 `syscall.Syscall` calls now capture errno instead of discarding (`ret, _, _` → `ret, _, errno`)

### Infrastructure
- **Repository Migrated**: github.com/timeless970614-lgtm/LenovoDispatcherToolkit → github.com/timeless970614-lgtm/LenovoToolkit
- **.gitignore**: Added `*.bak`, `test_*.ps1`

## v1.0.22 (2026-06-24)

### New Features
- EVTX file loading and level-based filtering
- EventData column with key=value format
- Table column dividers
- Critical events sorted to top by time

## v1.0.21 (2026-06-23)

### New Features
- **Dispatcher Dump**: Added "Enable Dispatcher Dump" button — auto-configures Windows Error Reporting LocalDumps for `LNV_DES.exe` and `LNVDispatcherService.exe`
- **Dispatcher Service Status**: Dashboard "Service Control" renamed to "Dispatcher Service Status"

### Performance
- **CPU Usage Reduction**: Idle CPU dropped from ~3% to <0.5%
  - Dashboard polling split: lightweight `refresh()` (registry only) vs full `fullRefresh()` (WMI)
  - Dashboard refresh interval 5s → 10s
  - App.vue mode polling 30s → 60s
  - `visibilitychange` pauses all polling when window hidden/minimized
  - WMI `GetSystemInfo` / `GetModeCheckInfo` cached, only called on first load and user actions
- **Startup**: Merged 8 backend function calls into single batch PS + Go registry reads (2s → ~350ms)

### UI Improvements
- PCM_Service / Vantage_Service: frontend displays N/A instead of service status
- Dispatcher Service: "Stopped" displays as "Not Supported"
- FunctionCheck labels: removed uppercase transform, added "Auto iGPU on battery" fallback text
- PPMDriver: title changed to "Processor Power Management Parameters"

### Bug Fixes
- Dashboard refresh button bound to `fullRefresh()` (was `refresh()` — missed WMI data)

---

## v1.0.19 (2026-05-25)

### New Features
- **ETL Trace**: Added "Load External ETL" card — choose `.etl` file from disk and auto-analyze in one click
- **Power Module**: Added RAPL power reading via `lenovo_power.sys` kernel driver (CPU package, Core, GFX, DRAM power)
- **Power Module**: Added IPF V1 connection via `ipfsrv.public` pipe (PL1/PL2, CPU temp), works without admin
- **PPM Driver**: Added driver file path display (PPM Package `.ppkg` path, full path visible without truncation)

### Bug Fixes
- IPF temp formula fix: raw value is deciKelvin, formula `val/10.0 - 273.15`
- RAPL `raplState` deadlock fix: moved `detect()` out of `readEnergyUnits()` to avoid recursive mutex lock
- Power Management UI: removed unreliable System Power and GPU Power cards

---

## v1.0.18 (2026-05-11)

### UI Fixes
- **Dispatcher GPU Mode Control**: Fixed parameter labels (PCM_Service, PCM_GPUStatus, PE_GPUPrefStatus, Vantage_Service, Vantage_GPUStatus, Vantage_DefaultMode) to display in mixed-case instead of all uppercase
- **Dispatcher GPU Mode Control**: Adjusted card layout - left-aligned with tab bar, increased card width for better visual alignment

### New Features
- **NPU Module**: Added 5 new Wails bindings - SetNPUPowerLimit, StartNPUScheduler, StopNPUScheduler, GetNPOSchedulerState, GetNPOSchedulerSettings - enabling full NPU scheduling UI functionality
- **NPU Module**: Fixed scheduler settings to use proper Go struct field names (PascalCase)

### Bug Fixes
- **ETL Trace**: Fixed trace capture timing issues

---