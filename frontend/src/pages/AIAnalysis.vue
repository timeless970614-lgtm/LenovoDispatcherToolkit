<template>
  <div class="ai-analysis-page">

    <!-- Log Files Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <polyline points="10 9 9 9 8 9"/>
          </svg>
          Log Files
        </span>
        <button class="btn-sm" @click="loadLogs">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 4 23 10 17 10"/><polyline points="1 20 1 14 7 14"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/></svg>
          Refresh
        </button>
      </div>
      <div v-if="logFiles.length" class="log-file-list">
        <div
          v-for="f in logFiles"
          :key="f.name"
          class="log-file-item"
          :class="{ active: selectedLog === f.name }"
          @click="selectLog(f.name)"
        >
          <span class="log-name">{{ f.name }}</span>
          <span class="log-meta">{{ formatSize(f.size) }} · {{ f.modTime }}</span>
        </div>
      </div>
      <div v-else class="empty-hint">No log files found in {{ logDir }}</div>
    </div>

    <!-- AI Analysis Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 8v4l3 3"/>
          </svg>
          AI Analysis
        </span>
        <button class="btn-analyze" @click="runAnalysis" :disabled="analyzing">
          <svg v-if="!analyzing" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg>
          <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
          {{ analyzing ? 'Analyzing...' : 'Analyze Latest Log' }}
        </button>
      </div>

      <!-- Summary Panels -->
      <div v-if="summary" class="summary-grid">
        <div class="summary-panel">
          <div class="panel-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/></svg>
            Workload Levels
          </div>
          <div class="log-lines">
            <div v-for="(l, i) in summary.workloadLevels" :key="i" class="log-line">{{ l }}</div>
            <div v-if="!summary.workloadLevels?.length" class="no-data">No data</div>
          </div>
        </div>
        <div class="summary-panel">
          <div class="panel-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/></svg>
            Turbo Events
          </div>
          <div class="log-lines">
            <div v-for="(l, i) in summary.turboEvents" :key="i" class="log-line">{{ l }}</div>
            <div v-if="!summary.turboEvents?.length" class="no-data">No data</div>
          </div>
        </div>
        <div class="summary-panel">
          <div class="panel-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
            Foreground Apps
          </div>
          <div class="log-lines">
            <div v-for="(l, i) in summary.appEvents" :key="i" class="log-line">{{ l }}</div>
            <div v-if="!summary.appEvents?.length" class="no-data">No data</div>
          </div>
        </div>
        <div class="summary-panel panel-errors">
          <div class="panel-title error-title">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            Errors / Warnings
          </div>
          <div class="log-lines">
            <div v-for="(l, i) in summary.errors" :key="i" class="log-line error-line">{{ l }}</div>
            <div v-if="!summary.errors?.length" class="no-data ok-hint">✓ No errors found</div>
          </div>
        </div>
      </div>
      <div v-else-if="!analyzing" class="empty-hint">Click "Analyze Latest Log" to start</div>
      <div v-else class="analyzing-hint">
        <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
        Reading and parsing log...
      </div>
    </div>

    <!-- Raw Log Viewer Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <polyline points="4 17 10 11 4 5"/>
            <line x1="12" y1="19" x2="20" y2="19"/>
          </svg>
          Raw Log (last {{ tailLines }} lines)
        </span>
        <div class="header-controls">
          <select v-model="tailLines" class="lines-select" @change="loadRawLog">
            <option :value="100">100 lines</option>
            <option :value="200">200 lines</option>
            <option :value="500">500 lines</option>
            <option :value="1000">1000 lines</option>
          </select>
          <button class="btn-sm" @click="loadRawLog">Load</button>
        </div>
      </div>
      <div v-if="rawLog" class="raw-log-box">
        <pre>{{ rawLog }}</pre>
      </div>
      <div v-else class="empty-hint">Click Load to view raw log</div>
    </div>

  </div>
</template>

<script>
export default {
  name: 'AIAnalysis',
  data() {
    return {
      logDir: 'C:\\ProgramData\\Lenovo\\LenovoDispatcher\\Logs',
      logFiles: [],
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
        if (window.go?.main?.App) {
          this.logFiles = await window.go.main.App.GetLogFiles() || []
        }
      } catch (e) { console.error(e) }
    },
    selectLog(name) {
      this.selectedLog = name
    },
    async runAnalysis() {
      this.analyzing = true
      this.summary = null
      try {
        if (window.go?.main?.App) {
          this.summary = await window.go.main.App.GetLogSummary()
        }
      } catch (e) {
        console.error(e)
      } finally {
        this.analyzing = false
      }
    },
    async loadRawLog() {
      try {
        if (window.go?.main?.App) {
          this.rawLog = await window.go.main.App.ReadLogTail(this.tailLines)
        }
      } catch (e) { console.error(e) }
    },
    formatSize(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    },
  }
}
</script>

<style scoped>
.ai-analysis-page {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

/* Log file list */
.log-file-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 220px;
  overflow-y: auto;
}

.log-file-item {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
  border: 1px solid transparent;
}

.log-file-item:hover {
  background: var(--bg-card-hover);
}

.log-file-item.active {
  background: rgba(230, 63, 50, 0.08);
  border-color: rgba(230, 63, 50, 0.25);
}

.log-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
  font-family: 'Consolas', 'Monaco', monospace;
}

.log-meta {
  font-size: 11px;
  color: var(--text-tertiary);
  white-space: nowrap;
  margin-left: 12px;
}

/* Summary grid */
.summary-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-top: 4px;
}

.summary-panel {
  background: var(--bg-tertiary);
  border-radius: 8px;
  padding: 14px;
  border: 1px solid var(--border-color);
}

.panel-errors {
  border-color: rgba(244, 67, 54, 0.2);
  background: rgba(244, 67, 54, 0.04);
}

.panel-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-secondary);
  margin-bottom: 10px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.error-title {
  color: #F44336;
}

.log-lines {
  display: flex;
  flex-direction: column;
  gap: 4px;
  max-height: 160px;
  overflow-y: auto;
}

.log-line {
  font-size: 11px;
  font-family: 'Consolas', 'Monaco', monospace;
  color: var(--text-secondary);
  line-height: 1.5;
  word-break: break-all;
  padding: 2px 0;
  border-bottom: 1px solid var(--border-color);
}

.log-line:last-child {
  border-bottom: none;
}

.error-line {
  color: #F44336;
}

.no-data {
  font-size: 12px;
  color: var(--text-tertiary);
  font-style: italic;
}

.ok-hint {
  color: var(--accent-green);
  font-style: normal;
  font-weight: 600;
}

/* Raw log */
.raw-log-box {
  background: var(--bg-tertiary);
  border-radius: 8px;
  padding: 16px;
  max-height: 400px;
  overflow-y: auto;
  border: 1px solid var(--border-color);
}

.raw-log-box pre {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 11px;
  color: var(--text-secondary);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

/* Buttons */
.btn-sm {
  padding: 6px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 500;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: var(--transition);
  font-family: inherit;
}

.btn-sm:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
  border-color: var(--border-light);
}

.btn-analyze {
  padding: 7px 16px;
  background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 6px;
  transition: var(--transition);
  font-family: inherit;
}

.btn-analyze:hover:not(:disabled) {
  opacity: 0.9;
  transform: translateY(-1px);
}

.btn-analyze:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.header-controls {
  display: flex;
  align-items: center;
  gap: 8px;
}

.lines-select {
  padding: 5px 8px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 12px;
  font-family: inherit;
  cursor: pointer;
}

.empty-hint {
  padding: 24px;
  text-align: center;
  color: var(--text-tertiary);
  font-size: 13px;
}

.analyzing-hint {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 32px;
  color: var(--text-secondary);
  font-size: 13px;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spinning {
  animation: spin 0.8s linear infinite;
}
</style>
