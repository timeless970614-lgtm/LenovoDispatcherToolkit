<template>
  <div class="mode-check-page">
    <!-- Header -->
    <div class="page-header">
      <h1 class="page-title">Mode Check</h1>
      <p class="page-subtitle">Dispatcher status and feature analysis</p>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Loading mode check data...</span>
    </div>

    <!-- Main Content -->
    <div v-else-if="info" class="content-grid">
      <!-- Fixed Thermal Mode Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
              <circle cx="12" cy="10" r="3"/>
            </svg>
            Fixed Thermal Mode
          </span>
          <span v-if="pinnedMode" class="pin-badge pinned">
            📌 {{ pinnedMode }}
          </span>
          <span v-else class="pin-badge auto">
            Auto
          </span>
        </div>

        <div class="pin-description">
          <p>Fix a specific thermal mode to prevent automatic switching. The mode will be restored on service restart.</p>
        </div>

        <div v-for="group in modeGroups" :key="group.key" class="mode-group">
          <div class="mode-group-title">{{ group.name }}</div>
          <div class="mode-grid">
            <div 
              v-for="mode in availableModes.filter(m => m.group === group.key)" 
              :key="mode.id"
              :class="['mode-item', { 'mode-active': pinnedMode === mode.id, 'mode-current': currentMode === mode.id }]"
              @click="selectMode(mode.id)"
            >
              <div class="mode-id">{{ mode.id }}</div>
              <div class="mode-name">{{ mode.name }}</div>
            </div>
          </div>
        </div>

        <div class="pin-actions">
          <button class="btn btn-primary" @click="pinMode" :disabled="!selectedMode || pinning">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/>
              <circle cx="12" cy="10" r="3"/>
            </svg>
            {{ pinning ? 'Fixing...' : 'Fix Mode' }}
          </button>
          <button class="btn btn-secondary" @click="unpinMode" :disabled="!pinnedMode || pinning">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 6L6 18M6 6l12 12"/>
            </svg>
            Unfix (Auto)
          </button>
        </div>

        <div v-if="pinResult" :class="['pin-result', pinResult.success ? 'success' : 'error']">
          {{ pinResult.message }}
        </div>
      </div>

      <!-- DYTC Function Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <polyline points="16 18 22 12 16 6"/>
              <polyline points="8 6 2 12 8 18"/>
            </svg>
            DYTC Dispatcher Function
          </span>
        </div>

        <div class="dytc-display">
          <div class="dytc-value">
            <span class="dytc-label">DISPATCHER_FUNCTION</span>
            <span class="dytc-hex">{{ info.dytcValue ? '0x' + info.dytcValue.toString(16).toUpperCase().padStart(8, '0') : 'N/A' }}</span>
          </div>
          <div class="dytc-binary" v-if="info.dytcBinary">
            <span class="binary-label">Binary:</span>
            <span class="binary-value">{{ info.dytcBinary }}</span>
          </div>
        </div>
      </div>

      <!-- Policy Enable Function Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            Policy Enable Function
          </span>
          <span class="func-value">{{ info.enableFuncHex }}</span>
        </div>

        <div class="policy-list">
          <div 
            v-for="policy in info.enableFuncPolicies" 
            :key="policy.bit"
            :class="['policy-row', policy.enabled ? 'policy-enabled' : 'policy-disabled']"
          >
            <div class="policy-bit">
              <span class="bit-num">bit{{ policy.bit }}</span>
            </div>
            <div class="policy-indicator">
              <span v-if="policy.enabled" class="led led-on"></span>
              <span v-else class="led led-off"></span>
            </div>
            <div class="policy-info">
              <span class="policy-name">{{ policy.name }}</span>
              <span class="policy-desc">{{ policy.desc }}</span>
            </div>
            <div class="policy-status">
              <span :class="['status-tag', policy.enabled ? 'tag-on' : 'tag-off']">
                {{ policy.enabled ? 'ON' : 'OFF' }}
              </span>
            </div>
          </div>
        </div>

        <div class="policy-summary">
          <span class="summary-item">
            <span class="led led-on"></span>
            {{ info.enableFuncPolicies.filter(p => p.enabled).length }} policies enabled
          </span>
          <span class="summary-item">
            <span class="led led-off"></span>
            {{ info.enableFuncPolicies.filter(p => !p.enabled).length }} policies disabled
          </span>
        </div>
      </div>

      <!-- Features Card -->
      <div class="card" v-if="info.features && info.features.length">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
            </svg>
            Supported Features
          </span>
        </div>

        <div class="features-list">
          <div v-for="feature in info.features" :key="feature.name" class="feature-row">
            <div class="feature-name">{{ feature.name }}</div>
            <div class="feature-support" :class="feature.value === 'Y' ? 'support-yes' : 'support-na'">
              {{ feature.value === 'Y' ? 'Supported' : 'N/A' }}
            </div>
            <div class="feature-desc">{{ feature.support }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else class="error-state">
      <p>Failed to load mode check data</p>
      <button class="btn btn-primary" @click="refresh">Retry</button>
    </div>
  </div>
</template>

<script>
import { GetModeCheckInfo, GetServiceStatus, GetPinnedDYTCMode, PinDYTCMode, UnpinDYTCMode, GetDispatcherInfo, StartService, StopService } from '../../wailsjs/go/main/App'

export default {
  name: 'ModeCheck',
  data() {
    return {
      info: null,
      loading: true,
      serviceStatus: 'Unknown',
      serviceRunning: false,
      serviceInterval: null,
      pinnedMode: '',
      currentMode: 'N/A',
      selectedMode: '',
      pinning: false,
      pinResult: null,
      availableModes: [
        { id: 'BSM', name: 'Battery Saving', group: 'basic' },
        { id: 'EPM', name: 'Extreme Performance', group: 'basic' },
        { id: 'DCC', name: 'DCC Mode', group: 'intelligent' },
        { id: 'IBSM', name: 'Intelligent Battery Saving', group: 'intelligent' },
        { id: 'AQM', name: 'Intelligent Auto Quiet', group: 'intelligent' },
        { id: 'STD', name: 'Intelligent Stand Mode', group: 'intelligent' },
        { id: 'APM', name: 'Intelligent Auto Performance', group: 'intelligent' },
        { id: 'IEPM', name: 'Intelligent Extreme', group: 'intelligent' },
      ],
      modeGroups: [
        { key: 'basic', name: 'Standard Modes' },
        { key: 'intelligent', name: 'Intelligent Mode' },
      ]
    }
  },
  async mounted() {
    await this.refresh()
    this.startServiceWatcher()
  },
  beforeUnmount() {
    this.stopServiceWatcher()
  },
  methods: {
    async refresh() {
      this.loading = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const [info, status, pinned, dispatcher] = await Promise.all([
            GetModeCheckInfo(),
            GetServiceStatus(),
            GetPinnedDYTCMode(),
            GetDispatcherInfo(),
          ])
          this.info = info
          this.serviceStatus = status
          this.pinnedMode = pinned || ''
          if (dispatcher && dispatcher.CurrentMode) {
            // Extract mode abbreviation - handle "Name (num)" format
            const modeMap = {
              'Battery Saving': 'BSM',
              'Intelligent Battery Saving': 'IBSM',
              'Intelligent Auto Quiet': 'AQM',
              'Intelligent Stand Mode': 'STD',
              'Intelligent Auto Performance': 'APM',
              'Intelligent Extreme': 'IEPM',
              'Extreme Performance': 'EPM',
              'Yoga Tablet': 'Tablet',
              'Yoga Tent': 'Tent',
              'Yoga Flat': 'Flat',
              'Geek Mode': 'GEEK',
            }
            const raw = dispatcher.CurrentMode
            // Try to match by name first
            let matched = Object.keys(modeMap).find(k => raw.startsWith(k))
            if (matched) {
              this.currentMode = modeMap[matched]
            } else {
              // If no match, extract the number from "Name (num)" format
              const numMatch = raw.match(/\((\d+)\)$/)
              if (numMatch) {
                const num = parseInt(numMatch[1])
                const numToMode = { 1: 'BSM', 2: 'IBSM', 3: 'AQM', 4: 'STD', 5: 'APM', 6: 'IEPM', 7: 'EPM', 8: 'Tablet', 9: 'Tent', 10: 'Flat', 11: 'GEEK' }
                this.currentMode = numToMode[num] || raw
              } else {
                this.currentMode = raw
              }
            }
          } else {
            // Even if no dispatcher info, keep showing 'N/A'
            this.currentMode = 'N/A'
          }
        }
      } catch (e) {
        console.error('Failed to load mode check info:', e)
      } finally {
        this.loading = false
      }
    },
    selectMode(modeId) {
      this.selectedMode = modeId
    },
    async pinMode() {
      if (!this.selectedMode) return
      this.pinning = true
      this.pinResult = null
      
      // Step 1: Stop service first
      try {
        const status = await GetServiceStatus()
        if (status.toLowerCase().includes('running')) {
          console.log('Service is running, stopping...')
          const stopResult = await StopService()
          console.log('Stop service result:', stopResult)
          // Wait for service to fully stop
          await new Promise(resolve => setTimeout(resolve, 1500))
        }
      } catch (e) {
        console.warn('Failed to stop service:', e)
      }
      
      // Step 2: Pin the mode
      try {
        await PinDYTCMode(this.selectedMode)
        this.pinnedMode = this.selectedMode
        this.pinResult = { success: true, message: `Mode pinned to ${this.selectedMode}. Service stopped for pinning.` }
      } catch (e) {
        this.pinResult = { success: false, message: 'Failed to pin mode: ' + e }
      } finally {
        this.pinning = false
      }
    },
    async unpinMode() {
      this.pinning = true
      this.pinResult = null
      
      // Step 1: Check service status and start if stopped
      try {
        const status = await GetServiceStatus()
        if (status.toLowerCase().includes('stopped') || status.toLowerCase().includes('not running')) {
          console.log('Service is stopped, starting...')
          const startResult = await StartService()
          console.log('Start service result:', startResult)
          // Wait a moment for service to start
          await new Promise(resolve => setTimeout(resolve, 1000))
        }
      } catch (e) {
        console.warn('Failed to check/start service:', e)
      }
      
      // Step 2: Unpin the mode
      try {
        await UnpinDYTCMode()
        this.pinnedMode = ''
        this.pinResult = { success: true, message: 'Pin removed. Mode will be auto-selected on next restart.' }
      } catch (e) {
        this.pinResult = { success: false, message: 'Failed to unpin mode: ' + e }
      } finally {
        this.pinning = false
      }
    },
    
    // Real-time service status and mode polling
    async pollServiceAndMode() {
      try {
        // Poll service status
        const status = await GetServiceStatus()
        this.serviceRunning = status.toLowerCase().includes('running')
        
        // Poll current mode from dispatcher info
        const info = await GetDispatcherInfo()
        if (info && info.CurrentMode) {
          // Parse and set current mode
          const modeMap = {
            'Battery Saving': 'BSM',
            'Intelligent Battery Saving': 'IBSM',
            'Intelligent Auto Quiet': 'AQM',
            'Intelligent Stand Mode': 'STD',
            'Intelligent Auto Performance': 'APM',
            'Intelligent Extreme': 'IEPM',
            'Extreme Performance': 'EPM',
            'Yoga Tablet': 'Tablet',
            'Yoga Tent': 'Tent',
            'Yoga Flat': 'Flat',
            'Geek Mode': 'GEEK',
          }
          const raw = info.CurrentMode
          let matched = Object.keys(modeMap).find(k => raw.startsWith(k))
          if (matched) {
            this.currentMode = modeMap[matched]
          } else {
            const numMatch = raw.match(/\((\d+)\)$/)
            if (numMatch) {
              const num = parseInt(numMatch[1])
              const numToMode = { 1: 'BSM', 2: 'IBSM', 3: 'AQM', 4: 'STD', 5: 'APM', 6: 'IEPM', 7: 'EPM', 8: 'Tablet', 9: 'Tent', 10: 'Flat', 11: 'GEEK' }
              this.currentMode = numToMode[num] || raw
            } else {
              this.currentMode = raw
            }
          }
        }
        // If no info, keep the existing currentMode value (don't change)
      } catch (e) {
        console.warn('Failed to poll service/mode:', e)
      }
    },
    
    startServiceWatcher() {
      // Poll every 1 second
      this.serviceInterval = setInterval(() => this.pollServiceAndMode(), 1000)
      // Initial poll
      this.pollServiceAndMode()
    },
    
    stopServiceWatcher() {
      if (this.serviceInterval) {
        clearInterval(this.serviceInterval)
        this.serviceInterval = null
      }
    }
  }
}
</script>

<style scoped>
.mode-check-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 16px;
}

.status-bar {
  display: flex;
  gap: 12px;
  margin-bottom: 24px;
  align-items: center;
}

.service-status-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-tertiary);
  border-radius: 20px;
  font-size: 13px;
}

.svc-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  animation: pulse 2s infinite;
}

.svc-dot-running {
  background: #4ade80;
  box-shadow: 0 0 6px #4ade80;
}

.svc-dot-stopped {
  background: #ef4444;
  box-shadow: 0 0 6px #ef4444;
  animation: none;
}

.svc-label {
  color: var(--text-primary);
  font-weight: 500;
}

.current-mode-pill {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: var(--bg-tertiary);
  border-radius: 20px;
  font-size: 13px;
}

.mode-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: #60a5fa;
}

.mode-label {
  color: var(--text-secondary);
}

.mode-value {
  color: var(--text-primary);
  font-weight: 600;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 4px 0;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px;
  color: var(--text-secondary);
}

.error-state {
  text-align: center;
  padding: 60px;
  color: var(--text-secondary);
}

/* Card Styles */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.info-value.mono {
  font-family: 'Consolas', monospace;
  font-size: 12px;
}

/* Status Grid */
.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.status-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.status-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.status-value.highlight {
  color: var(--lenovo-red);
  font-weight: 600;
}

.status-value.badge {
  background: rgba(230, 63, 50, 0.1);
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  color: var(--lenovo-red);
  display: inline-block;
}

.status-value.mono {
  font-family: 'Consolas', monospace;
  font-size: 12px;
}

/* Status Badge */
.status-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
}

.status-running {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.status-stopped {
  background: rgba(244, 67, 54, 0.15);
  color: #F44336;
}

/* DYTC Display */
.dytc-display {
  background: var(--bg-tertiary);
  border-radius: 8px;
  padding: 16px;
}

.dytc-value {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.dytc-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.dytc-hex {
  font-family: 'Consolas', monospace;
  font-size: 20px;
  font-weight: 700;
  color: var(--lenovo-red);
}

.dytc-binary {
  display: flex;
  align-items: center;
  gap: 8px;
}

.binary-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

.binary-value {
  font-family: 'Consolas', monospace;
  font-size: 11px;
  color: var(--text-secondary);
  letter-spacing: 0.5px;
}

/* Policy List */
.func-value {
  font-family: 'Consolas', monospace;
  font-size: 12px;
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.1);
  padding: 3px 10px;
  border-radius: 6px;
}

.policy-list {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  max-height: 400px;
  overflow-y: auto;
}

.policy-row {
  display: grid;
  grid-template-columns: 60px 28px 1fr 60px;
  gap: 12px;
  padding: 10px 14px;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.15s;
}

.policy-row:last-child {
  border-bottom: none;
}

.policy-row:hover {
  background: var(--bg-tertiary);
}

.policy-row.policy-enabled {
  background: rgba(76, 175, 80, 0.03);
}

.policy-row.policy-disabled {
  opacity: 0.6;
}

.bit-num {
  font-family: 'Consolas', monospace;
  font-size: 10px;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
}

.led {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.led-on {
  background: #4CAF50;
  box-shadow: 0 0 5px rgba(76, 175, 80, 0.6);
}

.led-off {
  background: var(--text-tertiary);
  opacity: 0.4;
}

.policy-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.policy-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'Consolas', monospace;
}

.policy-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}

.status-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  text-align: center;
}

.tag-on {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.tag-off {
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
}

.policy-summary {
  display: flex;
  gap: 20px;
  margin-top: 12px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Features List */
.features-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feature-row {
  display: grid;
  grid-template-columns: 1fr 100px 1fr;
  gap: 12px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  align-items: center;
}

.feature-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.feature-support {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 12px;
  text-align: center;
}

.support-yes {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.support-na {
  background: var(--bg-primary);
  color: var(--text-tertiary);
}

.feature-desc {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Button */
.btn-icon {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 7px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: var(--transition);
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-icon:hover:not(:disabled) {
  color: var(--lenovo-red);
  border-color: var(--lenovo-red);
}

.btn-icon:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Responsive */
@media (max-width: 900px) {
  .info-grid,
  .status-grid {
    grid-template-columns: 1fr;
  }
}

/* Pin Mode Styles */
.pin-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
}

.pin-badge.pinned {
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red);
}

.pin-badge.auto {
  background: rgba(100, 100, 100, 0.15);
  color: var(--text-secondary);
}

.pin-description {
  margin-bottom: 16px;
}

.pin-description p {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
}

.mode-group {
  margin-bottom: 8px;
}

.mode-group-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 8px;
  padding-left: 2px;
}

.mode-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
  margin-bottom: 16px;
}

.mode-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: 12px 8px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: var(--transition);
}

.mode-item:hover {
  border-color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.05);
}

.mode-item.mode-active {
  border-color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.1);
}

.mode-item.mode-current {
  box-shadow: 0 0 0 2px rgba(74, 222, 128, 0.3);
}

.mode-id {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.mode-name {
  font-size: 10px;
  color: var(--text-secondary);
  text-align: center;
  margin-top: 4px;
}

.pin-actions {
  display: flex;
  gap: 12px;
  margin-bottom: 12px;
}

.pin-result {
  padding: 12px;
  border-radius: var(--radius-md);
  font-size: 13px;
}

.pin-result.success {
  background: rgba(76, 175, 80, 0.1);
  color: #4CAF50;
  border: 1px solid rgba(76, 175, 80, 0.3);
}

.pin-result.error {
  background: rgba(244, 67, 54, 0.1);
  color: #F44336;
  border: 1px solid rgba(244, 67, 54, 0.3);
}

@media (max-width: 800px) {
  .mode-grid {
    grid-template-columns: repeat(2, 1fr);
  }
}
</style>
