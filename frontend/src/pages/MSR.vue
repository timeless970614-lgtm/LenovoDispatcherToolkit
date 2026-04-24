<template>
  <div class="msr-page">
    <div class="msr-header">
      <h2 class="msr-title">MSR (Model Specific Registers)</h2>
      <p class="msr-subtitle">Intel Power Management & MSR Access Tool</p>
    </div>

    <!-- MSR Settings - Password Protected -->
    <div class="msr-card advanced-section">
      <div class="card-header advanced-toggle" @click="toggleAdvanced">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          MSR Settings
        </span>
        <svg :class="['chevron-icon', { 'chevron-open': advancedUnlocked }]" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="6 9 12 15 18 9"/>
        </svg>
      </div>

      <!-- Password Prompt (shown when collapsed & not unlocked) -->
      <div v-if="!advancedUnlocked" class="password-prompt">
        <div class="password-row">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px; flex-shrink:0;">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0110 0v4"/>
          </svg>
          <input 
            type="password" 
            class="password-input" 
            v-model="advancedPassword" 
            placeholder="Enter password to unlock"
            @keydown.enter="unlockAdvanced"
            ref="passwordInput"
          />
          <button class="btn-unlock" @click="unlockAdvanced">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M15 3h4a2 2 0 012 2v14a2 2 0 01-2 2h-4"/>
              <polyline points="10 17 15 12 10 7"/>
              <line x1="15" y1="12" x2="3" y2="12"/>
            </svg>
            Unlock
          </button>
        </div>
        <div v-if="passwordError" class="password-error">{{ passwordError }}</div>
      </div>

      <!-- Unlocked Content -->
      <div v-if="advancedUnlocked" class="advanced-content">
    <!-- MSR Status Card -->
    <div class="msr-card status-card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2L2 7l10 5 10-5-10-5z"/>
            <path d="M2 17l10 5 10-5"/>
            <path d="M2 12l10 5 10-5"/>
          </svg>
          MSR Driver Status
        </span>
        <span :class="['status-badge', driverStatus]">{{ driverStatusText }}</span>
      </div>
      <div class="card-body">
        <div class="status-grid">
          <div class="status-item">
            <span class="status-label">WinRing0</span>
            <span :class="['status-value', winring0Available ? 'ok' : 'not-ok']">
              {{ winring0Available ? 'Available' : 'Not Found' }}
            </span>
          </div>
          <div class="status-item">
            <span class="status-label">MSR Driver</span>
            <span :class="['status-value', msrDriverAvailable ? 'ok' : 'not-ok']">
              {{ msrDriverAvailable ? 'Available' : 'Not Found' }}
            </span>
          </div>
          <div class="status-item">
            <span class="status-label">Admin Rights</span>
            <span :class="['status-value', isAdmin ? 'ok' : 'warning']">
              {{ isAdmin ? 'Yes' : 'No' }}
            </span>
          </div>
        </div>
        <div class="status-actions" v-if="!winring0Available && !msrDriverAvailable">
          <button class="btn-install" @click="showInstallHelp = true">
            Install Driver
          </button>
        </div>
      </div>
    </div>

    <!-- MSR Registers Grid -->
    <div class="msr-grid">
      <!-- Energy Performance Bias -->
      <div class="msr-card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
            </svg>
            Energy Perf Bias (0x1B0)
          </span>
        </div>
        <div class="card-body">
          <div class="msr-value-display">
            <span class="msr-hex">0x{{ energyBiasValue.toString(16).toUpperCase().padStart(2, '0') }}</span>
            <span class="msr-dec">{{ energyBiasValue }} / 15</span>
          </div>
          <div class="msr-slider">
            <input 
              type="range" 
              min="0" 
              max="15" 
              v-model="energyBiasValue"
              @change="updateEnergyBias"
              :disabled="!canWriteMSR"
            />
            <div class="slider-labels">
              <span>Performance</span>
              <span>Balance</span>
              <span>Power Save</span>
            </div>
          </div>
          <div class="msr-desc">
            0=Max Perf, 6=Balance, 15=Max Energy Save
          </div>
        </div>
      </div>

      <!-- Turbo Boost -->
      <div class="msr-card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
            </svg>
            Turbo Boost (0x1A0)
          </span>
        </div>
        <div class="card-body">
          <div class="toggle-row">
            <span class="toggle-label">Turbo Boost</span>
            <label class="toggle-switch">
              <input 
                type="checkbox" 
                v-model="turboBoostEnabled"
                @change="updateTurboBoost"
                :disabled="!canWriteMSR"
              />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="msr-desc">
            IA32_MISC_ENABLE bit 38 controls Turbo Boost
          </div>
        </div>
      </div>

      <!-- P-State Control -->
      <div class="msr-card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <path d="M12 6v6l4 2"/>
            </svg>
            P-State Control (0x198)
          </span>
        </div>
        <div class="card-body">
          <div class="msr-readonly">
            <span class="readonly-label">Current P-State</span>
            <span class="readonly-value">{{ currentPState }}</span>
          </div>
          <div class="msr-readonly">
            <span class="readonly-label">Target P-State</span>
            <span class="readonly-value">{{ targetPState }}</span>
          </div>
          <button class="btn-refresh" @click="readPState" :disabled="!canReadMSR">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/>
              <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
            </svg>
            Refresh
          </button>
        </div>
      </div>

      <!-- RAPL Power Limit -->
      <div class="msr-card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18.36 6.64a9 9 0 1 1-12.73 0"/>
              <line x1="12" y1="2" x2="12" y2="12"/>
            </svg>
            RAPL Power Limit (0x610)
          </span>
        </div>
        <div class="card-body">
          <div class="power-input-row">
            <span class="input-label">Package Power (W)</span>
            <input 
              type="number" 
              v-model="packagePowerLimit"
              class="power-input"
              :disabled="!canWriteMSR"
            />
          </div>
          <div class="power-input-row">
            <span class="input-label">Time Window (ms)</span>
            <input 
              type="number" 
              v-model="timeWindow"
              class="power-input"
              :disabled="!canWriteMSR"
            />
          </div>
          <button class="btn-apply" @click="applyPowerLimit" :disabled="!canWriteMSR">
            Apply Limit
          </button>
        </div>
      </div>

      <!-- Clock Modulation -->
      <div class="msr-card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="2" x2="12" y2="4"/>
              <line x1="12" y1="20" x2="12" y2="22"/>
              <line x1="4.93" y1="4.93" x2="6.34" y2="6.34"/>
              <line x1="17.66" y1="17.66" x2="19.07" y2="19.07"/>
              <line x1="2" y1="12" x2="4" y2="12"/>
              <line x1="20" y1="12" x2="22" y2="12"/>
            </svg>
            Clock Modulation (0x19A)
          </span>
        </div>
        <div class="card-body">
          <div class="toggle-row">
            <span class="toggle-label">Enable Modulation</span>
            <label class="toggle-switch">
              <input 
                type="checkbox" 
                v-model="clockModEnabled"
                @change="updateClockMod"
                :disabled="!canWriteMSR"
              />
              <span class="toggle-slider"></span>
            </label>
          </div>
          <div class="duty-slider" v-if="clockModEnabled">
            <span class="slider-label">Duty Cycle: {{ dutyCycle }}%</span>
            <input 
              type="range" 
              min="1" 
              max="14" 
              v-model="dutyCycleValue"
              @change="updateClockMod"
              :disabled="!canWriteMSR"
            />
          </div>
          <div class="msr-desc">
            T-state throttling for thermal control
          </div>
        </div>
      </div>

      <!-- Uncore Frequency -->
      <div class="msr-card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="4" y="4" width="16" height="16" rx="2"/>
              <rect x="9" y="9" width="6" height="6"/>
              <line x1="9" y1="1" x2="9" y2="4"/>
              <line x1="15" y1="1" x2="15" y2="4"/>
              <line x1="9" y1="20" x2="9" y2="23"/>
              <line x1="15" y1="20" x2="15" y2="23"/>
              <line x1="20" y1="9" x2="23" y2="9"/>
              <line x1="20" y1="14" x2="23" y2="14"/>
              <line x1="1" y1="9" x2="4" y2="9"/>
              <line x1="1" y1="14" x2="4" y2="14"/>
            </svg>
            Uncore Ratio (0x620)
          </span>
        </div>
        <div class="card-body">
          <div class="ratio-inputs">
            <div class="ratio-input">
              <span class="input-label">Max Ratio</span>
              <input 
                type="number" 
                v-model="uncoreMaxRatio"
                min="0" 
                max="127"
                :disabled="!canWriteMSR"
              />
            </div>
            <div class="ratio-input">
              <span class="input-label">Min Ratio</span>
              <input 
                type="number" 
                v-model="uncoreMinRatio"
                min="0" 
                max="127"
                :disabled="!canWriteMSR"
              />
            </div>
          </div>
          <div class="lock-warning" v-if="uncoreMaxRatio === uncoreMinRatio">
            ⚠️ Frequency locked (max = min)
          </div>
          <button class="btn-apply" @click="applyUncoreRatio" :disabled="!canWriteMSR">
            Apply Ratio
          </button>
        </div>
      </div>
    </div>

    <!-- MSR Register List -->
    <div class="msr-card msr-fullwidth">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <line x1="8" y1="6" x2="21" y2="6"/>
            <line x1="8" y1="12" x2="21" y2="12"/>
            <line x1="8" y1="18" x2="21" y2="18"/>
            <line x1="3" y1="6" x2="3.01" y2="6"/>
            <line x1="3" y1="12" x2="3.01" y2="12"/>
            <line x1="3" y1="18" x2="3.01" y2="18"/>
          </svg>
          MSR Register Reference
        </span>
      </div>
      <div class="card-body">
        <table class="msr-table">
          <thead>
            <tr>
              <th>Address</th>
              <th>Name</th>
              <th>Description</th>
              <th>Access</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="reg in msrRegisters" :key="reg.address">
              <td class="mono">0x{{ reg.address.toString(16).toUpperCase().padStart(3, '0') }}</td>
              <td>{{ reg.name }}</td>
              <td>{{ reg.description }}</td>
              <td>
                <span :class="['access-badge', reg.readonly ? 'ro' : 'rw']">
                  {{ reg.readonly ? 'RO' : 'RW' }}
                </span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
      </div>
    </div>

    <!-- Install Help Modal -->
    <div class="modal-overlay" v-if="showInstallHelp" @click="showInstallHelp = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Install MSR Driver</h3>
          <button class="modal-close" @click="showInstallHelp = false">&times;</button>
        </div>
        <div class="modal-body">
          <p>To access MSR registers, you need to install a kernel driver:</p>
          <h4>Option 1: WinRing0 (Recommended)</h4>
          <ol>
            <li>Download WinRing0 from GitHub</li>
            <li>Copy WinRing0x64.sys to C:\Windows\System32\drivers\</li>
            <li>Run as Administrator:
              <code>sc create WinRing0 type= kernel binPath= C:\Windows\System32\drivers\WinRing0x64.sys</code>
            </li>
            <li>Start driver: <code>sc start WinRing0</code></li>
          </ol>
          <h4>Option 2: MSR Driver (Custom)</h4>
          <ol>
            <li>Build driver from C:\LenovoDispatcher\MSR\driver\</li>
            <li>Sign driver or enable test signing</li>
            <li>Install using provided INF file</li>
          </ol>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MSR',
  props: {
    theme: {
      type: String,
      default: 'dark'
    }
  },
  data() {
    return {
      // Driver status
      winring0Available: false,
      msrDriverAvailable: false,
      isAdmin: false,
      
      // MSR Values
      energyBiasValue: 6,
      turboBoostEnabled: true,
      currentPState: 'N/A',
      targetPState: 'N/A',
      packagePowerLimit: 65,
      timeWindow: 1000,
      clockModEnabled: false,
      dutyCycleValue: 8,
      uncoreMaxRatio: 30,
      uncoreMinRatio: 8,
      
      // UI State
      showInstallHelp: false,
      
      // Password Protection
      advancedUnlocked: false,
      advancedPassword: '',
      passwordError: '',
      
      // MSR Register List (from PPT)
      msrRegisters: [
        { address: 0x1B0, name: 'IA32_ENERGY_PERF_BIAS', description: 'Energy Performance Bias Hint (0-15)', readonly: false },
        { address: 0x198, name: 'IA32_PERF_CTL', description: 'Performance Control - P-state target', readonly: false },
        { address: 0x199, name: 'IA32_PERF_STATUS', description: 'Current P-state (read-only)', readonly: true },
        { address: 0x1A0, name: 'IA32_MISC_ENABLE', description: 'Turbo Boost control', readonly: false },
        { address: 0x19A, name: 'IA32_CLOCK_MODULATION', description: 'Clock Modulation (T-state)', readonly: false },
        { address: 0xE7, name: 'IA32_APERF', description: 'Actual Performance Counter', readonly: true },
        { address: 0xE8, name: 'IA32_MPERF', description: 'Maximum Performance Counter', readonly: true },
        { address: 0x620, name: 'MSR_UNCORE_RATIO_LIMIT', description: 'Uncore Frequency Limits', readonly: false },
        { address: 0x606, name: 'MSR_RAPL_POWER_UNIT', description: 'RAPL Units (power/energy/time)', readonly: true },
        { address: 0x610, name: 'MSR_PKG_POWER_LIMIT', description: 'Package Power Limit', readonly: false },
        { address: 0x611, name: 'MSR_PKG_ENERGY_STATUS', description: 'Package Energy Consumed', readonly: true },
        { address: 0x638, name: 'MSR_PP0_POWER_LIMIT', description: 'CPU Cores Power Limit', readonly: false },
        { address: 0x618, name: 'MSR_DRAM_POWER_LIMIT', description: 'DRAM Power Limit', readonly: false },
      ]
    }
  },
  computed: {
    driverStatus() {
      if (this.winring0Available || this.msrDriverAvailable) return 'ok'
      return 'warning'
    },
    driverStatusText() {
      if (this.winring0Available || this.msrDriverAvailable) return 'Ready'
      return 'Driver Required'
    },
    canReadMSR() {
      return this.winring0Available || this.msrDriverAvailable
    },
    canWriteMSR() {
      return (this.winring0Available || this.msrDriverAvailable) && this.isAdmin
    },
    dutyCycle() {
      return this.dutyCycleValue * 6.25
    }
  },
  mounted() {
    this.checkDriverStatus()
    this.checkAdminRights()
  },
  methods: {
    async checkDriverStatus() {
      // Check for WinRing0
      try {
        // This would be implemented via Go backend
        // For now, simulate
        this.winring0Available = false
        this.msrDriverAvailable = false
      } catch (e) {
        console.error('Driver check failed:', e)
      }
    },
    checkAdminRights() {
      // Check if running as admin
      // This would be implemented via Go backend
      this.isAdmin = false
    },
    updateEnergyBias() {
      console.log('Setting Energy Perf Bias to:', this.energyBiasValue)
      // Would call backend to write MSR
    },
    updateTurboBoost() {
      console.log('Setting Turbo Boost to:', this.turboBoostEnabled)
      // Would call backend to write MSR
    },
    readPState() {
      console.log('Reading P-State...')
      // Would call backend to read MSR
      this.currentPState = '0x1D00'
      this.targetPState = '0x1E00'
    },
    applyPowerLimit() {
      console.log('Applying power limit:', this.packagePowerLimit, 'W')
      // Would call backend to write MSR
    },
    updateClockMod() {
      console.log('Clock modulation:', this.clockModEnabled, 'Duty:', this.dutyCycleValue)
      // Would call backend to write MSR
    },
    applyUncoreRatio() {
      console.log('Uncore ratio - Max:', this.uncoreMaxRatio, 'Min:', this.uncoreMinRatio)
      // Would call backend to write MSR
    },
    toggleAdvanced() {
      if (this.advancedUnlocked) {
        this.advancedUnlocked = !this.advancedUnlocked
      }
    },
    unlockAdvanced() {
      if (this.advancedPassword === 'Lenovo2026') {
        this.advancedUnlocked = true
        this.passwordError = ''
        this.advancedPassword = ''
      } else {
        this.passwordError = 'Need Dispatcher owner check or contact zhoushang2'
        setTimeout(() => { this.passwordError = '' }, 3000)
      }
    }
  }
}
</script>

<style scoped>
.msr-page {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;
}

.msr-header {
  margin-bottom: 24px;
}

.msr-title {
  font-size: 24px;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 8px 0;
}

.msr-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.msr-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
  gap: 16px;
  margin-bottom: 16px;
}

.msr-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}

.msr-fullwidth {
  grid-column: 1 / -1;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-color);
}

.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-body {
  padding: 16px;
}

.status-badge {
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
}

.status-badge.ok {
  background: rgba(74, 222, 128, 0.2);
  color: #4ADE80;
}

.status-badge.warning {
  background: rgba(251, 191, 36, 0.2);
  color: #FBBF24;
}

.status-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 16px;
  margin-bottom: 16px;
}

.status-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-label {
  font-size: 12px;
  color: var(--text-tertiary);
}

.status-value {
  font-size: 14px;
  font-weight: 600;
}

.status-value.ok {
  color: #4ADE80;
}

.status-value.not-ok {
  color: #EF4444;
}

.status-value.warning {
  color: #FBBF24;
}

.status-actions {
  display: flex;
  justify-content: flex-end;
}

.btn-install {
  padding: 8px 16px;
  background: var(--lenovo-red);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}

.btn-install:hover {
  background: var(--lenovo-red-light);
}

.msr-value-display {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
}

.msr-hex {
  font-family: 'Consolas', monospace;
  font-size: 18px;
  font-weight: 600;
  color: var(--lenovo-red);
}

.msr-dec {
  font-size: 14px;
  color: var(--text-secondary);
}

.msr-slider {
  margin-bottom: 12px;
}

.msr-slider input[type="range"] {
  width: 100%;
  height: 6px;
  -webkit-appearance: none;
  background: var(--bg-tertiary);
  border-radius: 3px;
  outline: none;
}

.msr-slider input[type="range"]::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 18px;
  height: 18px;
  background: var(--lenovo-red);
  border-radius: 50%;
  cursor: pointer;
}

.slider-labels {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

.msr-desc {
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.5;
}

.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.toggle-label {
  font-size: 14px;
  color: var(--text-primary);
}

.toggle-switch {
  position: relative;
  width: 48px;
  height: 24px;
}

.toggle-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: var(--bg-tertiary);
  border-radius: 24px;
  transition: 0.3s;
}

.toggle-slider:before {
  position: absolute;
  content: "";
  height: 18px;
  width: 18px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: 0.3s;
}

input:checked + .toggle-slider {
  background: var(--lenovo-red);
}

input:checked + .toggle-slider:before {
  transform: translateX(24px);
}

input:disabled + .toggle-slider {
  opacity: 0.5;
  cursor: not-allowed;
}

.msr-readonly {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid var(--border-color);
}

.readonly-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.readonly-value {
  font-family: 'Consolas', monospace;
  font-size: 14px;
  color: var(--text-primary);
}

.btn-refresh {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  padding: 8px 16px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 13px;
  cursor: pointer;
  transition: var(--transition);
}

.btn-refresh:hover:not(:disabled) {
  background: var(--bg-card-hover);
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.power-input-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}

.input-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.power-input {
  width: 100px;
  padding: 6px 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 14px;
  text-align: right;
}

.power-input:focus {
  outline: none;
  border-color: var(--lenovo-red);
}

.btn-apply {
  width: 100%;
  padding: 10px;
  background: var(--lenovo-red);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}

.btn-apply:hover:not(:disabled) {
  background: var(--lenovo-red-light);
}

.btn-apply:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.duty-slider {
  margin: 12px 0;
}

.slider-label {
  display: block;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.ratio-inputs {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
}

.ratio-input {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.ratio-input input {
  padding: 8px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 14px;
}

.lock-warning {
  padding: 8px 12px;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: #FBBF24;
  margin-bottom: 12px;
}

.msr-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 13px;
}

.msr-table th {
  text-align: left;
  padding: 10px 12px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-weight: 600;
  border-bottom: 1px solid var(--border-color);
}

.msr-table td {
  padding: 10px 12px;
  border-bottom: 1px solid var(--border-color);
  color: var(--text-primary);
}

.msr-table tr:hover td {
  background: var(--bg-card-hover);
}

.mono {
  font-family: 'Consolas', monospace;
}

.access-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}

.access-badge.ro {
  background: rgba(96, 165, 250, 0.2);
  color: #60A5FA;
}

.access-badge.rw {
  background: rgba(74, 222, 128, 0.2);
  color: #4ADE80;
}

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-content {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  width: 90%;
  max-width: 500px;
  max-height: 80vh;
  overflow: auto;
}

.modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
}

.modal-header h3 {
  margin: 0;
  font-size: 18px;
  color: var(--text-primary);
}

.modal-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  font-size: 24px;
  cursor: pointer;
}

.modal-body {
  padding: 20px;
  color: var(--text-primary);
  line-height: 1.6;
}

.modal-body h4 {
  margin: 16px 0 8px 0;
  color: var(--lenovo-red);
}

.modal-body ol {
  margin: 8px 0;
  padding-left: 20px;
}

.modal-body li {
  margin: 8px 0;
}

.modal-body code {
  display: block;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  font-family: 'Consolas', monospace;
  font-size: 12px;
  margin: 8px 0;
  word-break: break-all;
}

/* Advanced Section - Password Protected */
.advanced-section {
  border-color: rgba(245, 158, 11, 0.3);
}

.advanced-toggle {
  cursor: pointer;
  user-select: none;
  transition: background 0.15s;
  border-radius: var(--radius-lg);
  margin: -16px -16px 0 -16px;
  padding: 16px 20px;
}

.advanced-toggle:hover {
  background: rgba(245, 158, 11, 0.05);
}

.chevron-icon {
  transition: transform 0.25s ease;
  color: var(--text-tertiary);
}

.chevron-icon.chevron-open {
  transform: rotate(180deg);
}

.password-prompt {
  padding: 16px 0 0 0;
}

.password-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.password-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 13px;
  font-family: 'Consolas', monospace;
  outline: none;
  transition: border-color 0.2s;
}

.password-input:focus {
  border-color: #F59E0B;
}

.password-input::placeholder {
  color: var(--text-tertiary);
  font-family: inherit;
}

.btn-unlock {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.3);
  border-radius: var(--radius-md);
  color: #F59E0B;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  white-space: nowrap;
}

.btn-unlock:hover {
  background: rgba(245, 158, 11, 0.2);
  border-color: #F59E0B;
}

.password-error {
  margin-top: 8px;
  font-size: 12px;
  color: #EF4444;
  padding: 6px 10px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: var(--radius-md);
}

.advanced-content {
  padding-top: 16px;
}
</style>