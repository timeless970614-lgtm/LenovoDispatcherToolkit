/*
 * MSRApi.cs - User-Mode API for MSR Access
 * =========================================
 * Based on: C:\Users\Zhou\OneDrive - Lenovo\MSR.pptx
 * =========================================
 * 
 * This class provides a managed wrapper for MSR access through either:
 * 1. WinRing0 driver (if installed)
 * 2. Our custom MSRDriver (if installed)
 * 3. Simulation mode (read-only, for testing)
 */

using System;
using System.Runtime.InteropServices;
using System.Text;

namespace MSR_Test_App
{
    /// <summary>
    /// MSR Access API wrapper
    /// </summary>
    public class MSRApi : IDisposable
    {
        private IntPtr _driverHandle = IntPtr.Zero;
        private bool _isSimulationMode = true;
        private bool _disposed = false;

        // WinRing0 DLL imports
        private const string WinRing0Dll = "WinRing0x64.dll";

        [DllImport("kernel32", SetLastError = true)]
        private static extern IntPtr LoadLibrary(string lpFileName);

        [DllImport("kernel32", SetLastError = true)]
        private static extern IntPtr GetProcAddress(IntPtr hModule, string lpProcName);

        [DllImport("kernel32", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool FreeLibrary(IntPtr hModule);

        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        private delegate int RdmsrDelegate(uint cpu, uint msr, out ulong value);

        [UnmanagedFunctionPointer(CallingConvention.Cdecl)]
        private delegate int WrmsrDelegate(uint cpu, uint msr, ulong value);

        private RdmsrDelegate? _rdmsr;
        private WrmsrDelegate? _wrmsr;

        /// <summary>
        /// Initialize MSR API
        /// </summary>
        /// <returns>True if driver is available, false if simulation mode</returns>
        public bool Initialize()
        {
            // Try WinRing0 first
            if (TryLoadWinRing0())
            {
                _isSimulationMode = false;
                Console.WriteLine("[OK] WinRing0 driver loaded");
                return true;
            }

            // Try our custom driver
            if (TryLoadCustomDriver())
            {
                _isSimulationMode = false;
                Console.WriteLine("[OK] MSRDriver loaded");
                return true;
            }

            // Simulation mode
            _isSimulationMode = true;
            Console.WriteLine("[INFO] Running in simulation mode (no driver)");
            Console.WriteLine("       Install WinRing0 or MSRDriver for real MSR access");
            return false;
        }

        private bool TryLoadWinRing0()
        {
            try
            {
                IntPtr hModule = LoadLibrary(WinRing0Dll);
                if (hModule == IntPtr.Zero)
                {
                    // Try system paths
                    string[] paths = {
                        "WinRing0x64.dll",
                        "C:\\Windows\\System32\\WinRing0x64.dll",
                        "C:\\Windows\\SysWOW64\\WinRing0x64.dll",
                        ".\\WinRing0x64.dll"
                    };

                    foreach (var path in paths)
                    {
                        hModule = LoadLibrary(path);
                        if (hModule != IntPtr.Zero) break;
                    }
                }

                if (hModule == IntPtr.Zero) return false;

                IntPtr rdmsrPtr = GetProcAddress(hModule, "Rdmsr");
                IntPtr wrmsrPtr = GetProcAddress(hModule, "Wrmsr");

                if (rdmsrPtr != IntPtr.Zero && wrmsrPtr != IntPtr.Zero)
                {
                    _rdmsr = Marshal.GetDelegateForFunctionPointer<RdmsrDelegate>(rdmsrPtr);
                    _wrmsr = Marshal.GetDelegateForFunctionPointer<WrmsrDelegate>(wrmsrPtr);
                    return true;
                }

                FreeLibrary(hModule);
                return false;
            }
            catch
            {
                return false;
            }
        }

        private bool TryLoadCustomDriver()
        {
            try
            {
                _driverHandle = CreateFile(
                    "\\\\.\\MSRDriver",
                    GENERIC_READ | GENERIC_WRITE,
                    0,
                    IntPtr.Zero,
                    OPEN_EXISTING,
                    FILE_ATTRIBUTE_NORMAL,
                    IntPtr.Zero
                );

                if (_driverHandle == (IntPtr)(-1))
                {
                    _driverHandle = IntPtr.Zero;
                    return false;
                }

                return true;
            }
            catch
            {
                return false;
            }
        }

        /// <summary>
        /// Read MSR value
        /// </summary>
        public bool ReadMSR(uint cpu, uint address, out ulong value)
        {
            value = 0;

            if (_isSimulationMode)
            {
                // Return simulated values based on PPT documentation
                value = GetSimulatedValue(address);
                return true;
            }

            if (_rdmsr != null)
            {
                int result = _rdmsr(cpu, address, out value);
                return result == 0;
            }

            return false;
        }

        /// <summary>
        /// Write MSR value
        /// </summary>
        public bool WriteMSR(uint cpu, uint address, ulong value)
        {
            if (_isSimulationMode)
            {
                Console.WriteLine($"[SIM] Would write 0x{value:X16} to MSR 0x{address:X3}");
                return true;
            }

            if (_wrmsr != null)
            {
                int result = _wrmsr(cpu, address, value);
                return result == 0;
            }

            return false;
        }

        /// <summary>
        /// Get simulated MSR value for testing
        /// </summary>
        private ulong GetSimulatedValue(uint address)
        {
            return address switch
            {
                // IA32_ENERGY_PERF_BIAS - Default is 0x6 (Balance)
                0x1B0 => 0x6,

                // IA32_PERF_CTL - Example value
                0x198 => 0x1E00, // Target ratio 30

                // IA32_PERF_STATUS - Current state
                0x199 => 0x1D00, // Current ratio 29

                // IA32_MISC_ENABLE - Turbo enabled (bit 38 = 0)
                0x1A0 => 0x850089,

                // IA32_CLOCK_MODULATION - Disabled
                0x19A => 0x0,

                // RAPL_POWER_UNIT - Default values
                0x606 => 0xE1E02A10,

                // Other MSRs return 0
                _ => 0x0
            };
        }

        /// <summary>
        /// Check if running in simulation mode
        /// </summary>
        public bool IsSimulationMode => _isSimulationMode;

        /// <summary>
        /// Read Energy Performance Bias
        /// </summary>
        public byte ReadEnergyPerfBias(uint cpu)
        {
            if (ReadMSR(cpu, 0x1B0, out ulong value))
            {
                return (byte)(value & 0xF);
            }
            return 6; // Default balance
        }

        /// <summary>
        /// Write Energy Performance Bias
        /// </summary>
        public bool WriteEnergyPerfBias(uint cpu, byte bias)
        {
            if (bias > 15) return false;
            return WriteMSR(cpu, 0x1B0, bias);
        }

        /// <summary>
        /// Get Turbo Boost status
        /// </summary>
        public bool GetTurboBoostEnabled(uint cpu)
        {
            if (ReadMSR(cpu, 0x1A0, out ulong value))
            {
                // Bit 38 = Turbo Boost enable
                return (value & (1UL << 38)) == 0;
            }
            return true; // Assume enabled by default
        }

        /// <summary>
        /// Set Turbo Boost status
        /// </summary>
        public bool SetTurboBoostEnabled(uint cpu, bool enable)
        {
            if (ReadMSR(cpu, 0x1A0, out ulong value))
            {
                if (enable)
                    value &= ~(1UL << 38); // Clear bit to enable
                else
                    value |= (1UL << 38);  // Set bit to disable

                return WriteMSR(cpu, 0x1A0, value);
            }
            return false;
        }

        /// <summary>
        /// Read RAPL power units
        /// </summary>
        public (double Power, double Energy, double Time) GetRAPLUnits(uint cpu)
        {
            if (ReadMSR(cpu, 0x606, out ulong value))
            {
                byte powerUnit = (byte)(value & 0xF);
                byte energyUnit = (byte)((value >> 8) & 0x1F);
                byte timeUnit = (byte)((value >> 16) & 0xF);

                return (
                    Math.Pow(2, powerUnit) / 8.0,  // Power unit in Watts
                    Math.Pow(2, energyUnit) / 1e6, // Energy unit in Joules
                    Math.Pow(2, timeUnit) / 1e6     // Time unit in Seconds
                );
            }

            // Default values
            return (0.125, 15.3e-6, 976e-6);
        }

        /// <summary>
        /// Read Package Power Limit
        /// </summary>
        public (uint LimitMw, uint TimeWindowUs) GetPackagePowerLimit(uint cpu)
        {
            if (ReadMSR(cpu, 0x610, out ulong value))
            {
                byte powerUnit = (byte)(ReadMSR(cpu, 0x606, out ulong unitVal) ? (unitVal & 0xF) : 4);
                byte timeUnit = (byte)(ReadMSR(cpu, 0x606, out unitVal) ? ((unitVal >> 16) & 0xF) : 10);

                uint limit = (uint)(value & 0x7FFF);
                uint window = (uint)((value >> 17) & 0x7F);

                double power = limit * Math.Pow(2, powerUnit) / 8.0 * 1000; // mW
                double time = window * Math.Pow(2, timeUnit) / 1e6; // us

                return ((uint)power, (uint)time);
            }

            return (0, 0);
        }

        /// <summary>
        /// Calculate current frequency from APERF/MPERF
        /// </summary>
        public uint CalculateCurrentFrequency(uint cpu, uint nominalFreqMhz)
        {
            if (!ReadMSR(cpu, 0xE7, out ulong aperf) || !ReadMSR(cpu, 0xE8, out ulong mperf))
                return 0;

            if (mperf == 0) return 0;

            double ratio = (double)aperf / mperf;
            return (uint)(nominalFreqMhz * ratio);
        }

        /// <summary>
        /// Dispose resources
        /// </summary>
        public void Dispose()
        {
            Dispose(true);
            GC.SuppressFinalize(this);
        }

        protected virtual void Dispose(bool disposing)
        {
            if (!_disposed)
            {
                if (_driverHandle != IntPtr.Zero)
                {
                    CloseHandle(_driverHandle);
                    _driverHandle = IntPtr.Zero;
                }
                _disposed = true;
            }
        }

        // P/Invoke declarations
        [DllImport("kernel32.dll", SetLastError = true)]
        private static extern IntPtr CreateFile(
            string lpFileName,
            uint dwDesiredAccess,
            uint dwShareMode,
            IntPtr lpSecurityAttributes,
            uint dwCreationDisposition,
            uint dwFlagsAndAttributes,
            IntPtr hTemplateFile);

        [DllImport("kernel32.dll", SetLastError = true)]
        [return: MarshalAs(UnmanagedType.Bool)]
        private static extern bool CloseHandle(IntPtr hObject);

        private const uint GENERIC_READ = 0x80000000;
        private const uint GENERIC_WRITE = 0x40000000;
        private const uint OPEN_EXISTING = 3;
        private const uint FILE_ATTRIBUTE_NORMAL = 0x80;
    }
}
