<template>
  <div class="ppm-page">
    <!-- Combined Platform & Drivers Card -->
    <div class="card info-card">
      <div class="card-header">
        <div class="card-icon">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="4" y="4" width="16" height="16" rx="2"/>
            <rect x="9" y="9" width="6" height="6"/>
            <line x1="9" y1="2" x2="9" y2="4"/>
            <line x1="15" y1="2" x2="15" y2="4"/>
            <line x1="9" y1="20" x2="9" y2="22"/>
            <line x1="15" y1="20" x2="15" y2="22"/>
          </svg>
        </div>
        <div class="card-title-info">
          <h2>Platform & PPM Drivers</h2>
        </div>
        <button class="btn-refresh" @click="loadPPMInfo" :disabled="loading">
          <svg :class="{ 'spin': loading }" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M23 4v6h-6M1 20v-6h6"/>
            <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
          </svg>
          <span>{{ loading ? 'Scanning...' : 'Refresh' }}</span>
        </button>
      </div>
      
      <div class="info-content">
        <!-- CPU Info Section -->
        <div class="section cpu-section">
          <div class="section-label">CPU</div>
          <div class="cpu-row">
            <span class="cpu-name">{{ platformInfo.cpuName || 'Loading...' }}</span>
            <span class="cpu-cores">{{ platformInfo.cores }} Cores | {{ platformInfo.threads }} Threads</span>
          </div>
        </div>

        <div class="section-divider"></div>

        <!-- Drivers Section -->
        <div class="section drivers-section">
          <div class="section-label">PPM Drivers</div>
          <div class="drivers-grid">
            <!-- IPF Framework -->
            <div class="driver-item" v-if="ipfDriver">
              <div class="driver-name">IPF Framework</div>
              <div class="driver-version">{{ ipfDriver.version }}</div>
              <div class="driver-date">{{ formatDate(ipfDriver.date) }}</div>
            </div>
            <div class="driver-item na" v-else>
              <div class="driver-name">IPF Framework</div>
              <div class="driver-version">N/A</div>
            </div>

            <!-- DTT -->
            <div class="driver-item" v-if="dttDriver">
              <div class="driver-name">DTT</div>
              <div class="driver-version">{{ dttDriver.version }}</div>
              <div class="driver-date">{{ formatDate(dttDriver.date) }}</div>
            </div>
            <div class="driver-item na" v-else>
              <div class="driver-name">DTT</div>
              <div class="driver-version">N/A</div>
            </div>

            <!-- PPM Provisioning -->
            <div class="driver-item" v-if="ppmProvisioning">
              <div class="driver-name">PPM Provisioning</div>
              <div class="driver-version">{{ ppmProvisioning.version }}</div>
              <div class="driver-date">{{ formatDate(ppmProvisioning.date) }}</div>
            </div>
            <div class="driver-item na" v-else>
              <div class="driver-name">PPM Provisioning</div>
              <div class="driver-version">N/A</div>
            </div>
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
            // Find IPF Framework Manager
            this.ipfDriver = drivers.find(d => 
              d.name && d.name.includes('Framework Manager')
            )
            
            // Find DTT (Intel Dynamic Tuning Technology, not Updater)
            this.dttDriver = drivers.find(d => 
              d.name && d.name.includes('Dynamic Tuning') && !d.name.includes('Updater')
            )
            
            // Find PPM Provisioning
            this.ppmProvisioning = drivers.find(d => 
              d.name && d.name.includes('PPM Provisioning')
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
      // PowerShell date format: "20241216000000.******+***"
      if (dateStr.length >= 8) {
        return dateStr.substring(0, 4) + '-' + dateStr.substring(4, 6) + '-' + dateStr.substring(6, 8)
      }
      // Fallback: try to match YYYYMMDD format
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
  gap: 12px;
  padding: 12px;
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
  padding: 14px 20px;
  border-bottom: 1px solid var(--border-color);
  background: var(--bg-tertiary);
}

.card-icon {
  width: 36px;
  height: 36px;
  border-radius: var(--radius-md);
  background: rgba(230, 63, 50, 0.1);
  color: var(--lenovo-red);
  display: flex;
  align-items: center;
  justify-content: center;
}

.card-title-info h2 {
  margin: 0;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-title-info .card-subtitle {
  margin: 2px 0 0 0;
  font-size: 11px;
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
  font-size: 11px;
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

/* Combined Info Card */
.info-card {
  flex-shrink: 0;
}

.info-content {
  padding: 16px 20px;
}

.section {
  display: flex;
  align-items: center;
  gap: 16px;
}

.section-label {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  min-width: 100px;
}

.section-divider {
  height: 1px;
  background: var(--border-color);
  margin: 14px 0;
}

/* CPU Section */
.cpu-section {
  padding: 8px 0;
}

.cpu-row {
  display: flex;
  align-items: center;
  gap: 16px;
  flex: 1;
}

.cpu-name {
  font-size: 15px;
  font-weight: 700;
  color: var(--text-primary);
}

.cpu-cores {
  font-size: 12px;
  color: var(--text-secondary);
  padding: 4px 10px;
  background: var(--bg-secondary);
  border-radius: var(--radius-sm);
}

/* Drivers Section */
.drivers-section {
  flex-direction: column;
  align-items: flex-start;
  gap: 10px;
}

.drivers-section .section-label {
  margin-bottom: 4px;
}

.drivers-grid {
  display: flex;
  gap: 16px;
  width: 100%;
}

.driver-item {
  flex: 1;
  padding: 14px 16px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.driver-item.na {
  opacity: 0.5;
}

.driver-name {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
}

.driver-version {
  font-size: 14px;
  font-weight: 700;
  color: var(--accent-green);
  font-family: 'Consolas', monospace;
}

.driver-item.na .driver-version {
  color: var(--text-tertiary);
}

.driver-date {
  font-size: 10px;
  color: var(--text-tertiary);
  margin-top: 2px;
}

/* Parameters Card */
.params-card {
  flex: 1;
  min-height: 300px;
}

.params-content {
  padding: 12px 16px;
  overflow-x: auto;
}

.param-table {
  width: 100%;
}

.param-header-row,
.param-row {
  display: grid;
  grid-template-columns: 200px 1fr 280px;
  gap: 12px;
  padding: 8px 12px;
  align-items: center;
}

.param-header-row {
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
  font-size: 10px;
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
  gap: 3px;
}

.param-key {
  font-size: 11px;
  font-weight: 700;
  color: var(--lenovo-red);
  font-family: 'Consolas', monospace;
}

.param-category {
  font-size: 9px;
  font-weight: 600;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  padding: 2px 6px;
  border-radius: 3px;
  width: fit-content;
}

.param-col-desc {
  font-size: 11px;
  color: var(--text-secondary);
  line-height: 1.4;
}

.param-col-impact {
  text-align: right;
}

.impact-badge {
  font-size: 10px;
  font-weight: 600;
  padding: 3px 8px;
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
