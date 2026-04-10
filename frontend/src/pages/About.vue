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
          <p class="hero-subtitle">Power Management & Dispatcher Control</p>
          <div class="version-chip">v1.0.15</div>
        </div>
      </div>
    </div>

    <!-- Info Cards -->
    <div class="grid-2">
      <div class="card">
        <div class="card-header">
          <span class="card-title">Build Information</span>
        </div>
        <div class="info-list">
          <div class="info-row">
            <span class="info-key">Version</span>
            <span class="info-val">1.0.15</span>
          </div>
          <div class="info-row">
            <span class="info-key">Framework</span>
            <span class="info-val">Wails v2.12.0</span>
          </div>
          <div class="info-row">
            <span class="info-key">Frontend</span>
            <span class="info-val">Vue 3 + Vite</span>
          </div>
          <div class="info-row">
            <span class="info-key">Backend</span>
            <span class="info-val">Go 1.23</span>
          </div>
          <div class="info-row">
            <span class="info-key">Platform</span>
            <span class="info-val">Windows x64</span>
          </div>
          <div class="info-row">
            <span class="info-key">Developer</span>
            <span class="info-val dev-name">zhoushang2</span>
          </div>
        </div>
      </div>

      <div class="card">
        <div class="card-header">
          <span class="card-title">Features</span>
        </div>
        <div class="feature-list">
          <div class="feature-item" v-for="f in features" :key="f">
            <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5">
              <polyline points="20 6 9 17 4 12"/>
            </svg>
            {{ f }}
          </div>
        </div>
      </div>
    </div>

    <!-- Description Card -->
    <div class="card">
      <div class="card-header">
        <span class="card-title">About This Tool</span>
      </div>
      <p class="about-desc">
        Lenovo Dispatcher Toolkit is a lightweight utility for managing Lenovo laptop power and thermal settings.
        It provides direct access to the Lenovo Dispatcher driver, allowing fine-grained control over CPU scheduling,
        energy performance preferences, and thermal modes — all through a modern, intuitive interface.
      </p>
      <div class="copyright">
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
          <circle cx="12" cy="12" r="10"/>
          <path d="M14.83 14.83a4 4 0 1 1 0-5.66"/>
        </svg>
        2023–2026 Lenovo Group LTD. All rights reserved.
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'About',
  data() {
    return {
      features: [
        'Real-time system information via WMI',
        'Dispatcher driver status monitoring',
        'PPM Hetero scheduling control',
        'Energy Performance Preference (EPP)',
        'Max processor frequency tuning',
        'Soft Park Latency configuration',
        'DYTC thermal mode switching',
        'Windows service management'
      ]
    }
  }
}
</script>

<style scoped>
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
  background: radial-gradient(circle, rgba(230, 63, 50, 0.15) 0%, transparent 70%);
  pointer-events: none;
}

.hero-content {
  display: flex;
  align-items: center;
  gap: 24px;
}

.hero-logo {
  width: 80px;
  height: 80px;
  background: linear-gradient(135deg, var(--lenovo-red) 0%, var(--lenovo-red-dark) 100%);
  border-radius: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 8px 24px rgba(230, 63, 50, 0.4);
  flex-shrink: 0;
}

.hero-title {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  margin-bottom: 6px;
  letter-spacing: -0.5px;
}

.hero-subtitle {
  font-size: 14px;
  color: var(--text-secondary);
  margin-bottom: 12px;
}

.version-chip {
  display: inline-block;
  padding: 4px 14px;
  background: rgba(230, 63, 50, 0.15);
  border: 1px solid rgba(230, 63, 50, 0.3);
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
  color: var(--lenovo-red);
}

.info-list {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.info-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
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

.feature-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.feature-item {
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 13px;
  color: var(--text-secondary);
  padding: 8px 12px;
  border-radius: var(--radius-sm);
  transition: var(--transition);
}

.feature-item:hover {
  background: var(--bg-tertiary);
  color: var(--text-primary);
}

.feature-item svg {
  color: var(--lenovo-red);
  flex-shrink: 0;
}

.about-desc {
  font-size: 14px;
  color: var(--text-secondary);
  line-height: 1.8;
  margin-bottom: 20px;
}

.copyright {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 12px;
  color: var(--text-tertiary);
  padding-top: 16px;
  border-top: 1px solid var(--border-color);
}
</style>
