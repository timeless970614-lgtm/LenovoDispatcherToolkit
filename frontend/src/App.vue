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
        <div class="theme-toggle" @click="toggleTheme" :title="theme === 'dark' ? 'Switch to Light Mode' : 'Switch to Dark Mode'">
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
          <svg v-else width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/>
          </svg>
        </div>
        <div class="version-tag">v1.0.0</div>
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
          <button class="btn-ml-capture" :class="{ capturing: mlCapturing }" @click="toggleMLCapture" :title="mlCapturing ? 'Stop ML Log Capture' : 'Start ML Log Capture'">
            <svg v-if="!mlCapturing" width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
              <circle cx="12" cy="12" r="6"/>
              <rect x="10" y="2" width="4" height="6" rx="1"/>
            </svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none" class="rec-dot">
              <circle cx="12" cy="12" r="8"/>
            </svg>
            <span>{{ mlCapturing ? 'Stop ML' : 'ML Log' }}</span>
            <span v-if="mlCapturing && mlEventCount > 0" class="ml-count">{{ mlEventCount }}</span>
          </button>

          <div class="service-status-pill" v-if="serviceRunning !== null">
            <span :class="['svc-dot', serviceRunning ? 'svc-dot-running' : 'svc-dot-stopped']"></span>
            <span class="svc-label">{{ serviceRunning ? 'Running' : 'Stopped' }}</span>
          </div>
          <div class="current-mode-pill" v-if="currentMode || pinnedMode">
            <span class="mode-dot"></span>
            <span class="mode-label">Current Mode</span>
            <span class="mode-value">{{ pinnedMode || currentMode }}</span>
            <span v-if="pinnedMode" class="pin-indicator">📌</span>
          </div>
          <div class="header-time">{{ currentTime }}</div>
        </div>
      </header>
      
      <div class="content-body">
        <Dashboard v-if="currentPage === 'dashboard'" :theme="theme" />
        <PPMDriver v-else-if="currentPage === 'ppm'" :theme="theme" />

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
import PPMDriver from './pages/PPMDriver.vue'
import FunctionCheck from './pages/FunctionCheck.vue'
import ModeCheck from './pages/ModeCheck.vue'
import AIAnalysis from './pages/AIAnalysis.vue'
import Settings from './pages/Settings.vue'
import About from './pages/About.vue'
import { StartMLScenarioCapture, StopMLScenarioCapture, GetMLLogStatus } from '../wailsjs/go/main/App'

export default {
  name: 'App',
  components: {
    Dashboard,
    PPMDriver,
    FunctionCheck,
    ModeCheck,
    AIAnalysis,
    Settings,
    About
  },
  data() {
    return {
      currentPage: 'dashboard',
      theme: 'dark',
      lang: 'en',
      currentTime: '',
      currentMode: '',
      pinnedMode: '',
      serviceRunning: null,
      mlCapturing: false,
      mlEventCount: 0,
      mlInterval: null,
      timeInterval: null,
      modeInterval: null,
      mainNavItems: [
        { 
          id: 'dashboard', 
          label: 'Dashboard', 
          subtitle: 'System overview & status',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="3" y="3" width="7" height="7" rx="1"/><rect x="14" y="3" width="7" height="7" rx="1"/><rect x="3" y="14" width="7" height="7" rx="1"/><rect x="14" y="14" width="7" height="7" rx="1"/></svg>'
        },
        { 
          id: 'ppm', 
          label: 'PPM Driver', 
          subtitle: 'Processor power management',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/><circle cx="12" cy="12" r="4"/></svg>'
        },

        { 
          id: 'funccheck', 
          label: 'Function Check', 
          subtitle: 'GPU & system diagnostics',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M9 11l3 3L22 4M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/></svg>'
        },
        { 
          id: 'modecheck', 
          label: 'Mode Check', 
          subtitle: 'Dispatcher status & features',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/></svg>'
        },
        {
          id: 'aianalysis',
          label: 'AI Analysis',
          subtitle: 'Log analysis & diagnostics',
          icon: '<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/></svg>'
        }
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
          if (info && info.CurrentMode) {
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
            const matched = Object.keys(modeMap).find(k => raw.startsWith(k))
            this.currentMode = matched ? modeMap[matched] : raw
          }
          const status = await window.go.main.App.GetServiceStatus()
          this.serviceRunning = (status === 'Running')
        }
      } catch (e) {
        // silent
      }
    },
    toggleTheme() {
      this.theme = this.theme === 'dark' ? 'light' : 'dark'
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
          if (result && result.outputFile) {
            console.log('ML Log saved to:', result.outputFile)
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
</style>

<style scoped>
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
  margin-bottom: 12px;
}

.theme-toggle:hover {
  background: var(--lenovo-red);
  color: white;
}
</style>
