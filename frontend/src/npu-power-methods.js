// NPU Power Management methods for FunctionCheck.vue

export const npuMethods = {
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

  async refreshNPU() {
    this.npuLoading = true
    this.npuDriverError = ''
    this.npuSDKInfo = {}
    this.npuDeviceList = []
    this.npuDeviceCount = 0
    try {
      const report = await window._GetNPUFullReport()
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
        ctcGroupId: d.ctcPhyInfo && d.ctcPhyInfo.groupId !== undefined ? d.ctcPhyInfo.groupId : -1,
        ctcChipId: d.ctcPhyInfo && d.ctcPhyInfo.chipId !== undefined ? d.ctcPhyInfo.chipId : -1,
        dvfsResult: '',
      }))
      if (this.npuDeviceCount > 0) this.startNPUPowerPolling(0)
    } catch(e) {
      const msg = String(e)
      if (msg.includes('not found') || msg.includes('not exported') || msg.includes('dll')) this.npuDriverError = msg
      console.error('NPU refresh error:', e)
    } finally {
      this.npuLoading = false
    }
  },

  async setNPUPowerLimit(devIndex) {
    const pl = this.npuPowerLimit[devIndex]
    if (!pl) return
    try {
      const result = await window._SetNPUPowerLimit(devIndex, pl.maxW || 25, pl.minW || 5)
      this.$set(this.npuPowerLimitResult, devIndex, result)
    } catch(e) {
      this.$set(this.npuPowerLimitResult, devIndex, { Success: false, Message: String(e) })
    }
  },

  async setNPUDVFS(devIndex, mode) {
    this.npuSettingDVFS = devIndex
    const dev = this.npuDeviceList.find(d => d.index === devIndex)
    if (dev) dev.dvfsResult = ''
    try {
      const msg = await window._NPUSetDVFSMode(devIndex, mode)
      if (dev) dev.dvfsResult = msg
      if (!msg.startsWith('Error') && dev) dev.dvfsMode = mode
      await this.refreshNPU()
    } catch(e) {
      if (dev) dev.dvfsResult = 'Error: ' + String(e)
    } finally {
      this.npuSettingDVFS = null
    }
  },

  async setNPUClockLock(devIndex) {
    const lock = this.npuClockLocking[devIndex]
    if (!lock) return
    const max = parseInt(lock.maxMhz) || 1400
    const min = parseInt(lock.minMhz) || 700
    if (max < 700 || max > 1400 || min < 700 || min > 1400 || min > max) {
      this.$set(this.npuPowerResult, devIndex, { Success: false, Message: 'Clock must be 700-1400 MHz and min <= max' })
      return
    }
    this.$set(this.npuClockLocking[devIndex], 'setting', true)
    this.$set(this.npuPowerResult, devIndex, null)
    try {
      const result = await window._SetNPUClockLock(devIndex, max, min)
      this.$set(this.npuPowerResult, devIndex, result)
      this.$set(this.npuClockLocking[devIndex], 'result', result)
      await this.pollNPUPower(devIndex)
    } catch(e) {
      this.$set(this.npuPowerResult, devIndex, { Success: false, Message: String(e) })
    } finally {
      this.$set(this.npuClockLocking[devIndex], 'setting', false)
    }
  },

  async resetNPUDefaults(devIndex) {
    const lock = this.npuClockLocking[devIndex]
    if (lock) this.$set(lock, 'setting', true)
    this.$set(this.npuPowerResult, devIndex, null)
    try {
      const result = await window._ResetNPUDefaults(devIndex)
      this.$set(this.npuPowerResult, devIndex, result)
      if (result.Success && this.npuClockLocking[devIndex]) {
        this.npuClockLocking[devIndex].maxMhz = 1400
        this.npuClockLocking[devIndex].minMhz = 700
      }
      await this.pollNPUPower(devIndex)
      await this.refreshNPU()
    } catch(e) {
      this.$set(this.npuPowerResult, devIndex, { Success: false, Message: String(e) })
    } finally {
      if (lock) this.$set(lock, 'setting', false)
    }
  },

  async pollNPUPower(devIndex) {
    try {
      const status = await window._GetNPUPowerStatus(devIndex)
      this.$set(this.npuPowerStatus, devIndex, status)
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

  async showNPUDebug() {
    this.showNPUDebug = !this.showNPUDebug
    if (this.showNPUDebug && !this.npuRawProbe) {
      try {
        this.npuRawProbe = await window._NPURawProbe()
      } catch(e) {
        this.npuRawProbe = 'Error: ' + String(e)
      }
    }
  },

  async startNpuScheduler() {
    this.npuSchedStarting = true
    try {
      await window._StartNPUScheduler(this.npuSchedDev, {
        utilHighPct: this.npuSchedSettings.utilHighPct || 85,
        utilLowPct: this.npuSchedSettings.utilLowPct || 20,
        tempWarnC: this.npuSchedSettings.tempWarnC || 80,
        tempCritC: this.npuSchedSettings.tempCritC || 90,
        checkSec: this.npuSchedSettings.checkSec || 5,
      })
      this.npuSchedRunning = true
      this.npuSchedPoll()
    } catch(e) {
      alert('Scheduler start failed: ' + String(e))
    } finally {
      this.npuSchedStarting = false
    }
  },

  async stopNpuScheduler() {
    try {
      await window._StopNPUScheduler()
      this.npuSchedRunning = false
    } catch(e) {
      alert('Scheduler stop failed: ' + String(e))
    }
  },

  async npuSchedPoll() {
    if (!this.npuSchedRunning) return
    try {
      this.npuSchedState = await window._GetNPUSchedulerState()
    } catch(e) { /* silent */ }
    // Keep polling
    setTimeout(() => this.npuSchedPoll(), 3000)
  },
}
