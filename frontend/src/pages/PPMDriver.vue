<template>
  <div class="ppm-page">
    <!-- PPM Parameters Card -->
    <div class="card params-card">
      <!-- Header -->
      <div class="card-header ppm-header">
        <div class="card-title-info">
          <h2>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px;vertical-align:middle;color:var(--lenovo-red);">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
            </svg>
            Processor Power Management Parameters
          </h2>
          <p class="card-subtitle">Arrow Lake HX · PPMP-ARL-v1007.20250118 · Family 6 Model 198</p>
          <!-- Driver Paths -->
          <div class="driver-paths" v-if="ppmProvisioning && (ppmProvisioning.ppkgPath || ppmProvisioning.infPath)">
            <div class="path-row">
              <span class="path-label">PPM Package</span>
              <code class="path-value">{{ ppmProvisioning.ppkgPath || ppmProvisioning.infPath }}</code>
            </div>
          </div>
        </div>
        <div class="cpu-badge">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="4" y="4" width="16" height="16" rx="2"/>
            <rect x="9" y="9" width="6" height="6"/>
          </svg>
          {{ platformInfo.cpuName || 'Intel Core Ultra' }}
        </div>
      </div>

      <!-- Current Active Indicator -->
      <div class="active-indicator-bar">
        <div class="indicator-left">
          <span class="indicator-dot"></span>
          <span class="indicator-label">Current Viewing:</span>
          <span class="indicator-scheme">{{ activeScheme }}</span>
          <span class="indicator-arrow">/</span>
          <span class="indicator-profile">{{ activeProfile }}</span>
        </div>
        <div class="indicator-right">
          <span class="indicator-hint">AC Power</span>
          <span class="indicator-sep">|</span>
          <span class="indicator-hint">DC Power</span>
        </div>
      </div>

      <div class="params-content">
        <!-- Scheme Tabs -->
        <div class="scheme-tabs">
          <button 
            v-for="scheme in schemes" 
            :key="scheme.id"
            :class="['scheme-tab', { active: activeScheme === scheme.id }]"
            @click="activeScheme = scheme.id; activeProfile = 'Default'"
          >
            <span class="scheme-icon" v-if="scheme.id === 'Balanced'">⚖</span>
            <span class="scheme-icon" v-else-if="scheme.id === 'HighPerformance'">⚡</span>
            <span class="scheme-icon" v-else-if="scheme.id === 'BetterBatteryLifeOverlay'">🔋</span>
            <span class="scheme-icon" v-else>🚀</span>
            {{ scheme.name }}
          </button>
        </div>

        <!-- Profile Tabs -->
        <div class="profile-tabs">
          <span class="profile-tabs-label">Profile:</span>
          <button 
            v-for="profile in getProfilesForScheme(activeScheme)" 
            :key="profile"
            :class="['profile-tab', { active: activeProfile === profile }]"
            @click="activeProfile = profile"
          >
            {{ profile }}
          </button>
        </div>

        <!-- Parameters Table -->
        <div class="params-table-container">
          <table class="params-table">
            <thead>
              <tr>
                <th class="col-param">
                  <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:4px;vertical-align:middle;">
                    <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
                  </svg>
                  Parameter
                </th>
                <th class="col-ac">
                  <span class="th-ac-dot"></span> AC
                </th>
                <th class="col-dc">
                  <span class="th-dc-dot"></span> DC
                </th>
                <th class="col-desc">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="(param, idx) in getCurrentParams" :key="param.key" :class="['param-row', { 'row-alt': idx % 2 === 1 }]">
                <td class="col-param">
                  <span class="param-key">{{ param.key }}</span>
                </td>
                <td class="col-ac">
                  <span :class="['value-pill', 'ac-pill', { filled: param.ac && param.ac !== '-' }]">{{ formatValue(param.ac) }}</span>
                </td>
                <td class="col-dc">
                  <span :class="['value-pill', 'dc-pill', { filled: param.dc && param.dc !== '-' }]">{{ formatValue(param.dc) }}</span>
                </td>
                <td class="col-desc">{{ param.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- ===== PPM Tuner Module ===== -->
        <div class="tuner-section">
          <div class="tuner-header">
            <div class="tuner-header-left">
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;color:var(--lenovo-red);">
                <circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 00.33 1.82l.06.06a2 2 0 010 2.83 2 2 0 01-2.83 0l-.06-.06a1.65 1.65 0 00-1.82-.33 1.65 1.65 0 00-1 1.51V21a2 2 0 01-2 2 2 2 0 01-2-2v-.09A1.65 1.65 0 009 19.4a1.65 1.65 0 00-1.82.33l-.06.06a2 2 0 01-2.83 0 2 2 0 010-2.83l.06-.06A1.65 1.65 0 004.68 15a1.65 1.65 0 00-1.51-1H3a2 2 0 01-2-2 2 2 0 012-2h.09A1.65 1.65 0 004.6 9a1.65 1.65 0 00-.33-1.82l-.06-.06a2 2 0 010-2.83 2 2 0 012.83 0l.06.06A1.65 1.65 0 009 4.68a1.65 1.65 0 001-1.51V3a2 2 0 012-2 2 2 0 012 2v.09a1.65 1.65 0 001 1.51 1.65 1.65 0 001.82-.33l.06-.06a2 2 0 012.83 0 2 2 0 010 2.83l-.06.06A1.65 1.65 0 0019.4 9a1.65 1.65 0 001.51 1H21a2 2 0 012 2 2 2 0 01-2 2h-.09a1.65 1.65 0 00-1.51 1z"/>
              </svg>
              <span class="tuner-title">PPM Tuner</span>
              <span class="tuner-badge">Live Edit</span>
            </div>
            <div class="tuner-header-actions">
              <button class="btn-tuner-refresh" @click="refreshLivePPM" :disabled="tunerLoading">
                {{ tunerLoading ? '⟳ Reading...' : '🔄 Refresh' }}
              </button>
              <button class="btn-tuner-restore" @click="restoreDefaults" :disabled="tunerApplying">
                {{ tunerApplying ? 'Restoring...' : '⏪ Restore Defaults' }}
              </button>
            </div>
          </div>

          <div v-if="tunerError" class="tuner-error">
            <span>{{ tunerError }}</span>
            <button class="btn-dismiss" @click="tunerError=''">✕</button>
          </div>

          <div v-if="tunerSuccess" class="tuner-success">
            <span>{{ tunerSuccess }}</span>
          </div>

          <!-- Current scheme info -->
          <div class="tuner-scheme-info">
            <span class="scheme-info-label">Active Scheme:</span>
            <span class="scheme-info-value">{{ livePPM.schemeName || 'Loading...' }}</span>
            <span class="scheme-info-guid mono">{{ livePPM.schemeGUID }}</span>
          </div>

          <!-- Tuner Grid -->
          <div class="tuner-grid">
            <!-- EPP Card -->
            <div class="tuner-card">
              <div class="tuner-card-header">
                <span class="tuner-card-title">⚡ EPP (Energy Perf)</span>
                <span class="tuner-card-guid mono">PERFEPP / PERFEPP1</span>
              </div>
              <div class="tuner-card-desc">Energy Performance Preference: 0=max perf, 100=max efficiency</div>
              <div class="tuner-params">
                <div class="tuner-row">
                  <label class="tuner-label">P-Core EPP</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.epp && livePPM.epp.found">AC: {{ livePPM.epp.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.epp" min="0" max="100" placeholder="0-100" />
                  </div>
                </div>
                <div class="tuner-row">
                  <label class="tuner-label">E-Core EPP</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.epp1 && livePPM.epp1.found">AC: {{ livePPM.epp1.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.epp1" min="0" max="100" placeholder="0-100" />
                  </div>
                </div>
              </div>
              <button class="btn-apply" @click="applyEPP" :disabled="tunerApplying || tunerValues.epp === null">
                {{ tunerApplying ? 'Applying...' : 'Apply EPP' }}
              </button>
            </div>

            <!-- Hetero Card -->
            <div class="tuner-card">
              <div class="tuner-card-header">
                <span class="tuner-card-title">🔀 Hetero Threshold</span>
                <span class="tuner-card-guid mono">HETEROINCREASE / HETERODECREASE</span>
              </div>
              <div class="tuner-card-desc">Core type scheduling thresholds. 254=max (hard to switch).</div>
              <div class="tuner-params">
                <div class="tuner-row">
                  <label class="tuner-label">Increase Threshold</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.heteroInc && livePPM.heteroInc.found">AC: {{ livePPM.heteroInc.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.heteroInc" min="0" max="254" placeholder="0-254" />
                  </div>
                </div>
                <div class="tuner-row">
                  <label class="tuner-label">Decrease Threshold</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.heteroDec && livePPM.heteroDec.found">AC: {{ livePPM.heteroDec.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.heteroDec" min="0" max="254" placeholder="0-254" />
                  </div>
                </div>
              </div>
              <button class="btn-apply" @click="applyHetero" :disabled="tunerApplying || tunerValues.heteroInc === null">
                {{ tunerApplying ? 'Applying...' : 'Apply Hetero' }}
              </button>
            </div>

            <!-- Max Frequency Card -->
            <div class="tuner-card">
              <div class="tuner-card-header">
                <span class="tuner-card-title">📈 Max Frequency</span>
                <span class="tuner-card-guid mono">PROCFREQMAX / PROCFREQMAX1</span>
              </div>
              <div class="tuner-card-desc">Max CPU frequency cap (MHz). 0=unlimited.</div>
              <div class="tuner-params">
                <div class="tuner-row">
                  <label class="tuner-label">P-Core Max (MHz)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.maxFreq && livePPM.maxFreq.found">AC: {{ livePPM.maxFreq.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.maxFreq" min="0" max="6000" placeholder="0=unlimited" />
                  </div>
                </div>
                <div class="tuner-row">
                  <label class="tuner-label">E-Core Max (MHz)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.maxFreq1 && livePPM.maxFreq1.found">AC: {{ livePPM.maxFreq1.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.maxFreq1" min="0" max="6000" placeholder="0=unlimited" />
                  </div>
                </div>
              </div>
              <button class="btn-apply" @click="applyMaxFreq" :disabled="tunerApplying || tunerValues.maxFreq === null">
                {{ tunerApplying ? 'Applying...' : 'Apply Freq' }}
              </button>
            </div>

            <!-- Soft Park Card -->
            <div class="tuner-card">
              <div class="tuner-card-header">
                <span class="tuner-card-title">💤 Soft Park Latency</span>
                <span class="tuner-card-guid mono">SOFTPARKLATENCY</span>
              </div>
              <div class="tuner-card-desc">Core unpark latency (ms). Lower=faster wake, Higher=power saving.</div>
              <div class="tuner-params">
                <div class="tuner-row">
                  <label class="tuner-label">AC Latency (ms)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.softPark && livePPM.softPark.found">AC: {{ livePPM.softPark.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.softParkAC" min="0" max="50000" placeholder="ms" />
                  </div>
                </div>
                <div class="tuner-row">
                  <label class="tuner-label">DC Latency (ms)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.softPark && livePPM.softPark.found">DC: {{ livePPM.softPark.dcValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.softParkDC" min="0" max="50000" placeholder="ms" />
                  </div>
                </div>
              </div>
              <button class="btn-apply" @click="applySoftPark" :disabled="tunerApplying || tunerValues.softParkAC === null">
                {{ tunerApplying ? 'Applying...' : 'Apply Soft Park' }}
              </button>
            </div>

            <!-- Min/Max Processor State Card -->
            <div class="tuner-card">
              <div class="tuner-card-header">
                <span class="tuner-card-title">🔧 Processor State (%)</span>
                <span class="tuner-card-guid mono">PROCTHROTTLEMIN / PROCTHROTTLEMAX</span>
              </div>
              <div class="tuner-card-desc">Min/Max processor performance state (%). 100=full speed.</div>
              <div class="tuner-params">
                <div class="tuner-row">
                  <label class="tuner-label">Min Perf State (%)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.minPerf && livePPM.minPerf.found">AC: {{ livePPM.minPerf.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.minPerf" min="0" max="100" placeholder="0-100" />
                  </div>
                </div>
                <div class="tuner-row">
                  <label class="tuner-label">Max Perf State (%)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.maxPerf && livePPM.maxPerf.found">AC: {{ livePPM.maxPerf.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.maxPerf" min="0" max="100" placeholder="0-100" />
                  </div>
                </div>
              </div>
              <button class="btn-apply" @click="applyProcessorState" :disabled="tunerApplying || tunerValues.minPerf === null">
                {{ tunerApplying ? 'Applying...' : 'Apply State' }}
              </button>
            </div>

            <!-- Min Cores Card -->
            <div class="tuner-card">
              <div class="tuner-card-header">
                <span class="tuner-card-title">🔬 Min Cores (%)</span>
                <span class="tuner-card-guid mono">CPMINCORES</span>
              </div>
              <div class="tuner-card-desc">Minimum active cores (%). Higher=more cores always awake.</div>
              <div class="tuner-params">
                <div class="tuner-row">
                  <label class="tuner-label">Min Cores (%)</label>
                  <div class="tuner-value-group">
                    <span class="tuner-current" v-if="livePPM.cpMinCores && livePPM.cpMinCores.found">AC: {{ livePPM.cpMinCores.acValue }}</span>
                    <span class="tuner-current" v-else>—</span>
                    <input type="number" class="tuner-input" v-model.number="tunerValues.cpMinCores" min="0" max="100" placeholder="0-100" />
                  </div>
                </div>
              </div>
              <button class="btn-apply" @click="applyRawSetting('cpMinCores', '0cc5b647-c1df-4637-891a-dec35c318583')" :disabled="tunerApplying || tunerValues.cpMinCores === null">
                {{ tunerApplying ? 'Applying...' : 'Apply Min Cores' }}
              </button>
            </div>
          </div>
        </div>

        <!-- Definitions Section -->
        <div class="definitions-section">
          <div class="defs-header">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;color:var(--lenovo-red);">
              <circle cx="12" cy="12" r="10"/>
              <line x1="12" y1="16" x2="12" y2="12"/>
              <line x1="12" y1="8" x2="12.01" y2="8"/>
            </svg>
            Hetero Thresholds
          </div>
          <div class="defs-grid">
            <div class="def-item">
              <span class="def-key">HeteroDecreaseThreshold</span>
              <span class="def-value">1347440720</span>
              <span class="def-desc">P-Core → E-Core downgrade threshold</span>
            </div>
            <div class="def-item">
              <span class="def-key">HeteroIncreaseThreshold</span>
              <span class="def-value">2139259522</span>
              <span class="def-desc">E-Core → P-Core upgrade threshold</span>
            </div>
          </div>
        </div>

        <!-- Scheme Summary -->
        <div class="scheme-summary">
          <div class="summary-header">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;color:var(--lenovo-red);">
              <path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
            </svg>
            Scheme Summary
          </div>
          <div class="summary-grid">
            <div class="summary-item" v-for="s in schemeSummaries" :key="s.id">
              <div class="summary-name">{{ s.name }}</div>
              <div class="summary-details">
                <span>EPP: {{ s.epp }}</span>
                <span>CPMinCores: {{ s.cpMinCores }}</span>
                <span>HeteroPolicy: {{ s.heteroPolicy }}</span>
              </div>
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
      platformInfo: { cpuName: 'Intel Core Ultra 9 290HX Plus', cores: 24, threads: 24, platform: 'Intel', architecture: 'x64' },
      ipfDriver: { name: 'Intel Performance Framework Manager', version: 'v2.2.10204.8', date: '', infPath: '', sysPath: '' },
      dttDriver: { name: 'Intel Dynamic Tuning Technology', version: 'v9.0.11905.54373', date: '', infPath: '', sysPath: '' },
      ppmProvisioning: { name: 'PPM Provisioning', version: 'v1.0.0.97', date: '', ppkgPath: '', infPath: '' },
      activeScheme: 'Balanced',
      activeProfile: 'Default',
      // Tuner
      tunerLoading: false,
      tunerApplying: false,
      tunerError: '',
      tunerSuccess: '',
      livePPM: {
        schemeName: '',
        schemeGUID: '',
        epp: null, epp1: null,
        heteroInc: null, heteroDec: null,
        maxFreq: null, maxFreq1: null,
        softPark: null,
        minPerf: null, maxPerf: null,
        cpMinCores: null
      },
      tunerValues: {
        epp: null, epp1: null,
        heteroInc: null, heteroDec: null,
        maxFreq: null, maxFreq1: null,
        softParkAC: null, softParkDC: null,
        minPerf: null, maxPerf: null,
        cpMinCores: null
      },
      
      schemes: [
        { id: 'Balanced', name: 'Balanced' },
        { id: 'HighPerformance', name: 'High Performance' },
        { id: 'BetterBatteryLifeOverlay', name: 'Better Battery' },
        { id: 'MaxPerformanceOverlay', name: 'Max Performance' }
      ],

      schemeSummaries: [
        { id: 'Balanced', name: 'Balanced', epp: '25→50', cpMinCores: '50%→10%', heteroPolicy: '0' },
        { id: 'HighPerformance', name: 'High Perf', epp: '0→0', cpMinCores: '100%→100%', heteroPolicy: '0' },
        { id: 'BetterBatteryLifeOverlay', name: 'Better Battery', epp: '60→72', cpMinCores: '10%→4%', heteroPolicy: '0→4' },
        { id: 'MaxPerformanceOverlay', name: 'Max Perf', epp: '25→33', cpMinCores: '50%→10%', heteroPolicy: '0' }
      ],

      // Complete parameter data from customizations.xml
      ppmData: {
        Balanced: {
          Default: [
            { key: 'CPConcurrency', ac: '95', dc: '95', desc: 'Core parking concurrency threshold' },
            { key: 'CPDecreaseTime', ac: '3', dc: '3', desc: 'Core parking decrease time (seconds)' },
            { key: 'CPDistribution', ac: '90', dc: '90', desc: 'Core parking distribution' },
            { key: 'CPHeadroom', ac: '50', dc: '50', desc: 'Core parking headroom' },
            { key: 'CPIncreaseTime', ac: '1', dc: '1', desc: 'Core parking increase time (seconds)' },
            { key: 'CPMinCores', ac: '50', dc: '10', desc: 'Minimum active cores (%), AC=12 cores, DC=2-3 cores' },
            { key: 'CpLatencyHintUnpark1', ac: '50', dc: '50', desc: 'E-Core latency hint for unpark' },
            { key: 'HeteroClass1InitialPerf', ac: '100', dc: '100', desc: 'E-Core initial performance (%)' },
            { key: 'HeteroDecreaseThreshold', ac: '254', dc: '254', desc: 'P-Core→E-Core demotion threshold (254=max, hard to demote)' },
            { key: 'HeteroIncreaseThreshold', ac: '254', dc: '254', desc: 'E-Core→P-Core promotion threshold (254=max)' },
            { key: 'HeteroPolicy', ac: '0', dc: '0', desc: 'Heterogeneous policy: 0=all cores available' },
            { key: 'IdleDemoteThreshold', ac: '40', dc: '40', desc: 'Idle demote threshold' },
            { key: 'IdlePromoteThreshold', ac: '60', dc: '60', desc: 'Idle promote threshold' },
            { key: 'ModuleUnparkPolicy', ac: 'Sequential', dc: 'Sequential', desc: 'Module unpark policy' },
            { key: 'PerfDecreaseThreshold', ac: '10', dc: '-', desc: 'Performance decrease threshold' },
            { key: 'PerfEnergyPreference', ac: '25', dc: '50', desc: 'Energy Performance Preference: 0=max perf, 100=max efficiency' },
            { key: 'PerfEnergyPreference1', ac: '25', dc: '50', desc: 'E-Core EPP' },
            { key: 'PerfIncreasePolicy', ac: 'Rocket', dc: '-', desc: 'Frequency increase policy: Rocket=aggressive' },
            { key: 'PerfIncreaseThreshold', ac: '30', dc: '-', desc: 'Performance increase threshold' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SoftParkLatency', ac: '10', dc: '1000', desc: 'Soft park latency (ms): AC=fast wake, DC=slow wake' }
          ],
          Multimedia: [
            { key: 'LatencyHintEpp', ac: '19', dc: '19', desc: 'Latency hint EPP for multimedia' },
            { key: 'LatencyHintEpp1', ac: '19', dc: '19', desc: 'E-Core latency hint EPP' },
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core max frequency (MHz): 0=unlimited, DC limited to 1.5GHz' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core max frequency (MHz)' },
            { key: 'PerfEnergyPreference', ac: '100', dc: '100', desc: 'EPP: 100=maximum efficiency' },
            { key: 'PerfEnergyPreference1', ac: '100', dc: '100', desc: 'E-Core EPP' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core performance latency hint' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Efficient E', desc: 'Scheduling: Prefer efficient cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Efficient E', desc: 'Short-term scheduling policy' }
          ],
          LowLatency: [
            { key: 'PerfEnergyPreference', ac: '10', dc: '10', desc: 'EPP: 10=near-max performance' },
            { key: 'PerfEnergyPreference1', ac: '10', dc: '10', desc: 'E-Core EPP for low latency' }
          ],
          Eco: [
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core max frequency (MHz)' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core max frequency (MHz)' },
            { key: 'PerfEnergyPreference', ac: '50', dc: '70', desc: 'EPP: balanced→efficiency' },
            { key: 'PerfEnergyPreference1', ac: '50', dc: '70', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Efficient E', desc: 'Scheduling: prefer E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Efficient E', desc: 'Short-term scheduling' }
          ],
          Utility: [
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core max frequency (MHz)' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core max frequency (MHz)' },
            { key: 'PerfEnergyPreference', ac: '50', dc: '70', desc: 'EPP: balanced→efficiency' },
            { key: 'PerfEnergyPreference1', ac: '50', dc: '70', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Efficient E', desc: 'Scheduling: prefer E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Efficient E', desc: 'Short-term scheduling' }
          ],
          Background: [
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core max for background (MHz)' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core max for background (MHz)' },
            { key: 'PerfEnergyPreference', ac: '50', dc: '70', desc: 'Background EPP: power saving' },
            { key: 'PerfEnergyPreference1', ac: '50', dc: '70', desc: 'E-Core background EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: '-', desc: 'Background prefers E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: '-', desc: 'Short-term scheduling' }
          ],
          Constrained: [
            { key: 'CPConcurrency', ac: '95', dc: '95', desc: 'Core parking concurrency' },
            { key: 'CPDecreaseTime', ac: '3', dc: '3', desc: 'Core parking decrease time' },
            { key: 'CPDistribution', ac: '90', dc: '90', desc: 'Core parking distribution' },
            { key: 'CPHeadroom', ac: '50', dc: '50', desc: 'Core parking headroom' },
            { key: 'CPIncreaseTime', ac: '1', dc: '1', desc: 'Core parking increase time' },
            { key: 'CPMinCores', ac: '10', dc: '4', desc: 'Min cores: DC=4% (~1 core)' },
            { key: 'CpLatencyHintUnpark1', ac: '50', dc: '50', desc: 'E-Core latency hint' },
            { key: 'HeteroClass1InitialPerf', ac: '100', dc: '100', desc: 'E-Core initial perf' },
            { key: 'HeteroDecreaseThreshold', ac: '254', dc: '0', desc: 'DC=0 means no P-cores' },
            { key: 'HeteroDecreaseTime', ac: '-', dc: '1', desc: 'Hetero decrease time (s)' },
            { key: 'HeteroIncreaseThreshold', ac: '254', dc: '0', desc: 'DC=0 means no P-cores' },
            { key: 'HeteroIncreaseTime', ac: '-', dc: '6', desc: 'Hetero increase time (s)' },
            { key: 'HeteroPolicy', ac: '0', dc: '4', desc: 'DC=4: E-Core only, P-Cores sleep' },
            { key: 'IdleDemoteThreshold', ac: '40', dc: '40', desc: 'Idle demote threshold' },
            { key: 'IdlePromoteThreshold', ac: '60', dc: '60', desc: 'Idle promote threshold' },
            { key: 'ModuleUnparkPolicy', ac: 'Sequential', dc: 'Sequential', desc: 'Module unpark policy' },
            { key: 'PerfEnergyPreference', ac: '60', dc: '72', desc: 'EPP: power saving' },
            { key: 'PerfEnergyPreference1', ac: '60', dc: '72', desc: 'E-Core EPP' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SoftParkLatency', ac: '1000', dc: '5000', desc: 'Soft park latency: very slow wake' }
          ]
        },
        HighPerformance: {
          Default: [
            { key: 'CPConcurrency', ac: '95', dc: '95', desc: 'Core parking concurrency' },
            { key: 'CPDecreaseTime', ac: '3', dc: '3', desc: 'Core parking decrease time' },
            { key: 'CPHeadroom', ac: '50', dc: '50', desc: 'Core parking headroom' },
            { key: 'CPIncreaseTime', ac: '1', dc: '1', desc: 'Core parking increase time' },
            { key: 'CPMinCores', ac: '100', dc: '100', desc: 'All cores active (100%)' },
            { key: 'CpLatencyHintUnpark1', ac: '50', dc: '50', desc: 'E-Core latency hint' },
            { key: 'HeteroClass1InitialPerf', ac: '100', dc: '100', desc: 'E-Core initial performance' },
            { key: 'HeteroDecreaseThreshold', ac: '254', dc: '254', desc: 'Max threshold' },
            { key: 'HeteroIncreaseThreshold', ac: '254', dc: '254', desc: 'Max threshold' },
            { key: 'HeteroPolicy', ac: '0', dc: '0', desc: 'All cores available' },
            { key: 'PerfEnergyPreference', ac: '0', dc: '0', desc: 'EPP=0: Maximum performance' },
            { key: 'PerfEnergyPreference1', ac: '0', dc: '0', desc: 'E-Core EPP=0' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SoftParkLatency', ac: '1000', dc: '1000', desc: 'Soft park latency' }
          ]
        },
        BetterBatteryLifeOverlay: {
          Default: [
            { key: 'CPConcurrency', ac: '95', dc: '95', desc: 'Core parking concurrency' },
            { key: 'CPDecreaseTime', ac: '3', dc: '3', desc: 'Core parking decrease time' },
            { key: 'CPDistribution', ac: '90', dc: '90', desc: 'Core parking distribution' },
            { key: 'CPHeadroom', ac: '50', dc: '50', desc: 'Core parking headroom' },
            { key: 'CPIncreaseTime', ac: '1', dc: '1', desc: 'Core parking increase time' },
            { key: 'CPMinCores', ac: '10', dc: '4', desc: 'Min cores: AC=10%, DC=4% (~1 core)' },
            { key: 'CpLatencyHintUnpark1', ac: '50', dc: '50', desc: 'E-Core latency hint' },
            { key: 'HeteroClass1InitialPerf', ac: '100', dc: '100', desc: 'E-Core initial performance' },
            { key: 'HeteroDecreaseThreshold', ac: '254', dc: '0', desc: 'DC=0: no P-Cores' },
            { key: 'HeteroDecreaseTime', ac: '-', dc: '1', desc: 'Hetero decrease time (s)' },
            { key: 'HeteroIncreaseThreshold', ac: '254', dc: '0', desc: 'DC=0: no P-Cores' },
            { key: 'HeteroIncreaseTime', ac: '-', dc: '6', desc: 'Hetero increase time (s)' },
            { key: 'HeteroPolicy', ac: '0', dc: '4', desc: 'DC=4: E-Core only' },
            { key: 'IdleDemoteThreshold', ac: '40', dc: '40', desc: 'Idle demote threshold' },
            { key: 'IdlePromoteThreshold', ac: '60', dc: '60', desc: 'Idle promote threshold' },
            { key: 'ModuleUnparkPolicy', ac: 'Sequential', dc: 'Sequential', desc: 'Module unpark policy' },
            { key: 'PerfEnergyPreference', ac: '60', dc: '72', desc: 'EPP: power saving' },
            { key: 'PerfEnergyPreference1', ac: '60', dc: '72', desc: 'E-Core EPP' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SoftParkLatency', ac: '1000', dc: '5000', desc: 'Soft park latency: slow wake' }
          ],
          Multimedia: [
            { key: 'LatencyHintEpp', ac: '19', dc: '19', desc: 'Latency hint EPP' },
            { key: 'LatencyHintEpp1', ac: '19', dc: '19', desc: 'E-Core latency hint EPP' },
            { key: 'MaxFrequency', ac: '1500', dc: '1500', desc: 'P-Core limited to 1.5GHz' },
            { key: 'MaxFrequency1', ac: '1500', dc: '1500', desc: 'E-Core limited to 1.5GHz' },
            { key: 'PerfEnergyPreference', ac: '100', dc: '100', desc: 'EPP: max efficiency' },
            { key: 'PerfEnergyPreference1', ac: '100', dc: '100', desc: 'E-Core EPP' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SchedulingPolicy', ac: 'Efficient E', dc: 'Efficient E', desc: 'Force E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Efficient E', dc: 'Efficient E', desc: 'Short-term: E-Cores' }
          ],
          LowLatency: [
            { key: 'PerfEnergyPreference', ac: '10', dc: '10', desc: 'EPP: near-max performance' },
            { key: 'PerfEnergyPreference1', ac: '10', dc: '10', desc: 'E-Core EPP' }
          ],
          Eco: [
            { key: 'MaxFrequency', ac: '1500', dc: '1500', desc: 'P-Core 1.5GHz limit' },
            { key: 'MaxFrequency1', ac: '1500', dc: '1500', desc: 'E-Core 1.5GHz limit' },
            { key: 'PerfEnergyPreference', ac: '70', dc: '77', desc: 'EPP: power saving' },
            { key: 'PerfEnergyPreference1', ac: '70', dc: '77', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Efficient E', dc: 'Efficient E', desc: 'E-Cores only' },
            { key: 'ShortSchedulingPolicy', ac: 'Efficient E', dc: 'Efficient E', desc: 'Short-term: E-Cores' }
          ],
          Utility: [
            { key: 'MaxFrequency', ac: '1500', dc: '1500', desc: 'P-Core 1.5GHz limit' },
            { key: 'MaxFrequency1', ac: '1500', dc: '1500', desc: 'E-Core 1.5GHz limit' },
            { key: 'PerfEnergyPreference', ac: '70', dc: '77', desc: 'EPP: power saving' },
            { key: 'PerfEnergyPreference1', ac: '70', dc: '77', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Efficient E', dc: 'Efficient E', desc: 'E-Cores only' },
            { key: 'ShortSchedulingPolicy', ac: 'Efficient E', dc: 'Efficient E', desc: 'Short-term: E-Cores' }
          ],
          EntryLevelPerf: [
            { key: 'PerfEnergyPreference', ac: '-', dc: '77', desc: 'EPP: DC only' },
            { key: 'PerfEnergyPreference1', ac: '-', dc: '77', desc: 'E-Core EPP' }
          ],
          Background: [
            { key: 'MaxFrequency', ac: '1500', dc: '1500', desc: 'P-Core 1.5GHz limit' },
            { key: 'MaxFrequency1', ac: '1500', dc: '1500', desc: 'E-Core 1.5GHz limit' },
            { key: 'PerfEnergyPreference', ac: '70', dc: '77', desc: 'Background EPP' },
            { key: 'PerfEnergyPreference1', ac: '70', dc: '77', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' }
          ]
        },
        MaxPerformanceOverlay: {
          Default: [
            { key: 'CPConcurrency', ac: '95', dc: '95', desc: 'Core parking concurrency' },
            { key: 'CPDecreaseTime', ac: '3', dc: '3', desc: 'Core parking decrease time' },
            { key: 'CPDistribution', ac: '90', dc: '90', desc: 'Core parking distribution' },
            { key: 'CPHeadroom', ac: '50', dc: '50', desc: 'Core parking headroom' },
            { key: 'CPIncreaseTime', ac: '1', dc: '1', desc: 'Core parking increase time' },
            { key: 'CPMinCores', ac: '50', dc: '10', desc: 'Min cores: AC=50%, DC=10%' },
            { key: 'CpLatencyHintUnpark1', ac: '50', dc: '50', desc: 'E-Core latency hint' },
            { key: 'HeteroClass1InitialPerf', ac: '100', dc: '100', desc: 'E-Core initial performance' },
            { key: 'HeteroDecreaseThreshold', ac: '254', dc: '254', desc: 'Max threshold' },
            { key: 'HeteroIncreaseThreshold', ac: '254', dc: '254', desc: 'Max threshold' },
            { key: 'HeteroPolicy', ac: '0', dc: '0', desc: 'All cores available' },
            { key: 'IdleDemoteThreshold', ac: '40', dc: '40', desc: 'Idle demote threshold' },
            { key: 'IdlePromoteThreshold', ac: '60', dc: '60', desc: 'Idle promote threshold' },
            { key: 'ModuleUnparkPolicy', ac: 'Sequential', dc: 'Sequential', desc: 'Module unpark policy' },
            { key: 'PerfEnergyPreference', ac: '25', dc: '33', desc: 'EPP: biased to performance' },
            { key: 'PerfEnergyPreference1', ac: '25', dc: '33', desc: 'E-Core EPP' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SoftParkLatency', ac: '10', dc: '10', desc: 'Fast wake (10ms)' }
          ],
          Multimedia: [
            { key: 'LatencyHintEpp', ac: '19', dc: '19', desc: 'Latency hint EPP' },
            { key: 'LatencyHintEpp1', ac: '19', dc: '19', desc: 'E-Core latency hint EPP' },
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core: AC unlimited, DC 1.5GHz' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core: AC unlimited, DC 1.5GHz' },
            { key: 'PerfEnergyPreference', ac: '100', dc: '100', desc: 'EPP: max efficiency' },
            { key: 'PerfEnergyPreference1', ac: '100', dc: '100', desc: 'E-Core EPP' },
            { key: 'PerfLatencyHint', ac: '100', dc: '100', desc: 'Performance latency hint' },
            { key: 'PerfLatencyHint1', ac: '100', dc: '100', desc: 'E-Core latency hint' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Prefer E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Short-term: prefer E-Cores' }
          ],
          LowLatency: [
            { key: 'PerfEnergyPreference', ac: '10', dc: '10', desc: 'EPP: near-max performance' },
            { key: 'PerfEnergyPreference1', ac: '10', dc: '10', desc: 'E-Core EPP' }
          ],
          Eco: [
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core max frequency' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core max frequency' },
            { key: 'PerfEnergyPreference', ac: '50', dc: '50', desc: 'EPP: balanced' },
            { key: 'PerfEnergyPreference1', ac: '50', dc: '50', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Prefer E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Short-term: prefer E-Cores' }
          ],
          Utility: [
            { key: 'MaxFrequency', ac: '0', dc: '1500', desc: 'P-Core max frequency' },
            { key: 'MaxFrequency1', ac: '0', dc: '1500', desc: 'E-Core max frequency' },
            { key: 'PerfEnergyPreference', ac: '50', dc: '50', desc: 'EPP: balanced' },
            { key: 'PerfEnergyPreference1', ac: '50', dc: '50', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Prefer E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Short-term: prefer E-Cores' }
          ],
          Background: [
            { key: 'MaxFrequency', ac: '0', dc: '0', desc: 'Background: unlimited' },
            { key: 'MaxFrequency1', ac: '0', dc: '0', desc: 'E-Core: unlimited' },
            { key: 'PerfEnergyPreference', ac: '50', dc: '50', desc: 'Background EPP: balanced' },
            { key: 'PerfEnergyPreference1', ac: '50', dc: '50', desc: 'E-Core EPP' },
            { key: 'ResourcePriority', ac: '50', dc: '50', desc: 'Resource priority' },
            { key: 'SchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Prefer E-Cores' },
            { key: 'ShortSchedulingPolicy', ac: 'Prefer E', dc: 'Prefer E', desc: 'Short-term: prefer E-Cores' }
          ]
        }
      }
    }
  },
  mounted() {
    this.loadPPMInfo()
    this.refreshLivePPM()
  },
  computed: {
    getCurrentParams() {
      const schemeData = this.ppmData[this.activeScheme]
      if (!schemeData) return []
      const profileData = schemeData[this.activeProfile]
      return profileData || []
    }
  },
  methods: {
    getProfilesForScheme(scheme) {
      const schemeData = this.ppmData[scheme]
      return schemeData ? Object.keys(schemeData) : ['Default']
    },
    formatValue(val) {
      if (val === '-' || val === undefined || val === null) return '—'
      return val
    },
    async loadPPMInfo() {
      this.loading = true
      try {
        if (window.go && window.go.main && window.go.main.App) {
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
          const drivers = await window.go.main.App.GetPPMDrivers()
          if (drivers && drivers.length > 0) {
            this.ipfDriver = drivers.find(d => d.name && d.name.includes('Framework Manager'))
            this.dttDriver = drivers.find(d => d.name && d.name.includes('Dynamic Tuning') && !d.name.includes('Updater'))
            this.ppmProvisioning = drivers.find(d => d.name && d.name.includes('PPM Provisioning'))
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
      if (dateStr.length >= 8) {
        return dateStr.substring(0, 4) + '-' + dateStr.substring(4, 6) + '-' + dateStr.substring(6, 8)
      }
      const match = dateStr.match(/^(\d{4})(\d{2})(\d{2})/)
      if (match) {
        return `${match[1]}-${match[2]}-${match[3]}`
      }
      return dateStr
    },
    // ===== PPM Tuner Methods =====
    async refreshLivePPM() {
      this.tunerLoading = true
      this.tunerError = ''
      try {
        if (window.go && window.go.main && window.go.main.App) {
          const raw = await window.go.main.App.GetPPMSettings()
          if (raw) {
            this.livePPM = raw
            // Sync current values into editor fields
            if (raw.epp && raw.epp.found) this.tunerValues.epp = raw.epp.acValue
            if (raw.epp1 && raw.epp1.found) this.tunerValues.epp1 = raw.epp1.acValue
            if (raw.heteroInc && raw.heteroInc.found) this.tunerValues.heteroInc = raw.heteroInc.acValue
            if (raw.heteroDec && raw.heteroDec.found) this.tunerValues.heteroDec = raw.heteroDec.acValue
            if (raw.maxFreq && raw.maxFreq.found) this.tunerValues.maxFreq = raw.maxFreq.acValue
            if (raw.maxFreq1 && raw.maxFreq1.found) this.tunerValues.maxFreq1 = raw.maxFreq1.acValue
            if (raw.softPark && raw.softPark.found) {
              this.tunerValues.softParkAC = raw.softPark.acValue
              this.tunerValues.softParkDC = raw.softPark.dcValue
            }
            if (raw.minPerf && raw.minPerf.found) this.tunerValues.minPerf = raw.minPerf.acValue
            if (raw.maxPerf && raw.maxPerf.found) this.tunerValues.maxPerf = raw.maxPerf.acValue
            if (raw.cpMinCores && raw.cpMinCores.found) this.tunerValues.cpMinCores = raw.cpMinCores.acValue
          }
        }
      } catch (e) {
        this.tunerError = 'Failed to read PPM: ' + (e.message || e)
      } finally {
        this.tunerLoading = false
      }
    },
    showTunerSuccess(msg) {
      this.tunerSuccess = msg
      setTimeout(() => { this.tunerSuccess = '' }, 3000)
      setTimeout(() => this.refreshLivePPM(), 500)
    },
    async applyEPP() {
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const result = await window.go.main.App.ApplyEPP(this.tunerValues.epp, this.tunerValues.epp1)
        if (result && result.startsWith('OK')) {
          this.showTunerSuccess('✅ EPP applied: P-Core=' + this.tunerValues.epp + ' E-Core=' + this.tunerValues.epp1)
        } else {
          this.tunerError = result || 'Failed'
        }
      } catch (e) {
        this.tunerError = 'EPP error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
    },
    async applyHetero() {
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const result = await window.go.main.App.ApplyHetero(this.tunerValues.heteroInc, this.tunerValues.heteroDec)
        if (result && result.startsWith('OK')) {
          this.showTunerSuccess('✅ Hetero applied: Inc=' + this.tunerValues.heteroInc + ' Dec=' + this.tunerValues.heteroDec)
        } else {
          this.tunerError = result || 'Failed'
        }
      } catch (e) {
        this.tunerError = 'Hetero error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
    },
    async applyMaxFreq() {
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const result = await window.go.main.App.ApplyMaxFrequency(this.tunerValues.maxFreq, this.tunerValues.maxFreq1)
        if (result && result.startsWith('OK')) {
          this.showTunerSuccess('✅ Max Frequency applied: P=' + this.tunerValues.maxFreq + ' E=' + this.tunerValues.maxFreq1 + ' MHz')
        } else {
          this.tunerError = result || 'Failed'
        }
      } catch (e) {
        this.tunerError = 'Freq error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
    },
    async applySoftPark() {
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const result = await window.go.main.App.ApplySoftParkLatency(this.tunerValues.softParkAC, this.tunerValues.softParkDC)
        if (result && result.startsWith('OK')) {
          this.showTunerSuccess('✅ Soft Park Latency applied: AC=' + this.tunerValues.softParkAC + 'ms DC=' + this.tunerValues.softParkDC + 'ms')
        } else {
          this.tunerError = result || 'Failed'
        }
      } catch (e) {
        this.tunerError = 'Soft Park error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
    },
    async applyProcessorState() {
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const minResult = await window.go.main.App.SetPowerSettingRaw('893dee8e-2bef-41e0-89c6-b55d092996a4', this.tunerValues.minPerf, this.tunerValues.minPerf)
        const maxResult = await window.go.main.App.SetPowerSettingRaw('bc5038f7-23e0-4960-96da-33abaf5935ec', this.tunerValues.maxPerf, this.tunerValues.maxPerf)
        if (minResult.startsWith('OK') && maxResult.startsWith('OK')) {
          this.showTunerSuccess('✅ Processor State applied: Min=' + this.tunerValues.minPerf + '% Max=' + this.tunerValues.maxPerf + '%')
        } else {
          this.tunerError = minResult + ' | ' + maxResult
        }
      } catch (e) {
        this.tunerError = 'State error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
    },
    async applyRawSetting(key, guid) {
      const val = this.tunerValues[key]
      if (val === null) return
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const result = await window.go.main.App.SetPowerSettingRaw(guid, val, val)
        if (result.startsWith('OK')) {
          this.showTunerSuccess('✅ ' + key + ' applied: ' + val)
        } else {
          this.tunerError = result || 'Failed'
        }
      } catch (e) {
        this.tunerError = key + ' error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
    },
    async restoreDefaults() {
      this.tunerApplying = true
      this.tunerError = ''
      try {
        const result = await window.go.main.App.RestoreDefaults()
        if (result && result.startsWith('OK')) {
          this.showTunerSuccess('✅ Default power settings restored')
        } else {
          this.tunerError = result || 'Failed'
        }
      } catch (e) {
        this.tunerError = 'Restore error: ' + (e.message || e)
      } finally {
        this.tunerApplying = false
      }
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

/* ===== Card ===== */
.card {
  background: var(--bg-card);
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-color);
  overflow: hidden;
}

.ppm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  background: var(--bg-tertiary);
  border-bottom: 1px solid var(--border-color);
}

.card-title-info h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary);
  display: flex;
  align-items: center;
}

.card-title-info .card-subtitle {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--text-secondary);
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 0.02em;
}

.cpu-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: rgba(230, 63, 50, 0.1);
  border: 1px solid rgba(230, 63, 50, 0.2);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: var(--lenovo-red);
  white-space: nowrap;
}

/* ===== Active Indicator Bar ===== */
.active-indicator-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 24px;
  background: var(--bg-secondary);
  border-bottom: 1px solid var(--border-color);
}

.indicator-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.indicator-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: var(--accent-green);
  box-shadow: 0 0 6px rgba(74, 222, 128, 0.5);
  animation: pulse-dot 2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { box-shadow: 0 0 6px rgba(74, 222, 128, 0.5); }
  50% { box-shadow: 0 0 12px rgba(74, 222, 128, 0.8); }
}

.indicator-label {
  font-size: 12px;
  color: var(--text-secondary);
}

.indicator-scheme {
  font-size: 13px;
  font-weight: 700;
  color: var(--lenovo-red);
}

.indicator-arrow {
  color: var(--text-tertiary);
  font-weight: 300;
}

.indicator-profile {
  font-size: 13px;
  font-weight: 600;
  color: var(--lenovo-red);
}

.indicator-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.indicator-hint {
  font-size: 11px;
  color: var(--text-tertiary);
  font-weight: 500;
}

.indicator-sep {
  color: var(--border-color);
  font-weight: 300;
}

/* ===== Content ===== */
.params-card {
  /* no flex: 1 — let content determine height for parent scroll */
  overflow: visible;
  min-height: 400px;
}

.params-content {
  padding: 16px 20px 20px;
}

/* ===== Scheme Tabs ===== */
.scheme-tabs {
  display: flex;
  gap: 8px;
  margin-bottom: 10px;
}

.scheme-tab {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 10px 18px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}

.scheme-icon {
  font-size: 14px;
}

.scheme-tab:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
}

.scheme-tab.active {
  background: var(--lenovo-red);
  border-color: var(--lenovo-red);
  color: #fff;
  box-shadow: 0 2px 8px rgba(230, 63, 50, 0.3);
}

/* ===== Profile Tabs ===== */
.profile-tabs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 16px;
  padding: 8px 12px;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.profile-tabs-label {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-right: 4px;
}

.profile-tab {
  padding: 4px 12px;
  border: 1px solid transparent;
  border-radius: 14px;
  background: transparent;
  color: var(--text-tertiary);
  font-size: 10px;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
}

.profile-tab:hover {
  color: var(--text-secondary);
  background: var(--bg-tertiary);
}

.profile-tab.active {
  background: rgba(230, 63, 50, 0.1);
  border-color: rgba(230, 63, 50, 0.3);
  color: var(--lenovo-red);
  font-weight: 700;
}

/* ===== Parameters Table ===== */
.params-table-container {
  overflow-x: auto;
  margin-bottom: 16px;
  border: 1px solid var(--border-color);
  border-radius: 8px;
  overflow: hidden;
}

.params-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
}

.params-table thead {
  background: var(--bg-tertiary);
}

.params-table th {
  padding: 10px 14px;
  text-align: left;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-size: 10px;
  border-bottom: 2px solid var(--border-color);
}

.params-table td {
  padding: 9px 14px;
  border-bottom: 1px solid var(--border-color);
  vertical-align: middle;
}

.param-row:last-child td {
  border-bottom: none;
}

.row-alt {
  background: rgba(128, 128, 128, 0.03);
}

.col-param {
  width: 200px;
}

.col-ac, .col-dc {
  width: 70px;
  text-align: center;
}

.col-desc {
  font-size: 11px;
  color: var(--text-tertiary);
}

.param-key {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 700;
  color: var(--lenovo-red);
  font-size: 10px;
}

/* ===== Value Pills ===== */
.value-pill {
  display: inline-block;
  padding: 3px 8px;
  border-radius: 4px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 600;
  font-size: 10px;
  min-width: 40px;
}

.value-pill.filled.ac-pill {
  background: rgba(230, 63, 50, 0.1);
  color: var(--lenovo-red);
  border: 1px solid rgba(230, 63, 50, 0.2);
}

.value-pill.filled.dc-pill {
  background: rgba(245, 158, 11, 0.1);
  color: var(--accent-yellow);
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.value-pill:not(.filled) {
  color: var(--text-tertiary);
}

.th-ac-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--lenovo-red);
  vertical-align: middle;
  margin-right: 4px;
}

.th-dc-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: var(--accent-yellow);
  vertical-align: middle;
  margin-right: 4px;
}

/* ===== Definitions ===== */
.definitions-section {
  margin-bottom: 16px;
  padding: 16px;
  background: rgba(230, 63, 50, 0.04);
  border-radius: 8px;
  border: 1px solid rgba(230, 63, 50, 0.15);
}

.defs-header {
  font-size: 12px;
  font-weight: 700;
  color: var(--lenovo-red);
  margin-bottom: 10px;
  display: flex;
  align-items: center;
}

.defs-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  gap: 12px;
}

.def-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border: 1px solid var(--border-color);
}

.def-key {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-family: 'Consolas', 'Monaco', monospace;
}

.def-value {
  font-size: 18px;
  font-weight: 800;
  color: var(--lenovo-red);
  font-family: 'Consolas', 'Monaco', monospace;
}

.def-desc {
  font-size: 10px;
  color: var(--text-tertiary);
}

/* ===== Scheme Summary ===== */
.scheme-summary {
  padding: 16px;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
}

.summary-header {
  font-size: 12px;
  font-weight: 700;
  color: var(--lenovo-red);
  margin-bottom: 10px;
  display: flex;
  align-items: center;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 10px;
}

.summary-item {
  padding: 12px;
  background: var(--bg-tertiary);
  border-radius: 6px;
  border: 1px solid var(--border-color);
  transition: var(--transition);
}

.summary-item:hover {
  border-color: rgba(230, 63, 50, 0.3);
  transform: translateY(-1px);
}

.summary-name {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 6px;
}

.summary-details {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.summary-details span {
  font-size: 9px;
  color: var(--text-tertiary);
  font-family: 'Consolas', 'Monaco', monospace;
}

/* ===== Driver Paths ===== */
.driver-paths {
  margin-top: 10px;
  padding: 10px 14px;
  background: var(--bg-secondary);
  border-radius: 6px;
  border: 1px solid var(--border-color);
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
}

.path-row {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 11px;
}

.path-label {
  font-weight: 700;
  color: var(--lenovo-red);
  font-size: 10px;
  text-transform: uppercase;
  white-space: nowrap;
  min-width: 70px;
}

.path-value {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 10px;
  padding: 3px 8px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 4px;
  color: var(--text-secondary);
  word-break: break-all;
  flex: 1;
}

.path-value.sys-path {
  color: var(--accent-yellow);
  border-color: rgba(245, 158, 11, 0.2);
}

/* ===== PPM Tuner ===== */
.tuner-section {
  margin-bottom: 16px;
  padding: 20px;
  background: var(--bg-secondary);
  border-radius: 10px;
  border: 1px solid var(--border-color);
}

.tuner-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
}

.tuner-header-left {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tuner-title {
  font-size: 14px;
  font-weight: 700;
  color: var(--text-primary);
}

.tuner-badge {
  font-size: 9px;
  padding: 2px 8px;
  border-radius: 10px;
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red);
  font-weight: 700;
  text-transform: uppercase;
  letter-spacing: 0.5px;
}

.tuner-header-actions {
  display: flex;
  gap: 8px;
}

.btn-tuner-refresh {
  padding: 6px 14px;
  background: transparent;
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-secondary);
  font-size: 11px;
  cursor: pointer;
  transition: var(--transition);
}
.btn-tuner-refresh:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
}
.btn-tuner-refresh:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-tuner-restore {
  padding: 6px 14px;
  background: rgba(245, 158, 11, 0.1);
  border: 1px solid rgba(245, 158, 11, 0.25);
  border-radius: 6px;
  color: var(--accent-yellow);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}
.btn-tuner-restore:hover {
  background: rgba(245, 158, 11, 0.2);
}
.btn-tuner-restore:disabled { opacity: 0.5; cursor: not-allowed; }

.tuner-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 10px;
  background: rgba(230, 63, 50, 0.08);
  border: 1px solid rgba(230, 63, 50, 0.25);
  border-radius: 6px;
  color: #fca5a5;
  font-size: 11px;
}

.tuner-success {
  padding: 8px 12px;
  margin-bottom: 10px;
  background: rgba(74, 222, 128, 0.08);
  border: 1px solid rgba(74, 222, 128, 0.25);
  border-radius: 6px;
  color: #4ade80;
  font-size: 11px;
   animation: fade-in 0.3s ease;
}

@keyframes fade-in {
  from { opacity: 0; transform: translateY(-4px); }
  to { opacity: 1; transform: translateY(0); }
}

.tuner-scheme-info {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  margin-bottom: 14px;
  background: rgba(0, 0, 0, 0.15);
  border-radius: 6px;
  border: 1px solid var(--border-color);
  font-size: 11px;
}

.scheme-info-label {
  color: var(--text-tertiary);
  font-weight: 600;
}

.scheme-info-value {
  color: var(--lenovo-red);
  font-weight: 700;
}

.scheme-info-guid {
  color: var(--text-tertiary);
  font-size: 9px;
}

.tuner-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
  gap: 12px;
}

.tuner-card {
  background: var(--bg-card);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 14px;
  transition: var(--transition);
}

.tuner-card:hover {
  border-color: rgba(230, 63, 50, 0.3);
}

.tuner-card-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
}

.tuner-card-title {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary);
}

.tuner-card-guid {
  font-size: 8px;
  color: var(--text-tertiary);
}

.tuner-card-desc {
  font-size: 10px;
  color: var(--text-tertiary);
  margin-bottom: 10px;
}

.tuner-params {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 12px;
}

.tuner-row {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tuner-label {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary);
  min-width: 120px;
  flex-shrink: 0;
}

.tuner-value-group {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.tuner-current {
  font-size: 10px;
  color: var(--text-tertiary);
  font-family: 'Consolas', 'Monaco', monospace;
  padding: 2px 8px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 4px;
  border: 1px solid var(--border-color);
  min-width: 60px;
  text-align: center;
  white-space: nowrap;
}

.tuner-input {
  flex: 1;
  padding: 6px 10px;
  background: rgba(0, 0, 0, 0.2);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  color: var(--text-primary);
  font-size: 12px;
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 600;
  outline: none;
  transition: border-color 0.2s;
  min-width: 80px;
}

.tuner-input:focus {
  border-color: var(--lenovo-red);
  box-shadow: 0 0 0 2px rgba(230, 63, 50, 0.15);
}

.tuner-input::placeholder {
  color: var(--text-tertiary);
  font-weight: 400;
}

.btn-apply {
  width: 100%;
  padding: 8px 16px;
  background: linear-gradient(135deg, var(--lenovo-red), var(--lenovo-red-dark, #c0281e));
  color: #fff;
  border: none;
  border-radius: 6px;
  font-size: 12px;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s;
  box-shadow: 0 2px 6px rgba(230, 63, 50, 0.25);
}

.btn-apply:hover:not(:disabled) {
  box-shadow: 0 4px 12px rgba(230, 63, 50, 0.4);
  transform: translateY(-1px);
}

.btn-apply:disabled {
  opacity: 0.5;
  cursor: not-allowed;
  transform: none;
}
</style>
