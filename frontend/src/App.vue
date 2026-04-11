<template>
  <div id="app" :class="theme">
    <!-- Sidebar -->
    <aside class="sidebar">
      <div class="sidebar-brand">
        <div class="brand-logo">
          <svg width="32" height="32" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="brand-text">
          <span class="brand-name">Lenovo</span>
          <span class="brand-sub">Dispatcher Toolkit</span>
        </div>
      </div>
      
      <nav class="sidebar-nav">
        <div 
          v-for="item in mainNavItems" 
          :key="item.id"
          :class="['nav-item', { active: currentPage === item.id }]"
          @click="currentPage = item.id"
        >
          <span class="nav-icon" v-html="item.icon"></span>
          <span class="nav-label">{{ item.label }}</span>
          <span class="nav-indicator" v-if="currentPage === item.id"></span>
        </div>
      </nav>

      <nav class="sidebar-nav-bottom">
        <div 
          v-for="item in bottomNavItems" 
          :key="item.id"
          :class="['nav-item', { active: currentPage === item.id }]"
          @click="currentPage = item.id"
        >
          <span class="nav-icon" v-html="item.icon"></span>
          <span class="nav-label">{{ item.label }}</span>
          <span class="nav-indicator" v-if="currentPage === item.id"></span>
        </div>
      </nav>

      <div class="sidebar-footer">
        <div class="sidebar-footer-icons">
          <div class="icon-btn" @click="toggleLang" :title="lang === 'en' ? '切换到中文' : 'Switch to English'">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/>
              <line x1="2" y1="12" x2="22" y2="12"/>
              <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z"/>
            </svg>
          </div>
          <div class="theme-toggle" @click="cycleTheme" :title="'Theme: ' + themeLabel">
            <svg v-if="theme === 'dark'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="5"/>
              <line x1="12" y1="1" x2="12" y2="3"/>
              <line x1="12" y1="21" x2="12" y2="23"/>
              <line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/>
              <line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/>
              <line x1="1" y1="12" x2="3" y2="12"/>
              <line x1="21" y1="12" x2="23" y2="12"/>
              <line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/>
              <line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/>
            </svg>
            <svg v-else-if="theme === 'light'" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
            </svg>
            <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="5"/>
              <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
            </svg>
          </div>
        </div>
        <div class="version-tag">v1.0.15</div>
      </div>
    </aside>

    <!-- Main Content -->
    <main class="main-content">
      <header class="content-header">
        <div class="header-left">
          <h1 class="page-title">{{ currentPageTitle }}</h1>
          <p class="page-subtitle">{{ currentPageSubtitle }}</p>
        </div>
        <div class="header-right">
          <!-- ML Log Capture Button -->
          <button class="btn-ml-capture" :class="{ capturing: mlCapturing }" @click="toggleMLCapture" :title="mlCapturing ? 'Stop capture (LOG+CSV)' : 'Start capture (LOG+CSV)'">
            <svg v-if="!mlCapturing" width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
              <circle cx="12" cy="12" r="6"/>
              <rect x="10" y="2" width="4" height="6" rx="1"/>
            </svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none" class="rec-dot">
              <circle cx="12" cy="12" r="8"/>
            </svg>
            <span>{{ mlCapturing ? 'Stop' : 'ML Log' }}</span>
            <span class="ml-format">LOG+CSV</span>
          </button>

          <div class="service-status-pill" v-if="serviceRunning !== null">
            <span :class="['svc-dot', serviceRunning ? 'svc-dot-running' : 'svc-dot-stopped']"></span>
            <span class="svc-label">{{ serviceRunning ? 'Running' : 'Stopped' }}</span>
          </div>
          <div class="current-mode-pill" v-if="currentMode !== ''">
            <span class="mode-dot"></span>
            <span class="mode-label">Current Mode</span>
            <span class="mode-value">{{ pinnedMode || currentMode }}</span>
            <span v-if="pinnedMode" class="pin-indicator">📌</span>
          </div>
          <!-- Uninstall Button -->
          <button class="btn-uninstall" @click="handleUninstall" :disabled="uninstalling" title="Uninstall Dispatcher Driver">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M3 6h18M19 6v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6m3 0V4a2 2 0 012-2h4a2 2 0 012 2v2"/>
              <line x1="10" y1="11" x2="10" y2="17"/>
              <line x1="14" y1="11" x2="14" y2="17"/>
            </svg>
            <span>{{ uninstalling ? 'Uninstalling...' : 'Uninstall' }}</span>
          </button>
        </div>
      </header>
      
      <div class="content-body">
        <Dashboard v-if="currentPage === 'dashboard'" :theme="theme" />
        <PPMDriver v-else-if="currentPage === 'ppm'" :theme="theme" />
        <MSR v-else-if="currentPage === 'msr'" :theme="theme" />

        <FunctionCheck v-else-if="currentPage === 'funccheck'" :theme="theme" />
        <ModeCheck v-else-if="currentPage === 'modecheck'" :theme="theme" />
        <AIAnalysis v-else-if="currentPage === 'aianalysis'" :theme="theme" />
        <Settings v-else-if="currentPage === 'settings'" :theme="theme" :lang="lang" @update:theme="setTheme" @update:lang="setLang" />
        <About v-else-if="currentPage === 'about'" :theme="theme" />
      </div>
    </main>
  </div>
</template>

<script>
import Dashboard from './pages/Dashboard.vue'
import FunctionCheck from './pages/FunctionCheck.vue'
import ModeCheck from './pages/ModeCheck.vue'
import AIAnalysis from './pages/AIAnalysis.vue'
import MSR from './pages/MSR.vue'
import Settings from './pages/Settings.vue'
import About from './pages/About.vue'
import { StartMLScenarioCapture, StopMLScenarioCapture, GetMLLogStatus, OpenFolder, UninstallDispatcher } from '../wailsjs/go/main/App'

export default {
  name: 'App',
  components: {
    Dashboard,
    FunctionCheck,
    ModeCheck,
    AIAnalysis,
    MSR,
    Settings,
    About
  },
  data() {
    return {
      currentPage: 'dashboard',
      theme: 'dark',
      lang: 'en',
      currentTime: '',
      currentMode: 'N/A',
      pinnedMode: '',
      serviceRunning: null,
      mlCapturing: false,
      mlEventCount: 0,
      mlResult: null,
      mlInterval: null,
      timeInterval: null,
      modeInterval: null,
      uninstalling: false,
      mainNavItems: [
        { 
          id: 'dashboard', 
          label: 'Dashboard', 
          subtitle: 'System overview & status',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>'
        },
        { 
          id: 'modecheck', 
          label: 'Mode Check', 
          subtitle: 'Dispatcher status & features',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>'
        },
        { 
          id: 'funccheck', 
          label: 'Function Check', 
          subtitle: 'GPU & system diagnostics',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>'
        },
        {
          id: 'aianalysis',
          label: 'AI Analysis',
          subtitle: 'Log analysis & diagnostics',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/></svg>'
        },

        {
          id: 'ppm', 
          label: 'PPM Driver', 
          subtitle: 'Processor power management',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/><circle cx="12" cy="12" r="4"/></svg>'
        },
        { 
          id: 'msr', 
          label: 'MSR', 
          subtitle: 'Model Specific Registers',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/><line x1="8" y1="7" x2="16" y2="7"/><line x1="8" y1="11" x2="16" y2="11"/><line x1="8" y1="15" x2="12" y2="15"/></svg>'
        },
      ],
      bottomNavItems: [
        { 
          id: 'settings', 
          label: 'Settings', 
          subtitle: 'Application preferences',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/></svg>'
        },
        { 
          id: 'about', 
          label: 'About', 
          subtitle: 'Application information',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/></svg>'
        }
      ]
    }
  },
  computed: {
    currentPageTitle() {
      const allItems = [...this.mainNavItems, ...this.bottomNavItems]
      const item = allItems.find(i => i.id === this.currentPage)
      return item ? item.label : ''
    },
    currentPageSubtitle() {
      const allItems = [...this.mainNavItems, ...this.bottomNavItems]
      const item = allItems.find(i => i.id === this.currentPage)
      return item ? item.subtitle : ''
    },
    themeLabel() {
      const labels = { dark: 'Dark', light: 'Light', warm: 'Warm' }
      return labels[this.theme] || 'Dark'
    }
  },
  mounted() {
    // Load saved theme from localStorage
    const savedTheme = localStorage.getItem('lenovo-toolkit-theme')
    if (savedTheme) {
      this.theme = savedTheme
    }
    // Load saved language from localStorage
    const savedLang = localStorage.getItem('lenovo-toolkit-lang')
    if (savedLang) {
      this.lang = savedLang
    }
    this.updateTime()
    this.timeInterval = setInterval(this.updateTime, 1000)
    this.updateMode()
    this.modeInterval = setInterval(this.updateMode, 2000)
  },
  beforeUnmount() {
    if (this.timeInterval) clearInterval(this.timeInterval)
    if (this.modeInterval) clearInterval(this.modeInterval)
  },
  methods: {
    updateTime() {
      const now = new Date()
      this.currentTime = now.toLocaleTimeString('en-US', { 
        hour: '2-digit', 
        minute: '2-digit',
        hour12: true 
      })
    },
    async updateMode() {
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const [info, pinned] = await Promise.all([
            window.go.main.App.GetDispatcherInfo(),
            window.go.main.App.GetPinnedDYTCMode(),
          ])
          this.pinnedMode = pinned || ''
          
          // Use AutoMode directly (it's the raw number from registry)
          if (info && info.AutoMode) {
            const num = info.AutoMode
            const numToMode = { 
              1: 'BSM', 2: 'IBSM', 3: 'AQM', 4: 'STD', 
              5: 'APM', 6: 'IEPM', 7: 'EPM', 
              8: 'Tablet', 9: 'Tent', 10: 'Flat', 11: 'GEEK' 
            }
            this.currentMode = numToMode[num] || ('Unknown(' + num + ')')
          }
          
          const status = await window.go.main.App.GetServiceStatus()
          this.serviceRunning = (status === 'Running')
        }
      } catch (e) {
        // silent
      }
    },
    cycleTheme() {
      const themes = ['dark', 'light', 'warm']
      const currentIndex = themes.indexOf(this.theme)
      const nextIndex = (currentIndex + 1) % themes.length
      this.theme = themes[nextIndex]
      localStorage.setItem('lenovo-toolkit-theme', this.theme)
    },
    setTheme(newTheme) {
      this.theme = newTheme
      localStorage.setItem('lenovo-toolkit-theme', this.theme)
    },
    setLang(newLang) {
      this.lang = newLang
      localStorage.setItem('lenovo-toolkit-lang', this.lang)
    },
    toggleLang() {
      this.lang = this.lang === 'en' ? 'zh' : 'en'
      localStorage.setItem('lenovo-toolkit-lang', this.lang)
    },

    async toggleMLCapture() {
      if (this.mlCapturing) {
        // Stop capture
        try {
          const result = await StopMLScenarioCapture()
          this.mlCapturing = false
          if (this.mlInterval) {
            clearInterval(this.mlInterval)
            this.mlInterval = null
          }
        } catch (e) {
          console.error('StopMLCapture error:', e)
          this.mlCapturing = false
        }
      } else {
        // Start capture
        try {
          const result = await StartMLScenarioCapture()
          if (result && result.error) {
            console.error('StartMLCapture error:', result.error)
            return
          }
          this.mlCapturing = true
          this.mlEventCount = 0
          this.mlResult = null
          // Poll status every second
          if (this.mlInterval) clearInterval(this.mlInterval)
          this.mlInterval = setInterval(async () => {
            try {
              const status = await GetMLLogStatus()
              this.mlEventCount = status.eventCount || 0
            } catch (e) { /* ignore */ }
          }, 1000)
        } catch (e) {
          console.error('StartMLCapture error:', e)
        }
      }
    },

    openMLLogDir() {
      if (!this.mlResult) return
      const dir = this.mlResult.outputFile
      const lastSep = dir.lastIndexOf('\\')
      const folder = lastSep > 0 ? dir.substring(0, lastSep) : dir.substring(0, dir.lastIndexOf('/'))
      OpenFolder(folder)
    },

    async handleUninstall() {
      // Show confirmation dialog
      const confirmed = confirm('Are you sure you want to uninstall the Lenovo Process Management Dispatcher driver?\n\nThis will remove the driver from your system.')
      if (!confirmed) {
        return
      }

      this.uninstalling = true
      try {
        const result = await UninstallDispatcher()
        if (result.success) {
          alert(result.message)
        } else {
          alert('Uninstall failed: ' + result.message)
        }
      } catch (e) {
        alert('Uninstall error: ' + e)
      } finally {
        this.uninstalling = false
      }
    }
  }
}
</script>

<style>
#app.dark {
  --bg-primary: #0D0D0D;
  --bg-secondary: #161616;
  --bg-tertiary: #1E1E1E;
  --bg-card: #232323;
  --bg-card-hover: #2A2A2A;
  --text-primary: #FFFFFF;
  --text-secondary: #A0A0A0;
  --text-tertiary: #666666;
  --border-color: #333333;
  --border-light: #404040;
}

#app.light {
  --bg-primary: #F5F5F5;
  --bg-secondary: #FFFFFF;
  --bg-tertiary: #EEEEEE;
  --bg-card: #FFFFFF;
  --bg-card-hover: #F0F0F0;
  --text-primary: #1A1A1A;
  --text-secondary: #666666;
  --text-tertiary: #999999;
  --border-color: #E0E0E0;
  --border-light: #CCCCCC;
}

#app.warm {
  --bg-primary: #FDF6E3;
  --bg-secondary: #F5E6C8;
  --bg-tertiary: #EDE0C0;
  --bg-card: #FAF0D7;
  --bg-card-hover: #F5E8CC;
  --text-primary: #5C4B37;
  --text-secondary: #8B7355;
  --text-tertiary: #A89080;
  --border-color: #E8D5B5;
  --border-light: #D4C4A8;
}

#app {
  --lenovo-red: #E63F32;
  --lenovo-red-light: #FF5A4D;
  --lenovo-red-dark: #C4352A;
  --accent-green: #4ADE80;
  --accent-blue: #60A5FA;
  --accent-yellow: #FBBF24;
  --shadow-sm: 0 1px 2px rgba(0, 0, 0, 0.1);
  --shadow-md: 0 4px 12px rgba(0, 0, 0, 0.15);
  --shadow-lg: 0 8px 24px rgba(0, 0, 0, 0.2);
  --radius-sm: 6px;
  --radius-md: 10px;
  --radius-lg: 14px;
  --transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
}

/* ML Log Capture Button */
.btn-ml-capture {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
  white-space: nowrap;
  min-width: 100px;
  justify-content: center;
}

.btn-ml-capture:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.05);
}

.btn-ml-capture.capturing {
  border-color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.1);
  color: var(--lenovo-red);
}

.btn-ml-capture .rec-dot {
  animation: ml-pulse 1s ease-in-out infinite;
}

@keyframes ml-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}

.ml-count {
  font-size: 11px;
  font-weight: 700;
  background: var(--lenovo-red);
  color: white;
  border-radius: 10px;
  padding: 1px 6px;
  min-width: 20px;
  text-align: center;
}

.ml-format {
  font-size: 10px;
  font-weight: 600;
  color: var(--accent-blue);
  opacity: 0.7;
  letter-spacing: 0.5px;
}

/* ML Result Toast */
.ml-result-toast {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  border-radius: var(--radius-md);
  background: rgba(74, 222, 128, 0.1);
  border: 1px solid rgba(74, 222, 128, 0.3);
  color: var(--accent-green);
  font-size: 12px;
  white-space: nowrap;
  animation: ml-toast-in 0.3s ease;
}

.ml-toast-title {
  font-weight: 700;
}

.ml-toast-info {
  color: var(--text-secondary);
  font-size: 11px;
}

.ml-toast-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 26px;
  height: 26px;
  border: none;
  border-radius: 6px;
  background: rgba(255,255,255,0.06);
  color: var(--text-secondary);
  cursor: pointer;
  font-size: 12px;
  transition: all 0.15s ease;
}

.ml-toast-btn:hover {
  background: rgba(255,255,255,0.12);
  color: var(--text-primary);
}

.ml-toast-enter-active { transition: all 0.3s ease; }
.ml-toast-leave-active { transition: all 0.2s ease; }
.ml-toast-enter-from { opacity: 0; transform: translateY(-8px); }
.ml-toast-leave-to { opacity: 0; transform: translateY(-8px); }

/* Uninstall Button */
.btn-uninstall {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 14px;
  border: 1px solid rgba(230, 63, 50, 0.3);
  border-radius: var(--radius-md);
  background: rgba(230, 63, 50, 0.08);
  color: var(--lenovo-red);
  font-size: 13px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
  white-space: nowrap;
}

.btn-uninstall:hover:not(:disabled) {
  background: var(--lenovo-red);
  border-color: var(--lenovo-red);
  color: white;
}

.btn-uninstall:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}
</style>

<style scoped>
.sidebar-footer-icons {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.icon-btn {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-secondary);
  transition: var(--transition);
}

.icon-btn:hover {
  background: var(--lenovo-red);
  color: white;
}

.theme-toggle {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  color: var(--text-secondary);
  transition: var(--transition);
}

.theme-toggle:hover {
  background: var(--lenovo-red);
  color: white;
}
</style>
