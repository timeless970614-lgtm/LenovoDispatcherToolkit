<template>
  <div class="dytc-page">
    <!-- Current Status Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <path d="M14 4v10.54a4 4 0 1 1-4 0V4a2 2 0 0 1 4 0Z"/>
          </svg>
          Lenovo Dispatcher DYTC Setting
        </span>
        <button class="btn-icon" @click="refreshAll" title="Refresh">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="23 4 23 10 17 10"/>
            <polyline points="1 20 1 14 7 14"/>
            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
          </svg>
        </button>
      </div>

      <div class="current-mode-display" v-if="dytcInfo">
        <div class="mode-info-grid">
          <div class="mode-info-item">
            <span class="mode-info-label">DYTC Mode</span>
            <span class="mode-info-value highlight">{{ dytcInfo.currentMode || 'Unknown' }}</span>
          </div>
          <div class="mode-info-item">
            <span class="mode-info-label">Dispatcher Mode</span>
            <span class="mode-info-value">{{ dytcInfo.currentDispatcherMode || 'Unknown' }}</span>
          </div>
          <div class="mode-info-item">
            <span class="mode-info-label">AI Engine</span>
            <span class="mode-info-value">{{ dytcInfo.aiEngineMode || 'Unknown' }}</span>
          </div>
        </div>
      </div>
      <div class="loading" v-else>
        <div class="spinner"></div>
      </div>
    </div>

    <!-- Thermal Mode Selection - New 3-Section Layout -->
    <div class="card thermal-card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <circle cx="12" cy="12" r="10"/>
            <polyline points="12 6 12 12 16 14"/>
          </svg>
          Thermal Mode Selection
        </span>
        <div class="header-right-group">
          <div v-if="pinnedMode" class="pin-badge">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
            Pinned: {{ pinnedMode }}
            <button class="btn-unpin" @click="unpinMode" title="Remove pin">✕</button>
          </div>
          <span class="hint-text">Stop Dispatcher service before changing modes</span>
        </div>
      </div>

      <div class="thermal-layout">
        <!-- BSM Section -->
        <div class="thermal-col thermal-col-bsm">
          <div class="section-label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            BSM
          </div>
          <div 
            class="mode-tile mode-tile-bsm"
            :class="{ active: currentMode === 'BSM', disabled: !capabilities.BSM }"
            @click="setDYTCMode('BSM')"
          >
            <div class="tile-icon" style="background: #6B7280;">
              <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
              </svg>
            </div>
            <div class="tile-name">BSM</div>
            <div class="tile-desc">Basic Performance</div>
            <div class="tile-badge">Base</div>
            <div class="tile-check" v-if="currentMode === 'BSM'">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
            </div>
            <button class="tile-pin-btn" :class="{ pinned: pinnedMode === 'BSM' }" @click.stop="togglePin('BSM')" title="Pin this mode">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
            </button>
          </div>
        </div>

        <!-- Intelligent Section -->
        <div class="thermal-col thermal-col-intelligent">
          <div class="section-label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
            Intelligent
          </div>
          <div class="intelligent-grid">
            <div 
              v-for="mode in intelligentModes" 
              :key="mode.id"
              class="mode-tile mode-tile-intel"
              :class="{ active: currentMode === mode.id, disabled: !capabilities[mode.id] }"
              @click="setDYTCMode(mode.id)"
              :title="mode.desc"
            >
              <div class="tile-icon" :style="{ background: mode.color }">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
                </svg>
              </div>
              <div class="tile-name">{{ mode.id }}</div>
              <div class="tile-desc">{{ mode.shortDesc }}</div>
              <div class="tile-check" v-if="currentMode === mode.id">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
              <button class="tile-pin-btn" :class="{ pinned: pinnedMode === mode.id }" @click.stop="togglePin(mode.id)" title="Pin this mode">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
              </button>
            </div>
          </div>
        </div>

        <!-- EPM / Geek Section -->
        <div class="thermal-col thermal-col-epm">
          <div class="section-label">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
            </svg>
            EPM / Geek
          </div>
          <div class="epm-stack">
            <div 
              class="mode-tile mode-tile-epm"
              :class="{ active: currentMode === 'EPM', disabled: !capabilities.EPM }"
              @click="setDYTCMode('EPM')"
            >
              <div class="tile-icon" style="background: #DC2626;">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
                </svg>
              </div>
              <div class="tile-name">EPM</div>
              <div class="tile-desc">Extreme Performance</div>
              <div class="tile-badge epm">Max Power</div>
              <div class="tile-check" v-if="currentMode === 'EPM'">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
              <button class="tile-pin-btn" :class="{ pinned: pinnedMode === 'EPM' }" @click.stop="togglePin('EPM')" title="Pin this mode">
                <svg width="12" height="12" viewBox="0 0 24 24" fill="currentColor" stroke="none"><path d="M16 12V4h1V2H7v2h1v8l-2 2v2h5.2v6h1.6v-6H18v-2l-2-2z"/></svg>
              </button>
            </div>

            <div 
              class="mode-tile mode-tile-geek"
              :class="{ active: currentMode === 'GEEK', disabled: !capabilities.GEEK }"
              @click="setDYTCMode('GEEK')"
            >
              <div class="tile-icon" style="background: #E63F32;">
                <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" stroke-width="2">
                  <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
                </svg>
              </div>
              <div class="tile-name">GEEK</div>
              <div class="tile-desc">Geek Mode</div>
              <div class="tile-badge geek">Advanced</div>
              <div class="tile-check" v-if="currentMode === 'GEEK'">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                  <polyline points="20 6 9 17 4 12"/>
                </svg>
              </div>
            </div>
          </div>
        </div>
      </div>

      <!-- Active Mode Indicator Bar -->
      <div class="active-indicator" v-if="currentMode">
        <span class="active-indicator-dot"></span>
        <span>Current: <strong>{{ currentMode }}</strong></span>
      </div>
    </div>

    <!-- Capability Info Card -->
    <div class="card" v-if="dytcInfo">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline points="22 4 12 14.01 9 11.01"/>
          </svg>
          Device Capabilities
        </span>
      </div>
      <div class="capability-grid">
        <div class="capability-item">
          <span class="cap-label">DCC Support</span>
          <span :class="['cap-value', dytcInfo.dccCapability !== 0 ? 'supported' : 'unsupported']">
            {{ dytcInfo.dccCapability !== 0 ? 'Supported' : 'Not Supported' }}
          </span>
        </div>
        <div class="capability-item">
          <span class="cap-label">GEEK Mode Support</span>
          <span :class="['cap-value', dytcInfo.geekCapability !== 0 ? 'supported' : 'unsupported']">
            {{ dytcInfo.geekCapability !== 0 ? 'Supported' : 'Not Supported' }}
          </span>
        </div>
        <div class="capability-item full-width">
          <span class="cap-label">Dispatcher Function</span>
          <span class="cap-value mono">0x{{ dytcInfo.dispatcherFunction.toString(16).toUpperCase().padStart(8, '0') }}</span>
        </div>
        <div class="capability-item full-width">
          <span class="cap-label">Enable Function</span>
          <span class="cap-value mono">0x{{ dytcInfo.enableFunc.toString(16).toUpperCase().padStart(8, '0') }}</span>
        </div>
      </div>
    </div>

    <!-- Dispatcher Function Features Card -->
    <div class="card" v-if="dytcInfo && dytcInfo.dispatcherFeatures && dytcInfo.dispatcherFeatures.length">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <polyline points="16 18 22 12 16 6"/>
            <polyline points="8 6 2 12 8 18"/>
          </svg>
          Dispatcher Function Features
        </span>
        <span class="func-raw-val">
          DISPATCHER_FUNCTION = <span class="mono-red">0x{{ dytcInfo.dispatcherFunction.toString(16).toUpperCase().padStart(8, '0') }}</span>
        </span>
      </div>

      <!-- Group tabs -->
      <div class="feat-group-tabs">
        <button 
          v-for="g in featureGroups" :key="g"
          :class="['feat-group-tab', { active: activeFeatureGroup === g }]"
          @click="activeFeatureGroup = g"
        >
          <span class="feat-group-dot" :class="'dot-' + g.toLowerCase()"></span>
          {{ g }}
          <span class="feat-group-count">
            {{ dytcInfo.dispatcherFeatures.filter(f => f.group === g && f.enabled).length }}/{{ dytcInfo.dispatcherFeatures.filter(f => f.group === g).length }}
          </span>
        </button>
      </div>

      <!-- Feature list -->
      <div class="feat-list">
        <div 
          v-for="feat in filteredFeatures" 
          :key="feat.bit"
          :class="['feat-row', feat.enabled ? 'feat-on' : 'feat-off']"
        >
          <div class="feat-bit">
            <span class="bit-badge">bit{{ feat.bit }}</span>
          </div>
          <div class="feat-status">
            <span v-if="feat.enabled" class="feat-led led-on"></span>
            <span v-else class="feat-led led-off"></span>
          </div>
          <div class="feat-info">
            <span class="feat-name">{{ feat.name }}</span>
            <span class="feat-desc">{{ feat.desc }}</span>
          </div>
          <div class="feat-state">
            <span :class="['feat-tag', feat.enabled ? 'tag-enabled' : 'tag-disabled']">
              {{ feat.enabled ? 'ON' : 'OFF' }}
            </span>
          </div>
        </div>
      </div>

      <!-- Summary bar -->
      <div class="feat-summary">
        <span class="feat-summary-item">
          <span class="feat-led led-on" style="display:inline-block;"></span>
          {{ dytcInfo.dispatcherFeatures.filter(f => f.enabled).length }} features enabled
        </span>
        <span class="feat-summary-item">
          <span class="feat-led led-off" style="display:inline-block;"></span>
          {{ dytcInfo.dispatcherFeatures.filter(f => !f.enabled).length }} features disabled
        </span>
      </div>
    </div>

    <!-- Service Control Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
          Dispatcher Service Control
        </span>
        <span :class="['status-badge', serviceStatus === 'Running' ? 'status-running' : 'status-stopped']">
          {{ serviceStatus }}
        </span>
      </div>
      <div class="btn-group">
        <button class="btn btn-primary" @click="startDispatcher" :disabled="serviceStatus === 'Running'">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
          Start Dispatcher
        </button>
        <button class="btn btn-danger" @click="stopDispatcher" :disabled="serviceStatus !== 'Running'">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
          Stop Dispatcher
        </button>
      </div>
      <p class="service-hint">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="16" x2="12" y2="12"/><line x1="12" y1="8" x2="12.01" y2="8"/></svg>
        Stop Dispatcher service before changing DYTC modes for changes to take effect
      </p>
    </div>

    <!-- Status Message -->
    <transition name="fade">
      <div v-if="statusMsg" :class="['status-msg', statusOk ? 'success' : 'error']">
        <svg v-if="statusOk" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
          <polyline points="22 4 12 14.01 9 11.01"/>
        </svg>
        <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <line x1="12" y1="8" x2="12" y2="12"/>
          <line x1="12" y1="16" x2="12.01" y2="16"/>
        </svg>
        {{ statusMsg }}
      </div>
    </transition>

    <!-- Confirm: Stop Service & Set Mode -->
    <transition name="modal-fade">
      <div v-if="confirmVisible" class="modal-overlay" @click.self="cancelConfirm">
        <div class="modal-box">
          <div class="modal-icon">
            <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="8" x2="12" y2="12"/>
              <line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
          </div>
          <div class="modal-title">Dispatcher Service is Running</div>
          <div class="modal-body">
            To apply <strong>{{ confirmPendingMode }}</strong> mode, the Dispatcher service must be stopped first.<br><br>
            Click <strong>Stop &amp; Apply</strong> to stop the service and set the mode automatically.
          </div>
          <div class="modal-actions">
            <button class="modal-btn-cancel" @click="cancelConfirm" :disabled="confirmStopping">
              Cancel
            </button>
            <button class="modal-btn-confirm" @click="confirmStopAndSet" :disabled="confirmStopping">
              <svg v-if="confirmStopping" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
              {{ confirmStopping ? `Waiting for service to stop... (${serviceStatus})` : 'Stop & Apply' }}
            </button>
          </div>
        </div>
      </div>
    </transition>
  </div>
</template>

<script>
export default {
  name: 'DYTCMode',
  data() {
    return {
      currentMode: '',
      serviceStatus: 'Unknown',
      dytcInfo: null,
      capabilities: {
        BSM: true, IBSM: true, AQM: true, STD: true,
        APM: true, IEPM: true, EPM: true, DCC: true, GEEK: true
      },
      intelligentModes: [
        { id: 'IBSM', shortDesc: 'Intelligent Basic', color: '#3B82F6' },
        { id: 'AQM',  shortDesc: 'Active Quiet',      color: '#F59E0B' },
        { id: 'STD',  shortDesc: 'Standard',           color: '#10B981' },
        { id: 'APM',  shortDesc: 'Advanced Perf',      color: '#EF4444' },
        { id: 'DCC',  shortDesc: 'Dynamic Ctrl',       color: '#0EA5E9' },
        { id: 'IEPM', shortDesc: 'Intell. Perf',       color: '#8B5CF6' }
      ],
      featureGroups: ['Core', 'Perf', 'Turbo', 'AI', 'GPU', 'Memory', 'SSD', 'Audio', 'UI', 'Debug'],
      activeFeatureGroup: 'Core',
      statusMsg: '',
      statusOk: true,
      refreshInterval: null,
      pinnedMode: '',
      // confirm dialog state
      confirmVisible: false,
      confirmPendingMode: null,
      confirmStopping: false,
    }
  },
  async mounted() {
    await this.refreshAll()
    await this.loadPinnedMode()
    this.refreshInterval = setInterval(this.refreshStatus, 3000)
  },
  computed: {
    filteredFeatures() {
      if (!this.dytcInfo || !this.dytcInfo.dispatcherFeatures) return []
      return this.dytcInfo.dispatcherFeatures.filter(f => f.group === this.activeFeatureGroup)
    }
  },
  beforeUnmount() {
    if (this.refreshInterval) {
      clearInterval(this.refreshInterval)
    }
  },
  methods: {
    async refreshAll() {
      await this.refreshDYTCInfo()
      await this.refreshCapabilities()
      await this.refreshStatus()
    },
    async refreshDYTCInfo() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const info = await window.go.main.App.GetDYTCInfo()
          if (info) {
            this.dytcInfo = info
            this.currentMode = info.currentMode || ''
          }
        }
      } catch (e) {
        console.error('Refresh DYTC info error:', e)
      }
    },
    async refreshCapabilities() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const caps = await window.go.main.App.CheckDYTCCapabilities()
          if (caps) {
            this.capabilities = caps
          }
        }
      } catch (e) {
        console.error('Refresh capabilities error:', e)
      }
    },
    async refreshStatus() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const status = await window.go.main.App.GetServiceStatus()
          if (status) this.serviceStatus = status
        }
      } catch (e) {
        console.error('Refresh status error:', e)
      }
    },
    async setDYTCMode(modeId) {
      if (!this.capabilities[modeId]) {
        this.showStatus(`${modeId} mode is not supported on this device`, false)
        return
      }
      // If service is running, ask user to stop it first
      if (this.serviceStatus === 'Running') {
        this.confirmPendingMode = modeId
        this.confirmVisible = true
        return
      }
      // Service already stopped — set mode directly
      await this._applyDYTCMode(modeId)
    },

    async _applyDYTCMode(modeId) {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const result = await window.go.main.App.SetDYTCMode(modeId)
          this.showStatus(result, true)
          await this.refreshDYTCInfo()
        }
      } catch (e) {
        this.showStatus(`Failed to set ${modeId}: ${e}`, false)
      }
    },

    async confirmStopAndSet() {
      this.confirmStopping = true
      const pendingMode = this.confirmPendingMode
      try {
        if (window.go && window.go.main && window.go.main.App) {
          // Send stop command (non-blocking — don't await the full 30s wait)
          window.go.main.App.StopService().catch(() => {})

          // Poll service status until stopped (max 30s)
          let stopped = false
          for (let i = 0; i < 60; i++) {
            await new Promise(r => setTimeout(r, 500))
            try {
              const st = await window.go.main.App.GetServiceStatus()
              if (st === 'Stopped') { stopped = true; break }
            } catch (_) {}
          }

          if (!stopped) {
            this.showStatus('Service did not stop in time, please try manually', false)
            return
          }

          await this.refreshStatus()
          // Apply the mode
          const result = await window.go.main.App.SetDYTCMode(pendingMode)
          this.showStatus(result || `${pendingMode} mode applied`, true)
          await this.refreshDYTCInfo()
        }
      } catch (e) {
        this.showStatus(`Operation failed: ${e}`, false)
      } finally {
        this.confirmStopping = false
        this.confirmVisible = false
        this.confirmPendingMode = null
      }
    },

    cancelConfirm() {
      this.confirmVisible = false
      this.confirmPendingMode = null
    },

    async loadPinnedMode() {
      try {
        if (window.go?.main?.App) {
          this.pinnedMode = await window.go.main.App.GetPinnedDYTCMode() || ''
        }
      } catch (e) { /* silent */ }
    },

    async togglePin(modeId) {
      try {
        if (!window.go?.main?.App) return
        if (this.pinnedMode === modeId) {
          // Already pinned → unpin
          await window.go.main.App.UnpinDYTCMode()
          this.pinnedMode = ''
          this.showStatus(`${modeId} unpinned`, true)
        } else {
          // Pin this mode
          await window.go.main.App.PinDYTCMode(modeId)
          this.pinnedMode = modeId
          this.showStatus(`${modeId} pinned — Dispatcher will restore this mode on restart`, true)
        }
      } catch (e) {
        this.showStatus(`Pin failed: ${e}`, false)
      }
    },

    async unpinMode() {
      try {
        if (window.go?.main?.App) {
          await window.go.main.App.UnpinDYTCMode()
          this.pinnedMode = ''
          this.showStatus('Mode unpinned', true)
        }
      } catch (e) {
        this.showStatus(`Unpin failed: ${e}`, false)
      }
    },
    async startDispatcher() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.StartService()
          this.showStatus('Dispatcher service started', true)
          await this.refreshStatus()
        }
      } catch (e) {
        this.showStatus(e, false)
      }
    },
    async stopDispatcher() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.StopService()
          this.showStatus('Dispatcher service stopped', true)
          await this.refreshStatus()
        }
      } catch (e) {
        this.showStatus(e, false)
      }
    },
    showStatus(msg, ok = true) {
      this.statusMsg = msg
      this.statusOk = ok
      setTimeout(() => { this.statusMsg = '' }, 4000)
    }
  }
}
</script>

<style scoped>
.current-mode-display {
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 20px;
  margin-bottom: 20px;
}

.mode-info-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 20px;
}

.mode-info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.mode-info-label {
  font-size: 12px;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.mode-info-value {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.mode-info-value.highlight {
  color: var(--lenovo-red);
}

.hint-text {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Thermal Layout - 3 columns */
.thermal-card {
  padding: 20px;
}

.thermal-layout {
  display: grid;
  grid-template-columns: 1fr 2fr 1fr;
  gap: 20px;
  margin-top: 16px;
  align-items: start;
}

.thermal-col {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.section-label {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-tertiary);
  padding-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
}

/* BSM Column */
.thermal-col-bsm .mode-tile-bsm {
  flex-direction: column;
  text-align: center;
  padding: 16px 12px;
  gap: 6px;
}

.thermal-col-bsm .tile-icon {
  width: 36px;
  height: 36px;
  min-width: 36px;
  border-radius: 8px;
}

/* Intelligent Grid */
.intelligent-grid {
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
}

.thermal-col-intelligent .section-label {
  color: #3B82F6;
  border-color: rgba(59, 130, 246, 0.3);
}

/* EPM Stack */
.epm-stack {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

/* Mode Tiles */
.mode-tile {
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 14px;
  cursor: pointer;
  transition: var(--transition);
  position: relative;
  display: flex;
  align-items: center;
  gap: 12px;
}

.mode-tile:hover:not(.disabled) {
  border-color: var(--border-light);
  transform: translateY(-1px);
}

.mode-tile.active {
  border-color: var(--lenovo-red);
  background: linear-gradient(135deg, rgba(230, 63, 50, 0.1) 0%, rgba(230, 63, 50, 0.04) 100%);
}

.mode-tile.disabled {
  opacity: 0.35;
  cursor: not-allowed;
}

.tile-icon {
  width: 40px;
  height: 40px;
  min-width: 40px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

.mode-tile.active .tile-icon {
  box-shadow: 0 3px 12px rgba(230, 63, 50, 0.35);
}

.mode-tile-intel {
  flex-direction: column;
  text-align: center;
  padding: 14px 10px;
  gap: 6px;
}

.mode-tile-intel .tile-icon {
  width: 36px;
  height: 36px;
  min-width: 36px;
  border-radius: 8px;
}

.tile-name {
  font-size: 13px;
  font-weight: 700;
  color: var(--text-primary);
}

.tile-desc {
  font-size: 10px;
  color: var(--text-tertiary);
  line-height: 1.3;
}

.tile-badge {
  font-size: 9px;
  font-weight: 700;
  text-transform: uppercase;
  padding: 2px 6px;
  border-radius: 4px;
  background: rgba(107, 114, 128, 0.2);
  color: #9CA3AF;
  letter-spacing: 0.5px;
}

.tile-badge.epm {
  background: rgba(220, 38, 38, 0.15);
  color: #DC2626;
}

.tile-badge.geek {
  background: rgba(230, 63, 50, 0.15);
  color: #E63F32;
}

.tile-check {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 20px;
  height: 20px;
  background: var(--lenovo-red);
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: white;
}

/* Active Indicator */
.active-indicator {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 10px 14px;
  background: rgba(230, 63, 50, 0.08);
  border: 1px solid rgba(230, 63, 50, 0.2);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-secondary);
}

.active-indicator strong {
  color: var(--lenovo-red);
}

.active-indicator-dot {
  width: 8px;
  height: 8px;
  background: var(--lenovo-red);
  border-radius: 50%;
  animation: pulse 2s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; transform: scale(1); }
  50% { opacity: 0.5; transform: scale(0.85); }
}

/* Capability Grid */
.capability-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-top: 16px;
}

.capability-item {
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  padding: 14px 16px;
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.capability-item.full-width {
  grid-column: span 2;
}

.cap-label {
  font-size: 13px;
  color: var(--text-secondary);
}

.cap-value {
  font-size: 13px;
  font-weight: 600;
}

.cap-value.supported {
  color: var(--accent-green);
}

.cap-value.unsupported {
  color: #F44336;
}

.cap-value.mono {
  font-family: 'Consolas', 'Monaco', monospace;
  color: var(--text-primary);
}

.btn-danger {
  background: linear-gradient(135deg, #DC2626 0%, #B91C1C 100%);
  color: white;
  padding: 12px 24px;
  border: none;
  border-radius: var(--radius-md);
  cursor: pointer;
  font-size: 14px;
  font-weight: 600;
  display: inline-flex;
  align-items: center;
  gap: 8px;
  transition: var(--transition);
  font-family: inherit;
}

.btn-danger:hover:not(:disabled) {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(220, 38, 38, 0.3);
}

.btn-danger:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.service-hint {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 16px;
  padding: 12px 16px;
  background: rgba(251, 191, 36, 0.1);
  border: 1px solid rgba(251, 191, 36, 0.3);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--accent-yellow);
}

.btn-icon {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 8px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: var(--transition);
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-icon:hover {
  background: var(--bg-card-hover);
  color: var(--lenovo-red);
  border-color: var(--lenovo-red);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Dispatcher Function Features */
.func-raw-val {
  font-size: 12px;
  color: var(--text-tertiary);
}

.mono-red {
  font-family: 'Consolas', monospace;
  color: var(--lenovo-red);
  font-weight: 700;
}

.feat-group-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 16px;
}

.feat-group-tab {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  font-size: 12px;
  color: var(--text-secondary);
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.feat-group-tab:hover {
  border-color: var(--border-light);
  color: var(--text-primary);
}

.feat-group-tab.active {
  background: rgba(230, 63, 50, 0.1);
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  font-weight: 600;
}

.feat-group-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: currentColor;
}

.feat-group-count {
  font-size: 10px;
  opacity: 0.7;
  background: var(--bg-primary);
  padding: 1px 5px;
  border-radius: 8px;
}

.feat-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.feat-row {
  display: grid;
  grid-template-columns: 60px 28px 1fr 60px;
  gap: 12px;
  padding: 10px 14px;
  align-items: center;
  border-top: 1px solid var(--border-color);
  transition: background 0.15s;
}

.feat-row:first-child {
  border-top: none;
}

.feat-row:hover {
  background: var(--bg-tertiary);
}

.feat-row.feat-on {
  background: rgba(76, 175, 80, 0.03);
}

.feat-row.feat-off {
  opacity: 0.6;
}

.bit-badge {
  font-family: 'Consolas', monospace;
  font-size: 10px;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
  border: 1px solid var(--border-color);
}

.feat-led {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  flex-shrink: 0;
}

.led-on {
  background: #4CAF50;
  box-shadow: 0 0 5px rgba(76, 175, 80, 0.6);
}

.led-off {
  background: var(--text-tertiary);
  opacity: 0.4;
}

.feat-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.feat-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'Consolas', monospace;
}

.feat-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}

.feat-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  text-align: center;
}

.tag-enabled {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.tag-disabled {
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
}

.feat-summary {
  display: flex;
  gap: 20px;
  margin-top: 12px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.feat-summary-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

@media (max-width: 900px) {
  .thermal-layout {
    grid-template-columns: 1fr 1fr;
  }
  .thermal-col-intelligent {
    grid-column: span 2;
  }
  .intelligent-grid {
    grid-template-columns: repeat(3, 1fr);
  }
  .mode-info-grid {
    grid-template-columns: 1fr;
  }
  .capability-grid {
    grid-template-columns: 1fr;
  }
  .capability-item.full-width {
    grid-column: span 1;
  }
}

/* ── Pin Mode ──────────────────────────────────────────────── */
.header-right-group {
  display: flex;
  align-items: center;
  gap: 12px;
}

.pin-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 10px;
  background: rgba(230, 63, 50, 0.1);
  border: 1px solid rgba(230, 63, 50, 0.35);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: var(--lenovo-red);
}

.btn-unpin {
  background: none;
  border: none;
  color: var(--lenovo-red);
  cursor: pointer;
  font-size: 12px;
  padding: 0 2px;
  line-height: 1;
  opacity: 0.7;
  transition: opacity 0.15s;
}
.btn-unpin:hover { opacity: 1; }

.tile-pin-btn {
  position: absolute;
  top: 6px;
  right: 6px;
  width: 22px;
  height: 22px;
  border-radius: 50%;
  border: 1px solid rgba(255,255,255,0.2);
  background: rgba(0,0,0,0.25);
  color: rgba(255,255,255,0.5);
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: all 0.15s;
  z-index: 2;
}

.tile-pin-btn:hover {
  background: rgba(230, 63, 50, 0.6);
  color: white;
  border-color: rgba(230, 63, 50, 0.8);
}

.tile-pin-btn.pinned {
  background: var(--lenovo-red);
  color: white;
  border-color: var(--lenovo-red);
  box-shadow: 0 0 8px rgba(230, 63, 50, 0.6);
}

/* ── Confirm Modal ─────────────────────────────────────────── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.modal-box {
  background: var(--bg-card);
  border: 1px solid var(--border-light);
  border-radius: 16px;
  padding: 32px 28px 24px;
  width: 380px;
  max-width: 90vw;
  box-shadow: 0 24px 64px rgba(0, 0, 0, 0.5);
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
  text-align: center;
}

.modal-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: rgba(230, 63, 50, 0.12);
  border: 1px solid rgba(230, 63, 50, 0.3);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--lenovo-red);
  margin-bottom: 4px;
}

.modal-title {
  font-size: 17px;
  font-weight: 700;
  color: var(--text-primary);
}

.modal-body {
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.7;
}

.modal-body strong {
  color: var(--text-primary);
}

.modal-actions {
  display: flex;
  gap: 10px;
  margin-top: 8px;
  width: 100%;
}

.modal-btn-cancel {
  flex: 1;
  padding: 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  transition: var(--transition);
}

.modal-btn-cancel:hover:not(:disabled) {
  background: var(--bg-card-hover);
  color: var(--text-primary);
}

.modal-btn-confirm {
  flex: 1;
  padding: 10px;
  background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border: none;
  border-radius: 8px;
  color: white;
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  font-family: inherit;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  transition: var(--transition);
}

.modal-btn-confirm:hover:not(:disabled) {
  opacity: 0.9;
}

.modal-btn-cancel:disabled,
.modal-btn-confirm:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.modal-fade-enter-active,
.modal-fade-leave-active {
  transition: opacity 0.2s ease;
}
.modal-fade-enter-from,
.modal-fade-leave-to {
  opacity: 0;
}
.modal-fade-enter-active .modal-box,
.modal-fade-leave-active .modal-box {
  transition: transform 0.2s ease;
}
.modal-fade-enter-from .modal-box {
  transform: scale(0.92) translateY(10px);
}

@keyframes spin {
  to { transform: rotate(360deg); }
}
.spinning {
  animation: spin 0.8s linear infinite;
}
</style>
