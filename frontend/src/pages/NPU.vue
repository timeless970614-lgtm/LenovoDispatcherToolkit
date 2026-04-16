<template>
  <div class="npu-page">
    <!-- NPU Header Card -->
    <div class="card">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/>
            <line x1="9" y1="2" x2="9" y2="4"/><line x1="15" y1="2" x2="15" y2="4"/>
            <line x1="9" y1="20" x2="9" y2="22"/><line x1="15" y1="20" x2="15" y2="22"/>
            <line x1="20" y1="9" x2="22" y2="9"/><line x1="20" y1="14" x2="22" y2="14"/>
            <line x1="2" y1="9" x2="4" y2="9"/><line x1="2" y1="14" x2="4" y2="14"/>
          </svg>
          NPU Smart Scheduler
        </h3>
        <div class="live-indicator" v-if="!npuLoading">
          <span class="live-dot"></span> {{ npuDeviceCount }} Device(s)
        </div>
        <button class="btn btn-secondary btn-sm" @click="refreshNPU" :disabled="npuLoading">
          {{ npuLoading ? 'Refreshing...' : 'Refresh' }}
        </button>
      </div>

      <!-- Driver Error / No Device Warning -->
      <div v-if="npuDriverError" style="margin:16px; padding:20px; background:#7f1d1d; border:2px solid #ef4444; border-radius:8px; text-align:center">
        <div style="font-size:18px; font-weight:bold; color:#fca5a5; margin-bottom:8px">⚠ NPU Device Not Detected</div>
        <div style="font-size:13px; color:#fecaca; line-height:1.6">{{ npuDriverError }}</div>
        <div style="margin-top:12px; font-size:12px; color:#f87171">Please verify: ① Driver installed ② NPU enabled in BIOS ③ No other program using NPU</div>
        <button class="btn btn-secondary btn-sm" style="margin-top:12px; background:#991b1b; color:#fef2f2; border-color:#b91c1c" @click="showNPUREport">View Diagnostic Report</button>
      </div>

      <!-- Loading -->
      <div v-if="npuLoading" style="margin:16px; padding:40px; text-align:center; color:var(--text-secondary)">
        <div style="font-size:16px">Connecting to NPU driver...</div>
        <div style="margin-top:12px">
          <svg width="32" height="32" viewBox="0 0 24 24" style="animation:spin 1s linear infinite; fill:none; stroke:var(--text-secondary); stroke-width:2">
            <circle cx="12" cy="12" r="10" stroke-opacity="0.3"/>
            <path d="M12 2a10 10 0 0 1 10 10"/>
          </svg>
        </div>
      </div>

      <!-- SDK Info Banner -->
      <div v-if="!npuLoading && npuSDKInfo.sdkVersion" class="sdk-banner">
        <div class="sdk-icon">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/></svg>
        </div>
        <div class="sdk-row">
          <span class="sdk-field">SDK</span>
          <span class="sdk-val">{{ npuSDKInfo.sdkVersion }}</span>
        </div>
        <div class="sdk-divider"></div>
        <div class="sdk-row">
          <span class="sdk-field">Driver</span>
          <span class="sdk-val">{{ npuSDKInfo.driverVersion }}</span>
        </div>
        <div class="sdk-divider"></div>
        <div class="sdk-row">
          <span class="sdk-field">Built</span>
          <span class="sdk-val">{{ npuSDKInfo.buildtime }}</span>
        </div>
      </div>

      <!-- No device -->
      <div v-if="!npuLoading && npuDeviceCount === 0 && !npuDriverError" style="margin:16px; padding:30px; text-align:center; color:var(--text-secondary); background:var(--bg-secondary); border:2px dashed var(--border-color); border-radius:8px">
        No NPU device detected. Please verify the driver is properly installed.
      </div>

      <!-- Device List -->
      <div v-if="npuDeviceList.length > 0" style="margin-top:16px">
        <div v-for="dev in npuDeviceList" :key="dev.index" class="npu-device-card">
          <!-- Device Header -->
          <div class="npu-dev-header">
            <div class="npu-dev-title">
              <strong>Device {{ dev.index }}</strong>
              <span class="npu-dev-name">{{ dev.modelName || 'NPU XH2A' }}</span>
              <span class="npu-dvfs-badge" :class="'npu-dvfs-' + (dev.dvfsMode || 'unknown').toLowerCase()">{{ dev.dvfsMode || 'N/A' }}</span>
            </div>
            <div class="npu-dev-actions">
              <button class="btn btn-primary btn-sm" @click="setNPUDVFS(dev.index, 'PERFORMANCE')" :disabled="npuSettingDVFS === dev.index">⚡ PERFORMANCE</button>
              <button class="btn btn-secondary btn-sm" @click="setNPUDVFS(dev.index, 'ONDEMAND')" :disabled="npuSettingDVFS === dev.index">🔄 ONDEMAND</button>
              <button class="btn btn-warning btn-sm" @click="setNPUDVFS(dev.index, 'POWERLIMIT')" :disabled="npuSettingDVFS === dev.index">⚡ POWERLIMIT</button>
              <select v-model="npuSelectedDev" class="form-select form-select-sm" style="width:auto;margin-left:8px">
                <option v-for="d in npuDeviceList" :key="d.index" :value="d.index">Device {{ d.index }}</option>
              </select>
            </div>
          </div>

          <!-- Metrics Grid -->
          <div class="npu-metrics-grid">
            <div class="npu-metric">
              <span class="npu-metric-label">IPU Utilization</span>
              <div class="npu-metric-bar-wrap">
                <div class="npu-metric-bar" :style="{width: Math.round((dev.ipuUtiliRate||0)*100) + '%', background: utilColor(dev.ipuUtiliRate)}"></div>
              </div>
              <span class="npu-metric-val">{{ Math.round((dev.ipuUtiliRate||0)*100) }}%</span>
            </div>
            <div class="npu-metric">
              <span class="npu-metric-label">Chip Temperature</span>
              <div class="npu-metric-bar-wrap">
                <div class="npu-metric-bar" :style="{width: Math.min(100, (dev.temperatureC||0)/1.0) + '%', background: tempColor(dev.temperatureC)}"></div>
              </div>
              <span class="npu-metric-val">{{ dev.temperatureC > 0 ? dev.temperatureC.toFixed(1) + '°C' : 'N/A' }}</span>
            </div>
            <div class="npu-metric">
              <span class="npu-metric-label">IPU Frequency</span>
              <div class="npu-metric-bar-wrap">
                <div class="npu-metric-bar" :style="{width: Math.min(100, (dev.ipuFrequencyHz||0)/14e8*100) + '%', background: '#3b82f6'}"></div>
              </div>
              <span class="npu-metric-val">{{ dev.ipuFrequencyHz > 0 ? (dev.ipuFrequencyHz/1e9).toFixed(3) + ' GHz' : 'N/A' }}</span>
            </div>
            <div class="npu-metric">
              <span class="npu-metric-label">Board Power</span>
              <div class="npu-metric-bar-wrap">
                <div class="npu-metric-bar" :style="{width: Math.min(100, (dev.boardPowerW||0)/50*100) + '%', background: '#a855f7'}"></div>
              </div>
              <span class="npu-metric-val">{{ dev.boardPowerW > 0 ? dev.boardPowerW.toFixed(2) + ' W' : 'N/A' }}</span>
            </div>
            <div class="npu-metric">
              <span class="npu-metric-label">IPU Voltage</span>
              <div class="npu-metric-bar-wrap">
                <div class="npu-metric-bar" :style="{width: Math.min(100, (dev.ipuVoltageMV||0)/1200*100) + '%', background: '#f59e0b'}"></div>
              </div>
              <span class="npu-metric-val">{{ dev.ipuVoltageMV > 0 ? dev.ipuVoltageMV.toFixed(2) + ' mV' : 'N/A' }}</span>
            </div>
            <div class="npu-metric">
              <span class="npu-metric-label">DDR Memory</span>
              <div class="npu-metric-bar-wrap">
                <div class="npu-metric-bar" :style="{width: dev.memTotalMB > 0 ? Math.round((dev.memUsedMB||0)/(dev.memTotalMB)*100) + '%' : '0%', background: '#64748b'}"></div>
              </div>
              <span class="npu-metric-val">{{ dev.memTotalMB > 0 ? (dev.memUsedMB||0) + ' / ' + dev.memTotalMB + ' MB' : 'N/A' }}</span>
            </div>
          </div>

          <!-- Basic Info Row -->
          <div class="npu-info-row">
            <div class="npu-info-item"><span class="npu-info-label">Vendor ID</span><span class="npu-info-val">{{ dev.vendorId > 0 ? '0x' + dev.vendorId.toString(16).toUpperCase() : 'N/A' }}</span></div>
            <div class="npu-info-item"><span class="npu-info-label">Computing Power</span><span class="npu-info-val">{{ dev.computingPower > 0 ? dev.computingPower + ' TOPS' : 'N/A' }}</span></div>
            <div class="npu-info-item"><span class="npu-info-label">Core Count</span><span class="npu-info-val">{{ dev.coreCount > 0 ? dev.coreCount : 'N/A' }}</span></div>
            <div class="npu-info-item"><span class="npu-info-label">DDR Capacity</span><span class="npu-info-val">{{ dev.ddrSizeMB > 0 ? (dev.ddrSizeMB/1024).toFixed(1) + ' GB' : 'N/A' }}</span></div>
            <div class="npu-info-item"><span class="npu-info-label">Firmware Ver</span><span class="npu-info-val">{{ dev.firmwareVer || 'N/A' }}</span></div>
          </div>

          <!-- Per-Core Utilization -->
          <div v-if="dev.coreCount > 0" class="npu-cores-section">
            <div class="npu-cores-title">Per-Core Utilization</div>
            <div class="npu-cores-grid">
              <div v-for="(pct, idx) in (npuPowerStatus[dev.index] ? getCoreUtilList(dev.index, dev.coreCount) : [])" :key="idx" class="npu-core-row">
                <span class="npu-core-lbl">Core {{ idx }}</span>
                <div class="npu-core-bwrap">
                  <div class="npu-core-bfill" :style="{width: pct + '%', background: utilColorColor(pct/100)}"></div>
                </div>
                <span class="npu-core-pct">{{ pct.toFixed(1) }}%</span>
              </div>
            </div>
          </div>

          <!-- Power Limit Panel -->
          <div class="npu-panel">
            <div class="npu-panel-title">⚡ Power Limit (POWERLIMIT Mode)</div>
            <div class="npu-power-limit-row">
              <div class="npu-power-limit-field">
                <label>Max Power (W)</label>
                <input type="number" v-model.number="npuPowerLimit[dev.index].maxW" min="5" max="50" class="form-input form-input-sm" style="width:80px">
              </div>
              <div class="npu-power-limit-field">
                <label>Min Power (W)</label>
                <input type="number" v-model.number="npuPowerLimit[dev.index].minW" min="3" max="50" class="form-input form-input-sm" style="width:80px">
              </div>
              <button class="btn btn-warning btn-sm" @click="setNPUPowerLimit(dev.index)" :disabled="npuPowerLimit[dev.index]?.setting">Apply Power Limit</button>
            </div>
            <div class="npu-power-limit-tip">After setting, DVFS will switch to POWERLIMIT mode. Device will dynamically adjust frequency within the power range.</div>
            <div v-if="npuPowerResult[dev.index]" :class="['result-message', npuPowerResult[dev.index].Success ? 'success' : 'error']" style="margin-top:8px">
              {{ npuPowerResult[dev.index].Message }}
            </div>
          </div>

          <!-- Power Status from hm_smi -->
          <div v-if="npuPowerStatus[dev.index]" class="npu-power-status">
            <div class="npu-power-title">hm_smi Power Status</div>
            <div class="npu-power-grid">
              <div class="npu-power-item"><span>DVFS Mode</span><span>{{ npuPowerStatus[dev.index].dvfsMode || 'N/A' }}</span></div>
              <div class="npu-power-item"><span>IPU Frequency</span><span>{{ npuPowerStatus[dev.index].curIpuFreqMHz > 0 ? npuPowerStatus[dev.index].curIpuFreqMHz + ' MHz' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>Freq Range</span><span>{{ npuPowerStatus[dev.index].lockIpuMinMHz || '—' }} ~ {{ npuPowerStatus[dev.index].lockIpuMaxMHz || '—' }} MHz</span></div>
              <div class="npu-power-item"><span>Board Power</span><span>{{ npuPowerStatus[dev.index].boardPowerW > 0 ? npuPowerStatus[dev.index].boardPowerW + ' W' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>IPU Load</span><span>{{ npuPowerStatus[dev.index].ipuLoadPct > 0 ? npuPowerStatus[dev.index].ipuLoadPct + '%' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>DDR Available</span><span>{{ npuPowerStatus[dev.index].ddrFreeMB > 0 ? npuPowerStatus[dev.index].ddrFreeMB + ' MB' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>Core0 Temp</span><span>{{ npuPowerStatus[dev.index].core0TempC > 0 ? npuPowerStatus[dev.index].core0TempC + '°C' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>Core1 Temp</span><span>{{ npuPowerStatus[dev.index].core1TempC > 0 ? npuPowerStatus[dev.index].core1TempC + '°C' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>DDR0 Temp</span><span>{{ npuPowerStatus[dev.index].ddr0TempC > 0 ? npuPowerStatus[dev.index].ddr0TempC + '°C' : 'N/A' }}</span></div>
              <div class="npu-power-item"><span>DDR2 Temp</span><span>{{ npuPowerStatus[dev.index].ddr2TempC > 0 ? npuPowerStatus[dev.index].ddr2TempC + '°C' : 'N/A' }}</span></div>
            </div>
          </div>

          <!-- Clock Lock Panel -->
          <div class="npu-panel">
            <div class="npu-panel-title">🔒 Clock Frequency Lock</div>
            <div class="npu-clock-row">
              <div class="npu-clock-field">
                <label>Max Frequency (MHz)</label>
                <input type="number" v-model.number="npuClockLocking[dev.index].maxMhz" min="700" max="1400" class="form-input form-input-sm" placeholder="1400" style="width:100px">
              </div>
              <div class="npu-clock-field">
                <label>Min Frequency (MHz)</label>
                <input type="number" v-model.number="npuClockLocking[dev.index].minMhz" min="700" max="1400" class="form-input form-input-sm" placeholder="700" style="width:100px">
              </div>
              <button class="btn btn-primary btn-sm" @click="setNPUClockLock(dev.index)" :disabled="npuClockLocking[dev.index]?.setting">Apply</button>
              <button class="btn btn-secondary btn-sm" @click="resetNPUDefaults(dev.index)" :disabled="npuClockLocking[dev.index]?.setting">Reset</button>
            </div>
            <div v-if="npuPowerResult[dev.index]" :class="['result-message', npuPowerResult[dev.index].Success ? 'success' : 'error']" style="margin-top:8px">
              {{ npuPowerResult[dev.index].Message }}
            </div>
          </div>

          <!-- DVFS Result -->
          <div v-if="dev.dvfsResult" :class="['result-message', dev.dvfsResult.startsWith('Error') ? 'error' : 'success']" style="margin-top:8px">
            {{ dev.dvfsResult }}
          </div>
        </div>
      </div>

      <!-- Debug Panel -->
      <div style="margin-top:16px">
        <button class="btn btn-secondary btn-sm" @click="npuShowDebug = !npuShowDebug">
          {{ npuShowDebug ? 'Hide' : 'Show' }} Diagnostic Report
        </button>
        <div v-if="npuShowDebug" class="result-message" style="margin-top:8px; background:#1a1a2e; color:#0f0; font-family:monospace; font-size:11px; max-height:300px; overflow-y:auto; white-space:pre-wrap; text-align:left">
          {{ npuRawProbe || 'Loading...' }}
        </div>
      </div>
    </div>

    <!-- Smart Scheduler -->
    <div class="card" style="margin-top:16px">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"/>
          </svg>
          Smart Auto Scheduler
        </h3>
        <div class="live-indicator" v-if="npuSchedRunning">
          <span class="live-dot"></span> Running
        </div>
      </div>

      <div class="npu-sched-grid">
        <div class="npu-sched-field">
          <label>Device</label>
          <select v-model="npuSchedDev" class="form-select form-select-sm">
            <option v-for="d in npuDeviceList" :key="d.index" :value="d.index">Device {{ d.index }}</option>
          </select>
        </div>
        <div class="npu-sched-field">
          <label>High Load Threshold %</label>
          <input type="number" v-model.number="npuSchedSettings.utilHighPct" min="10" max="100" class="form-input form-input-sm">
        </div>
        <div class="npu-sched-field">
          <label>Low Load Threshold %</label>
          <input type="number" v-model.number="npuSchedSettings.utilLowPct" min="1" max="99" class="form-input form-input-sm">
        </div>
        <div class="npu-sched-field">
          <label>Temp Warning °C</label>
          <input type="number" v-model.number="npuSchedSettings.tempWarnC" min="40" max="100" class="form-input form-input-sm">
        </div>
        <div class="npu-sched-field">
          <label>Temp Critical °C</label>
          <input type="number" v-model.number="npuSchedSettings.tempCritC" min="50" max="110" class="form-input form-input-sm">
        </div>
        <div class="npu-sched-field">
          <label>Check Interval (sec)</label>
          <input type="number" v-model.number="npuSchedSettings.checkSec" min="1" max="60" class="form-input form-input-sm">
        </div>
      </div>

      <div class="npu-sched-actions">
        <button v-if="!npuSchedRunning" class="btn btn-primary" @click="startNpuScheduler" :disabled="npuSchedStarting || npuDeviceList.length === 0">
          ▶ {{ npuSchedStarting ? 'Starting...' : 'Start Scheduler' }}
        </button>
        <button v-else class="btn btn-warning" @click="stopNpuScheduler">■ Stop Scheduler</button>
      </div>

      <!-- Scheduler State -->
      <div v-if="npuSchedRunning" class="npu-sched-state">
        <div class="npu-sched-state-title">Scheduler Running — {{ npuSchedState.curMode || '—' }}</div>
        <div class="npu-sched-state-grid">
          <div class="npu-sched-state-item"><span>IPU Utilization</span><span>{{ npuSchedState.curUtilPct >= 0 ? npuSchedState.curUtilPct.toFixed(1) + '%' : '—' }}</span></div>
          <div class="npu-sched-state-item"><span>Chip Temp</span><span>{{ npuSchedState.curTempC > 0 ? npuSchedState.curTempC.toFixed(1) + '°C' : '—' }}</span></div>
          <div class="npu-sched-state-item"><span>Board Power</span><span>{{ npuSchedState.curPowerW > 0 ? npuSchedState.curPowerW.toFixed(2) + ' W' : '—' }}</span></div>
          <div class="npu-sched-state-item"><span>IPU Frequency</span><span>{{ npuSchedState.curFreqMHz > 0 ? npuSchedState.curFreqMHz.toFixed(0) + ' MHz' : '—' }}</span></div>
          <div class="npu-sched-state-item"><span>Freq Range</span><span>{{ npuSchedState.curLockMinMHz || '—' }} ~ {{ npuSchedState.curLockMaxMHz || '—' }} MHz</span></div>
          <div class="npu-sched-state-item"><span>Decision</span><span>{{ npuSchedState.decision || '—' }}</span></div>
          <div class="npu-sched-state-item"><span>Last Switch</span><span>{{ npuSchedState.lastSwitch || '—' }}</span></div>
        </div>
      </div>
    </div>

    <!-- Parameter Reference -->
    <div class="card" style="margin-top:16px">
      <div class="card-header">
        <h3 class="card-title">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/>
            <line x1="8" y1="7" x2="16" y2="7"/><line x1="8" y1="11" x2="16" y2="11"/><line x1="8" y1="15" x2="12" y2="15"/>
          </svg>
          Parameter Reference
        </h3>
        <button class="btn btn-secondary btn-sm" @click="npuShowRef = !npuShowRef">
          {{ npuShowRef ? 'Collapse' : 'Expand' }}
        </button>
      </div>
      <div v-if="npuShowRef" class="npu-ref-content">
        <div class="npu-ref-section">
          <div class="npu-ref-title">DVFS Mode Description</div>
          <table class="npu-ref-table">
            <thead><tr><th>Mode</th><th>Frequency Range</th><th>Use Case</th></tr></thead>
            <tbody>
              <tr><td><strong>PERFORMANCE</strong></td><td>Fixed 1400 MHz</td><td>Batch inference, compute-intensive tasks</td></tr>
              <tr><td><strong>ONDEMAND</strong></td><td>Dynamic 200-1400 MHz</td><td>Idle, low-load scenarios</td></tr>
            </tbody>
          </table>
        </div>
        <div class="npu-ref-section">
          <div class="npu-ref-title">Smart Scheduler Policy</div>
          <table class="npu-ref-table">
            <thead><tr><th>Condition</th><th>Action</th></tr></thead>
            <tbody>
              <tr><td>Utilization &gt; {{ npuSchedSettings.utilHighPct }}%</td><td>Switch to PERFORMANCE</td></tr>
              <tr><td>Utilization &lt; {{ npuSchedSettings.utilLowPct }}%</td><td>Switch to ONDEMAND</td></tr>
              <tr><td>Temp &gt; {{ npuSchedSettings.tempWarnC }}°C</td><td>Force ONDEMAND</td></tr>
              <tr><td>Temp &gt; {{ npuSchedSettings.tempCritC }}°C</td><td>Force lowest power mode</td></tr>
            </tbody>
          </table>
        </div>
        <div class="npu-ref-section">
          <div class="npu-ref-title">API Quick Reference</div>
          <table class="npu-ref-table">
            <thead><tr><th>API</th><th>Description</th></tr></thead>
            <tbody>
              <tr><td>hm_sys_get_ipu_utili_rate</td><td>Get IPU overall utilization (0.0~1.0)</td></tr>
              <tr><td>hm_sys_get_ipu_core_utili_rate</td><td>Get per-core utilization</td></tr>
              <tr><td>hm_sys_get_temperature</td><td>Get chip temperature (°C)</td></tr>
              <tr><td>hm_sys_get_board_power</td><td>Get board power (W)</td></tr>
              <tr><td>hm_sys_get_ipu_frequency</td><td>Get IPU frequency (Hz)</td></tr>
              <tr><td>hm_sys_set_dvfs_mode</td><td>Set DVFS mode</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'NPU',
  data() {
    return {
      npuLoading: false,
      npuDriverError: '',
      npuRawProbe: '',
      npuShowDebug: false,
      npuDeviceList: [],
      npuDeviceCount: 0,
      npuSDKInfo: {},
      npuSelectedDev: 0,
      npuClockLocking: {},
      npuPowerLimit: {},
      npuPowerResult: {},
      npuSettingDVFS: null,
      npuShowRef: true,
      npuPowerStatus: {},
      npuPowerTimer: null,
      npuPowerRefreshDev: 0,
      npuSchedRunning: false,
      npuSchedStarting: false,
      npuSchedDev: 0,
      npuSchedState: { curMode: '', decision: '', curUtilPct: -1, curTempC: 0, curPowerW: 0, curFreqMHz: 0, curLockMaxMHz: 0, curLockMinMHz: 0, lastSwitch: '' },
      npuSchedSettings: { utilHighPct: 70, utilLowPct: 20, tempWarnC: 80, tempCritC: 90, checkSec: 5 },
    }
  },
  mounted() {
    // Give Wails time to initialize the Go bindings
    this.npuLoading = false // show loading immediately
    this.npuDriverError = '正在连接 NPU 驱动...'
    setTimeout(() => { this.refreshNPU() }, 300)
  },
  beforeUnmount() {
    this.stopNPUPowerPolling()
  },
  methods: {
    utilColor(v) {
      if (v === undefined || v < 0) return 'var(--text-tertiary)'
      if (v >= 0.8) return '#ef4444'
      if (v >= 0.5) return '#f59e0b'
      return '#22c55e'
    },
    tempColor(c) {
      if (!c || c <= 0) return 'var(--text-tertiary)'
      if (c >= 80) return '#ef4444'
      if (c >= 60) return '#f59e0b'
      return '#22c55e'
    },
    utilColorColor(v) {
      // same as utilColor but accepts 0-100 percentage instead of 0.0-1.0
      if (v === undefined || v < 0) return 'var(--text-tertiary)'
      if (v >= 80) return '#ef4444'
      if (v >= 50) return '#f59e0b'
      return '#22c55e'
    },
    getCoreUtilList(devIndex, coreCount) {
      // Extract core utilization from npuPowerStatus (core0UtilPct / core1UtilPct)
      const status = this.npuPowerStatus[devIndex]
      if (!status) return []
      const list = []
      if (status.core0UtilPct > 0) list.push(status.core0UtilPct)
      if (status.core1UtilPct > 0) list.push(status.core1UtilPct)
      // Pad with zeros for missing cores
      while (list.length < coreCount) list.push(0)
      return list
    },
    async refreshNPU() {
      this.npuLoading = true
      this.npuDriverError = ''
      try {
        // Wait for Wails to be ready
        const waitForWails = () => new Promise(resolve => {
          if (window.go && window.go.main && window.go.main.App) { resolve(); return }
          let waited = 0
          const check = setInterval(() => {
            waited += 100
            if (window.go && window.go.main && window.go.main.App) { clearInterval(check); resolve(); return }
            if (waited > 10000) { clearInterval(check); resolve('timeout'); return }
          }, 100)
        })

        const wailsReady = await waitForWails()
        if (wailsReady === 'timeout' || !window.go.main.App.GetNPUFullReport) {
          this.npuDriverError = 'Wails 后端未就绪，请重启应用'
          this.npuDeviceCount = 0
          this.npuDeviceList = []
          this.npuSDKInfo = {}
        } else {
          this.npuSDKInfo = report.sdkInfo || {}
          this.npuDeviceCount = report.deviceCount || 0
          this.npuDeviceList = (report.devices || []).map(d => ({
            index: d.index,
            vendorId: d.properties ? d.properties.vendorId || 0 : 0,
            serialNumber: d.properties ? d.properties.serialNumber || '' : '',
            modelName: d.properties ? d.properties.modelName || '' : '',
            computingPower: d.properties ? d.properties.computingPowerTOPS || 0 : 0,
            coreCount: d.properties ? d.properties.coreCount || 0 : 0,
            ddrSizeMB: d.properties ? d.properties.ddrSizeMB || 0 : 0,
            firmwareVer: d.properties ? d.properties.firmwareVer || '' : '',
            ipuUtiliRate: d.metrics ? d.metrics.ipuUtiliRate || 0 : 0,
            ipuVoltageMV: d.metrics ? d.metrics.ipuVoltageMV || 0 : 0,
            ipuFrequencyHz: d.metrics ? d.metrics.ipuFrequencyHz || 0 : 0,
            boardPowerW: d.metrics ? d.metrics.boardPowerW || 0 : 0,
            temperatureC: d.metrics ? d.metrics.temperatureC || 0 : 0,
            memTotalMB: d.metrics ? d.metrics.memTotalMB || 0 : 0,
            memUsedMB: d.metrics ? d.metrics.memUsedMB || 0 : 0,
            dvfsMode: d.dvfsMode || '',
            dvfsResult: '',
          }))

          this.npuDeviceList.forEach(dev => {
            if (!this.npuClockLocking[dev.index]) {
              this.npuClockLocking[dev.index] = { maxMhz: 1400, minMhz: 700, setting: false, result: null }
            }
            if (!this.npuPowerLimit[dev.index]) {
              this.npuPowerLimit[dev.index] = { maxW: 25, minW: 5 }
            }
          })

          if (this.npuDeviceCount === 0) {
            this.npuDriverError = '未检测到 NPU 设备。请检查驱动是否安装、BIOS 中 NPU 是否启用，或是否有其他程序占用'
          } else if (this.npuDeviceCount > 0 && !this.npuPowerTimer) {
            this.startNPUPowerPolling(this.npuDeviceList[0].index)
          }
        }
      } catch(e) {
        console.error('[NPU] refreshNPU error:', e)
        this.npuDriverError = '调用 NPU API 出错: ' + String(e)
        this.npuDeviceCount = 0
        this.npuDeviceList = []
      } finally {
        this.npuLoading = false
      }
    },
    async setNPUDVFS(devIndex, mode) {
      this.npuSettingDVFS = devIndex
      const dev = this.npuDeviceList.find(d => d.index === devIndex)
      if (dev) dev.dvfsResult = ''
      try {
        const msg = await window.go.main.App.NPUSetDVFSMode(devIndex, mode)
        if (dev) dev.dvfsResult = msg
        if (!msg.startsWith('Error') && dev) dev.dvfsMode = mode
        await this.refreshNPU()
      } catch(e) {
        if (dev) dev.dvfsResult = 'Error: ' + String(e)
      } finally {
        this.npuSettingDVFS = null
      }
    },
    async setNPUPowerLimit(devIndex) {
      const pl = this.npuPowerLimit[devIndex]
      if (!pl) return
      this.npuPowerLimit[devIndex].setting = true
      this.npuPowerResult[devIndex] = null
      try {
        const result = await window.go.main.App.SetNPUPowerLimit(devIndex, pl.maxW || 25, pl.minW || 5)
        this.npuPowerResult[devIndex] = result
        if (result.Success) {
          await this.refreshNPU()
        }
      } catch(e) {
        this.npuPowerResult[devIndex] = { Success: false, Message: String(e) }
      } finally {
        this.npuPowerLimit[devIndex].setting = false
      }
    },
    async setNPUClockLock(devIndex) {
      const lock = this.npuClockLocking[devIndex]
      if (!lock) return
      const max = parseInt(lock.maxMhz) || 1400
      const min = parseInt(lock.minMhz) || 700
      if (max < 700 || max > 1400 || min < 700 || min > 1400 || min > max) {
        this.npuPowerResult[devIndex] = { Success: false, Message: '频率必须在 700-1400 MHz 之间，且最小值不能超过最大值' }
        return
      }
      this.npuClockLocking[devIndex].setting = true
      this.npuPowerResult[devIndex] = null
      try {
        const result = await window.go.main.App.SetNPUClockLock(devIndex, max, min)
        this.npuPowerResult[devIndex] = result
        await this.pollNPUPower(devIndex)
      } catch(e) {
        this.npuPowerResult[devIndex] = { Success: false, Message: String(e) }
      } finally {
        this.npuClockLocking[devIndex].setting = false
      }
    },
    async resetNPUDefaults(devIndex) {
      this.npuClockLocking[devIndex].setting = true
      this.npuPowerResult[devIndex] = null
      try {
        const result = await window.go.main.App.ResetNPUDefaults(devIndex)
        this.npuPowerResult[devIndex] = result
        if (result.Success) {
          this.npuClockLocking[devIndex].maxMhz = 1400
          this.npuClockLocking[devIndex].minMhz = 700
        }
        await this.pollNPUPower(devIndex)
        await this.refreshNPU()
      } catch(e) {
        this.npuPowerResult[devIndex] = { Success: false, Message: String(e) }
      } finally {
        this.npuClockLocking[devIndex].setting = false
      }
    },
    async pollNPUPower(devIndex) {
      try {
        const status = await window.go.main.App.GetNPUPowerStatus(devIndex)
        this.npuPowerStatus[devIndex] = status
      } catch(e) { /* silent */ }
    },
    startNPUPowerPolling(devIndex) {
      this.stopNPUPowerPolling()
      this.npuPowerRefreshDev = devIndex
      this.pollNPUPower(devIndex)
      this.npuPowerTimer = setInterval(() => { this.pollNPUPower(this.npuPowerRefreshDev) }, 3000)
    },
    stopNPUPowerPolling() {
      if (this.npuPowerTimer) { clearInterval(this.npuPowerTimer); this.npuPowerTimer = null }
    },
    async showNPUREport() {
      this.npuShowDebug = true
      try {
        const report = await window.go.main.App.GetNPUREport()
        this.npuRawProbe = report || '无诊断报告'
      } catch(e) {
        this.npuRawProbe = '加载报告失败: ' + String(e)
      }
    },
    async startNpuScheduler() {
      this.npuSchedStarting = true
      try {
        await window.go.main.App.StartNPUScheduler(this.npuSchedDev, JSON.stringify({
          utilHighPct: this.npuSchedSettings.utilHighPct || 85,
          utilLowPct: this.npuSchedSettings.utilLowPct || 20,
          tempWarnC: this.npuSchedSettings.tempWarnC || 80,
          tempCritC: this.npuSchedSettings.tempCritC || 90,
          checkSec: this.npuSchedSettings.checkSec || 5,
        }))
        this.npuSchedRunning = true
        this.pollNpuSchedulerState()
      } catch(e) {
        alert('启动调度器失败: ' + String(e))
      } finally {
        this.npuSchedStarting = false
      }
    },
    async stopNpuScheduler() {
      try {
        await window.go.main.App.StopNPUScheduler()
        this.npuSchedRunning = false
      } catch(e) {
        alert('停止调度器失败: ' + String(e))
      }
    },
    async pollNpuSchedulerState() {
      if (!this.npuSchedRunning) return
      try {
        const r = await window.go.main.App.GetNPUSchedulerState()
        this.npuSchedState = {
          curMode: r.Running ? (r.CurMode || '') : '',
          decision: r.Decision || '',
          curUtilPct: r.CurUtilPct || -1,
          curTempC: r.CurTempC || 0,
          curPowerW: r.CurPowerW || 0,
          curFreqMHz: r.CurFreqMHz || 0,
          curLockMaxMHz: r.CurLockMaxMHz || 0,
          curLockMinMHz: r.CurLockMinMHz || 0,
          lastSwitch: r.LastSwitch || '—',
        }
        setTimeout(() => this.pollNpuSchedulerState(), 3000)
      } catch(e) {
        setTimeout(() => this.pollNpuSchedulerState(), 3000)
      }
    },
  }
}
</script>

<style scoped>
/* NPU Function Page Styles */
.npu-page { padding: 0; }
.npu-device-card {
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
  background: var(--bg-secondary);
}
.npu-dev-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 8px;
}
.npu-dev-title {
  display: flex;
  align-items: center;
  gap: 10px;
}
.npu-dev-name {
  font-size: 13px;
  color: var(--text-secondary);
}
.npu-dev-actions {
  display: flex;
  align-items: center;
  gap: 6px;
}
.npu-dvfs-badge {
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
}
.npu-dvfs-performance { background: #dbeafe; color: #1d4ed8; }
.npu-dvfs-ondemand { background: #dcfce7; color: #15803d; }
.npu-dvfs-unknown { background: var(--bg-tertiary); color: var(--text-secondary); }
.npu-metrics-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
  gap: 10px;
  margin-bottom: 12px;
}
.npu-metric {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 8px 10px;
}
.npu-metric-label {
  font-size: 11px;
  color: var(--text-secondary);
  display: block;
  margin-bottom: 4px;
}
.npu-metric-bar-wrap {
  height: 4px;
  background: var(--border-color);
  border-radius: 2px;
  margin-bottom: 4px;
  overflow: hidden;
}
.npu-metric-bar {
  height: 100%;
  border-radius: 2px;
  transition: width .3s ease;
}
.npu-metric-val {
  font-size: 12px;
  font-family: monospace;
  color: var(--text-primary);
}
.npu-info-row {
  display: flex;
  flex-wrap: wrap;
  gap: 8px 20px;
  margin-bottom: 12px;
}
.npu-info-item { display: flex; flex-direction: column; gap: 2px; }
.npu-info-label { font-size: 10px; color: var(--text-secondary); text-transform: uppercase; letter-spacing: .04em; }
.npu-info-val { font-size: 12px; color: var(--text-primary); font-family: monospace; }
.npu-power-status {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 10px;
  margin-bottom: 12px;
}
.npu-power-title { font-size: 12px; font-weight: 600; margin-bottom: 8px; color: var(--text-secondary); }
.npu-power-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 6px;
}
.npu-power-item {
  display: flex;
  justify-content: space-between;
  font-size: 11px;
  padding: 3px 0;
  border-bottom: 1px solid var(--border-color);
}
.npu-power-item span:first-child { color: var(--text-secondary); }
.npu-power-item span:last-child { color: var(--text-primary); font-family: monospace; }
.npu-panel {
  background: var(--bg-primary);
  border: 1px solid var(--border-color);
  border-radius: 6px;
  padding: 12px;
  margin-bottom: 12px;
}
.npu-panel-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 10px;
  color: var(--text-secondary);
}
.npu-clock-row {
  display: flex;
  align-items: flex-end;
  gap: 10px;
  flex-wrap: wrap;
}
.npu-clock-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.npu-clock-field label {
  font-size: 11px;
  color: var(--text-secondary);
}
.npu-sched-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(130px, 1fr));
  gap: 10px;
  margin-bottom: 12px;
}
.npu-sched-field {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.npu-sched-field label {
  font-size: 11px;
  color: var(--text-secondary);
}
.npu-sched-actions {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}
.npu-sched-state {
  background: var(--bg-tertiary);
  border-radius: 6px;
  padding: 12px;
}
.npu-sched-state-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
}
.npu-sched-state-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: 6px;
}
.npu-sched-state-item {
  display: flex;
  justify-content: space-between;
  font-size: 12px;
  padding: 4px 0;
  border-bottom: 1px solid var(--border-color);
}
.npu-sched-state-item span:first-child { color: var(--text-secondary); }
.npu-sched-state-item span:last-child { color: var(--text-primary); font-family: monospace; }
.npu-ref-content {
  background: var(--bg-secondary);
  border: 1px solid var(--border-color);
  border-radius: 8px;
  padding: 16px;
  margin-top: 12px;
}
.npu-ref-section { margin-bottom: 16px; }
.npu-ref-section:last-child { margin-bottom: 0; }
.npu-ref-title {
  font-size: 13px;
  font-weight: 600;
  margin-bottom: 8px;
  color: var(--text-secondary);
}
.npu-ref-table {
  width: 100%;
  border-collapse: collapse;
  font-size: 12px;
}
.npu-ref-table th {
  text-align: left;
  padding: 6px 8px;
  background: var(--bg-tertiary);
  color: var(--text-secondary);
  font-size: 11px;
  text-transform: uppercase;
  letter-spacing: .04em;
}
.npu-ref-table td {
  padding: 6px 8px;
  border-top: 1px solid var(--border-color);
  color: var(--text-primary);
}
.npu-no-device {
  text-align: center;
  padding: 20px;
  color: var(--text-secondary);
  font-style: italic;
}

/* SDK Info Banner */
.sdk-banner {
  display: flex;
  align-items: center;
  gap: 16px;
  margin: 12px 0 16px;
  padding: 12px 20px;
  background: linear-gradient(135deg, var(--bg-secondary) 0%, var(--bg-tertiary) 100%);
  border: 1px solid var(--border-color);
  border-radius: 10px;
  font-size: 13px;
}
.sdk-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: linear-gradient(135deg, #3b82f6 0%, #6366f1 100%);
  color: #fff;
  flex-shrink: 0;
}
.sdk-row {
  display: flex;
  align-items: center;
  gap: 8px;
}
.sdk-field {
  font-size: 11px;
  font-weight: 600;
  text-transform: uppercase;
  letter-spacing: .06em;
  color: var(--text-tertiary);
}
.sdk-val {
  font-family: 'Consolas', 'SF Mono', monospace;
  font-size: 13px;
  color: var(--text-primary);
}
.sdk-divider {
  width: 1px;
  height: 20px;
  background: var(--border-color);
}
</style>
