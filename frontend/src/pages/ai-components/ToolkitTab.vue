<template>
  <div class="toolkit-section">
    <!-- Header Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
          </svg>
          System Tools - One-Click Install
        </span>
        <button class="btn-sm" @click="openToolkitFolder">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
          </svg>
          Open Folder
        </button>
      </div>
      <div class="toolkit-path">Install Directory: {{ toolkitInstallDir }}</div>
    </div>

    <!-- Category Filter -->
    <div class="category-filter">
      <button :class="['cat-btn', { active: toolkitCategory === 'all' }]" @click="toolkitCategory = 'all'">All ({{ toolkitTools.length }})</button>
      <button :class="['cat-btn', { active: toolkitCategory === 'monitor' }]" @click="toolkitCategory = 'monitor'">Monitor</button>
      <button :class="['cat-btn', { active: toolkitCategory === 'system' }]" @click="toolkitCategory = 'system'">System Info</button>
      <button :class="['cat-btn', { active: toolkitCategory === 'benchmark' }]" @click="toolkitCategory = 'benchmark'">Benchmark</button>
      <button :class="['cat-btn', { active: toolkitCategory === 'diagnostic' }]" @click="toolkitCategory = 'diagnostic'">Diagnostic</button>
    </div>

    <!-- Tools Grid -->
    <div class="tools-grid">
      <div v-for="tool in filteredTools" :key="tool.id" class="tool-card" :class="{ installed: toolInstallStatus[tool.id]?.installed }">
        <div class="tool-header">
          <div class="tool-icon" :class="tool.category">
            <svg v-if="tool.category === 'monitor'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
            </svg>
            <svg v-else-if="tool.category === 'system'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/>
            </svg>
            <svg v-else-if="tool.category === 'benchmark'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
            </svg>
            <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
          </div>
          <div class="tool-info">
            <span class="tool-name">{{ tool.name }}</span>
            <span class="tool-vendor">{{ tool.vendor }}</span>
          </div>
          <span v-if="toolInstallStatus[tool.id]?.installed" class="installed-badge">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            Installed
          </span>
        </div>

        <div class="tool-desc">{{ tool.description }}</div>

        <div class="tool-meta">
          <span class="tool-version">v{{ tool.version }}</span>
          <span class="tool-size">{{ tool.sizeMb }} MB</span>
          <span class="tool-winget" v-if="tool.wingetId">已安装 winget</span>
        </div>

        <div v-if="isToolBusy(tool.id)" class="tool-installing">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          <span>{{ toolProgress[tool.id]?.message || 'Installing...' }}</span>
        </div>

        <div v-if="toolProgress[tool.id]?.status === 'error'" class="tool-error">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          {{ toolProgress[tool.id].message }}
        </div>

        <div v-if="toolProgress[tool.id]?.status === 'completed'" class="tool-success">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="20 6 9 17 4 12"/>
          </svg>
          {{ toolProgress[tool.id].message }}
        </div>

        <div class="tool-actions">
          <button v-if="!toolInstallStatus[tool.id]?.installed && !isToolBusy(tool.id)" class="btn-install" @click="installTool(tool.id)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            Install & Run
          </button>
          <button v-if="isToolBusy(tool.id)" class="btn-installing" disabled>
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            Installing...
          </button>
          <button v-if="toolInstallStatus[tool.id]?.installed" class="btn-run" @click="runTool(tool.id)">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            Run
          </button>
          <a :href="tool.website" target="_blank" class="btn-link">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
              <polyline points="15 3 21 3 21 9"/>
              <line x1="10" y1="14" x2="21" y2="3"/>
            </svg>
            Website
          </a>
        </div>
      </div>
    </div>

    <!-- Quick Install All -->
    <div class="card" v-if="!allToolsInstalled">
      <div class="card-header">
        <span class="card-title">Quick Actions</span>
      </div>
      <div class="quick-actions">
        <button class="btn-batch" @click="installEssentials" :disabled="batchInstalling">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
            <polyline points="7 10 12 15 17 10"/>
            <line x1="12" y1="15" x2="12" y2="3"/>
          </svg>
          {{ batchInstalling ? 'Installing...' : 'Install Essential Tools (HWiNFO, CPU-Z, GPU-Z)' }}
        </button>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ToolkitTab',
  props: {
    theme: { type: String, default: 'dark' },
  },
  data() {
    return {
      toolkitTools: [],
      toolkitInstallDir: 'C:\\LenovoDispatcherToolkit\\Tools',
      toolkitCategory: 'all',
      toolInstallStatus: {},
      toolProgress: {},
      batchInstalling: false,
    }
  },
  async mounted() {
    await this.loadToolkitData()
  },
  methods: {
    async loadToolkitData() {
      try {
        if (window.go?.main?.App) {
          this.toolkitTools = await window.go.main.App.GetToolkitTools() || []
          this.toolkitInstallDir = await window.go.main.App.GetToolkitInstallDir() || this.toolkitInstallDir
          await this.refreshToolkitStatus()
        }
      } catch (e) { console.error(e) }
    },
    async refreshToolkitStatus() {
      try {
        if (window.go?.main?.App) {
          const statuses = await window.go.main.App.CheckAllToolkitInstalled() || []
          this.toolInstallStatus = {}
          statuses.forEach(s => {
            this.toolInstallStatus[s.toolId] = s
          })
        }
      } catch (e) { console.error(e) }
    },
    async installTool(toolId) {
      if (!window.go?.main?.App) return
      this.toolProgress[toolId] = { status: 'installing', progress: 0, message: 'Installing via winget...' }
      try {
        await window.go.main.App.InstallToolkitTool(toolId)
        const pollProgress = async () => {
          if (!this.toolProgress[toolId] || this.toolProgress[toolId].status === 'completed' || this.toolProgress[toolId].status === 'error') {
            return
          }
          try {
            const progress = await window.go.main.App.GetToolkitProgress(toolId)
            this.toolProgress[toolId] = progress
            if (progress.status === 'installing') {
              setTimeout(pollProgress, 1000)
            } else if (progress.status === 'completed') {
              await this.refreshToolkitStatus()
            }
          } catch (e) {
            console.error('Poll error:', e)
          }
        }
        setTimeout(pollProgress, 500)
      } catch (e) {
        this.toolProgress[toolId] = { status: 'error', message: e.message || String(e) }
      }
    },
    async runTool(toolId) {
      try {
        if (window.go?.main?.App) {
          const result = await window.go.main.App.RunToolkitTool(toolId)
          if (result.startsWith('Error')) {
            alert(result)
          }
        }
      } catch (e) { console.error(e) }
    },
    async openToolkitFolder() {
      try {
        if (window.go?.main?.App) {
          await window.go.main.App.OpenToolkitFolder()
        }
      } catch (e) { console.error(e) }
    },
    isToolBusy(toolId) {
      const p = this.toolProgress[toolId]
      return p && p.status === 'installing'
    },
    async installEssentials() {
      this.batchInstalling = true
      const essentials = ['hwinfo64', 'cpuz', 'gpuz']
      for (const id of essentials) {
        if (!this.toolInstallStatus[id]?.installed) {
          await this.installTool(id)
          let attempts = 0
          while (this.isToolBusy(id) && attempts < 120) {
            await new Promise(r => setTimeout(r, 1000))
            attempts++
          }
        }
      }
      this.batchInstalling = false
    },
  },
  computed: {
    filteredTools() {
      if (this.toolkitCategory === 'all') return this.toolkitTools
      return this.toolkitTools.filter(t => t.category === this.toolkitCategory)
    },
    allToolsInstalled() {
      return this.toolkitTools.every(t => this.toolInstallStatus[t.id]?.installed)
    },
  },
}
</script>

<style scoped>
.toolkit-section { display: flex; flex-direction: column; gap: 16px; }
.toolkit-path { font-size: 12px; color: var(--text-tertiary); padding: 8px 12px; background: var(--bg-tertiary); border-radius: 6px; font-family: 'Consolas', monospace; }

.category-filter { display: flex; gap: 8px; flex-wrap: wrap; }
.cat-btn {
  padding: 6px 14px; background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 16px; color: var(--text-secondary); font-size: 12px; font-weight: 500;
  cursor: pointer; transition: var(--transition); font-family: inherit;
}
.cat-btn:hover { border-color: var(--lenovo-red); color: var(--text-primary); }
.cat-btn.active { background: var(--lenovo-red); border-color: var(--lenovo-red); color: white; }

.tools-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
.tool-card {
  background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 10px; padding: 16px; display: flex; flex-direction: column; gap: 10px;
  transition: var(--transition);
}
.tool-card:hover { border-color: var(--border-light); background: var(--bg-card-hover); }
.tool-card.installed { border-color: rgba(76, 175, 80, 0.3); }
.tool-header { display: flex; align-items: center; gap: 12px; }
.tool-icon {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-tertiary); color: var(--text-secondary);
}
.tool-icon.monitor { background: rgba(33, 150, 243, 0.15); color: #2196F3; }
.tool-icon.system { background: rgba(156, 39, 176, 0.15); color: #9C27B0; }
.tool-icon.benchmark { background: rgba(255, 152, 0, 0.15); color: #FF9800; }
.tool-icon.diagnostic { background: rgba(76, 175, 80, 0.15); color: #4CAF50; }
.tool-info { flex: 1; display: flex; flex-direction: column; gap: 2px; }
.tool-name { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.tool-vendor { font-size: 11px; color: var(--text-tertiary); }
.installed-badge {
  display: flex; align-items: center; gap: 4px;
  padding: 3px 8px; background: rgba(76, 175, 80, 0.15);
  border-radius: 10px; color: #4CAF50; font-size: 11px; font-weight: 600;
}
.tool-desc { font-size: 12px; color: var(--text-secondary); line-height: 1.5; }
.tool-meta { display: flex; gap: 12px; font-size: 11px; color: var(--text-tertiary); }
.tool-version { font-family: 'Consolas', monospace; }
.tool-winget { background: rgba(33, 150, 243, 0.15); color: #2196F3; padding: 2px 6px; border-radius: 4px; }
.tool-installing { display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: rgba(255, 152, 0, 0.08); border-radius: 6px; color: #FF9800; font-size: 12px; }
.tool-error, .tool-success { display: flex; align-items: center; gap: 6px; padding: 8px 10px; border-radius: 6px; font-size: 11px; }
.tool-error { background: rgba(244, 67, 54, 0.08); color: #F44336; border: 1px solid rgba(244, 67, 54, 0.2); }
.tool-success { background: rgba(76, 175, 80, 0.08); color: #4CAF50; border: 1px solid rgba(76, 175, 80, 0.2); }
.tool-actions { display: flex; gap: 8px; margin-top: 4px; }
.btn-install, .btn-installing {
  flex: 1; padding: 8px 12px; border: none; border-radius: 6px;
  font-size: 12px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
}
.btn-install { background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%); color: white; }
.btn-install:hover { opacity: 0.9; transform: translateY(-1px); }
.btn-installing { background: var(--bg-tertiary); color: var(--text-secondary); cursor: not-allowed; }
.btn-run {
  flex: 1; padding: 8px 12px; border: none; border-radius: 6px;
  background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%);
  color: white; font-size: 12px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
}
.btn-run:hover { opacity: 0.9; transform: translateY(-1px); }
.btn-link {
  padding: 8px 12px; background: var(--bg-tertiary); border: 1px solid var(--border-color);
  border-radius: 6px; color: var(--text-secondary); font-size: 12px; font-weight: 500;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit; text-decoration: none;
}
.btn-link:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }
.quick-actions { display: flex; gap: 12px; flex-wrap: wrap; }
.btn-batch {
  padding: 10px 20px; background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border: none; border-radius: 8px; color: white; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 8px;
  transition: var(--transition); font-family: inherit;
}
.btn-batch:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-batch:disabled { opacity: 0.6; cursor: not-allowed; }
.btn-sm { padding: 6px 12px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 6px; color: var(--text-secondary); font-size: 12px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit; }
.btn-sm:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }
@keyframes spin { to { transform: rotate(360deg); } }
.spinning { animation: spin 0.8s linear infinite; }
</style>
