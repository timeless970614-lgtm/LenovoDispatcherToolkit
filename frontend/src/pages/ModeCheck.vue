<template>
  <div class="mode-check-page">
    <!-- Header -->
    <div class="page-header">
      <h1 class="page-title">Mode Check</h1>
      <p class="page-subtitle">Dispatcher status and feature analysis</p>
    </div>

    <!-- Loading State -->
    <div v-if="loading" class="loading-state">
      <div class="spinner"></div>
      <span>Loading mode check data...</span>
    </div>

    <!-- Main Content -->
    <div v-else-if="info" class="content-grid">
      <!-- DYTC Function Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <polyline points="16 18 22 12 16 6"/>
              <polyline points="8 6 2 12 8 18"/>
            </svg>
            DYTC Dispatcher Function
          </span>
        </div>

        <div class="dytc-display">
          <div class="dytc-value">
            <span class="dytc-label">DISPATCHER_FUNCTION</span>
            <span class="dytc-hex">{{ info.dytcValue ? '0x' + info.dytcValue.toString(16).toUpperCase().padStart(8, '0') : 'N/A' }}</span>
          </div>
          <div class="dytc-binary" v-if="info.dytcBinary">
            <span class="binary-label">Binary:</span>
            <span class="binary-value">{{ info.dytcBinary }}</span>
          </div>
        </div>
      </div>

      <!-- Policy EnableFunc Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            Policy EnableFunc
          </span>
          <span class="func-value">{{ info.enableFuncHex }}</span>
        </div>

        <div class="policy-list">
          <div 
            v-for="policy in info.enableFuncPolicies" 
            :key="policy.bit"
            :class="['policy-row', policy.enabled ? 'policy-enabled' : 'policy-disabled']"
          >
            <div class="policy-bit">
              <span class="bit-num">bit{{ policy.bit }}</span>
            </div>
            <div class="policy-indicator">
              <span v-if="policy.enabled" class="led led-on"></span>
              <span v-else class="led led-off"></span>
            </div>
            <div class="policy-info">
              <span class="policy-name">{{ policy.name }}</span>
              <span class="policy-desc">{{ policy.desc }}</span>
            </div>
            <div class="policy-status">
              <span :class="['status-tag', policy.enabled ? 'tag-on' : 'tag-off']">
                {{ policy.enabled ? 'ON' : 'OFF' }}
              </span>
            </div>
          </div>
        </div>

        <div class="policy-summary">
          <span class="summary-item">
            <span class="led led-on"></span>
            {{ info.enableFuncPolicies.filter(p => p.enabled).length }} policies enabled
          </span>
          <span class="summary-item">
            <span class="led led-off"></span>
            {{ info.enableFuncPolicies.filter(p => !p.enabled).length }} policies disabled
          </span>
        </div>
      </div>

      <!-- Features Card -->
      <div class="card" v-if="info.features && info.features.length">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
              <polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/>
            </svg>
            Supported Features
          </span>
        </div>

        <div class="features-list">
          <div v-for="feature in info.features" :key="feature.name" class="feature-row">
            <div class="feature-name">{{ feature.name }}</div>
            <div class="feature-support" :class="feature.value === 'Y' ? 'support-yes' : 'support-na'">
              {{ feature.value === 'Y' ? 'Supported' : 'N/A' }}
            </div>
            <div class="feature-desc">{{ feature.support }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Error State -->
    <div v-else class="error-state">
      <p>Failed to load mode check data</p>
      <button class="btn btn-primary" @click="refresh">Retry</button>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ModeCheck',
  data() {
    return {
      info: null,
      loading: true,
      serviceStatus: 'Unknown'
    }
  },
  async mounted() {
    await this.refresh()
  },
  methods: {
    async refresh() {
      this.loading = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          this.info = await window.go.main.App.GetModeCheckInfo()
          this.serviceStatus = await window.go.main.App.GetServiceStatus()
        }
      } catch (e) {
        console.error('Failed to load mode check info:', e)
      } finally {
        this.loading = false
      }
    }
  }
}
</script>

<style scoped>
.mode-check-page {
  padding: 24px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 24px;
}

.page-title {
  font-size: 28px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 4px 0;
}

.page-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin: 0;
}

.content-grid {
  display: grid;
  grid-template-columns: 1fr;
  gap: 20px;
}

.loading-state {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 12px;
  padding: 60px;
  color: var(--text-secondary);
}

.error-state {
  text-align: center;
  padding: 60px;
  color: var(--text-secondary);
}

/* Card Styles */
.card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-lg);
  padding: 20px;
}

.card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 16px;
}

.card-title {
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
  display: flex;
  align-items: center;
}

/* Info Grid */
.info-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
}

.info-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.info-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.info-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.info-value.mono {
  font-family: 'Consolas', monospace;
  font-size: 12px;
}

/* Status Grid */
.status-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 16px;
}

.status-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.status-label {
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.status-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--text-primary);
}

.status-value.highlight {
  color: var(--lenovo-red);
  font-weight: 600;
}

.status-value.badge {
  background: rgba(230, 63, 50, 0.1);
  padding: 2px 10px;
  border-radius: 12px;
  font-size: 12px;
  color: var(--lenovo-red);
  display: inline-block;
}

.status-value.mono {
  font-family: 'Consolas', monospace;
  font-size: 12px;
}

/* Status Badge */
.status-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 20px;
}

.status-running {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.status-stopped {
  background: rgba(244, 67, 54, 0.15);
  color: #F44336;
}

/* DYTC Display */
.dytc-display {
  background: var(--bg-tertiary);
  border-radius: 8px;
  padding: 16px;
}

.dytc-value {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.dytc-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.dytc-hex {
  font-family: 'Consolas', monospace;
  font-size: 20px;
  font-weight: 700;
  color: var(--lenovo-red);
}

.dytc-binary {
  display: flex;
  align-items: center;
  gap: 8px;
}

.binary-label {
  font-size: 11px;
  color: var(--text-tertiary);
}

.binary-value {
  font-family: 'Consolas', monospace;
  font-size: 11px;
  color: var(--text-secondary);
  letter-spacing: 0.5px;
}

/* Policy List */
.func-value {
  font-family: 'Consolas', monospace;
  font-size: 12px;
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.1);
  padding: 3px 10px;
  border-radius: 6px;
}

.policy-list {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
  max-height: 400px;
  overflow-y: auto;
}

.policy-row {
  display: grid;
  grid-template-columns: 60px 28px 1fr 60px;
  gap: 12px;
  padding: 10px 14px;
  align-items: center;
  border-bottom: 1px solid var(--border-color);
  transition: background 0.15s;
}

.policy-row:last-child {
  border-bottom: none;
}

.policy-row:hover {
  background: var(--bg-tertiary);
}

.policy-row.policy-enabled {
  background: rgba(76, 175, 80, 0.03);
}

.policy-row.policy-disabled {
  opacity: 0.6;
}

.bit-num {
  font-family: 'Consolas', monospace;
  font-size: 10px;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 4px;
}

.led {
  width: 10px;
  height: 10px;
  border-radius: 50%;
}

.led-on {
  background: #4CAF50;
  box-shadow: 0 0 5px rgba(76, 175, 80, 0.6);
}

.led-off {
  background: var(--text-tertiary);
  opacity: 0.4;
}

.policy-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.policy-name {
  font-size: 12px;
  font-weight: 600;
  color: var(--text-primary);
  font-family: 'Consolas', monospace;
}

.policy-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}

.status-tag {
  font-size: 10px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 4px;
  text-align: center;
}

.tag-on {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.tag-off {
  background: var(--bg-tertiary);
  color: var(--text-tertiary);
}

.policy-summary {
  display: flex;
  gap: 20px;
  margin-top: 12px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  font-size: 12px;
  color: var(--text-secondary);
}

.summary-item {
  display: flex;
  align-items: center;
  gap: 6px;
}

/* Features List */
.features-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.feature-row {
  display: grid;
  grid-template-columns: 1fr 100px 1fr;
  gap: 12px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  align-items: center;
}

.feature-name {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}

.feature-support {
  font-size: 11px;
  font-weight: 600;
  padding: 3px 10px;
  border-radius: 12px;
  text-align: center;
}

.support-yes {
  background: rgba(76, 175, 80, 0.15);
  color: #4CAF50;
}

.support-na {
  background: var(--bg-primary);
  color: var(--text-tertiary);
}

.feature-desc {
  font-size: 12px;
  color: var(--text-tertiary);
}

/* Button */
.btn-icon {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  padding: 7px;
  cursor: pointer;
  color: var(--text-secondary);
  transition: var(--transition);
  display: flex;
  align-items: center;
  justify-content: center;
}

.btn-icon:hover:not(:disabled) {
  color: var(--lenovo-red);
  border-color: var(--lenovo-red);
}

.btn-icon:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

/* Responsive */
@media (max-width: 900px) {
  .info-grid,
  .status-grid {
    grid-template-columns: 1fr;
  }
}
</style>
