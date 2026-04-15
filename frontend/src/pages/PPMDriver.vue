<template>
  <div class="ppm-page">
    <!-- PPM Parameters Card - Comprehensive Analysis -->
    <div class="card params-card">
      <div class="card-header">
        <div class="card-title-info">
          <h2>PPM Parameters (Arrow Lake HX)</h2>
          <p class="card-subtitle">PPM-ARL-v1007.20250118 | Family 6 Model 198</p>
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
            {{ scheme.name }}
          </button>
        </div>

        <!-- Profile Tabs -->
        <div class="profile-tabs">
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
                <th class="col-param">Parameter</th>
                <th class="col-ac">AC</th>
                <th class="col-dc">DC</th>
                <th class="col-desc">Description</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="param in getCurrentParams" :key="param.key">
                <td class="col-param">
                  <span class="param-key">{{ param.key }}</span>
                </td>
                <td class="col-ac">
                  <span :class="['value', { highlight: param.ac && param.ac !== '-' }]">{{ formatValue(param.ac) }}</span>
                </td>
                <td class="col-dc">
                  <span :class="['value', { highlight: param.dc && param.dc !== '-' }]">{{ formatValue(param.dc) }}</span>
                </td>
                <td class="col-desc">{{ param.desc }}</td>
              </tr>
            </tbody>
          </table>
        </div>

        <!-- Definitions Section -->
        <div class="definitions-section">
          <div class="defs-header">Definitions (Hetero Thresholds)</div>
          <div class="defs-grid">
            <div class="def-item">
              <span class="def-key">HeteroDecreaseThreshold</span>
              <span class="def-value">1347440720</span>
              <span class="def-desc">P-Core→E-Core downgrade threshold</span>
            </div>
            <div class="def-item">
              <span class="def-key">HeteroIncreaseThreshold</span>
              <span class="def-value">2139259522</span>
              <span class="def-desc">E-Core→P-Core upgrade threshold</span>
            </div>
          </div>
        </div>

        <!-- Scheme Summary -->
        <div class="scheme-summary">
          <div class="summary-header">Scheme Summary</div>
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
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
}

.card-title-info .card-subtitle {
  margin: 3px 0 0 0;
  font-size: 13px;
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

.cpu-info-row {
  display: flex;
  align-items: baseline;
  gap: 16px;
  margin-bottom: 12px;
  padding-bottom: 12px;
  border-bottom: 1px solid var(--border-color);
}

.cpu-name {
  font-size: 16px;
  font-weight: 700;
  color: var(--text-primary);
}

.cpu-cores {
  font-size: 13px;
  color: var(--text-secondary);
  font-weight: 500;
}

.drivers-table {
  width: 100%;
  border-collapse: collapse;
  table-layout: fixed;
}

.drivers-table th {
  padding: 10px 16px;
  text-align: center;
  font-size: 11px;
  font-weight: 700;
  color: var(--text-secondary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border: 1px solid var(--border-color);
  background: var(--bg-secondary);
}

.drivers-table td {
  padding: 10px 16px;
  text-align: center;
  font-size: 14px;
  font-weight: 700;
  color: var(--accent-green);
  font-family: 'Consolas', monospace;
  border: 1px solid var(--border-color);
}

.params-card {
  flex: 1;
  min-height: 400px;
}

.params-content {
  padding: 12px 16px;
}

.scheme-tabs {
  display: flex;
  gap: 4px;
  margin-bottom: 8px;
  border-bottom: 1px solid var(--border-color);
  padding-bottom: 8px;
}

.scheme-tab {
  padding: 8px 16px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: var(--bg-secondary);
  color: var(--text-secondary);
  font-size: 11px;
  font-weight: 600;
  cursor: pointer;
  transition: var(--transition);
}

.scheme-tab:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
}

.scheme-tab.active {
  background: var(--lenovo-red);
  border-color: var(--lenovo-red);
  color: white;
}

.profile-tabs {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-bottom: 12px;
}

.profile-tab {
  padding: 5px 10px;
  border: 1px solid var(--border-color);
  border-radius: var(--radius-sm);
  background: transparent;
  color: var(--text-tertiary);
  font-size: 10px;
  font-weight: 500;
  cursor: pointer;
  transition: var(--transition);
}

.profile-tab:hover {
  border-color: var(--text-secondary);
  color: var(--text-secondary);
}

.profile-tab.active {
  background: var(--bg-tertiary);
  border-color: var(--accent-green);
  color: var(--accent-green);
}

.params-table-container {
  overflow-x: auto;
  margin-bottom: 16px;
}

.params-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 11px;
}

.params-table th {
  padding: 8px 12px;
  text-align: left;
  font-weight: 700;
  color: var(--text-tertiary);
  background: var(--bg-tertiary);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  border-bottom: 1px solid var(--border-color);
}

.params-table td {
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color);
  vertical-align: middle;
}

.params-table tr:last-child td {
  border-bottom: none;
}

.col-param {
  width: 180px;
}

.col-ac, .col-dc {
  width: 80px;
  text-align: center;
}

.col-desc {
  width: auto;
}

.param-key {
  font-family: 'Consolas', monospace;
  font-weight: 700;
  color: var(--lenovo-red);
  font-size: 10px;
}

.value {
  display: inline-block;
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', monospace;
  font-weight: 600;
  font-size: 10px;
}

.value.highlight {
  background: rgba(230, 63, 50, 0.1);
  color: var(--lenovo-red);
}

.definitions-section {
  margin-bottom: 16px;
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.defs-header {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.defs-grid {
  display: flex;
  gap: 24px;
}

.def-item {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.def-key {
  font-size: 9px;
  font-weight: 600;
  color: var(--text-tertiary);
  text-transform: uppercase;
}

.def-value {
  font-size: 14px;
  font-weight: 700;
  color: var(--lenovo-red);
  font-family: 'Consolas', monospace;
}

.def-desc {
  font-size: 9px;
  color: var(--text-tertiary);
}

.scheme-summary {
  padding: 12px;
  background: var(--bg-secondary);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-color);
}

.summary-header {
  font-size: 11px;
  font-weight: 700;
  color: var(--text-secondary);
  margin-bottom: 8px;
}

.summary-grid {
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
}

.summary-item {
  padding: 8px;
  background: var(--bg-tertiary);
  border-radius: var(--radius-sm);
}

.summary-name {
  font-size: 10px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 4px;
}

.summary-details {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.summary-details span {
  font-size: 9px;
  color: var(--text-tertiary);
}
</style>
