<template>
  <div class="ai-agent-page">

    <!-- Settings Toggle -->
    <div class="settings-bar">
      <div class="mode-switch">
        <button :class="['mode-btn', cloudMode ? 'active-cloud' : 'active-local']" @click="toggleMode">
          <span v-if="cloudMode">☁️ 云端 AI (NVIDIA)</span>
          <span v-else>💻 本地分析</span>
        </button>
        <span class="mode-status" :class="cloudMode ? 'status-on' : 'status-off'">
          {{ cloudMode ? '已启用' : '本地模式' }}
        </span>
      </div>
      <button class="btn-sm btn-settings" @click="showSettings = !showSettings" title="API 设置">
        ⚙️ 设置
      </button>
    </div>

    <!-- Settings Panel (overlay) -->
    <div v-if="showSettings" class="settings-overlay" @click.self="showSettings = false">
      <div class="settings-panel card">
      <div class="settings-header">
        <strong>☁️ NVIDIA API 配置</strong>
        <button class="btn-sm" @click="showSettings = false">✕</button>
      </div>
      <div class="settings-body">
        <div class="setting-row">
          <label>API Key</label>
          <div class="input-group">
            <input
              type="password"
              v-model="apiConfig.apiKey"
              placeholder="nvapi-..."
              class="setting-input"
            />
            <button class="btn-sm" @click="testConnection" :disabled="testingConn">
              {{ testingConn ? '测试中...' : '🔗 测试' }}
            </button>
          </div>
          <a href="https://build.nvidia.com" target="_blank" class="hint-link">获取 API Key →</a>
        </div>
        <div class="setting-row">
          <label>模型</label>
          <select v-model="apiConfig.model" class="setting-select">
            <option v-for="m in modelList" :key="m.id" :value="m.id">
              {{ m.name }} — {{ m.desc }}
            </option>
          </select>
        </div>
        <div class="setting-row">
          <label>启用云端 AI</label>
          <label class="toggle">
            <input type="checkbox" v-model="apiConfig.enabled" @change="onConfigChange" />
            <span class="toggle-slider"></span>
          </label>
        </div>
        <div v-if="testResult" class="test-result" :class="testResult.success ? 'success' : 'error'">
          {{ testResult.message || testResult.error }}
        </div>
      </div>
      <div class="settings-footer">
          <button class="btn-sm btn-save" @click="saveConfig">💾 保存配置</button>
        </div>
      </div>
    </div>

    <!-- Quick Ask -->
    <div class="quick-ask-section card">
      <div class="quick-ask-chips">
        <span class="chips-label">Quick Ask:</span>
        <button class="chip" @click="askQuick('CPU使用率多少？')">CPU使用率</button>
        <button class="chip" @click="askQuick('内存占用情况')">内存占用</button>
        <button class="chip" @click="askQuick('磁盘空间还剩多少')">磁盘空间</button>
        <button class="chip" @click="askQuick('系统概览')">系统概览</button>
        <button class="chip" @click="askQuick('开机多久了')">运行时间</button>
        <button class="chip" @click="askQuick('显卡信息')">显卡信息</button>
        <button class="chip" @click="askQuick('电源和电池状态')">电源状态</button>
        <button class="chip" @click="askQuick('帮我分析下系统性能，给出优化建议')">⚡ 性能优化建议</button>
        <button v-if="cloudMode" class="chip chip-cloud" @click="askQuick('我的CPU使用率偏高，帮我排查可能的原因并给出优化方案')">🧠 AI 排查</button>
        <button class="chip" @click="askQuick('亮度调到70%')">💡 调亮度</button>
      </div>
    </div>

    <!-- Chat Messages -->
    <div class="agent-chat" ref="agentChat">
      <!-- Welcome -->
      <div v-if="!agentMessages.length" class="agent-welcome">
        <div class="welcome-icon">{{ cloudMode ? '☁️' : '🤖' }}</div>
        <div class="welcome-text">
          <strong>{{ cloudMode ? 'AI Agent 云端助手' : 'AI Agent 系统助手' }}</strong>
          <p v-if="cloudMode">云端模式已启用，AI 可以智能分析你的系统状态、日志文件、ETL Trace，并给出专业的性能优化建议和 PowerShell 脚本。</p>
          <p v-else>我可以查看你电脑的基本信息，回答系统状态相关的问题，还能调节屏幕亮度。</p>
          <p>点击上方快捷提问，或在下方输入框输入你的问题。点击 ➕ 可加载本地文件进行分析。</p>
        </div>
      </div>

      <!-- Messages -->
      <div v-for="(msg, idx) in agentMessages" :key="idx" :class="['agent-msg', 'msg-' + msg.role]">
        <div class="msg-avatar">
          <svg v-if="msg.role === 'user'" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
          <span v-else-if="msg.cloud">☁️</span>
          <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 0 2-2h14a2 2 0 0 0 2 2z"/></svg>
        </div>
        <div class="msg-body">
          <div class="msg-source" v-if="msg.cloud && msg.role === 'agent'">
            <span class="source-badge">☁️ NVIDIA {{ currentModelName }}</span>
          </div>
          <div class="msg-content" v-html="renderMarkdown(msg.content)"></div>
          <div class="msg-time">{{ msg.timestamp }}</div>
        </div>
      </div>

      <!-- Thinking -->
      <div v-if="agentThinking" class="agent-msg msg-agent">
        <div class="msg-avatar">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>
        </div>
        <div class="msg-body">
          <div class="msg-content thinking">{{ cloudMode ? '☁️ 云端 AI 正在分析...' : '正在分析...' }}</div>
        </div>
      </div>
    </div>

    <!-- Input Area -->
    <div class="agent-input-area">
      <!-- Load File Button -->
      <button class="btn-load-file" @click="triggerFileLoad" :disabled="agentThinking" title="加载本地文件">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
          <line x1="12" y1="5" x2="12" y2="19"/>
          <line x1="5" y1="12" x2="19" y2="12"/>
        </svg>
      </button>
      <input
        type="file"
        ref="fileInput"
        style="display:none"
        @change="onFileSelected"
        accept=".txt,.log,.csv,.json,.xml,.md,.ini,.cfg,.conf,.yaml,.yml"
      />
      <input
        type="text"
        class="agent-input"
        v-model="agentQuestion"
        :placeholder="cloudMode ? '输入问题，AI 将结合系统状态智能分析...' : '输入你的问题，例如：CPU温度是多少？'"
        @keydown.enter="sendAgentMessage"
        :disabled="agentThinking"
        ref="agentInput"
      />
      <button class="btn-send" @click="sendAgentMessage" :disabled="!agentQuestion.trim() || agentThinking">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <line x1="22" y1="2" x2="11" y2="13"/>
          <polygon points="22 2 15 22 11 13 2 9 22 2"/>
        </svg>
      </button>
    </div>

  </div>
</template>

<script>
export default {
  name: 'AIAgent',
  props: {
    theme: { type: String, default: 'dark' }
  },
  data() {
    return {
      agentQuestion: '',
      agentMessages: [],
      agentThinking: false,
      uploadedFile: null,
      // Cloud mode
      cloudMode: false,
      showSettings: false,
      apiConfig: {
        apiKey: '',
        model: 'z-ai/glm-5.1',
        enabled: false
      },
      modelList: [],
      testingConn: false,
      testResult: null
    }
  },
  computed: {
    currentModelName() {
      const m = this.modelList.find(x => x.id === this.apiConfig.model)
      return m ? m.name : 'AI'
    }
  },
  async mounted() {
    this.$nextTick(() => {
      this.$refs.agentInput?.focus()
    })
    await this.loadConfig()
  },
  methods: {
    // === Config ===
    async loadConfig() {
      try {
        if (window.go?.main?.App) {
          // Load model list
          const modelsStr = await window.go.main.App.GetNVIDIAModelList()
          if (modelsStr) {
            try { this.modelList = JSON.parse(modelsStr) } catch(e) {}
          }
          // Load saved config
          const cfg = await window.go.main.App.LoadNVIDIAConfig()
          if (cfg) {
            this.apiConfig = {
              apiKey: cfg.apiKey || '',
              model: cfg.model || 'z-ai/glm-5.1',
              enabled: cfg.enabled || false
            }
          }
          // Check if enabled
          const enabled = await window.go.main.App.IsNVIDIAEnabled()
          this.cloudMode = enabled
        }
      } catch (e) { console.error('Failed to load NVIDIA config:', e) }
    },
    async saveConfig() {
      try {
        if (window.go?.main?.App) {
          await window.go.main.App.SaveNVIDIAConfig({
            apiKey: this.apiConfig.apiKey,
            model: this.apiConfig.model,
            enabled: this.apiConfig.enabled,
            baseUrl: 'https://integrate.api.nvidia.com/v1'
          })
          this.cloudMode = this.apiConfig.enabled && this.apiConfig.apiKey !== ''
          this.showSettings = false
          this.$nextTick(() => {
            this.scrollAgentChat()
          })
        }
      } catch (e) {
        alert('保存配置失败: ' + e)
      }
    },
    onConfigChange() {
      // Auto-apply when toggling
    },
    toggleMode() {
      if (this.apiConfig.enabled && this.apiConfig.apiKey) {
        this.cloudMode = !this.cloudMode
      } else {
        this.showSettings = true
      }
    },
    async testConnection() {
      if (!this.apiConfig.apiKey) {
        this.testResult = { success: false, message: '请先输入 API Key' }
        return
      }
      this.testingConn = true
      this.testResult = null
      try {
        if (window.go?.main?.App) {
          const result = await window.go.main.App.TestNVIDIAConnection(this.apiConfig.apiKey, this.apiConfig.model)
          try { this.testResult = JSON.parse(result) } catch(e) { this.testResult = { success: false, message: result } }
        }
      } catch (e) {
        this.testResult = { success: false, message: '测试失败: ' + e }
      } finally {
        this.testingConn = false
      }
    },

    // === File Loading ===
    triggerFileLoad() {
      this.$refs.fileInput.click()
    },
    async onFileSelected(e) {
      const file = e.target.files?.[0]
      if (!file) return
      const MAX_SIZE = 5 * 1024 * 1024
      if (file.size > MAX_SIZE) {
        alert('File too large (' + (file.size/1024/1024).toFixed(1) + 'MB). Max 5MB.')
        e.target.value = ''
        return
      }
      const reader = new FileReader()
      reader.onload = (ev) => {
        const content = ev.target.result
        const MAX_CHARS = 100000
        const truncated = content.length > MAX_CHARS
          ? content.substring(0, MAX_CHARS) + '\n\n...(file truncated, showing first 100K chars)'
          : content
        this.uploadedFile = { name: file.name, content: truncated, size: file.size }
        this.agentMessages.push({
          role: 'user',
          content: '📎 已加载文件: **' + file.name + '** (' + (file.size/1024).toFixed(1) + 'KB)',
          timestamp: new Date().toLocaleTimeString(),
          cloud: false
        })
        this.scrollAgentChat()
      }
      reader.readAsText(file)
      e.target.value = ''
    },

    // === Send Message ===
    async sendAgentMessage() {
      const q = this.agentQuestion.trim()
      if (!q || this.agentThinking) return

      let fullQuestion = q
      let displayQuestion = q
      if (this.uploadedFile) {
        fullQuestion = '[File: ' + this.uploadedFile.name + ' (' + (this.uploadedFile.size/1024).toFixed(1) + 'KB)]\n```\n' + this.uploadedFile.content + '\n```\n\nUser question: ' + q
        displayQuestion = '📎 ' + this.uploadedFile.name + ': ' + q
        this.uploadedFile = null
      }

      this.agentMessages.push({ role: 'user', content: displayQuestion, timestamp: new Date().toLocaleTimeString(), cloud: false })
      this.agentQuestion = ''
      this.agentThinking = true
      this.$nextTick(() => { this.scrollAgentChat() })

      try {
        let answer = ''
        let isCloud = false

        if (this.cloudMode && window.go?.main?.App?.AskNVIDIACloud) {
          // Cloud mode
          try {
            answer = await window.go.main.App.AskNVIDIACloud(fullQuestion)
            isCloud = true
          } catch (cloudErr) {
            // Fallback to local on cloud error
            answer = '⚠️ 云端 AI 暂时不可用: ' + cloudErr + '\n\n已切换到本地模式回答：\n'
            if (window.go?.main?.App?.AskAIAgent) {
              answer += await window.go.main.App.AskAIAgent(fullQuestion)
            }
            isCloud = false
          }
        } else {
          // Local mode
          if (window.go?.main?.App) {
            answer = await window.go.main.App.AskAIAgent(fullQuestion)
          }
        }

        this.agentMessages.push({
          role: 'agent',
          content: answer,
          timestamp: new Date().toLocaleTimeString(),
          cloud: isCloud
        })
      } catch (e) {
        this.agentMessages.push({
          role: 'agent',
          content: '抱歉，处理请求时出错：' + e,
          timestamp: new Date().toLocaleTimeString(),
          cloud: false
        })
      } finally {
        this.agentThinking = false
        this.$nextTick(() => { this.scrollAgentChat() })
      }
    },

    askQuick(question) {
      this.agentQuestion = question
      this.sendAgentMessage()
    },
    scrollAgentChat() {
      const el = this.$refs.agentChat
      if (el) el.scrollTop = el.scrollHeight
    },
    renderMarkdown(text) {
      if (!text) return ''
      text = text.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>')
      text = text.replace(/\n/g, '<br>')
      text = text.replace(/`([^`]+)`/g, '<code>$1</code>')
      return text
    }
  }
}
</script>

<style scoped>
.ai-agent-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
  min-height: 0;
  flex: 1;
}

/* === Settings Bar === */
.settings-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 16px;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.mode-switch {
  display: flex;
  align-items: center;
  gap: 10px;
}

.mode-btn {
  padding: 6px 14px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.2s;
  font-family: inherit;
}

.mode-btn.active-cloud {
  background: rgba(76,175,80,0.15);
  border-color: rgba(76,175,80,0.4);
  color: #4CAF50;
}

.mode-btn.active-local {
  background: rgba(66,133,244,0.15);
  border-color: rgba(66,133,244,0.4);
  color: #4285F4;
}

.mode-btn:hover {
  opacity: 0.85;
}

.mode-status {
  font-size: 11px;
  font-weight: 600;
  padding: 2px 8px;
  border-radius: 10px;
}

.mode-status.status-on {
  background: rgba(76,175,80,0.15);
  color: #4CAF50;
}

.mode-status.status-off {
  background: rgba(255,255,255,0.05);
  color: var(--text-tertiary);
}

.btn-settings {
  padding: 4px 10px;
}

/* === Settings Overlay === */
.settings-overlay {
  position: fixed;
  inset: 0;
 background: rgba(0,0,0,0.5);
 display: flex;
  align-items: center;
  justify-content: center;
 z-index: 1000;
}

.settings-panel {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  width: 480px;
  max-width: 90vw;
  overflow: hidden;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color);
}

.settings-header strong {
  font-size: 13px;
  color: var(--text-primary);
}

.settings-body {
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.setting-row {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.setting-row label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
}

.input-group {
  display: flex;
  gap: 8px;
}

.setting-input {
  flex: 1;
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 12px;
  font-family: 'Consolas', monospace;
  outline: none;
}

.setting-input:focus {
  border-color: #4CAF50;
}

.setting-input::placeholder {
  color: var(--text-tertiary);
}

.setting-select {
  padding: 8px 12px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  font-size: 12px;
  outline: none;
  cursor: pointer;
}

.hint-link {
  font-size: 11px;
  color: #4CAF50;
  text-decoration: none;
}

.hint-link:hover {
  text-decoration: underline;
}

/* Toggle switch */
.toggle {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 22px;
  cursor: pointer;
}

.toggle input {
  opacity: 0;
  width: 0;
  height: 0;
}

.toggle-slider {
  position: absolute;
  inset: 0;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 22px;
  transition: 0.3s;
}

.toggle-slider::before {
  content: '';
  position: absolute;
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background: var(--text-secondary);
  border-radius: 50%;
  transition: 0.3s;
}

.toggle input:checked + .toggle-slider {
  background: rgba(76,175,80,0.3);
  border-color: #4CAF50;
}

.toggle input:checked + .toggle-slider::before {
  transform: translateX(18px);
  background: #4CAF50;
}

.test-result {
  padding: 8px 12px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
}

.test-result.success {
  background: rgba(76,175,80,0.1);
  color: #4CAF50;
  border: 1px solid rgba(76,175,80,0.3);
}

.test-result.error {
  background: rgba(230,63,50,0.1);
  color: var(--lenovo-red);
  border: 1px solid rgba(230,63,50,0.3);
}

.settings-footer {
  padding: 10px 16px;
  border-top: 1px solid var(--border-color);
  display: flex;
  justify-content: flex-end;
}

.btn-save {
  background: rgba(76,175,80,0.15);
  border-color: rgba(76,175,80,0.3);
  color: #4CAF50;
}

.btn-save:hover {
  background: rgba(76,175,80,0.25);
}

/* === Quick Ask === */
.quick-ask-section {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
}

.quick-ask-chips {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 12px 16px;
}

.chips-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  padding-right: 8px;
  align-self: center;
}

.chip {
  padding: 6px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 16px;
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
  white-space: nowrap;
}

.chip:hover {
  background: var(--bg-card-hover);
  color: var(--text-primary);
  border-color: var(--lenovo-red);
}

.chip-cloud {
  background: rgba(76,175,80,0.1);
  border-color: rgba(76,175,80,0.3);
  color: #4CAF50;
}

.chip-cloud:hover {
  background: rgba(76,175,80,0.2);
  border-color: #4CAF50;
}

/* === Chat === */
.agent-chat {
  flex: 1;
  min-height: 320px;
  max-height: 480px;
  overflow-y: auto;
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  padding: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.agent-welcome {
  display: flex;
  align-items: flex-start;
  gap: 16px;
  padding: 24px;
}

.welcome-icon {
  font-size: 48px;
  line-height: 1;
}

.welcome-text {
  flex: 1;
}

.welcome-text strong {
  font-size: 16px;
  color: var(--text-primary);
  display: block;
  margin-bottom: 8px;
}

.welcome-text p {
  margin: 4px 0;
  font-size: 13px;
  color: var(--text-secondary);
  line-height: 1.6;
}

.agent-msg {
  display: flex;
  gap: 10px;
}

.agent-msg.msg-user {
  flex-direction: row-reverse;
}

.msg-avatar {
  width: 32px;
  height: 32px;
  border-radius: 8px;
  background: var(--bg-tertiary);
  display: flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
  border: 1px solid var(--border-color);
  font-size: 14px;
}

.msg-user .msg-avatar {
  background: rgba(230,63,50,0.1);
  border-color: rgba(230,63,50,0.3);
  color: var(--lenovo-red);
}

.msg-agent .msg-avatar {
  background: rgba(76,175,80,0.1);
  border-color: rgba(76,175,80,0.3);
  color: #4CAF50;
}

.msg-body {
  max-width: 80%;
}

.msg-user .msg-body {
  text-align: right;
}

.msg-source {
  margin-bottom: 4px;
}

.source-badge {
  font-size: 10px;
  color: #4CAF50;
  font-weight: 600;
}

.msg-content {
  padding: 10px 14px;
  border-radius: 12px;
  font-size: 13px;
  line-height: 1.6;
  background: var(--bg-tertiary);
  color: var(--text-primary);
  text-align: left;
  display: inline-block;
}

.msg-user .msg-content {
  background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  color: white;
}

.msg-agent .msg-content {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
}

.msg-content code {
  background: rgba(0,0,0,0.1);
  padding: 1px 4px;
  border-radius: 3px;
  font-family: 'Consolas', monospace;
  font-size: 12px;
}

.msg-user .msg-content code {
  background: rgba(255,255,255,0.2);
}

.msg-content.thinking {
  color: var(--text-tertiary);
  font-style: italic;
}

.msg-time {
  font-size: 10px;
  color: var(--text-tertiary);
  margin-top: 4px;
}

/* === Input Area === */
.agent-input-area {
  display: flex;
  gap: 8px;
}

.btn-load-file {
  padding: 0 12px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--bg-card);
  color: var(--text-secondary);
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
  flex-shrink: 0;
}

.btn-load-file:hover:not(:disabled) {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.05);
}

.btn-load-file:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.agent-input {
  flex: 1;
  padding: 12px 16px;
  border: 1px solid var(--border-color);
  border-radius: 10px;
  background: var(--bg-card);
  color: var(--text-primary);
  font-size: 13px;
  outline: none;
  transition: border-color 0.2s;
  font-family: inherit;
}

.agent-input:focus {
  border-color: var(--lenovo-red);
}

.agent-input::placeholder {
  color: var(--text-tertiary);
}

.agent-input:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-send {
  padding: 0 16px;
  border: none;
  border-radius: 10px;
  background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  color: white;
  font-size: 14px;
  cursor: pointer;
  display: flex;
  align-items: center;
  justify-content: center;
  transition: var(--transition);
}

.btn-send:hover:not(:disabled) {
  opacity: 0.9;
  transform: scale(1.02);
}

.btn-send:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

/* Shared */
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

@keyframes spin {
  to { transform: rotate(360deg); }
}

.spinning {
  animation: spin 0.8s linear infinite;
}
</style>
