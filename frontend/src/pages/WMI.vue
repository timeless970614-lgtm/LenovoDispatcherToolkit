<template>
  <div class="wmi-page">

    <!-- Header -->
    <div class="wmi-header">
      <div class="wmi-header-left">
        <div class="wmi-logo">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <rect x="2" y="3" width="20" height="14" rx="2" ry="2"/>
            <line x1="8" y1="21" x2="16" y2="21"/>
            <line x1="12" y1="17" x2="12" y2="21"/>
          </svg>
        </div>
        <div class="wmi-header-text">
          <h3 class="wmi-title">WMI Explorer</h3>
          <span class="wmi-subtitle">Based on WMICodeCreator</span>
        </div>
      </div>
      <button class="btn-icon" @click="loadNamespaces" :disabled="nsLoading" :title="nsLoading ? 'Loading...' : 'Refresh'">
        <span :class="{ spinning: nsLoading }">🔄</span>
      </button>
    </div>

    <!-- Error Banner -->
    <Transition name="slide-down">
      <div v-if="error" class="wmi-error">
        <span class="error-icon">⚠️</span>
        <span>{{ error }}</span>
        <button class="btn-dismiss" @click="error=''">✕</button>
      </div>
    </Transition>

    <!-- Tab Bar -->
    <div class="wmi-tabs">
      <button v-for="tab in tabs" :key="tab.key"
              :class="['wmi-tab', { active: activeTab === tab.key }]"
              @click="activeTab = tab.key">
        <span class="tab-emoji">{{ tab.emoji }}</span>
        <span class="tab-label">{{ tab.label }}</span>
      </button>
    </div>

    <!-- ==================== Query Tab ==================== -->
    <div v-if="activeTab === 'query'" class="wmi-query-layout">

      <!-- Left Panel: Namespace + Classes -->
      <div class="panel-col panel-col-left">
        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">📂 Namespace</span>
            <span class="panel-badge">{{ namespaces.length }}</span>
          </div>
          <div class="panel-list scroll-thin">
            <div v-for="ns in namespaces" :key="ns"
                 :class="['panel-item', { selected: ns === selectedNamespace }]"
                 @click="selectNamespace(ns)">
              <span class="mono">{{ ns }}</span>
            </div>
            <div v-if="namespaces.length === 0 && !nsLoading" class="panel-empty">No namespaces</div>
            <div v-if="nsLoading" class="panel-empty">Loading...</div>
          </div>
        </div>

        <div class="panel-box flex-2">
          <div class="panel-header">
            <span class="panel-title">📋 Classes</span>
            <span class="panel-badge">{{ classesInNs.length }}</span>
          </div>
          <div class="panel-search">
            <input v-model="classSearch" placeholder="Filter classes..." class="input-search" />
          </div>
          <div class="panel-list scroll-thin">
            <div v-for="c in filteredClasses" :key="c.name"
                 :class="['panel-item', { selected: c.name === selectedClass }]"
                 @click="selectClass(c)">
              <span class="mono text-ellipsis">{{ c.name }}</span>
              <span v-if="c.isDynamic" class="tag tag-blue">D</span>
              <span v-if="c.isStatic" class="tag tag-amber">S</span>
            </div>
            <div v-if="classesInNs.length === 0 && !classLoading" class="panel-empty">Select a namespace</div>
            <div v-if="classLoading" class="panel-empty">Loading classes...</div>
          </div>
        </div>
      </div>

      <!-- Center Panel: Properties + Options -->
      <div class="panel-col panel-col-center">
        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">🔧 Properties</span>
            <span class="panel-badge">{{ classDetails.properties.length }}</span>
            <div class="panel-header-actions">
              <button class="btn-tiny" @click="toggleAllProps(true)">All</button>
              <button class="btn-tiny" @click="toggleAllProps(false)">None</button>
            </div>
          </div>
          <div class="panel-list scroll-thin">
            <div v-for="p in classDetails.properties" :key="p.name"
                 :class="['prop-item', { checked: selectedProps.includes(p.name) }]"
                 @click="toggleProp(p.name)">
              <span class="prop-check">{{ selectedProps.includes(p.name) ? '✓' : '' }}</span>
              <span class="prop-name mono">{{ p.name }}</span>
              <span class="prop-type">{{ p.type }}{{ p.isArray ? '[]' : '' }}</span>
              <span v-if="p.isKey" class="tag tag-amber">key</span>
            </div>
            <div v-if="classDetails.properties.length === 0 && !detailsLoading" class="panel-empty">Select a class</div>
            <div v-if="detailsLoading" class="panel-empty">Loading...</div>
          </div>
        </div>

        <div class="panel-box">
          <div class="panel-header"><span class="panel-title">🔍 WHERE</span></div>
          <input v-model="whereCondition" placeholder='e.g. Name="MyDevice"' class="input-code" />
        </div>

        <div class="panel-box">
          <div class="panel-header"><span class="panel-title">🌐 Language</span></div>
          <div class="lang-grid">
            <button v-for="lang in languages" :key="lang"
                    :class="['btn-lang', { active: lang === selectedLang }]"
                    @click="selectedLang = lang">
              {{ lang }}
            </button>
          </div>
        </div>

        <button class="btn-primary" @click="generateQueryCode" :disabled="!selectedClass || genLoading">
          {{ genLoading ? 'Generating...' : '⚡ Generate Code' }}
        </button>
      </div>

      <!-- Right Panel: Generated Code -->
      <div class="panel-col panel-col-right">
        <div class="panel-box code-box flex-1">
          <div class="panel-header">
            <span class="panel-title">📝 Output ({{ selectedLang }})</span>
            <button class="btn-tiny" @click="copyCode">{{ copied ? '✓ Copied' : '📋 Copy' }}</button>
          </div>
          <pre v-if="generatedCode" class="code-block">{{ generatedCode }}</pre>
          <div v-else class="code-placeholder">
            <span class="placeholder-icon">💻</span>
            <span>Select a class and properties, then generate code</span>
          </div>
        </div>
      </div>
    </div>

    <!-- ==================== Method Tab ==================== -->
    <div v-if="activeTab === 'method'" class="wmi-method-layout">

      <!-- Left: Namespace > Class > Method -->
      <div class="panel-col panel-col-left-sm">
        <div class="panel-box">
          <div class="panel-header"><span class="panel-title">📂 Namespace</span></div>
          <div class="panel-list scroll-thin" style="max-height: 140px;">
            <div v-for="ns in namespaces" :key="'m_' + ns"
                 :class="['panel-item', { selected: ns === method_ns }]"
                 @click="method_ns = ns; methodLoadClasses()">
              <span class="mono">{{ ns }}</span>
            </div>
          </div>
        </div>

        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">📋 Methods</span>
            <span class="panel-badge">{{ methodClasses.length }}</span>
          </div>
          <div class="panel-list scroll-thin" style="max-height: 200px;">
            <div v-for="c in methodClasses" :key="'mc_' + c.name"
                 :class="['panel-item', { selected: c.name === method_class }]"
                 @click="methodSelectClass(c)">
              <span class="mono">{{ c.name }}</span>
            </div>
            <div v-if="methodClasses.length === 0 && !methodClassLoading" class="panel-empty">Select namespace</div>
          </div>
        </div>

        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">⚡ Methods</span>
            <span class="panel-badge">{{ methodDetails.methods.length }}</span>
          </div>
          <div class="panel-list scroll-thin" style="max-height: 200px;">
            <div v-for="m in methodDetails.methods" :key="'mm_' + m.name"
                 :class="['method-item', { selected: method_selected && m.name === method_selected.name }]"
                 @click="methodSelectMethod(m)">
              <div class="mono method-name">{{ m.name }}</div>
              <div class="method-sig mono">{{ m.returnType }} {{ m.name }}({{ methodParamSig(m) }})</div>
            </div>
            <div v-if="methodDetails.methods.length === 0" class="panel-empty">Select a class</div>
          </div>
        </div>
      </div>

      <!-- Right: Details + Invoke + Code -->
      <div class="panel-col panel-col-right">
        <div v-if="method_selected" class="method-detail-card">
          <div class="method-detail-header">
            <span class="mono method-detail-name">{{ method_selected.name }}</span>
            <span class="tag tag-muted">{{ method_selected.returnType }}</span>
            <span v-if="method_selected.isStatic" class="tag tag-amber">static</span>
          </div>
          <div v-if="method_selected.description" class="method-detail-desc">{{ method_selected.description }}</div>

          <div v-if="method_selected.parameters && method_selected.parameters.filter(p => p.isIn).length > 0" class="method-params">
            <div class="section-label">Parameters</div>
            <div v-for="p in method_selected.parameters.filter(p => p.isIn)" :key="p.name" class="param-row">
              <span class="param-name mono">{{ p.name }}</span>
              <span class="param-type">{{ p.type }}{{ p.isArray ? '[]' : '' }}</span>
              <input v-model="methodParams[p.name]" :placeholder="p.type" class="input-param" />
            </div>
          </div>
          <div v-else class="section-label muted">No input parameters</div>

          <div class="method-actions">
            <button class="btn-primary" @click="invokeMethod" :disabled="methodInvoking">
              {{ methodInvoking ? 'Invoking...' : '▶ Execute' }}
            </button>
            <button class="btn-secondary" @click="generateMethodCode" :disabled="methodGenLoading">
              {{ methodGenLoading ? 'Generating...' : '⚡ Generate Code' }}
            </button>
          </div>
        </div>
        <div v-else class="panel-box empty-state">
          <span class="empty-icon">⚡</span>
          <span>Select a method from the left panel</span>
        </div>

        <!-- Invoke Result -->
        <Transition name="slide-down">
          <div v-if="methodResult" class="panel-box code-box">
            <div class="panel-header"><span class="panel-title">📤 Invoke Result</span></div>
            <pre class="code-block code-green">{{ methodResult }}</pre>
          </div>
        </Transition>

        <!-- Generated Code -->
        <Transition name="slide-down">
          <div v-if="methodGeneratedCode" class="panel-box code-box flex-1">
            <div class="panel-header">
              <span class="panel-title">📝 Generated Code</span>
              <button class="btn-tiny" @click="copyMethodCode">{{ methodCopied ? '✓ Copied' : '📋 Copy' }}</button>
            </div>
            <pre class="code-block">{{ methodGeneratedCode }}</pre>
          </div>
        </Transition>
      </div>
    </div>

    <!-- ==================== Event Tab ==================== -->
    <div v-if="activeTab === 'event'" class="wmi-event-layout">
      <div class="panel-col panel-col-left-sm">
        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">📡 Event Classes</span>
            <span class="panel-badge">{{ eventClasses.length }}</span>
            <button class="btn-tiny" @click="loadEventClasses" :disabled="eventLoading">↻</button>
          </div>
          <div class="panel-list scroll-thin">
            <div v-for="ec in eventClasses" :key="ec.name"
                 :class="['panel-item', { selected: ec.name === eventSelectedClass }]"
                 @click="selectEventClass(ec)">
              <div class="mono">{{ ec.name }}</div>
              <div class="item-sub">{{ ec.namespace }}</div>
            </div>
            <div v-if="eventClasses.length === 0 && !eventLoading" class="panel-empty">Click ↻ to load</div>
          </div>
        </div>
      </div>

      <div class="panel-col panel-col-right">
        <div v-if="eventSelectedClass" class="panel-box">
          <div class="method-detail-header">
            <span class="mono method-detail-name">{{ eventSelectedClass }}</span>
          </div>
          <div v-if="eventClassDetails.description" class="method-detail-desc">{{ eventClassDetails.description }}</div>

          <div class="section-label">Filter Properties</div>
          <div class="tag-grid">
            <span v-for="p in eventClassDetails.properties" :key="p.name"
                  :class="['tag', eventFilterProps.includes(p.name) ? 'tag-red' : 'tag-muted']"
                  @click="toggleEventProp(p.name)">
              {{ p.name }}
            </span>
          </div>

          <div v-if="eventFilterProps.length > 0" class="mt-2">
            <div class="section-label">WQL Condition</div>
            <input v-model="eventCondition" placeholder='e.g. TargetInstance.Name="test"' class="input-code" />
          </div>

          <div class="method-actions mt-2">
            <button class="btn-primary" @click="generateEventCode" :disabled="eventGenLoading">
              {{ eventGenLoading ? 'Generating...' : '⚡ Generate Event Code' }}
            </button>
          </div>
        </div>
        <div v-else class="panel-box empty-state">
          <span class="empty-icon">📡</span>
          <span>Select an event class</span>
        </div>

        <Transition name="slide-down">
          <div v-if="eventGeneratedCode" class="panel-box code-box flex-1">
            <div class="panel-header">
              <span class="panel-title">📝 Generated Event Code</span>
              <button class="btn-tiny" @click="copyEventCode">{{ eventCopied ? '✓ Copied' : '📋 Copy' }}</button>
            </div>
            <pre class="code-block">{{ eventGeneratedCode }}</pre>
          </div>
        </Transition>
      </div>
    </div>

    <!-- ==================== Browse Tab ==================== -->
    <div v-if="activeTab === 'browse'" class="wmi-browse-layout">
      <div class="panel-col panel-col-left">
        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">📂 Namespaces</span>
            <span class="panel-badge">{{ namespaces.length }}</span>
          </div>
          <div class="panel-list scroll-thin">
            <div v-for="ns in namespaces" :key="'b_' + ns"
                 :class="['panel-item', { selected: ns === browse_ns }]"
                 @click="browse_ns = ns; browseLoadClasses()">
              <span class="mono text-ellipsis">{{ ns }}</span>
            </div>
          </div>
        </div>
      </div>

      <div class="panel-col panel-col-mid">
        <div class="panel-box flex-1">
          <div class="panel-header">
            <span class="panel-title">📋 Classes</span>
            <span class="panel-badge">{{ browseClasses.length }}</span>
          </div>
          <div class="panel-search">
            <input v-model="browseClassSearch" placeholder="Filter..." class="input-search" />
          </div>
          <div class="panel-list scroll-thin">
            <div v-for="c in filteredBrowseClasses" :key="'bc_' + c.name"
                 :class="['panel-item', { selected: c.name === browse_class }]"
                 @click="browseSelectClass(c)">
              <span class="mono text-ellipsis">{{ c.name }}</span>
              <span v-if="c.isDynamic" class="tag tag-blue">D</span>
            </div>
          </div>
        </div>
      </div>

      <div class="panel-col panel-col-right">
        <div v-if="browseDetails.name" class="browse-detail scroll-thin">
          <!-- Class header -->
          <div class="browse-class-header">
            <span class="mono browse-class-name">{{ browseDetails.name }}</span>
            <span v-if="browseDetails.isDynamic" class="tag tag-blue">dynamic</span>
            <span v-if="browseDetails.isStatic" class="tag tag-amber">static</span>
          </div>
          <div v-if="browseDetails.description" class="browse-desc">{{ browseDetails.description }}</div>
          <div v-if="browseDetails.derivation && browseDetails.derivation.length" class="browse-derivation">
            Derivation: {{ browseDetails.derivation.join(' → ') }}
          </div>

          <!-- Properties -->
          <div class="browse-section">
            <div class="section-label">Properties ({{ browseDetails.properties.length }})</div>
            <div class="browse-prop-list">
              <div v-for="p in browseDetails.properties" :key="'bp_' + p.name" class="browse-prop">
                <div class="browse-prop-top">
                  <span class="mono browse-prop-name">{{ p.name }}</span>
                  <span class="browse-prop-type">{{ p.type }}{{ p.isArray ? '[]' : '' }}</span>
                  <span v-if="p.isKey" class="tag tag-amber">key</span>
                </div>
                <div v-if="p.description" class="browse-prop-desc">{{ p.description }}</div>
              </div>
            </div>
          </div>

          <!-- Methods -->
          <div class="browse-section">
            <div class="section-label">Methods ({{ browseDetails.methods.length }})</div>
            <div class="browse-prop-list">
              <div v-for="m in browseDetails.methods" :key="'bm_' + m.name" class="browse-prop">
                <div class="browse-prop-top">
                  <span class="mono browse-prop-name">{{ m.name }}</span>
                  <span class="browse-prop-type">{{ m.returnType }}</span>
                  <span v-if="m.isStatic" class="tag tag-amber">static</span>
                </div>
                <div class="browse-prop-sig mono">{{ m.returnType }} {{ m.name }}({{ methodParamSigBrowse(m) }})</div>
                <div v-if="m.description" class="browse-prop-desc">{{ m.description }}</div>
              </div>
            </div>
          </div>
        </div>
        <div v-else class="panel-box empty-state flex-1">
          <span class="empty-icon">📂</span>
          <span>Select a namespace and class to browse</span>
        </div>
      </div>
    </div>

  </div>
</template>

<script>
export default {
  name: 'WMI',
  props: {
    theme: { type: String, default: 'dark' }
  },
  data() {
    return {
      activeTab: 'query',
      tabs: [
        { key: 'query', label: 'Query', emoji: '🔍' },
        { key: 'method', label: 'Method', emoji: '⚡' },
        { key: 'event', label: 'Event', emoji: '📡' },
        { key: 'browse', label: 'Browse', emoji: '📂' }
      ],
      // Shared
      namespaces: [],
      nsLoading: false,
      error: '',
      // Query tab
      selectedNamespace: '',
      selectedClass: '',
      classSearch: '',
      classesInNs: [],
      classLoading: false,
      classDetails: { properties: [], methods: [] },
      detailsLoading: false,
      selectedProps: [],
      whereCondition: '',
      languages: ['PowerShell', 'VBScript', 'C#', 'VB.NET'],
      selectedLang: 'PowerShell',
      generatedCode: '',
      genLoading: false,
      copied: false,
      // Method tab
      method_ns: '',
      methodClasses: [],
      methodClassLoading: false,
      method_class: '',
      methodDetails: { methods: [] },
      methodDetailsLoading: false,
      method_selected: null,
      methodParams: {},
      methodInvoking: false,
      methodResult: '',
      methodGenLoading: false,
      methodGeneratedCode: '',
      methodCopied: false,
      // Event tab
      eventClasses: [],
      eventLoading: false,
      eventSelectedClass: '',
      eventClassDetails: { properties: [], description: '' },
      eventFilterProps: [],
      eventCondition: '',
      eventGenLoading: false,
      eventGeneratedCode: '',
      eventCopied: false,
      // Browse tab
      browse_ns: '',
      browseClasses: [],
      browseClassSearch: '',
      browse_class: '',
      browseDetails: {},
      browseDetailsLoading: false,
    }
  },
  computed: {
    filteredClasses() {
      if (!this.classSearch.trim()) return this.classesInNs
      const q = this.classSearch.toLowerCase()
      return this.classesInNs.filter(c => c.name.toLowerCase().includes(q))
    },
    filteredBrowseClasses() {
      if (!this.browseClassSearch.trim()) return this.browseClasses
      const q = this.browseClassSearch.toLowerCase()
      return this.browseClasses.filter(c => c.name.toLowerCase().includes(q))
    }
  },
  mounted() {
    this.waitForRuntime()
  },
  methods: {
    // ---- Runtime wait ----
    waitForRuntime(retry) {
      retry = retry || 0
      if (window?.go?.main?.App?.EnumerateAllNamespaces) {
        this.loadNamespaces()
        return
      }
      if (retry > 40) { this.error = 'Wails runtime not ready'; return }
      setTimeout(() => this.waitForRuntime(retry + 1), 500)
    },

    // ---- Namespaces ----
    async loadNamespaces() {
      this.nsLoading = true
      this.error = ''
      try {
        const raw = await window.go.main.App.EnumerateAllNamespaces()
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.namespaces = data.namespaces || []
        const preferred = this.namespaces.find(n => n.toLowerCase().includes('cimv2')) || this.namespaces[0]
        if (preferred) { this.selectedNamespace = preferred; this.method_ns = preferred; this.browse_ns = preferred; this.loadClassesInNs() }
      } catch (e) {
        this.error = 'Failed to load namespaces: ' + (e.message || e)
      } finally { this.nsLoading = false }
    },

    selectNamespace(ns) {
      this.selectedNamespace = ns
      this.selectedClass = ''
      this.classDetails = { properties: [], methods: [] }
      this.selectedProps = []
      this.generatedCode = ''
      this.loadClassesInNs()
    },

    // ---- Classes in namespace ----
    async loadClassesInNs() {
      if (!this.selectedNamespace) return
      this.classLoading = true
      try {
        const raw = await window.go.main.App.GetClassesInNamespace(this.selectedNamespace)
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.classesInNs = data || []
      } catch (e) {
        this.error = 'Failed to load classes: ' + (e.message || e)
      } finally { this.classLoading = false }
    },

    async selectClass(c) {
      this.selectedClass = c.name
      this.selectedProps = []
      this.generatedCode = ''
      this.detailsLoading = true
      try {
        const raw = await window.go.main.App.GetClassDetails(this.selectedNamespace, c.name)
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.classDetails = data
      } catch (e) {
        this.error = 'Failed to load class details: ' + (e.message || e)
      } finally { this.detailsLoading = false }
    },

    toggleProp(name) {
      const i = this.selectedProps.indexOf(name)
      if (i >= 0) this.selectedProps.splice(i, 1)
      else this.selectedProps.push(name)
    },
    toggleAllProps(on) {
      if (on) this.selectedProps = this.classDetails.properties.map(p => p.name)
      else this.selectedProps = []
    },

    // ---- Code generation ----
    async generateQueryCode() {
      this.genLoading = true
      this.generatedCode = ''
      try {
        const req = {
          namespace: this.selectedNamespace,
          className: this.selectedClass,
          properties: this.selectedProps,
          where: this.whereCondition,
          language: this.selectedLang.toLowerCase()
        }
        const raw = await window.go.main.App.GenerateWMIQueryCode(JSON.stringify(req))
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.generatedCode = data.code
      } catch (e) {
        this.error = 'Code generation failed: ' + (e.message || e)
      } finally { this.genLoading = false }
    },

    async copyCode() {
      if (!this.generatedCode) return
      await navigator.clipboard.writeText(this.generatedCode)
      this.copied = true
      setTimeout(() => this.copied = false, 2000)
    },

    // ---- Method tab ----
    async methodLoadClasses() {
      if (!this.method_ns) return
      this.methodClassLoading = true
      try {
        const raw = await window.go.main.App.GetClassesInNamespace(this.method_ns)
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.methodClasses = (data || []).filter(c => c.methodCount > 0 || c.name.toLowerCase().includes('lenovo') || c.name.toLowerCase().includes('lnv'))
        const raw2 = await window.go.main.App.GetWMIExplorer()
        const exData = JSON.parse(raw2)
        if (exData.classes) {
          const found = exData.classes.filter(c => c.namespace === this.method_ns)
          if (found.length) this.methodClasses = found
        }
      } catch (e) {
        this.error = 'Failed to load method classes: ' + (e.message || e)
      } finally { this.methodClassLoading = false }
    },

    async methodSelectClass(c) {
      this.method_class = c.name
      this.method_selected = null
      this.methodDetailsLoading = true
      try {
        const raw = await window.go.main.App.GetClassDetails(this.method_ns, c.name)
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.methodDetails = data
      } catch (e) {
        this.error = 'Failed: ' + (e.message || e)
      } finally { this.methodDetailsLoading = false }
    },

    methodSelectMethod(m) {
      this.method_selected = m
      this.methodParams = {}
      this.methodResult = ''
      this.methodGeneratedCode = ''
    },

    methodParamSig(m) {
      if (!m.parameters) return ''
      return m.parameters.filter(p => p.isIn).map(p => p.type + ' ' + p.name).join(', ')
    },

    async invokeMethod() {
      this.methodInvoking = true
      this.methodResult = ''
      try {
        const paramsStr = Object.entries(this.methodParams).map(([k, v]) => k + '=' + v).join(';')
        const result = await window.go.main.App.InvokeWMI(this.method_ns, this.method_class, this.method_selected.name, paramsStr)
        this.methodResult = result
      } catch (e) {
        this.methodResult = 'Error: ' + (e.message || e)
      } finally { this.methodInvoking = false }
    },

    async generateMethodCode() {
      this.methodGenLoading = true
      this.methodGeneratedCode = ''
      try {
        const req = {
          namespace: this.method_ns,
          className: this.method_class,
          method: this.method_selected.name,
          inParams: this.method_selected.parameters ? this.method_selected.parameters.filter(p => p.isIn) : [],
          language: this.selectedLang === 'PowerShell' ? 'powershell' : 'vbscript'
        }
        const raw = await window.go.main.App.GenerateWMIMethodCode(JSON.stringify(req))
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.methodGeneratedCode = data.code
      } catch (e) {
        this.error = 'Method code gen failed: ' + (e.message || e)
      } finally { this.methodGenLoading = false }
    },

    async copyMethodCode() {
      if (!this.methodGeneratedCode) return
      await navigator.clipboard.writeText(this.methodGeneratedCode)
      this.methodCopied = true
      setTimeout(() => this.methodCopied = false, 2000)
    },

    // ---- Event tab ----
    async loadEventClasses() {
      this.eventLoading = true
      try {
        const raw = await window.go.main.App.GetEventClasses()
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.eventClasses = data || []
      } catch (e) {
        this.error = 'Failed to load event classes: ' + (e.message || e)
      } finally { this.eventLoading = false }
    },

    async selectEventClass(ec) {
      this.eventSelectedClass = ec.name
      this.eventClassDetails = { properties: ec.properties || [], description: ec.description || '' }
      this.eventFilterProps = []
      this.eventGeneratedCode = ''
    },

    toggleEventProp(name) {
      const i = this.eventFilterProps.indexOf(name)
      if (i >= 0) this.eventFilterProps.splice(i, 1)
      else this.eventFilterProps.push(name)
    },

    async generateEventCode() {
      this.eventGenLoading = true
      this.eventGeneratedCode = ''
      try {
        const req = {
          eventClass: this.eventSelectedClass,
          namespace: this.eventClasses.find(e => e.name === this.eventSelectedClass)?.namespace || 'root\\cimv2',
          condition: this.eventCondition,
          language: this.selectedLang === 'PowerShell' ? 'powershell' : 'vbscript'
        }
        const raw = await window.go.main.App.GenerateWMIEventCode(JSON.stringify(req))
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.eventGeneratedCode = data.code
      } catch (e) {
        this.error = 'Event code gen failed: ' + (e.message || e)
      } finally { this.eventGenLoading = false }
    },

    async copyEventCode() {
      if (!this.eventGeneratedCode) return
      await navigator.clipboard.writeText(this.eventGeneratedCode)
      this.eventCopied = true
      setTimeout(() => this.eventCopied = false, 2000)
    },

    // ---- Browse tab ----
    async browseLoadClasses() {
      if (!this.browse_ns) return
      try {
        const raw = await window.go.main.App.GetClassesInNamespace(this.browse_ns)
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.browseClasses = data || []
      } catch (e) {
        this.error = 'Failed to load browse classes: ' + (e.message || e)
      }
    },

    async browseSelectClass(c) {
      this.browse_class = c.name
      this.browseDetailsLoading = true
      try {
        const raw = await window.go.main.App.GetClassDetails(this.browse_ns, c.name)
        const data = JSON.parse(raw)
        if (data.error) { this.error = data.error; return }
        this.browseDetails = data
      } catch (e) {
        this.error = 'Failed: ' + (e.message || e)
      } finally { this.browseDetailsLoading = false }
    },

    methodParamSigBrowse(m) {
      if (!m.parameters) return ''
      return m.parameters.filter(p => p.isIn).map(p => p.type + ' ' + p.name).join(', ')
    }
  }
}
</script>

<style scoped>
/* ========== Layout Variables ========== */
.wmi-page {
  display: flex;
  flex-direction: column;
  gap: 12px;
  color: var(--text-primary, #e0e0e0);
  font-size: 13px;
}

/* ========== Header ========== */
.wmi-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 2px;
}
.wmi-header-left {
  display: flex;
  align-items: center;
  gap: 12px;
}
.wmi-logo {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, var(--lenovo-red, #E63F32), var(--lenovo-red-dark, #c0281e));
  border-radius: 8px;
  color: #fff;
  box-shadow: 0 2px 8px rgba(230, 63, 50, 0.3);
}
.wmi-header-text {
  display: flex;
  flex-direction: column;
}
.wmi-title {
  font-size: 16px;
  font-weight: 700;
  color: var(--lenovo-red, #E63F32);
  line-height: 1.2;
  margin: 0;
}
.wmi-subtitle {
  font-size: 11px;
  color: var(--lenovo-red, #E63F32);
  opacity: 0.7;
}

/* ========== Error ========== */
.wmi-error {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  background: rgba(230, 63, 50, 0.08);
  border: 1px solid rgba(230, 63, 50, 0.25);
  border-radius: 8px;
  color: #fca5a5;
  font-size: 12px;
}
.error-icon { flex-shrink: 0; }
.btn-dismiss {
  margin-left: auto;
  padding: 2px 8px;
  background: transparent;
  border: 1px solid rgba(255,255,255,0.1);
  border-radius: 4px;
  color: #888;
  cursor: pointer;
  font-size: 11px;
}
.btn-dismiss:hover { background: rgba(255,255,255,0.05); }

/* ========== Tabs ========== */
.wmi-tabs {
  display: flex;
  gap: 2px;
  padding: 2px;
  background: var(--bg-card, rgba(255,255,255,0.04));
  border-radius: 10px;
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
}
.wmi-tab {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 9px 16px;
  background: transparent;
  border: none;
  border-radius: 8px;
  color: var(--text-muted, #888);
  cursor: pointer;
  font-size: 13px;
  font-weight: 500;
  transition: all 0.2s ease;
}
.wmi-tab:hover {
  background: rgba(255,255,255,0.04);
  color: var(--text-secondary, #ccc);
}
.wmi-tab.active {
  background: var(--lenovo-red, #E63F32);
  color: #fff;
  box-shadow: 0 2px 8px rgba(230, 63, 50, 0.35);
}
.tab-emoji { font-size: 14px; }
.tab-label { font-size: 13px; }

/* ========== Panel System ========== */
.panel-box {
  background: var(--bg-card, rgba(255,255,255,0.03));
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 10px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}
.panel-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid var(--border-color, rgba(255,255,255,0.06));
  background: rgba(255,255,255,0.02);
}
.panel-title {
  font-size: 11px;
  font-weight: 600;
  color: var(--text-secondary, #aaa);
  text-transform: uppercase;
  letter-spacing: 0.5px;
}
.panel-badge {
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  background: rgba(230, 63, 50, 0.15);
  color: var(--lenovo-red, #E63F32);
  font-weight: 600;
}
.panel-header-actions {
  margin-left: auto;
  display: flex;
  gap: 4px;
}
.panel-list {
  flex: 1;
  overflow-y: auto;
}
.panel-item {
  padding: 7px 12px;
  cursor: pointer;
  font-size: 11px;
  display: flex;
  align-items: center;
  gap: 6px;
  border-left: 2px solid transparent;
  transition: all 0.15s ease;
}
.panel-item:hover {
  background: rgba(255,255,255,0.04);
}
.panel-item.selected {
  background: rgba(230, 63, 50, 0.1);
  border-left-color: var(--lenovo-red, #E63F32);
  color: #fff;
}
.panel-empty {
  padding: 20px 12px;
  text-align: center;
  color: var(--text-muted, #555);
  font-size: 11px;
}
.item-sub {
  font-size: 9px;
  color: var(--text-muted, #555);
}

/* ========== Search Input ========== */
.panel-search {
  padding: 4px 8px;
  border-bottom: 1px solid var(--border-color, rgba(255,255,255,0.06));
}
.input-search {
  width: 100%;
  padding: 6px 10px;
  background: transparent;
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 6px;
  color: var(--text-primary, #e0e0e0);
  font-size: 11px;
  outline: none;
  transition: border-color 0.2s;
}
.input-search:focus {
  border-color: var(--lenovo-red, #E63F32);
}
.input-code {
  width: 100%;
  padding: 8px 10px;
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 6px;
  color: var(--text-primary, #e0e0e0);
  font-size: 11px;
  font-family: 'Consolas', 'Fira Code', monospace;
  outline: none;
  transition: border-color 0.2s;
}
.input-code:focus {
  border-color: var(--lenovo-red, #E63F32);
}

/* ========== Tags ========== */
.tag {
  font-size: 9px;
  padding: 1px 6px;
  border-radius: 4px;
  font-weight: 600;
  letter-spacing: 0.3px;
}
.tag-blue { background: rgba(59,130,246,0.15); color: #60a5fa; }
.tag-amber { background: rgba(245,158,11,0.15); color: #fbbf24; }
.tag-red { background: rgba(230,63,50,0.2); color: var(--lenovo-red, #E63F32); }
.tag-muted { background: rgba(255,255,255,0.05); color: var(--text-muted, #888); cursor: pointer; }
.tag-muted:hover { background: rgba(255,255,255,0.08); }
.tag-grid { display: flex; flex-wrap: wrap; gap: 4px; }

/* ========== Buttons ========== */
.btn-icon {
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--bg-card, rgba(255,255,255,0.04));
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 8px;
  cursor: pointer;
  font-size: 16px;
  transition: all 0.2s;
}
.btn-icon:hover { background: rgba(255,255,255,0.08); border-color: var(--lenovo-red, #E63F32); }
.btn-icon:disabled { opacity: 0.4; cursor: not-allowed; }

.btn-tiny {
  padding: 2px 8px;
  background: transparent;
  border: 1px solid var(--border-color, rgba(255,255,255,0.08));
  border-radius: 4px;
  color: var(--text-muted, #888);
  cursor: pointer;
  font-size: 10px;
  transition: all 0.15s;
}
.btn-tiny:hover { color: var(--lenovo-red, #E63F32); border-color: var(--lenovo-red, #E63F32); }

.btn-primary {
  padding: 10px 20px;
  background: linear-gradient(135deg, var(--lenovo-red, #E63F32), var(--lenovo-red-dark, #c0281e));
  color: #fff;
  border: none;
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  font-weight: 600;
  transition: all 0.2s;
  box-shadow: 0 2px 8px rgba(230, 63, 50, 0.3);
}
.btn-primary:hover { box-shadow: 0 4px 16px rgba(230, 63, 50, 0.45); transform: translateY(-1px); }
.btn-primary:active { transform: translateY(0); }
.btn-primary:disabled { opacity: 0.5; cursor: not-allowed; transform: none; }

.btn-secondary {
  padding: 10px 20px;
  background: var(--bg-card, rgba(255,255,255,0.04));
  color: var(--text-secondary, #ccc);
  border: 1px solid var(--border-color, rgba(255,255,255,0.08));
  border-radius: 8px;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
}
.btn-secondary:hover { border-color: var(--lenovo-red, #E63F32); color: #fff; }
.btn-secondary:disabled { opacity: 0.5; cursor: not-allowed; }

.btn-lang {
  padding: 5px 12px;
  background: transparent;
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 6px;
  color: var(--text-muted, #888);
  cursor: pointer;
  font-size: 11px;
  transition: all 0.2s;
}
.btn-lang:hover { border-color: rgba(255,255,255,0.15); color: var(--text-secondary, #ccc); }
.btn-lang.active {
  background: var(--lenovo-red, #E63F32);
  color: #fff;
  border-color: var(--lenovo-red, #E63F32);
}
.lang-grid { display: flex; flex-wrap: wrap; gap: 6px; }

/* ========== Properties ========== */
.prop-item {
  padding: 5px 10px;
  cursor: pointer;
  display: flex;
  align-items: center;
  gap: 8px;
  transition: background 0.15s;
}
.prop-item:hover { background: rgba(255,255,255,0.03); }
.prop-item.checked { background: rgba(59,130,246,0.08); }
.prop-check {
  width: 16px;
  height: 16px;
  border: 1.5px solid var(--border-color, rgba(255,255,255,0.15));
  border-radius: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 9px;
  color: #60a5fa;
  flex-shrink: 0;
  transition: all 0.15s;
}
.prop-item.checked .prop-check {
  background: rgba(59,130,246,0.2);
  border-color: #3b82f6;
}
.prop-name { color: #a5b4fc; font-size: 11px; }
.prop-type { color: var(--text-muted, #666); font-size: 10px; }

/* ========== Code Block ========== */
.code-box { flex: 1; }
.code-block {
  margin: 0;
  padding: 14px;
  font-size: 11px;
  font-family: 'Consolas', 'Fira Code', monospace;
  color: #c9d1d9;
  white-space: pre-wrap;
  word-break: break-all;
  overflow: auto;
  flex: 1;
  line-height: 1.6;
}
.code-green { color: #00ff88; }
.code-placeholder {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 8px;
  color: var(--text-muted, #444);
  font-size: 12px;
}
.placeholder-icon { font-size: 32px; opacity: 0.3; }

/* ========== Empty State ========== */
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 10px;
  padding: 32px;
  color: var(--text-muted, #555);
  font-size: 12px;
}
.empty-icon { font-size: 36px; opacity: 0.3; }

/* ========== Section Labels ========== */
.section-label {
  font-size: 10px;
  font-weight: 600;
  color: var(--text-muted, #777);
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin: 4px 0 8px 0;
}
.section-label.muted { color: var(--text-muted, #555); margin-top: 12px; }

/* ========== Flex Helpers ========== */
.flex-1 { flex: 1; }
.flex-2 { flex: 2; }
.mt-2 { margin-top: 12px; }
.mono { font-family: 'Consolas', 'Fira Code', monospace; }
.text-ellipsis { white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }

/* ========== Query Layout ========== */
.wmi-query-layout {
  display: grid;
  grid-template-columns: 220px 240px 1fr;
  gap: 10px;
  flex: 1;
  min-height: 0;
}
.panel-col {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
}

/* ========== Method Tab ========== */
.wmi-method-layout {
  display: grid;
  grid-template-columns: 260px 1fr;
  gap: 10px;
  flex: 1;
  min-height: 0;
}
.panel-col-left-sm {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
}

.method-item {
  padding: 6px 10px;
  cursor: pointer;
  border-left: 2px solid transparent;
  transition: all 0.15s;
}
.method-item:hover { background: rgba(255,255,255,0.04); }
.method-item.selected {
  background: rgba(230, 63, 50, 0.08);
  border-left-color: var(--lenovo-red, #E63F32);
}
.method-name { color: #e5e7eb; font-size: 11px; font-weight: 500; }
.method-sig { color: var(--text-muted, #555); font-size: 9px; }

.method-detail-card {
  background: var(--bg-card, rgba(255,255,255,0.03));
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 10px;
  padding: 14px;
}
.method-detail-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}
.method-detail-name { font-size: 15px; font-weight: 700; color: #fff; }
.method-detail-desc { font-size: 11px; color: var(--text-muted, #777); margin-bottom: 10px; }

.method-params { margin-bottom: 10px; }
.param-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.param-name { color: #a5b4fc; font-size: 11px; width: 110px; flex-shrink: 0; }
.param-type { color: var(--text-muted, #666); font-size: 10px; width: 80px; flex-shrink: 0; }
.input-param {
  flex: 1;
  padding: 5px 8px;
  background: rgba(0,0,0,0.2);
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 6px;
  color: var(--text-primary, #e0e0e0);
  font-size: 11px;
  font-family: 'Consolas', monospace;
  outline: none;
}
.input-param:focus { border-color: var(--lenovo-red, #E63F32); }

.method-actions {
  display: flex;
  gap: 8px;
  margin-top: 12px;
}

/* ========== Event Layout ========== */
.wmi-event-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 10px;
  flex: 1;
  min-height: 0;
}

/* ========== Browse Layout ========== */
.wmi-browse-layout {
  display: grid;
  grid-template-columns: 220px 240px 1fr;
  gap: 10px;
  flex: 1;
  min-height: 0;
}
.panel-col-mid {
  display: flex;
  flex-direction: column;
  gap: 8px;
  min-height: 0;
}

.browse-detail {
  background: var(--bg-card, rgba(255,255,255,0.03));
  border: 1px solid var(--border-color, rgba(255,255,255,0.06));
  border-radius: 10px;
  padding: 14px;
  overflow-y: auto;
}
.browse-class-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 6px;
}
.browse-class-name { font-size: 15px; font-weight: 700; color: #fff; }
.browse-desc { font-size: 11px; color: var(--text-muted, #777); margin-bottom: 4px; }
.browse-derivation { font-size: 10px; color: var(--text-muted, #555); margin-bottom: 12px; }
.browse-section { margin-top: 16px; }
.browse-prop-list { display: flex; flex-direction: column; }
.browse-prop {
  padding: 8px 0;
  border-bottom: 1px solid rgba(255,255,255,0.03);
}
.browse-prop:last-child { border-bottom: none; }
.browse-prop-top { display: flex; align-items: center; gap: 8px; }
.browse-prop-name { color: #a5b4fc; font-size: 11px; font-weight: 500; }
.browse-prop-type { color: var(--text-muted, #666); font-size: 10px; }
.browse-prop-desc { font-size: 10px; color: var(--text-muted, #555); margin-top: 2px; }
.browse-prop-sig { font-size: 10px; color: var(--text-muted, #555); }

/* ========== Scrollbar ========== */
.scroll-thin::-webkit-scrollbar { width: 5px; }
.scroll-thin::-webkit-scrollbar-track { background: transparent; }
.scroll-thin::-webkit-scrollbar-thumb { background: rgba(255,255,255,0.1); border-radius: 3px; }
.scroll-thin::-webkit-scrollbar-thumb:hover { background: rgba(255,255,255,0.18); }

/* ========== Transitions ========== */
.slide-down-enter-active,
.slide-down-leave-active {
  transition: all 0.3s ease;
  max-height: 500px;
  overflow: hidden;
}
.slide-down-enter-from,
.slide-down-leave-to {
  opacity: 0;
  max-height: 0;
  margin-top: 0 !important;
}

/* ========== Spinning Animation ========== */
@keyframes spin { to { transform: rotate(360deg); } }
.spinning { display: inline-block; animation: spin 1s linear infinite; }
</style>
