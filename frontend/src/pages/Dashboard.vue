<template>
  <div class="dashboard-page">
    <div class="dashboard-grid">

      <!-- System Info Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
              <line x1="8" y1="21" x2="16" y2="21"/>
              <line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            System Information
          </span>
          <button class="btn-icon" @click="fullRefresh" title="Refresh">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/>
              <polyline points="1 20 1 14 7 14"/>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
            </svg>
          </button>
        </div>
        <div class="card-content">
          <div class="info-item">
            <span class="info-label">CPU</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:70%"></span><span v-else>{{ sysInfo?.CPUName || 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label">BIOS Version</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:55%"></span><span v-else>{{ sysInfo?.BIOSVersion || 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label">Operating System</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:80%"></span><span v-else>{{ sysInfo?.OSCaption || 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label">Code Name</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:45%"></span><span v-else>{{ sysInfo?.CodeName || 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label">Total Memory</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:30%"></span><span v-else>{{ sysInfo?.TotalMemoryGB ? sysInfo.TotalMemoryGB.toFixed(2) + ' GB' : 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label">Memory Type</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:40%"></span><span v-else>{{ sysInfo?.MemoryType || 'LPDDR5 / DDR5' }}</span></span>
          </div>
        </div>
      </div>

      <!-- Dispatcher Device Information Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <rect x="2" y="7" width="20" height="14" rx="2" ry="2"/>
              <path d="M16 21V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v16"/>
            </svg>
            Dispatcher Device Information
          </span>
          <div style="display: flex; gap: 8px;">
            <button class="btn-enable-log" @click="enableLog" :disabled="enablingLog" :title="logEnabled ? 'Dispatcher log already enabled' : 'Enable Dispatcher Log'">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
                <polyline points="14 2 14 8 20 8"/>
                <line x1="16" y1="13" x2="8" y2="13"/>
                <line x1="16" y1="17" x2="8" y2="17"/>
              </svg>
              <span v-if="enablingLog">Enabling...</span>
              <span v-else-if="logEnabled">Dispatcher Log Enabled</span>
              <span v-else>Enable Dispatcher Log</span>
            </button>
            <button class="btn-enable-dump" @click="enableDump" :disabled="enablingDump" :title="dumpEnabled ? 'Dispatcher dump already enabled' : 'Enable Dispatcher Dump'">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                <line x1="12" y1="22.08" x2="12" y2="12"/>
              </svg>
              <span v-if="enablingDump">Enabling...</span>
              <span v-else-if="dumpEnabled">Dispatcher Dump Enabled</span>
              <span v-else>Enable Dispatcher Dump</span>
            </button>
            <button class="btn-test-mode" @click="openTestMode" title="Open Test Mode CMD">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="4 17 10 11 4 5"/>
                <line x1="12" y1="19" x2="20" y2="19"/>
              </svg>
              <span>TestMode</span>
            </button>
          </div>
        </div>
        <div class="card-content">
          <div class="info-item">
            <span class="info-label no-transform">Model</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:60%"></span><span v-else>{{ deviceInfo?.model || 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label no-transform">Driver Version</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:50%"></span><span v-else>{{ deviceInfo?.driverVersion || 'N/A' }}</span></span>
          </div>
          <div class="info-item">
            <span class="info-label no-transform">AI Engine</span>
            <span class="info-value"><span v-if="loading" class="skeleton-text" style="width:35%"></span><span v-else>{{ deviceInfo?.aiEngineMode || 'N/A' }}</span></span>
          </div>
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
            Dispatcher Service Status
          </span>
          <span :class="['status-badge', serviceStatus === 'Running' ? 'status-running' : 'status-stopped']">
            {{ serviceStatus }}
          </span>
        </div>
        <div class="btn-group">
          <button class="btn btn-primary" @click="startService" :disabled="serviceStatus === 'Running'">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
            Start
          </button>
          <button class="btn btn-secondary" @click="stopService" :disabled="serviceStatus !== 'Running'">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="6" y="4" width="4" height="16"/><rect x="14" y="4" width="4" height="16"/></svg>
            Stop
          </button>
          <button class="btn btn-secondary" @click="restartService">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/></svg>
            Restart
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'Dashboard',
  props: {
    theme: { type: String, default: 'dark' },
    serviceRunning: { type: Boolean, default: null },
    pollInterval: { type: Number, default: 10000 }
  },
  data() {
    return {
      sysInfo: null,
      deviceInfo: null,
      _timer: null,
      loading: true,
      enablingLog: false,
      logEnabled: false,
      enablingDump: false,
      dumpEnabled: false,

    }
  },
  computed: {
    serviceStatus() {
      if (this.serviceRunning === null) return 'Unknown'
      return this.serviceRunning ? 'Running' : 'Stopped'
    }
  },
  watch: {
    pollInterval() {
      if (this._timer) clearInterval(this._timer)
      this._timer = setInterval(this.refresh, this.pollInterval)
    }
  },
  mounted() {
    // Listen for backend pre-fetched data (emitted from onDomReady) for instant render.
    // Falls back to fullRefresh() if the event already fired (e.g. slow component mount).
    this._backendReadyHandler = (data) => {
      if (data) {
        if (data.sysInfo) this.sysInfo = data.sysInfo
        if (data.modeInfo) this.deviceInfo = data.modeInfo
        if (data.logEnabled !== undefined) this.logEnabled = data.logEnabled
        if (data.dumpEnabled !== undefined) this.dumpEnabled = data.dumpEnabled
        this.loading = false
      }
    }
    // Register listener — Wails EventsOn is available after DOM ready
    if (window.runtime && window.runtime.EventsOn) {
      window.runtime.EventsOn('backend:ready', this._backendReadyHandler)
    }
    // Fallback: if backend:ready already fired before listener was attached,
    // do a normal fetch.
    if (!this.sysInfo) {
      this.fullRefresh()
    }
    this._timer = setInterval(this.refresh, this.pollInterval)
    // Pause Dashboard polling when page is hidden
    this._visHandler = () => {
      if (document.hidden) {
        if (this._timer) { clearInterval(this._timer); this._timer = null }
      } else {
        if (!this._timer) {
          this.refresh()
          this._timer = setInterval(this.refresh, this.pollInterval)
        }
      }
    }
    document.addEventListener('visibilitychange', this._visHandler)
  },
  beforeUnmount() {
    if (this._timer) clearInterval(this._timer)
    if (this._visHandler) document.removeEventListener('visibilitychange', this._visHandler)
    if (this._backendReadyHandler && window.runtime && window.runtime.EventsOff) {
      window.runtime.EventsOff('backend:ready')
    }
  },
  methods: {
    async refresh() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          // Light refresh: only service status + log/dump (fast registry reads, no WMI)
          const [logStatus, dumpStatus] = await Promise.all([
            window.go.main.App.GetDynamicLogStatus(),
            window.go.main.App.GetDynamicDumpStatus(),
          ])
          if (logStatus !== undefined) this.logEnabled = logStatus
          if (dumpStatus !== undefined) this.dumpEnabled = dumpStatus
        }
      } catch (e) {
        console.error('Refresh error:', e)
      }
    },
    async fullRefresh() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const [sysInfo, modeInfo, logStatus, dumpStatus] = await Promise.all([
            window.go.main.App.GetSystemInfo(),
            window.go.main.App.GetModeCheckInfo(),
            window.go.main.App.GetDynamicLogStatus(),
            window.go.main.App.GetDynamicDumpStatus(),
          ])
          if (sysInfo) this.sysInfo = sysInfo
          if (modeInfo) this.deviceInfo = modeInfo
          if (logStatus !== undefined) this.logEnabled = logStatus
          if (dumpStatus !== undefined) this.dumpEnabled = dumpStatus
        }
      } catch (e) {
        console.error('Full refresh error:', e)
      } finally {
        this.loading = false
      }
    },
    async startService() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.StartService()
          this.$emit('service-changed')
          await this.fullRefresh()
        }
      } catch (e) { console.error('Start service error:', e) }
    },
    async stopService() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.StopService()
          this.$emit('service-changed')
          await this.fullRefresh()
        }
      } catch (e) { console.error('Stop service error:', e) }
    },
    async restartService() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.RestartService()
          this.$emit('service-changed')
          await this.fullRefresh()
        }
      } catch (e) { console.error('Restart service error:', e) }
    },
    async enableLog() {
      this.enablingLog = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const result = await window.go.main.App.EnableDynamicLog()
          if (result.success) {
            this.logEnabled = true
            alert(result.message)
          } else {
            alert('Failed to enable log: ' + result.message)
          }
        }
      } catch (e) {
        alert('Error: ' + e)
      } finally {
        this.enablingLog = false
      }
    },
    async enableDump() {
      this.enablingDump = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const result = await window.go.main.App.EnableDynamicDump()
          if (result.success) {
            this.dumpEnabled = true
            alert(result.message)
          } else {
            alert('Failed to enable dump: ' + result.message)
          }
        }
      } catch (e) {
        alert('Error: ' + e)
      } finally {
        this.enablingDump = false
      }
    },
    async openTestMode() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const result = await window.go.main.App.OpenTestMode()
          if (!result.success) {
            alert('Failed to open test mode: ' + result.message)
          }
        }
      } catch (e) {
        alert('Error: ' + e)
      }
    }
  }
}
</script>

<style scoped>
.info-label.no-transform { text-transform: none; }
.dashboard-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
  align-items: start;
}

.dashboard-grid .card {
  margin-bottom: 0;
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

.info-value.mono {
  font-family: 'Courier New', monospace;
  font-size: 13px;
}

.btn-enable-log {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.btn-enable-log:hover:not(:disabled) {
  border-color: var(--accent-green);
  color: var(--accent-green);
  background: rgba(74, 222, 128, 0.08);
}

.btn-enable-log:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-enable-dump {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.btn-enable-dump:hover:not(:disabled) {
  border-color: var(--accent-blue);
  color: var(--accent-blue);
  background: rgba(96, 165, 250, 0.08);
}

.btn-enable-dump:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-test-mode {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.btn-test-mode:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  background: rgba(227, 6, 19, 0.08);
}

/* Skeleton loading animation */
.skeleton-text {
  display: inline-block;
  height: 14px;
  border-radius: 4px;
  background: linear-gradient(90deg, var(--bg-tertiary) 25%, rgba(255,255,255,0.06) 50%, var(--bg-tertiary) 75%);
  background-size: 200% 100%;
  animation: skeleton-shimmer 1.5s ease-in-out infinite;
  vertical-align: middle;
}

@keyframes skeleton-shimmer {
  0% { background-position: 200% 0; }
  100% { background-position: -200% 0; }
}

</style>
