<template>
  <div class="salog-section">
    <!-- Capture System Event Log -->
    <div class="card event-log-card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
            <line x1="3" y1="9" x2="21" y2="9"/>
            <line x1="9" y1="21" x2="9" y2="9"/>
          </svg>
          Capture System Event Log
        </span>
        <span v-if="eventLogResult" class="capture-badge" style="background: rgba(76,175,80,0.1); color:#4CAF50; border-color:rgba(76,175,80,0.2);">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
          Captured
        </span>
      </div>
      <div class="event-log-body">
        <p class="event-log-hint">Capture Windows System event log (errors, warnings, crashes) for quick diagnostics — no ETL trace needed.</p>
        <div class="event-log-controls">
          <div class="event-log-presets">
            <button v-for="p in eventLogPresets" :key="p.label"
              :class="['preset-btn', { active: selectedEventLogPreset === p.label }]"
              @click="selectedEventLogPreset = p.label">
              {{ p.label }}
            </button>
          </div>
          <button class="btn-capture-eventlog" @click="captureEventLog" :disabled="capturingEventLog">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ spinning: capturingEventLog }">
              <circle v-if="capturingEventLog" cx="12" cy="12" r="10"/>
              <path v-if="capturingEventLog" d="M12 6v6l4 2"/>
              <polyline v-else points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
            {{ capturingEventLog ? 'Capturing...' : 'Capture Event Log' }}
          </button>
          <button class="btn-open-eventviewer" @click="openEventViewer">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
              <line x1="3" y1="9" x2="21" y2="9"/>
              <line x1="9" y1="21" x2="9" y2="9"/>
            </svg>
            Open Event Log
          </button>
          <button v-if="eventLogResult" class="btn-export-eventlog" @click="exportEventLog">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            Export CSV
          </button>
          <button class="btn-load-evtx" @click="loadEVTX" :disabled="loadingEVTX">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="12" y1="18" x2="12" y2="12"/>
              <line x1="9" y1="15" x2="15" y2="15"/>
            </svg>
            {{ loadingEVTX ? 'Loading...' : 'Load EVTX' }}
          </button>
          <button v-if="eventLogResult" class="btn-evtx-to-csv" @click="evtxToCSV" :disabled="!evtxFilePath">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            EVTX → CSV
          </button>
        </div>
        <!-- Level Filter -->
        <div v-if="eventLogResult" class="event-level-filter">
          <span class="filter-label">Filter by Level:</span>
          <button :class="['level-btn', 'level-critical', { active: evtxLevelFilter === 'Critical' }]" @click="evtxLevelFilter = evtxLevelFilter === 'Critical' ? '' : 'Critical'">Critical</button>
          <button :class="['level-btn', 'level-error', { active: evtxLevelFilter === 'Error' }]" @click="evtxLevelFilter = evtxLevelFilter === 'Error' ? '' : 'Error'">Error</button>
          <button :class="['level-btn', 'level-warning', { active: evtxLevelFilter === 'Warning' }]" @click="evtxLevelFilter = evtxLevelFilter === 'Warning' ? '' : 'Warning'">Warning</button>
          <button :class="['level-btn', 'level-info', { active: evtxLevelFilter === 'Information' }]" @click="evtxLevelFilter = evtxLevelFilter === 'Information' ? '' : 'Information'">Info</button>
          <button v-if="evtxLevelFilter" class="level-btn level-clear" @click="evtxLevelFilter = ''">✕ Clear</button>
        </div>
      </div>

      <!-- Event Log Results -->
      <div v-if="eventLogResult" class="event-log-results">
        <!-- Summary Bar -->
        <div class="event-log-summary">
          <div class="event-stat critical">
            <span class="stat-count">{{ eventLogResult.criticalCount }}</span>
            <span class="stat-label">Critical</span>
          </div>
          <div class="event-stat error">
            <span class="stat-count">{{ eventLogResult.errorCount }}</span>
            <span class="stat-label">Errors</span>
          </div>
          <div class="event-stat warning">
            <span class="stat-count">{{ eventLogResult.warningCount }}</span>
            <span class="stat-label">Warnings</span>
          </div>
          <div class="event-stat info">
            <span class="stat-count">{{ eventLogResult.infoCount }}</span>
            <span class="stat-label">Info</span>
          </div>
          <div class="event-stat total">
            <span class="stat-count">{{ eventLogResult.totalEvents }}</span>
            <span class="stat-label">{{ eventLogResult.timeRange }}</span>
          </div>
        </div>

        <!-- Events Table -->
        <div v-if="filteredEvents.length" class="event-error-table">
          <div class="table-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#F44336" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            {{ evtxLevelFilter ? evtxLevelFilter + ' Events' : 'Recent Critical / Errors' }}
            <span class="table-count">{{ filteredEvents.length }}</span>
          </div>
          <div class="table-scroll">
            <table class="event-table">
              <thead>
                <tr>
                  <th>Time</th>
                  <th>Level</th>
                  <th>Source</th>
                  <th>ID</th>
                  <th>Message</th>
                  <th>EventData</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(ev, i) in filteredEvents" :key="i" :class="{ 'row-critical': ev.level === 'Critical' }">
                  <td class="cell-time">{{ formatEventTime(ev.time) }}</td>
                  <td><span :class="'level-badge level-' + ev.level.toLowerCase()">{{ ev.level }}</span></td>
                  <td class="cell-source">{{ ev.providerName }}</td>
                  <td class="cell-id">{{ ev.eventId }}</td>
                  <td class="cell-msg">{{ ev.message }}</td>
                  <td class="cell-eventdata"><span v-if="ev.eventData" style="white-space:pre-line">{{ ev.eventData }}</span><span v-else>-</span></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>

        <!-- Top Providers -->
        <div v-if="eventLogResult.topProviders && eventLogResult.topProviders.length" class="event-providers">
          <div class="table-title">Top Event Sources</div>
          <div class="provider-list">
            <div v-for="(p, i) in eventLogResult.topProviders" :key="i" class="provider-row">
              <span class="provider-name">{{ p.name }}</span>
              <span class="provider-count">{{ p.count }}</span>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Event Log Analysis Card -->
    <div v-if="eventLogResult" class="card eventlog-analysis-card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <path d="M9.663 17h4.673M12 3v1m6.364 1.636l-.707.707M21 12h-1M4 12H3m3.343-5.657l-.707-.707m2.828 9.9a5 5 0 117.072 0l-.548.547A3.374 3.374 0 0014 18.469V19a2 2 0 11-4 0v-.531c0-.895-.356-1.754-.988-2.386l-.548-.547z"/>
          </svg>
          Event Log Analysis
        </span>
        <button class="btn-analyze-eventlog" @click="analyzeEventLog" :disabled="analyzingEventLog">
          <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ spinning: analyzingEventLog }">
            <circle v-if="analyzingEventLog" cx="12" cy="12" r="10"/>
            <path v-if="analyzingEventLog" d="M12 6v6l4 2"/>
            <path v-else d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/>
            <polyline v-if="!analyzingEventLog" points="22 4 12 14.01 9 11.01"/>
          </svg>
          {{ analyzingEventLog ? 'Analyzing...' : 'Analyze' }}
        </button>
      </div>
      <div v-if="eventLogAnalysis" class="eventlog-analysis-body">
        <!-- Health Badge -->
        <div :class="['health-badge', eventLogAnalysis.overallHealth.toLowerCase()]">
          <svg v-if="eventLogAnalysis.overallHealth === 'Good'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#4CAF50" stroke-width="2"><path d="M22 11.08V12a10 10 0 1 1-5.93-9.14"/><polyline points="22 4 12 14.01 9 11.01"/></svg>
          <svg v-else-if="eventLogAnalysis.overallHealth === 'Warning'" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#FF9800" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
          <svg v-else width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#F44336" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="15" y1="9" x2="9" y2="15"/><line x1="9" y1="9" x2="15" y2="15"/></svg>
          <span class="health-label">{{ eventLogAnalysis.overallHealth }}</span>
          <span class="health-summary">{{ eventLogAnalysis.summary }}</span>
        </div>

        <!-- Root Causes -->
        <div v-if="eventLogAnalysis.rootCauses && eventLogAnalysis.rootCauses.length" class="analysis-section">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
            Root Causes
          </div>
          <div class="root-cause-list">
            <div v-for="(rc, i) in eventLogAnalysis.rootCauses" :key="i" :class="['root-cause-item', rc.severity.toLowerCase()]">
              <span class="rc-category">{{ rc.category }}</span>
              <span class="rc-detail">{{ rc.detail }}</span>
              <span :class="['rc-severity', rc.severity.toLowerCase()]">{{ rc.severity }}</span>
            </div>
          </div>
        </div>

        <!-- Repeat Errors -->
        <div v-if="eventLogAnalysis.repeatErrors && eventLogAnalysis.repeatErrors.length" class="analysis-section">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
            Repeat Error Patterns
          </div>
          <div class="repeat-list">
            <div v-for="(rp, i) in eventLogAnalysis.repeatErrors" :key="i" class="repeat-item">
              <div class="repeat-header">
                <code class="repeat-source">{{ rp.providerName }} #{{ rp.eventId }}</code>
                <span class="repeat-count">&times;{{ rp.count }}</span>
                <span class="repeat-time">Last: {{ formatEventTime(rp.lastSeen) }}</span>
              </div>
              <div class="repeat-msg">{{ rp.sampleMsg }}</div>
            </div>
          </div>
        </div>

        <!-- Known Issues -->
        <div v-if="eventLogAnalysis.knownIssues && eventLogAnalysis.knownIssues.length" class="analysis-section">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
            Known Issues Detected
          </div>
          <div class="known-issue-list">
            <div v-for="(ki, i) in eventLogAnalysis.knownIssues" :key="i" class="known-issue-item">
              <div class="ki-header">
                <span class="ki-name">{{ ki.name }}</span>
                <span :class="['ki-confidence', ki.confidence >= 80 ? 'high' : ki.confidence >= 60 ? 'mid' : 'low']">{{ ki.confidence }}%</span>
              </div>
              <div class="ki-detail">{{ ki.detail }}</div>
            </div>
          </div>
        </div>

        <!-- Correlations -->
        <div v-if="eventLogAnalysis.correlations && eventLogAnalysis.correlations.length" class="analysis-section">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71"/></svg>
            Correlations
          </div>
          <div class="correlation-list">
            <div v-for="(c, i) in eventLogAnalysis.correlations" :key="i" class="correlation-item">{{ c.description }}</div>
          </div>
        </div>

        <!-- Timeline -->
        <div v-if="eventLogAnalysis.timeline && eventLogAnalysis.timeline.length" class="analysis-section">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg>
            Error Timeline
          </div>
          <div class="timeline-list">
            <div v-for="(te, i) in eventLogAnalysis.timeline" :key="i" :class="['timeline-item', te.level.toLowerCase()]">
              <span class="tl-time">{{ formatEventTime(te.time) }}</span>
              <span :class="['tl-level', te.level.toLowerCase()]">{{ te.level }}</span>
              <code class="tl-source">{{ te.source }} #{{ te.id }}</code>
              <span class="tl-msg">{{ te.message }}</span>
            </div>
          </div>
        </div>

        <!-- Recommendations -->
        <div v-if="eventLogAnalysis.recommendations && eventLogAnalysis.recommendations.length" class="analysis-section">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
            Recommendations
          </div>
          <ul class="recommendation-list">
            <li v-for="(r, i) in eventLogAnalysis.recommendations" :key="i">{{ r }}</li>
          </ul>
        </div>
      </div>
      <div v-else class="analysis-hint">
        <p>Click <strong>Analyze</strong> to identify error patterns, root causes, and get recommendations.</p>
      </div>
    </div>

    <!-- Capture Dispdiag Log -->
    <div class="card dispdiag-card">
      <div class="card-header">
        <span class="card-title">
          <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:8px">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
          </svg>
          Capture Dispdiag Log
        </span>
        <span v-if="dispdiagResult" class="capture-badge" style="background: rgba(76,175,80,0.1); color:#4CAF50; border-color:rgba(76,175,80,0.2);">
          <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
          Captured
        </span>
      </div>
      <div class="dispdiag-body">
        <p class="dispdiag-hint">Capture Windows display diagnostic data — EDID, link training status, brightness info, and driver details via <code>dispdiag.exe</code>.</p>
        <div class="dispdiag-controls">
          <div class="dispdiag-options">
            <label class="dispdiag-opt">
              <input type="checkbox" v-model="dispdiagDump" /> Dump mode (-d)
            </label>
            <label class="dispdiag-opt">
              Delay: <input type="number" v-model.number="dispdiagDelay" min="0" max="30" class="delay-input" />s
            </label>
          </div>
          <button class="btn-capture-dispdiag" @click="captureDispdiag" :disabled="capturingDispdiag">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" :class="{ spinning: capturingDispdiag }">
              <circle v-if="capturingDispdiag" cx="12" cy="12" r="10"/>
              <path v-if="capturingDispdiag" d="M12 6v6l4 2"/>
              <polyline v-else points="22 12 18 12 15 21 9 3 6 12 2 12"/>
            </svg>
            {{ capturingDispdiag ? 'Running dispdiag...' : 'Run Dispdiag' }}
          </button>
          <button class="btn-open-dispdiag" @click="openDispdiagLog">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
              <polyline points="14 2 14 8 20 8"/>
              <line x1="16" y1="13" x2="8" y2="13"/>
              <line x1="16" y1="17" x2="8" y2="17"/>
              <polyline points="10 9 9 9 8 9"/>
            </svg>
            Open Dispdiag Log
          </button>
        </div>
      </div>

      <!-- Dispdiag Results -->
      <div v-if="dispdiagResult" class="dispdiag-results">
        <div class="dispdiag-info-bar">
          <div class="dispdiag-info-item">
            <span class="info-label">File:</span>
            <code class="info-value">{{ dispdiagResult.fileName }}</code>
          </div>
          <div class="dispdiag-info-item">
            <span class="info-label">Size:</span>
            <span class="info-value">{{ dispdiagResult.outputSize }}</span>
          </div>
          <div class="dispdiag-info-item">
            <span class="info-label">Duration:</span>
            <span class="info-value">{{ dispdiagResult.durationSecs }}s</span>
          </div>
          <div class="dispdiag-info-item">
            <span class="info-label">Pass/Fail:</span>
            <span v-if="!dispdiagResult.errors || !dispdiagResult.errors.length" class="pass-badge">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#4CAF50" stroke-width="3"><polyline points="20 6 9 17 4 12"/></svg>
              PASS
            </span>
            <span v-else class="fail-badge">
              <svg width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="#F44336" stroke-width="3"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>
              FAIL
            </span>
          </div>
          <div class="dispdiag-info-item">
            <span class="info-label">EDID Blocks:</span>
            <span class="info-value" style="color:var(--lenovo-red);font-weight:700">{{ dispdiagResult.edidBlocks }}</span>
          </div>
        </div>

        <div v-if="dispdiagResult.errors && dispdiagResult.errors.length" class="dispdiag-errors">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#F44336" stroke-width="2">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            Key Failures / Errors
          </div>
          <ul class="error-list">
            <li v-for="(err, i) in dispdiagResult.errors.slice(0, 15)" :key="i">{{ err }}</li>
          </ul>
        </div>

        <div v-if="dispdiagResult.warnings && dispdiagResult.warnings.length" class="dispdiag-warnings">
          <div class="sub-title">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#FF9800" stroke-width="2">
              <path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z"/>
              <line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/>
            </svg>
            Warnings
          </div>
          <ul class="warn-list">
            <li v-for="(w, i) in dispdiagResult.warnings.slice(0, 10)" :key="i">{{ w }}</li>
          </ul>
        </div>

        <div v-if="dispdiagResult.summary" class="dispdiag-summary">
          <div class="sub-title">Summary</div>
          <div class="summary-grid">
            <div class="summary-row" v-if="dispdiagResult.summary.driverVersion">
              <span class="grid-key">Driver</span>
              <code class="grid-val">{{ dispdiagResult.summary.driverVersion }}</code>
            </div>
            <div class="summary-row" v-if="dispdiagResult.summary.buildVersion">
              <span class="grid-key">Build</span>
              <code class="grid-val">{{ dispdiagResult.summary.buildVersion }}</code>
            </div>
            <div class="summary-row" v-if="dispdiagResult.summary.datVersion">
              <span class="grid-key">Dat Version</span>
              <code class="grid-val">{{ dispdiagResult.summary.datVersion }}</code>
            </div>
          </div>
        </div>

        <div v-if="dispdiagResult.brightnessInfo && dispdiagResult.brightnessInfo.length" class="dispdiag-brightness">
          <div class="sub-title">Brightness Info</div>
          <div class="brightness-list">
            <code v-for="(b, i) in dispdiagResult.brightnessInfo.slice(0, 10)" :key="i" class="brightness-line">{{ b }}</code>
          </div>
        </div>

        <div v-if="dispdiagResult.fileContent" class="dispdiag-content">
          <div class="sub-title">Raw Output Preview</div>
          <pre class="dispdiag-preview">{{ dispdiagResult.fileContent.substring(0, 2000) }}
{{ dispdiagResult.fileContent.length > 2000 ? '...(truncated)' : '' }}</pre>
        </div>

        <div class="dispdiag-actions">
          <button class="btn-dispdiag-action" @click="openDispdiagFolder">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z"/>
            </svg>
            Open Output Folder
          </button>
          <button v-if="dispdiagResult.outputPath" class="btn-dispdiag-action" @click="exportDispdiagJSON">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
              <path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/>
              <polyline points="7 10 12 15 17 10"/>
              <line x1="12" y1="15" x2="12" y2="3"/>
            </svg>
            Export JSON
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { OpenFolder } from '../../../wailsjs/go/main/App'

export default {
  name: 'SALog',
  props: {
    theme: { type: String, default: 'dark' },
  },
  data() {
    return {
      eventLogPresets: [
        { label: 'Last 1h', hours: 1, maxEvents: 200 },
        { label: 'Last 500', hours: 0, maxEvents: 500 },
        { label: 'Last 1000', hours: 0, maxEvents: 1000 },
      ],
      selectedEventLogPreset: 'Last 1000',
      capturingEventLog: false,
      eventLogResult: null,
      analyzingEventLog: false,
      eventLogAnalysis: null,
      loadingEVTX: false,
      evtxFilePath: '',
      evtxLevelFilter: '',
      dispdiagDump: false,
      dispdiagDelay: 0,
      capturingDispdiag: false,
      dispdiagResult: null,
    }
  },
  methods: {
    async captureEventLog() {
      this.capturingEventLog = true
      this.eventLogResult = null
      try {
        const preset = this.eventLogPresets.find(p => p.label === this.selectedEventLogPreset)
        if (!preset) return
        if (window.go?.main?.App) {
          this.eventLogResult = await window.go.main.App.CaptureSystemEventLog(preset.hours, preset.maxEvents)
        }
      } catch (e) {
        console.error(e)
        this.eventLogResult = { error: e.message || String(e), totalEvents: 0, recentErrors: [] }
      } finally {
        this.capturingEventLog = false
      }
    },
    async exportEventLog() {
      try {
        const preset = this.eventLogPresets.find(p => p.label === this.selectedEventLogPreset)
        if (!preset) return
        if (window.go?.main?.App) {
          await window.go.main.App.ExportSystemEventLog('', preset.hours, preset.maxEvents)
          await OpenFolder('C:\\Users\\Public\\ETL_Traces')
        }
      } catch (e) { console.error(e) }
    },
    formatEventTime(ts) {
      if (!ts) return ''
      const m = ts.match(/(\d{4}-\d{2}-\d{2}\s?T?\d{2}:\d{2}:\d{2})/)
      return m ? m[1].replace('T', ' ') : ts.substring(0, 19)
    },
    openEventViewer() {
      try {
        if (window.go?.main?.App) {
          window.go.main.App.OpenEventViewer()
        }
      } catch (e) { console.error(e) }
    },
    async analyzeEventLog() {
      if (!this.eventLogResult) return
      this.analyzingEventLog = true
      this.eventLogAnalysis = null
      try {
        if (window.go?.main?.App) {
          this.eventLogAnalysis = await window.go.main.App.AnalyzeSystemEventLog(this.eventLogResult)
        }
      } catch (e) {
        this.eventLogAnalysis = { overallHealth: 'Error', summary: 'Analysis failed: ' + (e.message || String(e)), rootCauses: [], repeatErrors: [], knownIssues: [], correlations: [], recommendations: [], timeline: [] }
      } finally {
        this.analyzingEventLog = false
      }
    },
    async loadEVTX() {
      this.loadingEVTX = true
      try {
        if (window.go?.main?.App) {
          const path = await window.go.main.App.OpenEVTXFileDialog()
          if (!path) { this.loadingEVTX = false; return }
          this.evtxFilePath = path
          this.eventLogResult = await window.go.main.App.LoadEVTXFile(path)
          this.eventLogAnalysis = null
          this.evtxLevelFilter = ''
        }
      } catch (e) {
        console.error(e)
        this.eventLogResult = { error: e.message || String(e), totalEvents: 0, recentErrors: [], recentEvents: [] }
      } finally {
        this.loadingEVTX = false
      }
    },
    async evtxToCSV() {
      if (!this.evtxFilePath) return
      try {
        if (window.go?.main?.App) {
          const csvPath = await window.go.main.App.ExportEVTXToCSV(this.evtxFilePath, '')
          if (csvPath && !csvPath.includes('failed')) {
            alert('CSV exported to: ' + csvPath)
          } else if (csvPath) {
            alert(csvPath)
          }
        }
      } catch (e) { console.error(e); alert('Export failed: ' + (e.message || String(e))) }
    },
    async captureDispdiag() {
      this.capturingDispdiag = true
      this.dispdiagResult = null
      try {
        if (window.go?.main?.App) {
          this.dispdiagResult = await window.go.main.App.RunDispdiag('', this.dispdiagDelay, this.dispdiagDump)
        }
      } catch (e) {
        console.error(e)
        this.dispdiagResult = { errors: [e.message || String(e)], summary: {} }
      } finally {
        this.capturingDispdiag = false
      }
    },
    async openDispdiagFolder() {
      try {
        if (window.go?.main?.App) {
          const dir = await window.go.main.App.GetDispdiagOutputDir()
          await OpenFolder(dir)
        }
      } catch (e) { console.error(e) }
    },
    async openDispdiagLog() {
      try {
        if (window.go?.main?.App) {
          const result = await window.go.main.App.OpenDispdiagLog()
          if (result && result.includes('No dispdiag log')) {
            alert(result)
          }
        }
      } catch (e) { console.error(e) }
    },
    async exportDispdiagJSON() {
      try {
        if (window.go?.main?.App) {
          await window.go.main.App.ExportDispdiagResult(this.dispdiagResult, '')
        }
      } catch (e) { console.error(e) }
    },
  },
  computed: {
    filteredEvents() {
      if (!this.eventLogResult || !this.eventLogResult.recentEvents) return []
      let events
      if (!this.evtxLevelFilter) {
        events = this.eventLogResult.recentErrors || []
      } else {
        events = this.eventLogResult.recentEvents.filter(e => e.level === this.evtxLevelFilter)
      }
      const critical = events.filter(e => e.level === 'Critical').sort((a, b) => new Date(b.time) - new Date(a.time))
      const others = events.filter(e => e.level !== 'Critical').sort((a, b) => new Date(b.time) - new Date(a.time))
      return [...critical, ...others]
    }
  }
}
</script>

<style scoped>
.salog-section { display: flex; flex-direction: column; gap: 16px; }

.event-log-card { border: 1px dashed rgba(76,175,80,0.3); }
.event-log-body { padding: 16px 20px; }
.event-log-hint { margin: 0 0 12px 0; font-size: 12px; color: var(--text-tertiary); }
.event-log-controls { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.event-log-presets { display: flex; gap: 4px; }
.preset-btn {
  padding: 4px 10px; background: var(--bg-tertiary); border: 1px solid var(--border-color);
  border-radius: 5px; color: var(--text-secondary); font-size: 12px; font-weight: 500;
  cursor: pointer; transition: var(--transition); font-family: inherit;
}
.preset-btn.active { background: var(--lenovo-red); border-color: var(--lenovo-red); color: white; }
.preset-btn:hover:not(.active) { border-color: var(--lenovo-red); }
.btn-capture-eventlog {
  padding: 8px 16px; border-radius: 8px; color: white; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
  background: linear-gradient(135deg, #4CAF50 0%, #2E7D32 100%);
  border: none;
}
.btn-capture-eventlog:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-capture-eventlog:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-export-eventlog {
  padding: 8px 16px; border: 1px solid var(--border-color); border-radius: 8px;
  background: var(--bg-tertiary); color: var(--text-secondary); font-size: 13px; font-weight: 500;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
}
.btn-export-eventlog:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--border-light); }
.btn-open-eventviewer {
  padding: 8px 16px; border: 1px solid var(--border-color); border-radius: 8px;
  background: var(--bg-tertiary); color: var(--text-secondary); font-size: 13px; font-weight: 500;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
}
.btn-open-eventviewer:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--lenovo-red); }
.btn-load-evtx {
  padding: 8px 16px; border-radius: 8px; color: white; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
  background: linear-gradient(135deg, #FF9800 0%, #F57C00 100%); border: none;
}
.btn-load-evtx:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-load-evtx:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-evtx-to-csv {
  padding: 8px 16px; border-radius: 8px; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
  background: transparent; border: 1px solid var(--border-color); color: var(--text-secondary);
}
.btn-evtx-to-csv:hover:not(:disabled) { border-color: #FF9800; color: #FF9800; background: rgba(255,152,0,0.06); }
.btn-evtx-to-csv:disabled { opacity: 0.4; cursor: not-allowed; }
.event-level-filter {
  display: flex; align-items: center; gap: 8px; padding: 8px 0 0 0; flex-wrap: wrap;
}
.filter-label { font-size: 12px; color: var(--text-tertiary); font-weight: 600; }
.level-btn {
  padding: 4px 12px; border-radius: 6px; font-size: 12px; font-weight: 600;
  cursor: pointer; transition: var(--transition); font-family: inherit;
  border: 1px solid var(--border-color); background: transparent; color: var(--text-secondary);
}
.level-btn.active.level-critical { background: #F44336; color: white; border-color: #F44336; }
.level-btn.active.level-error { background: #FF5722; color: white; border-color: #FF5722; }
.level-btn.active.level-warning { background: #FF9800; color: white; border-color: #FF9800; }
.level-btn.active.level-info { background: #2196F3; color: white; border-color: #2196F3; }
.level-btn.level-clear { background: transparent; border-color: var(--text-tertiary); color: var(--text-tertiary); }
.level-btn.level-clear:hover { border-color: var(--lenovo-red); color: var(--lenovo-red); }
.table-count {
  margin-left: 8px; font-size: 11px; font-weight: 700; background: var(--bg-tertiary);
  padding: 2px 8px; border-radius: 10px; color: var(--lenovo-red);
}

.event-log-results { padding: 0 20px 16px 20px; display: flex; flex-direction: column; gap: 12px; }
.event-log-summary { display: flex; gap: 8px; flex-wrap: wrap; }
.event-stat {
  display: flex; flex-direction: column; align-items: center; gap: 2px;
  padding: 8px 16px; border-radius: 8px; min-width: 80px;
}
.event-stat.critical { background: rgba(244,67,54,0.1); border: 1px solid rgba(244,67,54,0.2); }
.event-stat.error { background: rgba(255,152,0,0.08); border: 1px solid rgba(255,152,0,0.2); }
.event-stat.warning { background: rgba(255,193,7,0.08); border: 1px solid rgba(255,193,7,0.2); }
.event-stat.info { background: rgba(33,150,243,0.08); border: 1px solid rgba(33,150,243,0.2); }
.event-stat.total { background: var(--bg-tertiary); border: 1px solid var(--border-color); }
.stat-count { font-size: 22px; font-weight: 700; font-family: 'Consolas','Monaco',monospace; color: var(--text-primary); }
.event-stat.critical .stat-count { color: #F44336; }
.event-stat.error .stat-count { color: #FF9800; }
.event-stat.warning .stat-count { color: #FFC107; }
.event-stat.info .stat-count { color: #2196F3; }
.stat-label { font-size: 10px; font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; color: var(--text-tertiary); }

.event-error-table { }
.table-title {
  display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 700;
  color: var(--text-secondary); margin-bottom: 8px; text-transform: uppercase; letter-spacing: 0.5px;
}
.table-scroll { overflow-y: auto; border-radius: 6px; border: 1px solid var(--border-color); }
.event-table { width: 100%; border-collapse: collapse; font-size: 11px; }
.event-table th {
  position: sticky; top: 0; background: var(--bg-tertiary);
  padding: 6px 8px; text-align: left; font-size: 10px; font-weight: 700;
  text-transform: uppercase; letter-spacing: 0.5px; color: var(--text-tertiary);
  border-bottom: 1px solid var(--border-color);
}
.event-table td { padding: 5px 8px; border-bottom: 1px solid var(--border-color); color: var(--text-secondary); }
.event-table tr:last-child td { border-bottom: none; }
.event-table tr:hover td { background: var(--bg-card-hover); }
.cell-time { white-space: nowrap; font-family: 'Consolas','Monaco',monospace; font-size: 10px; color: var(--text-tertiary); }
.cell-source { font-family: 'Consolas','Monaco',monospace; font-size: 11px; color: var(--text-primary); white-space: nowrap; }
.cell-id { font-family: 'Consolas','Monaco',monospace; font-size: 11px; color: var(--lenovo-red); font-weight: 600; text-align: center; }
.cell-msg { font-size: 11px; color: var(--text-secondary); max-width: 400px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.cell-eventdata { font-size: 11px; color: var(--text-tertiary); max-width: 300px; white-space: pre-line; word-break: break-word; line-height: 1.5; }
.level-badge {
  padding: 1px 6px; border-radius: 3px; font-size: 9px; font-weight: 700; text-transform: uppercase; white-space: nowrap;
}
.level-critical, .level-严重 { background: rgba(244,67,54,0.15); color: #F44336; }
.level-error, .level-错误 { background: rgba(255,152,0,0.15); color: #FF9800; }
.level-warning, .level-警告 { background: rgba(255,193,7,0.12); color: #FFC107; }
.level-information, .level-信息, .level-verbose { background: rgba(33,150,243,0.1); color: #2196F3; }
.row-critical td { color: #F44336 !important; font-weight: 700 !important; background: rgba(244,67,54,0.06); }
.event-table th, .event-table td { border-right: 1px solid rgba(255,255,255,0.08); }
.event-table th:last-child, .event-table td:last-child { border-right: none; }

.event-providers { }
.provider-list { display: flex; flex-wrap: wrap; gap: 4px; }
.provider-row {
  display: flex; align-items: center; gap: 6px; padding: 4px 10px;
  background: var(--bg-tertiary); border: 1px solid var(--border-color);
  border-radius: 5px; font-size: 11px;
}
.provider-name { color: var(--text-primary); font-family: 'Consolas','Monaco',monospace; }
.provider-count { color: var(--lenovo-red); font-weight: 700; }

.eventlog-analysis-card { border: 1px solid rgba(156,39,176,0.2); background: linear-gradient(135deg, rgba(156,39,176,0.02) 0%, transparent 100%); }
.eventlog-analysis-card .card-header { display: flex; justify-content: space-between; align-items: center; }
.btn-analyze-eventlog { padding: 6px 16px; border-radius: 6px; font-size: 13px; font-weight: 600; background: var(--lenovo-red); color: #fff; border: none; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: all 0.2s; }
.btn-analyze-eventlog:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-analyze-eventlog:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-analyze-eventlog svg { width: 14px; height: 14px; }
.eventlog-analysis-body { padding: 16px 20px; display: flex; flex-direction: column; gap: 14px; }
.analysis-hint { padding: 20px; text-align: center; color: var(--text-tertiary); font-size: 13px; }

.health-badge { display: flex; align-items: center; gap: 12px; padding: 14px 16px; border-radius: 8px; border: 1px solid; }
.health-badge.good { background: rgba(76,175,80,0.08); border-color: rgba(76,175,80,0.2); }
.health-badge.warning { background: rgba(255,152,0,0.08); border-color: rgba(255,152,0,0.2); }
.health-badge.critical { background: rgba(244,67,54,0.08); border-color: rgba(244,67,54,0.2); }
.health-label { font-size: 16px; font-weight: 700; text-transform: uppercase; letter-spacing: 0.5px; }
.health-badge.good .health-label { color: #4CAF50; }
.health-badge.warning .health-label { color: #FF9800; }
.health-badge.critical .health-label { color: #F44336; }
.health-summary { flex: 1; font-size: 13px; color: var(--text-secondary); }

.analysis-section { display: flex; flex-direction: column; gap: 8px; }
.analysis-section .sub-title { display: flex; align-items: center; gap: 8px; font-size: 13px; font-weight: 600; color: var(--text-primary); margin-bottom: 4px; }

.root-cause-list { display: flex; flex-direction: column; gap: 6px; }
.root-cause-item { display: flex; align-items: center; gap: 10px; padding: 10px 12px; border-radius: 6px; border: 1px solid var(--border-color); background: var(--bg-card); }
.rc-category { font-weight: 600; color: var(--text-primary); min-width: 100px; }
.rc-detail { flex: 1; font-size: 12px; color: var(--text-secondary); }
.rc-severity { font-size: 11px; padding: 2px 8px; border-radius: 10px; font-weight: 600; text-transform: uppercase; }
.rc-severity.critical { background: rgba(244,67,54,0.1); color: #F44336; }
.rc-severity.warning { background: rgba(255,152,0,0.1); color: #FF9800; }
.rc-severity.info { background: rgba(33,150,243,0.1); color: #2196F3; }

.repeat-list { display: flex; flex-direction: column; gap: 6px; }
.repeat-item { padding: 10px 12px; border-radius: 6px; border: 1px solid var(--border-color); background: var(--bg-card); }
.repeat-header { display: flex; align-items: center; gap: 8px; margin-bottom: 4px; }
.repeat-source { font-family: 'Consolas','Monaco',monospace; font-size: 12px; color: var(--lenovo-red); }
.repeat-count { background: var(--bg-tertiary); padding: 2px 8px; border-radius: 10px; font-size: 11px; font-weight: 600; color: var(--text-primary); }
.repeat-time { font-size: 11px; color: var(--text-tertiary); margin-left: auto; }
.repeat-msg { font-size: 11px; color: var(--text-secondary); padding-left: 4px; }

.known-issue-list { display: flex; flex-direction: column; gap: 8px; }
.known-issue-item { padding: 12px; border-radius: 6px; border: 1px solid rgba(255,152,0,0.3); background: rgba(255,152,0,0.05); }
.ki-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 6px; }
.ki-name { font-weight: 600; color: var(--text-primary); }
.ki-confidence { font-size: 11px; padding: 2px 10px; border-radius: 10px; font-weight: 600; }
.ki-confidence.high { background: rgba(244,67,54,0.1); color: #F44336; }
.ki-confidence.mid { background: rgba(255,152,0,0.1); color: #FF9800; }
.ki-confidence.low { background: rgba(33,150,243,0.1); color: #2196F3; }
.ki-detail { font-size: 12px; color: var(--text-secondary); line-height: 1.5; }

.correlation-list { display: flex; flex-direction: column; gap: 6px; }
.correlation-item { padding: 8px 12px; border-radius: 6px; border: 1px solid var(--border-color); background: var(--bg-card); font-size: 12px; color: var(--text-secondary); }

.timeline-list { display: flex; flex-direction: column; gap: 4px; max-height: 300px; overflow-y: auto; }
.timeline-item { display: flex; align-items: center; gap: 8px; padding: 6px 10px; border-radius: 4px; font-size: 11px; background: var(--bg-card); border-left: 3px solid; }
.timeline-item.critical { border-left-color: #F44336; background: rgba(244,67,54,0.03); }
.timeline-item.error { border-left-color: #FF9800; background: rgba(255,152,0,0.02); }
.tl-time { color: var(--text-tertiary); min-width: 85px; font-family: 'Consolas','Monaco',monospace; }
.tl-level { font-weight: 600; text-transform: uppercase; min-width: 55px; }
.tl-level.critical { color: #F44336; }
.tl-level.error { color: #FF9800; }
.tl-source { color: var(--lenovo-red); }
.tl-msg { flex: 1; color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.recommendation-list { margin: 0; padding-left: 20px; display: flex; flex-direction: column; gap: 6px; }
.recommendation-list li { font-size: 13px; color: var(--text-secondary); line-height: 1.5; }

.capture-badge {
  display: flex; align-items: center; gap: 6px; padding: 4px 10px;
  border-radius: 12px; font-size: 12px; font-weight: 600;
}

.dispdiag-card { border: 1px dashed rgba(33,150,243,0.3); }
.dispdiag-body { padding: 16px 20px; }
.dispdiag-hint { margin: 0 0 12px 0; font-size: 12px; color: var(--text-tertiary); }
.dispdiag-hint code { background: var(--bg-tertiary); padding: 1px 5px; border-radius: 3px; font-size: 11px; color: var(--lenovo-red); }
.dispdiag-controls { display: flex; align-items: center; gap: 12px; flex-wrap: wrap; }
.dispdiag-options { display: flex; align-items: center; gap: 16px; }
.dispdiag-opt { display: flex; align-items: center; gap: 6px; font-size: 12px; color: var(--text-secondary); cursor: pointer; user-select: none; }
.dispdiag-opt input[type="checkbox"] { accent-color: var(--lenovo-red); cursor: pointer; }
.delay-input { width: 42px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 4px; color: var(--text-primary); padding: 2px 4px; font-size: 12px; font-family: 'Consolas','Monaco',monospace; text-align: center; }
.btn-capture-dispdiag {
  padding: 8px 16px; border-radius: 8px; color: white; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
  background: linear-gradient(135deg, #2196F3 0%, #1565C0 100%); border: none;
}
.btn-capture-dispdiag:hover:not(:disabled) { opacity: 0.9; transform: translateY(-1px); }
.btn-capture-dispdiag:disabled { opacity: 0.5; cursor: not-allowed; }
.btn-open-dispdiag {
  padding: 8px 16px; border-radius: 8px; font-size: 13px; font-weight: 600;
  cursor: pointer; display: flex; align-items: center; gap: 6px;
  transition: var(--transition); font-family: inherit;
  background: transparent; border: 1px solid var(--border-color); color: var(--text-secondary);
}
.btn-open-dispdiag:hover { border-color: #2196F3; color: #2196F3; background: rgba(33,150,243,0.06); transform: translateY(-1px); }

.dispdiag-results { padding: 0 20px 16px 20px; display: flex; flex-direction: column; gap: 12px; }
.dispdiag-info-bar { display: flex; gap: 12px; flex-wrap: wrap; align-items: center; padding: 10px 14px; background: var(--bg-tertiary); border-radius: 8px; border: 1px solid var(--border-color); }
.dispdiag-info-item { display: flex; align-items: center; gap: 4px; font-size: 12px; }
.info-label { color: var(--text-tertiary); font-weight: 600; text-transform: uppercase; letter-spacing: 0.5px; font-size: 10px; }
.info-value { color: var(--text-primary); font-size: 12px; }
.info-value code { font-family: 'Consolas','Monaco',monospace; font-size: 11px; }
.pass-badge { display: flex; align-items: center; gap: 4px; color: #4CAF50; font-weight: 700; font-size: 12px; background: rgba(76,175,80,0.1); padding: 2px 8px; border-radius: 4px; }
.fail-badge { display: flex; align-items: center; gap: 4px; color: #F44336; font-weight: 700; font-size: 12px; background: rgba(244,67,54,0.1); padding: 2px 8px; border-radius: 4px; }

.sub-title { display: flex; align-items: center; gap: 6px; font-size: 12px; font-weight: 700; color: var(--text-secondary); margin-bottom: 6px; text-transform: uppercase; letter-spacing: 0.5px; }
.error-list, .warn-list { margin: 0; padding: 0 0 0 18px; font-size: 11px; color: var(--text-secondary); max-height: 200px; overflow-y: auto; }
.error-list li { color: #F44336; margin-bottom: 2px; font-family: 'Consolas','Monaco',monospace; }
.warn-list li { color: #FF9800; margin-bottom: 2px; font-family: 'Consolas','Monaco',monospace; }

.summary-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(200px, 1fr)); gap: 6px; }
.summary-row { display: flex; align-items: center; gap: 8px; padding: 6px 10px; background: var(--bg-tertiary); border: 1px solid var(--border-color); border-radius: 5px; }
.grid-key { color: var(--text-tertiary); font-size: 10px; text-transform: uppercase; font-weight: 600; letter-spacing: 0.5px; }
.grid-val { color: var(--text-primary); font-family: 'Consolas','Monaco',monospace; font-size: 11px; }

.brightness-list { display: flex; flex-direction: column; gap: 3px; }
.brightness-line { display: block; padding: 4px 8px; background: var(--bg-tertiary); border-radius: 4px; font-size: 11px; color: var(--text-secondary); font-family: 'Consolas','Monaco',monospace; }

.dispdiag-preview { margin: 0; padding: 10px 12px; background: #1a1a2e; border-radius: 6px; border: 1px solid var(--border-color); font-size: 10px; line-height: 1.4; color: #a0a0b0; font-family: 'Consolas','Monaco',monospace; max-height: 250px; overflow: auto; white-space: pre-wrap; word-break: break-all; }

.dispdiag-actions { display: flex; gap: 8px; flex-wrap: wrap; }
.btn-dispdiag-action { padding: 8px 14px; border: 1px solid var(--border-color); border-radius: 8px; background: var(--bg-tertiary); color: var(--text-secondary); font-size: 12px; font-weight: 500; cursor: pointer; display: flex; align-items: center; gap: 6px; transition: var(--transition); font-family: inherit; }
.btn-dispdiag-action:hover { background: var(--bg-card-hover); color: var(--text-primary); border-color: var(--lenovo-red); }

@keyframes spin { to { transform: rotate(360deg); } }
.spinning { animation: spin 0.8s linear infinite; }
</style>
