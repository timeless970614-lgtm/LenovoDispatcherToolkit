<template>
  <div class="settings-page">

    <!-- Appearance Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
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
            <div class="toggle-group">
              <button :class="['toggle-btn', theme === 'dark' ? 'active' : '']" @click="setTheme('dark')">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z"/></svg>
                {{ t.dark }}
              </button>
              <button :class="['toggle-btn', theme === 'light' ? 'active' : '']" @click="setTheme('light')">
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="5"/><line x1="12" y1="1" x2="12" y2="3"/><line x1="12" y1="21" x2="12" y2="23"/><line x1="4.22" y1="4.22" x2="5.64" y2="5.64"/><line x1="18.36" y1="18.36" x2="19.78" y2="19.78"/><line x1="1" y1="12" x2="3" y2="12"/><line x1="21" y1="12" x2="23" y2="12"/><line x1="4.22" y1="19.78" x2="5.64" y2="18.36"/><line x1="18.36" y1="5.64" x2="19.78" y2="4.22"/></svg>
                {{ t.light }}
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
            <select class="select-input" v-model="refreshInterval">
              <option value="1000">{{ t.sec1 }}</option>
              <option value="2000">{{ t.sec2 }}</option>
              <option value="5000">{{ t.sec5 }}</option>
              <option value="10000">{{ t.sec10 }}</option>
              <option value="30000">{{ t.sec30 }}</option>
            </select>
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
    refreshIntervalDesc: 'How often to refresh system data',
    sec1: '1 second',
    sec2: '2 seconds',
    sec5: '5 seconds',
    sec10: '10 seconds',
    sec30: '30 seconds',
    autoStart: 'Start with Windows',
    autoStartDesc: 'Launch toolkit automatically on startup',
    minimizeToTray: 'Minimize to Tray',
    minimizeToTrayDesc: 'Keep running in system tray when closed',
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
    refreshIntervalDesc: '系统数据的刷新频率',
    sec1: '1 秒',
    sec2: '2 秒',
    sec5: '5 秒',
    sec10: '10 秒',
    sec30: '30 秒',
    autoStart: '开机自启',
    autoStartDesc: '系统启动时自动运行工具箱',
    minimizeToTray: '最小化到托盘',
    minimizeToTrayDesc: '关闭窗口时保持在系统托盘运行',
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
    }
  },
  emits: ['update:theme', 'update:lang'],
  data() {
    return {
      refreshInterval: '1000',
      autoStart: false,
      minimizeToTray: true
    }
  },
  computed: {
    t() {
      return i18n[this.lang] || i18n.en
    }
  },
  methods: {
    setTheme(newTheme) {
      this.$emit('update:theme', newTheme)
    },
    setLang(newLang) {
      this.$emit('update:lang', newLang)
    }
  }
}
</script>

<style scoped>
.settings-list {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 20px;
  border-radius: var(--radius-md);
  transition: var(--transition);
}

.setting-row:hover {
  background: var(--bg-tertiary);
}

.setting-info {
  flex: 1;
}

.setting-name {
  font-size: 15px;
  font-weight: 500;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.setting-desc {
  font-size: 13px;
  color: var(--text-tertiary);
}

.setting-control {
  flex-shrink: 0;
  margin-left: 24px;
}

.toggle-group {
  display: flex;
  background: var(--bg-tertiary);
  border-radius: var(--radius-md);
  padding: 4px;
  gap: 4px;
}

.toggle-btn {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 16px;
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

.select-input {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  color: var(--text-primary);
  padding: 10px 16px;
  font-size: 14px;
  font-family: inherit;
  cursor: pointer;
  outline: none;
  transition: var(--transition);
  min-width: 160px;
}

.select-input:hover {
  border-color: var(--border-light);
}

.select-input:focus {
  border-color: var(--lenovo-red);
}

.switch {
  width: 52px;
  height: 28px;
  background: var(--bg-tertiary);
  border: 2px solid var(--border-color);
  border-radius: 14px;
  cursor: pointer;
  transition: var(--transition);
  position: relative;
}

.switch.on {
  background: var(--lenovo-red);
  border-color: var(--lenovo-red);
}

.switch-thumb {
  width: 20px;
  height: 20px;
  background: white;
  border-radius: 50%;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: var(--transition);
  box-shadow: 0 2px 4px rgba(0,0,0,0.3);
}

.switch.on .switch-thumb {
  left: 26px;
}
</style>
