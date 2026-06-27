<template>
  <div class="log-section">
    <!-- Log Files Card - Password Protected -->
    <div class="card advanced-section">
      <div class="card-header advanced-toggle" @click="toggleLogSection">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
          </svg>
          Log Files
        </span>
        <svg :class="['chevron', { open: logSectionExpanded }]" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
      </div>

      <!-- Password Prompt -->
      <div v-if="logSectionExpanded && !logUnlocked" class="password-prompt">
        <div class="password-row">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px; flex-shrink:0;">
            <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
            <path d="M7 11V7a5 5 0 0110 0v4"/>
          </svg>
          <input
            type="password"
            class="password-input"
            v-model="logPassword"
            placeholder="Enter password to unlock"
            @keydown.enter="unlockLogs"
            ref="logPasswordInput"
          />
          <button class="btn-unlock" @click="unlockLogs">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/>
            </svg>
            Unlock
          </button>
        </div>
        <div v-if="logPasswordError" class="password-error">{{ logPasswordError }}</div>
      </div>

      <!-- Unlocked Content -->
      <div v-if="logSectionExpanded && logUnlocked">
        <div style="display:flex;justify-content:flex-end;padding:0 0 8px 0;">
          <button class="btn-sm" @click="loadLogs">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
            </svg>
            Refresh
          </button>
        </div>
        <div v-if="logFiles.length" class="log-file-list">
          <div v-for="f in logFiles" :key="f.name" class="log-file-item"
            :class="{ active: selectedLog === f.name }" @click="selectLog(f.name)">
            <span class="log-name">{{ f.name }}</span>
            <span class="log-meta">{{ formatSize(f.size) }}  {{ f.modTime }}</span>
          </div>
        </div>
        <div v-else class="empty-hint">No log files found in {{ logDir }}</div>
      </div>
    </div>

    <!-- AI Analysis Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/>
          </svg>
          AI Analysis
        </span>
        <button class="btn-analyze" @click="runAnalysis" :disabled="analyzing">
          <svg v-if="!analyzing" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          {{ analyzing ? 'Analyzing...' : 'Analyze Latest Log' }}
        </button>
      </div>
      <div v-if="summary" class="summary-grid">
        <div class="summary-panel">
          <div class="panel-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
            Workload Levels
          </div>
          <div class="log-lines">
            <div v-for="(l,i) in summary.workloadLevels" :key="i" class="log-line">{{ l }}</div>
            <div v-if="!summary.workloadLevels?.length" class="no-data">No data</div>
          </div>
        </div>
        <div class="summary-panel">
          <div class="panel-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
            </svg>
            Turbo Events
          </div>
          <div class="log-lines">
            <div v-for="(l,i) in summary.turboEvents" :key="i" class="log-line">{{ l }}</div>
            <div v-if="!summary.turboEvents?.length" class="no-data">No data</div>
          </div>
        </div>
        <div class="summary-panel">
          <div class="panel-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            Foreground Apps
          </div>
          <div class="log-lines">
            <div v-for="(l,i) in summary.appEvents" :key="i" class="log-line">{{ l }}</div>
            <div v-if="!summary.appEvents?.length" class="no-data">No data</div>
          </div>
        </div>
        <div class="summary-panel panel-errors">
          <div class="panel-title error-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            Errors / Warnings
          </div>
          <div class="log-lines">
            <div v-for="(l,i) in summary.errors" :key="i" class="log-line error-line">{{ l }}</div>
            <div v-if="!summary.errors?.length" class="no-data ok-hint">OK - No errors found</div>
          </div>
        </div>
      </div>
      <div v-else-if="!analyzing" class="empty-hint">Click "Analyze Latest Log" to start</div>
      <div v-else class="analyzing-hint">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
          <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
        </svg>
        Reading and parsing log...
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'LogAnalysis',
  props: {
    theme: { type: String, default: 'dark' },
  },
  data() {
    return {
      logDir: 'C:\\ProgramData\\Lenovo\\LenovoDispatcher\\Logs',
      logFiles: [],
      logSectionExpanded: true,
      logUnlocked: false,
      logPassword: '',
      logPasswordError: '',
      selectedLog: null,
      summary: null,
      rawLog: '',
      tailLines: 200,
      analyzing: false,
    }
  },
  async mounted() {
    await this.loadLogs()
  },
  methods: {
    async loadLogs() {
      try {
        if (window.go?.main?.App) this.logFiles = await window.go.main.App.GetLogFiles() || []
      } catch (e) { console.error(e) }
    },
    selectLog(name) { this.selectedLog = name },
    toggleLogSection() {
      this.logSectionExpanded = !this.logSectionExpanded
    },
    unlockLogs() {
      if (this.logPassword === 'Lenovo2026') {
        this.logUnlocked = true
        this.logPasswordError = ''
        this.logPassword = ''
        this.loadLogs()
      } else {
        this.logPasswordError = 'Need Dispatcher owner check or contact zhoushang2'
        setTimeout(() => { this.logPasswordError = '' }, 3000)
      }
    },
    async runAnalysis() {
      this.analyzing = true
      this.summary = null
      try {
        if (window.go?.main?.App) this.summary = await window.go.main.App.GetLogSummary()
      } catch (e) { console.error(e) }
      finally { this.analyzing = false }
    },
    async loadRawLog() {
      try {
        if (window.go?.main?.App) this.rawLog = await window.go.main.App.ReadLogTail(this.tailLines)
      } catch (e) { console.error(e) }
    },
    formatSize(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    },
  },
}
</script>

<style scoped>
.log-section { display: flex; flex-direction: column; gap: 16px; }

.log-file-list { display: flex; flex-direction: column; gap: 4px; max-height: 220px; overflow-y: auto; }
.log-file-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; border-radius: 6px; cursor: pointer; transition: background 0.15s; border: 1px solid transparent; }
.log-file-item:hover { background: var(--bg-card-hover); }
.log-file-item.active { background: rgba(230,63,50,0.08); border-color: rgba(230,63,50,0.25); }
.log-name { font-size: 13px; font-weight: 500; color: var(--text-primary); font-family: 'Consolas','Monaco',monospace; }
.log-meta { font-size: 11px; color: var(--text-tertiary); white-space: nowrap; margin-left: 12px; }

.advanced-section { border-color: rgba(245, 158, 11, 0.3); }
.advanced-toggle { cursor: pointer; user-select: none; }
.advanced-toggle .chevron { transition: transform 0.2s; margin-left: auto; }
.advanced-toggle .chevron.open { transform: rotate(180deg); }
.password-prompt { padding: 16px 0 0 0; }
.password-row { display: flex; align-items: center; gap: 8px; }
.password-input { flex: 1; padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-primary); color: var(--text-primary); font-size: 13px; outline: none; transition: border-color 0.2s; }
.password-input:focus { border-color: #F59E0B; }
.password-input::placeholder { color: var(--text-tertiary); font-family: inherit; }
.btn-unlock { display: flex; align-items: center; gap: 4px; padding: 8px 16px; border: 1px solid #F59E0B; border-radius: 6px; background: rgba(245,158,11,0.1); color: #F59E0B; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; white-space: nowrap; }
.btn-unlock:hover { background: rgba(245,158,11,0.2); border-color: #F59E0B; }
.password-error { margin-top: 8px; font-size: 12px; color: #EF4444; }

.summary-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-top: 4px; }
.summary-panel { background: var(--bg-tertiary); border-radius: 8px; padding: 14px; border: 1px solid var(--border-color); }
.panel-errors { border-color: rgba(244,67,54,0.2); background: rgba(244,67,54,0.04); }
.panel-title { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-secondary); margin-bottom: 10px; display: flex; align-items: center; gap: 6px; }
.error-title { color: #F44336; }
.log-lines { display: flex; flex-direction: column; gap: 4px; max-height: 160px; overflow-y: auto; }
.log-line { font-size: 11px; font-family: 'Consolas','Monaco',monospace; color: var(--text-secondary); line-height: 1.5; word-break: break-all; padding: 2px 0; border-bottom: 1px solid var(--border-color); }
.log-line:last-child { border-bottom: none; }
.error-line { color: #F44336; }
.no-data { font-size: 12px; color: var(--text-tertiary); font-style: italic; }
.ok-hint { color: var(--accent-green); font-style: normal; font-weight: 600; }

.btn-sm { padding: 6px 12px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 6px; color: var(--text-secondary); font-size: 12px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit; }
.btn-sm:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }
.btn-analyze { padding: 7px 16px; background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%); border: none; border-radius: 6px; color: white; font-size: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit; }
.btn-analyze:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-analyze:disabled { opacity: 0.6; cursor: not-allowed; }
.empty-hint { padding: 24px; text-align: center; color: var(--text-tertiary); font-size: 13px; }
.analyzing-hint { display: flex; align-items: center; justify-content: center; gap: 10px; padding: 32px; color: var(--text-secondary); font-size: 13px; }
@keyframes spin { to { transform: rotate(360deg); } }
.spinning { animation: spin 0.8s linear infinite; }
</style>
