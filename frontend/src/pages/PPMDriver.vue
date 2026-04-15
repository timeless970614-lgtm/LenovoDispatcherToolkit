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
          <p class="card-subtitle">CPU and platform details</p>
        </div>
      </div>
      <div class="platform-content">
        <div class="cpu-info">
          <div class="cpu-name">{{ platformInfo.cpuName || 'Loading...' }}</div>
          <div class="cpu-details">
            <span class="cpu-detail-item">{{ platformInfo.cores }} Cores</span>
            <span class="cpu-detail-divider">|</span>
            <span class="cpu-detail-item">{{ platformInfo.threads }} Threads</span>
            <span class="cpu-detail-divider">|</span>
            <span class="cpu-detail-item">{{ platformInfo.platform }}</span>
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
      <div class="drivers-content">
        <!-- IPF Framework -->
        <div class="driver-row" v-if="ipfDriver">
          <div class="driver-label">IPF Framework</div>
          <div class="driver-value">v{{ ipfDriver.version }}</div>
          <div class="driver-date">{{ formatDate(ipfDriver.date) }}</div>
        </div>
        <div class="driver-row loading-row" v-else-if="!ipfDriver && !loading">
          <div class="driver-label">IPF Framework</div>
          <div class="driver-value na">N/A</div>
        </div>
        
        <!-- DTT -->
        <div class="driver-row" v-if="dttDriver">
          <div class="driver-label">DTT</div>
          <div class="driver-value">v{{ dttDriver.version }}</div>
          <div class="driver-date">{{ formatDate(dttDriver.date) }}</div>
        </div>
        <div class="driver-row loading-row" v-else-if="!dttDriver && !loading">
          <div class="driver-label">DTT</div>
          <div class="driver-value na">N/A</div>
        </div>
        
        <!-- PPM Provisioning -->
        <div class="driver-row" v-if="ppmProvisioning">
          <div class="driver-label">PPM Provisioning</div>
          <div class="driver-value">v{{ ppmProvisioning.version }}</div>
          <div class="driver-date">{{ formatDate(ppmProvisioning.date) }}</div>
        </div>
        <div class="driver-row loading-row" v-else-if="!ppmProvisioning && !loading">
          <div class="driver-label">PPM Provisioning</div>
          <div class="driver-value na">N/A</div>
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
          <p class="card-subtitle">Key power management parameters</p>
        </div>
      </div>
      <div class="params-content">
        <div class="param-table">
          <div class="param-header-row">
            <div class="param-col-key">Parameter</div>
            <div class="param-col-desc">Description</div>
            <div class="param-col-impact">Impact</div>
          </div>
          <div v-for="param in ppmParameters" :key="param.key" class="param-row">
            <div class="param-col-key">
              <span class="param-key">{{ param.key }}</span>
              <span class="param-category">{{ param.category }}</span>
            </div>
            <div class="param-col-desc">{{ param.description }}</div>
            <div class="param-col-impact">
              <span :class="['impact-badge', param.impactLevel]">{{ param.impact }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
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
      ipfDriver: null,
      dttDriver: null,
      ppmProvisioning: null,
      ppmParameters: [
        {
          key: 'PROCTHROTTLEMIN',
          category: 'Performance',
          description: 'Minimum processor performance state (%) - Lower bound for CPU frequency scaling',
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
          description: 'Energy Performance Preference (0-255) - Power vs performance balance hint',
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
            // Find IPF Framework (Processor Participant or Generic Participant)
            this.ipfDriver = drivers.find(d => 
              d.name.includes('Innovation Platform Framework Manager') ||
              d.name.includes('Processor Participant')
            ) || drivers.find(d => d.name.includes('Innovation Platform'))
            
            // Find DTT
            this.dttDriver = drivers.find(d => 
              d.name.includes('Dynamic Tuning Technology') && !d.name.includes('Updater')
            )
            
            // Find PPM Provisioning
            this.ppmProvisioning = drivers.find(d => 
              d.name.includes('PPM Provisioning')
            )
          }
        }
      } catch (e) {
        console.error('Failed to load PPM info:', e)
      } finally {
        this.loading = false
      }
    },
    formatDate(dateStr) {
      if (!dateStr) return ''
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
  display: flex;
  flex-direction: column;
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
  flex-shrink: 0;
}

.platform-content {
  padding: 24px 20px;
}

.cpu-info {
  text-align: center;
}

.cpu-name {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 8px;
}

.cpu-details {
  font-size: 13px;
  color: var(--text-secondary);
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
}

.cpu-detail-divider {
  color: var(--text-tertiary);
}

/* Drivers Card */
.drivers-card {
  flex-shrink: 0;
}

.drivers-content {
  padding: 16px 20px;
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.driver-row {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  gap: 16px;
}

.driver-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
  min-width: 140px;
}

.driver-value {
  font-size: 13px;
  font-weight: 700;
  color: var(--accent-green);
  font-family: 'Consolas', monospace;
}

.driver-value.na {
  color: var(--text-tertiary);
}

.driver-date {
  margin-left: auto;
  font-size: 12px;
  color: var(--text-tertiary);
}

.loading-row {
  opacity: 0.6;
}

/* Parameters Card */
.params-card {
  flex: 1;
  min-height: 300px;
}

.params-content {
  padding: 16px;
  overflow-x: auto;
}

.param-table {
  width: 100%;
  border-collapse: collapse;
}

.param-header-row,
.param-row {
  display: grid;
  grid-template-columns: 200px 1fr 280px;
  gap: 12px;
  padding: 10px 12px;
  align-items: center;
}

.param-header-row {
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  font-size: 11px;
  font-weight: 700;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.param-row {
  border-bottom: 1px solid var(--border-color);
}

.param-row:last-child {
  border-bottom: none;
}

.param-col-key {
  display: flex;
  flex-direction: column;
  gap: 4px;
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
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 3px;
  width: fit-content;
}

.param-col-desc {
  font-size: 12px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.param-col-impact {
  text-align: right;
}

.impact-badge {
  font-size: 11px;
  font-weight: 600;
  padding: 4px 8px;
  border-radius: 4px;
  display: inline-block;
}

.impact-badge.high {
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red);
}

.impact-badge.medium {
  background: rgba(251, 191, 36, 0.15);
  color: var(--accent-yellow);
}

.impact-badge.low {
  background: rgba(74, 222, 128, 0.15);
  color: var(--accent-green);
}
</style>
