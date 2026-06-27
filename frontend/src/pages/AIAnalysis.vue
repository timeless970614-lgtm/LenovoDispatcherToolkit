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
      <button :class="['ai-tab', { active: activeAIType === 'etl' }]" @click="activeAIType = 'etl'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>
        </svg>
        ETL Trace
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'salog' }]" @click="activeAIType = 'salog'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="3" y="3" width="18" height="18" rx="2" ry="2"/>
          <line x1="3" y1="9" x2="21" y2="9"/>
          <line x1="9" y1="21" x2="9" y2="9"/>
        </svg>
        SA Log
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'toolkit' }]" @click="activeAIType = 'toolkit'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/>
        </svg>
        Toolkit
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'msr' }]" @click="activeAIType = 'msr'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M4 19.5A2.5 2.5 0 0 1 6.5 17H20"/><path d="M6.5 2H20v20H6.5A2.5 2.5 0 0 1 4 19.5v-15A2.5 2.5 0 0 1 6.5 2z"/><line x1="8" y1="7" x2="16" y2="7"/><line x1="8" y1="11" x2="16" y2="11"/><line x1="8" y1="15" x2="12" y2="15"/>
        </svg>
        MSR
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'wmi' }]" @click="activeAIType = 'wmi'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
          <line x1="8" y1="21" x2="16" y2="21"/>
          <line x1="12" y1="17" x2="12" y2="21"/>
        </svg>
        WMI
      </button>
      <button :class="['ai-tab', { active: activeAIType === 'ppm' }]" @click="activeAIType = 'ppm'">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <path d="M13 2L3 14h9l-1 8 10-12h-9l1-8z"/>
        </svg>
        PPM Driver
      </button>
    </div>

    <!-- Tab Content -->
    <LogAnalysis v-if="activeAIType === 'log'" :theme="theme" />
    <ETLTrace v-if="activeAIType === 'etl'" :theme="theme" />
    <SALog v-if="activeAIType === 'salog'" :theme="theme" />
    <ToolkitTab v-if="activeAIType === 'toolkit'" :theme="theme" />
    <MSR v-if="activeAIType === 'msr'" :theme="theme" />
    <WMI v-if="activeAIType === 'wmi'" :theme="theme" />
    <PPMDriver v-if="activeAIType === 'ppm'" :theme="theme" />

  </div>
</template>

<script>
import LogAnalysis from './ai-components/LogAnalysis.vue'
import ETLTrace from './ai-components/ETLTrace.vue'
import SALog from './ai-components/SALog.vue'
import ToolkitTab from './ai-components/ToolkitTab.vue'
import MSR from './MSR.vue'
import WMI from './WMI.vue'
import PPMDriver from './PPMDriver.vue'

export default {
  name: 'AIAnalysis',
  props: {
    theme: { type: String, default: 'dark' },
  },
  components: {
    LogAnalysis,
    ETLTrace,
    SALog,
    ToolkitTab,
    MSR,
    WMI,
    PPMDriver,
  },
  data() {
    return {
      activeAIType: 'log',
    }
  },
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

@keyframes spin { to { transform: rotate(360deg); } }
.spinning { animation: spin 0.8s linear infinite; }
</style>
