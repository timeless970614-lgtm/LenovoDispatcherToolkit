<template>
  <div class="ppm-page">
    <!-- Platform Info Card -->
    <div class="card platform-card">
      <div class="card-header">
        <div class="card-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="4" y="4" width="16" height="16" rx="2"/>
            <rect x="9" y="9" width="6" height="6"/>
            <line x1="9" y1="2" x2="9" y2="4"/>
            <line x1="15" y1="2" x2="15" y2="4"/>
            <line x1="9" y1="20" x2="9" y2="22"/>
            <line x1="15" y1="20" x2="15" y2="22"/>
            <line x1="20" y1="9" x2="22" y2="9"/>
            <line x1="20" y1="14" x2="22" y2="14"/>
            <line x1="2" y1="9" x2="4" y2="9"/>
            <line x1="2" y1="14" x2="4" y2="14"/>
          </svg>
        </div>
        <div class="card-title-info">
          <h2>Platform Information</h2>
          <p class="card-subtitle">System platform and CPU details</p>
        </div>
      </div>
      <div class="platform-content">
        <div class="platform-grid">
          <div class="platform-item">
            <span class="platform-label">CPU</span>
            <span class="platform-value">{{ platformInfo.cpuName || 'Loading...' }}</span>
          </div>
          <div class="platform-item">
            <span class="platform-label">Cores / Threads</span>
            <span class="platform-value">{{ platformInfo.cores }} / {{ platformInfo.threads }}</span>
          </div>
          <div class="platform-item">
            <span class="platform-label">Platform</span>
            <span class="platform-value">{{ platformInfo.platform || 'Intel' }}</span>
          </div>
          <div class="platform-item">
            <span class="platform-label">Architecture</span>
            <span class="platform-value">{{ platformInfo.architecture || 'x64' }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- PPM Drivers Card -->
    <div class="card drivers-card">
      <div class="card-header">
        <div class="card-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M12 2v4M12 18v4M4.93 4.93l2.83 2.83M16.24 16.24l2.83 2.83M2 12h4M18 12h4M4.93 19.07l2.83-2.83M16.24 7.76l2.83-2.83"/>
            <circle cx="12" cy="12" r="4"/>
          </svg>
        </div>
        <div class="card-title-info">
          <h2>PPM Drivers</h2>
          <p class="card-subtitle">Intel Processor Power Management Components</p>
        </div>
        <button class="btn-refresh" @click="loadPPMInfo" :disabled="loading">
          <svg :class="{ 'spin': loading }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6"/>
            <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
          </svg>
          <span>{{ loading ? 'Scanning...' : 'Refresh' }}</span>
        </button>
      </div>
      <div class="drivers-list">
        <div v-for="driver in ppmDrivers" :key="driver.name" class="driver-item">
          <div class="driver-icon" :class="getDriverClass(driver.name)">
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M12 2L2 7l10 5 10-5-10-5zM2 17l10 5 10-5M2 12l10 5 10-5"/>
            </svg>
          </div>
          <div class="driver-info">
            <div class="driver-name">{{ driver.name }}</div>
            <div class="driver-meta">
              <span class="driver-version">v{{ driver.version }}</span>
              <span class="driver-date">{{ formatDate(driver.date) }}</span>
              <span class="driver-location" v-if="driver.location">{{ driver.location }}</span>
            </div>
          </div>
          <div class="driver-status">
            <span class="status-badge installed">Installed</span>
          </div>
        </div>
      </div>
    </div>

    <!-- PPM Parameters Card -->
    <div class="card params-card">
      <div class="card-header">
        <div class="card-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="3"/>
            <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
        </div>
        <div class="card-title-info">
          <h2>PPM Parameters Analysis</h2>
          <p class="card-subtitle">Key power management parameters and their meanings</p>
        </div>
      </div>
      <div class="params-grid">
        <div v-for="param in ppmParameters" :key="param.key" class="param-item">
          <div class="param-header">
            <span class="param-key">{{ param.key }}</span>
            <span class="param-category">{{ param.category }}</span>
          </div>
          <div class="param-description">{{ param.description }}</div>
          <div class="param-impact">
            <span class="impact-label">Impact:</span>
            <span :class="['impact-value', param.impactLevel]">{{ param.impact }}</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Intel IPF Architecture Card -->
    <div class="card architecture-card">
      <div class="card-header">
        <div class="card-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/>
            <path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
            <line x1="8" y1="7" x2="16" y2="7"/>
            <line x1="8" y1="11" x2="16" y2="11"/>
            <line x1="8" y1="15" x2="12" y2="15"/>
          </svg>
        </div>
        <div class="card-title-info">
          <h2>Intel Innovation Platform Framework</h2>
          <p class="card-subtitle">PPM Architecture Overview</p>
        </div>
      </div>
      <div class="architecture-content">
        <div class="arch-layer">
          <div class="layer-name">Application Layer</div>
          <div class="layer-items">
            <span class="layer-item">OS Power Policy</span>
            <span class="layer-item">Dynamic Tuning Client</span>
          </div>
        </div>
        <div class="arch-arrow">↓</div>
        <div class="arch-layer highlight">
          <div class="layer-name">Intel IPF Framework</div>
          <div class="layer-items">
            <span class="layer-item">Extensible Framework</span>
            <span class="layer-item">Processor Participant</span>
            <span class="layer-item">Generic Participant</span>
          </div>
        </div>
        <div class="arch-arrow">↓</div>
        <div class="arch-layer">
          <div class="layer-name">Driver Layer</div>
          <div class="layer-items">
            <span class="layer-item">Dynamic Tuning Technology</span>
            <span class="layer-item">PPM Provisioning</span>
          </div>
        </div>
        <div class="arch-arrow">↓</div>
        <div class="arch-layer hardware">
          <div class="layer-name">Hardware</div>
          <div class="layer-items">
            <span class="layer-item">CPU (P-Core/E-Core)</span>
            <span class="layer-item">Intel GPU</span>
            <span class="layer-item">NPU</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { GetPPMPlatformInfo, GetPPMDrivers } from '../../wailsjs/go/main/App'

export default {
  name: 'PPMDriver',
  props: ['theme'],
  data() {
    return {
      loading: false,
      platformInfo: {
        cpuName: '',
        cores: 0,
        threads: 0,
        platform: 'Intel',
        architecture: 'x64'
      },
      ppmDrivers: [],
      ppmParameters: [
        {
          key: 'PROCTHROTTLEMIN',
          category: 'Performance',
          description: 'Minimum processor performance state (%) - Sets the lower bound for CPU frequency scaling',
          impact: 'Lower values save power but may cause latency',
          impactLevel: 'medium'
        },
        {
          key: 'PROCTHROTTLEMAX',
          category: 'Performance',
          description: 'Maximum processor performance state (%) - Caps the maximum CPU frequency',
          impact: 'Lower values reduce peak performance and power',
          impactLevel: 'high'
        },
        {
          key: 'PERFEPP',
          category: 'Efficiency',
          description: 'Energy Performance Preference - Hints to CPU for power vs performance balance (0-255)',
          impact: 'Higher values prioritize efficiency over performance',
          impactLevel: 'high'
        },
        {
          key: 'HETEROINCREASETHRESHOLD',
          category: 'Heterogeneous',
          description: 'P-Core activation threshold - Workload level to promote from E-Core to P-Core',
          impact: 'Lower values activate P-Cores more aggressively',
          impactLevel: 'high'
        },
        {
          key: 'HETERODECREASETHRESHOLD',
          category: 'Heterogeneous',
          description: 'E-Core fallback threshold - Workload level to demote from P-Core to E-Core',
          impact: 'Higher values keep P-Cores active longer',
          impactLevel: 'medium'
        },
        {
          key: 'PROCFREQMAX',
          category: 'Frequency',
          description: 'Maximum frequency limit (MHz) for P-Cores - Hardware frequency cap',
          impact: 'Limits turbo boost frequency',
          impactLevel: 'high'
        },
        {
          key: 'CPMINCORES',
          category: 'Core Parking',
          description: 'Minimum active cores - Number of cores that must remain unparked',
          impact: 'Higher values reduce latency at cost of power',
          impactLevel: 'medium'
        },
        {
          key: 'SOFTPARKLATENCY',
          category: 'Core Parking',
          description: 'Soft parking latency tolerance - Time before cores are fully parked',
          impact: 'Higher values keep cores in shallow sleep longer',
          impactLevel: 'low'
        }
      ]
    }
  },
  mounted() {
    this.loadPPMInfo()
  },
  methods: {
    async loadPPMInfo() {
      this.loading = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          // Load platform info
          const platform = await window.go.main.App.GetPPMPlatformInfo()
          if (platform) {
            this.platformInfo = {
              cpuName: platform.cpuName || '',
              cores: platform.cores || 0,
              threads: platform.threads || 0,
              platform: platform.platform || 'Intel',
              architecture: platform.architecture || 'x64'
            }
          }

          // Load PPM drivers
          const drivers = await window.go.main.App.GetPPMDrivers()
          if (drivers && drivers.length > 0) {
            this.ppmDrivers = drivers
          }
        }
      } catch (e) {
        console.error('Failed to load PPM info:', e)
      } finally {
        this.loading = false
      }
    },
    getDriverClass(name) {
      if (name.includes('Dynamic Tuning')) return 'driver-dtt'
      if (name.includes('PPM')) return 'driver-ppm'
      if (name.includes('Innovation Platform')) return 'driver-ipf'
      return 'driver-other'
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
      // Parse WMI date format: 20241216000000.******+***
      const match = dateStr.match(/^(\d{4})(\d{2})(\d{2})/)
      if (match) {
        return `${match[1]}-${match[2]}-${match[3]}`
      }
      return dateStr
    }
  }
}
</script>

<style scoped>
.ppm-page {
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: auto auto;
  gap: 16px;
  padding: 16px;
  height: calc(100vh - 140px);
  overflow-y: auto;
}

.card {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.card-header {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 16px 20px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.card-icon {
  width: 40px;
  height: 40px;
  border-radius: var(--radius-md);
  background: rgba(230, 63, 50, 0.1);
  color: var(--lenovo-red);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-title-info h2 {
  margin: 0;
  font-size: 15px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-title-info .card-subtitle {
  margin: 2px 0 0 0;
  font-size: 12px;
  color: var(--text-secondary);
}

.btn-refresh {
  margin-left: auto;
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 6px 12px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}

.btn-refresh:hover:not(:disabled) {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
}

.btn-refresh:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.spin {
  animation: spin 1s linear infinite;
}

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

/* Platform Card */
.platform-card {
  grid-column: 1;
  grid-row: 1;
}

.platform-content {
  padding: 20px;
}

.platform-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.platform-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.platform-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.platform-value {
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

/* Drivers Card */
.drivers-card {
  grid-column: 2;
  grid-row: 1 / 3;
}

.drivers-list {
  padding: 12px;
  max-height: 500px;
  overflow-y: auto;
}

.driver-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  margin-bottom: 8px;
  transition: var(--transition);
}

.driver-item:hover {
  background: var(--bg-tertiary);
}

.driver-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-sm);
  display: flex;
  align-items: center;
  justify-content: center;
}

.driver-icon.driver-dtt {
  background: rgba(96, 165, 250, 0.15);
  color: var(--accent-blue);
}

.driver-icon.driver-ppm {
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red);
}

.driver-icon.driver-ipf {
  background: rgba(74, 222, 128, 0.15);
  color: var(--accent-green);
}

.driver-icon.driver-other {
  background: rgba(251, 191, 36, 0.15);
  color: var(--accent-yellow);
}

.driver-info {
  flex: 1;
  min-width: 0;
}

.driver-name {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.driver-meta {
  display: flex;
  gap: 8px;
  margin-top: 4px;
  font-size: 11px;
  color: var(--text-secondary);
}

.driver-version {
  color: var(--accent-green);
  font-weight: 600;
}

.driver-date {
  opacity: 0.8;
}

.driver-location {
  opacity: 0.6;
}

.driver-status .status-badge {
  font-size: 10px;
  font-weight: 700;
  padding: 3px 8px;
  border-radius: 4px;
  text-transform: uppercase;
}

.status-badge.installed {
  background: rgba(74, 222, 128, 0.15);
  color: var(--accent-green);
}

/* Parameters Card */
.params-card {
  grid-column: 1;
  grid-row: 2;
}

.params-grid {
  padding: 12px;
  display: flex;
  flex-direction: column;
  gap: 8px;
  max-height: 400px;
  overflow-y: auto;
}

.param-item {
  padding: 12px;
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  border-left: 3px solid var(--lenovo-red);
}

.param-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 6px;
}

.param-key {
  font-size: 12px;
  font-weight: 700;
  color: var(--lenovo-red);
  font-family: 'Consolas', monospace;
}

.param-category {
  font-size: 10px;
  font-weight: 600;
  padding: 2px 6px;
  border-radius: 4px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
}

.param-description {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
  margin-bottom: 6px;
}

.param-impact {
  font-size: 11px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.impact-label {
  color: var(--text-tertiary);
}

.impact-value {
  font-weight: 600;
}

.impact-value.high {
  color: var(--lenovo-red);
}

.impact-value.medium {
  color: var(--accent-yellow);
}

.impact-value.low {
  color: var(--accent-green);
}

/* Architecture Card */
.architecture-card {
  grid-column: 1 / 3;
  grid-row: 3;
}

.architecture-content {
  padding: 20px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.arch-layer {
  width: 100%;
  max-width: 800px;
  padding: 16px;
  border-radius: var(--radius-md);
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
}

.arch-layer.highlight {
  background: rgba(230, 63, 50, 0.05);
  border-color: rgba(230, 63, 50, 0.3);
}

.arch-layer.hardware {
  background: rgba(96, 165, 250, 0.05);
  border-color: rgba(96, 165, 250, 0.3);
}

.layer-name {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
}

.layer-items {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.layer-item {
  font-size: 12px;
  font-weight: 500;
  padding: 4px 10px;
  border-radius: 4px;
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.arch-arrow {
  font-size: 16px;
  color: var(--text-tertiary);
}
</style>
