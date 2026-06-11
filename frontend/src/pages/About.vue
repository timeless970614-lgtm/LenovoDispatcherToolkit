<template>
  <div class="about-page">

    <!-- Hero Card -->
    <div class="card hero-card">
      <div class="hero-content">
        <div class="hero-logo">
          <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="white" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
        </div>
        <div class="hero-text">
          <h2 class="hero-title">Lenovo Dispatcher Toolkit</h2>
          <p class="hero-subtitle">Power Management & Thermal Control Utility</p>
          <div class="hero-meta">
            <span class="version-chip">v1.0.21</span>
            <span class="build-chip">Build 2026.06.07</span>
          </div>
        </div>
      </div>
    </div>

    <!-- Info Grid: Build Info + System Requirements -->
    <div class="grid-2">

      <!-- Build Information -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;">
              <rect x="2" y="3" width="20" height="14" rx="2"/><line x1="8" y1="21" x2="16" y2="21"/><line x1="12" y1="17" x2="12" y2="21"/>
            </svg>
            Build Information
          </span>
        </div>
        <div class="info-list">
          <div class="info-row">
            <span class="info-key">Version</span>
            <span class="info-val">1.0.21</span>
          </div>
          <div class="info-row">
            <span class="info-key">Build Date</span>
            <span class="info-val">June 2026</span>
          </div>
          <div class="info-row">
            <span class="info-key">Framework</span>
            <span class="info-val">Wails v2.12.0</span>
          </div>
          <div class="info-row">
            <span class="info-key">Frontend</span>
            <span class="info-val">Vue 3 + Vite 4</span>
          </div>
          <div class="info-row">
            <span class="info-key">Backend</span>
            <span class="info-val">Go 1.22</span>
          </div>
          <div class="info-row">
            <span class="info-key">Platform</span>
            <span class="info-val">Windows 10 / 11 x64</span>
          </div>
          <div class="info-row">
            <span class="info-key">Developer</span>
            <span class="info-val dev-name">zhoushang2</span>
          </div>
        </div>
      </div>

      <!-- System Requirements -->
      <div class="card">
        <div class="card-header">
          <span class="card-title">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;">
              <circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/>
            </svg>
            System Requirements
          </span>
        </div>
        <div class="req-list">
          <div class="req-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            Windows 10 (21H2) or Windows 11
          </div>
          <div class="req-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            Lenovo laptop with Dispatcher driver installed
          </div>
          <div class="req-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            Administrator privileges (for service & driver operations)
          </div>
          <div class="req-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            Intel DTT (Dynamic Tuning Technology) — optional, for ITS mode display
          </div>
          <div class="req-item">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polyline points="20 6 9 17 4 12"/></svg>
            4 GB RAM minimum, 8 GB recommended
          </div>
        </div>
      </div>
    </div>

    <!-- Description Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;">
            <circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/>
          </svg>
          About This Tool
        </span>
      </div>
      <div class="about-body">
        <p class="about-desc">
          <strong>Lenovo Dispatcher Toolkit</strong> is a comprehensive utility for Lenovo laptops,
          providing deep access to the <strong>Lenovo Process Management Dispatcher</strong> driver.
          It enables fine-grained control over thermal modes, power plans, GPU behavior,
          and AI engine settings through a modern, intuitive interface.
        </p>
        <p class="about-desc">
          Built with <strong>Wails v2</strong> (Go + Vue), the tool combines native Windows performance
          with a responsive web-tech UI. It communicates directly with the Dispatcher driver via
          registry access and the <strong>LenovoDYTC.dll</strong> native library — no WMI overhead,
          no background bloat.
        </p>
        <p class="about-desc">
          Key capabilities include <strong>DYTC thermal mode switching</strong> (BSM / IBSM / AQM / STD /
          APM / IEPM / EPM / DCC / GEEK), <strong>mode pinning</strong> via ODV33 UserScenario,
          <strong>real-time GPU utilization</strong> monitoring, <strong>intel GPU frequency</strong> tuning,
          <strong>NPU management</strong> for Horizon Robotics NPUs, and
          <strong>ETL trace capture & analysis</strong> for deep system diagnostics.
        </p>
        <p class="about-desc">
          The tool also bundles a <strong>one-click toolkit installer</strong> (HWiNFO64, CPU-Z, GPU-Z,
          CrystalDiskInfo, FurMark, Prime95, and more), an <strong>AI Agent</strong> for system Q&A,
          and a <strong>WMI Explorer</strong> for advanced users.
        </p>
      </div>
    </div>

    <!-- Technologies Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" style="margin-right:6px;vertical-align:middle;">
            <polyline points="16 18 22 12 16 6"/><polyline points="8 6 2 12 8 18"/>
          </svg>
          Technologies &amp; Dependencies
        </span>
      </div>
      <div class="tech-tags">
        <span class="tech-tag" v-for="t in technologies" :key="t">{{ t }}</span>
      </div>
    </div>

    <!-- Links & Copyright -->
    <div class="card footer-card">
      <div class="footer-links">
        <a class="footer-link" href="#" @click.prevent="openGitHub">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M9 19c-5 1.5-5-2.5-7-3m14 6v-3.87a3.37 3.37 0 00-.94-2.61c3.14-.35 6.44-1.54 6.44-7A5.44 5.44 0 0020 4.77 5.07 5.07 0 0019.91 1S18.73.65 16 2.48a13.38 13.38 0 00-7 0C6.27.65 5.09 1 3.87 1.65 2.65 1 1.42 2.73.42 4.77 1.44 2.65 1.24 1.61.65 2.61A5.44 5.44 0 003 5.77a13.38 13.38 0 00-7 0C6.27 19.23 9 21.72 9 19z"/>
          </svg>
          GitHub Repository
        </a>
        <a class="footer-link" href="#" @click.prevent="openLenovoSupport">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <circle cx="12" cy="12" r="10"/><path d="M12 16v-4M12 8h.01"/>
          </svg>
          Lenovo Support
        </a>
      </div>
      <div class="copyright">
        <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/><path d="M14.83 14.83a4 4 0 11-5.66 0"/>
        </svg>
        &copy; 2026 Lenovo Group Ltd. All rights reserved.
        <span class="footer-sep">|</span>
        Developed by zhoushang2
      </div>
    </div>

  </div>
</template>

<script>
export default {
  name: 'About',
  data() {
    return {
      technologies: [
        'Wails v2.12.0',
        'Go 1.22',
        'Vue 3',
        'Vite 4',
        'Windows Service Manager',
        'WMI (Windows Management Instrumentation)',
        'LenovoDYTC.dll',
        'Intel Graphics Control API',
        'Horizon Robotics NPU HAL',
        'Winget',
        'WPR / ETL',
        'WPA',
        'golang.org/x/sys/windows',
        'golang.org/x/sys/windows/svc',
      ]
    }
  },
  methods: {
    openGitHub() {
      if (window.go?.main?.App) {
        window.go.main.App.OpenGitHub?.() || this._shellOpen('https://github.com/timeless970614-lgtm/LenovoDispatcherToolkit')
      } else {
        window.open('https://github.com/timeless970614-lgtm/LenovoDispatcherToolkit', '_blank')
      }
    },
    openLenovoSupport() {
      window.open('https://support.lenovo.com', '_blank')
    },
    _shellOpen(url) {
      // Fallback: use Wails runtime if available
      if (window.runtime) {
        window.runtime.BrowserOpenURL(url)
      } else {
        window.open(url, '_blank')
      }
    }
  }
}
</script>

<style scoped>
.about-page {
  display: flex;
  flex-direction: column;
  gap: 20px;
}

/* ── Hero Card ─────────────────────────────── */
.hero-card {
  background: linear-gradient(135deg, #1a0a09 0%, #2d1210 50%, #1a0a09 100%);
  border-color: rgba(230, 63, 50, 0.3);
  overflow: hidden;
  position: relative;
}
.hero-card::before {
  content: '';
  position: absolute;
  top: -50%;
  right: -10%;
  width: 300px;
  height: 300px;
  background: radial-gradient(circle, rgba(230,63,50,0.15) 0%, transparent 70%);
  pointer-events: none;
}
.hero-content {
  display: flex;
  align-items: center;
  gap: 24px;
  position: relative;
  z-index: 1;
}
.hero-logo {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(230,63,50,0.4);
  flex-shrink: 0;
}
.hero-title {
  font-size: 24px;
  font-weight: 700;
  color: #ffffff;
  margin-bottom: 4px;
  letter-spacing: -0.5px;
}
.hero-subtitle {
  font-size: 14px;
  color: rgba(255,255,255,0.65);
  margin-bottom: 12px;
}
.hero-meta {
  display: flex;
  gap: 8px;
  align-items: center;
}
.version-chip {
  display: inline-block;
  padding: 4px 14px;
  background: rgba(230,63,50,0.15);
  border: 1px solid rgba(230,63,50,0.3);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: var(--lenovo-red);
}
.build-chip {
  display: inline-block;
  padding: 4px 14px;
  background: rgba(96,165,250,0.1);
  border: 1px solid rgba(96,165,250,0.25);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: var(--accent-blue);
}

/* ── Grid ─────────────────────────────────── */
.grid-2 {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 20px;
}
@media (max-width: 900px) {
  .grid-2 { grid-template-columns: 1fr; }
}

/* ── Info List (Build Info) ──────────────── */
.info-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}
.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 14px;
  border-radius: var(--radius-sm);
  transition: var(--transition);
}
.info-row:hover {
  background: var(--bg-tertiary);
}
.info-key {
  font-size: 13px;
  color: var(--text-secondary);
}
.info-val {
  font-size: 13px;
  font-weight: 500;
  color: var(--text-primary);
}
.dev-name {
  color: var(--text-primary);
  font-weight: 600;
}

/* ── System Requirements ──────────────────── */
.req-list {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 4px 0;
}
.req-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-secondary);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  transition: var(--transition);
}
.req-item:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}
.req-item svg {
  color: var(--accent-green);
  flex-shrink: 0;
}

/* ── About Body ───────────────────────────── */
.about-body {
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.about-desc {
  font-size: 13.5px;
  color: var(--text-secondary);
  line-height: 1.8;
  margin: 0;
}
.about-desc strong {
  color: var(--text-primary);
}

/* ── Technology Tags ─────────────────────── */
.tech-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}
.tech-tag {
  display: inline-block;
  padding: 5px 12px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: 20px;
  font-size: 12px;
  color: var(--text-secondary);
  font-family: 'Consolas', 'Monaco', monospace;
  transition: var(--transition);
}
.tech-tag:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  background: rgba(230,63,50,0.05);
}

/* ── Footer Card ──────────────────────────── */
.footer-card {
  padding: 20px 24px;
}
.footer-links {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  flex-wrap: wrap;
}
.footer-link {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 8px 14px;
  background: var(--bg-tertiary);
  border: 1px solid var(--border-color);
  border-radius: var(--radius-md);
  font-size: 13px;
  color: var(--text-secondary);
  text-decoration: none;
  transition: var(--transition);
  cursor: pointer;
}
.footer-link:hover {
  border-color: var(--lenovo-red);
  color: var(--lenovo-red);
  background: rgba(230,63,50,0.05);
}
.copyright {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  font-size: 12px;
  color: var(--text-tertiary);
  padding-top: 14px;
  border-top: 1px solid var(--border-color);
}
.footer-sep {
  opacity: 0.4;
}
</style>
