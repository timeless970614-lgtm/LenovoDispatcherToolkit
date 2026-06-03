<template>
  <div class="msr-page">
    <!-- Page Header -->
    <div class="msr-header">
      <h2 class="msr-title">MSR (Model Specific Registers)</h2>
      <p class="msr-subtitle">Intel Power Management &amp; MSR Access Tool</p>
    </div>

    <!-- Driver Status Banner -->
    <div class="msr-banner">
      <div class="banner-icon">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 2L2 7l10 5 10-5-10-5z"/>
          <path d="M2 17l10 5 10-5"/>
          <path d="M2 12l10 5 10-5"/>
        </svg>
      </div>
      <div class="banner-info">
        <span class="banner-label">Driver Status</span>
        <span :class="['banner-value', driverStatus]">{{ driverStatusText }}</span>
      </div>
      <div class="banner-meta">
        <span class="meta-item">
          <span class="meta-label">WinRing0</span>
          <span :class="['meta-dot', winring0Available ? 'green' : 'red']"></span>
        </span>
        <span class="meta-item">
          <span class="meta-label">MSR Driver</span>
          <span :class="['meta-dot', msrDriverAvailable ? 'green' : 'red']"></span>
        </span>
        <span class="meta-item">
          <span class="meta-label">Admin</span>
          <span :class="['meta-dot', isAdmin ? 'green' : 'yellow']"></span>
        </span>
      </div>
      <button v-if="!winring0Available && !msrDriverAvailable" class="btn-install" @click="showInstallHelp = true">
        Install Driver
      </button>
    </div>

    <!-- Password Protected Advanced Settings -->
    <div :class="['msr-card', 'advanced-card', { unlocked: advancedUnlocked }]">
      <div class="card-header" :class="['advanced-toggle', { hoverable: !advancedUnlocked }]" @click="toggleAdvanced">
        <span class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
          </svg>
          Advanced MSR Settings
        </span>
        <div class="header-right">
          <span v-if="!advancedUnlocked" class="lock-hint">Password Protected</span>
          <span v-else class="unlocked-hint">Unlocked</span>
          <svg :class="['chevron-icon', { 'chevron-open': advancedUnlocked }]" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="6 9 12 15 18 9"/>
          </svg>
        </div>
      </div>

      <!-- Password Gate -->
      <div v-if="!advancedUnlocked" class="password-gate">
        <div class="password-row">
          <div class="password-icon">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
              <path d="M7 11V7a5 5 0 0110 0v4"/>
            </svg>
          </div>
          <input
            type="password"
            class="password-input"
            v-model="advancedPassword"
            placeholder="Enter password to unlock advanced settings"
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
        <div v-if="passwordError" class="password-error">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/>
            <line x1="12" y1="8" x2="12" y2="12"/>
            <line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          {{ passwordError }}
        </div>
      </div>

      <!-- Advanced Content (Unlocked) -->
      <div v-if="advancedUnlocked" class="advanced-content">
        <!-- 2-column grid: left=controls, right=status -->
        <div class="controls-grid">
          <!-- Left Column: Controls -->
          <div class="controls-col">
            <!-- Energy Performance Bias -->
            <div class="ctrl-card">
              <div class="ctrl-header">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
                </svg>
                <span class="ctrl-name">Energy Performance Bias</span>
                <span class="ctrl-addr">0x1B0</span>
              </div>
              <div class="ctrl-body">
                <div class="bias-display">
                  <span class="bias-hex">0x{{ energyBiasValue.toString(16).toUpperCase().padStart(2, '0') }}</span>
                  <span class="bias-val">{{ energyBiasValue }} / 15</span>
                </div>
                <div class="bias-bar">
                  <div class="bias-fill" :style="{ width: (energyBiasValue / 15 * 100) + '%' }"></div>
                </div>
                <div class="slider-labels">
                  <span>Performance</span>
                  <span>Balance</span>
                  <span>Power Save</span>
                </div>
                <input
                  type="range"
                  min="0"
                  max="15"
                  v-model="energyBiasValue"
                  @change="updateEnergyBias"
                  :disabled="!canWriteMSR"
                  class="range-input"
                />
              </div>
            </div>

            <!-- Turbo Boost -->
            <div class="ctrl-card">
              <div class="ctrl-header">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
                </svg>
                <span class="ctrl-name">Turbo Boost</span>
                <span class="ctrl-addr">0x1A0</span>
              </div>
              <div class="ctrl-body">
                <div class="toggle-row">
                  <span class="toggle-label">Enable Turbo Boost</span>
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="turboBoostEnabled" @change="updateTurboBoost" :disabled="!canWriteMSR"/>
                    <span class="toggle-slider"></span>
                  </label>
                </div>
                <p class="ctrl-desc">IA32_MISC_ENABLE MSR bit 38 controls Turbo Boost activation</p>
              </div>
            </div>

            <!-- P-State Control -->
            <div class="ctrl-card">
              <div class="ctrl-header">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <path d="M12 6v6l4 2"/>
                </svg>
                <span class="ctrl-name">P-State Control</span>
                <span class="ctrl-addr">0x198</span>
              </div>
              <div class="ctrl-body">
                <div class="pstate-row">
                  <div class="pstate-item">
                    <span class="pstate-label">Current P-State</span>
                    <span class="pstate-value mono">{{ currentPState }}</span>
                  </div>
                  <div class="pstate-item">
                    <span class="pstate-label">Target P-State</span>
                    <span class="pstate-value mono">{{ targetPState }}</span>
                  </div>
                </div>
                <button class="btn-action" @click="readPState" :disabled="!canReadMSR">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <polyline points="23 4 23 10 17 10"/>
                    <path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
                  </svg>
                  Read P-State
                </button>
              </div>
            </div>

            <!-- RAPL Power Limit -->
            <div class="ctrl-card">
              <div class="ctrl-header">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <path d="M18.36 6.64a9 9 0 1 1-12.73 0"/>
                  <line x1="12" y1="2" x2="12" y2="12"/>
                </svg>
                <span class="ctrl-name">RAPL Power Limit</span>
                <span class="ctrl-addr">0x610</span>
              </div>
              <div class="ctrl-body">
                <div class="rapl-fields">
                  <div class="rapl-field">
                    <label class="field-label">Package Power</label>
                    <div class="field-input-wrap">
                      <input type="number" v-model="packagePowerLimit" class="field-input" :disabled="!canWriteMSR"/>
                      <span class="field-unit">W</span>
                    </div>
                  </div>
                  <div class="rapl-field">
                    <label class="field-label">Time Window</label>
                    <div class="field-input-wrap">
                      <input type="number" v-model="timeWindow" class="field-input" :disabled="!canWriteMSR"/>
                      <span class="field-unit">ms</span>
                    </div>
                  </div>
                </div>
                <button class="btn-action primary" @click="applyPowerLimit" :disabled="!canWriteMSR">Apply Power Limit</button>
              </div>
            </div>

            <!-- Clock Modulation -->
            <div class="ctrl-card">
              <div class="ctrl-header">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                  <circle cx="12" cy="12" r="10"/>
                  <line x1="12" y1="2" x2="12" y2="4"/>
                  <line x1="12" y1="20" x2="12" y2="22"/>
                  <line x1="4.93" y1="4.93" x2="6.34" y2="6.34"/>
                  <line x1="17.66" y1="17.66" x2="19.07" y2="19.07"/>
                  <line x1="2" y1="12" x2="4" y2="12"/>
                  <line x1="20" y1="12" x2="22" y2="12"/>
                </svg>
                <span class="ctrl-name">Clock Modulation</span>
                <span class="ctrl-addr">0x19A</span>
              </div>
              <div class="ctrl-body">
                <div class="toggle-row">
                  <span class="toggle-label">Enable T-State Throttling</span>
                  <label class="toggle-switch">
                    <input type="checkbox" v-model="clockModEnabled" @change="updateClockMod" :disabled="!canWriteMSR"/>
                    <span class="toggle-slider"></span>
                  </label>
                </div>
                <div v-if="clockModEnabled" class="duty-section">
                  <div class="duty-label">
                    <span>Duty Cycle</span>
                    <span class="duty-pct">{{ dutyCycle }}%</span>
                  </div>
                  <div class="duty-bar">
                    <div class="duty-fill" :style="{ width: (dutyCycleValue / 14 * 100) + '%' }"></div>
                  </div>
                  <input
                    type="range"
                    min="1"
                    max="14"
                    v-model="dutyCycleValue"
                    @change="updateClockMod"
                    :disabled="!canWriteMSR"
                    class="range-input"
                  />
                </div>
                <p class="ctrl-desc">T-state throttling for thermal management via IA32_CLOCK_MODULATION</p>
              </div>
            </div>

            <!-- Uncore Ratio -->
            <div class="ctrl-card">
              <div class="ctrl-header">
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
                <span class="ctrl-name">Uncore Frequency Ratio</span>
                <span class="ctrl-addr">0x620</span>
              </div>
              <div class="ctrl-body">
                <div class="ratio-fields">
                  <div class="ratio-field">
                    <label class="field-label">Max Ratio</label>
                    <input type="number" v-model="uncoreMaxRatio" min="0" max="127" class="field-input" :disabled="!canWriteMSR"/>
                  </div>
                  <div class="ratio-sep">→</div>
                  <div class="ratio-field">
                    <label class="field-label">Min Ratio</label>
                    <input type="number" v-model="uncoreMinRatio" min="0" max="127" class="field-input" :disabled="!canWriteMSR"/>
                  </div>
                </div>
                <div v-if="uncoreMaxRatio === uncoreMinRatio" class="lock-warning">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                    <path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/>
                    <line x1="12" y1="9" x2="12" y2="13"/>
                    <line x1="12" y1="17" x2="12.01" y2="17"/>
                  </svg>
                  Frequency locked (max = min)
                </div>
                <button class="btn-action primary" @click="applyUncoreRatio" :disabled="!canWriteMSR">Apply Uncore Ratio</button>
              </div>
            </div>
          </div>

          <!-- Right Column: Register Reference Table -->
          <div class="regs-col">
            <div class="regs-header">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <line x1="8" y1="6" x2="21" y2="6"/>
                <line x1="8" y1="12" x2="21" y2="12"/>
                <line x1="8" y1="18" x2="21" y2="18"/>
                <line x1="3" y1="6" x2="3.01" y2="6"/>
                <line x1="3" y1="12" x2="3.01" y2="12"/>
                <line x1="3" y1="18" x2="3.01" y2="18"/>
              </svg>
              <span>MSR Register Reference</span>
            </div>
            <div class="regs-table-wrap">
              <table class="regs-table">
                <thead>
                  <tr>
                    <th>Address</th>
                    <th>Name</th>
                    <th>Access</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="reg in msrRegisters" :key="reg.address" :class="{ readonly: reg.readonly }">
                    <td class="mono">0x{{ reg.address.toString(16).toUpperCase().padStart(3, '0') }}</td>
                    <td class="reg-name" :title="reg.description">{{ reg.name }}</td>
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
    </div>

    <!-- Install Driver Modal -->
    <div class="modal-overlay" v-if="showInstallHelp" @click="showInstallHelp = false">
      <div class="modal-content" @click.stop>
        <div class="modal-header">
          <h3>Install MSR Driver</h3>
          <button class="modal-close" @click="showInstallHelp = false">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <line x1="18" y1="6" x2="6" y2="18"/>
              <line x1="6" y1="6" x2="18" y2="18"/>
            </svg>
          </button>
        </div>
        <div class="modal-body">
          <p>To access MSR registers, you need to install a kernel driver:</p>
          <h4>Option 1: WinRing0 (Recommended)</h4>
          <ol>
            <li>Download WinRing0 from GitHub</li>
            <li>Copy WinRing0x64.sys to <code>C:\Windows\System32\drivers\</code></li>
            <li>Run as Administrator:
              <code>sc create WinRing0 type= kernel binPath= C:\Windows\System32\drivers\WinRing0x64.sys</code>
            </li>
            <li>Start driver: <code>sc start WinRing0</code></li>
          </ol>
          <h4>Option 2: MSR Driver (Custom)</h4>
          <ol>
            <li>Build driver from <code>C:\LenovoDispatcher\MSR\driver\</code></li>
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
      winring0Available: false,
      msrDriverAvailable: false,
      isAdmin: false,
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
      showInstallHelp: false,
      advancedUnlocked: false,
      advancedPassword: '',
      passwordError: '',
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
      return Math.round(this.dutyCycleValue * 6.25)
    }
  },
  mounted() {
    this.checkDriverStatus()
    this.checkAdminRights()
  },
  methods: {
    async checkDriverStatus() {
      try {
        this.winring0Available = false
        this.msrDriverAvailable = false
      } catch (e) {
        console.error('Driver check failed:', e)
      }
    },
    checkAdminRights() {
      this.isAdmin = false
    },
    updateEnergyBias() {
      console.log('Setting Energy Perf Bias to:', this.energyBiasValue)
    },
    updateTurboBoost() {
      console.log('Setting Turbo Boost to:', this.turboBoostEnabled)
    },
    readPState() {
      this.currentPState = '0x1D00'
      this.targetPState = '0x1E00'
    },
    applyPowerLimit() {
      console.log('Applying power limit:', this.packagePowerLimit, 'W')
    },
    updateClockMod() {
      console.log('Clock modulation:', this.clockModEnabled, 'Duty:', this.dutyCycle)
    },
    applyUncoreRatio() {
      console.log('Uncore ratio - Max:', this.uncoreMaxRatio, 'Min:', this.uncoreMinRatio)
    },
    toggleAdvanced() {
      if (this.advancedUnlocked) {
        this.advancedUnlocked = false
      }
    },
    unlockAdvanced() {
      const now = new Date()
      const dateStr = `${now.getFullYear()}${String(now.getMonth() + 1).padStart(2, '0')}${String(now.getDate()).padStart(2, '0')}`
      const expected = `Lenovo${dateStr}`
      if (this.advancedPassword === expected) {
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
  padding: 20px 24px;
  max-width: 1400px;
  margin: 0 auto;
}

/* Header */
.msr-header {
  margin-bottom: 20px;
}
.msr-title {
  font-size: 13px;
  font-weight: 600;
  color: var(--lenovo-red);
  margin: 0 0 4px 0;
  text-transform: uppercase;
  letter-spacing: 0.8px;
}
.msr-subtitle {
  font-size: 12px;
  color: var(--text-secondary);
  margin: 0;
}

/* Driver Status Banner */
.msr-banner {
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 14px 20px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  margin-bottom: 16px;
}
.banner-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 40px;
  height: 40px;
  background: rgba(99, 102, 241, 0.1);
  border-radius: var(--radius-sm);
  color: #818CF8;
  flex-shrink: 0;
}
.banner-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
  flex-shrink: 0;
}
.banner-label {
  font-size: 11px;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.banner-value {
  font-size: 15px;
  font-weight: 700;
}
.banner-value.ok { color: #4ADE80; }
.banner-value.warning { color: #FBBF24; }
.banner-meta {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}
.meta-item {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}
.meta-label { color: var(--text-tertiary); }
.meta-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  flex-shrink: 0;
}
.meta-dot.green { background: #4ADE80; box-shadow: 0 0 6px rgba(74, 222, 128, 0.5); }
.meta-dot.red { background: #EF4444; box-shadow: 0 0 6px rgba(239, 68, 68, 0.5); }
.meta-dot.yellow { background: #FBBF24; box-shadow: 0 0 6px rgba(251, 191, 36, 0.5); }
.btn-install {
  padding: 8px 18px;
  background: var(--lenovo-red);
  color: white;
  border: none;
  border-radius: var(--radius-sm);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}
.btn-install:hover { background: var(--lenovo-red-light); }

/* Advanced Card */
.msr-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}
.advanced-card {
  border-color: rgba(245, 158, 11, 0.3);
}
.advanced-card.unlocked {
  border-color: rgba(74, 222, 128, 0.3);
}
.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-color);
}
.advanced-toggle.hoverable { cursor: pointer; }
.advanced-toggle.hoverable:hover { background: rgba(245, 158, 11, 0.06); }
.card-title {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}
.header-right {
  display: flex;
  align-items: center;
  gap: 10px;
}
.lock-hint {
  font-size: 11px;
  color: #F59E0B;
  background: rgba(245, 158, 11, 0.1);
  padding: 3px 10px;
  border-radius: 10px;
}
.unlocked-hint {
  font-size: 11px;
  color: #4ADE80;
  background: rgba(74, 222, 128, 0.1);
  padding: 3px 10px;
  border-radius: 10px;
}
.chevron-icon {
  transition: transform 0.25s ease;
  color: var(--text-tertiary);
}
.chevron-open { transform: rotate(180deg); }

/* Password Gate */
.password-gate {
  padding: 20px;
}
.password-row {
  display: flex;
  align-items: center;
  gap: 12px;
}
.password-icon {
  color: #F59E0B;
  flex-shrink: 0;
}
.password-input {
  flex: 1;
  padding: 10px 14px;
  border: 1px solid rgba(245, 158, 11, 0.4);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
}
.password-input:focus { border-color: #F59E0B; }
.password-input::placeholder { color: var(--text-tertiary); }
.btn-unlock {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.4);
  border-radius: var(--radius-md);
  color: #F59E0B;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
  white-space: nowrap;
}
.btn-unlock:hover { background: rgba(245, 158, 11, 0.2); border-color: #F59E0B; }
.password-error {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  padding: 10px 14px;
  background: rgba(239, 68, 68, 0.08);
  border: 1px solid rgba(239, 68, 68, 0.25);
  border-radius: var(--radius-md);
  font-size: 12px;
  color: #EF4444;
}

/* Advanced Content */
.advanced-content { padding: 20px; }

/* Controls Grid */
.controls-grid {
  display: grid;
  grid-template-columns: 1fr 380px;
  gap: 20px;
  align-items: start;
}
.controls-col {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
}

/* Control Cards */
.ctrl-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
}
.ctrl-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  color: var(--lenovo-red);
}
.ctrl-name {
  flex: 1;
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}
.ctrl-addr {
  font-family: 'Consolas', monospace;
  font-size: 10px;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
}
.ctrl-body { padding: 14px; }
.ctrl-desc {
  font-size: 11px;
  color: var(--text-tertiary);
  margin: 8px 0 0 0;
  line-height: 1.4;
}

/* Bias Display */
.bias-display {
  display: flex;
  align-items: baseline;
  gap: 10px;
  margin-bottom: 10px;
}
.bias-hex {
  font-family: 'Consolas', monospace;
  font-size: 22px;
  font-weight: 700;
  color: var(--lenovo-red);
}
.bias-val {
  font-size: 13px;
  color: var(--text-secondary);
}
.bias-bar {
  height: 4px;
  background: var(--bg-card);
  border-radius: 2px;
  margin-bottom: 6px;
  overflow: hidden;
}
.bias-fill {
  height: 100%;
  background: linear-gradient(90deg, #4ADE80, #F59E0B, #EF4444);
  border-radius: 2px;
  transition: width 0.3s;
}
.slider-labels {
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: var(--text-tertiary);
  margin-bottom: 8px;
}
.range-input {
  width: 100%;
  height: 5px;
  -webkit-appearance: none;
  background: var(--bg-card);
  border-radius: 3px;
  outline: none;
  cursor: pointer;
}
.range-input::-webkit-slider-thumb {
  -webkit-appearance: none;
  width: 16px;
  height: 16px;
  background: var(--lenovo-red);
  border-radius: 50%;
  cursor: pointer;
}
.range-input:disabled { opacity: 0.5; cursor: not-allowed; }

/* Toggle */
.toggle-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}
.toggle-label {
  font-size: 13px;
  color: var(--text-primary);
}
.toggle-switch {
  position: relative;
  width: 44px;
  height: 22px;
  flex-shrink: 0;
}
.toggle-switch input { opacity: 0; width: 0; height: 0; }
.toggle-slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background: var(--bg-card);
  border-radius: 22px;
  transition: 0.3s;
}
.toggle-slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 3px;
  bottom: 3px;
  background: white;
  border-radius: 50%;
  transition: 0.3s;
}
input:checked + .toggle-slider { background: var(--lenovo-red); }
input:checked + .toggle-slider:before { transform: translateX(22px); }
input:disabled + .toggle-slider { opacity: 0.4; cursor: not-allowed; }

/* P-State */
.pstate-row {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 12px;
}
.pstate-item {
  display: flex;
  flex-direction: column;
  gap: 3px;
}
.pstate-label {
  font-size: 11px;
  color: var(--text-tertiary);
}
.pstate-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}
.mono { font-family: 'Consolas', monospace; }

/* Buttons */
.btn-action {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  width: 100%;
  padding: 8px 14px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.2s;
}
.btn-action:hover:not(:disabled) { background: var(--bg-card-hover); }
.btn-action:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-action.primary {
  background: var(--lenovo-red);
  border-color: var(--lenovo-red);
  color: white;
}
.btn-action.primary:hover:not(:disabled) { background: var(--lenovo-red-light); }

/* RAPL Fields */
.rapl-fields {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-bottom: 12px;
}
.rapl-field { display: flex; flex-direction: column; gap: 4px; }
.field-label {
  font-size: 11px;
  color: var(--text-tertiary);
}
.field-input-wrap {
  display: flex;
  align-items: center;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  overflow: hidden;
}
.field-input {
  flex: 1;
  padding: 7px 10px;
  background: transparent;
  border: none;
  color: var(--text-primary);
  font-size: 14px;
  outline: none;
}
.field-input:focus { outline: none; }
.field-unit {
  padding: 0 10px 0 0;
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Duty Cycle */
.duty-section { margin-bottom: 10px; }
.duty-label {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
  font-size: 12px;
  color: var(--text-secondary);
}
.duty-pct {
  font-weight: 700;
  color: var(--lenovo-red);
  font-family: 'Consolas', monospace;
}
.duty-bar {
  height: 4px;
  background: var(--bg-card);
  border-radius: 2px;
  margin-bottom: 8px;
  overflow: hidden;
}
.duty-fill {
  height: 100%;
  background: var(--lenovo-red);
  border-radius: 2px;
  transition: width 0.3s;
}

/* Uncore Ratio */
.ratio-fields {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  margin-bottom: 12px;
}
.ratio-field { flex: 1; display: flex; flex-direction: column; gap: 4px; }
.ratio-sep {
  color: var(--text-tertiary);
  font-size: 16px;
  padding-bottom: 8px;
  flex-shrink: 0;
}
.field-input {
  padding: 7px 10px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  color: var(--text-primary);
  font-size: 14px;
  width: 100%;
  box-sizing: border-box;
}
.field-input:focus { outline: none; border-color: var(--lenovo-red); }
.field-input:disabled { opacity: 0.5; }
.lock-warning {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(251, 191, 36, 0.08);
  border: 1px solid rgba(251, 191, 36, 0.25);
  border-radius: var(--radius-sm);
  font-size: 12px;
  color: #FBBF24;
  margin-bottom: 12px;
}

/* Register Table Column */
.regs-col {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  overflow: hidden;
  position: sticky;
  top: 20px;
}
.regs-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  background: var(--bg-card);
  border-bottom: 1px solid var(--border-color);
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
}
.regs-table-wrap {
  overflow-y: auto;
  max-height: calc(100vh - 400px);
}
.regs-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.regs-table th {
  text-align: left;
  padding: 10px 12px;
  background: var(--bg-card);
  color: var(--text-tertiary);
  font-weight: 600;
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  position: sticky;
  top: 0;
  border-bottom: 1px solid var(--border-color);
}
.regs-table td {
  padding: 9px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.03);
  color: var(--text-primary);
  vertical-align: middle;
}
.regs-table tr:hover td { background: var(--bg-card); }
.regs-table tr.readonly td { opacity: 0.75; }
.reg-name {
  max-width: 150px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-family: 'Consolas', monospace;
  font-size: 11px;
}
.access-badge {
  display: inline-block;
  padding: 2px 7px;
  border-radius: 4px;
  font-size: 10px;
  font-weight: 700;
}
.access-badge.ro { background: rgba(96, 165, 250, 0.15); color: #60A5FA; }
.access-badge.rw { background: rgba(74, 222, 128, 0.15); color: #4ADE80; }

/* Modal */
.modal-overlay {
  position: fixed;
  top: 0; left: 0; right: 0; bottom: 0;
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
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}
.modal-close {
  background: none;
  border: none;
  color: var(--text-secondary);
  cursor: pointer;
  padding: 4px;
  display: flex;
  align-items: center;
}
.modal-close:hover { color: var(--text-primary); }
.modal-body {
  padding: 20px;
  color: var(--text-primary);
  line-height: 1.7;
  font-size: 13px;
}
.modal-body h4 {
  margin: 16px 0 8px 0;
  font-size: 13px;
  font-weight: 600;
  color: var(--lenovo-red);
}
.modal-body ol { margin: 8px 0; padding-left: 20px; }
.modal-body li { margin: 6px 0; }
.modal-body code {
  display: block;
  padding: 8px 12px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  font-family: 'Consolas', monospace;
  font-size: 12px;
  margin: 6px 0;
  word-break: break-all;
  color: #A5B4FC;
}
</style>
