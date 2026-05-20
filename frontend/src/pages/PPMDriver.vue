<template>
  <div class="ppm-page">
    <!-- PPM Parameters Card -->
    <div class="card params-card">
      <!-- Header -->
      <div class="card-header ppm-header">
        <div class="card-title-info">
          <h2>
            <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px;vertical-align:middle;color:#3b82f6;">
              <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
            </svg>
            PPM Parameters
          </h2>
          <p class="card-subtitle">Arrow Lake HX · PPMP-ARL-v1007.20250118 · Family 6 Model 198</p>
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

        <!-- Definitions Section -->
        <div class="definitions-section">
          <div class="defs-header">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;color:#8b5cf6;">
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
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;color:#3b82f6;">
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
      ipfDriver: { name: 'Intel Performance Framework Manager', version: 'v2.2.10204.8', date: '' },
      dttDriver: { name: 'Intel Dynamic Tuning Technology', version: 'v9.0.11905.54373', date: '' },
      ppmProvisioning: { name: 'PPM Provisioning', version: 'v1.0.0.97', date: '', path: '' },
      activeScheme: 'Balanced',
      activeProfile: 'Default',
      
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
  background: var(--bg-card, #1e1e2e);
  border-radius: 12px;
  border: 1px solid var(--border-color, #2e2e42);
  overflow: hidden;
}

.ppm-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 18px 24px;
  background: linear-gradient(135deg, rgba(59, 130, 246, 0.06) 0%, rgba(139, 92, 246, 0.06) 100%);
  border-bottom: 1px solid var(--border-color, #2e2e42);
}

.card-title-info h2 {
  margin: 0;
  font-size: 20px;
  font-weight: 700;
  color: var(--text-primary, #e0e0e8);
  display: flex;
  align-items: center;
}

.card-title-info .card-subtitle {
  margin: 4px 0 0 0;
  font-size: 12px;
  color: var(--text-secondary, #888);
  font-family: 'Consolas', 'Monaco', monospace;
  letter-spacing: 0.02em;
}

.cpu-badge {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: rgba(59, 130, 246, 0.1);
  border: 1px solid rgba(59, 130, 246, 0.2);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: #3b82f6;
  white-space: nowrap;
}

/* ===== Active Indicator Bar ===== */
.active-indicator-bar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 24px;
  background: rgba(139, 92, 246, 0.06);
  border-bottom: 1px solid var(--border-color, #2e2e42);
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
  background: #22c55e;
  box-shadow: 0 0 6px rgba(34, 197, 94, 0.5);
  animation: pulse-dot 2s ease-in-out infinite;
}

@keyframes pulse-dot {
  0%, 100% { box-shadow: 0 0 6px rgba(34, 197, 94, 0.5); }
  50% { box-shadow: 0 0 12px rgba(34, 197, 94, 0.8); }
}

.indicator-label {
  font-size: 12px;
  color: var(--text-secondary, #888);
}

.indicator-scheme {
  font-size: 13px;
  font-weight: 700;
  color: #8b5cf6;
}

.indicator-arrow {
  color: var(--text-tertiary, #666);
  font-weight: 300;
}

.indicator-profile {
  font-size: 13px;
  font-weight: 600;
  color: #3b82f6;
}

.indicator-right {
  display: flex;
  align-items: center;
  gap: 8px;
}

.indicator-hint {
  font-size: 11px;
  color: var(--text-tertiary, #666);
  font-weight: 500;
}

.indicator-sep {
  color: var(--border-color, #2e2e42);
  font-weight: 300;
}

/* ===== Content ===== */
.params-card {
  flex: 1;
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
  border: 1px solid var(--border-color, #2e2e42);
  border-radius: 8px;
  background: var(--bg-secondary, #1a1a2e);
  color: var(--text-secondary, #888);
  font-size: 12px;
  font-weight: 600;
  cursor: pointer;
  transition: all 0.25s ease;
}

.scheme-icon {
  font-size: 14px;
}

.scheme-tab:hover {
  border-color: rgba(139, 92, 246, 0.4);
  color: #8b5cf6;
  background: rgba(139, 92, 246, 0.05);
}

.scheme-tab.active {
  background: linear-gradient(135deg, rgba(139, 92, 246, 0.15), rgba(59, 130, 246, 0.15));
  border-color: rgba(139, 92, 246, 0.5);
  color: #c4b5fd;
  box-shadow: 0 2px 8px rgba(139, 92, 246, 0.15);
}

/* ===== Profile Tabs ===== */
.profile-tabs {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 6px;
  margin-bottom: 16px;
  padding: 8px 12px;
  background: var(--bg-secondary, #1a1a2e);
  border-radius: 8px;
  border: 1px solid var(--border-color, #2e2e42);
}

.profile-tabs-label {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-tertiary, #666);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  margin-right: 4px;
}

.profile-tab {
  padding: 4px 12px;
  border: 1px solid transparent;
  border-radius: 14px;
  background: transparent;
  color: var(--text-tertiary, #666);
  font-size: 10px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.profile-tab:hover {
  color: var(--text-secondary, #888);
  background: var(--bg-tertiary, #252540);
}

.profile-tab.active {
  background: rgba(59, 130, 246, 0.15);
  border-color: rgba(59, 130, 246, 0.4);
  color: #3b82f6;
  font-weight: 700;
}

/* ===== Parameters Table ===== */
.params-table-container {
  overflow-x: auto;
  margin-bottom: 16px;
  border: 1px solid var(--border-color, #2e2e42);
  border-radius: 8px;
  overflow: hidden;
}

.params-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
}

.params-table thead {
  background: linear-gradient(180deg, var(--bg-tertiary, #222240), var(--bg-secondary, #1a1a2e));
}

.params-table th {
  padding: 10px 14px;
  text-align: left;
  font-weight: 700;
  color: var(--text-secondary, #888);
  text-transform: uppercase;
  letter-spacing: 0.06em;
  font-size: 10px;
  border-bottom: 2px solid var(--border-color, #2e2e42);
}

.params-table td {
  padding: 9px 14px;
  border-bottom: 1px solid var(--border-color, #2e2e42);
  vertical-align: middle;
}

.param-row:last-child td {
  border-bottom: none;
}

.row-alt {
  background: rgba(255, 255, 255, 0.015);
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
  color: var(--text-tertiary, #666);
}

.param-key {
  font-family: 'Consolas', 'Monaco', monospace;
  font-weight: 700;
  color: #a78bfa;
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
  background: rgba(59, 130, 246, 0.1);
  color: #60a5fa;
  border: 1px solid rgba(59, 130, 246, 0.2);
}

.value-pill.filled.dc-pill {
  background: rgba(245, 158, 11, 0.1);
  color: #fbbf24;
  border: 1px solid rgba(245, 158, 11, 0.2);
}

.value-pill:not(.filled) {
  color: var(--text-tertiary, #666);
}

.th-ac-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #3b82f6;
  vertical-align: middle;
  margin-right: 4px;
}

.th-dc-dot {
  display: inline-block;
  width: 6px;
  height: 6px;
  border-radius: 50%;
  background: #f59e0b;
  vertical-align: middle;
  margin-right: 4px;
}

/* ===== Definitions ===== */
.definitions-section {
  margin-bottom: 16px;
  padding: 16px;
  background: rgba(139, 92, 246, 0.04);
  border-radius: 8px;
  border: 1px solid rgba(139, 92, 246, 0.15);
}

.defs-header {
  font-size: 12px;
  font-weight: 700;
  color: #a78bfa;
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
  background: var(--bg-secondary, #1a1a2e);
  border-radius: 6px;
  border: 1px solid var(--border-color, #2e2e42);
}

.def-key {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-tertiary, #666);
  text-transform: uppercase;
  letter-spacing: 0.04em;
  font-family: 'Consolas', 'Monaco', monospace;
}

.def-value {
  font-size: 18px;
  font-weight: 800;
  background: linear-gradient(135deg, #8b5cf6, #6366f1);
  -webkit-background-clip: text;
  -webkit-text-fill-color: transparent;
  background-clip: text;
  font-family: 'Consolas', 'Monaco', monospace;
}

.def-desc {
  font-size: 10px;
  color: var(--text-tertiary, #666);
}

/* ===== Scheme Summary ===== */
.scheme-summary {
  padding: 16px;
  background: var(--bg-secondary, #1a1a2e);
  border-radius: 8px;
  border: 1px solid var(--border-color, #2e2e42);
}

.summary-header {
  font-size: 12px;
  font-weight: 700;
  color: #60a5fa;
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
  background: var(--bg-tertiary, #252540);
  border-radius: 6px;
  border: 1px solid var(--border-color, #2e2e42);
  transition: all 0.2s ease;
}

.summary-item:hover {
  border-color: rgba(59, 130, 246, 0.3);
  transform: translateY(-1px);
}

.summary-name {
  font-size: 12px;
  font-weight: 700;
  color: var(--text-primary, #e0e0e8);
  margin-bottom: 6px;
}

.summary-details {
  display: flex;
  flex-direction: column;
  gap: 3px;
}

.summary-details span {
  font-size: 9px;
  color: var(--text-tertiary, #666);
  font-family: 'Consolas', 'Monaco', monospace;
}
</style>
