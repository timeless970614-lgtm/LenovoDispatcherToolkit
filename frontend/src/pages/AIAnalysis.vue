<template>
  <div class="ai-analysis-page">

    <!-- Tab Switcher (always visible) -->
    <div class="ai-tabs">
      <button :class="['ai-tab', { active: activeAIType === 'log' }]" @click="activeAIType = 'log'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
          <polyline points="14 2 14 8 20 8"/>
        </svg>
        Log Analysis
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'etl' }]" @click="switchToETL">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
        ETL Trace
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'toolkit' }]" @click="switchToToolkit">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
        </svg>
        Toolkit
      </button>
    </div>

    <!-- ETL Trace Tab -->
    <div v-if="activeAIType === 'etl'" class="etl-section">

      <!-- Elevated Warning -->
      <div v-if="!isElevated" class="elevated-warning">
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
        </svg>
        <div>
          <strong>Administrator Required</strong>
          <p>ETL trace capture requires running as Administrator. End this process in Task Manager, then right-click the exe and choose "Run as administrator".</p>
        </div>
      </div>

      <!-- Capture Control Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
              <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
            </svg>
            Trace Capture
          </span>
          <span v-if="captureState.isCapturing" class="capture-badge recording">
            <span class="rec-dot"></span> Recording
          </span>
        </div>

        <!-- Profile Selection -->
        <div class="profile-grid">
          <button
            v-for="p in etlProfiles" :key="p.id"
            :class="['profile-btn', { active: selectedProfile === p.id, disabled: captureState.isCapturing }]"
            @click="selectedProfile = p.id" :title="p.description">
            <span class="profile-name">{{ p.name }}</span>
            <span class="profile-desc">{{ p.description }}</span>
          </button>
        </div>

        <!-- Duration & Status Controls -->
        <div class="capture-controls">
          <div class="control-group">
            <label class="ctrl-label">Duration</label>
            <div class="duration-buttons">
              <button v-for="d in [10, 30, 60, 120]" :key="d"
                :class="['dur-btn', { active: captureDuration === d }]"
                @click="captureDuration = d" :disabled="captureState.isCapturing">
                {{ d < 60 ? d + 's' : (d/60) + 'm' }}
              </button>
            </div>
          </div>
          <div class="control-group">
            <label class="ctrl-label">Status</label>
            <span :class="['status-val', captureState.isCapturing ? 'recording' : 'idle']">
              {{ captureState.isCapturing ? 'Recording: ' + captureState.profile : 'Ready' }}
            </span>
          </div>
          <div class="control-group">
            <label class="ctrl-label">Output</label>
            <span class="status-val mono" v-if="captureState.outputPath">{{ truncatePath(captureState.outputPath) }}</span>
            <span class="status-val muted" v-else>C:\Users\Public\ETL_Traces\</span>
          </div>
        </div>

        <!-- Action Buttons -->
        <div class="capture-actions">
          <button v-if="!captureState.isCapturing" class="btn-start" @click="startCapture" :disabled="!isElevated || startingCapture">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none"><circle cx="12" cy="12" r="8"/></svg>
            {{ startingCapture ? 'Starting...' : 'Start Trace' }}
          </button>
          <button v-else class="btn-stop" @click="stopCapture" :disabled="stoppingCapture">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none"><rect x="6" y="6" width="12" height="12" rx="1"/></svg>
            {{ stoppingCapture ? 'Stopping...' : 'Stop Trace' }}
          </button>
          <button class="btn-outline" @click="refreshTraceList" :disabled="refreshing">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
            </svg>
            Refresh List
          </button>
          <button class="btn-outline" @click="openTraceFolder">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
            Open Folder
          </button>
        </div>

        <div v-if="captureError" class="capture-error">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
          </svg>
          {{ captureError }}
        </div>
      </div>

      <!-- Trace History -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
            </svg>
            Trace History ({{ traceList.length }})
          </span>
        </div>
        <div v-if="traceList.length" class="trace-list">
          <div v-for="t in traceList" :key="t.path" class="trace-item"
            :class="{ selected: selectedTrace === t.path }" @click="selectedTrace = t.path">
            <div class="trace-info">
              <span class="trace-path">{{ truncatePath(t.path) }}</span>
              <span class="trace-meta">{{ t.sizeMB }} MB  {{ t.capturedAt }}  {{ t.profileName }}</span>
            </div>
            <button class="btn-analyze-sm" @click.stop="analyzeTrace(t.path)" :disabled="analyzingTrace">Analyze</button>
          </div>
        </div>
        <div v-else class="empty-hint">No traces captured yet. Start a trace above.</div>
      </div>

      <!-- Analysis Results -->
      <div v-if="analysisResult" class="analysis-results">

        <div class="card">
          <div class="card-header">
            <span class="card-title">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
              </svg>
              Analysis Result
            </span>
            <span class="trace-meta mono">{{ truncatePath(analysisResult.traceInfo.path) }}</span>
          </div>
        </div>

        <!-- Summary Banner -->
        <div class="result-summary-banner" v-if="analysisResult.summary">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><path d="M12 8v4M12 16h.01"/>
          </svg>
          {{ analysisResult.summary }}
        </div>

        <!-- Analysis Grid -->
        <div class="analysis-grid">
          <!-- CPU -->
          <div class="analysis-card">
            <div class="card-title-sm">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/>
                <line x1="9" y1="2" x2="9" y2="4"/><line x1="15" y1="2" x2="15" y2="4"/>
                <line x1="9" y1="20" x2="9" y2="22"/><line x1="15" y1="20" x2="15" y2="22"/>
              </svg>
              CPU
            </div>
            <div class="metric-row" v-if="analysisResult.cpu.cpuUsagePct">
              <span class="metric-label">CPU Usage</span>
              <span class="metric-value highlight">{{ analysisResult.cpu.cpuUsagePct }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.cpu.contextSwitches">
              <span class="metric-label">Context Switches</span>
              <span class="metric-value">{{ formatNum(analysisResult.cpu.contextSwitches) }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.cpu.interrupts">
              <span class="metric-label">Interrupts</span>
              <span class="metric-value">{{ formatNum(analysisResult.cpu.interrupts) }}</span>
            </div>
            <div v-if="analysisResult.cpu.busyProcesses?.length" class="proc-list">
              <div class="proc-header">Top Processes</div>
              <div v-for="p in analysisResult.cpu.busyProcesses" :key="p.processName" class="proc-row">
                <span class="proc-name">{{ p.processName }}</span>
                <span class="proc-pct">{{ p.cpuPct.toFixed(1) }}%</span>
              </div>
            </div>
            <div v-else class="no-data-sm">Run with xperf profile for process details</div>
          </div>

          <!-- Disk I/O -->
          <div class="analysis-card">
            <div class="card-title-sm">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
              Disk I/O
            </div>
            <div class="metric-row" v-if="analysisResult.disk.totalReadMB">
              <span class="metric-label">Total Read</span>
              <span class="metric-value">{{ analysisResult.disk.totalReadMB }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.disk.totalWrittenMB">
              <span class="metric-label">Total Written</span>
              <span class="metric-value">{{ analysisResult.disk.totalWrittenMB }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.disk.readOpsPerSec">
              <span class="metric-label">Read Ops/s</span>
              <span class="metric-value">{{ analysisResult.disk.readOpsPerSec }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.disk.avgLatencyMs">
              <span class="metric-label">Avg Latency</span>
              <span class="metric-value">{{ analysisResult.disk.avgLatencyMs }}</span>
            </div>
          </div>

          <!-- Network -->
          <div class="analysis-card">
            <div class="card-title-sm">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M5 12.55a11 11 0 0 1 14.08 0M1.42 9a16 16 0 0 1 21.16 0M8.53 16.11a6 6 0 0 1 6.95 0M12 20h.01"/>
              </svg>
              Network
            </div>
            <div class="metric-row" v-if="analysisResult.network.totalSentMB">
              <span class="metric-label">Sent</span>
              <span class="metric-value">{{ analysisResult.network.totalSentMB }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.network.totalRecvMB">
              <span class="metric-label">Received</span>
              <span class="metric-value">{{ analysisResult.network.totalRecvMB }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.network.tcpConnections">
              <span class="metric-label">TCP Connections</span>
              <span class="metric-value">{{ analysisResult.network.tcpConnections }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.gpu.gpuEngineUtilPct">
              <span class="metric-label">GPU Util</span>
              <span class="metric-value">{{ analysisResult.gpu.gpuEngineUtilPct }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.gpu.gpuMemoryUsedMB">
              <span class="metric-label">GPU Memory</span>
              <span class="metric-value">{{ analysisResult.gpu.gpuMemoryUsedMB }}</span>
            </div>
          </div>

          <!-- Power -->
          <div class="analysis-card">
            <div class="card-title-sm">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
              </svg>
              Power
            </div>
            <div class="metric-row" v-if="analysisResult.power.cpuPower">
              <span class="metric-label">CPU Power</span>
              <span class="metric-value">{{ analysisResult.power.cpuPower }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.power.gpuPower">
              <span class="metric-label">GPU Power</span>
              <span class="metric-value">{{ analysisResult.power.gpuPower }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.power.packagePower">
              <span class="metric-label">Package Power</span>
              <span class="metric-value">{{ analysisResult.power.packagePower }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.power.s0ixDuration">
              <span class="metric-label">S0ix Duration</span>
              <span class="metric-value">{{ analysisResult.power.s0ixDuration }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.power.s0ixTransitions">
              <span class="metric-label">S0ix Transitions</span>
              <span class="metric-value">{{ analysisResult.power.s0ixTransitions }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.power.processorFreqMHz">
              <span class="metric-label">CPU Freq</span>
              <span class="metric-value">{{ analysisResult.power.processorFreqMHz }}</span>
            </div>
          </div>

          <!-- DPC/ISR -->
          <div class="analysis-card">
            <div class="card-title-sm">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/>
              </svg>
              DPC / ISR Latency
            </div>
            <div class="metric-row" v-if="analysisResult.dpcrisr.avgDpcMs">
              <span class="metric-label">Avg DPC</span>
              <span class="metric-value">{{ analysisResult.dpcrisr.avgDpcMs }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.dpcrisr.maxDpcMs">
              <span class="metric-label">Max DPC</span>
              <span class="metric-value warning">{{ analysisResult.dpcrisr.maxDpcMs }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.dpcrisr.avgIsrMs">
              <span class="metric-label">Avg ISR</span>
              <span class="metric-value">{{ analysisResult.dpcrisr.avgIsrMs }}</span>
            </div>
            <div class="metric-row" v-if="analysisResult.dpcrisr.maxIsrMs">
              <span class="metric-label">Max ISR</span>
              <span class="metric-value warning">{{ analysisResult.dpcrisr.maxIsrMs }}</span>
            </div>
            <div v-if="analysisResult.dpcrisr.highDpcLatencyProcs?.length" class="proc-list">
              <div class="proc-header">High DPC Latency</div>
              <div v-for="p in analysisResult.dpcrisr.highDpcLatencyProcs" :key="p.processName" class="proc-row">
                <span class="proc-name">{{ p.processName || p.module || 'Unknown' }}</span>
                <span class="proc-pct warning">{{ p.maxLatencyMs }}</span>
              </div>
            </div>
          </div>
        </div>

        <!-- Raw CSV -->
        <div class="card" v-if="analysisResult.rawCSVLines?.length">
          <div class="card-header">
            <span class="card-title">
              <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
                <polyline points="4 17 10 11 4 5"/><line x1="12" y1="19" x2="20" y2="19"/>
              </svg>
              Raw CSV Summary ({{ analysisResult.rawCSVLines.length }} lines)
            </span>
          </div>
          <div class="raw-csv-box">
            <pre>{{ analysisResult.rawCSVLines.join('\n') }}</pre>
          </div>
        </div>

      </div><!-- /analysis-results -->

      <!-- Analyzing Overlay -->
      <div v-if="analyzingTrace" class="analyzing-overlay">
        <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
          <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
        </svg>
        <span>Parsing ETL trace with tracerpt...</span>
      </div>

    </div><!-- /ETL section -->

    <!-- Log Analysis Tab -->
    <div v-if="activeAIType === 'log'" class="log-section">
      <!-- Log Files Card - Password Protected -->
      <div class="card advanced-section">
        <div class="card-header advanced-toggle" @click="toggleLogSection">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/>
            </svg>
            Log Files
          </span>
          <svg :class="['chevron', { open: logSectionExpanded }]" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
        </div>

        <!-- Password Prompt -->
        <div v-if="logSectionExpanded && !logUnlocked" class="password-prompt">
          <div class="password-row">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right: 8px; flex-shrink:0;">
              <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/>
              <path d="M7 11V7a5 5 0 0110 0v4"/>
            </svg>
            <input
              type="password"
              class="password-input"
              v-model="logPassword"
              placeholder="Enter password to unlock"
              @keydown.enter="unlockLogs"
              ref="logPasswordInput"
            />
            <button class="btn-unlock" @click="unlockLogs">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="3" y="11" width="18" height="11" rx="2" ry="2"/><path d="M7 11V7a5 5 0 0110 0v4"/>
              </svg>
              Unlock
            </button>
          </div>
          <div v-if="logPasswordError" class="password-error">{{ logPasswordError }}</div>
        </div>

        <!-- Unlocked Content -->
        <div v-if="logSectionExpanded && logUnlocked">
          <div style="display:flex;justify-content:flex-end;padding:0 0 8px 0;">
            <button class="btn-sm" @click="loadLogs">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="23 4 23 10 17 10"/><path d="M20.49 15a9 9 0 1 1-2.12-9.36L23 10"/>
              </svg>
              Refresh
            </button>
          </div>
          <div v-if="logFiles.length" class="log-file-list">
            <div v-for="f in logFiles" :key="f.name" class="log-file-item"
              :class="{ active: selectedLog === f.name }" @click="selectLog(f.name)">
              <span class="log-name">{{ f.name }}</span>
              <span class="log-meta">{{ formatSize(f.size) }}  {{ f.modTime }}</span>
            </div>
          </div>
          <div v-else class="empty-hint">No log files found in {{ logDir }}</div>
        </div>
      </div>

      <!-- AI Analysis Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
              <circle cx="12" cy="12" r="10"/><path d="M12 8v4l3 3"/>
            </svg>
            AI Analysis
          </span>
          <button class="btn-analyze" @click="runAnalysis" :disabled="analyzing">
            <svg v-if="!analyzing" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polygon points="5 3 19 12 5 21 5 3"/>
            </svg>
            <svg v-else width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            {{ analyzing ? 'Analyzing...' : 'Analyze Latest Log' }}
          </button>
        </div>
        <div v-if="summary" class="summary-grid">
          <div class="summary-panel">
            <div class="panel-title">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
              </svg>
              Workload Levels
            </div>
            <div class="log-lines">
              <div v-for="(l,i) in summary.workloadLevels" :key="i" class="log-line">{{ l }}</div>
              <div v-if="!summary.workloadLevels?.length" class="no-data">No data</div>
            </div>
          </div>
          <div class="summary-panel">
            <div class="panel-title">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
              </svg>
              Turbo Events
            </div>
            <div class="log-lines">
              <div v-for="(l,i) in summary.turboEvents" :key="i" class="log-line">{{ l }}</div>
              <div v-if="!summary.turboEvents?.length" class="no-data">No data</div>
            </div>
          </div>
          <div class="summary-panel">
            <div class="panel-title">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
              </svg>
              Foreground Apps
            </div>
            <div class="log-lines">
              <div v-for="(l,i) in summary.appEvents" :key="i" class="log-line">{{ l }}</div>
              <div v-if="!summary.appEvents?.length" class="no-data">No data</div>
            </div>
          </div>
          <div class="summary-panel panel-errors">
            <div class="panel-title error-title">
              <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              Errors / Warnings
            </div>
            <div class="log-lines">
              <div v-for="(l,i) in summary.errors" :key="i" class="log-line error-line">{{ l }}</div>
              <div v-if="!summary.errors?.length" class="no-data ok-hint">OK - No errors found</div>
            </div>
          </div>
        </div>
        <div v-else-if="!analyzing" class="empty-hint">Click "Analyze Latest Log" to start</div>
        <div v-else class="analyzing-hint">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
            <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
          </svg>
          Reading and parsing log...
        </div>
      </div>

    </div><!-- /log-section -->

    <!-- Toolkit Tab -->
    <div v-if="activeAIType === 'toolkit'" class="toolkit-section">
      
      <!-- Header Card -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
              <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
            </svg>
            System Tools - One-Click Install
          </span>
          <button class="btn-sm" @click="openToolkitFolder">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
            Open Folder
          </button>
        </div>
        <div class="toolkit-path">Install Directory: {{ toolkitInstallDir }}</div>
      </div>

      <!-- Category Filter -->
      <div class="category-filter">
        <button :class="['cat-btn', { active: toolkitCategory === 'all' }]" @click="toolkitCategory = 'all'">All ({{ toolkitTools.length }})</button>
        <button :class="['cat-btn', { active: toolkitCategory === 'monitor' }]" @click="toolkitCategory = 'monitor'">Monitor</button>
        <button :class="['cat-btn', { active: toolkitCategory === 'system' }]" @click="toolkitCategory = 'system'">System Info</button>
        <button :class="['cat-btn', { active: toolkitCategory === 'benchmark' }]" @click="toolkitCategory = 'benchmark'">Benchmark</button>
        <button :class="['cat-btn', { active: toolkitCategory === 'diagnostic' }]" @click="toolkitCategory = 'diagnostic'">Diagnostic</button>
      </div>

      <!-- Tools Grid -->
      <div class="tools-grid">
        <div v-for="tool in filteredTools" :key="tool.id" class="tool-card" :class="{ installed: toolInstallStatus[tool.id]?.installed }">
          <div class="tool-header">
            <div class="tool-icon" :class="tool.category">
              <svg v-if="tool.category === 'monitor'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M22 12h-4l-3 9L9 3l-3 9H2"/>
              </svg>
              <svg v-else-if="tool.category === 'system'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <rect x="4" y="4" width="16" height="16" rx="2"/><rect x="9" y="9" width="6" height="6"/>
              </svg>
              <svg v-else-if="tool.category === 'benchmark'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <polygon points="13 2 3 14 12 14 11 22 21 10 12 10 13 2"/>
              </svg>
              <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
            </div>
            <div class="tool-info">
              <span class="tool-name">{{ tool.name }}</span>
              <span class="tool-vendor">{{ tool.vendor }}</span>
            </div>
            <span v-if="toolInstallStatus[tool.id]?.installed" class="installed-badge">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3">
                <polyline points="20 6 9 17 4 12"/>
              </svg>
              Installed
            </span>
          </div>
          
          <div class="tool-desc">{{ tool.description }}</div>
          
          <div class="tool-meta">
            <span class="tool-version">v{{ tool.version }}</span>
            <span class="tool-size">{{ tool.sizeMb }} MB</span>
            <span class="tool-winget" v-if="tool.wingetId">📦 winget</span>
          </div>

          <!-- Installing Status -->
          <div v-if="isToolBusy(tool.id)" class="tool-installing">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
              <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
            </svg>
            <span>{{ toolProgress[tool.id]?.message || 'Installing...' }}</span>
          </div>

          <!-- Error Message -->
          <div v-if="toolProgress[tool.id]?.status === 'error'" class="tool-error">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ toolProgress[tool.id].message }}
          </div>

          <!-- Success Message -->
          <div v-if="toolProgress[tool.id]?.status === 'completed'" class="tool-success">
            <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            {{ toolProgress[tool.id].message }}
          </div>

          <div class="tool-actions">
            <button v-if="!toolInstallStatus[tool.id]?.installed && !isToolBusy(tool.id)" class="btn-install" @click="installTool(tool.id)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
                <polyline points="7 10 12 15 17 10"/>
                <line x1="12" y1="15" x2="12" y2="3"/>
              </svg>
              Install & Run
            </button>
            <button v-if="isToolBusy(tool.id)" class="btn-installing" disabled>
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" class="spinning">
                <path d="M21 12a9 9 0 1 1-6.219-8.56"/>
              </svg>
              Installing...
            </button>
            <button v-if="toolInstallStatus[tool.id]?.installed" class="btn-run" @click="runTool(tool.id)">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" stroke="none">
                <polygon points="5 3 19 12 5 21 5 3"/>
              </svg>
              Run
            </button>
            <a :href="tool.website" target="_blank" class="btn-link">
              <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/>
                <polyline points="15 3 21 3 21 9"/>
                <line x1="10" y1="14" x2="21" y2="3"/>
              </svg>
              Website
            </a>
          </div>
        </div>
      </div>

      <!-- Quick Install All -->
      <div class="card" v-if="!allToolsInstalled">
        <div class="card-header">
          <span class="card-title">Quick Actions</span>
        </div>
        <div class="quick-actions">
          <button class="btn-batch" @click="installEssentials" :disabled="batchInstalling">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            {{ batchInstalling ? 'Installing...' : 'Install Essential Tools (HWiNFO, CPU-Z, GPU-Z)' }}
          </button>
        </div>
      </div>

    </div><!-- /toolkit-section -->

  </div>
</template>

<script>
import { OpenFolder } from '../../wailsjs/go/main/App'

export default {
  name: 'AIAnalysis',
  beforeDestroy() {
    this._capturePollTimer && clearInterval(this._capturePollTimer)
  },
  data() {
    return {
      activeAIType: 'log',
      // ETL
      etlProfiles: [],
      selectedProfile: 'GeneralProfile',
      captureDuration: 30,
      captureState: { isCapturing: false, profile: '', outputPath: '', status: '', error: '' },
      captureError: '',
      startingCapture: false,
      stoppingCapture: false,
      traceList: [],
      selectedTrace: '',
      analysisResult: null,
      analyzingTrace: false,
      refreshing: false,
      isElevated: false,
      // Log
      logDir: 'C:\\ProgramData\\Lenovo\\LenovoDispatcher\\Logs',
      logFiles: [],
      logSectionExpanded: true,
      logUnlocked: false,
      logPassword: '',
      logPasswordError: '',
      selectedLog: null,
      summary: null,
      rawLog: '',
      tailLines: 200,
      analyzing: false,
      // Toolkit
      toolkitTools: [],
      toolkitInstallDir: 'C:\\LenovoDispatcherToolkit\\Tools',
      toolkitCategory: 'all',
      toolInstallStatus: {},
      toolProgress: {},
      batchInstalling: false,
    }
  },
  async mounted() {
    await this.loadLogs()
  },
  methods: {
    async switchToETL() {
      this.activeAIType = 'etl'
      if (!this.etlProfiles.length) await this.loadETLData()
      // Check elevation status from backend
      try {
        if (window.go?.main?.App) {
          this.isElevated = await window.go.main.App.IsElevated()
        }
      } catch (e) { console.error(e) }
    },
    async loadETLData() {
      try {
        if (window.go?.main?.App) {
          const [profiles, status, traces] = await Promise.all([
            window.go.main.App.GetETLProfiles(),
            window.go.main.App.GetETLCaptureStatus(),
            window.go.main.App.GetETLTraceList(),
          ])
          this.etlProfiles = profiles || []
          this.captureState = status || { isCapturing: false }
          this.traceList = traces || []
        }
      } catch (e) { console.error(e) }
    },
    async startCapture() {
      this.startingCapture = true
      this.captureError = ''
      this._capturePollTimer && clearInterval(this._capturePollTimer)
      try {
        if (window.go?.main?.App) {
          const result = await window.go.main.App.StartETLCapture(this.selectedProfile, this.captureDuration)
          this.captureState = result
          if (result.error) this.captureError = result.error
          else if (!result.isCapturing) this.captureError = result.error || 'Failed to start capture'
          else {
            // Poll backend every 2s to detect when goroutine auto-stops the trace
            this._capturePollTimer = setInterval(async () => {
              try {
                const status = await window.go.main.App.GetETLCaptureStatus()
                if (status && !status.isCapturing) {
                  clearInterval(this._capturePollTimer)
                  this._capturePollTimer = null
                  this.captureState = status
                }
              } catch (e) { /* ignore polling errors */ }
            }, 2000)
          }
        }
      } catch (e) { this.captureError = e.message || String(e) }
      finally { this.startingCapture = false }
    },
    async stopCapture() {
      this.stoppingCapture = true
      this._capturePollTimer && clearInterval(this._capturePollTimer)
      try {
        if (window.go?.main?.App) {
          const traceInfo = await window.go.main.App.StopETLCapture()
          this.captureState = { isCapturing: false }
          if (traceInfo.path) {
            await this.refreshTraceList()
            await this.analyzeTrace(traceInfo.path)
          }
        }
      } catch (e) { this.captureError = e.message || String(e) }
      finally { this.stoppingCapture = false }
    },
    async refreshTraceList() {
      this.refreshing = true
      try {
        if (window.go?.main?.App) this.traceList = await window.go.main.App.GetETLTraceList() || []
      } catch (e) { console.error(e) }
      finally { this.refreshing = false }
    },
    async openTraceFolder() {
      try { await OpenFolder('C:\\Users\\Public\\ETL_Traces') } catch (e) { console.error(e) }
    },
    async analyzeTrace(path) {
      this.analyzingTrace = true
      this.analysisResult = null
      try {
        if (window.go?.main?.App) this.analysisResult = await window.go.main.App.AnalyzeETLFile(path || this.selectedTrace)
      } catch (e) { console.error(e) }
      finally { this.analyzingTrace = false }
    },
    async loadLogs() {
      try {
        if (window.go?.main?.App) this.logFiles = await window.go.main.App.GetLogFiles() || []
      } catch (e) { console.error(e) }
    },
    selectLog(name) { this.selectedLog = name },
    toggleLogSection() {
      this.logSectionExpanded = !this.logSectionExpanded
    },
    unlockLogs() {
      if (this.logPassword === 'Lenovo2026') {
        this.logUnlocked = true
        this.logPasswordError = ''
        this.logPassword = ''
        this.loadLogs()
      } else {
        this.logPasswordError = 'Need Dispatcher owner check or contact zhoushang2'
        setTimeout(() => { this.logPasswordError = '' }, 3000)
      }
    },
    async runAnalysis() {
      this.analyzing = true
      this.summary = null
      try {
        if (window.go?.main?.App) this.summary = await window.go.main.App.GetLogSummary()
      } catch (e) { console.error(e) }
      finally { this.analyzing = false }
    },
    async loadRawLog() {
      try {
        if (window.go?.main?.App) this.rawLog = await window.go.main.App.ReadLogTail(this.tailLines)
      } catch (e) { console.error(e) }
    },
    formatSize(bytes) {
      if (bytes < 1024) return bytes + ' B'
      if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(1) + ' KB'
      return (bytes / 1024 / 1024).toFixed(1) + ' MB'
    },
    formatNum(n) {
      if (n >= 1e9) return (n / 1e9).toFixed(1) + 'B'
      if (n >= 1e6) return (n / 1e6).toFixed(1) + 'M'
      if (n >= 1000) return (n / 1000).toFixed(1) + 'K'
      return String(n)
    },
    truncatePath(path) {
      if (!path) return ''
      return path.length <= 65 ? path : '...' + path.slice(-62)
    },
    // ============ Toolkit Methods ============
    async switchToToolkit() {
      this.activeAIType = 'toolkit'
      if (!this.toolkitTools.length) await this.loadToolkitData()
    },
    async loadToolkitData() {
      try {
        if (window.go?.main?.App) {
          this.toolkitTools = await window.go.main.App.GetToolkitTools() || []
          this.toolkitInstallDir = await window.go.main.App.GetToolkitInstallDir() || this.toolkitInstallDir
          await this.refreshToolkitStatus()
        }
      } catch (e) { console.error(e) }
    },
    async refreshToolkitStatus() {
      try {
        if (window.go?.main?.App) {
          const statuses = await window.go.main.App.CheckAllToolkitInstalled() || []
          this.toolInstallStatus = {}
          statuses.forEach(s => {
            this.toolInstallStatus[s.toolId] = s
          })
        }
      } catch (e) { console.error(e) }
    },
    async installTool(toolId) {
      if (!window.go?.main?.App) return
      // Set installing state
      this.toolProgress[toolId] = { status: 'installing', progress: 0, message: 'Installing via winget...' }
      
      try {
        // Start installation (runs async in backend)
        await window.go.main.App.InstallToolkitTool(toolId)
        
        // Poll for progress
        const pollProgress = async () => {
          if (!this.toolProgress[toolId] || this.toolProgress[toolId].status === 'completed' || this.toolProgress[toolId].status === 'error') {
            return
          }
          try {
            const progress = await window.go.main.App.GetToolkitProgress(toolId)
            this.toolProgress[toolId] = progress
            if (progress.status === 'installing') {
              setTimeout(pollProgress, 1000)
            } else if (progress.status === 'completed') {
              await this.refreshToolkitStatus()
            }
          } catch (e) {
            console.error('Poll error:', e)
          }
        }
        setTimeout(pollProgress, 500)
      } catch (e) {
        this.toolProgress[toolId] = { status: 'error', message: e.message || String(e) }
      }
    },
    async runTool(toolId) {
      try {
        if (window.go?.main?.App) {
          const result = await window.go.main.App.RunToolkitTool(toolId)
          if (result.startsWith('Error')) {
            alert(result)
          }
        }
      } catch (e) { console.error(e) }
    },
    async openToolkitFolder() {
      try {
        if (window.go?.main?.App) {
          await window.go.main.App.OpenToolkitFolder()
        }
      } catch (e) { console.error(e) }
    },
    isToolBusy(toolId) {
      const p = this.toolProgress[toolId]
      return p && p.status === 'installing'
    },
    async installEssentials() {
      this.batchInstalling = true
      const essentials = ['hwinfo64', 'cpuz', 'gpuz']
      for (const id of essentials) {
        if (!this.toolInstallStatus[id]?.installed) {
          await this.installTool(id)
          // Wait for completion
          let attempts = 0
          while (this.isToolBusy(id) && attempts < 120) {
            await new Promise(r => setTimeout(r, 1000))
            attempts++
          }
        }
      }
      this.batchInstalling = false
    },
  },
  computed: {
    filteredTools() {
      if (this.toolkitCategory === 'all') return this.toolkitTools
      return this.toolkitTools.filter(t => t.category === this.toolkitCategory)
    },
    allToolsInstalled() {
      return this.toolkitTools.every(t => this.toolInstallStatus[t.id]?.installed)
    }
  }
}
</script>

<style scoped>
.ai-analysis-page { display: flex; flex-direction: column; gap: 16px; }

.ai-tabs {
  display: flex; gap: 8px;
}
.ai-tab {
  display: flex; align-items: center; gap: 6px;
  padding: 8px 16px;
  background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 8px; color: var(--text-secondary);
  font-size: 13px; font-weight: 500; cursor: pointer;
  transition: var(--transition); font-family: inherit;
}
.ai-tab:hover { border-color: var(--lenovo-red); color: var(--text-primary); }
.ai-tab.active {
  background: linear-gradient(90deg, rgba(230,63,50,0.15) 0%, rgba(230,63,50,0.05) 100%);
  border-color: var(--lenovo-red); color: var(--text-primary);
}

/* ETL */
.etl-section, .log-section { display: flex; flex-direction: column; gap: 16px; }

.elevated-warning {
  display: flex; gap: 12px; padding: 14px 16px;
  background: rgba(255,152,0,0.08); border: 1px solid rgba(255,152,0,0.3);
  border-radius: 8px; color: #FF9800; font-size: 12px; line-height: 1.6;
}
.elevated-warning svg { flex-shrink: 0; margin-top: 2px; }
.elevated-warning strong { display: block; margin-bottom: 4px; font-size: 13px; }
.elevated-warning p { margin: 0; opacity: 0.85; }

.profile-grid {
  display: grid; grid-template-columns: repeat(auto-fill, minmax(140px, 1fr));
  gap: 8px; margin-bottom: 16px;
}
.profile-btn {
  display: flex; flex-direction: column; align-items: center; gap: 4px;
  padding: 12px 8px; background: var(--bg-tertiary);
  border: 1px solid var(--border-color); border-radius: 8px;
  cursor: pointer; transition: var(--transition); text-align: center; font-family: inherit;
}
.profile-btn:hover:not(.disabled) { border-color: var(--lenovo-red); background: rgba(230,63,50,0.06); }
.profile-btn.active { border-color: var(--lenovo-red); background: rgba(230,63,50,0.12); }
.profile-btn.disabled { opacity: 0.5; cursor: not-allowed; }
.profile-icon { font-size: 20px; }
.profile-name { font-size: 12px; font-weight: 600; color: var(--text-primary); }
.profile-desc { font-size: 10px; color: var(--text-tertiary); line-height: 1.3; }

.capture-controls {
  display: flex; gap: 24px; flex-wrap: wrap; margin-bottom: 16px;
  padding: 12px 16px; background: var(--bg-tertiary);
  border-radius: 8px; border: 1px solid var(--border-color);
}
.control-group { display: flex; flex-direction: column; gap: 6px; }
.ctrl-label { font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-tertiary); }
.duration-buttons { display: flex; gap: 4px; }
.dur-btn {
  padding: 4px 10px; background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 5px; color: var(--text-secondary); font-size: 12px; font-weight: 500;
  cursor: pointer; transition: var(--transition); font-family: inherit;
}
.dur-btn.active { background: var(--lenovo-red); border-color: var(--lenovo-red); color: white; }
.dur-btn:disabled { opacity: 0.4; cursor: not-allowed; }
.status-val { font-size: 13px; font-weight: 500; color: var(--text-primary); }
.status-val.recording { color: #4CAF50; }
.status-val.idle { color: var(--text-tertiary); }
.status-val.mono { font-family: 'Consolas','Monaco',monospace; font-size: 11px; }

.capture-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.btn-start, .btn-stop {
  padding: 9px 20px; border: none; border-radius: 8px;
  color: white; font-size: 13px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; gap: 8px; transition: var(--transition); font-family: inherit;
}
.btn-start { background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%); }
.btn-stop { background: linear-gradient(135deg, #F44336 0%, #C62828 100%); }
.btn-start:hover:not(:disabled), .btn-stop:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-start:disabled, .btn-stop:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-outline {
  padding: 9px 16px; background: var(--bg-tertiary); border: 1px solid var(--border-color);
  border-radius: 8px; color: var(--text-secondary); font-size: 13px; font-weight: 500;
  cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit;
}
.btn-outline:hover:not(:disabled) { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }
.btn-outline:disabled { opacity: 0.4; cursor: not-allowed; }

.capture-badge {
  display: flex; align-items: center; gap: 6px; padding: 4px 10px;
  border-radius: 12px; font-size: 12px; font-weight: 600;
}
.capture-badge.recording { background: rgba(76,175,80,0.15); color: #4CAF50; border: 1px solid rgba(76,175,80,0.3); }
.rec-dot { width: 8px; height: 8px; border-radius: 50%; background: #4CAF50; animation: blink 1s infinite; }
@keyframes blink { 0%,100%{opacity:1} 50%{opacity:0.3} }

.capture-error {
  display: flex; align-items: flex-start; gap: 8px; margin-top: 12px;
  padding: 10px 14px; background: rgba(244,67,54,0.08);
  border: 1px solid rgba(244,67,54,0.2); border-radius: 6px;
  color: #F44336; font-size: 12px; line-height: 1.5;
}
.capture-error svg { flex-shrink: 0; margin-top: 1px; }

.trace-list { display: flex; flex-direction: column; gap: 4px; }
.trace-item {
  display: flex; justify-content: space-between; align-items: center;
  padding: 10px 12px; border-radius: 6px; cursor: pointer;
  transition: background 0.15s; border: 1px solid transparent;
}
.trace-item:hover { background: var(--bg-card-hover); }
.trace-item.selected { background: rgba(230,63,50,0.08); border-color: rgba(230,63,50,0.25); }
.trace-info { display: flex; flex-direction: column; gap: 3px; min-width: 0; }
.trace-path { font-size: 12px; font-family: 'Consolas','Monaco',monospace; color: var(--text-primary); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; max-width: 500px; }
.trace-meta { font-size: 11px; color: var(--text-tertiary); }
.btn-analyze-sm {
  padding: 5px 12px; background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border: none; border-radius: 5px; color: white; font-size: 11px; font-weight: 600;
  cursor: pointer; white-space: nowrap; transition: var(--transition); font-family: inherit;
}
.btn-analyze-sm:hover:not(:disabled) { opacity: 0.85; }
.btn-analyze-sm:disabled { opacity: 0.4; cursor: not-allowed; }

.analysis-results { display: flex; flex-direction: column; gap: 12px; }
.result-summary-banner {
  display: flex; align-items: center; gap: 10px; padding: 12px 16px;
  background: linear-gradient(90deg, rgba(230,63,50,0.08) 0%, rgba(230,63,50,0.03) 100%);
  border: 1px solid rgba(230,63,50,0.2); border-radius: 8px;
  color: var(--text-secondary); font-size: 13px; font-family: 'Consolas','Monaco',monospace;
}
.analysis-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(240px, 1fr)); gap: 12px; }
.analysis-card {
  background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 8px; padding: 14px; display: flex; flex-direction: column; gap: 6px;
}
.card-title-sm {
  display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.8px; color: var(--lenovo-red);
  margin-bottom: 6px; padding-bottom: 8px; border-bottom: 1px solid var(--border-color);
}
.metric-row { display: flex; justify-content: space-between; align-items: center; }
.metric-label { font-size: 12px; color: var(--text-secondary); }
.metric-value { font-size: 12px; font-weight: 600; font-family: 'Consolas','Monaco',monospace; color: var(--text-primary); }
.metric-value.highlight { color: var(--lenovo-red); }
.metric-value.warning { color: #FF9800; }
.proc-list { margin-top: 4px; }
.proc-header { font-size: 10px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-tertiary); margin-bottom: 4px; }
.proc-row { display: flex; justify-content: space-between; align-items: center; padding: 3px 0; border-bottom: 1px solid var(--border-color); }
.proc-row:last-child { border-bottom: none; }
.proc-name { font-size: 11px; font-family: 'Consolas','Monaco',monospace; color: var(--text-secondary); }
.proc-pct { font-size: 11px; font-weight: 600; font-family: 'Consolas','Monaco',monospace; color: var(--text-primary); }
.proc-pct.warning { color: #FF9800; }
.no-data-sm { font-size: 11px; color: var(--text-tertiary); font-style: italic; }

.analyzing-overlay {
  display: flex; align-items: center; justify-content: center; gap: 12px;
  padding: 32px; background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 8px; color: var(--text-secondary); font-size: 13px;
}
.raw-csv-box { background: var(--bg-tertiary); border-radius: 8px; padding: 16px; max-height: 300px; overflow-y: auto; border: 1px solid var(--border-color); }
.raw-csv-box pre { font-family: 'Consolas','Monaco',monospace; font-size: 11px; color: var(--text-secondary); line-height: 1.6; white-space: pre-wrap; word-break: break-all; margin: 0; }

/* Log */
.log-file-list { display: flex; flex-direction: column; gap: 4px; max-height: 220px; overflow-y: auto; }
.log-file-item { display: flex; justify-content: space-between; align-items: center; padding: 8px 12px; border-radius: 6px; cursor: pointer; transition: background 0.15s; border: 1px solid transparent; }
.log-file-item:hover { background: var(--bg-card-hover); }
.log-file-item.active { background: rgba(230,63,50,0.08); border-color: rgba(230,63,50,0.25); }
.log-name { font-size: 13px; font-weight: 500; color: var(--text-primary); font-family: 'Consolas','Monaco',monospace; }
.log-meta { font-size: 11px; color: var(--text-tertiary); white-space: nowrap; margin-left: 12px; }

/* Log Files - Password Protected */
.advanced-section { border-color: rgba(245, 158, 11, 0.3); }
.advanced-toggle { cursor: pointer; user-select: none; }
.advanced-toggle .chevron { transition: transform 0.2s; margin-left: auto; }
.advanced-toggle .chevron.open { transform: rotate(180deg); }
.password-prompt { padding: 16px 0 0 0; }
.password-row { display: flex; align-items: center; gap: 8px; }
.password-input { flex: 1; padding: 8px 12px; border: 1px solid var(--border-color); border-radius: 6px; background: var(--bg-primary); color: var(--text-primary); font-size: 13px; outline: none; transition: border-color 0.2s; }
.password-input:focus { border-color: #F59E0B; }
.password-input::placeholder { color: var(--text-tertiary); font-family: inherit; }
.btn-unlock { display: flex; align-items: center; gap: 4px; padding: 8px 16px; border: 1px solid #F59E0B; border-radius: 6px; background: rgba(245,158,11,0.1); color: #F59E0B; font-size: 13px; font-weight: 600; cursor: pointer; transition: all 0.2s; white-space: nowrap; }
.btn-unlock:hover { background: rgba(245,158,11,0.2); border-color: #F59E0B; }
.password-error { margin-top: 8px; font-size: 12px; color: #EF4444; }

.summary-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 16px; margin-top: 4px; }
.summary-panel { background: var(--bg-tertiary); border-radius: 8px; padding: 14px; border: 1px solid var(--border-color); }
.panel-errors { border-color: rgba(244,67,54,0.2); background: rgba(244,67,54,0.04); }
.panel-title { font-size: 11px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.8px; color: var(--text-secondary); margin-bottom: 10px; display: flex; align-items: center; gap: 6px; }
.error-title { color: #F44336; }
.log-lines { display: flex; flex-direction: column; gap: 4px; max-height: 160px; overflow-y: auto; }
.log-line { font-size: 11px; font-family: 'Consolas','Monaco',monospace; color: var(--text-secondary); line-height: 1.5; word-break: break-all; padding: 2px 0; border-bottom: 1px solid var(--border-color); }
.log-line:last-child { border-bottom: none; }
.error-line { color: #F44336; }
.no-data { font-size: 12px; color: var(--text-tertiary); font-style: italic; }
.ok-hint { color: var(--accent-green); font-style: normal; font-weight: 600; }

.raw-log-box { background: var(--bg-tertiary); border-radius: 8px; padding: 16px; max-height: 400px; overflow-y: auto; border: 1px solid var(--border-color); }
.raw-log-box pre { font-family: 'Consolas','Monaco',monospace; font-size: 11px; color: var(--text-secondary); line-height: 1.6; white-space: pre-wrap; word-break: break-all; margin: 0; }

.btn-sm { padding: 6px 12px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 6px; color: var(--text-secondary); font-size: 12px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit; }
.btn-sm:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }
.btn-analyze { padding: 7px 16px; background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%); border: none; border-radius: 6px; color: white; font-size: 12px; font-weight: 600; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit; }
.btn-analyze:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-analyze:disabled { opacity: 0.6; cursor: not-allowed; }
.header-controls { display: flex; align-items: center; gap: 8px; }
.lines-select { padding: 5px 8px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 6px; color: var(--text-primary); font-size: 12px; font-family: inherit; cursor: pointer; }
.empty-hint { padding: 24px; text-align: center; color: var(--text-tertiary); font-size: 13px; }
.analyzing-hint { display: flex; align-items: center; justify-content: center; gap: 10px; padding: 32px; color: var(--text-secondary); font-size: 13px; }
@keyframes spin { to { transform: rotate(360deg); } }
.spinning { animation: spin 0.8s linear infinite; }

/* Toolkit Styles */
.toolkit-section { display: flex; flex-direction: column; gap: 16px; }
.toolkit-path { font-size: 12px; color: var(--text-tertiary); padding: 8px 12px; background: var(--bg-tertiary); border-radius: 6px; font-family: 'Consolas', monospace; }

.category-filter { display: flex; gap: 8px; flex-wrap: wrap; }
.cat-btn {
  padding: 6px 14px; background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 16px; color: var(--text-secondary); font-size: 12px; font-weight: 500;
  cursor: pointer; transition: var(--transition); font-family: inherit;
}
.cat-btn:hover { border-color: var(--lenovo-red); color: var(--text-primary); }
.cat-btn.active { background: var(--lenovo-red); border-color: var(--lenovo-red); color: white; }

.tools-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 12px; }
.tool-card {
  background: var(--bg-card); border: 1px solid var(--border-color);
  border-radius: 10px; padding: 16px; display: flex; flex-direction: column; gap: 10px;
  transition: var(--transition);
}
.tool-card:hover { border-color: var(--border-light); background: var(--bg-card-hover); }
.tool-card.installed { border-color: rgba(76, 175, 80, 0.3); }

.tool-header { display: flex; align-items: center; gap: 12px; }
.tool-icon {
  width: 40px; height: 40px; border-radius: 10px;
  display: flex; align-items: center; justify-content: center;
  background: var(--bg-tertiary); color: var(--text-secondary);
}
.tool-icon.monitor { background: rgba(33, 150, 243, 0.15); color: #2196F3; }
.tool-icon.system { background: rgba(156, 39, 176, 0.15); color: #9C27B0; }
.tool-icon.benchmark { background: rgba(255, 152, 0, 0.15); color: #FF9800; }
.tool-icon.diagnostic { background: rgba(76, 175, 80, 0.15); color: #4CAF50; }

.tool-info { flex: 1; display: flex; flex-direction: column; gap: 2px; }
.tool-name { font-size: 14px; font-weight: 600; color: var(--text-primary); }
.tool-vendor { font-size: 11px; color: var(--text-tertiary); }

.installed-badge {
  display: flex; align-items: center; gap: 4px;
  padding: 3px 8px; background: rgba(76, 175, 80, 0.15);
  border-radius: 10px; color: #4CAF50; font-size: 11px; font-weight: 600;
}

.tool-desc { font-size: 12px; color: var(--text-secondary); line-height: 1.5; }
.tool-meta { display: flex; gap: 12px; font-size: 11px; color: var(--text-tertiary); }
.tool-version { font-family: 'Consolas', monospace; }
.tool-winget { background: rgba(33, 150, 243, 0.15); color: #2196F3; padding: 2px 6px; border-radius: 4px; }

.tool-installing { display: flex; align-items: center; gap: 8px; padding: 8px 10px; background: rgba(255, 152, 0, 0.08); border-radius: 6px; color: #FF9800; font-size: 12px; }

.tool-error, .tool-success { display: flex; align-items: center; gap: 6px; padding: 8px 10px; border-radius: 6px; font-size: 11px; }
.tool-error { background: rgba(244, 67, 54, 0.08); color: #F44336; border: 1px solid rgba(244, 67, 54, 0.2); }
.tool-success { background: rgba(76, 175, 80, 0.08); color: #4CAF50; border: 1px solid rgba(76, 175, 80, 0.2); }

.tool-actions { display: flex; gap: 8px; margin-top: 4px; }
.btn-install, .btn-installing {
  flex: 1; padding: 8px 12px; border: none; border-radius: 6px;
  font-size: 12px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
}
.btn-install { background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%); color: white; }
.btn-install:hover { opacity: 0.9; transform: translateY(-1px); }
.btn-installing { background: var(--bg-tertiary); color: var(--text-secondary); cursor: not-allowed; }
.btn-run {
  flex: 1; padding: 8px 12px; border: none; border-radius: 6px;
  background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%);
  color: white; font-size: 12px; font-weight: 600; cursor: pointer;
  display: flex; align-items: center; justify-content: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
}
.btn-run:hover { opacity: 0.9; transform: translateY(-1px); }
.btn-link {
  padding: 8px 12px; background: var(--bg-tertiary); border: 1px solid var(--border-color);
  border-radius: 6px; color: var(--text-secondary); font-size: 12px; font-weight: 500;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit; text-decoration: none;
}
.btn-link:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }

.quick-actions { display: flex; gap: 12px; flex-wrap: wrap; }
.btn-batch {
  padding: 10px 20px; background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border: none; border-radius: 8px; color: white; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 8px;
  transition: var(--transition); font-family: inherit;
}
.btn-batch:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-batch:disabled { opacity: 0.6; cursor: not-allowed; }
</style>
