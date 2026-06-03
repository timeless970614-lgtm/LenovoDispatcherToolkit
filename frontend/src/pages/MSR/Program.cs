using System;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text;

namespace MSR_Test_App
{
    /// <summary>
    /// MSR Register Definitions from PPT slides
    /// Source: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx
    /// </summary>
    public static class MSRRegisters
    {
        // IA32_ENERGY_PERF_BIAS (0x1B0)
        // Software hint to guide hardware power management
        // Bits 3:0: 0=highest performance, 6=balance, 15=max energy savings
        public const uint IA32_ENERGY_PERF_BIAS = 0x1B0;

        // IA32_PERF_CTL (0x198)
        // Controls P-state transitions
        public const uint IA32_PERF_CTL = 0x198;

        // IA32_PERF_STATUS (0x199)
        // Read-only: Current P-state
        public const uint IA32_PERF_STATUS = 0x199;

        // IA32_MISC_ENABLE (0x1A0)
        // Turbo boost enable/disable
        public const uint IA32_MISC_ENABLE = 0x1A0;

        // IA32_CLOCK_MODULATION (0x19A)
        // Clock modulation for power/thermal management
        public const uint IA32_CLOCK_MODULATION = 0x19A;

        // MSR_UNCORE_RATIO_LIMIT (0x620)
        // Uncore frequency scaling limits
        public const uint MSR_UNCORE_RATIO_LIMIT = 0x620;

        // RAPL registers
        public const uint MSR_RAPL_POWER_UNIT = 0x606;
        public const uint MSR_PKG_POWER_LIMIT = 0x610;
        public const uint MSR_PKG_ENERGY_STATUS = 0x611;
        public const uint MSR_PKG_POWER_INFO = 0x614;
        public const uint MSR_PP0_POWER_LIMIT = 0x638;
        public const uint MSR_PP0_ENERGY_STATUS = 0x639;
        public const uint MSR_DRAM_POWER_LIMIT = 0x618;
        public const uint MSR_DRAM_ENERGY_STATUS = 0x619;
        public const uint MSR_DRAM_POWER_INFO = 0x61C;

        // APERF/MPERF for frequency monitoring
        public const uint IA32_APERF = 0xE7;
        public const uint IA32_MPERF = 0xE8;
    }

    /// <summary>
    /// WinRing0 native methods for MSR access
    /// </summary>
    public static class WinRing0
    {
        public const string DllName = "WinRing0x64.dll";

        [DllImport("kernel32", SetLastError = true)]
        public static extern IntPtr LoadLibrary(string lpFileName);

        [DllImport("kernel32", SetLastError = true)]
        public static extern IntPtr GetProcAddress(IntPtr hModule, string lpProcName);

        [DllImport("kernel32", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        public static extern bool FreeLibrary(IntPtr hModule);

        // MSR Read function type
        public delegate int ReadMsrDelegate(uint cpu, uint msr, out ulong value);
        
        // MSR Write function type
        public delegate int WriteMsrDelegate(uint cpu, uint msr, ulong value);

        public static ReadMsrDelegate? Rdmsr;
        public static WriteMsrDelegate? Wrmsr;
    }

    /// <summary>
    /// CPU Information
    /// </summary>
    public class CPUInfo
    {
        public int Number { get; set; }
        public string Name { get; set; } = "";
        public uint MaxSpeed { get; set; }
        public uint CurrentSpeed { get; set; }
        public string Status { get; set; } = "";
    }

    /// <summary>
    /// MSR Register Info
    /// </summary>
    public class MSRInfo
    {
        public uint Address { get; set; }
        public string Name { get; set; } = "";
        public string Description { get; set; } = "";
        public ulong Value { get; set; }
        public bool IsReadOnly { get; set; }
    }

    /// <summary>
    /// Main application class
    /// </summary>
    public class MSRTestApp
    {
        private const string APP_TITLE = "Intel MSR Test Application (Based on MSR.pptx)";
        
        public static void Main(string[] args)
        {
            Console.WriteLine("╔════════════════════════════════════════════════════════════════╗");
            Console.WriteLine("║           Intel MSR Test Application - Power Management         ║");
            Console.WriteLine("║                    Based on MSR.pptx Content                   ║");
            Console.WriteLine("╚════════════════════════════════════════════════════════════════╝");
            Console.WriteLine();

            // Initialize WinRing0
            if (!InitializeWinRing0())
            {
                Console.WriteLine("[ERROR] Failed to initialize WinRing0 driver.");
                Console.WriteLine("        Please install WinRing0 driver for MSR access.");
                Console.WriteLine();
                Console.WriteLine("        Installation steps:");
                Console.WriteLine("        1. Copy WinRing0.sys to C:\\Windows\\System32\\drivers\\");
                Console.WriteLine("        2. Run: sc create WinRing0 type= kernel binPath= C:\\Windows\\System32\\drivers\\WinRing0.sys");
                Console.WriteLine("        3. Run: sc start WinRing0");
                Console.WriteLine();
                ShowSimulationMode();
                return;
            }

            // Get CPU info
            var cpus = GetCPUInfo();
            Console.WriteLine($"Found {cpus.Count} CPU(s)");
            Console.WriteLine();

            // Main menu loop
            bool exit = false;
            while (!exit)
            {
                ShowMenu();
                var key = Console.ReadKey(true);
                
                switch (key.Key)
                {
                    case ConsoleKey.D1:
                        ShowMSRRegisters();
                        break;
                    case ConsoleKey.D2:
                        ReadAllMSRs(0);
                        break;
                    case ConsoleKey.D3:
                        TestEnergyPerfBias();
                        break;
                    case ConsoleKey.D4:
                        TestPerfCTL();
                        break;
                    case ConsoleKey.D5:
                        TestTurboBoost();
                        break;
                    case ConsoleKey.D6:
                        TestRAPL();
                        break;
                    case ConsoleKey.D7:
                        TestClockModulation();
                        break;
                    case ConsoleKey.D8:
                        TestUncoreRatio();
                        break;
                    case ConsoleKey.D9:
                        MonitorFrequency();
                        break;
                    case ConsoleKey.Q:
                        exit = true;
                        break;
                }
            }

            Console.WriteLine("Exiting...");
        }

        private static bool InitializeWinRing0()
        {
            try
            {
                // Try to load WinRing0 DLL
                IntPtr hModule = WinRing0.LoadLibrary(WinRing0.DllName);
                if (hModule == IntPtr.Zero)
                {
                    // Try alternative paths
                    string[] paths = {
                        "WinRing0x64.dll",
                        "C:\\Windows\\System32\\WinRing0x64.dll",
                        ".\\WinRing0x64.dll"
                    };
                    
                    foreach (var path in paths)
                    {
                        hModule = WinRing0.LoadLibrary(path);
                        if (hModule != IntPtr.Zero) break;
                    }
                    
                    if (hModule == IntPtr.Zero)
                        return false;
                }

                // Get function pointers
                IntPtr rdmsrPtr = WinRing0.GetProcAddress(hModule, "Rdmsr");
                IntPtr wrmsrPtr = WinRing0.GetProcAddress(hModule, "Wrmsr");

                if (rdmsrPtr != IntPtr.Zero && wrmsrPtr != IntPtr.Zero)
                {
                    WinRing0.Rdmsr = Marshal.GetDelegateForFunctionPointer<WinRing0.ReadMsrDelegate>(rdmsrPtr);
                    WinRing0.Wrmsr = Marshal.GetDelegateForFunctionPointer<WinRing0.WriteMsrDelegate>(wrmsrPtr);
                    return true;
                }

                return false;
            }
            catch
            {
                return false;
            }
        }

        private static void ShowSimulationMode()
        {
            Console.WriteLine("[SIMULATION MODE] - Driver not available, showing expected values");
            Console.WriteLine();
            
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  MSR Registers from PPT (Expected values on real hardware):");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            
            // Simulated IA32_ENERGY_PERF_BIAS
            Console.WriteLine("[0x1B0] IA32_ENERGY_PERF_BIAS");
            Console.WriteLine("       Default value: 0x6 (Balance mode)");
            Console.WriteLine("       Bits 3:0 = Energy Performance Bias (0-15)");
            Console.WriteLine("       0 = Highest Performance, 6 = Balance, 15 = Max Energy Savings");
            Console.WriteLine();
            
            // Simulated IA32_PERF_CTL
            Console.WriteLine("[0x198] IA32_PERF_CTL");
            Console.WriteLine("       Controls P-state transitions");
            Console.WriteLine("       Bits 15:0 = Target Performance State");
            Console.WriteLine();
            
            // Simulated IA32_MISC_ENABLE
            Console.WriteLine("[0x1A0] IA32_MISC_ENABLE");
            Console.WriteLine("       Bit 38 = Turbo Boost enable/disable");
            Console.WriteLine("       Default: Turbo Boost Enabled");
            Console.WriteLine();
            
            // RAPL
            Console.WriteLine("[0x606] MSR_RAPL_POWER_UNIT");
            Console.WriteLine("       Power Units (Bits 3:0): Default = 1/8W");
            Console.WriteLine("       Energy Status (Bits 12:8): Default = 15.3 µJ");
            Console.WriteLine("       Time Units (Bits 19:16): Default = 976 µs");
            Console.WriteLine();
            
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
        }

        private static List<CPUInfo> GetCPUInfo()
        {
            var cpus = new List<CPUInfo>();
            
            // Get CPU count from system
            int cpuCount = Environment.ProcessorCount;
            
            for (int i = 0; i < cpuCount; i++)
            {
                cpus.Add(new CPUInfo
                {
                    Number = i,
                    Name = $"CPU Core {i}",
                    MaxSpeed = 4000, // Default MHz
                    CurrentSpeed = 3500,
                    Status = "Active"
                });
            }

            return cpus;
        }

        private static void ShowMenu()
        {
            Console.WriteLine();
            Console.WriteLine("╔════════════════════════════════════════════════════════════════╗");
            Console.WriteLine("║  Menu (Based on MSR.pptx Power Management Topics):            ║");
            Console.WriteLine("╠════════════════════════════════════════════════════════════════╣");
            Console.WriteLine("║  [1] Show MSR Register Definitions                             ║");
            Console.WriteLine("║  [2] Read All Monitored MSRs                                  ║");
            Console.WriteLine("║  [3] Test Energy Performance Bias (0x1B0)                     ║");
            Console.WriteLine("║  [4] Test Performance Control (0x198)                         ║");
            Console.WriteLine("║  [5] Test Turbo Boost (0x1A0)                                 ║");
            Console.WriteLine("║  [6] Test RAPL Power Limits                                    ║");
            Console.WriteLine("║  [7] Test Clock Modulation (0x19A)                             ║");
            Console.WriteLine("║  [8] Test Uncore Ratio Limit (0x620)                          ║");
            Console.WriteLine("║  [9] Monitor CPU Frequency (APERF/MPERF)                      ║");
            Console.WriteLine("║  [Q] Quit                                                      ║");
            Console.WriteLine("╚════════════════════════════════════════════════════════════════╝");
            Console.Write("Select option: ");
        }

        private static void ShowMSRRegisters()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  MSR Registers from PPT (Intel Power Management)");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            
            var registers = new[]
            {
                new { Address = "0x1B0", Name = "IA32_ENERGY_PERF_BIAS", Desc = "Energy Performance Bias Hint (0-15)" },
                new { Address = "0x198", Name = "IA32_PERF_CTL", Desc = "Performance Control (P-state target)" },
                new { Address = "0x199", Name = "IA32_PERF_STATUS", Desc = "Current P-state (read-only)" },
                new { Address = "0x1A0", Name = "IA32_MISC_ENABLE", Desc = "Turbo Boost control" },
                new { Address = "0x19A", Name = "IA32_CLOCK_MODULATION", Desc = "Clock modulation (T-state)" },
                new { Address = "0x620", Name = "MSR_UNCORE_RATIO_LIMIT", Desc = "Uncore frequency limits" },
                new { Address = "0x606", Name = "MSR_RAPL_POWER_UNIT", Desc = "RAPL units (power/energy/time)" },
                new { Address = "0x610", Name = "MSR_PKG_POWER_LIMIT", Desc = "Package power limit" },
                new { Address = "0x611", Name = "MSR_PKG_ENERGY_STATUS", Desc = "Package energy consumed" },
                new { Address = "0x638", Name = "MSR_PP0_POWER_LIMIT", Desc = "CPU power plane limit" },
                new { Address = "0x618", Name = "MSR_DRAM_POWER_LIMIT", Desc = "DRAM power limit" },
                new { Address = "0xE7", Name = "IA32_APERF", Desc = "Actual Performance Clock Counter" },
                new { Address = "0xE8", Name = "IA32_MPERF", Desc = "Maximum Performance Clock Counter" },
            };

            foreach (var reg in registers)
            {
                Console.WriteLine($"  [{reg.Address}] {reg.Name}");
                Console.WriteLine($"          {reg.Desc}");
                Console.WriteLine();
            }
        }

        private static void ReadAllMSRs(uint cpu)
        {
            Console.WriteLine();
            Console.WriteLine($"Reading MSRs on CPU {cpu}...");
            Console.WriteLine();

            uint[] registers = {
                MSRRegisters.IA32_ENERGY_PERF_BIAS,
                MSRRegisters.IA32_PERF_CTL,
                MSRRegisters.IA32_PERF_STATUS,
                MSRRegisters.IA32_MISC_ENABLE,
                MSRRegisters.MSR_RAPL_POWER_UNIT,
                MSRRegisters.MSR_PKG_POWER_LIMIT,
                MSRRegisters.MSR_PKG_ENERGY_STATUS
            };

            foreach (var reg in registers)
            {
                if (ReadMSR(cpu, reg, out ulong value))
                {
                    Console.WriteLine($"[0x{reg:X3}] = 0x{value:X16}");
                }
                else
                {
                    Console.WriteLine($"[0x{reg:X3}] = READ FAILED (need admin/driver)");
                }
            }
        }

        private static void TestEnergyPerfBias()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  IA32_ENERGY_PERF_BIAS (0x1B0) Test");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slide 5:");
            Console.WriteLine("  - Bits 3:0: Software hint for power management");
            Console.WriteLine("  - 0 = Highest Performance");
            Console.WriteLine("  - 6 = Balance (default)");
            Console.WriteLine("  - 15 = Maximum Energy Savings");
            Console.WriteLine("  - Affects Turbo boost and Uncore Frequency Scaling");
            Console.WriteLine();

            if (ReadMSR(0, MSRRegisters.IA32_ENERGY_PERF_BIAS, out ulong value))
            {
                byte bias = (byte)(value & 0xF);
                Console.WriteLine($"Current value: 0x{value:X}");
                Console.WriteLine($"Energy Perf Bias: {bias}");
                Console.WriteLine($"Mode: {(bias == 0 ? "Highest Performance" : bias == 6 ? "Balance" : bias == 15 ? "Max Energy Savings" : "Custom")}");
            }
            else
            {
                Console.WriteLine("[INFO] Simulation mode - default value would be 0x6 (Balance)");
            }

            Console.WriteLine();
            Console.WriteLine("Testing write (requires admin):");
            Console.WriteLine("  To set Maximum Performance: wrmsr 0x1B0 0x0");
            Console.WriteLine("  To set Balance mode: wrmsr 0x1B0 0x6");
            Console.WriteLine("  To set Max Energy Savings: wrmsr 0x1B0 0xF");
        }

        private static void TestPerfCTL()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  IA32_PERF_CTL (0x198) Test");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slide 6:");
            Console.WriteLine("  - Intel P-State driver writes 16-bit value for state transitions");
            Console.WriteLine("  - Controls target P-state (frequency/voltage)");
            Console.WriteLine("  - Write to set target, read current target");
            Console.WriteLine();
            Console.WriteLine("  Related registers:");
            Console.WriteLine("    0x198: IA32_PERF_CTL - Target P-state");
            Console.WriteLine("    0x199: IA32_PERF_STATUS - Current actual P-state");

            if (ReadMSR(0, MSRRegisters.IA32_PERF_CTL, out ulong perfCtl))
            {
                Console.WriteLine();
                Console.WriteLine($"IA32_PERF_CTL: 0x{perfCtl:X}");
                Console.WriteLine($"  Target P-state ratio: {(perfCtl & 0xFF)}");
            }

            if (ReadMSR(0, MSRRegisters.IA32_PERF_STATUS, out ulong perfStatus))
            {
                Console.WriteLine($"IA32_PERF_STATUS: 0x{perfStatus:X}");
            }
        }

        private static void TestTurboBoost()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  Turbo Boost Control (0x1A0) Test");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slide 6:");
            Console.WriteLine("  - IA32_MISC_ENABLE (0x1A0) controls Turbo Boost");
            Console.WriteLine("  - Bit 38: Turbo Boost enable/disable");
            Console.WriteLine();
            Console.WriteLine("WARNING: Disabling Turbo Boost may significantly impact performance!");
            Console.WriteLine();

            if (ReadMSR(0, MSRRegisters.IA32_MISC_ENABLE, out ulong value))
            {
                bool turboEnabled = (value & (1UL << 38)) == 0;
                Console.WriteLine($"IA32_MISC_ENABLE: 0x{value:X16}");
                Console.WriteLine($"Turbo Boost: {(turboEnabled ? "ENABLED" : "DISABLED")}");
            }
            else
            {
                Console.WriteLine("[INFO] Default: Turbo Boost would be ENABLED");
            }
        }

        private static void TestRAPL()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  RAPL (Running Average Power Limit) Test");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slides 54-58:");
            Console.WriteLine("  RAPL Domains:");
            Console.WriteLine("    - PKG: Entire package (socket)");
            Console.WriteLine("    - PP0: CPU cores power plane");
            Console.WriteLine("    - PP1: Uncore/Graphics (if present)");
            Console.WriteLine("    - DRAM: Memory (server only)");
            Console.WriteLine();
            Console.WriteLine("Key Registers:");
            Console.WriteLine("  0x606: MSR_RAPL_POWER_UNIT - Units for power/energy/time");
            Console.WriteLine("  0x610: MSR_PKG_POWER_LIMIT - Package power limit");
            Console.WriteLine("  0x611: MSR_PKG_ENERGY_STATUS - Energy consumed");
            Console.WriteLine("  0x638: MSR_PP0_POWER_LIMIT - CPU power limit");
            Console.WriteLine("  0x618: MSR_DRAM_POWER_LIMIT - DRAM power limit");
            Console.WriteLine();

            if (ReadMSR(0, MSRRegisters.MSR_RAPL_POWER_UNIT, out ulong raplUnit))
            {
                Console.WriteLine($"RAPL_POWER_UNIT: 0x{raplUnit:X}");
                byte powerUnit = (byte)(raplUnit & 0xF);
                byte energyUnit = (byte)((raplUnit >> 8) & 0x1F);
                byte timeUnit = (byte)((raplUnit >> 16) & 0xF);
                Console.WriteLine($"  Power Unit: 1/2^{powerUnit} Watts = 1/{(1 << powerUnit) / 8.0:F2}W");
                Console.WriteLine($"  Energy Unit: 1/2^{energyUnit} Joules");
                Console.WriteLine($"  Time Unit: 1/2^{timeUnit} Seconds");
            }

            Console.WriteLine();
            Console.WriteLine("Package Power Limit (0x610):");
            if (ReadMSR(0, MSRRegisters.MSR_PKG_POWER_LIMIT, out ulong pkgLimit))
            {
                Console.WriteLine($"  Value: 0x{pkgLimit:X16}");
                Console.WriteLine("  Bits [14:0]: Power Limit 1 (in power units)");
                Console.WriteLine("  Bits [30:16]: Time Window 1");
                Console.WriteLine("  Bit 31: Enable 1");
            }

            Console.WriteLine();
            Console.WriteLine("Package Energy Status (0x611):");
            if (ReadMSR(0, MSRRegisters.MSR_PKG_ENERGY_STATUS, out ulong energy))
            {
                Console.WriteLine($"  Total Energy: 0x{energy:X16}");
                Console.WriteLine("  (Read-only, in energy units since last clear)");
            }
        }

        private static void TestClockModulation()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  IA32_CLOCK_MODULATION (0x19A) Test");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slide 18:");
            Console.WriteLine("  - DDCM: Dynamic Duty Cycle Modulation (T-state)");
            Console.WriteLine("  - Used for power management and thermal control");
            Console.WriteLine();
            Console.WriteLine("Bits:");
            Console.WriteLine("  Bit 4: Enable/Disable");
            Console.WriteLine("  Bits 3:0: Duty Cycle");
            Console.WriteLine("    0x0 = No modulation");
            Console.WriteLine("    0x1-0xE = 6.25% increments");
            Console.WriteLine("    0xF = Reserved");
            Console.WriteLine();
            Console.WriteLine("Example: 0x9 = 50% duty cycle enabled");

            if (ReadMSR(0, MSRRegisters.IA32_CLOCK_MODULATION, out ulong value))
            {
                Console.WriteLine();
                Console.WriteLine($"Current value: 0x{value:X}");
                bool enabled = (value & 0x10) != 0;
                byte duty = (byte)(value & 0xF);
                Console.WriteLine($"Status: {(enabled ? "ENABLED" : "DISABLED")}");
                if (enabled && duty > 0)
                {
                    double percent = duty * 6.25;
                    Console.WriteLine($"Duty Cycle: {percent}%");
                }
            }
        }

        private static void TestUncoreRatio()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  MSR_UNCORE_RATIO_LIMIT (0x620) Test");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slide 10:");
            Console.WriteLine("  - Controls uncore frequency scaling");
            Console.WriteLine("  - Uncore = cache, memory controller, etc.");
            Console.WriteLine();
            Console.WriteLine("Bits:");
            Console.WriteLine("  Bits 14:8: Maximum ratio");
            Console.WriteLine("  Bits 6:0: Minimum ratio");
            Console.WriteLine();
            Console.WriteLine("Tip: If max = min, uncore frequency is locked!");

            if (ReadMSR(0, MSRRegisters.MSR_UNCORE_RATIO_LIMIT, out ulong value))
            {
                Console.WriteLine();
                Console.WriteLine($"Current value: 0x{value:X}");
                byte maxRatio = (byte)((value >> 8) & 0x7F);
                byte minRatio = (byte)(value & 0x7F);
                Console.WriteLine($"Max Ratio: {maxRatio} ({(maxRatio * 100)} MHz bus)");
                Console.WriteLine($"Min Ratio: {minRatio}");
                
                if (maxRatio == minRatio)
                    Console.WriteLine("Uncore frequency is LOCKED (max = min)");
            }
        }

        private static void MonitorFrequency()
        {
            Console.WriteLine();
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine("  CPU Frequency Monitor (APERF/MPERF)");
            Console.WriteLine("═══════════════════════════════════════════════════════════════════");
            Console.WriteLine();
            Console.WriteLine("From PPT Slide 7:");
            Console.WriteLine("  - IA32_APERF (0xE7): Actual clock cycles in C0");
            Console.WriteLine("  - IA32_MPERF (0xE8): Maximum cycles in C0");
            Console.WriteLine();
            Console.WriteLine("Formula: Current Freq = NominalFreq × (ΔAPERF / ΔMPERF)");
            Console.WriteLine();
            Console.WriteLine("Monitoring for 3 seconds...");
            Console.WriteLine();

            for (int i = 0; i < 3; i++)
            {
                if (ReadMSR(0, MSRRegisters.IA32_APERF, out ulong aperf) &&
                    ReadMSR(0, MSRRegisters.IA32_MPERF, out ulong mperf))
                {
                    Console.WriteLine($"[{DateTime.Now:HH:mm:ss}] APERF=0x{aperf:X}, MPERF=0x{mperf:X}");
                }
                else
                {
                    Console.WriteLine($"[{DateTime.Now:HH:mm:ss}] Simulation - APERF/MPERF would show current frequency");
                }
                System.Threading.Thread.Sleep(1000);
            }
        }

        private static bool ReadMSR(uint cpu, uint msr, out ulong value)
        {
            value = 0;
            
            if (WinRing0.Rdmsr == null)
                return false;

            try
            {
                int result = WinRing0.Rdmsr(cpu, msr, out value);
                return result == 0;
            }
            catch
            {
                return false;
            }
        }

        private static bool WriteMSR(uint cpu, uint msr, ulong value)
        {
            if (WinRing0.Wrmsr == null)
                return false;

            try
            {
                int result = WinRing0.Wrmsr(cpu, msr, value);
                return result == 0;
            }
            catch
            {
                return false;
            }
        }
    }
}
