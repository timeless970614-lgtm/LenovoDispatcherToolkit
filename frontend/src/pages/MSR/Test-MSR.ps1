<#
.SYNOPSIS
    MSR Test Script - Based on MSR.pptx
.DESCRIPTION
    Tests and configures Intel MSR parameters for power management.
    Uses various methods to access CPU information and MSR-like settings.
.NOTES
    Based on: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx
    Requires: Administrator privileges for some operations
#>

param(
    [switch]$ReadAll,
    [switch]$ShowRegisters,
    [switch]$TestEnergyBias,
    [switch]$TestPerfCTL,
    [switch]$TestTurboBoost,
    [switch]$TestRAPL,
    [switch]$TestClockMod,
    [switch]$TestUncore,
    [switch]$MonitorFreq
)

# Color output
function Write-Header($text) {
    $width = 78
    $padding = [Math]::Max(0, ($width - $text.Length - 4) / 2)
    Write-Host ""
    Write-Host ("=" * $width)
    Write-Host (" " * [int]$padding + "[ $text ]")
    Write-Host ("=" * $width)
}

function Write-Section($text) {
    Write-Host ""
    Write-Host "--- $text ---"
}

# MSR Register definitions from PPT
$MSRRegisters = @{
    "IA32_ENERGY_PERF_BIAS" = 0x1B0
    "IA32_PERF_CTL" = 0x198
    "IA32_PERF_STATUS" = 0x199
    "IA32_MISC_ENABLE" = 0x1A0
    "IA32_CLOCK_MODULATION" = 0x19A
    "IA32_APERF" = 0xE7
    "IA32_MPERF" = 0xE8
    "MSR_UNCORE_RATIO_LIMIT" = 0x620
    "MSR_RAPL_POWER_UNIT" = 0x606
    "MSR_PKG_POWER_LIMIT" = 0x610
    "MSR_PKG_ENERGY_STATUS" = 0x611
    "MSR_PP0_POWER_LIMIT" = 0x638
    "MSR_DRAM_POWER_LIMIT" = 0x618
}

# Check for admin rights
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)

Write-Header "Intel MSR Test Application"
Write-Host "Based on: MSR.pptx - Intel Power Management"
Write-Host ""
Write-Host "Administrator Mode: $(if($isAdmin){'YES'}else{'NO (Some features limited)'})"

# Get CPU Information
Write-Section "CPU Information"
$cpuInfo = Get-CimInstance -ClassName Win32_Processor | Select-Object Name, NumberOfCores, NumberOfLogicalProcessors, MaxClockSpeed, CurrentClockSpeed
foreach ($cpu in $cpuInfo) {
    Write-Host "CPU: $($cpu.Name)"
    Write-Host "  Cores: $($cpu.NumberOfCores) Physical, $($cpu.NumberOfLogicalProcessors) Logical"
    Write-Host "  Max Clock: $($cpu.MaxClockSpeed) MHz"
    Write-Host "  Current Clock: $($cpu.CurrentClockSpeed) MHz"
}

# Get Power Plans
Write-Section "Power Plans"
$powerPlans = powercfg /list 2>$null
if ($LASTEXITCODE -eq 0) {
    $currentPlan = $powerPlans | Where-Object { $_ -match "Active.*:" }
    Write-Host $currentPlan
}

# Get Power Settings via WMI
Write-Section "Current Power Settings (WMI)"
try {
    $powerSettings = Get-CimInstance -Namespace root/power -ClassName Win32_PowerMeter 2>$null
    if ($powerSettings) {
        $powerSettings | Format-List
    }
} catch {
    Write-Host "(Power Meter not available)"
}

# Try to read MSR-like information from registry
Write-Section "Processor Registry Settings"
$processorKey = "HKLM:\HARDWARE\DESCRIPTION\System\CentralProcessor\0"
if (Test-Path $processorKey) {
    $procInfo = Get-ItemProperty $processorKey
    Write-Host "Processor Name: $($procInfo.ProcessorNameString)"
    Write-Host "Identifier: $($procInfo.Identifier)"
}

# Energy Performance Bias from registry (if available)
Write-Section "Energy Performance Bias (from registry/sysfs equivalent)"
$energyBiasPath = "HKLM:\SYSTEM\CurrentControlSet\Control\Power\User\PowerSchemes"
if (Test-Path $energyBiasPath) {
    Write-Host "Power Scheme settings available at: $energyBiasPath"
}

# P-State information
Write-Section "P-State Information"
Write-Host "P-States (Performance States) control CPU frequency and voltage."
Write-Host "Managed by: ACPI P-State driver or Intel P-State driver"
Write-Host ""

# Show all registers if requested
if ($ShowRegisters) {
    Write-Header "MSR Register Definitions from PPT"
    Write-Host ""
    Write-Host "Address    Name                              Description"
    Write-Host "--------   --------------------------------  ----------------------------------"
    
    foreach ($reg in $MSRRegisters.GetEnumerator() | Sort-Object Value) {
        $addr = "0x{0:X3}" -f $reg.Value
        $name = $reg.Key
        $desc = switch ($reg.Key) {
            "IA32_ENERGY_PERF_BIAS" { "Energy Performance Bias Hint (0-15)" }
            "IA32_PERF_CTL" { "P-State Target Control" }
            "IA32_PERF_STATUS" { "Current P-State (read-only)" }
            "IA32_MISC_ENABLE" { "Turbo Boost control" }
            "IA32_CLOCK_MODULATION" { "Clock Modulation (T-state)" }
            "IA32_APERF" { "Actual Performance Counter" }
            "IA32_MPERF" { "Maximum Performance Counter" }
            "MSR_UNCORE_RATIO_LIMIT" { "Uncore Frequency Limits" }
            "MSR_RAPL_POWER_UNIT" { "RAPL Units (power/energy/time)" }
            "MSR_PKG_POWER_LIMIT" { "Package Power Limit" }
            "MSR_PKG_ENERGY_STATUS" { "Package Energy Consumed" }
            "MSR_PP0_POWER_LIMIT" { "CPU Cores Power Limit" }
            "MSR_DRAM_POWER_LIMIT" { "DRAM Power Limit" }
        }
        Write-Host "$addr   $name  $desc"
    }
}

# Test Energy Performance Bias
if ($TestEnergyBias) {
    Write-Header "IA32_ENERGY_PERF_BIAS (0x1B0) Test"
    Write-Host ""
    Write-Host "From PPT Slide 5:"
    Write-Host "  - Bits 3:0: Software hint for power management"
    Write-Host "  - 0 = Highest Performance"
    Write-Host "  - 6 = Balance (default)"
    Write-Host "  - 15 = Maximum Energy Savings"
    Write-Host "  - Affects Turbo boost and Uncore Frequency Scaling"
    Write-Host ""
    
    # Try to read via WMI
    $os = Get-CimInstance -ClassName Win32_OperatingSystem
    Write-Host "OS Power Profile may affect this setting."
    
    # Show related registry
    $biasPath = "HKLM:\SYSTEM\CurrentControlSet\Services\intel_pstate"
    if (Test-Path $biasPath) {
        Write-Host "Intel P-State driver settings at: $biasPath"
        Get-ItemProperty $biasPath | Format-List
    }
}

# Test Performance Control
if ($TestPerfCTL) {
    Write-Header "IA32_PERF_CTL (0x198) Test"
    Write-Host ""
    Write-Host "From PPT Slide 6:"
    Write-Host "  - Intel P-State driver writes 16-bit value for state transitions"
    Write-Host "  - Controls target P-state (frequency/voltage)"
    Write-Host ""
    Write-Host "Use Intel XTU or ThrottleStop to modify this register."
    Write-Host ""
    
    # Show current frequencies
    $freqPath = "HKLM:\HARDWARE\DESCRIPTION\System\CentralProcessor\0"
    if (Test-Path $freqPath) {
        $proc = Get-ItemProperty $freqPath
        Write-Host "Current MHz: $($proc.~Mhz)"
    }
}

# Test Turbo Boost
if ($TestTurboBoost) {
    Write-Header "Turbo Boost Control (0x1A0) Test"
    Write-Host ""
    Write-Host "From PPT Slide 6:"
    Write-Host "  - IA32_MISC_ENABLE (0x1A0) controls Turbo Boost"
    Write-Host "  - Bit 38: Turbo Boost enable/disable"
    Write-Host ""
    Write-Host "WARNING: Disabling Turbo Boost may significantly impact performance!"
    Write-Host ""
    
    # Check BIOS setting (registry proxy)
    $biosTBPath = "HKLM:\SYSTEM\CurrentControlSet\Services\intel_pstate"
    Write-Host "Turbo Boost is controlled by:"
    Write-Host "  1. BIOS settings"
    Write-Host "  2. Intel P-State driver"
    Write-Host "  3. ACPI tables"
}

# Test RAPL
if ($TestRAPL) {
    Write-Header "RAPL (Running Average Power Limit) Test"
    Write-Host ""
    Write-Host "From PPT Slides 54-58:"
    Write-Host "  RAPL Domains:"
    Write-Host "    - PKG: Entire package (socket)"
    Write-Host "    - PP0: CPU cores power plane"
    Write-Host "    - PP1: Uncore/Graphics (if present)"
    Write-Host "    - DRAM: Memory (server only)"
    Write-Host ""
    
    # Try to read energy status via WMI
    Write-Host "Checking for power monitoring capabilities..."
    $counters = Get-Counter -ListSet "*power*" 2>$null
    if ($counters) {
        Write-Host "Available power counters:"
        $counters | ForEach-Object { Write-Host "  - $($_.CounterSetName)" }
    }
    
    Write-Host ""
    Write-Host "Key Registers:"
    Write-Host "  0x606: MSR_RAPL_POWER_UNIT - Units for power/energy/time"
    Write-Host "  0x610: MSR_PKG_POWER_LIMIT - Package power limit"
    Write-Host "  0x611: MSR_PKG_ENERGY_STATUS - Energy consumed (read-only)"
}

# Test Clock Modulation
if ($TestClockMod) {
    Write-Header "IA32_CLOCK_MODULATION (0x19A) Test"
    Write-Host ""
    Write-Host "From PPT Slide 18:"
    Write-Host "  - DDCM: Dynamic Duty Cycle Modulation (T-state)"
    Write-Host "  - Used for power management and thermal control"
    Write-Host ""
    Write-Host "Bits:"
    Write-Host "  Bit 4: Enable/Disable"
    Write-Host "  Bits 3:0: Duty Cycle (6.25% increments)"
}

# Test Uncore Ratio
if ($TestUncore) {
    Write-Header "MSR_UNCORE_RATIO_LIMIT (0x620) Test"
    Write-Host ""
    Write-Host "From PPT Slide 10:"
    Write-Host "  - Controls uncore frequency scaling"
    Write-Host "  - Uncore = cache, memory controller, etc."
    Write-Host ""
    Write-Host "Tip: If max = min, uncore frequency is locked!"
}

# Monitor Frequency
if ($MonitorFreq) {
    Write-Header "CPU Frequency Monitor (APERF/MPERF)"
    Write-Host ""
    Write-Host "From PPT Slide 7:"
    Write-Host "  - IA32_APERF (0xE7): Actual clock cycles in C0"
    Write-Host "  - IA32_MPERF (0xE8): Maximum cycles in C0"
    Write-Host ""
    Write-Host "Formula: Current Freq = NominalFreq × (ΔAPERF / ΔMPERF)"
    Write-Host ""
    
    Write-Host "Monitoring for 10 seconds (Ctrl+C to stop)..."
    Write-Host ""
    
    $count = 0
    while ($count -lt 10) {
        $freq = (Get-CimInstance -ClassName Win32_Processor).CurrentClockSpeed
        $time = Get-Date -Format "HH:mm:ss"
        Write-Host "[$time] Current Frequency: $freq MHz"
        Start-Sleep -Seconds 1
        $count++
    }
}

# Default: Show all
if (-not $ReadAll -and -not $ShowRegisters -and -not $TestEnergyBias -and 
    -not $TestPerfCTL -and -not $TestTurboBoost -and -not $TestRAPL -and 
    -not $TestClockMod -and -not $TestUncore -and -not $MonitorFreq) {
    
    Write-Header "Usage"
    Write-Host ""
    Write-Host "Options:"
    Write-Host "  -ReadAll           Read all available MSR information"
    Write-Host "  -ShowRegisters     Show MSR register definitions"
    Write-Host "  -TestEnergyBias    Test Energy Performance Bias"
    Write-Host "  -TestPerfCTL       Test Performance Control"
    Write-Host "  -TestTurboBoost    Test Turbo Boost"
    Write-Host "  -TestRAPL          Test RAPL Power Limits"
    Write-Host "  -TestClockMod      Test Clock Modulation"
    Write-Host "  -TestUncore        Test Uncore Ratio Limit"
    Write-Host "  -MonitorFreq       Monitor CPU Frequency"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\Test-MSR.ps1 -ShowRegisters"
    Write-Host "  .\Test-MSR.ps1 -TestRAPL -MonitorFreq"
}

Write-Host ""
Write-Host "==========================================="
Write-Host "For full MSR access, install WinRing0 driver"
Write-Host "See README.md for installation instructions"
Write-Host "==========================================="
