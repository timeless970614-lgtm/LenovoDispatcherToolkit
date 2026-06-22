<template>
  <div class="test-func-page">

    <!-- Header Info -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polygon points="5 3 19 12 5 21 5 3"/>
          </svg>
          Test Function
        </h3>
        <button class="btn btn-secondary btn-sm" @click="runAllTests" :disabled="running">
          <span v-if="running" class="spinner-small"></span>
          <span v-else>▶ Run All</span>
        </button>
      </div>
      <p class="test-desc">对核心功能模块进行诊断测试，验证各子系统是否正常工作。</p>
    </div>

    <!-- App Launch Speed Test -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
          </svg>
          应用启动速度
        </h3>
        <span v-if="launch.done" :class="['test-badge', launch.fast ? 'pass' : 'warn']">
          {{ launch.fast ? 'PASS' : 'OK' }}
        </span>
      </div>
      <div class="test-body">
        <div class="test-info-grid">
          <div class="test-info-item">
            <span class="test-info-label">Build 版本</span>
            <span class="test-info-value mono">v1.0.21</span>
          </div>
          <div class="test-info-item">
            <span class="test-info-label">构建日期</span>
            <span class="test-info-value mono">2026-06-07</span>
          </div>
          <div class="test-info-item">
            <span class="test-info-label">Framework</span>
            <span class="test-info-value mono">Wails v2.12.0</span>
          </div>
          <div class="test-info-item">
            <span class="test-info-label">启动耗时</span>
            <span :class="['test-info-value', 'mono', 'highlight', launch.done ? '' : 'loading-text']">
              {{ launch.elapsed ? launch.elapsed + ' ms' : '—' }}
            </span>
          </div>
        </div>
        <div class="test-actions">
          <button class="btn btn-primary btn-sm" @click="testLaunchSpeed" :disabled="launch.running">
            <span v-if="launch.running" class="spinner-small"></span>
            <span v-else>测试启动速度</span>
          </button>
          <span v-if="launch.result" :class="['result-text', launch.result.success ? 'success' : 'error']">
            {{ launch.result.message }}
          </span>
        </div>
      </div>
    </div>

    <!-- Brightness Test -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="5"/>
            <path d="M12 1v2M12 21v2M4.22 4.22l1.42 1.42M18.36 18.36l1.42 1.42M1 12h2M21 12h2M4.22 19.78l1.42-1.42M18.36 5.64l1.42-1.42"/>
          </svg>
          屏幕亮度测试
        </h3>
        <span v-if="brightness.done" :class="['test-badge', brightness.result?.writable ? 'pass' : (brightness.result?.readable ? 'warn' : 'fail')]">
          {{ brightness.result?.writable ? 'PASS' : (brightness.result?.readable ? 'READ-ONLY' : 'FAIL') }}
        </span>
      </div>
      <div class="test-body">
        <div v-if="brightness.running" class="test-loading">
          <div class="spinner-small"></div>
          <span>正在读写亮度...</span>
        </div>
        <div v-else-if="brightness.result" class="test-result-block">
          <div class="result-row">
            <span class="result-label">当前亮度</span>
            <div class="brightness-bar-wrap">
              <div class="brightness-bar" :style="{ width: brightness.result.currentBrightness + '%' }"></div>
            </div>
            <span class="result-value mono">{{ brightness.result.currentBrightness }}%</span>
          </div>
          <div class="result-row" v-if="brightness.result.testBrightness">
            <span class="result-label">测试设置</span>
            <div class="brightness-bar-wrap">
              <div class="brightness-bar test-bar" :style="{ width: brightness.result.testBrightness + '%' }"></div>
            </div>
            <span class="result-value mono">{{ brightness.result.testBrightness }}%</span>
          </div>
          <div class="result-chips">
            <span :class="['chip', brightness.result.readable ? 'chip-ok' : 'chip-fail']">
              {{ brightness.result.readable ? '✓ 可读' : '✗ 不可读' }}
            </span>
            <span :class="['chip', brightness.result.writable ? 'chip-ok' : 'chip-fail']">
              {{ brightness.result.writable ? '✓ 可写' : '✗ 不可写' }}
            </span>
            <span v-if="brightness.result.error" class="chip chip-fail">
              {{ brightness.result.error }}
            </span>
          </div>
        </div>
        <div v-else class="test-hint">
          <p>测试屏幕亮度读取和设置功能（仅支持笔记本内置屏幕）</p>
        </div>
        <div class="test-actions">
          <button class="btn btn-primary btn-sm" @click="testBrightness" :disabled="brightness.running">
            <span v-if="brightness.running" class="spinner-small"></span>
            <span v-else>运行测试</span>
          </button>
        </div>
      </div>
    </div>

    <!-- FN+Q Test -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="7" width="20" height="14" rx="2"/>
            <path d="M16 7V5a2 2 0 0 0-2-2h-4a2 2 0 0 0-2 2v2"/>
          </svg>
          FN+Q 热键测试
        </h3>
        <span v-if="fnq.done" :class="['test-badge', fnq.result?.supported ? 'pass' : 'warn']">
          {{ fnq.result?.supported ? 'PASS' : 'N/A' }}
        </span>
      </div>
      <div class="test-body">
        <div v-if="fnq.running" class="test-loading">
          <div class="spinner-small"></div>
          <span>正在检测 FN+Q 功能...</span>
        </div>
        <div v-else-if="fnq.result" class="test-result-block">
          <div class="test-info-grid">
            <div class="test-info-item">
              <span class="test-info-label">当前模式</span>
              <span class="test-info-value mono">{{ fnq.result.currentMode || 'N/A' }}</span>
            </div>
            <div class="test-info-item">
              <span class="test-info-label">FN+Q 支持</span>
              <span :class="['test-info-value', 'mono', fnq.result.supported ? 'text-ok' : 'text-muted']">
                {{ fnq.result.supported ? '支持' : '不支持' }}
              </span>
            </div>
            <div class="test-info-item">
              <span class="test-info-label">热键可用</span>
              <span :class="['test-info-value', 'mono', fnq.result.hotkeyAvailable ? 'text-ok' : 'text-muted']">
                {{ fnq.result.hotkeyAvailable ? '可用' : '不可用' }}
              </span>
            </div>
            <div class="test-info-item">
              <span class="test-info-label">Dispatcher Func</span>
              <span class="test-info-value mono">{{ fnq.result.dispatcherFunction || 'N/A' }}</span>
            </div>
            <div class="test-info-item" v-if="fnq.result.hotkeyStatus">
              <span class="test-info-label">热键状态</span>
              <span class="test-info-value mono">{{ fnq.result.hotkeyStatus }}</span>
            </div>
            <div class="test-info-item">
              <span class="test-info-label">DCC 功能</span>
              <span :class="['test-info-value', 'mono', fnq.result.dccCapability ? 'text-ok' : 'text-muted']">
                {{ fnq.result.dccCapability ? '支持' : '不支持' }}
              </span>
            </div>
            <div class="test-info-item">
              <span class="test-info-label">GEEK 功能</span>
              <span :class="['test-info-value', 'mono', fnq.result.geekCapability ? 'text-ok' : 'text-muted']">
                {{ fnq.result.geekCapability ? '支持' : '不支持' }}
              </span>
            </div>
          </div>
          <div v-if="fnq.result.error" class="result-error">
            {{ fnq.result.error }}
          </div>
        </div>
        <div v-else class="test-hint">
          <p>检测 FN+Q 快捷键功能支持状态及热键可用性</p>
        </div>
        <div class="test-actions">
          <button class="btn btn-primary btn-sm" @click="testFNQ" :disabled="fnq.running">
            <span v-if="fnq.running" class="spinner-small"></span>
            <span v-else>运行测试</span>
          </button>
        </div>
      </div>
    </div>

    <!-- Auto Launch (open_all_files.bat replacement) -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
          </svg>
          自动化打开应用
        </h3>
        <div class="header-actions">
          <button class="btn btn-secondary btn-sm" @click="toggleSelectAll" :disabled="autoLaunch.running">
            {{ allSelected ? '取消全选' : '全选' }}
          </button>
          <button class="btn btn-primary btn-sm" @click="launchSelected" :disabled="autoLaunch.running">
            <span v-if="autoLaunch.running" class="spinner-small"></span>
            <span v-else>▶ 启动选中</span>
          </button>
        </div>
      </div>
      <div class="test-body">
        <!-- Browser section -->
        <div v-if="autoLaunch.browserItems.length" class="launch-section">
          <div class="section-label">🌐 浏览器标签页</div>
          <div class="launch-grid">
            <div v-for="item in autoLaunch.browserItems" :key="item.id"
              :class="['launch-item', item.enabled ? 'enabled' : 'disabled']"
              @click="toggleItem(item)">
              <div class="launch-check">
                <span :class="['check-box', item.enabled ? 'checked' : '']">✓</span>
              </div>
              <div class="launch-info">
                <span class="launch-name">{{ item.name }}</span>
                <span class="launch-desc">{{ item.description }}</span>
              </div>
              <div class="launch-wait">{{ item.waitSec }}s</div>
            </div>
          </div>
        </div>

        <!-- Protocol/App section -->
        <div v-if="autoLaunch.appItems.length" class="launch-section">
          <div class="section-label">📱 应用与协议</div>
          <div class="launch-grid">
            <div v-for="item in autoLaunch.appItems" :key="item.id"
              :class="['launch-item', item.enabled ? 'enabled' : 'disabled']"
              @click="toggleItem(item)">
              <div class="launch-check">
                <span :class="['check-box', item.enabled ? 'checked' : '']">✓</span>
              </div>
              <div class="launch-info">
                <span class="launch-name">{{ item.name }}</span>
                <span class="launch-desc">{{ item.description }}</span>
              </div>
              <div class="launch-wait">{{ item.waitSec }}s</div>
            </div>
          </div>
        </div>

        <!-- Folder section -->
        <div v-if="autoLaunch.folderItems.length" class="launch-section">
          <div class="section-label">
            📂 文件夹文件
            <span class="section-hint">（自动扫描目录下的文件）</span>
          </div>
          <div class="launch-grid">
            <div v-for="item in autoLaunch.folderItems" :key="item.id"
              :class="['launch-item', item.enabled ? 'enabled' : 'disabled']"
              @click="toggleItem(item)">
              <div class="launch-check">
                <span :class="['check-box', item.enabled ? 'checked' : '']">✓</span>
              </div>
              <div class="launch-info">
                <span class="launch-name">{{ item.name }}</span>
                <span class="launch-desc">{{ item.description }}</span>
              </div>
              <div class="launch-wait">{{ item.waitSec }}s</div>
            </div>
          </div>
        </div>

        <!-- Running progress -->
        <div v-if="autoLaunch.running" class="test-loading">
          <div class="spinner-small"></div>
          <span>正在启动应用... ({{ autoLaunch.doneCount }}/{{ autoLaunch.totalCount }})</span>
        </div>

        <!-- Results -->
        <div v-if="autoLaunch.results.length" class="test-result-block" style="margin-top: 10px;">
          <div class="result-chips">
            <span v-for="r in autoLaunch.results" :key="r.itemId" :class="['chip', r.success ? 'chip-ok' : 'chip-fail']">
              {{ r.name }}: {{ r.success ? '✓ 已启动' : '✗ ' + (r.error || '失败') }}
            </span>
          </div>
        </div>

        <div v-if="!autoLaunch.running && !autoLaunch.results.length" class="test-hint">
          <p>一键打开 Edge 浏览器标签页、Outlook、爱奇艺客户端及文件夹文件（替代 open_all_files.bat）</p>
        </div>
      </div>
    </div>

    <!-- Mode Switch Test -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>
          </svg>
          模式切换测试
        </h3>
        <span v-if="mode.done" :class="['test-badge', mode.anyVerified ? 'pass' : 'warn']">
          {{ mode.anyVerified ? 'PASS' : 'N/A' }}
        </span>
      </div>
      <div class="test-body">
        <div v-if="mode.running" class="test-loading">
          <div class="spinner"></div>
          <span>正在测试模式切换...</span>
        </div>
        <div v-else-if="mode.result" class="test-result-block">
          <div class="result-row">
            <span class="result-label">当前模式</span>
            <span class="result-value mono">{{ mode.result.currentMode || 'N/A' }}</span>
          </div>
          <div v-if="mode.result.attempts && mode.result.attempts.length > 0" class="mode-attempts">
            <div
              v-for="(attempt, idx) in mode.result.attempts"
              :key="idx"
              :class="['mode-attempt-row', attempt.skipped ? 'skipped' : (attempt.verified ? 'verified' : 'failed')]"
            >
              <span class="attempt-mode mono">{{ attempt.mode }}</span>
              <span class="attempt-status">
                <span v-if="attempt.skipped" class="status-text skipped">跳过 ({{ attempt.reason }})</span>
                <span v-else-if="attempt.verified" class="status-text verified">✓ 成功</span>
                <span v-else class="status-text failed">✗ {{ attempt.error || attempt.note }}</span>
              </span>
              <span v-if="attempt.switchTimeMs" class="attempt-time mono">
                {{ attempt.switchTimeMs }} ms
              </span>
            </div>
          </div>
          <div v-if="mode.result.error" class="result-error">
            {{ mode.result.error }}
          </div>
        </div>
        <div v-else class="test-hint">
          <p>测试热管理模式切换功能（BSM / STD / AQM）</p>
        </div>
        <div class="test-actions">
          <button class="btn btn-primary btn-sm" @click="testModeSwitch" :disabled="mode.running">
            <span v-if="mode.running" class="spinner-small"></span>
            <span v-else>运行测试</span>
          </button>
        </div>
      </div>
    </div>

  </div>
</template>

<script>
import {
  TestBrightness,
  TestFNQ,
  TestModeSwitch,
  GetAutoLaunchItems,
  GetFolderFiles,
  BatchLaunchAutoLaunchItems,
  LaunchAllEnabledItems,
  ToggleAutoLaunchItem,
} from '../../wailsjs/go/main/App'

export default {
  name: 'TestFunction',
  data() {
    return {
      launch: { running: false, done: false, elapsed: null, fast: false, result: null },
      brightness: { running: false, done: false, result: null },
      fnq: { running: false, done: false, result: null },
      mode: { running: false, done: false, result: null, anyVerified: false },
      autoLaunch: {
        browserItems: [],
        appItems: [],
        folderItems: [],
        running: false,
        results: [],
        doneCount: 0,
        totalCount: 0,
      },
    }
  },
  computed: {
    allItems() {
      return [...this.autoLaunch.browserItems, ...this.autoLaunch.appItems, ...this.autoLaunch.folderItems]
    },
    allSelected() {
      return this.allItems.length > 0 && this.allItems.every(i => i.enabled)
    },
    running() {
      return this.launch.running || this.brightness.running || this.fnq.running || this.mode.running
    },
  },
  methods: {
    async testLaunchSpeed() {
      this.launch.running = true
      this.launch.done = false
      this.launch.result = null
      const start = performance.now()
      try {
        const { GetSystemInfo } = await import('../../wailsjs/go/main/App')
        await GetSystemInfo()
        const elapsed = Math.round(performance.now() - start)
        this.launch.elapsed = elapsed
        this.launch.fast = elapsed < 500
        this.launch.result = { success: true, message: `系统信息调用耗时 ${elapsed}ms` }
      } catch (e) {
        this.launch.result = { success: false, message: '系统信息调用失败: ' + String(e) }
      }
      this.launch.running = false
      this.launch.done = true
    },

    async testBrightness() {
      this.brightness.running = true
      this.brightness.done = false
      try {
        const result = await TestBrightness()
        this.brightness.result = result
      } catch (e) {
        this.brightness.result = { error: String(e), readable: false, writable: false }
      }
      this.brightness.running = false
      this.brightness.done = true
    },

    async testFNQ() {
      this.fnq.running = true
      this.fnq.done = false
      try {
        const result = await TestFNQ()
        this.fnq.result = result
      } catch (e) {
        this.fnq.result = { error: String(e) }
      }
      this.fnq.running = false
      this.fnq.done = true
    },

    async testModeSwitch() {
      this.mode.running = true
      this.mode.done = false
      this.mode.anyVerified = false
      try {
        const result = await TestModeSwitch()
        this.mode.result = result
        if (result.attempts) {
          this.mode.anyVerified = result.attempts.some(a => a.verified)
        }
      } catch (e) {
        this.mode.result = { error: String(e) }
      }
      this.mode.running = false
      this.mode.done = true
    },

    async runAllTests() {
      await this.testLaunchSpeed()
      await this.testBrightness()
      await this.testFNQ()
      await this.testModeSwitch()
    },

    // === Auto Launch methods ===
    async loadAutoLaunchItems() {
      try {
        const [items, folderCfg] = await GetAutoLaunchItems()
        this.autoLaunch.browserItems = items.filter(i => i.category === 'browser').map(i => ({...i}))
        this.autoLaunch.appItems = items.filter(i => i.category === 'protocol' || i.category === 'app').map(i => ({...i}))

        // Load folder files
        try {
          const folderItems = await GetFolderFiles()
          this.autoLaunch.folderItems = folderItems.map(i => ({...i}))
        } catch (e) {
          console.warn('Folder scan skipped:', e)
        }
      } catch (e) {
        console.error('Failed to load auto launch items:', e)
      }
    },

    toggleItem(item) {
      item.enabled = !item.enabled
      ToggleAutoLaunchItem(item.id, item.enabled).catch(() => {})
    },

    toggleSelectAll() {
      const target = !this.allSelected
      this.allItems.forEach(item => {
        item.enabled = target
        ToggleAutoLaunchItem(item.id, target).catch(() => {})
      })
    },

    async launchSelected() {
      const selected = this.allItems.filter(i => i.enabled).map(i => i.id)
      if (selected.length === 0) return

      this.autoLaunch.running = true
      this.autoLaunch.results = []
      this.autoLaunch.doneCount = 0
      this.autoLaunch.totalCount = selected.length

      try {
        const results = await BatchLaunchAutoLaunchItems(selected)
        this.autoLaunch.results = results || []
        this.autoLaunch.doneCount = (results || []).length
      } catch (e) {
        this.autoLaunch.results = [{ itemId: 'error', name: '错误', error: String(e) }]
      }

      this.autoLaunch.running = false
    },
  },
  mounted() {
    this.loadAutoLaunchItems()
  }
}
</script>

<style scoped>
.test-func-page {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.test-desc {
  margin: 0;
  color: var(--text-muted);
  font-size: 13px;
}

.test-body {
  margin-top: 12px;
}

/* Test Info Grid */
.test-info-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 8px;
  margin-bottom: 12px;
}

.test-info-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 8px 12px;
  background: var(--bg-elevated);
  border-radius: 6px;
  border: 1px solid var(--border);
}

.test-info-label {
  font-size: 11px;
  color: var(--text-muted);
  text-transform: uppercase;
  letter-spacing: 0.05em;
}

.test-info-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.test-info-value.mono {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  font-size: 13px;
}

.test-info-value.highlight {
  color: var(--accent);
}

.test-info-value.text-ok {
  color: #4ade80;
}

.test-info-value.text-muted {
  color: var(--text-muted);
}

/* Loading state */
.test-loading {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 16px;
  color: var(--text-muted);
  font-size: 13px;
}

.test-loading .spinner {
  width: 20px;
  height: 20px;
  border-width: 2px;
}

/* Hint */
.test-hint {
  padding: 12px 16px;
  color: var(--text-muted);
  font-size: 13px;
  background: var(--bg-elevated);
  border-radius: 6px;
}

/* Result */
.test-result-block {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  background: var(--bg-elevated);
  border-radius: 6px;
}

.result-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.result-label {
  font-size: 12px;
  color: var(--text-muted);
  width: 80px;
  flex-shrink: 0;
}

.result-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.result-value.mono {
  font-family: 'Cascadia Code', 'Fira Code', monospace;
}

.brightness-bar-wrap {
  flex: 1;
  height: 8px;
  background: var(--bg);
  border-radius: 4px;
  overflow: hidden;
}

.brightness-bar {
  height: 100%;
  background: linear-gradient(90deg, #60a5fa, #3b82f6);
  border-radius: 4px;
  transition: width 0.5s ease;
}

.brightness-bar.test-bar {
  background: linear-gradient(90deg, #facc15, #f59e0b);
}

/* Chips */
.result-chips {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.chip {
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}

.chip-ok {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.3);
}

.chip-fail {
  background: rgba(248, 113, 113, 0.15);
  color: #f87171;
  border: 1px solid rgba(248, 113, 113, 0.3);
}

/* Result error */
.result-error {
  padding: 8px 12px;
  background: rgba(248, 113, 113, 0.1);
  border: 1px solid rgba(248, 113, 113, 0.3);
  border-radius: 6px;
  color: #f87171;
  font-size: 12px;
}

/* Mode attempts */
.mode-attempts {
  display: flex;
  flex-direction: column;
  gap: 4px;
  margin-top: 4px;
}

.mode-attempt-row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 12px;
  border-radius: 6px;
  font-size: 13px;
}

.mode-attempt-row.verified {
  background: rgba(74, 222, 128, 0.08);
  border: 1px solid rgba(74, 222, 128, 0.2);
}

.mode-attempt-row.failed {
  background: rgba(248, 113, 113, 0.08);
  border: 1px solid rgba(248, 113, 113, 0.3);
}

.mode-attempt-row.skipped {
  background: var(--bg);
  border: 1px solid var(--border);
  opacity: 0.7;
}

.attempt-mode {
  font-weight: 600;
  width: 60px;
  font-size: 13px;
}

.attempt-status {
  flex: 1;
  font-size: 12px;
}

.attempt-time {
  font-size: 11px;
  color: var(--text-muted);
}

.status-text.verified {
  color: #4ade80;
}

.status-text.failed {
  color: #f87171;
}

.status-text.skipped {
  color: var(--text-muted);
}

/* Test badge */
.test-badge {
  padding: 2px 10px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 700;
  letter-spacing: 0.05em;
}

.test-badge.pass {
  background: rgba(74, 222, 128, 0.15);
  color: #4ade80;
  border: 1px solid rgba(74, 222, 128, 0.3);
}

.test-badge.warn {
  background: rgba(250, 204, 21, 0.15);
  color: #fbbf24;
  border: 1px solid rgba(250, 204, 21, 0.3);
}

.test-badge.fail {
  background: rgba(248, 113, 113, 0.15);
  color: #f87171;
  border: 1px solid rgba(248, 113, 113, 0.3);
}

/* Actions */
.test-actions {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-top: 10px;
}

.header-actions {
  display: flex;
  gap: 8px;
}

.result-text {
  font-size: 12px;
}

.result-text.success {
  color: #4ade80;
}

.result-text.error {
  color: #f87171;
}

.loading-text {
  color: var(--text-muted);
}

/* Spinner small */
.spinner-small {
  display: inline-block;
  width: 14px;
  height: 14px;
  border: 2px solid rgba(255,255,255,0.15);
  border-top-color: var(--accent);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
  flex-shrink: 0;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

/* === Auto Launch Styles === */
.launch-section {
  margin-bottom: 16px;
}

.section-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  margin-bottom: 8px;
  padding-left: 4px;
}

.section-hint {
  font-size: 11px;
  font-weight: 400;
  color: var(--text-muted);
}

.launch-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 6px;
}

.launch-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 8px 12px;
  border-radius: 8px;
  border: 1px solid var(--border);
  background: var(--bg-elevated);
  cursor: pointer;
  transition: all 0.15s;
  user-select: none;
}

.launch-item:hover {
  border-color: var(--accent);
}

.launch-item.enabled {
  border-color: rgba(74, 222, 128, 0.3);
  background: rgba(74, 222, 128, 0.05);
}

.launch-item.disabled {
  opacity: 0.5;
}

.launch-check {
  flex-shrink: 0;
}

.check-box {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border: 2px solid var(--border);
  border-radius: 4px;
  font-size: 12px;
  color: transparent;
  transition: all 0.15s;
}

.check-box.checked {
  background: var(--accent);
  border-color: var(--accent);
  color: white;
}

.launch-info {
  flex: 1;
  min-width: 0;
}

.launch-name {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.launch-desc {
  display: block;
  font-size: 11px;
  color: var(--text-muted);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.launch-wait {
  flex-shrink: 0;
  font-size: 11px;
  color: var(--text-muted);
  font-family: 'Cascadia Code', 'Fira Code', monospace;
  background: var(--bg);
  padding: 2px 6px;
  border-radius: 4px;
}
</style>
