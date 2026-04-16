<template>
  <div class="func-check-page">
    <!-- Function Tabs -->
    <div class="func-tabs">
      <button 
        v-for="tab in functionTabs" 
        :key="tab.id"
        :class="['func-tab', { active: activeTab === tab.id }]"
        @click="activeTab = tab.id"
      >
        <span class="tab-icon" v-html="tab.icon"></span>
        <span class="tab-label">{{ tab.label }}</span>
      </button>
    </div>

    <!-- GPU Function Content -->
    <div v-if="activeTab === 'gpu'" class="func-content">
      <!-- IGPU Mode Control Card -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4"/>
            </svg>
            Dispatcher GPU Mode Control
          </h3>
        </div>
        
        <div class="igpu-status">
          <div class="status-col-left">
            <div class="status-item status-item-tall">
              <span class="status-label">Current Status <span class="live-dot"></span></span>
              <span :class="['status-value', gpuPref.available ? (gpuPref.value === 2 ? 'status-uma' : gpuPref.value === 1 ? 'status-dis' : 'status-smart') : 'status-na']">
                {{ gpuPref.label }}
              </span>
            </div>
          </div>
          <div class="status-col-center">
            <div class="status-item">
              <span class="status-label">PCM_GPUStatus</span>
              <span :class="['status-value mono', gpuPref.pcmStatusAvail ? 'status-ok' : 'status-na']">
                {{ gpuPref.pcmStatusAvail ? gpuPref.pcmStatus + ' - ' + gpuPref.pcmLabel : 'N/A' }}
              </span>
            </div>
            <div class="status-item">
              <span class="status-label">Vantage_GPUStatus</span>
              <span class="status-value mono">{{ gpuPref.vantageStatus !== undefined ? gpuPref.vantageStatus : 'N/A' }}</span>
            </div>
          </div>
          <div class="status-col-right">
            <div class="status-item">
              <span class="status-label">PE_GPUPrefStatus</span>
              <span class="status-value mono">{{ gpuPref.available ? gpuPref.value : (gpuPref.label === 'Dispatcher not Support' ? '0' : 'N/A') }}</span>
            </div>
            <div class="status-item">
              <span class="status-label">Vantage_2</span>
              <span class="status-value mono">--</span>
            </div>
          </div>
        </div>

        <div class="igpu-control">
          <div class="btn-group">
            <button class="btn btn-primary" @click="setIGPUMode(0)" :disabled="settingMode || igpuStatus.mode === 0">
              Enable DGPU (Mode 0)
            </button>
            <button class="btn btn-warning" @click="setIGPUMode(1)" :disabled="settingMode || igpuStatus.mode === 1">
              Enable UMA Only (Mode 1)
            </button>
          </div>
        </div>

        <div v-if="settingMode" class="loading-overlay">
          <div class="spinner"></div>
          <p>Setting IGPU mode...</p>
        </div>

        <div v-if="settingResult" :class="['result-message', settingResult.success ? 'success' : 'error']">
          {{ settingResult.message }}
        </div>
      </div>

      <!-- System Diagnostic -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 16v-4M12 8h.01"/>
            </svg>
            System Diagnostic
          </h3>
        </div>
        
        <div class="diag-grid">
          <div class="diag-item">
            <span class="diag-label">Total GPUs Found</span>
            <span class="diag-value">{{ gpuList.length }}</span>
          </div>
          <div class="diag-item">
            <span class="diag-label">Discrete GPUs</span>
            <span class="diag-value">{{ discreteCount }}</span>
          </div>
          <div class="diag-item">
            <span class="diag-label">GPU Processes</span>
            <span class="diag-value">{{ processList.length }}</span>
          </div>
          <div class="diag-item">
            <span class="diag-label">NVIDIA Detected</span>
            <span class="diag-value">{{ nvidiaStatus.detected ? 'Yes' : 'No' }}</span>
          </div>
        </div>
      </div>

      <!-- GPU Using Processes - After System Diagnostic -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
              <line x1="9" y1="14" x2="15" y2="14"/>
            </svg>
            GPU-Using Processes
          </h3>
          <button class="btn btn-secondary btn-sm" @click="refreshProcesses" :disabled="loadingProcesses">
            <span v-if="loadingProcesses" class="spinner-small"></span>
            <span v-else>Refresh</span>
          </button>
        </div>
        
        <div v-if="processList.length > 0" class="process-list">
          <div class="process-header">
            <span class="col-pid">PID</span>
            <span class="col-name">Name</span>
            <span class="col-memory">Memory</span>
          </div>
          <div v-for="proc in processList" :key="proc.pid" class="process-item">
            <span class="col-pid">{{ proc.pid }}</span>
            <span class="col-name">{{ proc.name }}</span>
            <span class="col-memory">{{ proc.memory }}</span>
          </div>
        </div>
        <div v-else class="empty-state">
          <p>No GPU-using processes found</p>
        </div>
      </div>
    </div>

    <!-- Auto Gear Content -->
    <div v-if="activeTab === 'd'" class="func-content">
      <!-- Auto Gear Control -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            Auto Gear Control
          </h3>
        </div>
        <div class="auto-gear-info">
          <div class="gear-status-row">
            <span class="status-label">Current Gear <span class="live-dot"></span></span>
            <span class="status-value mono">
              <span :class="['gear-badge', 'gear-epot']">
                Gear {{ epotStatus.epot }}
              </span>
              <span class="epot-badge">EPOT</span>
            </span>
          </div>
        </div>

        <div v-if="settingGear" class="loading-overlay">
          <div class="spinner"></div>
          <p>Setting Gear mode...</p>
        </div>

        <div v-if="gearResult" :class="['result-message', gearResult.success ? 'success' : 'error']">
          {{ gearResult.message }}
        </div>
      </div>

      <!-- EPOT Status -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 20V10"/><path d="M12 20V4"/><path d="M6 20v-6"/>
            </svg>
            EPOT Status (ML_Scenario)
          </h3>
          <div class="live-indicator" v-if="epotRefreshing">
            <span class="live-dot"></span> Refreshing...
          </div>
          <button class="btn btn-secondary btn-sm" @click="loadEPOTStatus" :disabled="epotRefreshing">Refresh</button>
        </div>
        <div class="epot-grid">
          <div class="epot-row">
            <span class="epot-label">EPP</span>
            <span class="epot-desc">P-Core Energy Performance Preference</span>
            <span class="epot-value">{{ epotStatus.epp }}</span>
          </div>
          <div class="epot-row">
            <span class="epot-label">EPP_1</span>
            <span class="epot-desc">E-Core Energy Performance Preference</span>
            <span class="epot-value">{{ epotStatus.epp1 }}</span>
          </div>
          <div class="epot-row">
            <span class="epot-label">PPM_FREQUENCY_LIMIT</span>
            <span class="epot-desc">P-Core Frequency Limit</span>
            <span class="epot-value">{{ epotStatus.ppmFrequencyLimit }}</span>
          </div>
          <div class="epot-row">
            <span class="epot-label">PPM_FREQUENCY_LIMIT_1</span>
            <span class="epot-desc">E-Core Frequency Limit</span>
            <span class="epot-value">{{ epotStatus.ppmFrequencyLimit1 }}</span>
          </div>
          <div class="epot-row">
            <span class="epot-label">PPM_CPMIN</span>
            <span class="epot-desc">Min Active Cores</span>
            <span class="epot-value">{{ epotStatus.ppmCpMin }}</span>
          </div>
          <div class="epot-row">
            <span class="epot-label">PPM_CPMAX</span>
            <span class="epot-desc">Max Active Cores</span>
            <span class="epot-value">{{ epotStatus.ppmCpMax }}</span>
          </div>
          <div class="epot-row">
            <span class="epot-label">SoftParking</span>
            <span class="epot-desc">Soft Parking Delay</span>
            <span class="epot-value">{{ epotStatus.softParking }}</span>
          </div>
        </div>
      </div>

            <!-- Setting Gear Status -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title card-title-normal">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 20V10"/><path d="M12 20V4"/><path d="M6 20v-6"/>
            </svg>
            Set Auto Gear Status
          </h3>
          <div class="live-indicator" v-if="epotRefreshing">
            <span class="live-dot"></span> Refreshing...
          </div>
          <button class="btn btn-secondary btn-sm" @click="loadEPOTStatus" :disabled="epotRefreshing">Refresh</button>
        </div>
        
        <div class="gear-control">
          <div class="btn-group gear-btn-group">
            <button 
              class="btn btn-gear btn-gear-auto" 
              @click="setGear(0)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 0)"
              :class="{ active: gearStatus.available && gearStatus.value === 0 }"
            >
              <span class="gear-icon">⚡</span> Gear1
            </button>
            <button 
              class="btn btn-gear btn-gear-dgpu" 
              @click="setGear(1)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 1)"
              :class="{ active: gearStatus.available && gearStatus.value === 1 }"
            >
              <span class="gear-icon">🎮</span> Gear2
            </button>
            <button 
              class="btn btn-gear btn-gear-igpu" 
              @click="setGear(2)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 2)"
              :class="{ active: gearStatus.available && gearStatus.value === 2 }"
            >
              <span class="gear-icon">🔋</span> Gear3
            </button>
                        <button 
              class="btn btn-gear btn-gear-auto" 
              @click="setGear(0)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 0)"
              :class="{ active: gearStatus.available && gearStatus.value === 0 }"
            >
              <span class="gear-icon">⚡</span> Gear4
            </button>
            <button 
              class="btn btn-gear btn-gear-dgpu" 
              @click="setGear(1)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 1)"
              :class="{ active: gearStatus.available && gearStatus.value === 1 }"
            >
              <span class="gear-icon">🎮</span> Gear5
            </button>
            <button 
              class="btn btn-gear btn-gear-igpu" 
              @click="setGear(2)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 2)"
              :class="{ active: gearStatus.available && gearStatus.value === 2 }"
            >
              <span class="gear-icon">🔋</span> Gear6
            </button>
                        <button 
              class="btn btn-gear btn-gear-igpu" 
              @click="setGear(2)" 
              :disabled="settingGear || (gearStatus.available && gearStatus.value === 2)"
              :class="{ active: gearStatus.available && gearStatus.value === 2 }"
            >
              <span class="gear-icon">🔋</span> Gear7
            </button>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab A Content -->
    <div v-if="activeTab === 'a'" class="func-content">

      <!-- SSD Info Card -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="6" width="20" height="12" rx="2"/>
              <path d="M6 10h4M6 14h2"/>
            </svg>
            SSD Information
          </h3>
          <button class="btn btn-secondary btn-sm" @click="refreshSSD" :disabled="loadingSSD">
            <span v-if="loadingSSD" class="spinner-small"></span>
            <span v-else>Refresh</span>
          </button>
        </div>

        <div v-if="ssdList.length === 0 && !loadingSSD" class="empty-state">
          <svg width="40" height="40" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" style="opacity:0.3; margin-bottom:8px">
            <rect x="2" y="6" width="20" height="12" rx="2"/>
            <path d="M6 10h4M6 14h2"/>
          </svg>
          <p>No NVMe SSDs detected</p>
        </div>

        <div v-for="(ssd, idx) in ssdList" :key="idx" class="ssd-item">
          <div class="ssd-header">
            <span class="ssd-index">Drive {{ idx }}</span>
            <span class="ssd-name">{{ ssd.name }}</span>
            <span :class="['ssd-badge', ssd.multiModeCapable ? 'badge-capable' : 'badge-limited']">
              {{ ssd.multiModeCapable ? 'MultiMode' : 'Standard' }}
            </span>
          </div>

          <div class="ssd-info-grid">
            <div class="info-item">
              <span class="info-label">Model</span>
              <span class="info-value">{{ ssd.model || 'N/A' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Capacity</span>
              <span class="info-value">{{ ssd.capacityStr }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Protocol</span>
              <span class="info-value">{{ ssd.protocol }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Serial Number</span>
              <span class="info-value mono-sm">{{ ssd.serialNumber || 'N/A' }}</span>
            </div>
            <div class="info-item">
              <span class="info-label">Current Mode</span>
              <span :class="['info-value', 'mode-badge', ssd.currentModeStr !== 'N/A' ? 'mode-active' : '']">
                {{ ssd.currentModeStr }}
              </span>
            </div>
            <div class="info-item">
              <span class="info-label">Physical Drive</span>
              <span class="info-value">\\.\PhysicalDrive{{ ssd.driveIndex }}</span>
            </div>
          </div>

          <!-- Mode Selector (only for MultiMode capable) -->
          <div v-if="ssd.multiModeCapable" class="ssd-mode-control">
            <div class="mode-section-label">SSD Performance Mode</div>
            <div class="ssd-mode-buttons">
              <button
                v-for="mode in ssdModes"
                :key="mode.value"
                :class="['ssd-mode-btn', { active: ssd.currentMode === mode.value, loading: settingModeDrive === ssd.driveIndex }]"
                :disabled="settingModeDrive === ssd.driveIndex"
                @click="setSSDMode(ssd.driveIndex, mode.value)"
              >
                <span class="mode-abbr">{{ mode.label }}</span>
                <span class="mode-full">{{ mode.name }}</span>
              </button>
            </div>
            <p class="mode-hint" v-if="ssd.currentModeStr !== 'N/A'">
              <span class="hint-dot"></span>
              Current: <strong>{{ ssd.currentModeStr }}</strong> — Service restart required to apply changes
            </p>
          </div>

          <div v-if="ssd.error" class="ssd-error">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ ssd.error }}
          </div>

          <div v-if="modeResult && modeResult.driveIndex === ssd.driveIndex" :class="['result-message', modeResult.success ? 'success' : 'error']">
            {{ modeResult.message }}
          </div>
        </div>
      </div>

      <!-- NVMe Protocol Info Card -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            NVMe Protocol
          </h3>
        </div>
        <div class="card-content">
          <div class="info-item">
            <span class="info-label">Vendor Opcode</span>
            <span class="info-value mono-sm">0xC6</span>
          </div>
          <div class="info-item">
            <span class="info-label">Magic ID</span>
            <span class="info-value mono-sm">0x4657</span>
          </div>
          <div class="info-item">
            <span class="info-label">GET Mode SubCmd</span>
            <span class="info-value mono-sm">0x0F</span>
          </div>
          <div class="info-item">
            <span class="info-label">SET Mode SubCmd</span>
            <span class="info-value mono-sm">0x0E</span>
          </div>
          <div class="info-item">
            <span class="info-label">CEL Log Page</span>
            <span class="info-value mono-sm">0x05</span>
          </div>
          <div class="info-item">
            <span class="info-label">IOCTL</span>
            <span class="info-value mono-sm">IOCTL_STORAGE_PROTOCOL_COMMAND</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Tab B Content -->
    <div v-if="activeTab === 'b'" class="func-content">
    </div>

    <!-- Tab C Content - GPU Frequency -->
    <div v-if="activeTab === 'c'" class="func-content">
      <!-- Intel GPU Frequency Control Card -->
      <div class="card">
        <div class="card-header">
          <h3 class="card-title">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
            </svg>
            IGPU Frequency
          </h3>
          <span class="card-badge">IGC API</span>
        </div>

        <!-- GPU Info rows -->
        <div class="igpu-freq-info">
          <div class="freq-row">
            <span class="freq-label">GPU</span>
            <span class="freq-value">{{ intelGPU.gpuName || 'Not Detected' }}</span>
          </div>
          <div class="freq-row" v-if="intelGPU.driverVersion">
            <span class="freq-label">Driver</span>
            <span :class="['freq-value mono', intelGPU.driverOK ? 'driver-ok' : 'driver-old']">
              {{ intelGPU.driverVersion }}
              <span v-if="intelGPU.driverOK" class="driver-badge ok">✓</span>
              <span v-else class="driver-badge old">!</span>
            </span>
          </div>
          <div class="freq-row" v-if="intelGPU.driverDate">
            <span class="freq-label">Driver Date</span>
            <span class="freq-value mono">{{ intelGPU.driverDate }}</span>
          </div>
          <div class="freq-row" v-if="intelGPU.minDriverVersion">
            <span class="freq-label">Min Required</span>
            <span class="freq-value mono muted">{{ intelGPU.minDriverVersion }}</span>
          </div>
        </div>

        <!-- Driver upgrade warning -->
        <div v-if="intelGPU.available && !intelGPU.driverOK" class="driver-warning">
          <div class="warning-icon">⚠️</div>
          <div class="warning-content">
            <p><strong>Driver version too old!</strong></p>
            <p>Current: {{ intelGPU.driverVersion }} | Required: {{ intelGPU.minDriverVersion }}</p>
            <button class="btn btn-primary btn-sm" @click="openDriverDownload">
              Upgrade Driver
            </button>
          </div>
        </div>

        <!-- Live frequency status grid (IGC available) -->
        <div v-if="intelGPU.available" class="igpu-freq-status">
          <div class="freq-stat-grid">
            <div class="freq-stat-item">
              <span class="freq-stat-label">HW Min</span>
              <span class="freq-stat-value">{{ intelGPU.minFreq.toFixed(0) }} MHz</span>
            </div>
            <div class="freq-stat-item">
              <span class="freq-stat-label">HW Max</span>
              <span class="freq-stat-value">{{ intelGPU.maxFreq.toFixed(0) }} MHz</span>
            </div>
            <div class="freq-stat-item">
              <span class="freq-stat-label">Limit Min</span>
              <span class="freq-stat-value">{{ intelGPU.currentMin.toFixed(0) }} MHz</span>
            </div>
            <div class="freq-stat-item">
              <span class="freq-stat-label">Limit Max</span>
              <span class="freq-stat-value highlight">{{ intelGPU.currentMax.toFixed(0) }} MHz</span>
            </div>
            <!-- Current Frequency: estimated from limit (registry-based, no IGC) -->
            <div class="freq-stat-item">
              <span class="freq-stat-label">Current Freq</span>
              <span class="freq-stat-value live">
                {{ currentFreqDisplay }}
              </span>
            </div>
            <!-- GPU Utilization from Windows Performance Counter -->
            <div class="freq-stat-item">
              <span class="freq-stat-label">GPU Load</span>
              <span :class="['freq-stat-value', gpuUtilClass]">
                {{ gpuUtilDisplay }}
              </span>
            </div>
          </div>
        </div>

        <!-- Frequency range control sliders -->
        <div class="freq-control" v-if="intelGPU.available">
          <div class="freq-slider-group">
            <label>Min Frequency (MHz) &nbsp;<span class="slider-val-inline">{{ freqMin.toFixed(0) }}</span></label>
            <div class="slider-row">
              <input type="range"
                :min="intelGPU.minFreq"
                :max="intelGPU.maxFreq"
                :step="freqStep"
                v-model.number="freqMin"
                class="freq-slider"
                @input="onFreqMinInput">
              <span class="slider-val">{{ freqMin.toFixed(0) }}</span>
            </div>
          </div>
          <div class="freq-slider-group">
            <label>Max Frequency (MHz) &nbsp;<span class="slider-val-inline">{{ freqMax.toFixed(0) }}</span></label>
            <div class="slider-row">
              <input type="range"
                :min="intelGPU.minFreq"
                :max="intelGPU.maxFreq"
                :step="freqStep"
                v-model.number="freqMax"
                class="freq-slider"
                @input="onFreqMaxInput">
              <span class="slider-val">{{ freqMax.toFixed(0) }}</span>
            </div>
          </div>
        </div>

        <!-- Action buttons -->
        <div class="freq-actions" v-if="intelGPU.available">
          <button class="btn btn-primary" @click="applyFreqRange" :disabled="freqTesting">
            Apply Range
          </button>
          <button class="btn btn-outline" @click="testFreq('min')" :disabled="freqTesting">
            Lock Min
          </button>
          <button class="btn btn-outline" @click="testFreq('max')" :disabled="freqTesting">
            Lock Max
          </button>
          <button class="btn btn-outline" @click="testFreq('dynamic')" :disabled="freqTesting">
            Restore Dynamic
          </button>
          <button class="btn btn-outline" @click="refreshIGPUFreq" :disabled="freqTesting">
            ↻ Refresh
          </button>
        </div>

        <div v-if="freqTestResult" :class="['result-message', freqTestResult.success ? 'success' : 'error']">
          {{ freqTestResult.message }}
        </div>

        <div v-if="!intelGPU.available && !intelGPU.error" class="empty-state">
          <p>Checking Intel GPU...</p>
        </div>
        <div v-if="intelGPU.error" class="error-state">
          <p>{{ intelGPU.error }}</p>
          <a href="https://www.intel.cn/content/www/cn/zh/download-center/home.html" target="_blank" class="download-link">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            下载 Intel 显卡驱动
          </a>
        </div>
      </div>
    </div>

  </div>
</template>
<script>
import { EventsOn, EventsOff } from '../../wailsjs/runtime/runtime'
import { EnumerateGPUs, EnumerateGPUProcesses, GetIGPUMode, SetIGPUMode, CheckNVIDIAStatus, GetSSDInfo, SetSSDMode, GetGPUPrefStatus, GetIntelGPUFrequency, SetIntelGPUFrequencyRange, TestIntelGPUFrequency, GetIntelDriverDownloadURL, GetIntelGPUUtilization, StartGPUStatusWatcher, StopGPUStatusWatcher, GetGPUPrefStatusFromCache, GetGPUAutoGear, SetGPUAutoGear, GetEPOTStatus, UninstallDTT, UninstallDTTUI } from '../../wailsjs/go/main/App'

export default {
  name: 'FunctionCheck',
  data() {
    return {
      activeTab: 'gpu',
      functionTabs: [
        { id: 'gpu', label: 'DGPU Function', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2"/><line x1="12" y1="12" x2="12" y2="16"/><line x1="10" y1="14" x2="14" y2="14"/></svg>' },
        { id: 'd', label: 'Auto Gear', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2"/><line x1="12" y1="12" x2="12" y2="16"/><line x1="10" y1="14" x2="14" y2="14"/></svg>' },
        { id: 'a', label: 'SSD Turbo', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="6" width="20" height="12" rx="2"/><path d="M6 10h4M6 14h2"/></svg>' },
        { id: 'c', label: 'IGPU Frequency', icon: '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>' },
      ],
      gpuList: [],
      processList: [],
      igpuStatus: {
        available: false,
        mode: -1
      },
      gpuPref: {
        available: false,
        value: -1,
        label: 'Not Available'
      },
      intelGPU: {
        available: false,
        minFreq: 100,
        maxFreq: 2000,
        currentMin: 100,
        currentMax: 2000,
        requestedMHz: 0,
        actualMHz: 0,
        tdpMHz: 0,
        efficientMHz: 0,
        gpuName: '',
        driverVersion: '',
        driverDate: '',
        minDriverVersion: '',
        driverOK: false,
        adapterIndex: 0,
        error: ''
      },
      freqMin: 100,
      freqMax: 2000,
      freqTesting: false,
      freqTestResult: null,
      gpuUtilization: -1,   // -1 = not yet loaded
      utilTimer: null,       // setInterval handle for utilization polling
      nvidiaStatus: {
        detected: false,
        nvmlLoaded: false,
        serviceRunning: false
      },
      settingResult: null,
      loading: false,
      loadingSSD: false,
      ssdList: [],
      ssdModes: [
        { value: 0, label: 'STD', name: 'Standard' },
        { value: 1, label: 'PERF', name: 'Performance' },
        { value: 2, label: 'PWR', name: 'Power Saving' },
        { value: 3, label: 'DEF', name: 'Default' }
      ],
      settingModeDrive: null,
      modeResult: null,
      loadingProcesses: false,
      settingMode: false,
      checkingNvidia: false,
      gearStatus: { available: false, value: 0 },
      settingGear: false,
      gearResult: null,
      epotStatus: { epot: 0, epp: 0, epp1: 0, ppmFrequencyLimit: 0, ppmFrequencyLimit1: 0, ppmCpMin: 0, ppmCpMax: 0, softParking: 0 },
      epotRefreshing: false,
      uninstallingDTT: false,
      uninstallingDTTUI: false,
      dttResult: null,
    }
  },
  computed: {
    discreteCount() {
      return this.gpuList.filter(g => g.isDiscrete).length
    },
    igpuCount() {
      return this.gpuList.filter(g => !g.isDiscrete).length
    },
    // Slider step: 50 MHz for ranges > 500 MHz, else 25 MHz
    freqStep() {
      const range = this.intelGPU.maxFreq - this.intelGPU.minFreq
      return range > 500 ? 50 : 25
    },
    // Current Frequency display: registry-based limit (no real-time IGC available)
    currentFreqDisplay() {
      const max = this.intelGPU.currentMax
      if (!max || max <= 0) return 'N/A'
      return max.toFixed(0) + ' MHz'
    },
    // GPU utilization display
    gpuUtilDisplay() {
      if (this.gpuUtilization < 0) return 'N/A'
      return this.gpuUtilization.toFixed(1) + ' %'
    },
    // CSS class for utilization value
    gpuUtilClass() {
      if (this.gpuUtilization < 0) return ''
      if (this.gpuUtilization >= 80) return 'freq-stat-value util-high'
      if (this.gpuUtilization >= 40) return 'freq-stat-value util-mid'
      return 'freq-stat-value util-low'
    }
  },
  async mounted() {
    await this.refreshAll()
    // Load NPU device info
    // (NPU card removed)
    // Start registry watcher for GPU status (real-time, no polling)
    await this.startGPUStatusWatcher()
    // Start GPU utilization polling (every 3s, only when IGPU tab is active)
    this.startUtilPolling()
  },
  beforeUnmount() {
    this.stopGPUStatusWatcher()
    this.stopUtilPolling()
  },
  methods: {
    formatMemory(bytes) {
      if (!bytes || bytes === 0) return 'N/A'
      if (bytes >= 1024 * 1024 * 1024) {
        return (bytes / (1024 * 1024 * 1024)).toFixed(1) + ' GB'
      } else if (bytes >= 1024 * 1024) {
        return Math.round(bytes / (1024 * 1024)) + ' MB'
      }
      return bytes + ' B'
    },

    async refreshSSD() {
      this.loadingSSD = true
      this.modeResult = null
      try {
        const list = await GetSSDInfo()
        this.ssdList = list || []
      } catch (e) {
        this.ssdList = []
      } finally {
        this.loadingSSD = false
      }
    },

    async setSSDMode(driveIndex, modeValue) {
      this.settingModeDrive = driveIndex
      this.modeResult = null
      try {
        const result = await SetSSDMode(driveIndex, modeValue)
        this.modeResult = result
        if (result.success) {
          // Update current mode display optimistically
          const ssd = this.ssdList.find(s => s.driveIndex === driveIndex)
          if (ssd) {
            const modes = ['Standard', 'Performance', 'Power Saving', 'Default']
            ssd.currentMode = modeValue
            ssd.currentModeStr = modes[modeValue] || 'N/A'
          }
        }
      } catch (e) {
        this.modeResult = { driveIndex, success: false, message: String(e) }
      } finally {
        this.settingModeDrive = null
      }
    },

    async refreshGPU() {
      this.loading = true
      try {
        const result = await EnumerateGPUs()
        this.gpuList = result
      } catch (e) {
        console.error('Error getting GPU info:', e)
        this.gpuList = []
      }
      this.loading = false
    },
    
    async refreshProcesses() {
      this.loadingProcesses = true
      try {
        const result = await EnumerateGPUProcesses()
        this.processList = result
      } catch (e) {
        console.error('Error getting processes:', e)
        this.processList = []
      }
      this.loadingProcesses = false
    },
    
    async checkNvidia() {
      this.checkingNvidia = true
      try {
        const result = await CheckNVIDIAStatus()
        this.nvidiaStatus = result
      } catch (e) {
        console.error('Error checking NVIDIA:', e)
        this.nvidiaStatus = { detected: false, nvmlLoaded: false, serviceRunning: false }
      }
      this.checkingNvidia = false
    },
    async refreshAll() {
      await Promise.all([
        this.refreshGPU(),
        this.refreshProcesses(),
        this.getIGPUMode(),
        this.checkNvidia(),
        this.refreshSSD(),
        this.loadIntelGPU(),
        this.loadGearStatus(),
        this.loadEPOTStatus()
      ])
      // Initial GPU status read (only once at startup)
      await this.pollGPUPref()
    },
    
    async startGPUStatusWatcher() {
      try {
        // Start backend watcher (uses WaitForMultipleObjects, no polling)
        await StartGPUStatusWatcher()
        
        // Listen for GPU status change events from backend
        EventsOn('gpu:status-change', (status) => {
          this.gpuPref = status
        })
        
        // No polling interval needed - purely event-driven
      } catch (e) {
        console.error('Failed to start GPU status watcher:', e)
      }
    },
    
    async stopGPUStatusWatcher() {
      // Remove event listener
      EventsOff('gpu:status-change')
      // Don't stop the backend watcher - keep it running for other components
    },

    async pollGPUPref() {
      try {
        // Read from Go-side cache (instant, no process spawn)
        const result = await GetGPUPrefStatusFromCache()
        this.gpuPref = result
      } catch (e) { /* silent */ }
    },

    async loadIntelGPU() {
      try {
        const result = await GetIntelGPUFrequency()
        this.intelGPU = result
        if (result.available) {
          this.freqMin = result.currentMin > 0 ? result.currentMin : result.minFreq
          this.freqMax = result.currentMax > 0 ? result.currentMax : result.maxFreq
        }
      } catch (e) {
        this.intelGPU.error = 'Failed to query Intel GPU'
      }
    },

    async refreshIGPUFreq() {
      this.freqTesting = true
      this.freqTestResult = null
      await this.loadIntelGPU()
      this.freqTesting = false
    },

    // GPU utilization polling (3s interval, only when IGPU tab visible)
    startUtilPolling() {
      if (this.utilTimer) return
      this.pollGPUUtil()  // immediate first read
      this.utilTimer = setInterval(() => {
        if (this.activeTab === 'c') this.pollGPUUtil()
      }, 3000)
    },

    stopUtilPolling() {
      if (this.utilTimer) {
        clearInterval(this.utilTimer)
        this.utilTimer = null
      }
    },

    async pollGPUUtil() {
      try {
        const v = await GetIntelGPUUtilization()
        this.gpuUtilization = v
      } catch (e) {
        this.gpuUtilization = -1
      }
    },

    onFreqMinInput() {
      if (this.freqMin > this.freqMax) this.freqMax = this.freqMin
    },

    onFreqMaxInput() {
      if (this.freqMax < this.freqMin) this.freqMin = this.freqMax
    },

    async applyFreqRange() {
      if (this.freqMin > this.freqMax) {
        this.freqTestResult = { success: false, message: 'Min cannot exceed Max' }
        return
      }
      this.freqTesting = true
      this.freqTestResult = null
      try {
        const result = await SetIntelGPUFrequencyRange(this.freqMin, this.freqMax)
        this.freqTestResult = result
      } catch (e) {
        this.freqTestResult = { success: false, message: 'Failed: ' + e }
      }
      this.freqTesting = false
    },

    async testFreq(type) {
      this.freqTesting = true
      this.freqTestResult = null
      try {
        const result = await TestIntelGPUFrequency(type)
        this.freqTestResult = result
        if (result.success) {
          this.freqMin = result.minFreq
          this.freqMax = result.maxFreq
        }
      } catch (e) {
        this.freqTestResult = { success: false, message: 'Test failed: ' + e }
      }
      this.freqTesting = false
    },

    async openDriverDownload() {
      try {
        const url = await GetIntelDriverDownloadURL()
        window.open(url, '_blank')
      } catch (e) {
        window.open('https://www.intel.cn/content/www/cn/zh/download-center/home.html', '_blank')
      }
    },

    async getIGPUMode() {
      try {
        const result = await GetIGPUMode()
        this.igpuStatus = result
      } catch (e) {
        console.error('Error getting IGPU mode:', e)
        this.igpuStatus = { available: false, mode: -1 }
      }
    },
    
    async setIGPUMode(mode) {
      this.settingMode = true
      this.settingResult = null
      try {
        const result = await SetIGPUMode(mode)
        this.settingResult = result
        if (result.success) {
          await this.getIGPUMode()
        }
      } catch (e) {
        this.settingResult = { success: false, message: 'Error: ' + e }
      }
      this.settingMode = false
    },

    gearModeLabel(val) {
      switch(val) {
        case 0: return 'Auto'
        case 1: return 'DGPU'
        case 2: return 'iGPU'
        default: return 'Unknown (' + val + ')'
      }
    },

    async loadGearStatus() {
      try {
        this.gearStatus = await GetGPUAutoGear()
      } catch(e) {
        this.gearStatus = { available: false, value: 0 }
      }
    },

    async setGear(value) {
      this.settingGear = true
      this.gearResult = null
      try {
        const result = await SetGPUAutoGear(value)
        this.gearResult = result
        if (result.success) {
          this.gearStatus = { available: true, value: value }
        }
      } catch(e) {
        this.gearResult = { success: false, message: 'Error: ' + e }
      }
      this.settingGear = false
    },

    async loadEPOTStatus() {
      this.epotRefreshing = true
      try {
        this.epotStatus = await GetEPOTStatus()
      } catch(e) {
        console.error('Error loading EPOT status:', e)
      }
      this.epotRefreshing = false
    },

    async uninstallDTT() {
      this.uninstallingDTT = true
      this.dttResult = null
      try {
        const message = await UninstallDTT()
        this.dttResult = { success: !message.includes('Error') && !message.includes('Failed'), message: message }
      } catch(e) {
        this.dttResult = { success: false, message: 'Error: ' + e }
      }
      this.uninstallingDTT = false
    },

    async uninstallDTTUI() {
      this.uninstallingDTTUI = true
      this.dttResult = null
      try {
        const message = await UninstallDTTUI()
        this.dttResult = { success: !message.includes('Error') && !message.includes('Failed'), message: message }
      } catch(e) {
        this.dttResult = { success: false, message: 'Error: ' + e }
      }
      this.uninstallingDTTUI = false
    }
  }
}
</script>

<style scoped>
.func-check-page {
  padding: 0;
}

/* Function Tabs */
.func-tabs {
  display: flex;
  gap: 8px;
  padding: 16px 20px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  overflow-x: auto;
}

.func-tab {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 16px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  cursor: pointer;
  transition: all 0.2s;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 500;
  white-space: nowrap;
}

.func-tab:hover {
  background: var(--bg-secondary);
  border-color: var(--accent-color);
  color: var(--text-primary);
}

.func-tab.active {
  background: linear-gradient(90deg, rgba(230, 63, 50, 0.15) 0%, rgba(230, 63, 50, 0.05) 100%);
  border-color: var(--accent-color);
  color: var(--accent-color);
}

.tab-icon {
  font-size: 16px;
}

.tab-label {
  font-weight: 500;
}

.func-content {
  padding: 20px;
}

.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 12px;
  padding: 20px;
  margin-bottom: 20px;
}

.placeholder-card {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 60px 20px;
  text-align: center;
}

.placeholder-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.6;
}

.placeholder-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.placeholder-desc {
  font-size: 13px;
  color: var(--text-secondary);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.card-title {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.spinner-small {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

/* GPU List */
.gpu-list {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.gpu-item {
  background: var(--bg-tertiary);
  border-radius: 8px;
  padding: 16px;
}

.gpu-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.gpu-index {
  font-weight: 600;
  color: var(--accent-color);
}

.gpu-badge {
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
}

.badge-dgpu {
  background: rgba(0, 102, 204, 0.2);
  color: #0066CC;
}

.badge-igpu {
  background: rgba(230, 63, 50, 0.2);
  color: #E63F32;
}

.gpu-info-single-row {
  grid-template-columns: 2fr 1fr 1fr !important;
}

.gpu-info-single-row .info-item {
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  padding: 12px 16px;
}

.gpu-info-single-row .info-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
  margin-top: 4px;
}

.hw-id {
  font-size: 11px;
  word-break: break-all;
  color: var(--text-secondary);
}

.gpu-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 11px;
  color: var(--text-tertiary);
  text-transform: uppercase;
}

.info-value {
  font-size: 14px;
  color: var(--text-primary);
  word-break: break-word;
}

/* Process List */
.process-list {
  display: flex;
  flex-direction: column;
}

.process-header {
  display: grid;
  grid-template-columns: 80px 1fr 100px;
  gap: 12px;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  margin-bottom: 8px;
}

.process-item {
  display: grid;
  grid-template-columns: 80px 1fr 100px;
  gap: 12px;
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-color);
  font-size: 13px;
}

.process-item:last-child {
  border-bottom: none;
}

.col-memory {
  text-align: right;
  color: var(--text-secondary);
}

/* IGPU Control */
.igpu-status {
  display: grid;
  grid-template-columns: 1fr 1fr 1fr;
  gap: 16px;
  margin-bottom: 20px;
}

.status-col-left,
.status-col-center,
.status-col-right {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.status-item {
  background: var(--bg-tertiary);
  padding: 12px 16px;
  border-radius: 6px;
}

.status-item-tall {
  flex: 1;
  display: flex;
  flex-direction: column;
  justify-content: center;
  min-height: 80px;
}

.status-item-tall .status-label {
  font-size: 11px;
}

.status-item-tall .status-value {
  font-size: 18px;
  font-weight: 700;
}

.status-label {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  text-transform: uppercase;
  margin-bottom: 4px;
}

.status-value {
  font-size: 14px;
  font-weight: 600;
}

.status-dis {
  color: var(--success-color);
}

.status-uma {
  color: var(--warning-color);
}

.status-smart {
  color: var(--accent-blue);
}

.status-na {
  color: var(--text-muted);
  opacity: 0.6;
}

.status-ok {
  color: var(--success-color);
}

.status-error {
  color: var(--error-color);
}

.mono {
  font-family: 'Cascadia Code', 'Consolas', monospace;
  font-variant-numeric: tabular-nums;
}

.live-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--success-color);
  margin-left: 8px;
  vertical-align: middle;
  animation: live-pulse 1.5s ease-in-out infinite;
}

@keyframes live-pulse {
  0%, 100% { opacity: 1; box-shadow: 0 0 0 0 rgba(74, 222, 128, 0.6); }
  50% { opacity: 0.6; box-shadow: 0 0 0 4px rgba(74, 222, 128, 0); }
}

.igpu-control {
  position: relative;
}

.control-info {
  background: var(--bg-tertiary);
  padding: 12px 16px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 13px;
  line-height: 1.8;
  color: var(--text-secondary);
}

.control-info strong {
  color: var(--text-primary);
}

.btn-group {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.igpu-freq-info {
  display: grid;
  gap: 10px;
  margin-bottom: 16px;
}

.freq-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.freq-label {
  font-size: 13px;
  color: var(--text-secondary);
  min-width: 100px;
}

.freq-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

/* Live frequency status grid */
.igpu-freq-status {
  margin-bottom: 16px;
}

.freq-stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 8px;
}

.freq-stat-item {
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  padding: 10px 12px;
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.freq-stat-label {
  font-size: 11px;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
}

.freq-stat-value {
  font-size: 15px;
  font-weight: 700;
  font-variant-numeric: tabular-nums;
  color: var(--text-primary);
}

.freq-stat-value.highlight {
  color: var(--lenovo-red);
}

.freq-stat-value.live {
  color: #22c55e;
}

.freq-stat-value.util-low {
  color: #22c55e;
}

.freq-stat-value.util-mid {
  color: #f59e0b;
}

.freq-stat-value.util-high {
  color: #ef4444;
}

.slider-val-inline {
  font-weight: 700;
  color: var(--lenovo-red);
}

.freq-control {
  background: rgba(255, 255, 255, 0.02);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 16px;
}

.freq-slider-group {
  margin-bottom: 16px;
}

.freq-slider-group label {
  display: block;
  font-size: 12px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.slider-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.freq-slider {
  flex: 1;
  height: 6px;
  -webkit-appearance: none;
  background: var(--bg-tertiary);
  border-radius: 3px;
  outline: none;
}

.freq-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  border-radius: 50%;
  background: var(--lenovo-red);
  cursor: pointer;
  transition: transform 0.15s ease;
}

.freq-slider::-webkit-slider-thumb:hover {
  transform: scale(1.15);
}

.slider-val {
  min-width: 60px;
  text-align: center;
  font-size: 14px;
  font-weight: 600;
  font-variant-numeric: tabular-nums;
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.1);
  padding: 4px 8px;
  border-radius: 4px;
}

.freq-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

.freq-actions .btn {
  flex: 1;
  min-width: 100px;
}

.error-state {
  padding: 12px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: var(--radius-md);
  color: var(--error-color);
  font-size: 13px;
}

.error-state p {
  margin: 0 0 10px 0;
}

.download-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  color: var(--accent-blue);
  font-size: 12px;
  text-decoration: none;
  padding: 6px 12px;
  background: rgba(59, 130, 246, 0.1);
  border-radius: 6px;
  transition: all 0.2s ease;
}

.download-link:hover {
  background: rgba(59, 130, 246, 0.2);
  color: #60a5fa;
}

.driver-ok {
  color: var(--success-color);
}

.driver-old {
  color: var(--warning-color);
}

.driver-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
  border-radius: 50%;
  font-size: 10px;
  margin-left: 6px;
}

.driver-badge.ok {
  background: rgba(74, 222, 128, 0.2);
  color: var(--success-color);
}

.driver-badge.old {
  background: rgba(250, 204, 21, 0.2);
  color: var(--warning-color);
}

.muted {
  opacity: 0.6;
}

.driver-warning {
  display: flex;
  align-items: flex-start;
  gap: 12px;
  padding: 12px 16px;
  background: rgba(250, 204, 21, 0.1);
  border: 1px solid rgba(250, 204, 21, 0.3);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
}

.warning-icon {
  font-size: 20px;
  line-height: 1;
}

.warning-content p {
  margin: 0 0 8px 0;
  font-size: 13px;
  color: var(--text-primary);
}

.warning-content p:last-child {
  margin-bottom: 0;
}

.btn-sm {
  padding: 6px 12px;
  font-size: 12px;
}

.card-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 3px 8px;
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red);
  border-radius: 4px;
  letter-spacing: 0.5px;
}

.loading-overlay {
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(13, 13, 13, 0.8);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
}

.spinner {
  width: 32px;
  height: 32px;
  border: 3px solid var(--border-color);
  border-top-color: var(--accent-color);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

.loading-overlay p {
  margin-top: 12px;
  color: var(--text-secondary);
}

.result-message {
  margin-top: 16px;
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 13px;
}

.result-message.success {
  background: rgba(76, 175, 80, 0.1);
  border: 1px solid rgba(76, 175, 80, 0.3);
  color: var(--success-color);
}

.result-message.error {
  background: rgba(244, 67, 54, 0.1);
  border: 1px solid rgba(244, 67, 54, 0.3);
  color: var(--error-color);
}

/* NVIDIA Status */
.nvidia-status {
  text-align: center;
}

.nvidia-badge {
  padding: 16px;
  border-radius: 8px;
  font-size: 16px;
  font-weight: 600;
  margin-bottom: 12px;
}

.nvidia-badge.detected {
  background: rgba(118, 185, 0, 0.1);
  border: 1px solid rgba(118, 185, 0, 0.3);
  color: #76B900;
}

.nvidia-badge.not-detected {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  color: var(--text-secondary);
}

.nvidia-details {
  text-align: left;
  background: var(--bg-tertiary);
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 13px;
  color: var(--text-secondary);
}

.nvidia-details strong {
  color: var(--text-primary);
}

/* Diagnostic Grid */
.diag-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(150px, 1fr));
  gap: 12px;
}

.diag-item {
  background: var(--bg-tertiary);
  padding: 12px 16px;
  border-radius: 6px;
  text-align: center;
}

.diag-label {
  display: block;
  font-size: 11px;
  color: var(--text-tertiary);
  text-transform: uppercase;
  margin-bottom: 4px;
}

.diag-value {
  font-size: 18px;
  font-weight: 600;
  color: var(--text-primary);
}

.empty-state {
  text-align: center;
  padding: 40px;
  color: var(--text-tertiary);
}

/* ── SSD Tab ───────────────────────────────────────────── */
.ssd-item {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 16px;
  margin-bottom: 12px;
}
.ssd-header { display: flex; align-items: center; gap: 10px; margin-bottom: 14px; }
.ssd-index {
  font-size: 11px; font-weight: 700; color: var(--lenovo-red);
  background: rgba(230,63,50,0.1); padding: 2px 8px; border-radius: 10px; flex-shrink: 0;
}
.ssd-name { font-size: 15px; font-weight: 600; color: var(--text-primary); flex: 1; }
.ssd-badge { font-size: 11px; font-weight: 600; padding: 2px 10px; border-radius: 10px; }
.badge-capable { background: rgba(16,185,129,0.15); color: #10B981; border: 1px solid rgba(16,185,129,0.3); }
.badge-limited { background: rgba(107,114,128,0.1); color: var(--text-secondary); border: 1px solid rgba(107,114,128,0.2); }
.ssd-info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 8px; margin-bottom: 14px; }
.ssd-mode-control { background: var(--bg-tertiary); border-radius: var(--radius-md); padding: 14px; }
.mode-section-label { font-size: 12px; font-weight: 600; color: var(--text-secondary); margin-bottom: 10px; text-transform: uppercase; letter-spacing: 0.05em; }
.ssd-mode-buttons { display: flex; gap: 8px; flex-wrap: wrap; }
.ssd-mode-btn {
  display: flex; flex-direction: column; align-items: center; gap: 2px;
  padding: 10px 14px; border: 2px solid var(--border-color);
  border-radius: var(--radius-md); background: var(--bg-card);
  cursor: pointer; transition: var(--transition); min-width: 72px;
}
.ssd-mode-btn:hover:not(:disabled) { border-color: var(--lenovo-red); background: rgba(230,63,50,0.05); }
.ssd-mode-btn.active { border-color: var(--lenovo-red); background: rgba(230,63,50,0.1); box-shadow: 0 0 12px rgba(230,63,50,0.25); }
.ssd-mode-btn.loading { opacity: 0.6; cursor: wait; }
.ssd-mode-btn:disabled { cursor: not-allowed; opacity: 0.5; }
.mode-abbr { font-size: 13px; font-weight: 700; color: var(--text-primary); }
.mode-full { font-size: 10px; color: var(--text-secondary); }
.mode-hint { margin-top: 10px; font-size: 12px; color: var(--text-secondary); display: flex; align-items: center; gap: 6px; }
.hint-dot { width: 6px; height: 6px; border-radius: 50%; background: var(--lenovo-red); flex-shrink: 0; }
.mono-sm { font-family: 'Consolas','Monaco',monospace; font-size: 12px; }
.ssd-error {
  margin-top: 10px; padding: 8px 12px;
  background: rgba(239,68,68,0.08); border: 1px solid rgba(239,68,68,0.2);
  border-radius: var(--radius-sm); color: #EF4444; font-size: 12px; display: flex; align-items: center; gap: 6px;
}
.mode-badge.mode-active { color: var(--lenovo-red); font-weight: 700; }
.card-title-normal { text-transform: none !important; letter-spacing: normal !important; }
.placeholder-card { text-align: center; padding: 60px 40px; }
.placeholder-icon { font-size: 40px; margin-bottom: 16px; }
.placeholder-title { font-size: 18px; font-weight: 600; margin-bottom: 8px; }
.placeholder-desc { font-size: 14px; color: var(--text-secondary); }

@keyframes spin { to { transform: rotate(360deg); } }

/* Auto Gear & EPOT */
.auto-gear-info { padding: 16px; }
.gear-status-row { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; border-bottom: 1px solid var(--border-color); }
.gear-desc-row { display: flex; justify-content: space-between; align-items: center; padding: 12px 0; color: var(--text-secondary); font-size: 13px; }
.gear-control { padding: 16px; border-top: 1px solid var(--border-color); }
.gear-buttons { display: flex; gap: 12px; }
.epot-grid { padding: 16px; }
.epot-row { display: flex; align-items: center; gap: 12px; padding: 10px 0; border-bottom: 1px solid var(--border-color); font-size: 13px; }
.epot-row:last-child { border-bottom: none; }
.epot-label { width: 180px; font-weight: 500; color: var(--text-primary); font-family: 'Consolas', monospace; }
.epot-desc { flex: 1; color: var(--text-secondary); }
.epot-value { width: 80px; text-align: right; font-family: 'Consolas', monospace; font-weight: 600; color: var(--lenovo-red); }
.live-indicator { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-secondary); }
.epot-badge { display: inline-block; margin-left: 8px; padding: 2px 8px; background: rgba(230,63,50,0.15); border: 1px solid rgba(230,63,50,0.3); border-radius: 12px; color: var(--lenovo-red); font-size: 11px; font-weight: 600; font-family: 'Consolas', monospace; }

/* Gear Control Buttons */
.gear-btn-group { display: flex; gap: 12px; flex-wrap: wrap; }
.btn-gear {
  display: flex; align-items: center; gap: 6px;
  padding: 12px 20px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  cursor: pointer;
  transition: var(--transition);
  font-size: 14px;
  font-weight: 600;
}
.btn-gear:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
.btn-gear:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-gear.active { box-shadow: 0 0 16px rgba(0,0,0,0.2); }
.btn-gear-auto { border-color: #10B981; color: #10B981; }
.btn-gear-auto:hover:not(:disabled) { background: rgba(16,185,129,0.1); }
.btn-gear-auto.active { background: rgba(16,185,129,0.2); box-shadow: 0 0 16px rgba(16,185,129,0.4); }
.btn-gear-dgpu { border-color: #3B82F6; color: #3B82F6; }
.btn-gear-dgpu:hover:not(:disabled) { background: rgba(59,130,246,0.1); }
.btn-gear-dgpu.active { background: rgba(59,130,246,0.2); box-shadow: 0 0 16px rgba(59,130,246,0.4); }
.btn-gear-igpu { border-color: #F59E0B; color: #F59E0B; }
.btn-gear-igpu:hover:not(:disabled) { background: rgba(245,158,11,0.1); }
.btn-gear-igpu.active { background: rgba(245,158,11,0.2); box-shadow: 0 0 16px rgba(245,158,11,0.4); }
.gear-icon { font-size: 16px; }
.gear-badge {
  display: inline-flex; align-items: center;
  padding: 4px 12px;
  border-radius: 20px;
  font-size: 13px;
  font-weight: 600;
  margin-right: 8px;
}
.gear-badge.gear-auto { background: rgba(16,185,129,0.15); color: #10B981; border: 1px solid rgba(16,185,129,0.3); }
.gear-badge.gear-dgpu { background: rgba(59,130,246,0.15); color: #3B82F6; border: 1px solid rgba(59,130,246,0.3); }
.gear-badge.gear-igpu { background: rgba(245,158,11,0.15); color: #F59E0B; border: 1px solid rgba(245,158,11,0.3); }
.gear-badge.gear-na { background: rgba(107,114,128,0.1); color: var(--text-secondary); border: 1px solid rgba(107,114,128,0.2); }
.gear-badge.gear-epot { background: rgba(230,63,50,0.15); color: var(--lenovo-red); border: 1px solid rgba(230,63,50,0.3); }

/* DTT Uninstall */
.dtt-uninstall-content { padding: 16px; }
.dtt-desc { margin-bottom: 16px; color: var(--text-secondary); font-size: 13px; }
.dtt-desc p { margin: 0; }
.dtt-buttons { display: flex; gap: 12px; flex-wrap: wrap; }
.btn-dtt {
  display: flex; align-items: center; justify-content: center;
  padding: 12px 20px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-card);
  cursor: pointer;
  transition: var(--transition);
  font-size: 14px;
  font-weight: 600;
  min-width: 140px;
}
.btn-dtt:hover:not(:disabled) { transform: translateY(-2px); box-shadow: 0 4px 12px rgba(0,0,0,0.15); }
.btn-dtt:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-dtt-main { border-color: #EF4444; color: #EF4444; }
.btn-dtt-main:hover:not(:disabled) { background: rgba(239,68,68,0.1); }
.btn-dtt-ui { border-color: #F59E0B; color: #F59E0B; }
.btn-dtt-ui:hover:not(:disabled) { background: rgba(245,158,11,0.1); }

</style>
