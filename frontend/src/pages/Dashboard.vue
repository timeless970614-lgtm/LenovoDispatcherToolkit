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
          <button class="btn-icon" @click="refresh" title="Refresh">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/>
              <polyline points="1 20 1 14 7 14"/>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
            </svg>
          </button>
        </div>
        <div class="card-content" v-if="sysInfo">
          <div class="info-item">
            <span class="info-label">CPU</span>
            <span class="info-value">{{ sysInfo.CPUName || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">BIOS Version</span>
            <span class="info-value">{{ sysInfo.BIOSVersion || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Operating System</span>
            <span class="info-value">{{ sysInfo.OSCaption || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Code Name</span>
            <span class="info-value">{{ sysInfo.CodeName || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Total Memory</span>
            <span class="info-value">{{ sysInfo.TotalMemoryGB ? sysInfo.TotalMemoryGB.toFixed(2) + ' GB' : 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Memory Type</span>
            <span class="info-value">{{ sysInfo.MemoryType || 'LPDDR5 / DDR5' }}</span>
          </div>
        </div>
        <div class="loading" v-else><div class="spinner"></div></div>
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
          <button class="btn-enable-log" @click="enableLog" :disabled="enablingLog" :title="logEnabled ? 'Log already enabled' : 'Enable Dynamic Log'">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
            </svg>
            <span v-if="enablingLog">Enabling...</span>
            <span v-else-if="logEnabled">Log Enabled</span>
            <span v-else>EnableLog</span>
          </button>
        </div>
        <div class="card-content" v-if="deviceInfo">
          <div class="info-item">
            <span class="info-label">Model</span>
            <span class="info-value">{{ deviceInfo.model || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">Driver Version</span>
            <span class="info-value">{{ deviceInfo.driverVersion || 'N/A' }}</span>
          </div>
          <div class="info-item">
            <span class="info-label">AI Engine</span>
            <span class="info-value">{{ deviceInfo.aiEngineMode || 'N/A' }}</span>
          </div>
        </div>
        <div class="loading" v-else><div class="spinner"></div></div>
      </div>

      <!-- Service Control Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <circle cx="12" cy="12" r="3"/>
              <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
            </svg>
            Service Control
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
  data() {
    return {
      sysInfo: null,
      deviceInfo: null,
      serviceStatus: 'Unknown',
      refreshInterval: null,
      enablingLog: false,
      logEnabled: false,

    }
  },
  async mounted() {
    await this.refresh()
    this.refreshInterval = setInterval(this.refresh, 2000)
  },
  beforeUnmount() {
    if (this.refreshInterval) clearInterval(this.refreshInterval)
  },
  methods: {
    async refresh() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const [sysInfo, modeInfo, status, logStatus] = await Promise.all([
            window.go.main.App.GetSystemInfo(),
            window.go.main.App.GetModeCheckInfo(),
            window.go.main.App.GetServiceStatus(),
            window.go.main.App.GetDynamicLogStatus(),
          ])
          if (sysInfo) this.sysInfo = sysInfo
          if (modeInfo) this.deviceInfo = modeInfo
          if (status) this.serviceStatus = status
          if (logStatus !== undefined) this.logEnabled = logStatus
        }
      } catch (e) {
        console.error('Refresh error:', e)
      }
    },
    async startService() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.StartService()
          await this.refresh()
        }
      } catch (e) { console.error('Start service error:', e) }
    },
    async stopService() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.StopService()
          await this.refresh()
        }
      } catch (e) { console.error('Stop service error:', e) }
    },
    async restartService() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          await window.go.main.App.RestartService()
          await this.refresh()
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
    }
  }
}
</script>

<style scoped>
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

</style>
