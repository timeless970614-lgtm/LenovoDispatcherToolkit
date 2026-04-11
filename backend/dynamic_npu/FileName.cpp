///*
// * A consolidated example demonstrating the `hm_sys_get_*` family of
// * interfaces from `hm_sys.h`.
// *
// * The example groups APIs into logical categories:
// *  - Device enumeration & properties
// *  - Runtime metrics (utilization, memory, temperature, frequency)
// *  - PCIe / bus info and power/bandwidth
// *  - SDK / driver / firmware info and auxiliary getters
// *
// * Each group is implemented as a small helper function. The main()
// * enumerates devices and calls each helper to print results.
// *
// * Comments are in English and error handling is conservative: missing
// * values are reported but do not abort the whole program.
// */
//
//#include <stdio.h>
//#include <string.h>
//#include <stdint.h>
//#include "../Project1/hm_sys.h"
//#include <filesystem>
//#include <iostream>
//#include <windows.h>
//#include <string>
//#include <vector>
//#include <cstdint>
//#include <iomanip>
//#include <cstring>
//typedef int (*HAL_INIT_FUNC)();
//typedef int (*HAL_DEINIT_FUNC)();
//typedef int (*HAL_GET_VERSION_FUNC)(char* version, int size);
//typedef int (*HAL_GET_DEVICE_INFO_FUNC)(void* info);
//
//#define GET_FUNCTION(handle, func_name, func_ptr, func_type) \
//    func_ptr = (func_type)GetProcAddress(handle, func_name); \
//    if (!func_ptr) { \
//        std::cerr << "错误: 无法获取函数 " << func_name << std::endl; \
//    }
//
//class HoumoDriver 
//{
//private:
//    HMODULE m_hDll;
//
//    // 函数指针成员变量
//    HAL_INIT_FUNC m_HalInit;
//    HAL_DEINIT_FUNC m_HalDeinit;
//    HAL_GET_VERSION_FUNC m_HalGetVersion;
//    HAL_GET_DEVICE_INFO_FUNC m_HalGetDeviceInfo;
//    // 添加更多函数指针...
//
//public:
//    HoumoDriver() : m_hDll(nullptr), m_HalInit(nullptr),
//        m_HalDeinit(nullptr), m_HalGetVersion(nullptr),
//        m_HalGetDeviceInfo(nullptr) {
//    }
//
//    ~HoumoDriver() {
//        Unload();
//    }
//
//    bool Load(const std::string& dllPath) {
//        // 卸载之前加载的DLL
//        Unload();
//
//        std::cout << "Loading DLL: " << dllPath << std::endl;
//
//        // 加载DLL
//        m_hDll = LoadLibraryA(dllPath.c_str());
//        if (!m_hDll) {
//            DWORD error = GetLastError();
//            std::cerr << "load DLL fail,ErrorCode: " << error << std::endl;
//
//            // 获取错误描述
//            LPSTR errorMsg = nullptr;
//            FormatMessageA(
//                FORMAT_MESSAGE_ALLOCATE_BUFFER |
//                FORMAT_MESSAGE_FROM_SYSTEM |
//                FORMAT_MESSAGE_IGNORE_INSERTS,
//                nullptr,
//                error,
//                MAKELANGID(LANG_NEUTRAL, SUBLANG_DEFAULT),
//                (LPSTR)&errorMsg,
//                0, nullptr);
//
//            if (errorMsg) {
//                std::cerr << "Error description: " << errorMsg << std::endl;
//                LocalFree(errorMsg);
//            }
//
//            return false;
//        }
//
//        std::cout << "DLL loaded successfully!" << std::endl;
//
//        // 获取函数地址
//        // 注意：函数名需要根据实际DLL导出名称来调整
//
//        // 尝试不同的函数名格式
//        m_HalInit = (HAL_INIT_FUNC)GetProcAddress(m_hDll, "hal_init");
//        if (!m_HalInit) {
//            m_HalInit = (HAL_INIT_FUNC)GetProcAddress(m_hDll, "HalInit");
//        }
//        if (!m_HalInit) {
//            m_HalInit = (HAL_INIT_FUNC)GetProcAddress(m_hDll, "HAL_Init");
//        }
//
//        m_HalDeinit = (HAL_DEINIT_FUNC)GetProcAddress(m_hDll, "hal_deinit");
//        if (!m_HalDeinit) {
//            m_HalDeinit = (HAL_DEINIT_FUNC)GetProcAddress(m_hDll, "HalDeinit");
//        }
//        if (!m_HalDeinit) {
//            m_HalDeinit = (HAL_DEINIT_FUNC)GetProcAddress(m_hDll, "HAL_Deinit");
//        }
//
//        m_HalGetVersion = (HAL_GET_VERSION_FUNC)GetProcAddress(m_hDll, "hal_get_version");
//        if (!m_HalGetVersion) {
//            m_HalGetVersion = (HAL_GET_VERSION_FUNC)GetProcAddress(m_hDll, "HalGetVersion");
//        }
//        if (!m_HalGetVersion) {
//            m_HalGetVersion = (HAL_GET_VERSION_FUNC)GetProcAddress(m_hDll, "HAL_GetVersion");
//        }
//
//        // 尝试获取更多函数...
//        m_HalGetDeviceInfo = (HAL_GET_DEVICE_INFO_FUNC)GetProcAddress(m_hDll, "hal_get_device_info");
//
//        return true;
//    }
//
//    void Unload() {
//        if (m_hDll) {
//            std::cout << "Uninstall DLL..." << std::endl;
//            FreeLibrary(m_hDll);
//            m_hDll = nullptr;
//
//            // 清空函数指针
//            m_HalInit = nullptr;
//            m_HalDeinit = nullptr;
//            m_HalGetVersion = nullptr;
//            m_HalGetDeviceInfo = nullptr;
//        }
//    }
//
//    bool IsLoaded() const {
//        return m_hDll != nullptr;
//    }
//
//    // 包装的函数调用
//    int Init() {
//        if (m_HalInit) {
//            std::cout << "Call hal_init()..." << std::endl;
//            return m_HalInit();
//        }
//        std::cerr << "Error:hal_init not found" << std::endl;
//        return -1;
//    }
//
//    int Deinit() {
//        if (m_HalDeinit) {
//            std::cout << "call hal_deinit()..." << std::endl;
//            return m_HalDeinit();
//        }
//        std::cerr << "error: hal_deinit not found" << std::endl;
//        return -1;
//    }
//
//    std::string GetVersion() {
//        if (m_HalGetVersion) {
//            std::cout << "call hal_get_version()..." << std::endl;
//            char version[256] = { 0 };
//            if (m_HalGetVersion(version, sizeof(version)) == 0) {
//                return std::string(version);
//            }
//        }
//        else {
//            std::cerr << "error: hal_get_version not found" << std::endl;
//        }
//        return "Unknown version";
//    }
//
//    int GetDeviceInfo(void* info) {
//        if (m_HalGetDeviceInfo) {
//            std::cout << "call hal_get_device_info()..." << std::endl;
//            return m_HalGetDeviceInfo(info);
//        }
//        std::cerr << "Error: hal_get_device_info not found" << std::endl;
//        return -1;
//    }
//
//    // 显示DLL中所有导出的函数
//    /*void ListExportedFunctions() 
//    {
//        if (!m_hDll) {
//            std::cerr << "DLL not load" << std::endl;
//            return;
//        }
//
//        std::cout << "\n=== DLL Export Function List ===" << std::endl;
//        const char* functionNames[] = {
//            "hal_init", "HalInit", "HAL_Init",
//            "hal_deinit", "HalDeinit", "HAL_Deinit",
//            "hal_get_version", "HalGetVersion", "HAL_GetVersion",
//            "hal_get_device_info", "HalGetDeviceInfo", "HAL_GetDeviceInfo",
//            "hal_open", "hal_close", "hal_read", "hal_write",
//            "hal_ioctl", "hal_mmap", "hal_munmap"
//        };
//
//        for (const auto& funcName : functionNames) {
//            FARPROC funcAddr = GetProcAddress(m_hDll, funcName);
//            if (funcAddr) {
//                std::cout << "Find the function:" << funcName
//                    << " (Address: 0x" << std::hex << (size_t)funcAddr
//                    << std::dec << ")" << std::endl;
//            }
//        }
//        std::cout << "=== End of list ===\n" << std::endl;
//    }*/
//};
//
//bool FileExists(const std::string& filePath) 
//{
//    DWORD fileAttributes = GetFileAttributesA(filePath.c_str());
//    return (fileAttributes != INVALID_FILE_ATTRIBUTES &&
//        !(fileAttributes & FILE_ATTRIBUTE_DIRECTORY));
//}
//
///* Print basic device properties: vendor id, serial, name, TOPS, core count */
//static void print_device_properties(int dev)
//{
//    std::cout << "[dev " << dev << "] Basic Properties:" << std::endl;
//    int vendor = hm_sys_get_vendor_id(dev);
//    if (vendor < 0)
//        std::cout << "  vendor id: <unavailable>" << std::endl;
//    else
//        std::cout << "  vendor id: 0x" << std::hex << vendor << std::dec << std::endl;
//
//    char sn[HM_SYS_DEVICE_SN_LEN] = { 0 };
//    if (hm_sys_get_device_sn(dev, sn, sizeof(sn)) < 0)
//        std::cout << "  serial: <unavailable>" << std::endl;
//    else
//        std::cout << "  serial: " << sn << std::endl;
//
//    char name[HM_SYS_DEVICE_NAME_LEN] = { 0 };
//    if (hm_sys_get_device_name(dev, name, sizeof(name)) < 0)
//        std::cout << "  name: <unavailable>" << std::endl;
//    else
//        std::cout << "  name: " << name << std::endl;
//
//    int tops = hm_sys_get_computing_power(dev);
//    if (tops < 0)
//        std::cout << "  computing power: <unavailable>" << std::endl;
//    else
//        std::cout << "  computing power: " << tops << " TOPS" << std::endl;
//
//    int cores = hm_sys_get_core_count(dev);
//    if (cores < 0)
//        std::cout << "  core count: <unavailable>" << std::endl;
//    else
//        std::cout << "  core count: " << cores << std::endl;
//}
//
///*
// * Print runtime metrics: core utilization, per-core/core-frequency
// * (per-core frequency API is available too), memory and temperature.
// */
//static void print_device_metrics(int dev)
//{
//    std::cout << "[dev " << dev << "] Runtime Metrics:" << std::endl;
//
//    float util = hm_sys_get_ipu_utili_rate(dev);
//    if (util < 0.0f)
//        std::cout << "  core util rate: <unavailable>" << std::endl;
//    else
//        std::cout << "  core util rate: " << std::fixed << std::setprecision(2) << util << std::endl;
//
//    /* Demonstrate per-core util (if device exposes cores) */
//    int core_count = hm_sys_get_core_count(dev);
//    if (core_count > 0 && core_count <= 2) {
//        for (int c = 0; c < core_count; ++c) {
//            float core_util = hm_sys_get_ipu_core_utili_rate(dev, static_cast<uint32_t>(c));
//            if (core_util >= 0.0f)
//                std::cout << "    core[" << c << "] util: " << std::fixed << std::setprecision(2) << core_util << std::endl;
//            else
//                std::cout << "    core[" << c << "] util: <unavailable>" << std::endl;
//        }
//    }
//
//    uint64_t freq = 0;
//    if (hm_sys_get_ipu_frequency(dev, &freq) == 0)
//        std::cout << "  avg ipu frequency: " << freq << " Hz" << std::endl;
//    else
//        std::cout << "  avg ipu frequency: <unavailable>" << std::endl;
//
//    float voltage = 0.0f;
//    if (hm_sys_get_ipu_voltage(dev, &voltage) == 0)
//        std::cout << "  ipu voltage: " << std::fixed << std::setprecision(2) << voltage << " mV" << std::endl;
//    else
//        std::cout << "  ipu voltage: <unavailable>" << std::endl;
//
//    struct hm_mem_info mem = { 0 };
//    if (hm_sys_get_mem_info(dev, &mem) == 0)
//        std::cout << "  memory: total " << mem.mem_total << " MB, used "
//        << mem.mem_used << " MB, avail " << mem.mem_avail << " MB" << std::endl;
//    else
//        std::cout << "  memory: <unavailable>" << std::endl;
//
//    float temp = 0.0f;
//    if (hm_sys_get_temperature(dev, &temp) == 0)
//        std::cout << "  temperature: " << std::fixed << std::setprecision(2) << temp << " C" << std::endl;
//    else
//        std::cout << "  temperature: <unavailable>" << std::endl;
//}
//
///* Print PCIe / board info: BDF, bandwidth and board power */
//static void print_pcie_and_power_info(int dev)
//{
//    std::cout << "[dev " << dev << "] PCIe / Board Info:" << std::endl;
//
//    char bdf[64] = { 0 };
//    if (hm_sys_get_bdf(dev, bdf, sizeof(bdf)) == 0)
//        std::cout << "  PCIe BDF: " << bdf << std::endl;
//    else
//        std::cout << "  PCIe BDF: <unavailable>" << std::endl;
//
//    char bandwidth[64] = { 0 };
//    if (hm_sys_get_bandwidth(dev, bandwidth, sizeof(bandwidth)) == 0)
//        std::cout << "  PCIe bandwidth: " << bandwidth << std::endl;
//    else
//        std::cout << "  PCIe bandwidth: <unavailable>" << std::endl;
//
//    float power = 0.0f;
//    if (hm_sys_get_board_power(dev, &power) == 0)
//        std::cout << "  board power: " << std::fixed << std::setprecision(2) << power << " W" << std::endl;
//    else
//        std::cout << "  board power: <unavailable>" << std::endl;
//}
//
///* Print SDK / driver / firmware information */
//static void print_sdk_and_firmware_info(int dev)
//{
//    std::cout << "[dev " << dev << "] SDK and Firmware Info:" << std::endl;
//
//    char buf[128] = { 0 };
//    if (hm_sys_get_buildtime(buf, sizeof(buf)) == 0)
//        std::cout << "  SDK build time: " << buf << std::endl;
//    else
//        std::cout << "  SDK build time: <unavailable>" << std::endl;
//
//    std::memset(buf, 0, sizeof(buf));
//    if (hm_sys_get_version(buf, sizeof(buf)) == 0)
//        std::cout << "  SDK version: " << buf << std::endl;
//    else
//        std::cout << "  SDK version: <unavailable>" << std::endl;
//
//    std::memset(buf, 0, sizeof(buf));
//    if (hm_sys_get_driver_version(buf, sizeof(buf)) == 0)
//        std::cout << "  driver version: " << buf << std::endl;
//    else
//        std::cout << "  driver version: <unavailable>" << std::endl;
//
//    std::memset(buf, 0, sizeof(buf));
//    if (hm_sys_get_device_version(dev, buf, sizeof(buf)) == 0)
//        std::cout << "  device firmware version: " << buf << std::endl;
//    else
//        std::cout << "  device firmware version: <unavailable>" << std::endl;
//
//    uint64_t ddr = 0;
//    if (hm_sys_get_ddr_size(dev, &ddr) == 0)
//        std::cout << "  DDR size: " << ddr << " bytes" << std::endl;
//    else
//        std::cout << "  DDR size: <unavailable>" << std::endl;
//
//    /* DVFS mode */
//    enum hm_dvfs_mode mode;
//    if (hm_sys_get_dvfs_mode(dev, &mode) == 0)
//        std::cout << "  DVFS mode: " << static_cast<int>(mode) << std::endl;
//    else
//        std::cout << "  DVFS mode: <unavailable>" << std::endl;
//}
//
////int main(int argc, char** argv)
////{
////    std::cout << "=== Houmo XH2A Driver Test Program ===\n" << std::endl;
////    // DLL path
////    std::string dllPath = R"(C:\Program Files (x86)\houmo-drv-xh2_v0.7.0\hal\lib\libhal_xh2a.dll)";
////
////    // Check if DLL exists
////    if (!FileExists(dllPath))
////    {
////        std::cerr << "Error: DLL file does not exist!" << std::endl;
////        std::cerr << "Path: " << dllPath << std::endl;
////        // Try other possible paths
////        std::vector<std::string> possiblePaths = {
////            R"(C:\Program Files\houmo-drv-xh2_v0.7.0\hal\lib\libhal_xh2a.dll)",
////            R"(.\libhal_xh2a.dll)",
////            R"(libhal_xh2a.dll)",
////            R"(C:\houmo-drv-xh2_v0.7.0\hal\lib\libhal_xh2a.dll)"
////        };
////
////        for (const auto& path : possiblePaths) 
////        {
////            if (FileExists(path)) {
////                std::cout << "Found DLL: " << path << std::endl;
////                dllPath = path;
////                break;
////            }
////        }
////
////        if (!FileExists(dllPath))
////        {
////            std::cout << "\nPlease follow these steps:" << std::endl;
////            std::cout << "1. Copy libhal_xh2a.dll to one of the following locations:" << std::endl;
////            std::cout << "   - Current directory (" << std::filesystem::current_path().string() << ")" << std::endl;
////            std::cout << "   - C:\\Windows\\System32\\" << std::endl;
////            std::cout << "   - C:\\Windows\\SysWOW64\\ (for 32-bit applications)" << std::endl;
////            std::cout << "2. Or modify the dllPath variable in the code" << std::endl;
////
////            std::cout << "\nPress Enter to exit...";
////            std::cin.get();
////            return -1;
////        }
////    }
////
////    std::cout << "Found DLL file: " << dllPath << std::endl;
////    // Create driver instance
////    HoumoDriver driver;
////
////    // Load DLL
////    if (!driver.Load(dllPath)) 
////    {
////        std::cerr << "Failed to load driver!" << std::endl;
////        std::cout << "\nPress Enter to exit...";
////        std::cin.get();
////        return -1;
////    }
////    std::cout << "\n=== Testing Basic Functions ===" << std::endl;
////
////    (void)argc;(void)argv;
////    struct hm_device_info info = { 0 };
////    uint32_t ret = hm_sys_get_device_info(&info);
////    /* prefer info.num_devices if filled, otherwise use return value */
////    uint32_t dev_count = info.num_devices ? info.num_devices : ret;
////    if (dev_count == 0) 
////    {
////        std::cerr << "No devices found" << std::endl;
////        return -1;
////    }
////    std::cout << "Found " << dev_count << " device(s)" << std::endl;
////    // Convert C-style array to C++ vector for safer iteration
////    std::vector<int> device_ids;
////    for (uint32_t i = 0; i < dev_count && i < sizeof(info.device_ids) / sizeof(info.device_ids[0]); ++i) 
////    {
////        device_ids.push_back(info.device_ids[i]);
////    }
////
////    for (size_t i = 0; i < device_ids.size(); ++i) 
////    {
////        int dev = device_ids[i];
////        std::cout << "\n===== Device " << dev << " =====" << std::endl;
////        /*print_device_properties(dev);
////        print_device_metrics(dev);
////        print_pcie_and_power_info(dev);
////        print_sdk_and_firmware_info(dev);*/
////    }
////    return 0;
////}