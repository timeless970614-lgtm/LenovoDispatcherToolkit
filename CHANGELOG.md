# Changelog

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