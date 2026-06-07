<template>
  <div class="settings-page">

    <!-- Appearance Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <circle cx="12" cy="12" r="5"/>
            <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
          </svg>
          {{ t.appearance }}
        </span>
      </div>

      <div class="settings-list">
        <!-- Theme -->
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-name">{{ t.theme }}</div>
            <div class="setting-desc">{{ t.themeDesc }}</div>
          </div>
          <div class="setting-control">
            <div class="theme-grid">
              <button :class="['theme-btn', theme === 'dark' ? 'active' : '']" @click="setTheme('dark')">
                <span class="theme-preview dark"></span>
                <span class="theme-name">{{ t.dark }}</span>
              </button>
              <button :class="['theme-btn', theme === 'light' ? 'active' : '']" @click="setTheme('light')">
                <span class="theme-preview light"></span>
                <span class="theme-name">{{ t.light }}</span>
              </button>
            </div>
          </div>
        </div>

        <!-- Language -->
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-name">{{ t.language }}</div>
            <div class="setting-desc">{{ t.languageDesc }}</div>
          </div>
          <div class="setting-control">
            <div class="toggle-group">
              <button :class="['toggle-btn', lang === 'en' ? 'active' : '']" @click="setLang('en')">
                <span class="lang-flag">🇺🇸</span>
                English
              </button>
              <button :class="['toggle-btn', lang === 'zh' ? 'active' : '']" @click="setLang('zh')">
                <span class="lang-flag">🇨🇳</span>
                中文
              </button>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- General Settings Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
          {{ t.general }}
        </span>
      </div>

      <div class="settings-list">
        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-name">{{ t.refreshInterval }}</div>
            <div class="setting-desc">{{ t.refreshIntervalDesc }}</div>
          </div>
          <div class="setting-control">
            <div class="interval-group">
              <button
                v-for="opt in intervalOptions"
                :key="opt.value"
                :class="['interval-btn', String(pollInterval) === opt.value ? 'active' : '']"
                @click="setRefreshInterval(opt.value)"
              >{{ opt.label }}</button>
            </div>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-name">{{ t.autoStart }}</div>
            <div class="setting-desc">{{ t.autoStartDesc }}</div>
          </div>
          <div class="setting-control">
            <div :class="['switch', autoStart ? 'on' : '']" @click="autoStart = !autoStart">
              <div class="switch-thumb"></div>
            </div>
          </div>
        </div>

        <div class="setting-row">
          <div class="setting-info">
            <div class="setting-name">{{ t.minimizeToTray }}</div>
            <div class="setting-desc">{{ t.minimizeToTrayDesc }}</div>
          </div>
          <div class="setting-control">
            <div :class="['switch', minimizeToTray ? 'on' : '']" @click="minimizeToTray = !minimizeToTray">
              <div class="switch-thumb"></div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- About Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <circle cx="12" cy="12" r="10"/>
            <path d="M12 16v-4M12 8h.01"/>
          </svg>
          {{ t.aboutTitle }}
        </span>
      </div>
      <div class="settings-list">
        <div class="setting-row about-row">
          <div class="about-logo">
            <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="var(--lenovo-red)" stroke-width="1.5">
              <rect x="2" y="3" width="20" height="18" rx="2"/>
              <path d="M8 21h8M12 3v18"/>
            </svg>
          </div>
          <div class="about-info">
            <div class="about-name">Lenovo Dispatcher Toolkit</div>
            <div class="about-version">v1.0.20</div>
          </div>
        </div>
      </div>
    </div>

  </div>
</template>

<script>
const i18n = {
  en: {
    appearance: 'Appearance',
    theme: 'Theme',
    themeDesc: 'Choose your preferred color scheme',
    dark: 'Dark',
    light: 'Light',
    language: 'Language',
    languageDesc: 'Switch between English and Chinese',
    general: 'General Settings',
    refreshInterval: 'Auto Refresh Interval',
    refreshIntervalDesc: 'How often to refresh system data across all pages',
    autoStart: 'Start with Windows',
    autoStartDesc: 'Launch toolkit automatically on startup',
    minimizeToTray: 'Minimize to Tray',
    minimizeToTrayDesc: 'Keep running in system tray when closed',
    aboutTitle: 'About',
  },
  zh: {
    appearance: '外观',
    theme: '主题',
    themeDesc: '选择你偏好的配色方案',
    dark: '深色',
    light: '浅色',
    language: '语言',
    languageDesc: '切换中英文界面',
    general: '通用设置',
    refreshInterval: '自动刷新间隔',
    refreshIntervalDesc: '所有页面的系统数据刷新频率',
    autoStart: '开机自启',
    autoStartDesc: '系统启动时自动运行工具箱',
    minimizeToTray: '最小化到托盘',
    minimizeToTrayDesc: '关闭窗口时保持在系统托盘运行',
    aboutTitle: '关于',
  }
}

export default {
  name: 'Settings',
  props: {
    theme: {
      type: String,
      default: 'dark'
    },
    lang: {
      type: String,
      default: 'en'
    },
    pollInterval: {
      type: Number,
      default: 5000
    }
  },
  emits: ['update:theme', 'update:lang', 'update:poll-interval'],
  data() {
    return {
      autoStart: false,
      minimizeToTray: true
    }
  },
  computed: {
    t() {
      return i18n[this.lang] || i18n.en
    },
    intervalOptions() {
      return [
        { value: '2000', label: this.lang === 'zh' ? '2秒' : '2s' },
        { value: '5000', label: this.lang === 'zh' ? '5秒' : '5s' },
        { value: '10000', label: this.lang === 'zh' ? '10秒' : '10s' },
        { value: '30000', label: this.lang === 'zh' ? '30秒' : '30s' },
      ]
    }
  },
  methods: {
    setTheme(newTheme) {
      this.$emit('update:theme', newTheme)
    },
    setLang(newLang) {
      this.$emit('update:lang', newLang)
    },
    setRefreshInterval(ms) {
      this.$emit('update:poll-interval', ms)
    }
  }
}
</script>

<style scoped>
.settings-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px;
  border-radius: var(--radius-md);
  transition: var(--transition);
}

.setting-row:hover {
  background: var(--bg-tertiary);
}

.setting-info {
  flex: 1;
  min-width: 0;
}

.setting-name {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 3px;
}

.setting-desc {
  font-size: 12px;
  color: var(--text-tertiary);
  line-height: 1.4;
}

.setting-control {
  flex-shrink: 0;
  margin-left: 24px;
}

.toggle-group {
  display: flex;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 3px;
  gap: 2px;
}

.toggle-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 7px 14px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  background: transparent;
  color: var(--text-secondary);
  transition: var(--transition);
  font-family: inherit;
}

.toggle-btn.active {
  background: var(--lenovo-red);
  color: white;
  box-shadow: 0 2px 8px rgba(230, 63, 50, 0.3);
}

.lang-flag {
  font-size: 14px;
  line-height: 1;
}

/* Interval button group */
.interval-group {
  display: flex;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 3px;
  gap: 2px;
}

.interval-btn {
  padding: 7px 16px;
  border: none;
  border-radius: var(--radius-sm);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  background: transparent;
  color: var(--text-secondary);
  transition: var(--transition);
  font-family: inherit;
}

.interval-btn:hover {
  color: var(--text-primary);
}

.interval-btn.active {
  background: var(--lenovo-red);
  color: white;
  box-shadow: 0 2px 8px rgba(230, 63, 50, 0.3);
}

.switch {
  width: 48px;
  height: 26px;
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: 13px;
  cursor: pointer;
  transition: var(--transition);
  position: relative;
}

.switch.on {
  background: var(--lenovo-red);
  border-color: var(--lenovo-red);
}

.switch-thumb {
  width: 18px;
  height: 18px;
  background: white;
  border-radius: 50%;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: var(--transition);
  box-shadow: 0 2px 4px rgba(0,0,0,0.3);
}

.switch.on .switch-thumb {
  left: 24px;
}

/* Theme Grid */
.theme-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
}

.theme-btn {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 6px;
  padding: 10px 16px;
  border: 2px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-tertiary);
  cursor: pointer;
  transition: var(--transition);
}

.theme-btn:hover {
  border-color: var(--border-light);
}

.theme-btn.active {
  border-color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.1);
}

.theme-preview {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  border: 2px solid var(--border-color);
}

.theme-preview.dark {
  background: linear-gradient(135deg, #0D0D0D 0%, #232323 100%);
}

.theme-preview.light {
  background: linear-gradient(135deg, #F5F5F5 0%, #FFFFFF 100%);
}

.theme-name {
  font-size: 12px;
  font-weight: 500;
  color: var(--text-secondary);
}

.theme-btn.active .theme-name {
  color: var(--lenovo-red);
}

/* About row */
.about-row {
  gap: 16px;
  cursor: default;
}

.about-row:hover {
  background: transparent;
}

.about-logo {
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}

.about-info {
  flex: 1;
}

.about-name {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.about-version {
  font-size: 12px;
  color: var(--text-tertiary);
  margin-top: 2px;
}
</style>
