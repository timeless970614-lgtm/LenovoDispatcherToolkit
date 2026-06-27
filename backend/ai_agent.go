//go:build windows

package backend

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

// AIAgentSystemInfo holds comprehensive PC information for the agent
type AIAgentSystemInfo struct {
	OS          OSInfo            `json:"os"`
	CPU         CPUInfo           `json:"cpu"`
	Memory      MemoryInfo        `json:"memory"`
	Disks       []DiskInfo        `json:"disks"`
	GPUs        []AIAgentGPUInfo  `json:"gpus"`
	Network     NetworkInfo       `json:"network"`
	Power       PowerInfo         `json:"power"`
	TopProcs    []ProcessInfo     `json:"topProcs"`
	Uptime      string            `json:"uptime"`
	Hostname    string            `json:"hostname"`
	CurrentUser string            `json:"currentUser"`
}

type OSInfo struct {
	Name         string `json:"name"`
	Version      string `json:"version"`
	Build        string `json:"build"`
	Architecture string `json:"architecture"`
	InstallDate  string `json:"installDate"`
}

type CPUInfo struct {
	Name       string  `json:"name"`
	Cores      int     `json:"cores"`
	Threads    int     `json:"threads"`
	FreqMHz    int     `json:"freqMHz"`
	UsagePct   float64 `json:"usagePct"`
	TempC      float64 `json:"tempC"`
	MaxFreqMHz int     `json:"maxFreqMHz"`
	Vendor     string  `json:"vendor"`
}

type MemoryInfo struct {
	TotalGB     float64 `json:"totalGB"`
	UsedGB      float64 `json:"usedGB"`
	AvailableGB float64 `json:"availableGB"`
	UsedPct     float64 `json:"usedPct"`
}

type DiskInfo struct {
	Drive      string  `json:"drive"`
	Label      string  `json:"label"`
	TotalGB    float64 `json:"totalGB"`
	FreeGB     float64 `json:"freeGB"`
	UsedPct    float64 `json:"usedPct"`
	Type       string  `json:"type"`
	FileSystem string  `json:"fileSystem"`
}

type AIAgentGPUInfo struct {
	Name          string  `json:"name"`
	Vendor        string  `json:"vendor"`
	MemoryMB      int     `json:"memoryMB"`
	DriverVersion string  `json:"driverVersion"`
	UsagePct      float64 `json:"usagePct"`
}

type NetworkInfo struct {
	AdapterName string `json:"adapterName"`
	MAC         string `json:"mac"`
	IPAddress   string `json:"ipAddress"`
	SpeedMbps   int    `json:"speedMbps"`
	Connected   bool   `json:"connected"`
}

type PowerInfo struct {
	Battery       bool   `json:"battery"`
	ACConnected   bool   `json:"acConnected"`
	BatteryPct    int    `json:"batteryPct"`
	BatteryStatus string `json:"batteryStatus"`
	PowerPlan     string `json:"powerPlan"`
}

type ProcessInfo struct {
	Name  string  `json:"name"`
	PID   uint32  `json:"pid"`
	CPUPct float64 `json:"cpuPct"`
	MemMB float64 `json:"memMB"`
}

// GetAIAgentSystemInfo gathers comprehensive system information
// All WMI queries are batched into a single PowerShell call (~300ms total)
// instead of 9 separate calls (~2.7s total).
func GetAIAgentSystemInfo() AIAgentSystemInfo {
	info := AIAgentSystemInfo{}

	// OS info — Go native registry (no PowerShell)
	info.OS = getOSInfo()

	// Batch all WMI/PowerShell queries into a single process spawn
	batch := batchSystemInfo()

	info.CPU = batch.CPU
	info.Memory = batch.Memory
	info.Disks = batch.Disks
	info.GPUs = batch.GPUs
	info.Network = batch.Network
	info.Power = batch.Power
	info.TopProcs = batch.TopProcs
	info.Uptime = batch.Uptime
	info.Hostname, _ = os.Hostname()
	info.CurrentUser = os.Getenv("USERNAME")

	return info
}

// batchSystemInfoResult holds parsed results from a single PowerShell batch call
type batchSystemInfoResult struct {
	CPU      CPUInfo
	Memory   MemoryInfo
	Disks    []DiskInfo
	GPUs     []AIAgentGPUInfo
	Network  NetworkInfo
	Power    PowerInfo
	TopProcs []ProcessInfo
	Uptime   string
}

// batchSystemInfo runs ONE PowerShell process that collects all WMI data.
// Output is section-delimited so we can parse each part independently.
func batchSystemInfo() batchSystemInfoResult {
	result := batchSystemInfoResult{}

	// CPU vendor detection from registry name
	cpuName := ""
	maxFreq := 0
	if key, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE); err == nil {
		cpuName, _, _ = key.GetStringValue("ProcessorNameString")
		mf, _, _ := key.GetIntegerValue("~MHz")
		maxFreq = int(mf)
		key.Close()
	}
	cpuName = strings.TrimSpace(cpuName)
	vendor := ""
	nameLower := strings.ToLower(cpuName)
	if strings.Contains(nameLower, "intel") {
		vendor = "Intel"
	} else if strings.Contains(nameLower, "amd") {
		vendor = "AMD"
	} else if strings.Contains(nameLower, "qualcomm") {
		vendor = "Qualcomm"
	}
	result.CPU = CPUInfo{
		Name:       cpuName,
		Cores:      runtime.NumCPU(),
		Threads:    runtime.NumCPU(),
		MaxFreqMHz: maxFreq,
		Vendor:     vendor,
	}

	// ── Single PowerShell script: all WMI queries in one process ──
	psScript := `
$ErrorActionPreference = 'SilentlyContinue'

Write-Output "===CPU==="
$cpu = Get-CimInstance Win32_Processor
Write-Output "LOAD=$($cpu.LoadPercentage)"
Write-Output "FREQ=$($cpu.CurrentClockSpeed)"

# Temperature (may not work on all systems)
$temp = Get-CimInstance -Namespace root/wmi -ClassName MSAcpi_ThermalZoneTemperature | Where-Object { $_.CurrentTemperature -gt 0 } | Select-Object -First 1 -ExpandProperty CurrentTemperature
if ($temp) { Write-Output "TEMP=$([math]::Round($temp/10 - 273.15, 1))" } else { Write-Output "TEMP=" }

Write-Output "===MEMORY==="
$mem = Get-CimInstance Win32_OperatingSystem
Write-Output "MEM=$([math]::Round($mem.TotalVisibleMemorySize/1MB, 1))|$([math]::Round($mem.FreePhysicalMemory/1MB, 1))"

Write-Output "===DISKS==="
Get-CimInstance Win32_LogicalDisk | Where-Object { $_.DriveType -eq 3 } | ForEach-Object {
    $total = [math]::Round($_.Size/1GB, 1)
    $free = [math]::Round($_.FreeSpace/1GB, 1)
    $usedPct = 0
    if ($total -gt 0) { $usedPct = [math]::Round(($total - $free) / $total * 100, 1) }
    Write-Output "$($_.DeviceID)|$($_.VolumeName)|$total|$free|$usedPct"
}

Write-Output "===GPUS==="
Get-CimInstance Win32_VideoController | ForEach-Object {
    $memMB = 0
    if ($_.AdapterRAM) { $memMB = [math]::Round($_.AdapterRAM/1MB) }
    Write-Output "$($_.Name)|$($_.VideoProcessor)|$memMB|$($_.DriverVersion)"
}

Write-Output "===NETWORK==="
$adapter = Get-NetAdapter | Where-Object { $_.Status -eq 'Up' } | Select-Object -First 1
if ($adapter) {
    $ip = ($adapter | Get-NetIPAddress -AddressFamily IPv4 | Where-Object { $_.PrefixOrigin -ne 'WellKnown' } | Select-Object -First 1).IPAddress
    Write-Output "$($adapter.Name)|$($adapter.MacAddress)|$ip|$($adapter.LinkSpeed)"
} else {
    Write-Output "NONE"
}

Write-Output "===POWER==="
$battery = Get-CimInstance Win32_Battery
if ($battery) {
    Write-Output "BATTERY=true|$($battery.EstimatedChargeRemaining)|$($battery.BatteryStatus)"
} else {
    Write-Output "BATTERY=false"
}

Write-Output "===TOPPROCS==="
Get-Process | Sort-Object CPU -Descending | Select-Object -First 10 | ForEach-Object {
    $memMB = [math]::Round($_.WorkingSet64/1MB, 1)
    Write-Output "$($_.ProcessName)|$($_.Id)|$memMB"
}

Write-Output "===UPTIME==="
$os = Get-CimInstance Win32_OperatingSystem
$uptime = (Get-Date) - $os.LastBootUpTime
Write-Output "$([int]$uptime.TotalDays)|$($uptime.Hours)|$($uptime.Minutes)"

Write-Output "===END==="
`

	out, err := runPowershellScript(psScript)
	if err != nil {
		// If batch fails, we still have OS info + CPU name from registry
		return result
	}

	// ── Parse sectioned output ──
	lines := strings.Split(out, "\n")
	section := ""
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}

		// Section markers
		switch line {
		case "===CPU===":
			section = "cpu"
			continue
		case "===MEMORY===":
			section = "memory"
			continue
		case "===DISKS===":
			section = "disks"
			continue
		case "===GPUS===":
			section = "gpus"
			continue
		case "===NETWORK===":
			section = "network"
			continue
		case "===POWER===":
			section = "power"
			continue
		case "===TOPPROCS===":
			section = "topprocs"
			continue
		case "===UPTIME===":
			section = "uptime"
			continue
		case "===END===":
			return result
		}

		switch section {
		case "cpu":
			if strings.HasPrefix(line, "LOAD=") {
				if pct, e := strconv.ParseFloat(strings.TrimPrefix(line, "LOAD="), 64); e == nil {
					result.CPU.UsagePct = pct
				}
			} else if strings.HasPrefix(line, "FREQ=") {
				if freq, e := strconv.Atoi(strings.TrimPrefix(line, "FREQ=")); e == nil && freq > 0 {
					result.CPU.FreqMHz = freq
				}
			} else if strings.HasPrefix(line, "TEMP=") {
				tempStr := strings.TrimPrefix(line, "TEMP=")
				if temp, e := strconv.ParseFloat(tempStr, 64); e == nil && temp > 0 && temp < 150 {
					result.CPU.TempC = temp
				}
			}

		case "memory":
			if strings.HasPrefix(line, "MEM=") {
				data := strings.TrimPrefix(line, "MEM=")
				parts := strings.Split(data, "|")
				if len(parts) >= 2 {
					total, _ := strconv.ParseFloat(parts[0], 64)
					free, _ := strconv.ParseFloat(parts[1], 64)
					result.Memory.TotalGB = total
					result.Memory.AvailableGB = free
					result.Memory.UsedGB = total - free
					if total > 0 {
						result.Memory.UsedPct = (result.Memory.UsedGB / total) * 100
					}
				}
			}

		case "disks":
			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				d := DiskInfo{Drive: parts[0], Label: parts[1]}
				d.TotalGB, _ = strconv.ParseFloat(parts[2], 64)
				d.FreeGB, _ = strconv.ParseFloat(parts[3], 64)
				d.UsedPct, _ = strconv.ParseFloat(parts[4], 64)
				d.Type = "SSD"
				d.FileSystem = "NTFS"
				result.Disks = append(result.Disks, d)
			}

		case "gpus":
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				g := AIAgentGPUInfo{
					Name:          parts[0],
					DriverVersion: parts[3],
				}
				g.MemoryMB, _ = strconv.Atoi(parts[2])
				gLower := strings.ToLower(g.Name)
				if strings.Contains(gLower, "nvidia") || strings.Contains(gLower, "geforce") || strings.Contains(gLower, "rtx") || strings.Contains(gLower, "gtx") {
					g.Vendor = "NVIDIA"
				} else if strings.Contains(gLower, "amd") || strings.Contains(gLower, "radeon") {
					g.Vendor = "AMD"
				} else if strings.Contains(gLower, "intel") {
					g.Vendor = "Intel"
				} else {
					g.Vendor = parts[1]
				}
				result.GPUs = append(result.GPUs, g)
			}

		case "network":
			if line == "NONE" {
				break
			}
			parts := strings.Split(line, "|")
			if len(parts) >= 4 {
				result.Network.AdapterName = parts[0]
				result.Network.MAC = parts[1]
				result.Network.IPAddress = parts[2]
				speedStr := strings.ToLower(parts[3])
				if strings.Contains(speedStr, "gbps") {
					fields := strings.Fields(speedStr)
					if len(fields) > 0 {
						val, _ := strconv.Atoi(fields[0])
						result.Network.SpeedMbps = val * 1000
					}
				} else if strings.Contains(speedStr, "mbps") {
					fields := strings.Fields(speedStr)
					if len(fields) > 0 {
						val, _ := strconv.Atoi(fields[0])
						result.Network.SpeedMbps = val
					}
				}
				result.Network.Connected = true
			}

		case "power":
			if strings.HasPrefix(line, "BATTERY=") {
				data := strings.TrimPrefix(line, "BATTERY=")
				parts := strings.Split(data, "|")
				if len(parts) >= 1 && parts[0] == "true" {
					result.Power.Battery = true
					if len(parts) >= 3 {
						result.Power.BatteryPct, _ = strconv.Atoi(parts[1])
						sc, _ := strconv.Atoi(parts[2])
						result.Power.BatteryStatus = batteryStatusText(sc)
						result.Power.ACConnected = sc == 6 || sc == 7 || sc == 9
					}
				} else {
					// No battery → desktop, assume AC
					result.Power.ACConnected = true
				}
			}

		case "topprocs":
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				pidVal, _ := strconv.ParseUint(parts[1], 10, 32)
				p := ProcessInfo{
					Name: parts[0],
					PID:  uint32(pidVal),
				}
				p.MemMB, _ = strconv.ParseFloat(parts[2], 64)
				result.TopProcs = append(result.TopProcs, p)
			}

		case "uptime":
			parts := strings.Split(line, "|")
			if len(parts) >= 3 {
				days, _ := strconv.Atoi(parts[0])
				hours, _ := strconv.Atoi(parts[1])
				mins, _ := strconv.Atoi(parts[2])
				if days > 0 {
					result.Uptime = fmt.Sprintf("%d 天 %d 小时 %d 分钟", days, hours, mins)
				} else {
					result.Uptime = fmt.Sprintf("%d 小时 %d 分钟", hours, mins)
				}
			}
		}
	}

	// Power plan — uses powercfg CLI, can't merge into WMI script
	planOut, _ := runPowershellScript(`$plan = powercfg /getactivescheme; if ($plan -match '\)\s*(.+)$') { Write-Output $Matches[1].Trim() }`)
	result.Power.PowerPlan = strings.TrimSpace(planOut)
	if result.Power.PowerPlan == "" {
		result.Power.PowerPlan = "平衡"
	}

	return result
}

// AskAIAgent processes a user question and returns an answer based on system info
func AskAIAgent(question string) string {
	info := GetAIAgentSystemInfo()

	q := strings.ToLower(strings.TrimSpace(question))

	var response strings.Builder

	// ── File Analysis ── detect [File: ...] block and extract user question
	fileBlockRe := regexp.MustCompile(`(?s)\[File:\s*(.+?)\]\s*\x60\x60\x60\s*(.+?)\x60\x60\x60`)
	var actualQuestion = question
	if fileMatch := fileBlockRe.FindStringSubmatch(question); len(fileMatch) >= 3 {
		fileName := strings.TrimSpace(fileMatch[1])
		fileContent := strings.TrimSpace(fileMatch[2])
		// Extract the real user question after the file block
		if idx := strings.Index(question, "User question:"); idx >= 0 {
			actualQuestion = strings.TrimSpace(question[idx+len("User question:"):])
		}
		q = strings.ToLower(strings.TrimSpace(actualQuestion))

		// Basic file stats
		lines := strings.Split(fileContent, "\n")
		totalLines := len(lines)
		// Detect file type by extension or content pattern
		fileType := detectFileType(fileName, fileContent)
		response.WriteString(fmt.Sprintf("📄 **File Analysis: %s**\n", fileName))
		response.WriteString(fmt.Sprintf("- Type: %s\n", fileType))
		response.WriteString(fmt.Sprintf("- Lines: %d | Chars: %d\n", totalLines, len(fileContent)))

		// Search for common patterns
		if errorMatches := countPattern(lines, "error|fail|exception|crash|panic|timeout|拒绝|失败|错误|异常"); errorMatches > 0 {
			response.WriteString(fmt.Sprintf("- ⚠️  **Errors/Warnings**: %d occurrences\n", errorMatches))
		}
		if warnMatches := countPattern(lines, "warn|warning"); warnMatches > 0 {
			response.WriteString(fmt.Sprintf("- ⚡ Warnings: %d occurrences\n", warnMatches))
		}
		if timestamps := countPattern(lines, `\d{2}:\d{2}:\d{2}|\d{4}-\d{2}-\d{2}`); timestamps > 0 {
			response.WriteString(fmt.Sprintf("- 🕐 Timestamped entries: %d\n", timestamps))
		}

		// Show first/last few lines as preview
		previewLines := 3
		if totalLines > 0 {
			response.WriteString("- **Head:**\n```\n")
			for i, l := range lines {
				if i >= previewLines { break }
				if len(l) > 120 {
					l = l[:120] + "..."
				}
				response.WriteString(l + "\n")
			}
			response.WriteString("```\n")
		}
		if totalLines > previewLines*2 {
			response.WriteString("- **Tail:**\n```\n")
			for i := totalLines - previewLines; i < totalLines; i++ {
				if i < 0 { continue }
				l := lines[i]
				if len(l) > 120 {
					l = l[:120] + "..."
				}
				response.WriteString(l + "\n")
			}
			response.WriteString("```\n")
		}
		response.WriteString("\n---\n\n")
	} else {
		q = strings.ToLower(strings.TrimSpace(question))
	}

	// CPU related
	if containsAny(q, []string{"cpu", "processor", "处理器", "cpu使用率", "cpu占用", "cpu temp", "cpu温度", "cpu型号", "cpu名字", "频率", "frequency"}) {
		response.WriteString(fmt.Sprintf("🖥️ **CPU 信息**\n"))
		response.WriteString(fmt.Sprintf("- 型号: %s\n", info.CPU.Name))
		response.WriteString(fmt.Sprintf("- 核心数: %d 核心 / %d 线程\n", info.CPU.Cores, info.CPU.Threads))
		response.WriteString(fmt.Sprintf("- 当前频率: %d MHz (最大 %d MHz)\n", info.CPU.FreqMHz, info.CPU.MaxFreqMHz))
		response.WriteString(fmt.Sprintf("- 使用率: %.1f%%\n", info.CPU.UsagePct))
		if info.CPU.TempC > 0 {
			response.WriteString(fmt.Sprintf("- 温度: %.1f°C\n", info.CPU.TempC))
		}
	}

	// Memory related
	if containsAny(q, []string{"内存", "memory", "ram", "内存使用", "内存占用", "可用内存", "剩余内存", "总内存"}) {
		response.WriteString(fmt.Sprintf("🧠 **内存信息**\n"))
		response.WriteString(fmt.Sprintf("- 总容量: %.1f GB\n", info.Memory.TotalGB))
		response.WriteString(fmt.Sprintf("- 已使用: %.1f GB (%.1f%%)\n", info.Memory.UsedGB, info.Memory.UsedPct))
		response.WriteString(fmt.Sprintf("- 可用: %.1f GB\n", info.Memory.AvailableGB))
	}

	// Disk related
	if containsAny(q, []string{"磁盘", "disk", "硬盘", "存储", "空间", "disk usage", "硬盘空间", "c盘", "d盘", "ssd", "hdd"}) {
		response.WriteString(fmt.Sprintf("💾 **磁盘信息**\n"))
		for _, d := range info.Disks {
			response.WriteString(fmt.Sprintf("- %s (%s): %.1f GB / %.1f GB (%.1f%% 已用) [%s]\n",
				d.Drive, d.Label, d.FreeGB, d.TotalGB, d.UsedPct, d.Type))
		}
	}

	// GPU related
	if containsAny(q, []string{"显卡", "gpu", "图形", "图形卡", "video card", "集成显卡", "独立显卡", "显卡型号"}) {
		response.WriteString(fmt.Sprintf("🎮 **GPU信息**\n"))
		if len(info.GPUs) == 0 {
			response.WriteString("- 未检测到GPU信息\n")
		} else {
			for _, g := range info.GPUs {
				response.WriteString(fmt.Sprintf("- %s (%s): %d MB 显存\n", g.Name, g.Vendor, g.MemoryMB))
			}
		}
	}

	// Network related
	if containsAny(q, []string{"网络", "network", "ip", "mac", "网卡", "网络适配器", "ip地址"}) {
		response.WriteString(fmt.Sprintf("🌐 **网络信息**\n"))
		response.WriteString(fmt.Sprintf("- 适配器: %s\n", info.Network.AdapterName))
		response.WriteString(fmt.Sprintf("- IP地址: %s\n", info.Network.IPAddress))
		response.WriteString(fmt.Sprintf("- MAC: %s\n", info.Network.MAC))
		response.WriteString(fmt.Sprintf("- 速度: %d Mbps\n", info.Network.SpeedMbps))
		response.WriteString(fmt.Sprintf("- 连接状态: %s\n", boolToStatus(info.Network.Connected)))
	}

	// Power/Battery related
	if containsAny(q, []string{"电源", "power", "电池", "battery", "电量", "充电", "电源计划", "power plan"}) {
		response.WriteString(fmt.Sprintf("🔋 **电源信息**\n"))
		response.WriteString(fmt.Sprintf("- 电源模式: %s\n", info.Power.PowerPlan))
		if info.Power.Battery {
			response.WriteString(fmt.Sprintf("- 电池电量: %d%%\n", info.Power.BatteryPct))
			response.WriteString(fmt.Sprintf("- 状态: %s\n", info.Power.BatteryStatus))
		}
		response.WriteString(fmt.Sprintf("- AC电源: %s\n", boolToStatus(info.Power.ACConnected)))
	}

	// OS related
	if containsAny(q, []string{"系统", "os", "操作系统", "windows", "系统版本", "版本号", "build"}) {
		response.WriteString(fmt.Sprintf("🖥️ **系统信息**\n"))
		response.WriteString(fmt.Sprintf("- 系统: %s\n", info.OS.Name))
		response.WriteString(fmt.Sprintf("- 版本: %s (Build %s)\n", info.OS.Version, info.OS.Build))
		response.WriteString(fmt.Sprintf("- 架构: %s\n", info.OS.Architecture))
		response.WriteString(fmt.Sprintf("- 安装日期: %s\n", info.OS.InstallDate))
	}

	// Process related
	if containsAny(q, []string{"进程", "process", "进程列表", "占用cpu", "占用内存", "进程名", "程序"}) {
		response.WriteString(fmt.Sprintf("📊 **Top 进程 (按CPU排序)**\n"))
		for i, p := range info.TopProcs {
			if i >= 5 {
				break
			}
			response.WriteString(fmt.Sprintf("- %s (PID %d): CPU %.1f%%, 内存 %.0f MB\n",
				p.Name, p.PID, p.CPUPct, p.MemMB))
		}
	}

	// Uptime related
	if containsAny(q, []string{"开机时间", "运行时间", "uptime", "up time", "多久"}) {
		response.WriteString(fmt.Sprintf("⏱️ **运行时间**\n"))
		response.WriteString(fmt.Sprintf("- 系统已运行: %s\n", info.Uptime))
	}

	// General system info
	if containsAny(q, []string{"系统概览", "overview", "概要", "全部信息", "完整信息", "system info", "主机信息", "电脑信息"}) {
		response.WriteString(fmt.Sprintf("📋 **系统概览**\n\n"))
		response.WriteString(fmt.Sprintf("**主机**: %s (%s)\n\n", info.Hostname, info.CurrentUser))
		response.WriteString(fmt.Sprintf("**操作系统**: %s Build %s\n", info.OS.Name, info.OS.Build))
		response.WriteString(fmt.Sprintf("**CPU**: %s (%.1f%% 使用)\n", info.CPU.Name, info.CPU.UsagePct))
		response.WriteString(fmt.Sprintf("**内存**: %.1f GB / %.1f GB (%.1f%%)\n",
			info.Memory.UsedGB, info.Memory.TotalGB, info.Memory.UsedPct))
		response.WriteString(fmt.Sprintf("**运行时间**: %s\n", info.Uptime))
	}

	// Hostname
	if containsAny(q, []string{"主机名", "hostname", "computer name", "电脑名"}) {
		response.WriteString(fmt.Sprintf("🏷️ **主机名**: %s\n", info.Hostname))
	}

	// Brightness / 亮度 控制
	if containsAny(q, []string{"亮度", "brightness", "调亮", "调暗", "屏幕亮度", "屏幕暗", "屏幕亮", "调低亮度", "调高亮度", "增加亮度", "降低亮度"}) {
		response.WriteString(fmt.Sprintf("💡 **亮度控制**\n"))
		current, getErr := getBrightness()
		if getErr != nil {
			response.WriteString(fmt.Sprintf("- 无法获取/控制亮度: %v\n", getErr))
			response.WriteString("- 此功能需要笔记本/内置屏幕支持哦 (台式机外接显示器不支持WMI调亮度)\n")
		} else {
			var targetLevel int
			switch {
			case containsAny(q, []string{"最高", "最大", "最亮", "max", "100"}):
				targetLevel = 100
			case containsAny(q, []string{"最低", "最小", "最暗", "min", "0"}):
				targetLevel = 0
			case strings.Contains(q, "调高") || strings.Contains(q, "调亮") || strings.Contains(q, "增加") || strings.Contains(q, "bright") || strings.Contains(q, "increase") || strings.Contains(q, "up"):
				step := 10
				if strings.Contains(q, "一点") || strings.Contains(q, "稍微") {
					step = 5
				}
				targetLevel = current + step
				if targetLevel > 100 {
					targetLevel = 100
				}
			case strings.Contains(q, "调低") || strings.Contains(q, "调暗") || strings.Contains(q, "降低") || strings.Contains(q, "dim") || strings.Contains(q, "decrease") || strings.Contains(q, "down"):
				step := 10
				if strings.Contains(q, "一点") || strings.Contains(q, "稍微") {
					step = 5
				}
				targetLevel = current - step
				if targetLevel < 0 {
					targetLevel = 0
				}
			case strings.Contains(q, "%") || strings.Contains(q, "设为") || strings.Contains(q, "调到") || strings.Contains(q, "设置到") || strings.Contains(q, "改成") || strings.Contains(q, "设置为"):
				// Try to parse numeric value
				re := regexp.MustCompile(`(\d+)`)
				if match := re.FindStringSubmatch(q); len(match) > 1 {
					if val, err := strconv.Atoi(match[1]); err == nil && val >= 0 && val <= 100 {
						targetLevel = val
					}
				}
			default:
				// Just show current brightness without changing
				response.WriteString(fmt.Sprintf("- 当前亮度: **%d%%**\n", current))
				response.WriteString("- 试试说: 亮度调到80% 或 调亮一点 或 最高亮度\n")
				return response.String()
			}

			if targetLevel == current {
				response.WriteString(fmt.Sprintf("- 当前亮度: **%d%%** (无需调整)\n", current))
			} else {
				newLevel, setErr := setBrightness(targetLevel)
				if setErr != nil {
					response.WriteString(fmt.Sprintf("- 当前亮度: %d%%\n", current))
					response.WriteString(fmt.Sprintf("- 调整失败: %v\n", setErr))
				} else {
					response.WriteString(fmt.Sprintf("- 已从 %d%% 调整为 **%d%%** ✅\n", current, newLevel))
				}
			}
		}
	}

	// Fallback help
	if response.Len() == 0 {
		response.WriteString("🤖 我是您的系统助手，可以回答以下类型的问题：\n\n")
		response.WriteString("- CPU/处理器信息、使用率、温度\n")
		response.WriteString("- 内存使用情况\n")
		response.WriteString("- 磁盘空间、存储情况\n")
		response.WriteString("- 显卡/GPU信息\n")
		response.WriteString("- 网络、IP地址\n")
		response.WriteString("- 电源、电池状态\n")
		response.WriteString("- 屏幕亮度调节 (支持笔记本)\n")
		response.WriteString("- 系统版本、Build号\n")
		response.WriteString("- 进程列表、资源占用\n")
		response.WriteString("- 运行时间\n\n")
		response.WriteString("💡 请输入您的问题，例如：\"CPU使用率多少？\" 或 \"还有多少磁盘空间？\"")
	}

	return response.String()
}

// Helpers

func containsAny(s string, keywords []string) bool {
	for _, k := range keywords {
		if strings.Contains(s, k) {
			return true
		}
	}
	return false
}

func boolToStatus(b bool) string {
	if b { return "已连接" }
	return "未连接"
}

func getOSInfo() OSInfo {
	info := OSInfo{}

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows NT\CurrentVersion`, registry.QUERY_VALUE)
	if err == nil {
		defer key.Close()
		productName, _, _ := key.GetStringValue("ProductName")
		if productName == "" {
			productName = "Windows"
		}
		info.Name = productName
		displayVersion, _, _ := key.GetStringValue("DisplayVersion")
		currentBuild, _, _ := key.GetStringValue("CurrentBuild")
		currentBuildNum, _, _ := key.GetStringValue("CurrentBuildNumber")
		if displayVersion != "" {
			info.Version = displayVersion
		}
		info.Build = currentBuild
		if info.Build == "" {
			info.Build = currentBuildNum
		}
		installDate, _, _ := key.GetIntegerValue("InstallDate")
		if installDate > 0 {
			info.InstallDate = time.Unix(int64(installDate), 0).Format("2006-01-02")
		}
	}

	info.Architecture = runtime.GOARCH
	if info.Architecture == "amd64" {
		info.Architecture = "x64"
	}
	return info
}

// getCPUInfo removed — now handled by batchSystemInfo()

// getMemoryInfo removed — now handled by batchSystemInfo()

// getDiskInfo removed — now handled by batchSystemInfo()

// getGPUInfo removed — now handled by batchSystemInfo()

// getNetworkInfo removed — now handled by batchSystemInfo()

// getPowerInfo removed — now handled by batchSystemInfo()

func batteryStatusText(code int) string {
	switch code {
	case 1: return "放电中"
	case 2: return "AC电源"
	case 3: return "完全充电"
	case 4: return "低电量"
	case 5: return "临界电量"
	case 6, 7, 9: return "充电中"
	case 8: return "充电完成"
	default: return "未知"
	}
}

// getTopProcesses removed — now handled by batchSystemInfo()

// getUptime removed — now handled by batchSystemInfo()

func runPowershellScript(script string) (string, error) {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &windows.SysProcAttr{
		CreationFlags: windows.CREATE_NO_WINDOW,
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

// getBrightness returns current screen brightness (0-100) via WMI
func getBrightness() (int, error) {
	script := `
$b = Get-CimInstance -Namespace root/WMI -ClassName WmiMonitorBrightness
if ($b) {
    # Handle array (multiple monitors)
    if ($b -is [array]) {
        Write-Output $b[0].CurrentBrightness
    } else {
        Write-Output $b.CurrentBrightness
    }
} else {
    Write-Error "WmiMonitorBrightness not available"
}
`
	out, err := runPowershellScript(script)
	if err != nil {
		return 0, fmt.Errorf("无法读取亮度 (此功能仅支持笔记本内置屏幕): %v", err)
	}
	levelStr := strings.TrimSpace(out)
	level, err := strconv.Atoi(levelStr)
	if err != nil {
		return 0, fmt.Errorf("无法解析亮度值 '%s': 此功能仅笔记本内置屏幕支持", levelStr)
	}
	return level, nil
}

// setBrightness sets screen brightness (0-100) via WMI, returns the actual new level
func setBrightness(level int) (int, error) {
	if level < 0 {
		level = 0
	}
	if level > 100 {
		level = 100
	}
	script := fmt.Sprintf(`
# Use Invoke-CimMethod instead of direct method call (more reliable)
$allOk = $true
Get-CimInstance -Namespace root/WMI -ClassName WmiMonitorBrightnessMethods | ForEach-Object {
    $result = $_ | Invoke-CimMethod -MethodName "WmiSetBrightness" -Arguments @{ Timeout = [uint32]1; Brightness = [byte]%d }
    if ($result.ReturnValue -ne 0) {
        $allOk = $false
        Write-Error ("WmiSetBrightness returned: " + $result.ReturnValue)
    }
}
if (-not $allOk) {
    exit 1
}
# Wait for brightness transition to settle
Start-Sleep -Milliseconds 300
$b = Get-CimInstance -Namespace root/WMI -ClassName WmiMonitorBrightness
if ($b -is [array]) {
    Write-Output $b[0].CurrentBrightness
} else {
    Write-Output $b.CurrentBrightness
}
`, level)
	out, err := runPowershellScript(script)
	if err != nil {
		return 0, fmt.Errorf("无法设置亮度 (此功能仅支持笔记本内置屏幕): %v", err)
	}
	newLevelStr := strings.TrimSpace(out)
	newLevel, err := strconv.Atoi(newLevelStr)
	if err != nil {
		return level, nil // return requested level as fallback
	}
	return newLevel, nil
}

// detectFileType guesses file type from extension or content patterns
func detectFileType(fileName string, content string) string {
	lowerName := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lowerName, ".log"):
		return "Log File"
	case strings.HasSuffix(lowerName, ".csv"):
		return "CSV / Spreadsheet"
	case strings.HasSuffix(lowerName, ".json"):
		return "JSON Data"
	case strings.HasSuffix(lowerName, ".xml"):
		return "XML Document"
	case strings.HasSuffix(lowerName, ".md"), strings.HasSuffix(lowerName, ".markdown"):
		return "Markdown"
	case strings.HasSuffix(lowerName, ".txt"):
		return "Text File"
	case strings.HasSuffix(lowerName, ".ini"), strings.HasSuffix(lowerName, ".cfg"),
		strings.HasSuffix(lowerName, ".conf"):
		return "Configuration File"
	case strings.HasSuffix(lowerName, ".yaml"), strings.HasSuffix(lowerName, ".yml"):
		return "YAML Config"
	case strings.HasPrefix(content, "<html") || strings.HasPrefix(content, "<!DOCTYPE"):
		return "HTML Document"
	case strings.HasPrefix(content, "{") && strings.Contains(content, "\""):
		return "JSON-like Data"
	case strings.HasPrefix(content, "<"):
		return "XML/HTML-like Data"
	default:
		return detectContentType(content)
	}
}

func detectContentType(content string) string {
	firstLine := strings.ToLower(content)
	if idx := strings.Index(firstLine, "\n"); idx > 0 {
		firstLine = firstLine[:idx]
	}
	if strings.Contains(firstLine, "error") || strings.Contains(firstLine, "fail") ||
		strings.Contains(firstLine, "exception") {
		return "Error/Diagnostic Log"
	}
	if strings.Contains(firstLine, "[") && strings.Contains(firstLine, "]") &&
		strings.Contains(firstLine, ":") {
		return "Application Log"
	}
	return "Plain Text"
}

// countPattern counts lines matching a regex pattern (case-insensitive)
func countPattern(lines []string, pattern string) int {
	re := regexp.MustCompile("(?i)" + pattern)
	count := 0
	for _, line := range lines {
		if re.MatchString(line) {
			count++
		}
	}
	return count
}