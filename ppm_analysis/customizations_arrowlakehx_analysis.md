# Task Artifact: PPM Provisioning customizations.xml 深度分析 (ArrowLakeHX)

## 任务目标
解压 PPM-ARL-v1007.20250118.ppkg，分析 customizations.xml 中 Arrow Lake HX 专用电源策略配置，分析完成后清理临时文件。

## 文件元数据
- **包名**: Power.Settings.Processor.Intel
- **版本**: 1007.20250118
- **OwnerType**: SiliconVendor (Intel)
- **目标**: Family 6 Model 198 (Arrow Lake HX)
- **文件大小**: 490KB (customizations.xml)
- **覆盖架构**: 从 Clovertrail 到 PantherLake 共 48 种 CPU Target

## Arrow Lake HX (MobileArrowLakeHX) 专属配置分析

### Target 匹配条件
```
ProcessorType = Family 6 Model 198 (即 Arrow Lake HX)
ProcessorVendor = GenuineIntel
PowerPlatformRole = 2 (Mobile) 或 8 (Slate)
```

### Hetero 阈值定义 (Definitions)
| 参数 | 值 (十进制) | 解读 |
|------|------------|------|
| HeteroDecreaseThreshold | 1347440720 | P-Core→E-Core 降级阈值 (编码值) |
| HeteroIncreaseThreshold | 2139259522 | E-Core→P-Core 升级阈值 (编码值) |

---

### 🟢 Balanced (平衡模式)

#### Default Profile
| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| CPMinCores | 50 | 10 | 最少活跃核心数 (50%=12核 / 10%=2-3核) |
| CPConcurrency | 95 | 95 | 核心驻留并发阈值 |
| CPDecreaseTime | 3 | 3 | 核心驻留减少延迟(秒) |
| CPIncreaseTime | 1 | 1 | 核心驻留增加延迟(秒) |
| CPHeadroom | 50 | 50 | 核心驻留余量 |
| CPDistribution | 90 | 90 | 核心驻留分布 |
| HeteroIncreaseThreshold | 254 | 254 | P-Core激活阈值 (最大值=始终激活) |
| HeteroDecreaseThreshold | 254 | 254 | P-Core降级阈值 (最大值=不易降级) |
| HeteroPolicy | 0 | 0 | 异构策略 0=全核可用 |
| HeteroClass1InitialPerf | 100 | 100 | E-Core 初始性能 100% |
| PerfEnergyPreference | 25 | 50 | EPP: AC=偏性能, DC=均衡 |
| PerfEnergyPreference1 | 25 | 50 | E-Core EPP (同上) |
| PerfIncreasePolicy | Rocket | - | 频率提升策略=火箭式(激进) |
| PerfIncreaseThreshold | 30 | - | 频率提升触发阈值 |
| PerfDecreaseThreshold | 10 | - | 频率降低触发阈值 |
| PerfLatencyHint | 100 | 100 | 延迟敏感度提示 |
| SoftParkLatency | 10 | 1000 | 软驻留延迟(ms): AC快速唤醒, DC慢唤醒 |
| ModuleUnparkPolicy | Sequential | Sequential | 模块唤醒=顺序策略 |
| IdleDemoteThreshold | 40 | 40 | 空闲降级阈值 |
| IdlePromoteThreshold | 60 | 60 | 空闲升级阈值 |

#### Multimedia Profile
| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| MaxFrequency | 0 (不限) | 1500 MHz | P-Core 频率上限 |
| MaxFrequency1 | 0 (不限) | 1500 MHz | E-Core 频率上限 |
| PerfEnergyPreference | 100 | 100 | EPP=最省电 |
| SchedulingPolicy | Prefer E | Efficient E | 调度策略=偏好E-Core |
| ShortSchedulingPolicy | Prefer E | Efficient E | 短期调度同上 |
| LatencyHintEpp | 19 | 19 | 延迟提示EPP=接近最性能 |

#### LowLatency Profile
| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| PerfEnergyPreference | 10 | 10 | EPP=接近最性能 |

#### Eco / Utility Profile
| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| MaxFrequency | 0 | 1500 MHz | P-Core 上限 |
| MaxFrequency1 | 0 | 1500 MHz | E-Core 上限 |
| PerfEnergyPreference | 50 | 70 | EPP=省电方向 |
| SchedulingPolicy | Prefer E | Efficient E | 优先E-Core |
| ResourcePriority | 50 | 50 | 资源优先级 |

#### Background Profile
| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| MaxFrequency | 0 | 1500 MHz | 后台任务限制 |
| MaxFrequency1 | 0 | 1500 MHz | E-Core 限制 |
| PerfEnergyPreference | 50 | 70 | 后台=省电 |
| SchedulingPolicy | Prefer E | - | 后台优先E-Core |

#### Constrained Profile (受限/低功耗)
| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| CPMinCores | 10 | 4 | 极少核心保持活跃 |
| HeteroIncreaseThreshold | 254 | 0 | DC不激活P-Core |
| HeteroDecreaseThreshold | 254 | 0 | DC不降级(因无P-Core) |
| HeteroDecreaseTime | - | 1s | 降级延迟 |
| HeteroIncreaseTime | - | 6s | 升级延迟 |
| HeteroPolicy | 0 | 4 | DC=仅E-Core策略 |
| PerfEnergyPreference | 60 | 72 | EPP=省电 |
| SoftParkLatency | 1000 | 5000 | 极慢唤醒 |

---

### 🔴 HighPerformance (高性能模式)

| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| CPMinCores | 100 | 100 | 全部核心活跃 |
| HeteroPolicy | 0 | 0 | 全核可用 |
| HeteroIncreaseThreshold | 254 | 254 | 始终激活P-Core |
| PerfEnergyPreference | 0 | 0 | EPP=最性能 |
| PerfEnergyPreference1 | 0 | 0 | E-Core也最性能 |
| PerfLatencyHint | 100 | 100 | 延迟敏感度 |
| SoftParkLatency | 1000 | 1000 | 中等唤醒速度 |

---

### 🟡 BetterBatteryLifeOverlay (节电覆盖层)

| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| CPMinCores | 10 | 4 | 极少核心 |
| HeteroIncreaseThreshold | 254 | 0 | DC不激活P-Core |
| HeteroPolicy | 0 | 4 | DC=仅E-Core |
| MaxFrequency (Multimedia) | 1500 | 1500 | P-Core限1.5GHz |
| MaxFrequency1 (Multimedia) | 1500 | 1500 | E-Core限1.5GHz |
| PerfEnergyPreference (Default) | 60 | 72 | EPP=省电 |
| PerfEnergyPreference (Eco) | 70 | 77 | 更省电 |

---

### 🟠 MaxPerformanceOverlay (最佳性能覆盖层)

| 参数 | AC | DC | 说明 |
|------|-----|-----|------|
| CPMinCores | 50 | 10 | AC半核, DC少核 |
| HeteroPolicy | 0 | 0 | 全核可用 |
| PerfEnergyPreference (Default) | 25 | 33 | AC偏性能, DC略偏性能 |
| PerfEnergyPreference (Background) | 50 | 50 | 后台均衡 |
| PerfIncreasePolicy | - | - | 继承Rocket策略 |
| MaxFrequency (Background) | 0 | 0 | 不限频 |
| SoftParkLatency (Default) | 10 | 10 | 快速唤醒 |

---

## 关键发现总结

### 1. DC 模式 P-Core 频率硬限制
- **平衡/节电/多媒体/后台**: P-Core 和 E-Core 均限 1500 MHz
- **高性能**: 不限频 (EPP=0 即最大性能)
- **最佳性能覆盖**: 不限频

### 2. 异构调度策略分级
| 场景 | HeteroPolicy | 行为 |
|------|-------------|------|
| 高性能/平衡(AC) | 0 | 全核可用，P+E 同时工作 |
| 受限(DC) | 4 | 仅 E-Core，P-Core 休眠 |
| 节电(DC) | 4 | 仅 E-Core |

### 3. EPP 策略矩阵
| Profile | AC | DC | 解读 |
|---------|-----|-----|------|
| HighPerformance | 0 | 0 | 最大性能 |
| MaxPerformance Default | 25 | 33 | 偏性能 |
| Balanced Default | 25 | 50 | AC偏性能/DC均衡 |
| Eco/Utility | 50 | 70 | 均衡→省电 |
| Constrained | 60 | 72 | 省电 |
| Multimedia | 100 | 100 | 最省电(多媒体用E-Core) |
| LowLatency | 10 | 10 | 接近最性能(低延迟) |

### 4. 核心驻留 (Core Parking) 策略
- AC 平衡模式: 50% 核心保持活跃 (约12核)
- DC 平衡模式: 10% 核心活跃 (约2-3核)
- DC 受限/节电: 4% 核心 (约1核)
- 高性能: 100% 全核活跃

### 5. ModuleUnparkPolicy
- Arrow Lake HX 使用 **Module sequential policy** (顺序唤醒)
- 不同于某些架构的 round robin

## 清理
- ✅ C:\PPM-ARL-v1007.20250118 目录已删除
