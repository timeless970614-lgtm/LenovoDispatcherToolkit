<template>
  <div class="ppm-page">

    <!-- PPM Package Info Card -->
    <div class="card ppm-info-card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
            <polyline points="14 2 14 8 20 8"/>
            <line x1="16" y1="13" x2="8" y2="13"/>
            <line x1="16" y1="17" x2="8" y2="17"/>
            <polyline points="10 9 9 9 8 9"/>
          </svg>
          PPM Package Info
        </span>
        <span class="pkg-badge">Intel PTL · Win11 24H2</span>
      </div>

      <!-- Meta Row -->
      <div class="pkg-meta-row">
        <div class="pkg-meta-item">
          <span class="pkg-meta-label">Package Name</span>
          <span class="pkg-meta-value">Power.Settings.Processor.Intel</span>
        </div>
        <div class="pkg-meta-item">
          <span class="pkg-meta-label">Version</span>
          <span class="pkg-meta-value mono">2025.0814121440</span>
        </div>
        <div class="pkg-meta-item">
          <span class="pkg-meta-label">Owner</span>
          <span class="pkg-meta-value">SiliconVendor</span>
        </div>
        <div class="pkg-meta-item">
          <span class="pkg-meta-label">Rank</span>
          <span class="pkg-meta-value mono">0</span>
        </div>
      </div>

      <!-- Supported Processors -->
      <div class="pkg-section-title">Supported Processor Targets</div>
      <div class="cpu-targets-grid">
        <div class="cpu-target-group">
          <div class="cpu-group-label">Legacy</div>
          <div class="cpu-chip">Clovertrail <span class="cpu-model">M53</span></div>
          <div class="cpu-chip">Silvermont <span class="cpu-model">M55</span></div>
          <div class="cpu-chip">Airmont <span class="cpu-model">M76</span></div>
          <div class="cpu-chip">Haswell ULT <span class="cpu-model">M69</span></div>
          <div class="cpu-chip">Skylake <span class="cpu-model">M78/94</span></div>
          <div class="cpu-chip">Kabylake <span class="cpu-model">M142/158</span></div>
        </div>
        <div class="cpu-target-group">
          <div class="cpu-group-label">10th–11th Gen</div>
          <div class="cpu-chip">Lakefield <span class="cpu-model">M138</span></div>
          <div class="cpu-chip">Cannonlake <span class="cpu-model">M66</span></div>
          <div class="cpu-chip">Icelake <span class="cpu-model">M126</span></div>
          <div class="cpu-chip">Cometlake <span class="cpu-model">M166</span></div>
          <div class="cpu-chip">Tigerlake U <span class="cpu-model">M140</span></div>
          <div class="cpu-chip">Tigerlake H <span class="cpu-model">M141</span></div>
        </div>
        <div class="cpu-target-group">
          <div class="cpu-group-label">12th–13th Gen</div>
          <div class="cpu-chip">Rocketlake <span class="cpu-model">M167</span></div>
          <div class="cpu-chip">Alderlake S <span class="cpu-model">M151</span></div>
          <div class="cpu-chip">Alderlake P <span class="cpu-model">M154</span></div>
          <div class="cpu-chip">Alderlake N <span class="cpu-model">M190</span></div>
          <div class="cpu-chip">Raptorlake P <span class="cpu-model">M186</span></div>
          <div class="cpu-chip">Raptorlake S <span class="cpu-model">M183/191</span></div>
        </div>
        <div class="cpu-target-group">
          <div class="cpu-group-label">14th Gen+</div>
          <div class="cpu-chip">Meteorlake <span class="cpu-model">M170/172</span></div>
          <div class="cpu-chip">Lunarlake <span class="cpu-model">M189</span></div>
          <div class="cpu-chip">Arrowlake <span class="cpu-model">M197/181</span></div>
          <div class="cpu-chip">Arrowlake HX <span class="cpu-model">M198</span></div>
          <div class="cpu-chip highlight">Pantherlake U <span class="cpu-model">M204</span></div>
          <div class="cpu-chip highlight">Pantherlake H <span class="cpu-model">M204</span></div>
        </div>
      </div>

      <!-- Power Scheme Profiles -->
      <div class="pkg-section-title" style="margin-top: 20px;">Power Scheme Customizations</div>
      <div class="scheme-table">
        <div class="scheme-header">
          <span>Scheme</span>
          <span>Profile</span>
          <span>Key Parameters</span>
          <span>AC / DC</span>
        </div>
        <div class="scheme-row">
          <span class="scheme-name balanced">Balanced</span>
          <span class="profile-tag">Default</span>
          <span class="param-list">CPMinCores · ModuleUnparkPolicy · PerfAutonomousMode</span>
          <span class="ac-dc">Both</span>
        </div>
        <div class="scheme-row">
          <span class="scheme-name balanced">Balanced</span>
          <span class="profile-tag">Standby</span>
          <span class="param-list">PerfBoostMode=Disabled · EPP=100</span>
          <span class="ac-dc">Both</span>
        </div>
        <div class="scheme-row">
          <span class="scheme-name balanced">Balanced</span>
          <span class="profile-tag">Constrained</span>
          <span class="param-list">EPP DC=70</span>
          <span class="ac-dc">DC</span>
        </div>
        <div class="scheme-row">
          <span class="scheme-name battery">BetterBattery</span>
          <span class="profile-tag">Default</span>
          <span class="param-list">EPP DC=70</span>
          <span class="ac-dc">DC</span>
        </div>
        <div class="scheme-row">
          <span class="scheme-name maxperf">MaxPerformance</span>
          <span class="profile-tag">Default</span>
          <span class="param-list">EPP AC=0~33 · MinPerf AC=100 · PerfIncreasePolicy=Rocket</span>
          <span class="ac-dc">AC</span>
        </div>
        <div class="scheme-row">
          <span class="scheme-name maxperf">MaxPerformance</span>
          <span class="profile-tag">Background</span>
          <span class="param-list">EPP AC=0~25 (AoAc)</span>
          <span class="ac-dc">AC</span>
        </div>
        <div class="scheme-row highlight-row">
          <span class="scheme-name balanced">Balanced (PTL)</span>
          <span class="profile-tag">Default</span>
          <span class="param-list">EPP AC=45 · CPMinCores=4 · HeteroPolicy=0 · SoftPark=1000ms</span>
          <span class="ac-dc">Both</span>
        </div>
        <div class="scheme-row highlight-row">
          <span class="scheme-name maxperf">MaxPerf (PTL)</span>
          <span class="profile-tag">Default</span>
          <span class="param-list">EPP AC=33 · SchedulingPolicy=Prefer efficient</span>
          <span class="ac-dc">AC</span>
        </div>
      </div>

      <!-- Key Parameters Summary -->
      <div class="pkg-section-title" style="margin-top: 20px;">Key Parameter Defaults (PTL Balanced)</div>
      <div class="param-grid">
        <div class="param-card">
          <div class="param-name">EPP (P-Core)</div>
          <div class="param-val">45 <span class="param-unit">AC</span></div>
          <div class="param-sub">Energy Performance Preference</div>
        </div>
        <div class="param-card">
          <div class="param-name">EPP1 (E-Core)</div>
          <div class="param-val">45 <span class="param-unit">AC</span></div>
          <div class="param-sub">Energy Performance Preference</div>
        </div>
        <div class="param-card">
          <div class="param-name">CPMinCores</div>
          <div class="param-val">4 <span class="param-unit">cores</span></div>
          <div class="param-sub">Min unparked cores</div>
        </div>
        <div class="param-card">
          <div class="param-name">SoftParkLatency</div>
          <div class="param-val">1000 <span class="param-unit">ms</span></div>
          <div class="param-sub">Core park latency</div>
        </div>
        <div class="param-card">
          <div class="param-name">HeteroPolicy</div>
          <div class="param-val">0</div>
          <div class="param-sub">Scheduling policy</div>
        </div>
        <div class="param-card">
          <div class="param-name">CPConcurrency</div>
          <div class="param-val">95 <span class="param-unit">%</span></div>
          <div class="param-sub">Core parking concurrency</div>
        </div>
        <div class="param-card">
          <div class="param-name">IdleDemote</div>
          <div class="param-val">40 <span class="param-unit">DC</span></div>
          <div class="param-sub">Idle demote threshold</div>
        </div>
        <div class="param-card">
          <div class="param-name">IdlePromote</div>
          <div class="param-val">60 <span class="param-unit">DC</span></div>
          <div class="param-sub">Idle promote threshold</div>
        </div>
      </div>
    </div>

    <!-- Live Parameters Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px;">
            <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
          </svg>
          Live Power Settings
        </span>
        <div style="display:flex;gap:8px;align-items:center;">
          <span v-if="lastRefresh" class="refresh-time">{{ lastRefresh }}</span>
          <button class="btn-icon" @click="loadSettings" :disabled="loading" title="Refresh">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ spinning: loading }">
              <polyline points="23 4 23 10 17 10"/>
              <polyline points="1 20 1 14 7 14"/>
              <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>
            </svg>
          </button>
        </div>
      </div>

      <div v-if="loading" class="loading-state">
        <div class="spinner"></div>
        <span>Reading power settings...</span>
      </div>

      <div v-else-if="settings">
        <!-- Scheme Info -->
        <div class="scheme-info-bar">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
          Active Scheme: <strong>{{ settings.schemeName }}</strong>
          <span class="scheme-guid">{{ settings.schemeGUID }}</span>
        </div>

        <!-- Settings Table -->
        <div class="settings-table">
          <div class="settings-thead">
            <span>Parameter</span>
            <span>GUID</span>
            <span>AC Value</span>
            <span>DC Value</span>
            <span>Action</span>
          </div>

          <div v-for="param in paramList" :key="param.key" class="settings-row" :class="{ 'not-found': !settings[param.key].found }">
            <span class="param-name-cell">
              <span class="param-dot" :class="settings[param.key].found ? 'dot-active' : 'dot-inactive'"></span>
              {{ settings[param.key].name }}
            </span>
            <span class="param-guid-cell">{{ settings[param.key].guid.substring(0, 8) }}…</span>

            <span class="param-val-cell">
              <template v-if="settings[param.key].found">
                <template v-if="editingKey === param.key">
                  <input class="val-input" type="number" v-model.number="editAC" min="0" max="4294967295" />
                </template>
                <template v-else>
                  <span class="val-badge">{{ settings[param.key].acValue }}</span>
                </template>
              </template>
              <span v-else class="val-na">N/A</span>
            </span>

            <span class="param-val-cell">
              <template v-if="settings[param.key].found">
                <template v-if="editingKey === param.key">
                  <input class="val-input" type="number" v-model.number="editDC" min="0" max="4294967295" />
                </template>
                <template v-else>
                  <span class="val-badge">{{ settings[param.key].dcValue }}</span>
                </template>
              </template>
              <span v-else class="val-na">N/A</span>
            </span>

            <span class="param-action-cell">
              <template v-if="settings[param.key].found">
                <template v-if="editingKey === param.key">
                  <button class="btn-save" @click="saveEdit(param.key)" :disabled="saving">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><polyline points="20 6 9 17 4 12"/></svg>
                    Save
                  </button>
                  <button class="btn-cancel" @click="cancelEdit">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
                  </button>
                </template>
                <template v-else>
                  <button class="btn-edit" @click="startEdit(param.key)">
                    <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"/><path d="M18.5 2.5a2.121 2.121 0 0 1 3 3L12 15l-4 1 1-4 9.5-9.5z"/></svg>
                    Edit
                  </button>
                </template>
              </template>
              <span v-else class="val-na">—</span>
            </span>
          </div>
        </div>

        <!-- Status Message -->
        <transition name="fade">
          <div v-if="statusMsg" :class="['status-msg', statusOk ? 'success' : 'error']" style="margin-top:12px;">
            <svg v-if="statusOk" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
            <svg v-else width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            {{ statusMsg }}
          </div>
        </transition>
      </div>

      <div v-else class="empty-state">
        <p>Click refresh to load current power settings</p>
        <button class="btn btn-primary" @click="loadSettings" style="margin-top:12px;">Load Settings</button>
      </div>
    </div>

  </div>
</template>

<script>
export default {
  name: 'PPMDriver',
  data() {
    return {
      settings: null,
      loading: false,
      saving: false,
      editingKey: null,
      editAC: 0,
      editDC: 0,
      statusMsg: '',
      statusOk: true,
      lastRefresh: '',
      paramList: [
        { key: 'minPerf',    label: 'Min Processor State' },
        { key: 'maxPerf',    label: 'Max Processor State' },
        { key: 'cpMinCores', label: 'Min Processor Cores' },
        { key: 'epp',        label: 'EPP (P-Core)' },
        { key: 'epp1',       label: 'EPP1 (E-Core)' },
        { key: 'heteroInc',  label: 'Hetero Increase Threshold' },
        { key: 'heteroDec',  label: 'Hetero Decrease Threshold' },
        { key: 'maxFreq',    label: 'Max Frequency (P-Core)' },
        { key: 'maxFreq1',   label: 'Max Frequency (E-Core)' },
        { key: 'softPark',   label: 'Soft Park Latency' },
      ]
    }
  },
  async mounted() {
    await this.loadSettings()
  },
  methods: {
    async loadSettings() {
      this.loading = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          this.settings = await window.go.main.App.GetPPMSettings()
          const now = new Date()
          this.lastRefresh = now.toLocaleTimeString()
        }
      } catch (e) {
        this.showStatus('Failed to load settings: ' + e, false)
      } finally {
        this.loading = false
      }
    },
    startEdit(key) {
      this.editingKey = key
      this.editAC = this.settings[key].acValue
      this.editDC = this.settings[key].dcValue
    },
    cancelEdit() {
      this.editingKey = null
    },
    async saveEdit(key) {
      this.saving = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const guid = this.settings[key].guid
          const result = await window.go.main.App.SetPowerSettingRaw(guid, this.editAC, this.editDC)
          if (result === 'OK') {
            this.showStatus(`${this.settings[key].name} updated: AC=${this.editAC}, DC=${this.editDC}`, true)
            this.editingKey = null
            await this.loadSettings()
          } else {
            this.showStatus(result, false)
          }
        }
      } catch (e) {
        this.showStatus('Failed: ' + e, false)
      } finally {
        this.saving = false
      }
    },
    showStatus(msg, ok = true) {
      this.statusMsg = msg
      this.statusOk = ok
      setTimeout(() => { this.statusMsg = '' }, 5000)
    }
  }
}
</script>

<style scoped>
/* PPM Info Card */
.ppm-info-card {
  border-color: rgba(230, 63, 50, 0.2);
}

.pkg-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 3px 10px;
  background: rgba(230, 63, 50, 0.12);
  border: 1px solid rgba(230, 63, 50, 0.3);
  border-radius: 20px;
  color: var(--lenovo-red);
  letter-spacing: 0.3px;
}

.pkg-meta-row {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  margin-bottom: 20px;
  padding: 14px 16px;
  background: var(--bg-tertiary);
  border-radius: 8px;
}

.pkg-meta-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.pkg-meta-label {
  font-size: 10px;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.pkg-meta-value {
  font-size: 13px;
  font-weight: 600;
  color: var(--text-primary);
}

.pkg-meta-value.mono {
  font-family: 'Consolas', monospace;
  color: var(--lenovo-red);
}

.pkg-section-title {
  font-size: 11px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 1px;
  color: var(--text-tertiary);
  margin-bottom: 12px;
  padding-bottom: 6px;
  border-bottom: 1px solid var(--border-color);
}

/* CPU Targets Grid */
.cpu-targets-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.cpu-target-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.cpu-group-label {
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  color: var(--text-tertiary);
  margin-bottom: 4px;
  letter-spacing: 0.5px;
}

.cpu-chip {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 5px 10px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  font-size: 12px;
  color: var(--text-secondary);
  transition: var(--transition);
}

.cpu-chip:hover {
  border-color: var(--border-light);
  color: var(--text-primary);
}

.cpu-chip.highlight {
  background: rgba(230, 63, 50, 0.08);
  border-color: rgba(230, 63, 50, 0.3);
  color: var(--lenovo-red);
  font-weight: 600;
}

.cpu-model {
  font-size: 10px;
  font-family: 'Consolas', monospace;
  color: var(--text-tertiary);
  background: var(--bg-primary);
  padding: 1px 5px;
  border-radius: 3px;
}

.cpu-chip.highlight .cpu-model {
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red);
}

/* Scheme Table */
.scheme-table {
  display: flex;
  flex-direction: column;
  gap: 2px;
  border-radius: 8px;
  overflow: hidden;
  border: 1px solid var(--border-color);
}

.scheme-header {
  display: grid;
  grid-template-columns: 140px 100px 1fr 80px;
  gap: 12px;
  padding: 8px 14px;
  background: var(--bg-tertiary);
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.scheme-row {
  display: grid;
  grid-template-columns: 140px 100px 1fr 80px;
  gap: 12px;
  padding: 9px 14px;
  font-size: 12px;
  border-top: 1px solid var(--border-color);
  align-items: center;
  transition: background 0.15s;
}

.scheme-row:hover {
  background: var(--bg-tertiary);
}

.scheme-row.highlight-row {
  background: rgba(230, 63, 50, 0.04);
}

.scheme-row.highlight-row:hover {
  background: rgba(230, 63, 50, 0.08);
}

.scheme-name {
  font-weight: 600;
  font-size: 11px;
  padding: 2px 8px;
  border-radius: 4px;
  display: inline-block;
}

.scheme-name.balanced {
  background: rgba(59, 130, 246, 0.12);
  color: #3B82F6;
}

.scheme-name.battery {
  background: rgba(16, 185, 129, 0.12);
  color: #10B981;
}

.scheme-name.maxperf {
  background: rgba(230, 63, 50, 0.12);
  color: var(--lenovo-red);
}

.profile-tag {
  font-size: 11px;
  color: var(--text-secondary);
  background: var(--bg-tertiary);
  padding: 2px 7px;
  border-radius: 4px;
  display: inline-block;
}

.param-list {
  font-size: 11px;
  color: var(--text-secondary);
  font-family: 'Consolas', monospace;
}

.ac-dc {
  font-size: 11px;
  color: var(--text-tertiary);
  text-align: center;
}

/* Param Grid */
.param-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
}

.param-card {
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 12px 14px;
  transition: var(--transition);
}

.param-card:hover {
  border-color: rgba(230, 63, 50, 0.3);
}

.param-name {
  font-size: 11px;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 6px;
}

.param-val {
  font-size: 22px;
  font-weight: 700;
  color: var(--lenovo-red);
  line-height: 1;
  margin-bottom: 4px;
}

.param-unit {
  font-size: 11px;
  font-weight: 400;
  color: var(--text-tertiary);
}

.param-sub {
  font-size: 10px;
  color: var(--text-tertiary);
}

.fade-enter-active,
.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
  opacity: 0;
}

/* Live Settings Card */
.loading-state {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 24px;
  color: var(--text-secondary);
  font-size: 14px;
}

.refresh-time {
  font-size: 11px;
  color: var(--text-tertiary);
}

.scheme-info-bar {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: var(--bg-tertiary);
  border-radius: 8px;
  font-size: 13px;
  color: var(--text-secondary);
  margin-bottom: 16px;
}

.scheme-info-bar strong {
  color: var(--text-primary);
}

.scheme-guid {
  font-family: 'Consolas', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
  margin-left: 8px;
}

/* Settings Table */
.settings-table {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.settings-thead {
  display: grid;
  grid-template-columns: 200px 120px 110px 110px 140px;
  gap: 0;
  padding: 8px 14px;
  background: var(--bg-tertiary);
  font-size: 10px;
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.8px;
  color: var(--text-tertiary);
}

.settings-row {
  display: grid;
  grid-template-columns: 200px 120px 110px 110px 140px;
  gap: 0;
  padding: 10px 14px;
  border-top: 1px solid var(--border-color);
  align-items: center;
  transition: background 0.15s;
}

.settings-row:hover {
  background: var(--bg-tertiary);
}

.settings-row.not-found {
  opacity: 0.45;
}

.param-name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  color: var(--text-primary);
  font-weight: 500;
}

.param-dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  flex-shrink: 0;
}

.dot-active {
  background: var(--accent-green);
  box-shadow: 0 0 4px rgba(76, 175, 80, 0.5);
}

.dot-inactive {
  background: var(--text-tertiary);
}

.param-guid-cell {
  font-family: 'Consolas', monospace;
  font-size: 11px;
  color: var(--text-tertiary);
}

.param-val-cell {
  display: flex;
  align-items: center;
}

.val-badge {
  font-family: 'Consolas', monospace;
  font-size: 13px;
  font-weight: 600;
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.08);
  padding: 3px 10px;
  border-radius: 5px;
  min-width: 50px;
  text-align: center;
}

.val-na {
  font-size: 12px;
  color: var(--text-tertiary);
}

.val-input {
  width: 80px;
  padding: 4px 8px;
  background: var(--bg-primary);
  border: 1px solid var(--lenovo-red);
  border-radius: 5px;
  color: var(--text-primary);
  font-size: 13px;
  font-family: 'Consolas', monospace;
  outline: none;
}

.val-input:focus {
  box-shadow: 0 0 0 2px rgba(230, 63, 50, 0.2);
}

.param-action-cell {
  display: flex;
  align-items: center;
  gap: 6px;
}

.btn-edit {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 12px;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.btn-edit:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  background: rgba(230, 63, 50, 0.06);
}

.btn-save {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  padding: 5px 12px;
  background: var(--lenovo-red);
  border: none;
  border-radius: 6px;
  color: white;
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.btn-save:hover:not(:disabled) {
  opacity: 0.85;
}

.btn-save:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.btn-cancel {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-tertiary);
  cursor: pointer;
  transition: var(--transition);
  font-family: inherit;
}

.btn-cancel:hover {
  border-color: #F44336;
  color: #F44336;
}

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

@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}

.spinning {
  animation: spin 0.8s linear infinite;
}
</style>
