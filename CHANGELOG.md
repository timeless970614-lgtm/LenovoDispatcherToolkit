# Changelog

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